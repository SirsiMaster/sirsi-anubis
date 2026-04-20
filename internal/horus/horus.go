// Package horus implements a structural code graph that extracts symbols
// (types, functions, methods, interfaces) from Go source and serves
// compact outlines and signatures instead of full files.
//
// Subsumes the external Code Review Graph tool as native Go inside Pantheon.
// Phase 1 uses go/ast for Go files. Phase 2 adds Tree-sitter for multi-language support.
package horus

import "time"

// SymbolKind categorizes code symbols.
type SymbolKind string

const (
	KindPackage   SymbolKind = "package"
	KindImport    SymbolKind = "import"
	KindType      SymbolKind = "type"
	KindInterface SymbolKind = "interface"
	KindStruct    SymbolKind = "struct"
	KindFunc      SymbolKind = "func"
	KindMethod    SymbolKind = "method"
	KindConst     SymbolKind = "const"
	KindVar       SymbolKind = "var"
	KindField     SymbolKind = "field"
)

// Symbol represents a named code entity extracted from source.
type Symbol struct {
	Name      string     `json:"name"`
	Kind      SymbolKind `json:"kind"`
	File      string     `json:"file"`
	Line      int        `json:"line"`
	EndLine   int        `json:"endLine"`
	Signature string     `json:"signature"`
	Doc       string     `json:"doc,omitempty"`
	Exported  bool       `json:"exported"`
	Parent    string     `json:"parent,omitempty"`
}

// SymbolGraph is the complete structural map of a codebase.
type SymbolGraph struct {
	Root     string     `json:"root"`
	Packages []string   `json:"packages"`
	Symbols  []Symbol   `json:"symbols"`
	Stats    GraphStats `json:"stats"`
	BuiltAt  string     `json:"builtAt"`
}

// GraphStats holds metrics about the analyzed code.
type GraphStats struct {
	Files      int `json:"files"`
	Packages   int `json:"packages"`
	Types      int `json:"types"`
	Functions  int `json:"functions"`
	Methods    int `json:"methods"`
	Interfaces int `json:"interfaces"`
	TotalLines int `json:"totalLines"`
}

// Parser extracts symbols from source code.
type Parser interface {
	Language() string
	ParseFile(path string, src []byte) ([]Symbol, error)
	ParseDir(root string) (*SymbolGraph, error)
}

// NewGraph creates an empty symbol graph for a project root.
func NewGraph(root string) *SymbolGraph {
	return &SymbolGraph{
		Root:    root,
		BuiltAt: time.Now().Format(time.RFC3339),
	}
}

// computeStats updates the Stats field based on current symbols.
func (g *SymbolGraph) computeStats() {
	pkgSet := make(map[string]bool)
	for _, s := range g.Symbols {
		switch s.Kind {
		case KindPackage:
			pkgSet[s.Name] = true
		case KindType, KindStruct:
			g.Stats.Types++
		case KindInterface:
			g.Stats.Interfaces++
		case KindFunc:
			g.Stats.Functions++
		case KindMethod:
			g.Stats.Methods++
		}
	}
	g.Stats.Packages = len(pkgSet)
	g.Packages = make([]string, 0, len(pkgSet))
	for pkg := range pkgSet {
		g.Packages = append(g.Packages, pkg)
	}
}
