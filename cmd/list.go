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
	"io"
	"log"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/longkey1/gcal/internal/gcal"
	"github.com/spf13/cobra"
	"google.golang.org/api/calendar/v3"
)

var (
	listDate            string
	listSince           string
	listTo              string
	listMaxResults      int64
	listOutput          string
	listSort            string
	listIncludeDeclined bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List calendar events",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		svc, err := gcal.NewService(ctx, GetConfig())
		if err != nil {
			log.Fatalf("Unable to create gcal service: %v", err)
		}

		var events []*calendar.Event

		if listSince != "" {
			events, err = fetchRangeEvents(svc)
		} else {
			events, err = fetchDayEvents(svc)
		}
		if err != nil {
			log.Fatalf("Unable to retrieve events: %v", err)
		}

		if !listIncludeDeclined {
			events = filterDeclinedEvents(events)
		}

		sortEvents(events, listSort)

		if err := outputEvents(os.Stdout, events, listOutput); err != nil {
			log.Fatalf("Unable to output events: %v", err)
		}
	},
}

func fetchDayEvents(svc *gcal.Service) ([]*calendar.Event, error) {
	targetDate, err := time.ParseInLocation("2006-01-02", listDate, time.Now().Location())
	if err != nil {
		return nil, fmt.Errorf("invalid date format (expected YYYY-MM-DD): %w", err)
	}

	tmin := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location()).Format(time.RFC3339)
	tmax := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 23, 59, 59, 59, targetDate.Location()).Format(time.RFC3339)

	events := make([]*calendar.Event, 0)
	for _, cid := range svc.CalendarIDList {
		result, err := svc.Calendar.Events.List(cid).ShowDeleted(false).
			SingleEvents(true).TimeMin(tmin).TimeMax(tmax).OrderBy("startTime").Do()
		if err != nil {
			return nil, err
		}
		events = append(events, result.Items...)
	}
	return events, nil
}

func fetchRangeEvents(svc *gcal.Service) ([]*calendar.Event, error) {
	sinceTime, err := time.ParseInLocation("2006-01-02", listSince, time.Now().Location())
	if err != nil {
		return nil, fmt.Errorf("invalid since date format (expected YYYY-MM-DD): %w", err)
	}
	tmin := time.Date(sinceTime.Year(), sinceTime.Month(), sinceTime.Day(), 0, 0, 0, 0, sinceTime.Location()).Format(time.RFC3339)

	var tmax string
	if listTo != "" {
		toTime, err := time.ParseInLocation("2006-01-02", listTo, time.Now().Location())
		if err != nil {
			return nil, fmt.Errorf("invalid to date format (expected YYYY-MM-DD): %w", err)
		}
		tmax = time.Date(toTime.Year(), toTime.Month(), toTime.Day(), 23, 59, 59, 59, toTime.Location()).Format(time.RFC3339)
	}

	events := make([]*calendar.Event, 0)
	for _, cid := range svc.CalendarIDList {
		call := svc.Calendar.Events.List(cid).ShowDeleted(false).
			SingleEvents(true).TimeMin(tmin).OrderBy("startTime")
		if tmax != "" {
			call = call.TimeMax(tmax)
		}
		if listMaxResults > 0 {
			call = call.MaxResults(listMaxResults)
		}
		result, err := call.Do()
		if err != nil {
			return nil, err
		}
		events = append(events, result.Items...)
	}
	return events, nil
}

func filterDeclinedEvents(events []*calendar.Event) []*calendar.Event {
	filtered := make([]*calendar.Event, 0, len(events))
	for _, e := range events {
		if !isDeclined(e) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

func isDeclined(event *calendar.Event) bool {
	for _, attendee := range event.Attendees {
		if attendee.Self && attendee.ResponseStatus == "declined" {
			return true
		}
	}
	return false
}

func sortEvents(events []*calendar.Event, sortBy string) {
	switch sortBy {
	case "start":
		sort.Slice(events, func(x, y int) bool {
			if events[x].Start.DateTime != "" && events[y].Start.DateTime != "" {
				return events[x].Start.DateTime < events[y].Start.DateTime
			}
			if events[y].Start.DateTime != "" {
				return true
			}
			return false
		})
	case "updated":
		sort.Slice(events, func(x, y int) bool {
			return events[x].Updated > events[y].Updated
		})
	}
}

func outputEvents(w io.Writer, events []*calendar.Event, format string) error {
	switch format {
	case "json":
		return outputJSON(w, events)
	case "table":
		return outputTable(w, events)
	default:
		return fmt.Errorf("unknown output format: %s", format)
	}
}

func outputJSON(w io.Writer, events []*calendar.Event) error {
	b, err := json.Marshal(events)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%s", b)
	return nil
}

func outputTable(w io.Writer, events []*calendar.Event) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "START\tEND\tTITLE")

	for _, e := range events {
		start := formatEventTime(e.Start)
		end := formatEventTime(e.End)
		fmt.Fprintf(tw, "%s\t%s\t%s\n", start, end, e.Summary)
	}

	return tw.Flush()
}

func formatEventTime(t *calendar.EventDateTime) string {
	if t == nil {
		return ""
	}
	if t.DateTime != "" {
		parsed, err := time.Parse(time.RFC3339, t.DateTime)
		if err != nil {
			return t.DateTime
		}
		return parsed.Format("15:04")
	}
	if t.Date != "" {
		return "(all-day)"
	}
	return ""
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&listDate, "date", "d", time.Now().Format("2006-01-02"), "Date to list events (YYYY-MM-DD)")
	listCmd.Flags().StringVarP(&listSince, "since", "s", "", "Start date for range query (YYYY-MM-DD)")
	listCmd.Flags().StringVarP(&listTo, "to", "t", "", "End date for range query (YYYY-MM-DD)")
	listCmd.Flags().Int64VarP(&listMaxResults, "max-results", "n", 0, "Maximum number of results")
	listCmd.Flags().StringVarP(&listOutput, "output", "o", "table", "Output format: table, json")
	listCmd.Flags().StringVar(&listSort, "sort", "start", "Sort by: start, updated")
	listCmd.Flags().BoolVar(&listIncludeDeclined, "include-declined", false, "Include declined events")
}
