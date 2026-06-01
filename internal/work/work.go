// Package work is a pull-model work queue between agent threads.
//
// One file per work item lives under <root>/items/, with YAML frontmatter
// recording from/to/status/opened and a free-form instructions/body section.
// Receivers poll their inbox on wake, do the work, and call Close. No daemons,
// no dispatch ledger, no launch agents — the file is the queue.
package work

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Item is one piece of work routed between two agent threads.
type Item struct {
	ID           string
	From         string
	To           string
	Title        string
	Status       string // "open" or "closed"
	Opened       string // RFC3339
	Closed       string // RFC3339, empty if open
	Instructions string
	Result       string
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "untitled"
	}
	if len(s) > 60 {
		s = s[:60]
	}
	return s
}

func itemsDir(root string) string { return filepath.Join(root, "items") }

// EnsureRoot creates the items/ directory if missing.
func EnsureRoot(root string) error {
	return os.MkdirAll(itemsDir(root), 0o755)
}

// quoteYAML wraps a value in double quotes and escapes embedded quotes,
// backslashes, and newlines so titles/agent ids containing YAML-sensitive
// characters (colons, leading -, &, *, !, |, etc.) round-trip cleanly.
func quoteYAML(v string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`)
	return `"` + r.Replace(v) + `"`
}

// unquoteYAML reverses quoteYAML. Values that aren't double-quoted pass through.
func unquoteYAML(v string) string {
	if len(v) < 2 || v[0] != '"' || v[len(v)-1] != '"' {
		return v
	}
	inner := v[1 : len(v)-1]
	r := strings.NewReplacer(`\"`, `"`, `\n`, "\n", `\r`, "\r", `\\`, `\`)
	return r.Replace(inner)
}

// Send writes a new open item from→to and returns its ID (filename stem).
func Send(root, from, to, title, instructions string) (string, error) {
	if from == "" || to == "" {
		return "", fmt.Errorf("from and to are required")
	}
	if err := EnsureRoot(root); err != nil {
		return "", err
	}
	now := time.Now().UTC()
	id := fmt.Sprintf("%s-%s-%s-%s", now.Format("20060102-150405"), slugify(from), slugify(to), slugify(title))
	path := filepath.Join(itemsDir(root), id+".md")
	body := fmt.Sprintf(`---
from: %s
to: %s
title: %s
status: open
opened: %s
---

## Instructions

%s
`, quoteYAML(from), quoteYAML(to), quoteYAML(title), now.Format(time.RFC3339), strings.TrimSpace(instructions))
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return "", err
	}
	return id, nil
}

// Get loads one item by ID.
func Get(root, id string) (Item, error) {
	path := filepath.Join(itemsDir(root), id+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		return Item{}, err
	}
	return parse(id, string(data))
}

// ListInbox returns open items addressed to the given agent, oldest first.
// If agent is empty, returns all open items.
func ListInbox(root, agent string) ([]Item, error) {
	entries, err := os.ReadDir(itemsDir(root))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var items []Item
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".md")
		it, err := Get(root, id)
		if err != nil {
			continue
		}
		if it.Status != "open" {
			continue
		}
		if agent != "" && it.To != agent {
			continue
		}
		items = append(items, it)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items, nil
}

// ListAll returns every item regardless of status.
func ListAll(root string) ([]Item, error) {
	entries, err := os.ReadDir(itemsDir(root))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var items []Item
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".md")
		it, err := Get(root, id)
		if err != nil {
			continue
		}
		items = append(items, it)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items, nil
}

// Close marks an item closed and appends a result section.
func Close(root, id, result string) error {
	it, err := Get(root, id)
	if err != nil {
		return err
	}
	if it.Status == "closed" {
		return fmt.Errorf("already closed")
	}
	path := filepath.Join(itemsDir(root), id+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	updated := strings.Replace(string(data), "status: open", "status: closed", 1)
	updated = strings.Replace(updated, "opened: "+it.Opened, "opened: "+it.Opened+"\nclosed: "+now, 1)
	if strings.TrimSpace(result) != "" {
		updated += fmt.Sprintf("\n## Result\n\n%s\n", strings.TrimSpace(result))
	} else {
		updated += "\n## Result\n\n(closed without result)\n"
	}
	return os.WriteFile(path, []byte(updated), 0o644)
}

// parse extracts an Item from frontmatter + body text.
func parse(id, content string) (Item, error) {
	it := Item{ID: id}
	if !strings.HasPrefix(content, "---\n") {
		return it, fmt.Errorf("missing frontmatter")
	}
	end := strings.Index(content[4:], "\n---\n")
	if end < 0 {
		return it, fmt.Errorf("unterminated frontmatter")
	}
	fm := content[4 : 4+end]
	body := content[4+end+5:]
	for _, line := range strings.Split(fm, "\n") {
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = unquoteYAML(strings.TrimSpace(v))
		switch k {
		case "from":
			it.From = v
		case "to":
			it.To = v
		case "title":
			it.Title = v
		case "status":
			it.Status = v
		case "opened":
			it.Opened = v
		case "closed":
			it.Closed = v
		}
	}
	if instr, rest, ok := strings.Cut(body, "## Instructions"); ok {
		_ = instr
		if rIdx := strings.Index(rest, "\n## Result"); rIdx >= 0 {
			it.Instructions = strings.TrimSpace(rest[:rIdx])
			it.Result = strings.TrimSpace(strings.TrimPrefix(rest[rIdx:], "\n## Result"))
		} else {
			it.Instructions = strings.TrimSpace(rest)
		}
	}
	return it, nil
}
