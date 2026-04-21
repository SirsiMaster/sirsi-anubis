package horus

import (
	"go/ast"
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

// --- New tests below ---

func TestGoParser_EmptyFile(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	_, err := p.ParseFile("empty.go", []byte(""))
	if err == nil {
		t.Error("expected error for empty file (no package declaration)")
	}
}

func TestGoParser_OnlyPackage(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, err := p.ParseFile("pkg.go", []byte("package foo\n"))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(symbols) != 1 {
		t.Errorf("expected 1 symbol (package), got %d", len(symbols))
	}
	if symbols[0].Kind != KindPackage {
		t.Errorf("expected package kind, got %s", symbols[0].Kind)
	}
	if symbols[0].Name != "foo" {
		t.Errorf("expected name 'foo', got %q", symbols[0].Name)
	}
}

func TestGoParser_OnlyImports(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, err := p.ParseFile("imports.go", []byte(`package example

import (
	"fmt"
	"strings"
)
`))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	// Should have package + imports (imports are GenDecl but extractGenDecl handles them as ValueSpec, not really).
	// Actually imports are skipped in extractGenDecl (no ImportSpec handler).
	hasPackage := false
	for _, s := range symbols {
		if s.Kind == KindPackage {
			hasPackage = true
		}
	}
	if !hasPackage {
		t.Error("expected package symbol")
	}
}

func TestGoParser_EmbeddedStructs(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, err := p.ParseFile("embed.go", []byte(`package example

type Base struct {
	ID int
}

type Child struct {
	Base
	Name string
}
`))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	structs := 0
	for _, s := range symbols {
		if s.Kind == KindStruct {
			structs++
		}
	}
	if structs != 2 {
		t.Errorf("struct count = %d, want 2", structs)
	}
}

func TestGoParser_ChannelTypes(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, err := p.ParseFile("chan.go", []byte(`package example

func Worker(ch chan int) {}
`))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	found := false
	for _, s := range symbols {
		if s.Name == "Worker" && s.Kind == KindFunc {
			found = true
			if !strings.Contains(s.Signature, "chan") {
				t.Errorf("Worker signature should contain 'chan': %q", s.Signature)
			}
		}
	}
	if !found {
		t.Error("Worker function not found")
	}
}

func TestGoParser_FunctionTypes(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, err := p.ParseFile("functype.go", []byte(`package example

type Handler func(string) error
`))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	found := false
	for _, s := range symbols {
		if s.Name == "Handler" && s.Kind == KindType {
			found = true
		}
	}
	if !found {
		t.Error("Handler type not found")
	}
}

func TestGoParser_MapTypes(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, err := p.ParseFile("maps.go", []byte(`package example

type Registry map[string]int
`))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	found := false
	for _, s := range symbols {
		if s.Name == "Registry" && s.Kind == KindType {
			found = true
		}
	}
	if !found {
		t.Error("Registry type not found")
	}
}

func TestGoParser_ParseDirNestedSubdirs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, "sub", "deep"), 0o755)
	os.WriteFile(filepath.Join(dir, "root.go"), []byte(`package main

func Root() {}
`), 0o644)
	os.WriteFile(filepath.Join(dir, "sub", "sub.go"), []byte(`package sub

func Sub() {}
`), 0o644)
	os.WriteFile(filepath.Join(dir, "sub", "deep", "deep.go"), []byte(`package deep

func Deep() {}
`), 0o644)

	p := NewGoParser()
	graph, err := p.ParseDir(dir)
	if err != nil {
		t.Fatalf("ParseDir: %v", err)
	}
	if graph.Stats.Files != 3 {
		t.Errorf("files = %d, want 3", graph.Stats.Files)
	}
}

func TestGoParser_ParseDirSkipVendor(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "main.go"), []byte(`package main

func Main() {}
`), 0o644)
	os.MkdirAll(filepath.Join(dir, "vendor", "lib"), 0o755)
	os.WriteFile(filepath.Join(dir, "vendor", "lib", "lib.go"), []byte(`package lib

func Lib() {}
`), 0o644)

	p := NewGoParser()
	graph, err := p.ParseDir(dir)
	if err != nil {
		t.Fatalf("ParseDir: %v", err)
	}
	if graph.Stats.Files != 1 {
		t.Errorf("files = %d, want 1 (vendor should be skipped)", graph.Stats.Files)
	}
}

