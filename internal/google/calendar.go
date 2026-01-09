package google

import (
	"context"
	"fmt"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// CalendarService wraps the Google Calendar API service
type CalendarService struct {
	*calendar.Service
}

// NewCalendarService creates a new Calendar service with the given authenticator
func NewCalendarService(ctx context.Context, auth Authenticator) (*CalendarService, error) {
	client, err := auth.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated client: %v", err)
	}

	var srv *calendar.Service
	if client != nil {
		srv, err = calendar.NewService(ctx, option.WithHTTPClient(client))
	} else {
		// Use Application Default Credentials (for Service Account)
		srv, err = calendar.NewService(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %v", err)
	}

	return &CalendarService{srv}, nil
}
