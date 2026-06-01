package dashboard

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// confirmTokenTTL bounds how long a prepared destructive action stays
// committable. Short enough that a stale token can't be replayed long after the
// operator saw the preview; long enough for a human to read and click.
const confirmTokenTTL = 2 * time.Minute

// ConfirmGuard implements server-enforced, single-use, tokenized two-phase
// confirmation for destructive dashboard actions (E2, codex freeze gate).
//
// It is the ONLY safety boundary for destructive HTTP actions. The client is
// never trusted to enforce confirmation — there is no reliance on a browser
// confirm() dialog or a client-set flag. The flow:
//
//	Prepare(action,target,params,...) -> PreparedAction{ConfirmToken, ActionHash, ExpiresAt}
//	Validate(token, action, target, params, echoedHash) -> nil iff committable
//
// Validate rejects a token that is missing, unknown, already-used, expired, or
// bound to a different action/target/params. On success the token is consumed
// (single-use) so it can never be replayed.
type ConfirmGuard struct {
	mu      sync.Mutex
	pending map[string]*pendingConfirm
	ttl     time.Duration
	now     func() time.Time          // injectable clock (Rule A16) — instance state, set once
	randFn  func([]byte) (int, error) // injectable entropy (Rule A16) for failure-path testing
}

type pendingConfirm struct {
	action    string
	target    string
	hash      string
	expiresAt time.Time
	consumed  bool
}

// NewConfirmGuard creates a confirm guard with the default 2-minute token TTL.
func NewConfirmGuard() *ConfirmGuard {
	return &ConfirmGuard{
		pending: make(map[string]*pendingConfirm),
		ttl:     confirmTokenTTL,
		now:     time.Now,
		randFn:  rand.Read,
	}
}

// ActionHash returns a stable fingerprint of an action over
// (action, target, sorted params). Deterministic: identical inputs always yield
// an identical hash, so a token bound to one action can never authorize another.
func ActionHash(action, target string, params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString(action)
	b.WriteByte(0)
	b.WriteString(target)
	for _, k := range keys {
		b.WriteByte(0)
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(params[k])
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

// Prepare records a pending destructive action and returns a single-use token.
// The caller supplies the dry-run preview (computed read-only, no side effects).
func (g *ConfirmGuard) Prepare(action, target string, params map[string]string, preview string, affected []string, impact string) (*PreparedAction, error) {
	if action == "" {
		return nil, fmt.Errorf("confirm: action required")
	}

	tokBytes := make([]byte, 16)
	if _, err := g.randFn(tokBytes); err != nil {
		return nil, fmt.Errorf("confirm: token generation failed: %w", err)
	}
	token := hex.EncodeToString(tokBytes)
	hash := ActionHash(action, target, params)
	now := g.now()
	exp := now.Add(g.ttl)

	g.mu.Lock()
	g.gcLocked(now)
	g.pending[token] = &pendingConfirm{
		action:    action,
		target:    target,
		hash:      hash,
		expiresAt: exp,
	}
	g.mu.Unlock()

	if affected == nil {
		affected = []string{}
	}
	return &PreparedAction{
		Action:            action,
		Target:            target,
		DryRun:            true,
		ConfirmToken:      token,
		ActionHash:        hash,
		ExpiresAt:         exp,
		Preview:           preview,
		AffectedResources: affected,
		EstimatedImpact:   impact,
	}, nil
}

// Validate consumes a confirm token, enforcing every rejection rule. It returns
// nil ONLY when the token exists, is unexpired, unused, and binds to exactly
// this action/target/params (and the echoed hash matches, when provided). On
// success the token is consumed and removed (single-use, no replay).
func (g *ConfirmGuard) Validate(token, action, target string, params map[string]string, echoedHash string) error {
	if token == "" {
		return fmt.Errorf("confirm: missing confirm_token")
	}
	hash := ActionHash(action, target, params)
	now := g.now()

	g.mu.Lock()
	defer g.mu.Unlock()

	pc, ok := g.pending[token]
	if !ok {
		return fmt.Errorf("confirm: unknown or already-used token")
	}
	if pc.consumed {
		return fmt.Errorf("confirm: token already used")
	}
	if now.After(pc.expiresAt) {
		delete(g.pending, token)
		return fmt.Errorf("confirm: token expired")
	}
	if pc.action != action || pc.target != target || pc.hash != hash {
		return fmt.Errorf("confirm: token does not match action/target")
	}
	if echoedHash != "" && echoedHash != pc.hash {
		return fmt.Errorf("confirm: action_hash mismatch")
	}

	pc.consumed = true
	delete(g.pending, token) // single-use: remove on consume
	return nil
}

// requireConfirm enforces the E2 confirm contract for a destructive HTTP
// handler. It returns true ONLY when the caller may execute for real (a valid
// confirm token was presented). Otherwise it writes the response itself — a
// PreparedAction (prepare/dry-run) when no token was supplied, or an error when
// the token is invalid — and returns false, so the caller simply returns.
//
// This is the single shared destructive-confirm contract codex required: every
// destructive endpoint funnels through it, so "omitted dry_run" can never
// destroy and no endpoint invents its own confirmation semantics.
func (s *Server) requireConfirm(w http.ResponseWriter, action, target string, params map[string]string, token, echoedHash, preview string, affected []string, impact string) bool {
	if s.confirm == nil {
		writeError(w, "confirm guard not available", http.StatusServiceUnavailable)
		return false
	}
	if token == "" {
		prep, err := s.confirm.Prepare(action, target, params, preview, affected, impact)
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return false
		}
		writeJSON(w, prep)
		return false
	}
	if err := s.confirm.Validate(token, action, target, params, echoedHash); err != nil {
		writeError(w, err.Error(), http.StatusForbidden)
		return false
	}
	return true
}

// gcLocked drops expired or consumed tokens. Caller must hold g.mu.
func (g *ConfirmGuard) gcLocked(now time.Time) {
	for tok, pc := range g.pending {
		if pc.consumed || now.After(pc.expiresAt) {
			delete(g.pending, tok)
		}
	}
}
