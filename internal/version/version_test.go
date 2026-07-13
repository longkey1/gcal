package version

import (
	"fmt"
	"runtime"
	"testing"
)

// setVersion overrides the build-time variables for the duration of a
// test and restores them afterwards.
func setVersion(t *testing.T, version, commit, buildTime string) {
	t.Helper()
	origVersion, origCommit, origBuildTime := Version, CommitSHA, BuildTime
	Version, CommitSHA, BuildTime = version, commit, buildTime
	t.Cleanup(func() {
		Version, CommitSHA, BuildTime = origVersion, origCommit, origBuildTime
	})
}

func TestInfo(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		commit    string
		buildTime string
	}{
		{
			name:      "defaults",
			version:   "dev",
			commit:    "unknown",
			buildTime: "unknown",
		},
		{
			name:      "release values",
			version:   "v1.2.3",
			commit:    "abc1234",
			buildTime: "2026-07-13T00:00:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setVersion(t, tt.version, tt.commit, tt.buildTime)

			want := fmt.Sprintf("Version: %s\nCommit: %s\nBuild Time: %s\nGo Version: %s",
				tt.version, tt.commit, tt.buildTime, runtime.Version())
			if got := Info(); got != want {
				t.Errorf("Info() = %q, want %q", got, want)
			}
		})
	}
}

func TestShort(t *testing.T) {
	setVersion(t, "v9.9.9", "unknown", "unknown")

	if got := Short(); got != "v9.9.9" {
		t.Errorf("Short() = %q, want %q", got, "v9.9.9")
	}
}
