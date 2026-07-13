package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

// setLoadConfigGlobals overrides the package-level flag variables used
// by loadConfig, resets viper, and restores everything after the test.
func setLoadConfigGlobals(t *testing.T, configPath string, calendarIDs []string) {
	t.Helper()
	origCfgFile, origCalendarIDList := cfgFile, calendarIDList
	cfgFile, calendarIDList = configPath, calendarIDs
	viper.Reset()
	t.Cleanup(func() {
		cfgFile, calendarIDList = origCfgFile, origCalendarIDList
		viper.Reset()
	})
}

// writeConfigFile writes a config file into a temp dir and returns its path.
func writeConfigFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

const validConfigTOML = `
auth_type = "oauth"
application_credentials = "/path/credentials.json"
user_credentials = "/path/token.json"
calendar_id_list = ["primary", "work@example.com"]
`

// loadConfig reads package-level flag variables and the global viper
// instance, so these tests cannot run in parallel.
func TestLoadConfig(t *testing.T) {
	t.Run("reads config file", func(t *testing.T) {
		path := writeConfigFile(t, validConfigTOML)
		setLoadConfigGlobals(t, path, nil)

		cfg, err := loadConfig()
		if err != nil {
			t.Fatalf("loadConfig() error = %v", err)
		}
		if cfg.GoogleApplicationCredentials != "/path/credentials.json" {
			t.Errorf("GoogleApplicationCredentials = %q, want %q", cfg.GoogleApplicationCredentials, "/path/credentials.json")
		}
		if cfg.GoogleUserCredentials != "/path/token.json" {
			t.Errorf("GoogleUserCredentials = %q, want %q", cfg.GoogleUserCredentials, "/path/token.json")
		}
		want := []string{"primary", "work@example.com"}
		if !reflect.DeepEqual(cfg.CalendarIDList, want) {
			t.Errorf("CalendarIDList = %v, want %v", cfg.CalendarIDList, want)
		}
	})

	t.Run("flag overrides calendar id list", func(t *testing.T) {
		path := writeConfigFile(t, validConfigTOML)
		setLoadConfigGlobals(t, path, []string{"override@example.com"})

		cfg, err := loadConfig()
		if err != nil {
			t.Fatalf("loadConfig() error = %v", err)
		}
		want := []string{"override@example.com"}
		if !reflect.DeepEqual(cfg.CalendarIDList, want) {
			t.Errorf("CalendarIDList = %v, want %v", cfg.CalendarIDList, want)
		}
	})

	t.Run("missing config file", func(t *testing.T) {
		setLoadConfigGlobals(t, filepath.Join(t.TempDir(), "nosuch.toml"), nil)

		if _, err := loadConfig(); err == nil {
			t.Error("loadConfig() error = nil, want error")
		}
	})
}
