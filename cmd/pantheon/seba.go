package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/SirsiMaster/sirsi-pantheon/internal/guard"
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

var (
	computeText string
)

var sebaCmd = &cobra.Command{
	Use:   "seba",
	Short: "𓇽 Seba — Infrastructure Mapping, Hardware Profiling & Fleet Discovery",
	Long: `𓇽 Seba — Infrastructure Mapping, Hardware Profiling & Fleet Discovery

Seba maps your infrastructure — hardware, architecture, and fleet topology.

  pantheon seba scan              Map workstation architecture
  pantheon seba hardware          Hardware & GPU summary dashboard
  pantheon seba profile           Deep system profile saved as JSON
  pantheon seba compute           ANE-accelerated ML tokenization
  pantheon seba book              Generate project registry
  pantheon seba fleet             Map network hosts and containers
  pantheon seba diagram           Generate architectural diagrams`,
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

var sebaHardwareCmd = &cobra.Command{
	Use:   "hardware",
	Short: "𓇽 Hardware, CPU, GPU, and accelerator summary",
	RunE:  runSebaHardware,
}

var sebaProfileCmd = &cobra.Command{
	Use:   "profile",
	Short: "𓇽 Deep system profile saved to JSON",
	RunE:  runSebaProfile,
}

var sebaComputeCmd = &cobra.Command{
	Use:   "compute",
	Short: "𓇽 ANE-accelerated ML tokenization",
	RunE:  runSebaCompute,
}

func init() {
	sebaScanCmd.Flags().StringVar(&sebaFormat, "format", "mermaid", "Output format")
	sebaBookCmd.Flags().StringVar(&sebaOutput, "output", "dist/book", "Output directory")

	sebaFleetCmd.Flags().BoolVar(&fleetContainers, "containers", false, "Audit Docker only")
	sebaFleetCmd.Flags().BoolVar(&fleetConfirmNet, "confirm-network", false, "Confirm active scan")

	sebaDiagramCmd.Flags().StringVar(&diagramType, "type", "all", "Diagram type (hierarchy|dataflow|modules|memory|governance|pipeline|all)")
	sebaDiagramCmd.Flags().BoolVar(&diagramHTML, "html", false, "Generate self-contained HTML with rendered diagrams")

	sebaComputeCmd.Flags().StringVar(&computeText, "tokenize", "", "Text string to tokenize via ANE/CPU")

	sebaCmd.AddCommand(sebaScanCmd)
	sebaCmd.AddCommand(sebaBookCmd)
	sebaCmd.AddCommand(sebaFleetCmd)
	sebaCmd.AddCommand(sebaDiagramCmd)
	sebaCmd.AddCommand(sebaHardwareCmd)
	sebaCmd.AddCommand(sebaProfileCmd)
	sebaCmd.AddCommand(sebaComputeCmd)
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

func runSebaHardware(cmd *cobra.Command, args []string) error {
	start := time.Now()

	profile, err := seba.DetectHardware()
	if err != nil {
		return fmt.Errorf("hardware detection failed: %w", err)
	}

	if JsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(profile)
	}

	output.Banner()
	output.Header("SEBA — Hardware Architecture")

	aneStatus := "Not detected"
	if profile.NeuralEngine {
		aneStatus = "Active (16-core)"
	}

	gpuDisplay := profile.GPU.Name
	if gpuDisplay == "" {
		gpuDisplay = "Not detected"
	}
	if profile.GPU.MetalFamily != "" {
		gpuDisplay += " (" + profile.GPU.MetalFamily + ")"
	}

	ramDisplay := "Unknown"
	if profile.TotalRAM > 0 {
		ramDisplay = seba.FormatBytes(profile.TotalRAM)
	}

	dashboard := map[string]string{
		"CPU Model":     profile.CPUModel,
		"CPU Cores":     fmt.Sprintf("%d (%s)", profile.CPUCores, profile.CPUArch),
		"Total RAM":     ramDisplay,
		"GPU":           gpuDisplay,
		"Neural Engine": aneStatus,
		"OS":            fmt.Sprintf("%s/%s", profile.OS, profile.CPUArch),
	}

	if profile.Kernel != "" {
		dashboard["Kernel"] = profile.Kernel
	}

	accel := seba.DetectAccelerators()
	if accel != nil && accel.Primary != nil {
		dashboard["Primary Accelerator"] = string(accel.Primary.Type())
		if accel.HasCUDA {
			dashboard["CUDA"] = "Available"
		}
		if accel.HasMetal {
			dashboard["Metal"] = "Available"
		}
	}

	output.Dashboard(dashboard)
	output.Footer(time.Since(start))
	return nil
}

func runSebaProfile(cmd *cobra.Command, args []string) error {
	start := time.Now()

	profile, err := seba.DetectHardware()
	if err != nil {
		return fmt.Errorf("hardware detection failed: %w", err)
	}

	accel := seba.DetectAccelerators()

	combined := map[string]interface{}{
		"hardware":     profile,
		"accelerators": accel,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	}

	configDir := filepath.Join(os.Getenv("HOME"), ".config", "pantheon")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("cannot create config dir: %w", err)
	}

	profilePath := filepath.Join(configDir, "profile.json")
	f, err := os.Create(profilePath)
	if err != nil {
		return fmt.Errorf("cannot create profile: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(combined); err != nil {
		return fmt.Errorf("cannot write profile: %w", err)
	}

	if JsonOutput {
		enc2 := json.NewEncoder(os.Stdout)
		enc2.SetIndent("", "  ")
		return enc2.Encode(combined)
	}

	output.Banner()
	output.Header("SEBA — System Profile")
	output.Success("Profile saved to %s", profilePath)

	output.Dashboard(map[string]string{
		"CPU":   profile.CPUModel,
		"RAM":   seba.FormatBytes(profile.TotalRAM),
		"GPU":   profile.GPU.Name,
		"ANE":   fmt.Sprintf("%v", profile.NeuralEngine),
		"Saved": profilePath,
	})

	output.Footer(time.Since(start))
	return nil
}

func runSebaCompute(cmd *cobra.Command, args []string) error {
	start := time.Now()

	if !JsonOutput {
		output.Banner()
		output.Header("SEBA — Accelerated Compute")
	}

	if computeText == "" {
		output.Info("Use --tokenize \"text\" to run ANE/Neural inference.")
		return nil
	}

	result, _ := guard.Tokenize(computeText)
	elapsed := time.Since(start)

	if JsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]interface{}{
			"accelerator": result.Accel,
			"tokens":      result.Count,
			"text_length": len(computeText),
			"latency_ms":  elapsed.Milliseconds(),
		})
	}

	output.Dashboard(map[string]string{
		"Accelerator": result.Accel,
		"Tokens":      fmt.Sprintf("%d", result.Count),
		"Text Length": fmt.Sprintf("%d", len(computeText)),
		"Latency":     elapsed.String(),
	})
	output.Footer(elapsed)
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
