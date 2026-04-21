# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

apprun-cli is a command-line interface for managing applications on Sakura Cloud's AppRun Shared service. It provides deployment, configuration management, and traffic control capabilities. AppRun Dedicated is a separate service with a different API and is out of scope.

## Development Commands

- Go 1.24+ required
- `make build` - Build the binary to `./apprun-cli`
- `make test` or `go test -v ./...` - Run all tests
- `go test -v -run TestLoadApplication ./...` - Run a single test
- `go test -race ./...` - Run tests with race detector (used in CI)
- `make install` - Install to `$GOPATH/bin`
- `make dist` - Build release binaries with goreleaser

## Architecture

### CLI Structure
The root package is `cli` (import path: `github.com/fujiwara/apprun-cli`). Main entry point is in `cmd/apprun-cli/main.go`, which calls `cli.CLI.Run()` in `cli.go:37`.

The project uses `alecthomas/kong` for CLI argument parsing. All commands are defined as methods on the `CLI` struct (e.g., `runList()`, `runDeploy()`), with each subcommand's options in a separate `*Option` struct. Each command implementation lives in its own file (deploy.go, diff.go, etc.).

### Application Model
The `Application` struct (app.go:17) is the core data model that combines:
- `v1.PostApplicationBody` - application configuration
- `v1.PatchPacketFilter` - packet filter settings

This struct is used for:
- Loading from Jsonnet/JSON files via `LoadApplication()` (app.go:109)
- Converting to API request bodies via `PostApplicationBody()` and `toUpdateV1Application()`
- Marshaling to JSON for output

### Jsonnet Integration
The project uses `fujiwara/jsonnet-armed` for enhanced Jsonnet support with custom native functions defined in `jsonnet.go`. `setupVM()` initializes the Jsonnet VM with:
1. Default native functions (`must_env`, `env` - from jsonnet-armed)
2. Secret Manager function (`secret(vault_id, name)` from sakura-secrets-cli, plus deprecated `secret_value` wrapper)
3. Terraform state lookup (`tfstate(path)`) - only when `--tfstate` is specified

### API Client
Uses `sacloud/apprun-api-go` (v0.8+, ogen-generated) for AppRun API interactions and `fujiwara/sakura-secrets-cli` for Secret Manager integration. `cli.New` constructs a `saclient.Client` (from `sacloud/saclient-go`), which reads credentials from `usacloud` profiles or environment variables, then wraps it via `apprun.NewClient` to get a `*v1.Client`. Relevant environment variables:
- `SAKURA_ACCESS_TOKEN` (or `SAKURACLOUD_ACCESS_TOKEN`)
- `SAKURA_ACCESS_TOKEN_SECRET` (or `SAKURACLOUD_ACCESS_TOKEN_SECRET`)

## Testing

Tests are in `*_test.go` files. The project uses:
- Standard `testing` package
- `google/go-cmp` for deep equality comparisons
- `testdata/` directory for test fixtures

Use `export_test.go` to export internal functions for testing (Go testing pattern).

## Key Implementation Details

### Type Conversions
The codebase uses JSON marshaling/unmarshaling for type conversions between internal types and API types (see `fromV1Application()`, `toUpdateV1Application()` in app.go:43-67).

### Error Handling
Application lookup returns `ErrNotFound` when an application with the specified name doesn't exist. This allows differentiation between "not found" and other errors.

### Iterator Pattern
The codebase uses Go 1.23+ range-over-func iterators for paginated API results (e.g., `allApplications()`, `AllTraffics()` in traffics.go).

### Traffic Management
Traffic shifting (`traffics.go`) supports gradual rollout with `--shift-to`, `--rate`, `--period`, and `--rollback-on-failure` flags.

## TODO

- Remove deprecated `secret_value()` Jsonnet function (`deprecatedSecretValueFunc` in jsonnet.go) when releasing v1.
