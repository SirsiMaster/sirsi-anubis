package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/SirsiMaster/sirsi-pantheon/internal/output"
	"github.com/SirsiMaster/sirsi-pantheon/internal/scarab"
)

var (
	scarabContainers bool
	scarabConfirmNet bool
)

var scarabCmd = &cobra.Command{
	Use:   "scarab",
	Short: "🪲 Discover and audit network hosts and containers",
	Long: `🪲 Scarab — The Transformer

Named after the sacred Egyptian beetle that rolls across the landscape,
discovering and transforming everything it touches.

  pantheon scarab                    Discover hosts on local subnet
  pantheon scarab --containers       Audit Docker containers
  pantheon scarab --confirm-network  Required for active network scanning

Network discovery uses ARP cache (passive) and ping sweep (active).
Container audit scans Docker for stopped containers, dangling images,
and unused volumes.`,
	Run: runScarab,
}

func init() {
	scarabCmd.Flags().BoolVar(&scarabContainers, "containers", false, "Audit Docker containers only")
	scarabCmd.Flags().BoolVar(&scarabConfirmNet, "confirm-network", false, "Confirm active network scanning (required for ping sweep)")
}

func runScarab(cmd *cobra.Command, args []string) {
	if scarabContainers {
		runScarabContainers()
		return
	}
	runScarabDiscover()
}

func runScarabDiscover() {
	output.Header("🪲 Scarab — Network Discovery")
	fmt.Println()

	if !scarabConfirmNet {
		output.Warn("⚠️  Active network scanning requires --confirm-network flag")
		output.Info("   This sends ICMP pings across your local subnet.")
		output.Info("   Use: pantheon scarab --confirm-network")
		fmt.Println()
		output.Info("Showing passive ARP cache only...")
		fmt.Println()
	}

	result, err := scarab.Discover()
	if err != nil {
		output.Error(fmt.Sprintf("Discovery failed: %v", err))
		os.Exit(1)
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(result)
		return
	}

	output.Info(fmt.Sprintf("Subnet:      %s", result.Subnet))
	output.Info(fmt.Sprintf("Hosts found: %d", len(result.Hosts)))
	output.Info(fmt.Sprintf("Alive:       %d", result.TotalAlive))
	output.Info(fmt.Sprintf("Scan time:   %s", result.Duration.Round(1e6)))
	fmt.Println()

	if len(result.Hosts) == 0 {
		output.Info("No hosts discovered")
		return
	}

	// Table header
	fmt.Printf("    %-16s  %-18s  %-6s  %s\n", "IP", "MAC", "Alive", "Hostname")
	fmt.Printf("    %-16s  %-18s  %-6s  %s\n", "──────────────", "─────────────────", "─────", "────────")

	for _, h := range result.Hosts {
		alive := "  ✅"
		if !h.Alive {
			alive = "  ❓"
		}
		mac := h.MAC
		if mac == "" {
			mac = "—"
		}
		hostname := h.Hostname
		if hostname == "" {
			hostname = "—"
		}
		fmt.Printf("    %-16s  %-18s  %-6s  %s\n", h.IP, mac, alive, hostname)
	}
	fmt.Println()

	output.Info("💡 Eye of Horus upgrade: sweep entire fleet with agent deployment")
	output.Info("   → sirsi.dev/eye-of-horus")
}

func runScarabContainers() {
	output.Header("🪲 Scarab — Container Audit")
	fmt.Println()

	audit, err := scarab.AuditContainers()
	if err != nil {
		output.Error(fmt.Sprintf("Container audit failed: %v", err))
		os.Exit(1)
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(audit)
		return
	}

	if !audit.DockerRunning {
		output.Info("🐳 Docker is not running")
		return
	}

	output.Info("🐳 Docker:          Running")
	output.Info(fmt.Sprintf("   Containers:      %d total", len(audit.Containers)))
	output.Info(fmt.Sprintf("   Running:         %d", audit.RunningCount))
	output.Info(fmt.Sprintf("   Stopped:         %d", audit.StoppedCount))
	output.Info(fmt.Sprintf("   Dangling images: %d", audit.DanglingImages))
	output.Info(fmt.Sprintf("   Unused volumes:  %d", audit.UnusedVolumes))
	fmt.Println()

	if len(audit.Containers) == 0 {
		output.Info("No containers found")
		return
	}

	for _, c := range audit.Containers {
		status := scarab.FormatContainerStatus(c)
		fmt.Printf("    %-25s  %-30s  %s\n", c.Name, c.Image, status)
	}

	// Warnings
	fmt.Println()
	if audit.StoppedCount > 0 {
		output.Warn(fmt.Sprintf("⚠️  %d stopped containers — consider: docker container prune", audit.StoppedCount))
	}
	if audit.DanglingImages > 0 {
		output.Warn(fmt.Sprintf("⚠️  %d dangling images — consider: docker image prune", audit.DanglingImages))
	}
	if audit.UnusedVolumes > 0 {
		output.Warn(fmt.Sprintf("⚠️  %d unused volumes — consider: docker volume prune", audit.UnusedVolumes))
	}
}
