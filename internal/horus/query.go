package horus

import (
	"fmt"
	"strings"
)

// Query retrieves specific symbols from a SymbolGraph.
type Query struct {
	Graph *SymbolGraph
}

// NewQuery creates a query interface for the given graph.
func NewQuery(g *SymbolGraph) *Query {
	return &Query{Graph: g}
}

// Lookup returns the symbol with the given name. For methods, use "Type.Method".
func (q *Query) Lookup(name string) (*Symbol, bool) {
	for i, s := range q.Graph.Symbols {
		qualified := s.Name
		if s.Parent != "" {
			qualified = s.Parent + "." + s.Name
		}
		if qualified == name || s.Name == name {
			return &q.Graph.Symbols[i], true
		}
	}
	return nil, false
}

// ByKind returns all symbols of a given kind.
func (q *Query) ByKind(kind SymbolKind) []Symbol {
	var result []Symbol
	for _, s := range q.Graph.Symbols {
		if s.Kind == kind {
			result = append(result, s)
		}
	}
	return result
}

// ByFile returns all symbols in a given file.
func (q *Query) ByFile(path string) []Symbol {
	var result []Symbol
	for _, s := range q.Graph.Symbols {
		if s.File == path {
			result = append(result, s)
		}
	}
	return result
}

// FileOutline returns a compact outline of a file: package, imports,
// type declarations, function signatures. No function bodies.
// This is the primary token-saving mechanism: 8-49x smaller than full source.
func (q *Query) FileOutline(path string) string {
	symbols := q.ByFile(path)
	if len(symbols) == 0 {
		return fmt.Sprintf("// No symbols found in %s", path)
	}

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("// Outline: %s\n", path))

	for _, s := range symbols {
		switch s.Kind {
		case KindPackage:
			buf.WriteString(s.Signature)
			buf.WriteString("\n\n")
		case KindImport:
			// Skip individual imports in outline.
		case KindType, KindStruct, KindInterface:
			if s.Doc != "" {
				buf.WriteString("// ")
				buf.WriteString(strings.Split(s.Doc, "\n")[0])
				buf.WriteString("\n")
			}
			buf.WriteString(s.Signature)
			buf.WriteString("\n\n")
		case KindFunc, KindMethod:
			if s.Doc != "" {
				buf.WriteString("// ")
				buf.WriteString(strings.Split(s.Doc, "\n")[0])
				buf.WriteString("\n")
			}
			buf.WriteString(s.Signature)
			buf.WriteString("\n\n")
		case KindConst, KindVar:
			if s.Exported {
				buf.WriteString(fmt.Sprintf("%s %s\n", s.Kind, s.Name))
			}
		}
	}

	return strings.TrimSpace(buf.String())
}

// ContextFor returns the minimal context needed to understand a symbol:
// its declaration, doc, parent type (if method), and sibling methods.
func (q *Query) ContextFor(name string) string {
	sym, ok := q.Lookup(name)
	if !ok {
		return fmt.Sprintf("// Symbol %q not found", name)
	}

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("// Context for: %s (%s)\n", name, sym.Kind))
	buf.WriteString(fmt.Sprintf("// File: %s:%d-%d\n", sym.File, sym.Line, sym.EndLine))

	if sym.Doc != "" {
		buf.WriteString("// ")
		buf.WriteString(sym.Doc)
		buf.WriteString("\n")
	}

	buf.WriteString(sym.Signature)
	buf.WriteString("\n")

	// If it's a method, show the parent type and sibling methods.
	if sym.Parent != "" {
		buf.WriteString(fmt.Sprintf("\n// Parent type: %s\n", sym.Parent))
		for _, s := range q.Graph.Symbols {
			if s.Parent == sym.Parent && s.Name != sym.Name && s.Kind == KindMethod {
				buf.WriteString(fmt.Sprintf("//   %s\n", s.Signature))
			}
		}
	}

	return strings.TrimSpace(buf.String())
}

// MatchSymbols returns symbols whose names match the given glob-like pattern.
// Supports * as wildcard.
func (q *Query) MatchSymbols(pattern string) []Symbol {
	var result []Symbol
	for _, s := range q.Graph.Symbols {
		if matchGlob(pattern, s.Name) {
			result = append(result, s)
		}
	}
	return result
}

// matchGlob performs simple glob matching with * as wildcard.
func matchGlob(pattern, name string) bool {
	if pattern == "*" {
		return true
	}
	if !strings.Contains(pattern, "*") {
		return strings.EqualFold(pattern, name)
	}
	// Split on * and check that all parts appear in order.
	parts := strings.Split(strings.ToLower(pattern), "*")
	lower := strings.ToLower(name)
	pos := 0
	for _, part := range parts {
		if part == "" {
			continue
		}
		idx := strings.Index(lower[pos:], part)
		if idx < 0 {
			return false
		}
		pos += idx + len(part)
	}
	return true
}
