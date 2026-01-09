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
