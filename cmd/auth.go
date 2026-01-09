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
	"fmt"
	"log"
	"os"

	"github.com/longkey1/gcal/internal/gcal"
	"github.com/longkey1/gcal/internal/google"
	"github.com/spf13/cobra"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Google Calendar API using OAuth",
	Long: `Authenticate with Google Calendar API using OAuth.
This command initiates the OAuth flow to obtain and save access tokens.
Only applicable when auth_type is set to "oauth" in config.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		if cfg.AuthType != gcal.AuthTypeOAuth {
			log.Fatalf("auth command is only available for OAuth authentication (current: %s)", cfg.AuthType)
		}

		// Check if token already exists
		if _, err := os.Stat(cfg.GoogleUserCredentials); err == nil {
			fmt.Printf("Token file already exists: %s\n", cfg.GoogleUserCredentials)
			fmt.Print("Do you want to re-authenticate? [y/N]: ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Cancelled.")
				return
			}
		}

		// Run OAuth flow
		auth := google.NewOAuthAuthenticator(
			cfg.GoogleApplicationCredentials,
			cfg.GoogleUserCredentials,
		)

		if err := auth.Authenticate(); err != nil {
			log.Fatalf("Authentication failed: %v", err)
		}

		fmt.Println("Authentication successful!")
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
}
