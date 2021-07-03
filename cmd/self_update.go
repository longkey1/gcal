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
	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"log"

	"github.com/spf13/cobra"
)

// selfUpdateCmd represents the selfupdate command
var selfUpdateCmd = &cobra.Command{
	Use:   "self_update",
	Short: "self update binary file",
	Run: func(cmd *cobra.Command, args []string) {
		v := semver.MustParse(rootCmd.Version)
		latest, err := selfupdate.UpdateSelf(v, "longkey1/gcal")
		if err != nil {
			log.Fatalf("Unable to self update, %v", err)
		}

		if latest.Version.Equals(v) {
			// latest version is the same as current version. It means current binary is up to date.
			log.Println("Current binary is the latest version", rootCmd.Version)
		} else {
			log.Println("Successfully updated to version", latest.Version)
			log.Println("Release note:\n", latest.ReleaseNotes)
		}
	},
}

func init() {
	rootCmd.AddCommand(selfUpdateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// selfUpdateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// selfUpdateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
