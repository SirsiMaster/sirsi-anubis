package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/SirsiMaster/sirsi-pantheon/internal/output"
	"github.com/SirsiMaster/sirsi-pantheon/internal/rtk"
)

var (
	rtkMaxLines int
	rtkNoANSI   bool
	rtkNoDedup  bool
)

var rtkCmd = &cobra.Command{
	Use:   "rtk",
	Short: "Output filter — strip ANSI, dedup, truncate (subsumes RTK)",
	Long: `RTK — Output Filter

Filters terminal and tool output to reduce AI context window consumption.
Strips ANSI escape codes, deduplicates repeated lines, collapses blank runs,
and truncates oversized output with tail preservation.

  sirsi rtk filter < output.log    Filter stdin
  sirsi rtk stats < output.log     Show reduction statistics`,
}

var rtkFilterCmd = &cobra.Command{
	Use:   "filter",
	Short: "Filter output from stdin",
	RunE:  runRTKFilter,
}

var rtkStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show reduction statistics without outputting filtered text",
	RunE:  runRTKStats,
}

func init() {
	rtkFilterCmd.Flags().IntVar(&rtkMaxLines, "max-lines", 0, "Truncate after N lines (0 = unlimited)")
	rtkFilterCmd.Flags().BoolVar(&rtkNoANSI, "no-strip-ansi", false, "Don't strip ANSI escapes")
	rtkFilterCmd.Flags().BoolVar(&rtkNoDedup, "no-dedup", false, "Don't deduplicate lines")

	rtkStatsCmd.Flags().IntVar(&rtkMaxLines, "max-lines", 0, "Truncate after N lines (0 = unlimited)")

	rtkCmd.AddCommand(rtkFilterCmd, rtkStatsCmd)
}

func runRTKFilter(_ *cobra.Command, _ []string) error {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}

	cfg := rtk.DefaultConfig()
	if rtkMaxLines > 0 {
		cfg.MaxLines = rtkMaxLines
	}
	if rtkNoANSI {
		cfg.StripANSI = false
	}
	if rtkNoDedup {
		cfg.Dedup = false
	}

	f := rtk.New(cfg)
	result := f.Apply(string(input))
	fmt.Print(result.Output)
	return nil
}

func runRTKStats(_ *cobra.Command, _ []string) error {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}

	cfg := rtk.DefaultConfig()
	if rtkMaxLines > 0 {
		cfg.MaxLines = rtkMaxLines
	}

	f := rtk.New(cfg)
	result := f.Apply(string(input))

	output.Banner()
	fmt.Printf("RTK Output Filter Statistics\n\n")
	fmt.Printf("  Original:  %d bytes\n", result.OriginalBytes)
	fmt.Printf("  Filtered:  %d bytes\n", result.FilteredBytes)
	fmt.Printf("  Reduction: %.1f%%\n", (1-result.Ratio)*100)
	fmt.Printf("  Lines removed: %d\n", result.LinesRemoved)
	fmt.Printf("  Duplicates collapsed: %d\n", result.DupsCollapsed)
	fmt.Printf("  Truncated: %v\n", result.Truncated)
	return nil
}
