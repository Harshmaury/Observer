# Makefile — Observer
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -X main.observerVersion=$(VERSION)
BINDIR  := $(HOME)/bin

.PHONY: all build clean

all: build

build:
	@echo "  → observer $(VERSION)"
	@CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BINDIR)/observer ./cmd/observer/

clean:
	@rm -f $(BINDIR)/observer
