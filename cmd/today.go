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
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/api/calendar/v3"
	"log"
	"sort"
	"time"

	"github.com/spf13/cobra"
)

// todayCmd represents the events command
var todayCmd = &cobra.Command{
	Use:   "today",
	Short: "today's event list",
	Long: `today's event list, by markdown format',
`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		srv, err := calendar.NewService(ctx)
		if err != nil {
			log.Fatalf("Unable to retrieve Calendar client: %v", err)
		}

		tmin := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location()).Format(time.RFC3339)
		tmax := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 23, 59, 59, 59, time.Now().Location()).Format(time.RFC3339)
		var es []*calendar.Event
		for _, cid := range config.CalendarIdList {
			events, err := srv.Events.List(cid).ShowDeleted(false).
				SingleEvents(true).TimeMin(tmin).TimeMax(tmax).MaxResults(10).OrderBy("startTime").Do()
			if err != nil {
				log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
			}
			es = append(es, events.Items...)
		}

		sort.Slice(es, func(x, y int) bool {
			if es[x].Start.DateTime != "" && es[y].Start.DateTime != "" {
				return es[x].Start.DateTime < es[y].Start.DateTime
			}
			if es[y].Start.DateTime != "" {
				return true
			}

			return false
		})

		b, err := json.Marshal(es)
		if err != nil {
			log.Fatalf("Unable to marshal json: %v", err)
		}
		fmt.Printf("%s", b)
	},
}

func init() {
	rootCmd.AddCommand(todayCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// eventsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// eventsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
