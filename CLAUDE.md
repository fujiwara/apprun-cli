# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

apprun-cli is a command-line interface for managing applications on Sakura Cloud's AppRun β service. It provides deployment, configuration management, and traffic control capabilities.

## Development Commands

### Build and Test
- `make build` - Build the binary to `./apprun-cli`
- `make test` or `go test -v ./...` - Run all tests
- `go test -race ./...` - Run tests with race detector (used in CI)
- `make install` - Install to `$GOPATH/bin`
- `make dist` - Build release binaries with goreleaser

### Testing with Real Application
The binary accepts application definition files via `--app` flag or `APPRUN_CLI_APP` environment variable. Use `testdata/app.jsonnet` as a reference for testing.

## Architecture

### CLI Structure
The project uses `alecthomas/kong` for CLI argument parsing. Main entry point is in `cmd/apprun-cli/main.go`, which calls `cli.CLI.Run()` in `cli.go:37`.

All commands are defined as methods on the `CLI` struct:
- `runList()`, `runInit()`, `runDeploy()`, `runDiff()`, etc.
- Each command corresponds to a subcommand (list, init, deploy, etc.)

### Application Model
The `Application` struct (app.go:17) is the core data model that combines:
- `v1.PostApplicationBody` - application configuration
- `v1.PatchPacketFilter` - packet filter settings

This struct is used for:
- Loading from Jsonnet/JSON files via `LoadApplication()` (app.go:109)
- Converting to API request bodies via `PostApplicationBody()` and `toUpdateV1Application()`
- Marshaling to JSON for output

### Jsonnet Integration
The project uses `fujiwara/jsonnet-armed` for enhanced Jsonnet support with custom native functions:

**Native Functions** (defined in jsonnet.go:14):
- `must_env(key)` - Read environment variable (error if not set)
- `env(key, default)` - Read environment variable with default
- `tfstate(path)` - Lookup values from Terraform state (requires `--tfstate` flag)
- `secret_value(vault_id, secret_name, version)` - Retrieve secrets from Sakura Cloud Secret Manager

Setup happens in `setupVM()` (jsonnet.go:14), which:
1. Initializes default native functions
2. Adds Secret Manager functions
3. Optionally adds tfstate lookup functions if `--tfstate` is specified

### API Client
Uses `sacloud/apprun-api-go` for AppRun API interactions and `sacloud/secretmanager-api-go` for Secret Manager. The client is initialized with credentials from environment variables:
- `SAKURACLOUD_ACCESS_TOKEN`
- `SAKURACLOUD_ACCESS_TOKEN_SECRET`

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

### Traffic Management
Traffic shifting (`traffics.go`) supports gradual rollout with:
- `--shift-to` - target version
- `--rate` - percentage per period
- `--period` - time interval
- `--rollback-on-failure` - auto-rollback on failure
