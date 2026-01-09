/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Config struct {
	AuthType                     string   `mapstructure:"auth_type"` // "service_account" or "oauth"
	GoogleApplicationCredentials string   `mapstructure:"application_credentials"`
	GoogleUserCredentials        string   `mapstructure:"user_credentials"`
	CalendarIdList               []string `mapstructure:"calendar_id_list"`
}

var (
	cfgFile        string
	config         Config
	calendarIdList []string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gcal",
	Short: "Google Calendar cli client",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/gcal/config.toml)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().StringSliceVarP(&calendarIdList, "calendar-id-list", "c", []string{}, "Calendar ID List")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(filepath.Join(home, ".config/gcal"))
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Unable to read config file, %v", err)
	}
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	// Set calendarIdList
	if len(calendarIdList) == 0 {
		calendarIdList = config.CalendarIdList
	}
}

func SetVersionInfo(version, commit, date string) {
	rootCmd.Version = fmt.Sprintf("%s (Built on %s from Git SHA %s)", version, date, commit)
}

// NewCalendarService creates a Calendar service based on the configured auth type
func NewCalendarService(ctx context.Context) (*calendar.Service, error) {
	switch config.AuthType {
	case "service_account":
		return newServiceAccountCalendarService(ctx)
	case "oauth":
		return newOAuthCalendarService(ctx)
	default:
		// Default to oauth for backward compatibility
		return newOAuthCalendarService(ctx)
	}
}

func newServiceAccountCalendarService(ctx context.Context) (*calendar.Service, error) {
	if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", config.GoogleApplicationCredentials); err != nil {
		return nil, fmt.Errorf("unable to set GOOGLE_APPLICATION_CREDENTIALS: %v", err)
	}
	return calendar.NewService(ctx)
}

func newOAuthCalendarService(ctx context.Context) (*calendar.Service, error) {
	b, err := os.ReadFile(config.GoogleApplicationCredentials)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	oauthConfig, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	client := getClient(oauthConfig, config.GoogleUserCredentials)
	return calendar.NewService(ctx, option.WithHTTPClient(client))
}

func getClient(config *oauth2.Config, tokenFile string) *http.Client {
	token, err := tokenFromFile(tokenFile)
	if err != nil {
		token = getTokenFromWeb(config)
		saveToken(tokenFile, token)
	}
	return config.Client(context.Background(), token)
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
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

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
