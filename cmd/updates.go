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
	"sort"
	"time"

	"github.com/longkey1/gcal/internal/gcal"
	"github.com/spf13/cobra"
	"google.golang.org/api/calendar/v3"
)

var sinceDate string

// updatesCmd represents the events command
var updatesCmd = &cobra.Command{
	Use:   "updates",
	Short: "recent updates events",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		svc, err := gcal.NewService(ctx, GetConfig())
		if err != nil {
			log.Fatalf("Unable to create gcal service: %v", err)
		}

		tmin := time.Now().Format(time.RFC3339)
		sinceTime, err := time.ParseInLocation("2006-01-02", sinceDate, time.Now().Location())
		if err != nil {
			log.Fatalf("Invalid date format (expected YYYY-MM-DD): %v", err)
		}
		umin := sinceTime.Format(time.RFC3339)
		var es []*calendar.Event
		for _, cid := range svc.CalendarIDList {
			events, err := svc.Calendar.Events.List(cid).ShowDeleted(false).
				SingleEvents(true).TimeMin(tmin).UpdatedMin(umin).MaxResults(10).OrderBy("updated").Do()
			if err != nil {
				log.Fatalf("Unable to retrieve events: %v", err)
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
	rootCmd.AddCommand(updatesCmd)
	updatesCmd.Flags().StringVarP(&sinceDate, "since", "s", time.Now().Format("2006-01-02"), "Since date (YYYY-MM-DD)")
}
