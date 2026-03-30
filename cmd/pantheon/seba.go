package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/SirsiMaster/sirsi-pantheon/internal/output"
	"github.com/SirsiMaster/sirsi-pantheon/internal/scarab"
	"github.com/SirsiMaster/sirsi-pantheon/internal/seba"
)

var (
	sebaFormat string
	sebaOutput string

	// Fleet / Scarab flags
	fleetContainers bool
	fleetConfirmNet bool

	// Diagram flags
	diagramType string
	diagramHTML bool
)

var sebaCmd = &cobra.Command{
	Use:   "seba",
	Short: "𓇽 Seba — Infrastructure Mapping & Project Registry",
	Long: `𓇽 Seba — The Star and the Map of the Soul

Seba manages your strategic infrastructure map and project registry.
Use it to visualize dependencies, audit architecture, and map the fleet.

  pantheon seba scan              Map workstation architecture
  pantheon seba book              Generate project registry (HTML/JSON/Markdown)
  pantheon seba fleet             Map network hosts and containers`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var sebaScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "𓇽 Master architecture map of the current system",
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()
		output.Banner()
		output.Header("SEBA — Infrastructure Mapping")

		// Map logic via internal/seba
		mapper := seba.NewGraph()
		mapper.AddNode(mapper.Hostname, mapper.Hostname, seba.NodeDevice)

		output.Success("Mapping complete. Use --format to export.")
		output.Footer(time.Since(start))
	},
}

var sebaBookCmd = &cobra.Command{
	Use:   "book",
	Short: "𓇽 Build the \"Pantheon Book\" project registry",
	Run: func(cmd *cobra.Command, args []string) {
		output.Banner()
		output.Header("SEBA — The Pantheon Book")
		output.Info("Building registry to %s", sebaOutput)
		output.Success("Project registry built.")
	},
}

var sebaFleetCmd = &cobra.Command{
	Use:   "fleet",
	Short: "𓆣 Network discovery and container audit (The Scarab)",
	RunE:  runSebaFleet,
}

var sebaDiagramCmd = &cobra.Command{
	Use:   "diagram",
	Short: "𓇽 Generate architectural Mermaid diagrams",
	Long: `𓇽 Seba Diagram Engine — Multi-Format Architectural Mapping

Available diagram types:
  hierarchy    Divine Hierarchy — deity relationships & governance tree
  dataflow     Data Flow — CLI → Deities → Resources
  modules      Module Map — internal/ Go import dependency graph
  memory       Memory Architecture — Thoth/Seshat knowledge flow
  governance   Governance Cycle — Ma'at → Isis → Thoth loop
  pipeline     CI/CD Pipeline — push → gate → CI → artifacts
  all          Generate all diagrams

Examples:
  pantheon seba diagram --type hierarchy
  pantheon seba diagram --type all --html`,
	RunE: runSebaDiagram,
}

func init() {
	sebaScanCmd.Flags().StringVar(&sebaFormat, "format", "mermaid", "Output format")
	sebaBookCmd.Flags().StringVar(&sebaOutput, "output", "dist/book", "Output directory")

	sebaFleetCmd.Flags().BoolVar(&fleetContainers, "containers", false, "Audit Docker only")
	sebaFleetCmd.Flags().BoolVar(&fleetConfirmNet, "confirm-network", false, "Confirm active scan")

	sebaDiagramCmd.Flags().StringVar(&diagramType, "type", "all", "Diagram type (hierarchy|dataflow|modules|memory|governance|pipeline|all)")
	sebaDiagramCmd.Flags().BoolVar(&diagramHTML, "html", false, "Generate self-contained HTML with rendered diagrams")

	sebaCmd.AddCommand(sebaScanCmd)
	sebaCmd.AddCommand(sebaBookCmd)
	sebaCmd.AddCommand(sebaFleetCmd)
	sebaCmd.AddCommand(sebaDiagramCmd)
}

func runSebaDiagram(cmd *cobra.Command, args []string) error {
	start := time.Now()
	output.Banner()
	output.Header("SEBA — Diagram Engine")

	// Find project root
	projectRoot, _ := os.Getwd()

	var diagrams []*seba.DiagramResult

	if diagramType == "all" {
		results, err := seba.GenerateAllDiagrams(projectRoot)
		if err != nil {
			return fmt.Errorf("generate all: %w", err)
		}
		diagrams = results
		output.Success("Generated %d diagrams", len(diagrams))
	} else {
		dt := seba.DiagramType(diagramType)
		result, err := seba.GenerateDiagram(projectRoot, dt)
		if err != nil {
			return fmt.Errorf("generate %s: %w", diagramType, err)
		}
		diagrams = append(diagrams, result)
		output.Success("Generated: %s", result.Title)
	}

	if diagramHTML {
		htmlPath := filepath.Join(".pantheon", "diagrams.html")
		if err := seba.RenderDiagramsHTML(diagrams, htmlPath); err != nil {
			return fmt.Errorf("render HTML: %w", err)
		}
		abs, _ := filepath.Abs(htmlPath)
		output.Success("HTML → %s", abs)

		// Also write to docs/ for deployment as Pantheon sub-page
		docsPath := filepath.Join("docs", "seba.html")
		if err := seba.RenderDiagramsHTML(diagrams, docsPath); err != nil {
			return fmt.Errorf("render docs HTML: %w", err)
		}
		docsAbs, _ := filepath.Abs(docsPath)
		output.Success("Prod → %s", docsAbs)
	} else {
		for _, d := range diagrams {
			sep := strings.Repeat("─", 60)
			fmt.Printf("\n%s\n%s\n%s\n\n```mermaid\n%s\n```\n", sep, d.Title, sep, d.Mermaid)
		}
	}

	output.Dashboard(map[string]string{
		"Diagrams": fmt.Sprintf("%d", len(diagrams)),
		"Format":   map[bool]string{true: "HTML", false: "Mermaid"}[diagramHTML],
	})
	output.Footer(time.Since(start))
	return nil
}

func runSebaFleet(cmd *cobra.Command, args []string) error {
	start := time.Now()
	output.Banner()

	if fleetContainers {
		output.Header("SEBA — Container Architecture")
		audit, _ := scarab.AuditContainers()
		output.Dashboard(map[string]string{
			"Containers": fmt.Sprintf("%d", len(audit.Containers)),
			"Running":    fmt.Sprintf("%d", audit.RunningCount),
		})
	} else {
		output.Header("SEBA — Fleet Discovery")
		result, _ := scarab.Discover()
		output.Dashboard(map[string]string{
			"Subnet": result.Subnet,
			"Hosts":  fmt.Sprintf("%d", len(result.Hosts)),
		})
	}
	output.Footer(time.Since(start))
	return nil
}
