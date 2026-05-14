// Package router implements the idea-router filesystem protocol for
// cross-agent collaboration. Codex and Claude exchange proposals,
// reviews, and decisions through .agents/idea-router/ in the repo.
//
// When an agent submits a document, the router can optionally notify
// the other agent by spawning a CLI session (Codex or Claude Code).
package router

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// DocType identifies the kind of router document.
type DocType string

const (
	DocProposal DocType = "proposal"
	DocReview   DocType = "review"
	DocDecision DocType = "decision"
)

// Document represents a single router file.
type Document struct {
	ID       string    // filename without extension
	Type     DocType   // proposal, review, decision
	Path     string    // full filesystem path
	Author   string    // codex or claude
	Title    string    // extracted from first heading
	ModTime  time.Time // file modification time
	Content  string    // full file content
}

// State represents the router state.json file.
type State struct {
	Version         int      `json:"version"`
	ActiveTopics    []string `json:"active_topics"`
	CompletedTopics []string `json:"completed_topics,omitempty"`
	LastCodexRead   string   `json:"last_codex_read"`
	LastClaudeRead  string   `json:"last_claude_read"`
	Rules           map[string]bool `json:"rules"`
}

// Router provides access to the idea-router filesystem.
type Router struct {
	root string // path to .agents/idea-router/
}

// New creates a Router rooted at the given repo path.
// It looks for .agents/idea-router/ under repoRoot.
func New(repoRoot string) (*Router, error) {
	root := filepath.Join(repoRoot, ".agents", "idea-router")
	if info, err := os.Stat(root); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("idea-router not found at %s", root)
	}
	return &Router{root: root}, nil
}

// FindRepoRoot walks up from cwd to find a .agents/idea-router/ directory.
func FindRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, ".agents", "idea-router")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no .agents/idea-router/ found in any parent directory")
		}
		dir = parent
	}
}

// ReadState returns the current router state.
func (r *Router) ReadState() (*State, error) {
	data, err := os.ReadFile(filepath.Join(r.root, "state.json"))
	if err != nil {
		return nil, fmt.Errorf("read state: %w", err)
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parse state: %w", err)
	}
	return &state, nil
}

// WriteState persists the router state.
func (r *Router) WriteState(state *State) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	return os.WriteFile(filepath.Join(r.root, "state.json"), data, 0o644)
}

// validAuthors is the whitelist of allowed author values.
var validAuthors = map[string]bool{
	"codex": true,
	"claude": true,
}

// ValidateAuthor checks that author is an allowed value.
// The whitelist is the sole defense against path traversal — only "codex"
// and "claude" are accepted, so no path separators or ".." can appear.
func ValidateAuthor(author string) error {
	if !validAuthors[author] {
		return fmt.Errorf("author %q is not allowed (must be 'codex' or 'claude')", author)
	}
	return nil
}

// Submit writes a new document to the router and updates the state.
// Returns the document ID (filename stem).
func (r *Router) Submit(docType DocType, author, title, content string) (string, error) {
	if err := ValidateAuthor(author); err != nil {
		return "", err
	}

	ts := time.Now().Format("20060102-1504")
	slug := slugify(title)
	id := fmt.Sprintf("%s-%s-%s", ts, author, slug)

	var dir string
	switch docType {
	case DocProposal:
		dir = "proposals"
	case DocReview:
		dir = "reviews"
	case DocDecision:
		dir = "decisions"
	default:
		return "", fmt.Errorf("unknown doc type: %s", docType)
	}

	path := filepath.Join(r.root, dir, id+".md")

	// Path containment: verify the resolved path stays under the expected directory
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	absDir, err := filepath.Abs(filepath.Join(r.root, dir))
	if err != nil {
		return "", fmt.Errorf("resolve dir: %w", err)
	}
	if !strings.HasPrefix(absPath, absDir+string(os.PathSeparator)) {
		return "", fmt.Errorf("path traversal blocked: %q escapes %q", absPath, absDir)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write %s: %w", path, err)
	}

	// Update state: mark last read time for the submitting agent
	state, err := r.ReadState()
	if err != nil {
		return id, nil // file written, state update failed — non-fatal
	}
	now := time.Now().Format(time.RFC3339)
	switch author {
	case "claude":
		state.LastClaudeRead = now
	case "codex":
		state.LastCodexRead = now
	}
	_ = r.WriteState(state)

	return id, nil
}

// List returns all documents across proposals, reviews, and decisions.
func (r *Router) List() ([]Document, error) {
	var docs []Document

	for _, dt := range []DocType{DocProposal, DocReview, DocDecision} {
		dir := filepath.Join(r.root, string(dt)+"s")
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			info, _ := e.Info()
			doc := Document{
				ID:      strings.TrimSuffix(e.Name(), ".md"),
				Type:    dt,
				Path:    filepath.Join(dir, e.Name()),
				ModTime: info.ModTime(),
			}
			// Extract title from first heading
			if data, err := os.ReadFile(doc.Path); err == nil {
				doc.Content = string(data)
				for _, line := range strings.Split(doc.Content, "\n") {
					if strings.HasPrefix(line, "# ") {
						doc.Title = strings.TrimPrefix(line, "# ")
						break
					}
				}
				// Extract author from content
				for _, line := range strings.Split(doc.Content, "\n") {
					trimmed := strings.TrimSpace(line)
					if strings.HasPrefix(trimmed, "author:") || strings.HasPrefix(trimmed, "reviewer:") {
						parts := strings.SplitN(trimmed, ":", 2)
						if len(parts) == 2 {
							doc.Author = strings.TrimSpace(parts[1])
						}
						break
					}
				}
			}
			docs = append(docs, doc)
		}
	}

	sort.Slice(docs, func(i, j int) bool {
		return docs[i].ModTime.After(docs[j].ModTime)
	})
	return docs, nil
}

// Get returns a single document by ID (searches all directories).
func (r *Router) Get(id string) (*Document, error) {
	docs, err := r.List()
	if err != nil {
		return nil, err
	}
	for _, d := range docs {
		if d.ID == id {
			return &d, nil
		}
	}
	return nil, fmt.Errorf("document %q not found", id)
}

// PollSince returns documents modified after the given timestamp.
func (r *Router) PollSince(since time.Time, limit int) ([]Document, error) {
	docs, err := r.List()
	if err != nil {
		return nil, err
	}
	var filtered []Document
	for _, d := range docs {
		if d.ModTime.After(since) {
			filtered = append(filtered, d)
			if limit > 0 && len(filtered) >= limit {
				break
			}
		}
	}
	return filtered, nil
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		if r == ' ' || r == '-' || r == '_' {
			return '-'
		}
		return -1
	}, s)
	// Collapse multiple dashes
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	if len(s) > 40 {
		s = s[:40]
	}
	return s
}
