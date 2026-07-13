package gcal

import (
	"testing"

	"github.com/longkey1/gcal/internal/google"
)

func TestNewAuthenticator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		authType AuthType
		wantType string
	}{
		{
			name:     "service account",
			authType: AuthTypeServiceAccount,
			wantType: "*google.ServiceAccountAuthenticator",
		},
		{
			name:     "oauth",
			authType: AuthTypeOAuth,
			wantType: "*google.OAuthAuthenticator",
		},
		{
			name:     "empty defaults to oauth",
			authType: "",
			wantType: "*google.OAuthAuthenticator",
		},
		{
			name:     "unknown defaults to oauth",
			authType: "something-else",
			wantType: "*google.OAuthAuthenticator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			auth := newAuthenticator(&Config{
				AuthType:                     tt.authType,
				GoogleApplicationCredentials: "/path/credentials.json",
				GoogleUserCredentials:        "/path/token.json",
			})

			switch tt.wantType {
			case "*google.ServiceAccountAuthenticator":
				if _, ok := auth.(*google.ServiceAccountAuthenticator); !ok {
					t.Errorf("newAuthenticator() = %T, want %s", auth, tt.wantType)
				}
			case "*google.OAuthAuthenticator":
				if _, ok := auth.(*google.OAuthAuthenticator); !ok {
					t.Errorf("newAuthenticator() = %T, want %s", auth, tt.wantType)
				}
			}
		})
	}
}
