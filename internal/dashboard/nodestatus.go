// Package dashboard — nodestatus.go
//
// ADR-026 Horus ops-dashboard read endpoint. Serves the typed router.NodeStatus
// read-model at GET /api/node-status, plus a bounded OpsSummary projection at
// ?view=summary for the menubar (NSMenu budget — top-N + "N more" overflow).
//
// Boundary: claude-home defines this read contract; claude-pantheon owns
// surface chrome (menubar rows, TUI pane) that decodes into these types.
// Surfaces never re-aggregate — they consume one read-model, N read-only
// projections.
package dashboard

import (
	"net/http"
	"sort"

	"github.com/SirsiMaster/sirsi-pantheon/internal/router"
)

// DefaultOpsSummaryMax bounds the OpsSummary agent list to fit a typical
// NSMenu (~12-20 rows comfortably; we pick 12 to leave headroom for fixed rows
// like counts and the "Open full dashboard" link). The remainder collapse into
// `more_agents` so the menubar can render an "N more…" overflow row.
const DefaultOpsSummaryMax = 12

// OpsSummary is the bounded menubar projection of router.NodeStatus.
// It is a pure reduction — every field is derived from the source NodeStatus
// in dashboardSummarize(); no field is sourced independently. This keeps
// "one read-model" honest even when the menubar's NSMenu can't show the full
// thread list (claude-pantheon constraint #2 from boundary ack 235652).
type OpsSummary struct {
	SchemaVersion string `json:"schema_version"` // mirrors router.NodeStatus

	// Roll-ups (reduce loops, not separate sources)
	LiveThreadCount    int `json:"live_thread_count"`
	StaleThreadCount   int `json:"stale_thread_count"`
	SuspendedThreads   int `json:"suspended_threads,omitempty"` // populated when ADR-025 records visible
	QueueOpenItems     int `json:"queue_open_items"`            // = TotalPending
	RecentFailureCount int `json:"recent_failure_count"`

	// Drift / health flags — true if any agent CLI has auth issues, daemon misconfig,
	// or the binary drift detector (ADR-023) found a stale sibling.
	HasDriftOrAuthIssue bool   `json:"has_drift_or_auth_issue"`
	WorstIcon           string `json:"worst_icon,omitempty"` // glyph hint for the menubar's lead row

	// Bounded agent list — top-N agents by pending+live signal, others collapsed.
	Agents     []AgentSummary `json:"agents"`
	MoreAgents int            `json:"more_agents,omitempty"` // count of agents NOT shown

	// Echo source identifiers so the menubar can deep-link.
	RouterHome string `json:"router_home"`
}

// AgentSummary is a single menubar row.
type AgentSummary struct {
	AgentID      string `json:"agent_id"`
	LiveThreads  int    `json:"live_threads"`
	StaleThreads int    `json:"stale_threads"`
	PendingItems int    `json:"pending_items"`
	NeedsLogin   bool   `json:"needs_login,omitempty"`
}

