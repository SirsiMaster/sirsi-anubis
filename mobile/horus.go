package mobile

import (
	"sync"

	"github.com/SirsiMaster/sirsi-pantheon/internal/horus"
)

// horusCache is a package-level cache for parsed SymbolGraphs.
// Avoids re-parsing large codebases on repeated mobile calls.
var (
	horusCacheMu sync.Mutex
	horusCache   *horus.Cache
)

func getHorusCache() *horus.Cache {
	horusCacheMu.Lock()
	defer horusCacheMu.Unlock()
	if horusCache == nil {
		horusCache = horus.NewCache()
	}
	return horusCache
}

// parseOrCached parses a directory or returns the cached graph.
func parseOrCached(root string) (*horus.SymbolGraph, error) {
	cache := getHorusCache()

	if g, ok := cache.Get(root); ok {
		return g, nil
	}

	parser := horus.NewGoParser()
	g, err := parser.ParseDir(root)
	if err != nil {
		return nil, err
	}

	// Best-effort cache write; ignore errors.
	_ = cache.Put(root, g)
	return g, nil
}

// HorusParseDir parses a directory and returns the full SymbolGraph as JSON.
// Returns Response JSON with SymbolGraph data.
func HorusParseDir(root string) string {
	if root == "" {
		return errorJSON("root path is required")
	}

	graph, err := parseOrCached(root)
	if err != nil {
		return errorJSON("horus parse: " + err.Error())
	}

	return successJSON(graph)
}

// HorusFileOutline returns a compact outline of a single file's symbols.
// Parses the project at root, then extracts the outline for filePath.
// Returns Response JSON with {"outline": "..."} string.
func HorusFileOutline(root, filePath string) string {
	if root == "" || filePath == "" {
		return errorJSON("root and filePath are required")
	}

	graph, err := parseOrCached(root)
	if err != nil {
		return errorJSON("horus parse: " + err.Error())
	}

	q := horus.NewQuery(graph)
	outline := q.FileOutline(filePath)

	return successJSON(map[string]string{"outline": outline})
}

// HorusContextFor returns the minimal context needed to understand a symbol.
// Returns Response JSON with {"context": "..."} string.
func HorusContextFor(root, symbolName string) string {
	if root == "" || symbolName == "" {
		return errorJSON("root and symbolName are required")
	}

	graph, err := parseOrCached(root)
	if err != nil {
		return errorJSON("horus parse: " + err.Error())
	}

	q := horus.NewQuery(graph)
	ctx := q.ContextFor(symbolName)

	return successJSON(map[string]string{"context": ctx})
}

// HorusMatchSymbols returns symbols whose names match a glob pattern.
// Returns Response JSON with []Symbol data.
func HorusMatchSymbols(root, pattern string) string {
	if root == "" || pattern == "" {
		return errorJSON("root and pattern are required")
	}

	graph, err := parseOrCached(root)
	if err != nil {
		return errorJSON("horus parse: " + err.Error())
	}

	q := horus.NewQuery(graph)
	matches := q.MatchSymbols(pattern)

	return successJSON(matches)
}
