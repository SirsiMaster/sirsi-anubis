package horus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testGoSrc = `package example

import "fmt"

// Greeter defines something that greets.
type Greeter interface {
	Greet(name string) string
}

// Person holds a person's info.
type Person struct {
	Name string
	Age  int
}

// Greet implements Greeter.
func (p *Person) Greet(name string) string {
	return fmt.Sprintf("Hello %s, I'm %s", name, p.Name)
}

// String implements fmt.Stringer.
func (p *Person) String() string {
	return p.Name
}

// NewPerson creates a new person.
func NewPerson(name string, age int) *Person {
	return &Person{Name: name, Age: age}
}

// MaxAge is the maximum allowed age.
const MaxAge = 150

var defaultName = "World"
`

func TestGoParser_ParseFile(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, err := p.ParseFile("example.go", []byte(testGoSrc))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	kinds := make(map[SymbolKind]int)
	names := make(map[string]bool)
	for _, s := range symbols {
		kinds[s.Kind]++
		names[s.Name] = true
	}

	tests := []struct {
		kind  SymbolKind
		count int
	}{
		{KindPackage, 1},
		{KindInterface, 1},
		{KindStruct, 1},
		{KindMethod, 2},
		{KindFunc, 1},
		{KindConst, 1},
		{KindVar, 1},
	}
	for _, tt := range tests {
		if kinds[tt.kind] != tt.count {
			t.Errorf("%s count = %d, want %d", tt.kind, kinds[tt.kind], tt.count)
		}
	}

	expectedNames := []string{"example", "Greeter", "Person", "Greet", "String", "NewPerson", "MaxAge", "defaultName"}
	for _, n := range expectedNames {
		if !names[n] {
			t.Errorf("missing symbol: %s", n)
		}
	}
}

func TestGoParser_ExportedFlag(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, _ := p.ParseFile("example.go", []byte(testGoSrc))

	exported := map[string]bool{
		"example":     true,
		"Greeter":     true,
		"Person":      true,
		"Greet":       true,
		"String":      true,
		"NewPerson":   true,
		"MaxAge":      true,
		"defaultName": false,
	}

	for _, s := range symbols {
		if want, ok := exported[s.Name]; ok {
			if s.Exported != want {
				t.Errorf("%s: exported = %v, want %v", s.Name, s.Exported, want)
			}
		}
	}
}

func TestGoParser_Signatures(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, _ := p.ParseFile("example.go", []byte(testGoSrc))

	for _, s := range symbols {
		if s.Kind == KindFunc && s.Name == "NewPerson" {
			if !strings.Contains(s.Signature, "func NewPerson") {
				t.Errorf("NewPerson signature = %q, expected func signature", s.Signature)
			}
		}
		if s.Kind == KindMethod && s.Name == "Greet" {
			if s.Parent != "*Person" && s.Parent != "Person" {
				t.Errorf("Greet parent = %q, want *Person or Person", s.Parent)
			}
		}
	}
}

func TestGoParser_ParseDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "main.go"), []byte(`package main

func main() {}
`), 0o644)
	os.WriteFile(filepath.Join(dir, "util.go"), []byte(`package main

// Add returns a + b.
func Add(a, b int) int { return a + b }
`), 0o644)
	// Test files should be skipped.
	os.WriteFile(filepath.Join(dir, "main_test.go"), []byte(`package main

func TestAdd(t *testing.T) {}
`), 0o644)

	p := NewGoParser()
	graph, err := p.ParseDir(dir)
	if err != nil {
		t.Fatalf("ParseDir: %v", err)
	}

	if graph.Stats.Files != 2 {
		t.Errorf("files = %d, want 2 (test file should be skipped)", graph.Stats.Files)
	}
	if graph.Stats.Functions < 2 {
		t.Errorf("functions = %d, want >= 2", graph.Stats.Functions)
	}
}

