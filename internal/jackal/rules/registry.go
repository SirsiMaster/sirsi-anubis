// Package rules provides all built-in scan rules for the Jackal engine.
//
// Rule registration: All rules are registered via AllRules().
// New rules should be added there.
package rules

import (
	"runtime"

	"github.com/SirsiMaster/sirsi-anubis/internal/jackal"
)

// AllRules returns all built-in scan rules for the current platform.
// New rules are added here as they're implemented.
func AllRules() []jackal.ScanRule {
	var rules []jackal.ScanRule

	// Platform-specific rules
	switch runtime.GOOS {
	case "darwin":
		rules = append(rules, darwinRules()...)
	case "linux":
		rules = append(rules, linuxRules()...)
	}

	// Cross-platform rules
	rules = append(rules, crossPlatformRules()...)

	return rules
}

// darwinRules returns macOS-specific scan rules.
func darwinRules() []jackal.ScanRule {
	return []jackal.ScanRule{
		// General Mac
		NewSystemCachesRule(),
		NewSystemLogsRule(),
		NewCrashReportsRule(),
		NewDownloadsJunkRule(),
		NewTrashRule(),
		NewBrowserCachesRule(),

		// IDEs (macOS-specific)
		NewXcodeDerivedDataRule(),
	}
}

// linuxRules returns Linux-specific scan rules.
func linuxRules() []jackal.ScanRule {
	// TODO: Phase 2
	return []jackal.ScanRule{}
}

// crossPlatformRules returns rules that work on all platforms.
func crossPlatformRules() []jackal.ScanRule {
	return []jackal.ScanRule{
		// Developer Frameworks
		NewNodeModulesRule(),
		NewGoModCacheRule(),
		NewPythonCachesRule(),
		NewRustTargetRule(),
		NewDockerRule(),
	}
}
