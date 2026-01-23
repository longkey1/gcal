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
	"sort"
	"time"

	"github.com/longkey1/gcal/internal/gcal"
	"github.com/spf13/cobra"
	"google.golang.org/api/calendar/v3"
)

var dateStr string

// dayCmd represents the events command
var dayCmd = &cobra.Command{
	Use:   "day",
	Short: "day's events",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		svc, err := gcal.NewService(ctx, cfg)
		if err != nil {
			return fmt.Errorf("unable to create gcal service: %w", err)
		}

		targetDate, err := time.ParseInLocation("2006-01-02", dateStr, time.Now().Location())
		if err != nil {
			return fmt.Errorf("invalid date format (expected YYYY-MM-DD): %w", err)
		}

		tmin := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location()).Format(time.RFC3339)
		tmax := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 23, 59, 59, 59, targetDate.Location()).Format(time.RFC3339)
		es := make([]*calendar.Event, 0)
		for _, cid := range svc.CalendarIDList {
			events, err := svc.Calendar.Events.List(cid).ShowDeleted(false).
				SingleEvents(true).TimeMin(tmin).TimeMax(tmax).OrderBy("startTime").Do()
			if err != nil {
				return fmt.Errorf("unable to retrieve events: %w", err)
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
			return fmt.Errorf("unable to marshal json: %w", err)
		}
		fmt.Printf("%s", b)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dayCmd)
	dayCmd.Flags().StringVarP(&dateStr, "date", "d", time.Now().Format("2006-01-02"), "Date (YYYY-MM-DD)")
}
