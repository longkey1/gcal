# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`gcal` is a Google Calendar CLI client written in Go that supports both OAuth and Service Account authentication. The project uses cobra for CLI commands, viper for configuration management, and the Google Calendar API v3.

## Commands

### Building

```bash
make build          # Build binary to ./bin/gcal
go build -o ./bin/gcal
```

### Release Management

```bash
# Preview release (dry run)
make release type=patch dryrun=true   # v1.2.3 -> v1.2.4
make release type=minor dryrun=true   # v1.2.3 -> v1.3.0
make release type=major dryrun=true   # v1.2.3 -> v2.0.0

# Create and push release tag
make release type=patch dryrun=false

# Re-release existing tag (rebuild on GitHub Actions)
make re-release dryrun=true           # Preview re-release of latest tag
make re-release tag=v1.2.3 dryrun=false  # Re-release specific tag
```

After pushing a tag, GitHub Actions (.github/workflows/gorelease.yml) automatically builds multi-platform binaries using GoReleaser (.goreleaser.yaml).

### Installing GoReleaser

```bash
make tools  # Install goreleaser
```

### Testing the Binary

```bash
# Authenticate (OAuth only)
./bin/gcal auth

# Get today's events
./bin/gcal day

# Get specific date events
./bin/gcal day -d 2024-01-15

# Get updated events
./bin/gcal updates -s 2024-01-01

# Override calendar IDs
./bin/gcal -c "cal1@group.calendar.google.com,cal2@group.calendar.google.com" day
```

## Architecture

### Directory Structure

```
.
├── cmd/                    # Cobra command definitions
│   ├── root.go            # Root command and config initialization
│   ├── auth.go            # OAuth authentication flow
│   ├── day.go             # List events for a specific day
│   ├── updates.go         # List recently updated events
│   └── version.go         # Version information
├── internal/
│   ├── gcal/              # Application layer
│   │   ├── config.go      # Configuration structs and loading
│   │   └── service.go     # Service initialization
│   ├── google/            # Google API layer
│   │   ├── auth.go        # Authentication (OAuth and Service Account)
│   │   └── calendar.go    # Calendar API client wrapper
│   └── version/           # Version info package
└── main.go                # Entry point with version variables
```

### Configuration

Config file: `~/.config/gcal/config.toml`

The config structure supports two authentication types defined in `internal/gcal/config.go`:

- **OAuth** (`auth_type = "oauth"`): For personal use with user credentials
  - Requires `application_credentials` (OAuth client secret JSON)
  - Requires `user_credentials` (token.json path for storing user tokens)

- **Service Account** (`auth_type = "service_account"`): For automated/server use
  - Requires `application_credentials` (service account JSON)
  - Uses Application Default Credentials

Both require `calendar_id_list` array with calendar IDs to query.

### Authentication Flow

The authentication layer (`internal/google/auth.go`) uses the **Authenticator** interface pattern:

```go
type Authenticator interface {
    GetClient(ctx context.Context) (*http.Client, error)
}
```

Two implementations:
- **OAuthAuthenticator**: Reads token from file, returns configured HTTP client
- **ServiceAccountAuthenticator**: Sets GOOGLE_APPLICATION_CREDENTIALS env var, returns nil (signals to use ADC)

OAuth authentication (`gcal auth` command):
1. Finds random available port and starts local HTTP server
2. Opens browser with OAuth consent URL pointing to localhost callback
3. Waits for Google redirect with authorization code
4. Exchanges code for access/refresh token
5. Saves token to configured `user_credentials` path
6. Shuts down local server

### Service Initialization

`internal/gcal/service.go` creates the application service:

1. `NewService()` reads config and selects appropriate authenticator based on `auth_type`
2. Creates `google.CalendarService` wrapper around Google Calendar API client
3. Returns `Service` with calendar client and calendar ID list

### Event Retrieval Pattern

Commands like `day` and `updates`:
1. Create gcal Service with config
2. Iterate over all calendar IDs in `CalendarIDList`
3. Call Calendar API for each calendar ID
4. Collect and merge events from all calendars
5. Sort by start time
6. Output as JSON array

### Version Information

Version is injected at build time via ldflags in `.goreleaser.yaml`:
- `-X main.ver={{.Version}}`
- `-X main.commit={{.Commit}}`
- `-X main.date={{.Date}}`

Variables are set in `main.go` and transferred to `internal/version` package.

## Date Format

All date inputs use **YYYY-MM-DD** format (e.g., `2024-01-15`), not relative dates or RFC3339.
