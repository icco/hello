# AGENTS.md

Guidance for coding agents working on hello.

## Project Overview

A lightweight HTTP hello/health microservice in Go (`github.com/icco/hello`).

## Commands

```sh
go test ./...    # Run tests
go vet ./...     # Vet code
go run main.go   # Run locally (port 8080 by default)
go build .       # Build binary
```

## Conventions

- Follow icco Go conventions (`github.com/icco/gutil` for logging and HTTP helpers).
- PR titles and commits must follow Conventional Commits with lowercase subjects.
