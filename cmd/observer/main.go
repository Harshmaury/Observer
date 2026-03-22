// @observer-project: observer
// @observer-path: cmd/observer/main.go
// observer is the platform distributed tracing daemon (ADR-014).
//
// Startup sequence:
//  1. Config
//  2. Collectors (Nexus, Forge)
//  3. Trace store
//  4. HTTP server (:8086)
//  5. Polling loop — discovers new trace IDs from Nexus events
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Harshmaury/Observer/internal/api"
	"github.com/Harshmaury/Observer/internal/collector"
	"github.com/Harshmaury/Observer/internal/config"
	"github.com/Harshmaury/Observer/internal/trace"
)

const observerVersion = "0.1.0"

func main() {
	logger := log.New(os.Stdout, "[observer] ", log.LstdFlags)
	logger.Printf("Observer v%s starting", observerVersion)
	if err := run(logger); err != nil {
		logger.Fatalf("fatal: %v", err)
	}
	logger.Println("Observer stopped cleanly")
}

func run(logger *log.Logger) error {
	// ── 1. CONFIG ────────────────────────────────────────────────────────────
	httpAddr     := config.EnvOrDefault("OBSERVER_HTTP_ADDR", config.DefaultHTTPAddr)
	nexusAddr    := config.EnvOrDefault("NEXUS_HTTP_ADDR", config.DefaultNexusAddr)
	forgeAddr    := config.EnvOrDefault("FORGE_HTTP_ADDR", config.DefaultForgeAddr)
	serviceToken := config.EnvOrDefault("OBSERVER_SERVICE_TOKEN", "")
	if serviceToken == "" {
				if os.Getenv("ENGX_AUTH_REQUIRED") == "true" {
			logger.Fatalf("FATAL: ENGX_AUTH_REQUIRED=true but OBSERVER_SERVICE_TOKEN not set — refusing to start insecurely. Set OBSERVER_SERVICE_TOKEN in ~/.nexus/service-tokens or disable with ENGX_AUTH_REQUIRED=false")
		}
		logger.Println("WARNING: OBSERVER_SERVICE_TOKEN not set — inter-service auth disabled. Set ENGX_AUTH_REQUIRED=true to enforce strict mode.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// ── 2. COLLECTORS ────────────────────────────────────────────────────────
	nexusColl := collector.NewNexusCollector(nexusAddr, serviceToken)
	forgeColl := collector.NewForgeCollector(forgeAddr, serviceToken)

	// ── 3. TRACE STORE ───────────────────────────────────────────────────────
	traceStore := trace.NewStore()

	// ── 4. HTTP SERVER ───────────────────────────────────────────────────────
	srv := api.NewServer(httpAddr, traceStore, nexusColl, forgeColl, logger)

	logger.Printf("✓ Observer ready — http=%s nexus=%s forge=%s",
		httpAddr, nexusAddr, forgeAddr)

	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.Run(ctx); err != nil && ctx.Err() == nil {
			errCh <- fmt.Errorf("api server: %w", err)
		}
	}()

	// ── 5. POLLING LOOP — trace discovery ────────────────────────────────────
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				events := nexusColl.PollRecent(ctx)
				traceIDs := map[string]int{}
				for _, e := range events {
					if e.TraceID != "" {
						traceIDs[e.TraceID]++
					}
				}
				for id, count := range traceIDs {
					traceStore.Record(id, count)
				}
				if len(traceIDs) > 0 {
					logger.Printf("discovered %d new trace(s)", len(traceIDs))
				}
			}
		}
	}()

	select {
	case sig := <-sigCh:
		logger.Printf("received %s — shutting down", sig)
	case err := <-errCh:
		logger.Printf("component error: %v — shutting down", err)
	}

	cancel()
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	<-done
	return nil
}
