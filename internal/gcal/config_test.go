package gcal

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

// writeFile creates a file with the given content, creating parent
// directories as needed.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// LoadConfig reads from the global viper instance, so these tests
// reset it per case and cannot run in parallel.
func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    *Config
	}{
		{
			name: "full oauth config",
			content: `
auth_type = "oauth"
application_credentials = "/path/credentials.json"
user_credentials = "/path/token.json"
calendar_id_list = ["primary", "work@example.com"]
`,
			want: &Config{
				AuthType:                     AuthTypeOAuth,
				GoogleApplicationCredentials: "/path/credentials.json",
				GoogleUserCredentials:        "/path/token.json",
				CalendarIDList:               []string{"primary", "work@example.com"},
			},
		},
		{
			name: "service account config",
			content: `
auth_type = "service_account"
application_credentials = "/path/sa.json"
calendar_id_list = ["primary"]
`,
			want: &Config{
				AuthType:                     AuthTypeServiceAccount,
				GoogleApplicationCredentials: "/path/sa.json",
				CalendarIDList:               []string{"primary"},
			},
		},
		{
			name: "auth_type defaults to oauth",
			content: `
application_credentials = "/path/credentials.json"
user_credentials = "/path/token.json"
calendar_id_list = ["primary"]
`,
			want: &Config{
				AuthType:                     AuthTypeOAuth,
				GoogleApplicationCredentials: "/path/credentials.json",
				GoogleUserCredentials:        "/path/token.json",
				CalendarIDList:               []string{"primary"},
			},
		},
		{
			name:    "empty file defaults",
			content: "",
			want: &Config{
				AuthType: AuthTypeOAuth,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			t.Cleanup(viper.Reset)

			path := filepath.Join(t.TempDir(), "config.toml")
			writeFile(t, path, tt.content)
			viper.SetConfigFile(path)
			if err := viper.ReadInConfig(); err != nil {
				t.Fatalf("ReadInConfig() error = %v", err)
			}

			got, err := LoadConfig()
			if err != nil {
				t.Fatalf("LoadConfig() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadConfig() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  Config
		wantErr string
	}{
		{
			name: "valid oauth",
			config: Config{
				AuthType:                     AuthTypeOAuth,
				GoogleApplicationCredentials: "/path/credentials.json",
				GoogleUserCredentials:        "/path/token.json",
				CalendarIDList:               []string{"primary"},
			},
		},
		{
			name: "valid service account without user credentials",
			config: Config{
				AuthType:                     AuthTypeServiceAccount,
				GoogleApplicationCredentials: "/path/sa.json",
				CalendarIDList:               []string{"primary"},
			},
		},
		{
			name: "missing application credentials",
			config: Config{
				AuthType:              AuthTypeOAuth,
				GoogleUserCredentials: "/path/token.json",
				CalendarIDList:        []string{"primary"},
			},
			wantErr: "application_credentials is required",
		},
		{
			name: "oauth missing user credentials",
			config: Config{
				AuthType:                     AuthTypeOAuth,
				GoogleApplicationCredentials: "/path/credentials.json",
				CalendarIDList:               []string{"primary"},
			},
			wantErr: "user_credentials is required for OAuth authentication",
		},
		{
			name: "empty calendar id list",
			config: Config{
				AuthType:                     AuthTypeServiceAccount,
				GoogleApplicationCredentials: "/path/sa.json",
			},
			wantErr: "calendar_id_list is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() error = nil, want %q", tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Errorf("Validate() error = %q, want %q", err, tt.wantErr)
			}
		})
	}
}