func TestGoParser_ParseDirSkipTestdata(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "main.go"), []byte(`package main

func Main() {}
`), 0o644)
	os.MkdirAll(filepath.Join(dir, "testdata"), 0o755)
	os.WriteFile(filepath.Join(dir, "testdata", "fixture.go"), []byte(`package testdata

func Fixture() {}
`), 0o644)

	p := NewGoParser()
	graph, err := p.ParseDir(dir)
	if err != nil {
		t.Fatalf("ParseDir: %v", err)
	}
	if graph.Stats.Files != 1 {
		t.Errorf("files = %d, want 1 (testdata should be skipped)", graph.Stats.Files)
	}
}

func TestGoParser_ParseDirEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	p := NewGoParser()
	graph, err := p.ParseDir(dir)
	if err != nil {
		t.Fatalf("ParseDir: %v", err)
	}
	if graph.Stats.Files != 0 {
		t.Errorf("files = %d, want 0 for empty dir", graph.Stats.Files)
	}
	if len(graph.Symbols) != 0 {
		t.Errorf("symbols = %d, want 0 for empty dir", len(graph.Symbols))
	}
}

func TestQuery_ByFile(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, _ := p.ParseFile("example.go", []byte(testGoSrc))
	graph := &SymbolGraph{Root: ".", Symbols: symbols}
	q := NewQuery(graph)

	// Existing file.
	result := q.ByFile("example.go")
	if len(result) == 0 {
		t.Error("expected symbols for example.go")
	}

	// Nonexistent file.
	result = q.ByFile("nonexistent.go")
	if len(result) != 0 {
		t.Errorf("expected 0 symbols for nonexistent file, got %d", len(result))
	}
}

func TestQuery_FileOutline_NoSymbols(t *testing.T) {
	t.Parallel()
	graph := &SymbolGraph{Root: "."}
	q := NewQuery(graph)

	outline := q.FileOutline("missing.go")
	if !strings.Contains(outline, "No symbols found") {
		t.Errorf("expected 'No symbols found' message, got %q", outline)
	}
}

func TestQuery_FileOutline_ConstVarOnly(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, _ := p.ParseFile("consts.go", []byte(`package example

const ExportedConst = 42
const unexportedConst = 10
var ExportedVar = "hello"
var unexportedVar = "world"
`))
	graph := &SymbolGraph{Root: ".", Symbols: symbols}
	q := NewQuery(graph)

	outline := q.FileOutline("consts.go")
	if !strings.Contains(outline, "ExportedConst") {
		t.Error("outline should contain ExportedConst")
	}
	if !strings.Contains(outline, "ExportedVar") {
		t.Error("outline should contain ExportedVar")
	}
	// Unexported consts/vars should not appear in outline.
	if strings.Contains(outline, "unexportedConst") {
		t.Error("outline should not contain unexportedConst")
	}
	if strings.Contains(outline, "unexportedVar") {
		t.Error("outline should not contain unexportedVar")
	}
}

func TestQuery_ContextFor_NonMethod(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, _ := p.ParseFile("example.go", []byte(testGoSrc))
	graph := &SymbolGraph{Root: ".", Symbols: symbols}
	q := NewQuery(graph)

	ctx := q.ContextFor("NewPerson")
	if !strings.Contains(ctx, "func") {
		t.Error("context for NewPerson should contain 'func'")
	}
	// Should NOT contain parent type info.
	if strings.Contains(ctx, "Parent type") {
		t.Error("context for standalone func should not show parent type")
	}
}

func TestQuery_ContextFor_NotFound(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, _ := p.ParseFile("example.go", []byte(testGoSrc))
	graph := &SymbolGraph{Root: ".", Symbols: symbols}
	q := NewQuery(graph)

	ctx := q.ContextFor("DoesNotExist")
	if !strings.Contains(ctx, "not found") {
		t.Errorf("expected 'not found', got %q", ctx)
	}
}

