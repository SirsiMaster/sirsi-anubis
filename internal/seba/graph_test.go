package seba

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// ── NewGraph ────────────────────────────────────────────────────────────

func TestNewGraph(t *testing.T) {
	g := NewGraph()
	if g == nil {
		t.Fatal("NewGraph returned nil")
	}
	if g.GeneratedAt == "" {
		t.Error("GeneratedAt should be set")
	}
	if g.Platform == "" {
		t.Error("Platform should be set")
	}
	if !strings.Contains(g.Platform, runtime.GOOS) {
		t.Errorf("Platform should contain %s, got: %s", runtime.GOOS, g.Platform)
	}
	if len(g.Nodes) != 0 {
		t.Errorf("Initial nodes should be empty, got %d", len(g.Nodes))
	}
	if len(g.Edges) != 0 {
		t.Errorf("Initial edges should be empty, got %d", len(g.Edges))
	}
}

// ── AddNode ─────────────────────────────────────────────────────────────

func TestAddNode(t *testing.T) {
	g := NewGraph()
	g.AddNode("dev1", "MacBook Pro", NodeDevice)
	g.AddNode("app1", "VS Code", NodeApp)
	g.AddNode("ghost1", "Phantom Process", NodeGhost)

	if len(g.Nodes) != 3 {
		t.Fatalf("Expected 3 nodes, got %d", len(g.Nodes))
	}

	// Verify device node
	dev := g.Nodes[0]
	if dev.ID != "dev1" {
		t.Errorf("ID: expected dev1, got %s", dev.ID)
	}
	if dev.Label != "MacBook Pro" {
		t.Errorf("Label: expected MacBook Pro, got %s", dev.Label)
	}
	if dev.Type != NodeDevice {
		t.Errorf("Type: expected device, got %s", dev.Type)
	}
	if dev.Color != "#C8A951" {
		t.Errorf("Color: expected #C8A951, got %s", dev.Color)
	}

	// Ghost should get red
	ghost := g.Nodes[2]
	if ghost.Color != "#E74C3C" {
		t.Errorf("Ghost color: expected #E74C3C, got %s", ghost.Color)
	}
}

func TestAddNode_UnknownType(t *testing.T) {
	g := NewGraph()
	g.AddNode("unk1", "Unknown Entity", NodeType("alien"))

	if len(g.Nodes) != 1 {
		t.Fatal("Expected 1 node")
	}
	if g.Nodes[0].Color != "#CCCCCC" {
		t.Errorf("Unknown type should get default color #CCCCCC, got %s", g.Nodes[0].Color)
	}
}

// ── AddEdge ─────────────────────────────────────────────────────────────

func TestAddEdge(t *testing.T) {
	g := NewGraph()
	g.AddNode("a", "Node A", NodeApp)
	g.AddNode("b", "Node B", NodeProcess)
	g.AddEdge("a", "b", "spawns")

	if len(g.Edges) != 1 {
		t.Fatalf("Expected 1 edge, got %d", len(g.Edges))
	}

	e := g.Edges[0]
	if e.ID != "a->b" {
		t.Errorf("Edge ID: expected a->b, got %s", e.ID)
	}
	if e.Source != "a" {
		t.Errorf("Source: expected a, got %s", e.Source)
	}
	if e.Target != "b" {
		t.Errorf("Target: expected b, got %s", e.Target)
	}
	if e.Label != "spawns" {
		t.Errorf("Label: expected spawns, got %s", e.Label)
	}
}

func TestAddEdge_Multiple(t *testing.T) {
	g := NewGraph()
	g.AddNode("x", "X", NodeService)
	g.AddNode("y", "Y", NodeService)
	g.AddNode("z", "Z", NodeService)
	g.AddEdge("x", "y", "calls")
	g.AddEdge("x", "z", "depends")
	g.AddEdge("y", "z", "proxies")

	if len(g.Edges) != 3 {
		t.Errorf("Expected 3 edges, got %d", len(g.Edges))
	}
}

// ── ToJSON ──────────────────────────────────────────────────────────────

func TestToJSON(t *testing.T) {
	g := NewGraph()
	g.AddNode("dev1", "MacBook", NodeDevice)
	g.AddNode("docker1", "nginx", NodeContainer)
	g.AddEdge("dev1", "docker1", "runs")

	data, err := g.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var parsed InfraGraph
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("ToJSON produced invalid JSON: %v", err)
	}

	if len(parsed.Nodes) != 2 {
		t.Errorf("Expected 2 nodes in JSON, got %d", len(parsed.Nodes))
	}
	if len(parsed.Edges) != 1 {
		t.Errorf("Expected 1 edge in JSON, got %d", len(parsed.Edges))
	}
	if parsed.Hostname == "" {
		t.Error("Hostname should be populated")
	}
}

func TestToJSON_EmptyGraph(t *testing.T) {
	g := NewGraph()
	data, err := g.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON empty graph failed: %v", err)
	}
	if !strings.Contains(string(data), "nodes") {
		t.Error("JSON should contain 'nodes' key")
	}
}

// ── RenderHTML ──────────────────────────────────────────────────────────

