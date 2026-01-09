package gcal

import (
	"context"

	"github.com/longkey1/gcal/internal/google"
)

// Service represents the gcal application service
type Service struct {
	Calendar       *google.CalendarService
	CalendarIDList []string
}

// NewService creates a new gcal service based on the configuration
func NewService(ctx context.Context, config *Config) (*Service, error) {
	auth := newAuthenticator(config)

	calSvc, err := google.NewCalendarService(ctx, auth)
	if err != nil {
		return nil, err
	}

	return &Service{
		Calendar:       calSvc,
		CalendarIDList: config.CalendarIDList,
	}, nil
}

func newAuthenticator(config *Config) google.Authenticator {
	switch config.AuthType {
	case AuthTypeServiceAccount:
		return google.NewServiceAccountAuthenticator(config.GoogleApplicationCredentials)
	case AuthTypeOAuth:
		fallthrough
	default:
		return google.NewOAuthAuthenticator(
			config.GoogleApplicationCredentials,
			config.GoogleUserCredentials,
		)
	}
}