func TestQuery_FileOutline(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, _ := p.ParseFile("example.go", []byte(testGoSrc))
	graph := &SymbolGraph{Root: ".", Symbols: symbols}
	q := NewQuery(graph)

	outline := q.FileOutline("example.go")
	if !strings.Contains(outline, "package example") {
		t.Error("outline should contain package declaration")
	}
	if !strings.Contains(outline, "Greeter") {
		t.Error("outline should contain Greeter interface")
	}
	if !strings.Contains(outline, "NewPerson") {
		t.Error("outline should contain NewPerson function")
	}
	// Outline should NOT contain function bodies.
	if strings.Contains(outline, "fmt.Sprintf") {
		t.Error("outline should not contain function body code")
	}

	// Verify it's significantly shorter than the source.
	ratio := float64(len(outline)) / float64(len(testGoSrc))
	if ratio > 0.9 {
		t.Errorf("outline ratio = %.2f, expected significant reduction", ratio)
	}
}

func TestQuery_ContextFor(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, _ := p.ParseFile("example.go", []byte(testGoSrc))
	graph := &SymbolGraph{Root: ".", Symbols: symbols}
	q := NewQuery(graph)

	ctx := q.ContextFor("Greet")
	if !strings.Contains(ctx, "method") {
		t.Error("context should identify Greet as a method")
	}
	if !strings.Contains(ctx, "Person") {
		t.Error("context should reference parent type (contains Person)")
	}
	// Sibling method String should appear in context.
	if !strings.Contains(ctx, "String") {
		t.Error("context should show sibling method String")
	}
}

func TestQuery_Lookup(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, _ := p.ParseFile("example.go", []byte(testGoSrc))
	graph := &SymbolGraph{Root: ".", Symbols: symbols}
	q := NewQuery(graph)

	sym, ok := q.Lookup("NewPerson")
	if !ok {
		t.Fatal("NewPerson not found")
	}
	if sym.Kind != KindFunc {
		t.Errorf("kind = %s, want func", sym.Kind)
	}

	_, ok = q.Lookup("nonexistent")
	if ok {
		t.Error("should not find nonexistent symbol")
	}
}

func TestQuery_ByKind(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, _ := p.ParseFile("example.go", []byte(testGoSrc))
	graph := &SymbolGraph{Root: ".", Symbols: symbols}
	q := NewQuery(graph)

	methods := q.ByKind(KindMethod)
	if len(methods) != 2 {
		t.Errorf("method count = %d, want 2", len(methods))
	}
}

func TestQuery_MatchSymbols(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, _ := p.ParseFile("example.go", []byte(testGoSrc))
	graph := &SymbolGraph{Root: ".", Symbols: symbols}
	q := NewQuery(graph)

	tests := []struct {
		pattern string
		wantMin int
	}{
		{"*Person*", 2}, // Person and NewPerson
		{"Greet*", 2},   // Greet and Greeter
		{"MaxAge", 1},
		{"*", len(symbols)},
	}
	for _, tt := range tests {
		matches := q.MatchSymbols(tt.pattern)
		if len(matches) < tt.wantMin {
			t.Errorf("MatchSymbols(%q) = %d matches, want >= %d", tt.pattern, len(matches), tt.wantMin)
		}
	}
}

func TestCache_PutGet(t *testing.T) {
	t.Parallel()
	c := &Cache{dir: filepath.Join(t.TempDir(), "cache")}

	graph := &SymbolGraph{
		Root:     "/test/project",
		Packages: []string{"main"},
		Symbols: []Symbol{
			{Name: "Foo", Kind: KindFunc, File: "main.go", Line: 1},
		},
	}

	if err := c.Put("/test/project", graph); err != nil {
		t.Fatalf("Put: %v", err)
	}

	got, ok := c.Get("/test/project")
	if !ok {
		t.Fatal("cache miss after Put")
	}
	if len(got.Symbols) != 1 || got.Symbols[0].Name != "Foo" {
		t.Errorf("cached graph mismatch: %+v", got)
	}

	c.Invalidate("/test/project")
	_, ok = c.Get("/test/project")
	if ok {
		t.Error("expected cache miss after Invalidate")
	}
}
