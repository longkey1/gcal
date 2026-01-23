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

	"github.com/longkey1/gcal/internal/version"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display the version, commit SHA, and build time of gcal.`,
	Example: `  # Show full version information
  gcal version

  # Show only version number
  gcal version --short`,
	Args: cobra.NoArgs,
	RunE: runVersion,
}

func runVersion(cmd *cobra.Command, args []string) error {
	short, err := cmd.Flags().GetBool("short")
	if err != nil {
		return err
	}

	if short {
		fmt.Println(version.Short())
	} else {
		fmt.Println(version.Info())
	}

	return nil
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolP("short", "s", false, "Show only version number")
}