func TestQuery_ContextFor_MethodSiblings(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, _ := p.ParseFile("example.go", []byte(testGoSrc))
	graph := &SymbolGraph{Root: ".", Symbols: symbols}
	q := NewQuery(graph)

	// String is a method on *Person; its sibling is Greet.
	ctx := q.ContextFor("String")
	if !strings.Contains(ctx, "Greet") {
		t.Error("context for String should show sibling method Greet")
	}
	if !strings.Contains(ctx, "Parent type") {
		t.Error("context for method should show parent type")
	}
}

func TestQuery_MatchSymbols_EmptyPattern(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, _ := p.ParseFile("example.go", []byte(testGoSrc))
	graph := &SymbolGraph{Root: ".", Symbols: symbols}
	q := NewQuery(graph)

	matches := q.MatchSymbols("")
	// Empty pattern — case-insensitive exact match with "". No symbol has empty name.
	if len(matches) != 0 {
		t.Errorf("expected 0 matches for empty pattern, got %d", len(matches))
	}
}

func TestQuery_MatchSymbols_MultipleWildcards(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, _ := p.ParseFile("example.go", []byte(testGoSrc))
	graph := &SymbolGraph{Root: ".", Symbols: symbols}
	q := NewQuery(graph)

	// "*e*r*" should match Greeter (G-r-e-e-t-e-r).
	matches := q.MatchSymbols("*e*r*")
	found := false
	for _, m := range matches {
		if m.Name == "Greeter" {
			found = true
		}
	}
	if !found {
		t.Error("expected Greeter to match '*e*r*'")
	}
}

func TestMatchGlob(t *testing.T) {
	t.Parallel()
	tests := []struct {
		pattern string
		name    string
		want    bool
	}{
		// Exact match case insensitive.
		{"Hello", "hello", true},
		{"HELLO", "hello", true},
		// Star matches everything.
		{"*", "anything", true},
		{"*", "", true},
		// A*B pattern.
		{"A*B", "AXB", true},
		{"A*B", "AB", true},
		{"A*B", "AXXXB", true},
		{"A*B", "AXC", false},
		// *A suffix.
		{"*son", "Person", true},
		{"*son", "person", true},
		{"*son", "Greeter", false},
		// A* prefix.
		{"New*", "NewPerson", true},
		{"New*", "new", true}, // case-insensitive: "new" matches "New*"
		{"New*", "NewPerson", true},
		// No match.
		{"xyz", "abc", false},
		// Empty pattern exact match.
		{"", "", true},
		{"", "nonempty", false},
	}
	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.name, func(t *testing.T) {
			t.Parallel()
			got := matchGlob(tt.pattern, tt.name)
			if got != tt.want {
				t.Errorf("matchGlob(%q, %q) = %v, want %v", tt.pattern, tt.name, got, tt.want)
			}
		})
	}
}

