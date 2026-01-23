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
	"os"
	"path/filepath"

	"github.com/longkey1/gcal/internal/gcal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile        string
	calendarIDList []string
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
	rootCmd.PersistentFlags().StringSliceVarP(&calendarIDList, "calendar-id-list", "c", []string{}, "Calendar ID List")
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

	// Don't fail if config file is not found - let individual commands handle it
	_ = viper.ReadInConfig()
}

// loadConfig loads configuration from viper and applies command line overrides
func loadConfig() (*gcal.Config, error) {
	// Check if config file was actually read
	if viper.ConfigFileUsed() == "" {
		if cfgFile != "" {
			return nil, fmt.Errorf("config file not found: %s", cfgFile)
		}
		home, _ := os.UserHomeDir()
		defaultPath := filepath.Join(home, ".config/gcal/config.toml")
		return nil, fmt.Errorf("config file not found at default location: %s", defaultPath)
	}

	config, err := gcal.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to load config: %w", err)
	}

	// Override calendar ID list from command line flag
	if len(calendarIDList) > 0 {
		config.CalendarIDList = calendarIDList
	}

	return config, nil
}
