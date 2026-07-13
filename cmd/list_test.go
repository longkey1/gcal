package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"google.golang.org/api/calendar/v3"
)

// resetListFlags restores every list flag to its default value and
// clears the Changed state so subtests do not leak into each other.
func resetListFlags(t *testing.T) {
	t.Helper()
	flags := listCmd.Flags()
	for _, name := range []string{"date", "since", "to", "max-results", "output", "sort", "include-declined"} {
		f := flags.Lookup(name)
		if f == nil {
			t.Fatalf("flag %q not registered", name)
		}
		if err := flags.Set(name, f.DefValue); err != nil {
			t.Fatalf("reset flag %q: %v", name, err)
		}
		f.Changed = false
	}
}

// validateListFlags reads package-level flag state, so these tests
// cannot run in parallel.
func TestValidateListFlags(t *testing.T) {
	tests := []struct {
		name    string
		flags   map[string]string
		wantErr string
	}{
		{
			name: "defaults are valid",
		},
		{
			name:  "date only",
			flags: map[string]string{"date": "2024-01-15"},
		},
		{
			name:  "since only",
			flags: map[string]string{"since": "2024-01-01"},
		},
		{
			name:  "since with to",
			flags: map[string]string{"since": "2024-01-01", "to": "2024-01-31"},
		},
		{
			name:    "date and since together",
			flags:   map[string]string{"date": "2024-01-15", "since": "2024-01-01"},
			wantErr: "cannot use --date and --since together",
		},
		{
			name:    "to without since",
			flags:   map[string]string{"to": "2024-01-31"},
			wantErr: "--to can only be used with --since",
		},
		{
			name:  "json output",
			flags: map[string]string{"output": "json"},
		},
		{
			name:    "invalid output format",
			flags:   map[string]string{"output": "yaml"},
			wantErr: "invalid output format: yaml (valid: table, json)",
		},
		{
			name:  "sort by updated",
			flags: map[string]string{"sort": "updated"},
		},
		{
			name:    "invalid sort option",
			flags:   map[string]string{"sort": "summary"},
			wantErr: "invalid sort option: summary (valid: start, updated)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetListFlags(t)
			for name, value := range tt.flags {
				if err := listCmd.Flags().Set(name, value); err != nil {
					t.Fatalf("set flag %q: %v", name, err)
				}
			}

			err := validateListFlags(listCmd, nil)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validateListFlags() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("validateListFlags() error = nil, want %q", tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Errorf("validateListFlags() error = %q, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestIsDeclined(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		event *calendar.Event
		want  bool
	}{
		{
			name:  "no attendees",
			event: &calendar.Event{},
			want:  false,
		},
		{
			name: "self declined",
			event: &calendar.Event{
				Attendees: []*calendar.EventAttendee{
					{Self: true, ResponseStatus: "declined"},
				},
			},
			want: true,
		},
		{
			name: "self accepted",
			event: &calendar.Event{
				Attendees: []*calendar.EventAttendee{
					{Self: true, ResponseStatus: "accepted"},
				},
			},
			want: false,
		},
		{
			name: "other attendee declined",
			event: &calendar.Event{
				Attendees: []*calendar.EventAttendee{
					{Self: false, ResponseStatus: "declined"},
				},
			},
			want: false,
		},
		{
			name: "self declined among others",
			event: &calendar.Event{
				Attendees: []*calendar.EventAttendee{
					{Self: false, ResponseStatus: "accepted"},
					{Self: true, ResponseStatus: "declined"},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := isDeclined(tt.event); got != tt.want {
				t.Errorf("isDeclined() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterDeclinedEvents(t *testing.T) {
	t.Parallel()

	kept := &calendar.Event{Summary: "kept"}
	declined := &calendar.Event{
		Summary: "declined",
		Attendees: []*calendar.EventAttendee{
			{Self: true, ResponseStatus: "declined"},
		},
	}

	tests := []struct {
		name   string
		events []*calendar.Event
		want   []string
	}{
		{
			name:   "empty input",
			events: []*calendar.Event{},
			want:   []string{},
		},
		{
			name:   "nothing declined",
			events: []*calendar.Event{kept},
			want:   []string{"kept"},
		},
		{
			name:   "declined removed",
			events: []*calendar.Event{kept, declined, kept},
			want:   []string{"kept", "kept"},
		},
		{
			name:   "all declined",
			events: []*calendar.Event{declined},
			want:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := filterDeclinedEvents(tt.events)
			if len(got) != len(tt.want) {
				t.Fatalf("filterDeclinedEvents() returned %d events, want %d", len(got), len(tt.want))
			}
			for i, e := range got {
				if e.Summary != tt.want[i] {
					t.Errorf("filterDeclinedEvents()[%d].Summary = %q, want %q", i, e.Summary, tt.want[i])
				}
			}
		})
	}
}

func TestSortEvents(t *testing.T) {
	t.Parallel()

	timed := func(summary, start string) *calendar.Event {
		return &calendar.Event{
			Summary: summary,
			Start:   &calendar.EventDateTime{DateTime: start},
		}
	}
	allDay := func(summary, date string) *calendar.Event {
		return &calendar.Event{
			Summary: summary,
			Start:   &calendar.EventDateTime{Date: date},
		}
	}
	updated := func(summary, ts string) *calendar.Event {
		return &calendar.Event{
			Summary: summary,
			Start:   &calendar.EventDateTime{DateTime: ts},
			Updated: ts,
		}
	}

	tests := []struct {
		name   string
		events []*calendar.Event
		sortBy string
		want   []string
	}{
		{
			name: "start sorts by start time",
			events: []*calendar.Event{
				timed("late", "2024-01-15T15:00:00+09:00"),
				timed("early", "2024-01-15T09:00:00+09:00"),
				timed("mid", "2024-01-15T12:00:00+09:00"),
			},
			sortBy: "start",
			want:   []string{"early", "mid", "late"},
		},
		{
			name: "start places all-day events first",
			events: []*calendar.Event{
				timed("meeting", "2024-01-15T09:00:00+09:00"),
				allDay("offsite", "2024-01-15"),
			},
			sortBy: "start",
			want:   []string{"offsite", "meeting"},
		},
		{
			name: "updated sorts descending",
			events: []*calendar.Event{
				updated("old", "2024-01-01T00:00:00Z"),
				updated("new", "2024-03-01T00:00:00Z"),
				updated("mid", "2024-02-01T00:00:00Z"),
			},
			sortBy: "updated",
			want:   []string{"new", "mid", "old"},
		},
		{
			name: "unknown sort keeps order",
			events: []*calendar.Event{
				timed("b", "2024-01-15T15:00:00+09:00"),
				timed("a", "2024-01-15T09:00:00+09:00"),
			},
			sortBy: "none",
			want:   []string{"b", "a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sortEvents(tt.events, tt.sortBy)
			for i, e := range tt.events {
				if e.Summary != tt.want[i] {
					t.Errorf("sortEvents()[%d].Summary = %q, want %q", i, e.Summary, tt.want[i])
				}
			}
		})
	}
}

func TestFormatEventTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		time *calendar.EventDateTime
		want string
	}{
		{
			name: "nil",
			time: nil,
			want: "",
		},
		{
			name: "datetime",
			time: &calendar.EventDateTime{DateTime: "2024-01-15T09:30:00+09:00"},
			want: "09:30",
		},
		{
			name: "datetime utc",
			time: &calendar.EventDateTime{DateTime: "2024-01-15T23:05:00Z"},
			want: "23:05",
		},
		{
			name: "invalid datetime returned as-is",
			time: &calendar.EventDateTime{DateTime: "not-a-time"},
			want: "not-a-time",
		},
		{
			name: "all-day date",
			time: &calendar.EventDateTime{Date: "2024-01-15"},
			want: "(all-day)",
		},
		{
			name: "empty",
			time: &calendar.EventDateTime{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := formatEventTime(tt.time); got != tt.want {
				t.Errorf("formatEventTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOutputTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		events []*calendar.Event
		want   string
	}{
		{
			name:   "no events",
			events: []*calendar.Event{},
			want:   "START  END  TITLE\n",
		},
		{
			name: "timed and all-day events",
			events: []*calendar.Event{
				{
					Summary: "Standup",
					Start:   &calendar.EventDateTime{DateTime: "2024-01-15T09:00:00+09:00"},
					End:     &calendar.EventDateTime{DateTime: "2024-01-15T10:00:00+09:00"},
				},
				{
					Summary: "Team Offsite",
					Start:   &calendar.EventDateTime{Date: "2024-01-15"},
					End:     &calendar.EventDateTime{Date: "2024-01-16"},
				},
			},
			want: "START      END        TITLE\n" +
				"09:00      10:00      Standup\n" +
				"(all-day)  (all-day)  Team Offsite\n",
		},
		{
			name: "event without times",
			events: []*calendar.Event{
				{Summary: "Mystery"},
			},
			want: "START  END  TITLE\n" +
				"            Mystery\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			if err := outputTable(&buf, tt.events); err != nil {
				t.Fatalf("outputTable() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("outputTable() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOutputJSON(t *testing.T) {
	t.Parallel()

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		if err := outputJSON(&buf, []*calendar.Event{}); err != nil {
			t.Fatalf("outputJSON() error = %v", err)
		}
		if got := buf.String(); got != "[]" {
			t.Errorf("outputJSON() = %q, want %q", got, "[]")
		}
	})

	t.Run("events round-trip", func(t *testing.T) {
		t.Parallel()

		events := []*calendar.Event{
			{
				Summary: "Standup",
				Start:   &calendar.EventDateTime{DateTime: "2024-01-15T09:00:00+09:00"},
			},
			{
				Summary: "Offsite",
				Start:   &calendar.EventDateTime{Date: "2024-01-15"},
			},
		}

		var buf bytes.Buffer
		if err := outputJSON(&buf, events); err != nil {
			t.Fatalf("outputJSON() error = %v", err)
		}

		var got []*calendar.Event
		if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatalf("output is not valid JSON: %v", err)
		}
		if len(got) != len(events) {
			t.Fatalf("decoded %d events, want %d", len(got), len(events))
		}
		if got[0].Summary != "Standup" || got[0].Start.DateTime != "2024-01-15T09:00:00+09:00" {
			t.Errorf("decoded[0] = %+v, want summary Standup with original start", got[0])
		}
		if got[1].Summary != "Offsite" || got[1].Start.Date != "2024-01-15" {
			t.Errorf("decoded[1] = %+v, want summary Offsite with original date", got[1])
		}
	})
}

func TestOutputEvents(t *testing.T) {
	t.Parallel()

	events := []*calendar.Event{
		{
			Summary: "Standup",
			Start:   &calendar.EventDateTime{DateTime: "2024-01-15T09:00:00+09:00"},
			End:     &calendar.EventDateTime{DateTime: "2024-01-15T10:00:00+09:00"},
		},
	}

	tests := []struct {
		name         string
		format       string
		wantContains string
		wantErr      string
	}{
		{
			name:         "table",
			format:       "table",
			wantContains: "START",
		},
		{
			name:         "json",
			format:       "json",
			wantContains: `"summary":"Standup"`,
		},
		{
			name:    "unknown format",
			format:  "yaml",
			wantErr: "unknown output format: yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := outputEvents(&buf, events, tt.format)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("outputEvents() error = nil, want %q", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("outputEvents() error = %q, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("outputEvents() error = %v", err)
			}
			if !strings.Contains(buf.String(), tt.wantContains) {
				t.Errorf("outputEvents() = %q, want it to contain %q", buf.String(), tt.wantContains)
			}
		})
	}
}
