package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/SirsiMaster/sirsi-pantheon/internal/horus"
	"github.com/SirsiMaster/sirsi-pantheon/internal/output"
)

var (
	horusKind   string
	horusFilter string
)

var horusCmd = &cobra.Command{
	Use:   "horus",
	Short: "Structural code graph — symbols, outlines, context (subsumes Code Review Graph)",
	Long: `𓂀 Horus — Structural Code Graph

Extracts symbols (types, functions, methods, interfaces) from Go source
and serves compact outlines and signatures instead of full files.
8-49x smaller than reading full source.

  sirsi horus scan .                   Build symbol graph
  sirsi horus outline internal/mcp/tools.go   File outline (no bodies)
  sirsi horus symbols --kind=func      List all functions
  sirsi horus context NewServer        Show symbol context`,
}

var horusScanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Build symbol graph for a project",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runHorusScan,
}

var horusOutlineCmd = &cobra.Command{
	Use:   "outline <file>",
	Short: "Print compact file outline (declarations only, no bodies)",
	Args:  cobra.ExactArgs(1),
	RunE:  runHorusOutline,
}

var horusSymbolsCmd = &cobra.Command{
	Use:   "symbols",
	Short: "List symbols matching filters",
	RunE:  runHorusSymbols,
}

var horusContextCmd = &cobra.Command{
	Use:   "context <symbol>",
	Short: "Show minimal context for a symbol",
	Args:  cobra.ExactArgs(1),
	RunE:  runHorusContext,
}

var horusStatsCmd = &cobra.Command{
	Use:   "stats [path]",
	Short: "Print graph statistics",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runHorusStats,
}

func init() {
	horusSymbolsCmd.Flags().StringVar(&horusKind, "kind", "", "Filter by kind: type, func, method, interface, struct, const, var")
	horusSymbolsCmd.Flags().StringVar(&horusFilter, "filter", "", "Filter by name pattern (glob with *)")

	horusCmd.AddCommand(horusScanCmd, horusOutlineCmd, horusSymbolsCmd, horusContextCmd, horusStatsCmd)
}

func horusParseDir(path string) (*horus.SymbolGraph, error) {
	if path == "" || path == "." {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("determine working directory: %w", err)
		}
	}

	p := horus.NewGoParser()
	return p.ParseDir(path)
}

func runHorusScan(_ *cobra.Command, args []string) error {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	output.Banner()
	fmt.Printf("𓂀 Horus scanning %s...\n\n", path)

	graph, err := horusParseDir(path)
	if err != nil {
		return err
	}

	fmt.Printf("Symbol Graph Built\n")
	fmt.Printf("  Files:      %d\n", graph.Stats.Files)
	fmt.Printf("  Packages:   %d\n", graph.Stats.Packages)
	fmt.Printf("  Types:      %d\n", graph.Stats.Types)
	fmt.Printf("  Interfaces: %d\n", graph.Stats.Interfaces)
	fmt.Printf("  Functions:  %d\n", graph.Stats.Functions)
	fmt.Printf("  Methods:    %d\n", graph.Stats.Methods)
	fmt.Printf("  Lines:      %d\n", graph.Stats.TotalLines)
	fmt.Printf("  Symbols:    %d total\n", len(graph.Symbols))
	return nil
}

func runHorusOutline(_ *cobra.Command, args []string) error {
	path := args[0]
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	p := horus.NewGoParser()
	symbols, err := p.ParseFile(path, src)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	graph := &horus.SymbolGraph{Root: ".", Symbols: symbols}
	q := horus.NewQuery(graph)
	fmt.Println(q.FileOutline(path))

	ratio := float64(len(q.FileOutline(path))) / float64(len(src)) * 100
	fmt.Printf("\n── %.0f%% of original (%d → %d bytes) ──\n", ratio, len(src), len(q.FileOutline(path)))
	return nil
}

func runHorusSymbols(_ *cobra.Command, _ []string) error {
	graph, err := horusParseDir(".")
	if err != nil {
		return err
	}

	q := horus.NewQuery(graph)
	symbols := graph.Symbols

	if horusKind != "" {
		symbols = q.ByKind(horus.SymbolKind(horusKind))
	}
	if horusFilter != "" {
		symbols = q.MatchSymbols(horusFilter)
	}

	for _, s := range symbols {
		if s.Kind == horus.KindPackage {
			continue
		}
		prefix := " "
		if s.Exported {
			prefix = "▸"
		}
		sig := s.Signature
		if sig == "" {
			sig = fmt.Sprintf("%s %s", s.Kind, s.Name)
		}
		fmt.Printf("%s %s:%d  %s\n", prefix, s.File, s.Line, sig)
	}
	return nil
}

func runHorusContext(_ *cobra.Command, args []string) error {
	graph, err := horusParseDir(".")
	if err != nil {
		return err
	}

	q := horus.NewQuery(graph)
	fmt.Println(q.ContextFor(args[0]))
	return nil
}

func runHorusStats(_ *cobra.Command, args []string) error {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	graph, err := horusParseDir(path)
	if err != nil {
		return err
	}

	output.Banner()
	fmt.Printf("𓂀 Horus Graph Stats for %s\n\n", path)
	fmt.Printf("  Files:      %d\n", graph.Stats.Files)
	fmt.Printf("  Packages:   %d (%v)\n", graph.Stats.Packages, graph.Packages)
	fmt.Printf("  Types:      %d\n", graph.Stats.Types)
	fmt.Printf("  Interfaces: %d\n", graph.Stats.Interfaces)
	fmt.Printf("  Functions:  %d\n", graph.Stats.Functions)
	fmt.Printf("  Methods:    %d\n", graph.Stats.Methods)
	fmt.Printf("  Lines:      %d\n", graph.Stats.TotalLines)
	return nil
}
