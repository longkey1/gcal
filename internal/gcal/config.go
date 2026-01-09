package gcal

import (
	"fmt"

	"github.com/spf13/viper"
)

// AuthType represents the authentication type
type AuthType string

const (
	AuthTypeOAuth          AuthType = "oauth"
	AuthTypeServiceAccount AuthType = "service_account"
)

// Config holds the configuration for gcal
type Config struct {
	AuthType                     AuthType `mapstructure:"auth_type"`
	GoogleApplicationCredentials string   `mapstructure:"application_credentials"`
	GoogleUserCredentials        string   `mapstructure:"user_credentials"`
	CalendarIDList               []string `mapstructure:"calendar_id_list"`
}

// LoadConfig loads configuration from viper
func LoadConfig() (*Config, error) {
	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %v", err)
	}

	// Default to OAuth if not specified
	if config.AuthType == "" {
		config.AuthType = AuthTypeOAuth
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.GoogleApplicationCredentials == "" {
		return fmt.Errorf("application_credentials is required")
	}

	if c.AuthType == AuthTypeOAuth && c.GoogleUserCredentials == "" {
		return fmt.Errorf("user_credentials is required for OAuth authentication")
	}

	if len(c.CalendarIDList) == 0 {
		return fmt.Errorf("calendar_id_list is required")
	}

	return nil
}