// dashboardSummarize is the pure reduction NodeStatus → OpsSummary.
// Bounded by `max` (the agent list is truncated; remainder counted in
// MoreAgents). Stable, ordered output for diff-rendering by the menubar.
func dashboardSummarize(ns *router.NodeStatus, max int) OpsSummary {
	if max <= 0 {
		max = DefaultOpsSummaryMax
	}
	sum := OpsSummary{
		SchemaVersion:      ns.SchemaVersion,
		LiveThreadCount:    ns.LiveThreadCount,
		StaleThreadCount:   len(ns.StaleThreads),
		QueueOpenItems:     ns.TotalPending,
		RecentFailureCount: len(ns.RecentFailures),
		RouterHome:         ns.RouterHome,
	}

	// SuspendedThreads is the count of LiveThreads + StaleThreads whose Status
	// is the ADR-025 "suspended" lifecycle state. CollectNodeStatus drops
	// terminal statuses (closed/reaped) before populating these slices, so a
	// suspended record can land in either bucket depending on its idleness.
	for _, t := range ns.LiveThreads {
		if t.Status == router.ThreadStatusSuspended {
			sum.SuspendedThreads++
		}
	}
	for _, t := range ns.StaleThreads {
		if t.Status == router.ThreadStatusSuspended {
			sum.SuspendedThreads++
		}
	}

	// Drift / auth — true if any agent CLI has auth failures or any wake
	// mechanism is not ready, or daemon is installed but its configured binary
	// does not exist (ADR-023 drift class).
	for _, h := range ns.AgentHealth {
		if h.CLIFound && !h.AuthOK {
			sum.HasDriftOrAuthIssue = true
			break
		}
	}
	if !sum.HasDriftOrAuthIssue {
		for _, w := range ns.WakeHealth {
			if !w.Ready {
				sum.HasDriftOrAuthIssue = true
				break
			}
		}
	}
	if !sum.HasDriftOrAuthIssue && ns.DaemonInstalled && ns.ConfiguredBinary != "" && !ns.BinaryExists {
		sum.HasDriftOrAuthIssue = true
	}
	if sum.HasDriftOrAuthIssue {
		sum.WorstIcon = "🔴"
	} else if sum.StaleThreadCount > 0 || sum.RecentFailureCount > 0 {
		sum.WorstIcon = "🟡"
	} else {
		sum.WorstIcon = "🟢"
	}

	// Per-agent roll-up. Sources: ns.RegisteredAgents (membership),
	// ns.PendingByAgent (queue), ns.LiveThreads/StaleThreads (presence),
	// ns.AgentHealth (auth).
	type agg struct {
		live, stale, pending int
		needsLogin           bool
	}
	by := map[string]*agg{}
	get := func(id string) *agg {
		a, ok := by[id]
		if !ok {
			a = &agg{}
			by[id] = a
		}
		return a
	}
	for _, id := range ns.RegisteredAgents {
		_ = get(id) // ensure registered agents appear even with zero signal
	}
	for _, t := range ns.LiveThreads {
		get(t.AgentID).live++
	}
	for _, t := range ns.StaleThreads {
		get(t.AgentID).stale++
	}
	for agentID, ids := range ns.PendingByAgent {
		get(agentID).pending += len(ids)
	}
	for _, h := range ns.AgentHealth {
		// AgentHealth keys on AgentType ("claude"/"codex"), not agent_id; mark every
		// agent_id whose name contains the failing type as needs_login. This mirrors
		// CollectNodeStatus's own BlockedItems computation.
		if h.CLIFound && h.NeedsLogin {
			for id := range by {
				if containsType(id, h.AgentType) {
					by[id].needsLogin = true
				}
			}
		}
	}

	// Rank: pending desc, then live desc, then alpha for determinism.
	ids := make([]string, 0, len(by))
	for id := range by {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		ai, aj := by[ids[i]], by[ids[j]]
		if ai.pending != aj.pending {
			return ai.pending > aj.pending
		}
		if ai.live != aj.live {
			return ai.live > aj.live
		}
		return ids[i] < ids[j]
	})

	if len(ids) > max {
		sum.MoreAgents = len(ids) - max
		ids = ids[:max]
	}
	sum.Agents = make([]AgentSummary, 0, len(ids))
	for _, id := range ids {
		a := by[id]
		sum.Agents = append(sum.Agents, AgentSummary{
			AgentID:      id,
			LiveThreads:  a.live,
			StaleThreads: a.stale,
			PendingItems: a.pending,
			NeedsLogin:   a.needsLogin,
		})
	}
	return sum
}

// containsType reports whether an agent id contains a CLI type token
// ("claude" or "codex"). Exists so dashboardSummarize doesn't pull in strings.
func containsType(agentID, cliType string) bool {
	// agent ids are "claude-pantheon", "codex-finalwishes", etc.
	if len(cliType) == 0 || len(agentID) < len(cliType) {
		return false
	}
	for i := 0; i+len(cliType) <= len(agentID); i++ {
		if agentID[i:i+len(cliType)] == cliType {
			return true
		}
	}
	return false
}

// NodeStatusCollector is the producer hook for the GET /api/node-status
// endpoint. Tests inject a deterministic *router.NodeStatus without touching
// the host. Production wiring (cmd/sirsi-menubar) passes
// router.CollectNodeStatus's caller via Config.NodeStatusFn.
type NodeStatusCollector func() (*router.NodeStatus, error)

// apiNodeStatus serves GET /api/node-status (ADR-026).
//   - default: full router.NodeStatus
//   - ?view=summary: bounded OpsSummary (top-N agents + "more_agents")
//
// Read-only: no method-gating, no ConfirmGuard path, no side effects.
func (s *Server) apiNodeStatus(w http.ResponseWriter, r *http.Request) {
	if s.cfg.NodeStatusFn == nil {
		writeError(w, "node-status not available (collector not wired)", http.StatusServiceUnavailable)
		return
	}
	ns, err := s.cfg.NodeStatusFn()
	if err != nil {
		writeError(w, "node-status collection failed", http.StatusInternalServerError)
		return
	}
	if ns == nil {
		writeError(w, "node-status returned nil", http.StatusInternalServerError)
		return
	}
	if r.URL.Query().Get("view") == "summary" {
		writeJSON(w, dashboardSummarize(ns, DefaultOpsSummaryMax))
		return
	}
	writeJSON(w, ns)
}
