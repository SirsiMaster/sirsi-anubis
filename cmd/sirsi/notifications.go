package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/SirsiMaster/sirsi-pantheon/internal/notify"
	"github.com/SirsiMaster/sirsi-pantheon/internal/output"
)

var (
	notifAll    bool
	notifSource string
	notifSev    string
	notifLimit  int
	notifJSON   bool
)

var notificationsCmd = &cobra.Command{
	Use:   "notifications",
	Short: "View notification history — scan results, guard alerts, deployment outcomes",
	Long: `View the persistent history of all Pantheon operations.

Every menubar action, watchdog alert, and scan result is recorded here.
Use filters to narrow down by source deity or severity.

  sirsi notifications                   Last 20 notifications
  sirsi notifications --all             Full history
  sirsi notifications --source=anubis   Filter by deity
  sirsi notifications --severity=error  Filter by severity
  sirsi notifications --json            Machine-readable output
  sirsi notifications clear             Clear all history`,
	RunE: runNotifications,
}

var notifClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear notification history",
	RunE:  runNotifClear,
}

func init() {
	notificationsCmd.Flags().BoolVar(&notifAll, "all", false, "Show all notifications")
	notificationsCmd.Flags().StringVar(&notifSource, "source", "", "Filter by source (anubis, ka, maat, isis, ra)")
	notificationsCmd.Flags().StringVar(&notifSev, "severity", "", "Filter by severity (info, success, warning, error)")
	notificationsCmd.Flags().IntVar(&notifLimit, "limit", 20, "Number of notifications to show")
	notificationsCmd.Flags().BoolVar(&notifJSON, "json", false, "Output as JSON")
	notificationsCmd.AddCommand(notifClearCmd)
}

func runNotifications(_ *cobra.Command, _ []string) error {
	store, err := notify.Open(notify.DefaultPath())
	if err != nil {
		return fmt.Errorf("open notification store: %w", err)
	}
	defer store.Close()

	limit := notifLimit
	if notifAll {
		limit = 0
	}

	var results []notify.Notification

	switch {
	case notifSource != "":
		results, err = store.BySource(notifSource, limit)
	case notifSev != "":
		results, err = store.BySeverity(notifSev, limit)
	default:
		results, err = store.Recent(limit)
	}
	if err != nil {
		return fmt.Errorf("query notifications: %w", err)
	}

	if notifJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	if len(results) == 0 {
		fmt.Println("No notifications yet. Run a scan from the menubar or CLI.")
		return nil
	}

	output.Banner()
	count, _ := store.Count()
	fmt.Printf("Notification History (%d total)\n\n", count)

	for _, n := range results {
		icon := notify.SeverityIcon(n.Severity)
		ts := n.Timestamp.Format("Jan 02 15:04")
		dur := ""
		if n.DurationMs > 0 {
			dur = fmt.Sprintf(" (%dms)", n.DurationMs)
		}
		fmt.Printf("  %s [%s] %s — %s%s\n", icon, ts, n.Source, n.Summary, dur)
	}
	return nil
}

func runNotifClear(_ *cobra.Command, _ []string) error {
	store, err := notify.Open(notify.DefaultPath())
	if err != nil {
		return fmt.Errorf("open notification store: %w", err)
	}
	defer store.Close()

	removed, err := store.Clear()
	if err != nil {
		return fmt.Errorf("clear: %w", err)
	}
	fmt.Printf("Cleared %d notifications\n", removed)
	return nil
}
