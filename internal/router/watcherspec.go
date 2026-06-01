package router

import "fmt"

// WatcherSpec is the router's prescription for the single liveness/wake
// mechanism a registered thread must run (ADR-024 Decision 2). `register`
// returns it instead of guessing and spawning a watcher itself; the surface
// arms exactly the mechanism named here. This table IS the R4 capability
// inventory in code — consulted on every register so no session rebuilds a
// watcher that already has a canonical form.
type WatcherSpec struct {
	Type               string `json:"type"`                 // canonical watcher kind, e.g. "loop-monitor"
	Mechanism          string `json:"mechanism"`            // human description of the mechanism
	ArmInstruction     string `json:"arm_instruction"`      // exact instruction the surface follows to arm
	HeartbeatIntervalS int    `json:"heartbeat_interval_s"` // bounded heartbeat cadence
	WatchesInbox       bool   `json:"watches_inbox"`        // does this mechanism watch items/ for the agent?
	Resident           bool   `json:"resident"`             // resident UI surface (heartbeat-only allowed)
}

// surfaceClass groups surfaces that share a canonical watcher.
const (
	surfaceClaude    = "claude"
	surfaceCodex     = "codex"
	surfaceMCP       = "mcp"
	surfaceAPI       = "api"
	surfaceWebhook   = "webhook"
	surfaceWorker    = "worker"
	surfaceMenubar   = "menubar"
	surfaceTUI       = "tui"
	surfaceVSCode    = "vscode"
	surfaceJetBrains = "jetbrains"
	surfaceCursor    = "cursor"
	surfaceMacApp    = "macapp"
	surfaceGemini    = "gemini"
	surfaceGemma     = "gemma"
	surfaceQwen      = "qwen"
)

// loopArmInstruction renders the claude `/loop` arm instruction. The
// idempotency signature MUST key on the thread_id (`pgrep -f thr-<thread_id>`),
// never the shared loop body / `DIR=` string (it collides with other agents'
// loops on a shared host) and never TaskList (it falsely reports empty). See
// ADR-024 §3 + F2, and claude-deck's correction (commit 838ad66).
func loopArmInstruction(agentID, threadID string) string {
	return fmt.Sprintf(
		"Arm /loop watching items/ for `to: %s`; emit `sirsi thread heartbeat --thread %s` each tick. "+
			"Re-assert idempotently on every SessionStart/wakeup, keyed on the thread_id "+
			"(`pgrep -f %s`) — NOT the shared loop body / `DIR=.agents/idea-router/items` string "+
			"(it matches other agents' loops on a shared host), NOT TaskList (it falsely reports empty). "+
			"Re-arm only when zero matching watcher processes exist for this thread.",
		agentID, threadID, threadID,
	)
}

// WatcherFor returns the canonical watcher spec for a surface, templated for the
// given agent and thread. Unknown surfaces fall back to the daemon mechanism
// (safe: a polling daemon over items/ works for any headless surface).
func WatcherFor(surface, agentID, threadID string) WatcherSpec {
	switch surface {
	case surfaceClaude:
		return WatcherSpec{
			Type:               "loop-monitor",
			Mechanism:          "/loop + Monitor on .agents/idea-router/items/",
			ArmInstruction:     loopArmInstruction(agentID, threadID),
			HeartbeatIntervalS: 60,
			WatchesInbox:       true,
			Resident:           false,
		}
	case surfaceCodex:
		return WatcherSpec{
			Type:               "app-heartbeat",
			Mechanism:          "codex app heartbeat (ctr-thread-wake polling items/)",
			ArmInstruction:     "Use the codex app heartbeat automation (ctr-thread-wake); it polls items/ for `to: " + agentID + "` and heartbeats natively. No manual loop to arm.",
			HeartbeatIntervalS: 60,
			WatchesInbox:       true,
			Resident:           false,
		}
	case surfaceGemini, surfaceGemma, surfaceQwen:
		return WatcherSpec{
			Type:               "surface-loop",
			Mechanism:          "surface-native loop, else `sirsi router daemon`",
			ArmInstruction:     "Run a surface-native watch loop over items/ for `to: " + agentID + "`, heartbeating each tick; if none exists, run `sirsi router daemon`.",
			HeartbeatIntervalS: 60,
			WatchesInbox:       true,
			Resident:           false,
		}
	case surfaceMenubar, surfaceTUI, surfaceVSCode, surfaceJetBrains, surfaceCursor, surfaceMacApp:
		return WatcherSpec{
			Type:               "native-runloop",
			Mechanism:          "native runloop heartbeat ping (resident surface)",
			ArmInstruction:     "Heartbeat from the native runloop on a bounded interval (>=60s); do NOT spawn an inbox poller unless this surface acts on items. Close on graceful shutdown; hard kill falls back to OS-truth reaping (ADR-022).",
			HeartbeatIntervalS: 60,
			WatchesInbox:       false,
			Resident:           true,
		}
	case surfaceMCP, surfaceAPI, surfaceWebhook, surfaceWorker:
		return WatcherSpec{
			Type:               "daemon",
			Mechanism:          "`sirsi router daemon` (or resident launch agent)",
			ArmInstruction:     "Run `sirsi router daemon` (or the resident launch agent) to poll items/ for `to: " + agentID + "` and dispatch.",
			HeartbeatIntervalS: 60,
			WatchesInbox:       true,
			Resident:           false,
		}
	default:
		return WatcherSpec{
			Type:               "daemon",
			Mechanism:          "`sirsi router daemon` (fallback for unrecognized surface)",
			ArmInstruction:     "Unrecognized surface; run `sirsi router daemon` to poll items/ for `to: " + agentID + "`.",
			HeartbeatIntervalS: 60,
			WatchesInbox:       true,
			Resident:           false,
		}
	}
}
