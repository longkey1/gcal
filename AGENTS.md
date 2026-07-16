# AGENTS.md

This file provides guidance to AI coding agents (Claude Code, etc.) when working with code in this repository.

## Project Overview

gcal is a Google Calendar CLI client written in Go. It allows users to fetch and display calendar events from the command line.

## Build & Development Commands

```bash
# Build the application
go build

# Run the application
go run main.go

# Run lint (golangci-lint, version managed by go.mod tool directive)
make lint

# Release (dry run by default)
make release type=patch|minor|major

# Release (actually push)
make release type=patch dryrun=false

# Re-release existing tag
make re-release tag=v1.0.0 dryrun=false
```

## Architecture

### Command Structure (Cobra CLI)

- `main.go` - Entry point, sets version info and executes root command
- `cmd/root.go` - Root command setup, configuration loading, authentication (OAuth/Service Account)
- `cmd/list.go` - `list` subcommand: fetches events by date or date range with table/JSON output
- `cmd/auth.go` - `auth` subcommand: OAuth authentication flow
- `cmd/version.go` - `version` subcommand: displays version information

### Configuration

Config file location: `~/.config/gcal/config.toml`

Required config fields:
- `auth_type` - Authentication type: `"oauth"` or `"service_account"`
- `application_credentials` - Path to Google credentials JSON
- `user_credentials` - (OAuth only) Path to store/read user OAuth2 token
- `calendar_id_list` - List of Google Calendar IDs to query

### Google Calendar API Authentication

The application supports two authentication methods:

1. **OAuth** (default): For personal use
   - Reads OAuth client credentials from `application_credentials`
   - Stores user token in `user_credentials`
   - Initiates web flow if token doesn't exist

2. **Service Account**: For automated/server use
   - Sets `GOOGLE_APPLICATION_CREDENTIALS` environment variable
   - No user interaction required

### Key Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management
- `google.golang.org/api/calendar/v3` - Google Calendar API
- `golang.org/x/oauth2` - OAuth2 authentication
