package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/SirsiMaster/sirsi-pantheon/internal/output"
	"github.com/spf13/cobra"
)

var permissionsCmd = &cobra.Command{
	Use:   "permissions",
	Short: "Grant Full Disk Access for comprehensive scanning",
	Long: `Pantheon needs Full Disk Access to scan all directories on your machine.

Without it, macOS will repeatedly ask for permission to access Desktop,
Documents, Downloads, iCloud, and app containers.

This command opens System Settings to the correct pane so you can add
the sirsi binary once. After granting, all future scans work without prompts.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if runtime.GOOS != "darwin" {
			fmt.Println("Full Disk Access is a macOS feature. No action needed on this platform.")
			return nil
		}

		output.Header("Permissions Setup")
		fmt.Println()

		// Find the running binary path
		exe, err := os.Executable()
		if err != nil {
			exe = "sirsi"
		}
		exe, _ = filepath.EvalSymlinks(exe)

		// Check if we're running from Homebrew or local build
		fmt.Println("  To stop repeated permission prompts, grant Full Disk Access to:")
		fmt.Println()
		fmt.Printf("    %s\n", exe)
		fmt.Println()
		fmt.Println("  Steps:")
		fmt.Println("    1. System Settings will open to Privacy & Security")
		fmt.Println("    2. Click 'Full Disk Access' in the sidebar")
		fmt.Println("    3. Click the + button")
		fmt.Println("    4. Navigate to the path above and add it")
		fmt.Println("    5. Restart your terminal")
		fmt.Println()

		// Also mention the terminal app itself
		terminal := os.Getenv("TERM_PROGRAM")
		if terminal != "" {
			fmt.Printf("  Tip: Also grant Full Disk Access to %s (your terminal app)\n", terminal)
			fmt.Printf("  for complete access without any prompts.\n\n")
		}

		// Open System Settings
		fmt.Println("  Opening System Settings...")
		openCmd := exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_AllFiles")
		if err := openCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "  Could not open System Settings: %v\n", err)
			fmt.Println("  Open manually: System Settings → Privacy & Security → Full Disk Access")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(permissionsCmd)
}