func TestRenderHTML(t *testing.T) {
	g := NewGraph()
	g.AddNode("dev1", "WorkStation", NodeDevice)
	g.AddNode("app1", "VS Code", NodeApp)
	g.AddNode("ghost1", "Phantom", NodeGhost)
	g.AddNode("container1", "nginx", NodeContainer)
	g.AddNode("proc1", "node", NodeProcess)
	g.AddNode("cache1", "npm cache", NodeCache)
	g.AddNode("vol1", "docker volume", NodeVolume)
	g.AddNode("net1", "bridge", NodeNetwork)
	g.AddNode("svc1", "api-gateway", NodeService)
	g.AddEdge("dev1", "app1", "runs")
	g.AddEdge("dev1", "container1", "hosts")
	g.AddEdge("app1", "proc1", "spawns")
	g.AddEdge("container1", "net1", "attached")
	g.AddEdge("ghost1", "dev1", "haunts")

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output", "graph.html")

	err := g.RenderHTML(outputPath)
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}

	// Verify file exists and has content
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Output file not found: %v", err)
	}
	if info.Size() < 1000 {
		t.Errorf("HTML file seems too small: %d bytes", info.Size())
	}

	// Read and verify structure
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	html := string(content)

	// Check for critical HTML elements
	checks := []string{
		"<!DOCTYPE html>",
		"<canvas",
		"Seba",
		"Infrastructure Map",
		"requestAnimationFrame",
		"WorkStation", // node label
	}
	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Errorf("HTML should contain '%s'", check)
		}
	}

	t.Logf("RenderHTML: %d bytes, %d nodes, %d edges",
		info.Size(), len(g.Nodes), len(g.Edges))
}

func TestRenderHTML_NestedDir(t *testing.T) {
	g := NewGraph()
	g.AddNode("x", "Test", NodeApp)

	tmpDir := t.TempDir()
	deepPath := filepath.Join(tmpDir, "a", "b", "c", "graph.html")

	err := g.RenderHTML(deepPath)
	if err != nil {
		t.Fatalf("RenderHTML should create nested dirs: %v", err)
	}
	if _, err := os.Stat(deepPath); err != nil {
		t.Error("Output file not created in nested dir")
	}
}

// ── NodeColors ──────────────────────────────────────────────────────────

func TestNodeColors(t *testing.T) {
	allTypes := []NodeType{
		NodeDevice, NodeApp, NodeGhost, NodeContainer,
		NodeProcess, NodeCache, NodeVolume, NodeNetwork, NodeService,
	}

	for _, nt := range allTypes {
		color, ok := NodeColors[nt]
		if !ok {
			t.Errorf("NodeColors missing entry for %s", nt)
			continue
		}
		if !strings.HasPrefix(color, "#") {
			t.Errorf("Color for %s should start with #: %s", nt, color)
		}
		if len(color) != 7 {
			t.Errorf("Color for %s should be 7 chars (#RRGGBB): %s", nt, color)
		}
	}
}

// ── mustJSON ────────────────────────────────────────────────────────────

func TestMustJSON(t *testing.T) {
	result := mustJSON(map[string]int{"a": 1, "b": 2})
	if result == "" {
		t.Error("mustJSON returned empty string")
	}

	var parsed map[string]int
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("mustJSON produced invalid JSON: %v", err)
	}
	if parsed["a"] != 1 || parsed["b"] != 2 {
		t.Errorf("mustJSON round-trip failed: %v", parsed)
	}
}

// ── Integration: full pipeline ──────────────────────────────────────────

func TestFullPipeline(t *testing.T) {
	g := NewGraph()

	// Build a realistic graph
	g.AddNode("laptop", "Dev Laptop", NodeDevice)
	g.AddNode("vscode", "VS Code", NodeApp)
	g.AddNode("docker", "Docker Desktop", NodeApp)
	g.AddNode("nginx", "nginx:latest", NodeContainer)
	g.AddNode("redis", "redis:7", NodeContainer)
	g.AddNode("node", "node", NodeProcess)
	g.AddNode("ghost-xpc", "XPC Service (dead)", NodeGhost)
	g.AddNode("npm-cache", "npm cache", NodeCache)
	g.AddNode("db-vol", "postgres-data", NodeVolume)
	g.AddNode("bridge0", "docker0", NodeNetwork)

	g.AddEdge("laptop", "vscode", "runs")
	g.AddEdge("laptop", "docker", "runs")
	g.AddEdge("vscode", "node", "spawns")
	g.AddEdge("docker", "nginx", "hosts")
	g.AddEdge("docker", "redis", "hosts")
	g.AddEdge("nginx", "bridge0", "attached")
	g.AddEdge("redis", "bridge0", "attached")
	g.AddEdge("redis", "db-vol", "mounts")
	g.AddEdge("node", "npm-cache", "reads")
	g.AddEdge("ghost-xpc", "laptop", "haunts")

	// Export JSON
	jsonData, err := g.ToJSON()
	if err != nil {
		t.Fatal(err)
	}

	var roundTrip InfraGraph
	if err := json.Unmarshal(jsonData, &roundTrip); err != nil {
		t.Fatal(err)
	}

	if len(roundTrip.Nodes) != 10 {
		t.Errorf("Expected 10 nodes, got %d", len(roundTrip.Nodes))
	}
	if len(roundTrip.Edges) != 10 {
		t.Errorf("Expected 10 edges, got %d", len(roundTrip.Edges))
	}

	// Render HTML
	tmpDir := t.TempDir()
	htmlPath := filepath.Join(tmpDir, "infra.html")
	if err := g.RenderHTML(htmlPath); err != nil {
		t.Fatal(err)
	}

	info, _ := os.Stat(htmlPath)
	t.Logf("Pipeline: %d nodes, %d edges, %d bytes HTML", len(g.Nodes), len(g.Edges), info.Size())
}