func TestTypeExprString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		expr ast.Expr
		want string
	}{
		{
			name: "Ident",
			expr: &ast.Ident{Name: "string"},
			want: "string",
		},
		{
			name: "StarExpr",
			expr: &ast.StarExpr{X: &ast.Ident{Name: "Person"}},
			want: "*Person",
		},
		{
			name: "SelectorExpr",
			expr: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Stringer"}},
			want: "fmt.Stringer",
		},
		{
			name: "ArrayType",
			expr: &ast.ArrayType{Elt: &ast.Ident{Name: "int"}},
			want: "[]int",
		},
		{
			name: "MapType",
			expr: &ast.MapType{Key: &ast.Ident{Name: "string"}, Value: &ast.Ident{Name: "int"}},
			want: "map[string]int",
		},
		{
			name: "InterfaceType",
			expr: &ast.InterfaceType{},
			want: "interface{}",
		},
		{
			name: "FuncType",
			expr: &ast.FuncType{},
			want: "func(...)",
		},
		{
			name: "ChanType",
			expr: &ast.ChanType{Value: &ast.Ident{Name: "int"}},
			want: "chan int",
		},
		{
			name: "Unknown/default",
			expr: &ast.Ellipsis{},
			want: "...",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := typeExprString(tt.expr)
			if got != tt.want {
				t.Errorf("typeExprString(%s) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestIsExported(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want bool
	}{
		{"", false},
		{"a", false},
		{"A", true},
		{"_private", false},
		{"Exported", true},
		{"myFunc", false},
		{"MyFunc", true},
	}
	for _, tt := range tests {
		t.Run("isExported_"+tt.name, func(t *testing.T) {
			t.Parallel()
			got := isExported(tt.name)
			if got != tt.want {
				t.Errorf("isExported(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestNewGraph(t *testing.T) {
	t.Parallel()
	g := NewGraph("/test/root")
	if g.Root != "/test/root" {
		t.Errorf("Root = %q, want /test/root", g.Root)
	}
	if g.BuiltAt == "" {
		t.Error("BuiltAt should not be empty")
	}
	if len(g.Symbols) != 0 {
		t.Error("new graph should have zero symbols")
	}
}

func TestComputeStats(t *testing.T) {
	t.Parallel()
	g := &SymbolGraph{
		Root: ".",
		Symbols: []Symbol{
			{Name: "main", Kind: KindPackage},
			{Name: "Foo", Kind: KindType},
			{Name: "Bar", Kind: KindStruct},
			{Name: "Baz", Kind: KindInterface},
			{Name: "DoWork", Kind: KindFunc},
			{Name: "Run", Kind: KindMethod},
			{Name: "Walk", Kind: KindMethod},
			{Name: "MaxSize", Kind: KindConst},
			{Name: "DefaultVal", Kind: KindVar},
		},
	}
	g.computeStats()

	if g.Stats.Packages != 1 {
		t.Errorf("Packages = %d, want 1", g.Stats.Packages)
	}
	// Types counts both KindType and KindStruct.
	if g.Stats.Types != 2 {
		t.Errorf("Types = %d, want 2 (Type + Struct)", g.Stats.Types)
	}
	if g.Stats.Interfaces != 1 {
		t.Errorf("Interfaces = %d, want 1", g.Stats.Interfaces)
	}
	if g.Stats.Functions != 1 {
		t.Errorf("Functions = %d, want 1", g.Stats.Functions)
	}
	if g.Stats.Methods != 2 {
		t.Errorf("Methods = %d, want 2", g.Stats.Methods)
	}
	if len(g.Packages) != 1 {
		t.Errorf("Packages list = %d, want 1", len(g.Packages))
	}
}

func TestCache_GetNonExistent(t *testing.T) {
	t.Parallel()
	c := &Cache{dir: filepath.Join(t.TempDir(), "cache")}
	_, ok := c.Get("/does/not/exist")
	if ok {
		t.Error("expected cache miss for non-existent entry")
	}
}

func TestCache_GetCorruptedFile(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "cache")
	os.MkdirAll(dir, 0o755)
	c := &Cache{dir: dir}

	// Write garbage data as a cache file.
	cachePath := c.cachePath("/test/project")
	os.WriteFile(cachePath, []byte("not valid gob data"), 0o644)

	_, ok := c.Get("/test/project")
	if ok {
		t.Error("expected cache miss for corrupted file")
	}
}

func TestCachePath(t *testing.T) {
	t.Parallel()
	c := &Cache{dir: "/tmp/test-cache"}

	tests := []struct {
		root     string
		wantBase string
	}{
		{"/some/project", "project.gob"},
		{"", "default.gob"},
		{".", "default.gob"},
		{"/", "default.gob"},
	}
	for _, tt := range tests {
		t.Run(tt.root, func(t *testing.T) {
			t.Parallel()
			path := c.cachePath(tt.root)
			base := filepath.Base(path)
			if base != tt.wantBase {
				t.Errorf("cachePath(%q) base = %q, want %q", tt.root, base, tt.wantBase)
			}
		})
	}
}

func TestInterfaceSignature(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, _ := p.ParseFile("iface.go", []byte(`package example

type Empty interface {}

type Multi interface {
	Foo()
	Bar()
}
`))

	for _, s := range symbols {
		if s.Name == "Empty" && s.Kind == KindInterface {
			if !strings.Contains(s.Signature, "interface {") {
				t.Errorf("Empty interface signature = %q, expected 'interface {'", s.Signature)
			}
		}
		if s.Name == "Multi" && s.Kind == KindInterface {
			if !strings.Contains(s.Signature, "Foo()") {
				t.Errorf("Multi interface signature should contain Foo(): %q", s.Signature)
			}
			if !strings.Contains(s.Signature, "Bar()") {
				t.Errorf("Multi interface signature should contain Bar(): %q", s.Signature)
			}
		}
	}
}

func TestStructSignature(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, _ := p.ParseFile("strct.go", []byte(`package example

type Empty struct {}

type Multi struct {
	Name string
	Age  int
}
`))

	for _, s := range symbols {
		if s.Name == "Empty" && s.Kind == KindStruct {
			if !strings.Contains(s.Signature, "struct {") {
				t.Errorf("Empty struct signature = %q, expected 'struct {'", s.Signature)
			}
		}
		if s.Name == "Multi" && s.Kind == KindStruct {
			if !strings.Contains(s.Signature, "Name") {
				t.Errorf("Multi struct signature should contain Name: %q", s.Signature)
			}
			if !strings.Contains(s.Signature, "Age") {
				t.Errorf("Multi struct signature should contain Age: %q", s.Signature)
			}
		}
	}
}

func TestGoParser_Language(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	if p.Language() != "go" {
		t.Errorf("Language = %q, want 'go'", p.Language())
	}
}

func TestQuery_LookupQualified(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, _ := p.ParseFile("example.go", []byte(testGoSrc))
	graph := &SymbolGraph{Root: ".", Symbols: symbols}
	q := NewQuery(graph)

	// Lookup by qualified name: "*Person.Greet"
	sym, ok := q.Lookup("*Person.Greet")
	if !ok {
		t.Fatal("expected to find *Person.Greet by qualified name")
	}
	if sym.Kind != KindMethod {
		t.Errorf("kind = %s, want method", sym.Kind)
	}
}

func TestQuery_ByKind_AllKinds(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, _ := p.ParseFile("example.go", []byte(testGoSrc))
	graph := &SymbolGraph{Root: ".", Symbols: symbols}
	q := NewQuery(graph)

	tests := []struct {
		kind    SymbolKind
		wantMin int
	}{
		{KindPackage, 1},
		{KindInterface, 1},
		{KindStruct, 1},
		{KindFunc, 1},
		{KindMethod, 2},
		{KindConst, 1},
		{KindVar, 1},
		{KindField, 0}, // No field-level extraction in current parser.
	}
	for _, tt := range tests {
		result := q.ByKind(tt.kind)
		if len(result) < tt.wantMin {
			t.Errorf("ByKind(%s) = %d, want >= %d", tt.kind, len(result), tt.wantMin)
		}
	}
}

func TestNewCache(t *testing.T) {
	t.Parallel()
	c := NewCache()
	if c.dir == "" {
		t.Error("NewCache should have non-empty dir")
	}
	if !strings.Contains(c.dir, "horus") {
		t.Errorf("cache dir = %q, expected to contain 'horus'", c.dir)
	}
}

func TestGoParser_DocComments(t *testing.T) {
	t.Parallel()
	p := NewGoParser()
	symbols, err := p.ParseFile("doc.go", []byte(`package example

// HelloDoc is a documented function.
func HelloDoc() {}
`))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	for _, s := range symbols {
		if s.Name == "HelloDoc" {
			if !strings.Contains(s.Doc, "HelloDoc is a documented function") {
				t.Errorf("Doc = %q, expected doc comment", s.Doc)
			}
		}
	}
}

func TestGoParser_MultiplePackages(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, "pkgA"), 0o755)
	os.MkdirAll(filepath.Join(dir, "pkgB"), 0o755)
	os.WriteFile(filepath.Join(dir, "pkgA", "a.go"), []byte(`package pkgA

func A() {}
`), 0o644)
	os.WriteFile(filepath.Join(dir, "pkgB", "b.go"), []byte(`package pkgB

func B() {}
`), 0o644)

	p := NewGoParser()
	graph, err := p.ParseDir(dir)
	if err != nil {
		t.Fatalf("ParseDir: %v", err)
	}
	if graph.Stats.Packages < 2 {
		t.Errorf("packages = %d, want >= 2", graph.Stats.Packages)
	}
}
