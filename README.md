# gcal

Google Calendar CLI client

## Installation

Download from [Releases](https://github.com/longkey1/gcal/releases) or use `go install`:

```bash
go install github.com/longkey1/gcal@latest
```

### Supported Platforms

| OS | Architecture |
|---|---|
| Linux | amd64, arm64, armv6, armv7 |
| macOS | amd64 (Intel), arm64 (Apple Silicon) |

## Configuration

Create config file at `~/.config/gcal/config.toml`:

### OAuth (for personal use)

```toml
auth_type = "oauth"
application_credentials = "/path/to/oauth-credentials.json"
user_credentials = "/path/to/token.json"
calendar_id_list = ["primary", "your-calendar-id@group.calendar.google.com"]
```

### Service Account (for automated/server use)

```toml
auth_type = "service_account"
application_credentials = "/path/to/service-account.json"
calendar_id_list = ["your-calendar-id@group.calendar.google.com"]
```

## Usage

### auth

Authenticate with Google Calendar API (OAuth only):

```bash
gcal auth
```

1. Run `gcal auth`
2. Browser opens automatically (or copy the displayed URL manually)
3. Sign in with your Google account and grant access
4. Browser shows "Authentication successful!" - done

Token is saved to the path specified in `user_credentials`.

#### How it works

1. `gcal auth` starts a local HTTP server (e.g., `localhost:54321`)
2. OAuth URL includes `redirect_uri=http://localhost:54321/callback`
3. After Google authentication, browser redirects to `localhost` with auth code
4. Local server receives the code and exchanges it for access token
5. Token is saved to file

### day

Get events for today:

```bash
gcal day
```

Get events for tomorrow:

```bash
gcal day -d 1
```

Get events for yesterday:

```bash
gcal day -d -1
```

### updates

Get recently updated events:

```bash
gcal updates
```

Get events updated since a specific time:

```bash
gcal updates --since "2024-01-01T00:00:00+09:00"
```

### Options

```bash
# Specify config file
gcal --config /path/to/config.toml day

# Specify calendar IDs
gcal -c "calendar1@group.calendar.google.com,calendar2@group.calendar.google.com" day
```

## Output

Events are output as JSON array.
