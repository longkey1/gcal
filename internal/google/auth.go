package google

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// Authenticator provides HTTP client for Google API authentication
type Authenticator interface {
	GetClient(ctx context.Context) (*http.Client, error)
}

// OAuthAuthenticator implements Authenticator using OAuth2
type OAuthAuthenticator struct {
	credentialsFile string
	tokenFile       string
}

// NewOAuthAuthenticator creates a new OAuthAuthenticator
func NewOAuthAuthenticator(credentialsFile, tokenFile string) *OAuthAuthenticator {
	return &OAuthAuthenticator{
		credentialsFile: credentialsFile,
		tokenFile:       tokenFile,
	}
}

// GetClient returns an authenticated HTTP client using OAuth2
func (a *OAuthAuthenticator) GetClient(ctx context.Context) (*http.Client, error) {
	b, err := os.ReadFile(a.credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	token, err := a.tokenFromFile()
	if err != nil {
		token = a.getTokenFromWeb(config)
		if err := a.saveToken(token); err != nil {
			return nil, fmt.Errorf("unable to save token: %v", err)
		}
	}

	return config.Client(ctx, token), nil
}

func (a *OAuthAuthenticator) tokenFromFile() (*oauth2.Token, error) {
	f, err := os.Open(a.tokenFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

func (a *OAuthAuthenticator) getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the full url: \n%v\n", authURL)

	var fullURL string
	if _, err := fmt.Scan(&fullURL); err != nil {
		log.Fatalf("Unable to read authorization fullURL: %v", err)
	}

	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		log.Fatal(err)
	}

	queryParams := parsedURL.Query()
	code := queryParams.Get("code")
	fmt.Println("Authorization Code:", code)

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return token
}

func (a *OAuthAuthenticator) saveToken(token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", a.tokenFile)
	f, err := os.Create(a.tokenFile)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

// ServiceAccountAuthenticator implements Authenticator using Service Account
type ServiceAccountAuthenticator struct {
	credentialsFile string
}

// NewServiceAccountAuthenticator creates a new ServiceAccountAuthenticator
func NewServiceAccountAuthenticator(credentialsFile string) *ServiceAccountAuthenticator {
	return &ServiceAccountAuthenticator{
		credentialsFile: credentialsFile,
	}
}

// GetClient returns an authenticated HTTP client using Service Account
func (a *ServiceAccountAuthenticator) GetClient(ctx context.Context) (*http.Client, error) {
	if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", a.credentialsFile); err != nil {
		return nil, fmt.Errorf("unable to set GOOGLE_APPLICATION_CREDENTIALS: %v", err)
	}
	// Return nil to use Application Default Credentials
	return nil, nil
}
