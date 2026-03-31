package neith

import (
	"fmt"
	"strings"
	"time"
)

// Weave represents the unified development plan and timeline of the Pantheon.
// Owned by Net (Neith), the Weaver of existence.
type Weave struct {
	SessionID    string    `json:"session_id"`
	StartedAt    time.Time `json:"started_at"`
	Plan         []string  `json:"plan"`
	Achievements []string  `json:"achievements"`
	DriftFound   bool      `json:"drift_found"`
}

// AssessLogs compares the active project logs (BUILD_LOG.md) against the Development Plan.
// Each plan item is split into keywords; the score reflects how many keywords appear in the log.
func (w *Weave) AssessLogs(logContent string) (float64, error) {
	if len(w.Plan) == 0 {
		return 1.0, nil
	}

	logLower := strings.ToLower(logContent)
	var totalScore float64

	for _, item := range w.Plan {
		words := strings.Fields(strings.ToLower(item))
		if len(words) == 0 {
			totalScore += 1.0
			continue
		}
		matched := 0
		for _, word := range words {
			if strings.Contains(logLower, word) {
				matched++
			}
		}
		totalScore += float64(matched) / float64(len(words))
	}

	return totalScore / float64(len(w.Plan)), nil
}

// CheckDrift compares Plan against Achievements and sets DriftFound if any plan item is unachieved.
func (w *Weave) CheckDrift() {
	if len(w.Plan) == 0 {
		w.DriftFound = false
		return
	}
	achieved := make(map[string]bool, len(w.Achievements))
	for _, a := range w.Achievements {
		achieved[strings.ToLower(a)] = true
	}
	for _, p := range w.Plan {
		if !achieved[strings.ToLower(p)] {
			w.DriftFound = true
			return
		}
	}
	w.DriftFound = false
}

// Tapestry represents the interconnected state of all Pantheon deities.
type Tapestry struct {
	MaatConsistent  bool
	AnubisCorrect   bool
	KaExtinguished  bool
	ThothAccurate   bool
	SekhmetHardened bool
}

// Align ensures all deities submit to the tapestry. Checks are ordered by severity.
func (t *Tapestry) Align() error {
	if !t.MaatConsistent {
		return fmt.Errorf("the weave is unbalanced: Ma'at detects weight of untruth")
	}
	if !t.AnubisCorrect {
		return fmt.Errorf("the weave is torn: Anubis finds corruption in the scales")
	}
	if !t.KaExtinguished {
		return fmt.Errorf("the weave is haunted: Ka still lingers — ghosts remain")
	}
	if !t.ThothAccurate {
		return fmt.Errorf("the weave is unscribed: Thoth's records are incomplete")
	}
	if !t.SekhmetHardened {
		return fmt.Errorf("the weave is vulnerable: Sekhmet has not hardened the perimeter")
	}
	return nil
}
