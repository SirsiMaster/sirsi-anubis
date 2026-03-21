// Package mapper generates infrastructure graph visualizations.
// Produces a self-contained HTML file with interactive network graph
// using Sigma.js (WebGL) and Graphology (MIT licensed).
//
// No external dependencies required — the HTML file includes
// all JavaScript inline and opens in any browser.
package mapper

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// NodeType categorizes nodes in the infrastructure graph.
type NodeType string

const (
	NodeDevice    NodeType = "device"
	NodeApp       NodeType = "app"
	NodeGhost     NodeType = "ghost"
	NodeContainer NodeType = "container"
	NodeProcess   NodeType = "process"
	NodeCache     NodeType = "cache"
	NodeVolume    NodeType = "volume"
	NodeNetwork   NodeType = "network"
	NodeService   NodeType = "service"
)

// Node represents a single entity in the infrastructure graph.
type Node struct {
	ID       string            `json:"id"`
	Label    string            `json:"label"`
	Type     NodeType          `json:"type"`
	Size     int64             `json:"size,omitempty"` // bytes
	Color    string            `json:"color,omitempty"`
	X        float64           `json:"x"`
	Y        float64           `json:"y"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Edge represents a relationship between two nodes.
type Edge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label,omitempty"`
	Weight int    `json:"weight,omitempty"`
}

// InfraGraph is the complete infrastructure map.
type InfraGraph struct {
	Nodes       []Node `json:"nodes"`
	Edges       []Edge `json:"edges"`
	GeneratedAt string `json:"generated_at"`
	Hostname    string `json:"hostname"`
	Platform    string `json:"platform"`
}

// NodeColors maps node types to display colors.
var NodeColors = map[NodeType]string{
	NodeDevice:    "#C8A951", // Anubis gold
	NodeApp:       "#4A90D9", // Blue
	NodeGhost:     "#E74C3C", // Red — ghosts glow red
	NodeContainer: "#2ECC71", // Green
	NodeProcess:   "#F39C12", // Orange
	NodeCache:     "#95A5A6", // Gray
	NodeVolume:    "#8E44AD", // Purple
	NodeNetwork:   "#1ABC9C", // Teal
	NodeService:   "#3498DB", // Light blue
}

// NewGraph creates an empty infrastructure graph.
func NewGraph() *InfraGraph {
	hostname, _ := os.Hostname()
	return &InfraGraph{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Hostname:    hostname,
		Platform:    fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// AddNode adds a node to the graph.
func (g *InfraGraph) AddNode(id, label string, nodeType NodeType) {
	color, ok := NodeColors[nodeType]
	if !ok {
		color = "#CCCCCC"
	}
	g.Nodes = append(g.Nodes, Node{
		ID:    id,
		Label: label,
		Type:  nodeType,
		Color: color,
	})
}

// AddEdge connects two nodes.
func (g *InfraGraph) AddEdge(source, target, label string) {
	id := fmt.Sprintf("%s->%s", source, target)
	g.Edges = append(g.Edges, Edge{
		ID:     id,
		Source: source,
		Target: target,
		Label:  label,
	})
}

// ToJSON exports the graph as JSON.
func (g *InfraGraph) ToJSON() ([]byte, error) {
	return json.MarshalIndent(g, "", "  ")
}

// RenderHTML produces a self-contained HTML file with an interactive graph.
// Uses Sigma.js v2 (CDN) and Graphology for rendering.
func (g *InfraGraph) RenderHTML(outputPath string) error {
	graphJSON, err := json.Marshal(g)
	if err != nil {
		return fmt.Errorf("marshal graph: %w", err)
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>𓂀 Anubis Infrastructure Map — %s</title>
<script src="https://cdnjs.cloudflare.com/ajax/libs/graphology/0.25.4/graphology.umd.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/sigma.js/2.4.0/sigma.min.js"></script>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    background: #0D0D1A;
    color: #C8A951;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
    overflow: hidden;
  }
  #header {
    position: fixed; top: 0; left: 0; right: 0; z-index: 10;
    background: linear-gradient(180deg, rgba(13,13,26,0.95) 0%%, rgba(13,13,26,0) 100%%);
    padding: 20px 30px;
  }
  #header h1 { font-size: 24px; color: #C8A951; }
  #header p { font-size: 13px; color: #888; margin-top: 4px; }
  #graph { width: 100vw; height: 100vh; }
  #legend {
    position: fixed; bottom: 20px; left: 20px; z-index: 10;
    background: rgba(13,13,26,0.9); border: 1px solid #333;
    border-radius: 8px; padding: 16px; font-size: 12px;
  }
  .legend-item { display: flex; align-items: center; margin: 4px 0; }
  .legend-dot {
    width: 10px; height: 10px; border-radius: 50%%;
    margin-right: 8px; display: inline-block;
  }
  #stats {
    position: fixed; top: 20px; right: 20px; z-index: 10;
    background: rgba(13,13,26,0.9); border: 1px solid #333;
    border-radius: 8px; padding: 16px; font-size: 12px; text-align: right;
  }
  #tooltip {
    position: fixed; display: none; z-index: 20;
    background: rgba(13,13,26,0.95); border: 1px solid #C8A951;
    border-radius: 6px; padding: 10px 14px; font-size: 12px;
    pointer-events: none; max-width: 280px;
  }
</style>
</head>
<body>
<div id="header">
  <h1>𓂀 Anubis Infrastructure Map</h1>
  <p>%s — %s — Generated %s</p>
</div>
<div id="graph"></div>
<div id="legend"></div>
<div id="stats"></div>
<div id="tooltip"></div>

<script>
const data = %s;

// Build graph
const graph = new graphology.Graph();

// Add nodes with force-layout positions
const typeCount = {};
data.nodes.forEach((n, i) => {
  typeCount[n.type] = (typeCount[n.type] || 0) + 1;
  const angle = (2 * Math.PI * i) / data.nodes.length;
  const radius = 3 + Math.random() * 2;
  graph.addNode(n.id, {
    label: n.label,
    x: n.x || Math.cos(angle) * radius + (Math.random() - 0.5),
    y: n.y || Math.sin(angle) * radius + (Math.random() - 0.5),
    size: Math.max(6, Math.min(20, Math.log2((n.size || 1024) / 1024) + 4)),
    color: n.color || '#C8A951',
    type: n.type
  });
});

data.edges.forEach(e => {
  if (graph.hasNode(e.source) && graph.hasNode(e.target)) {
    graph.addEdge(e.source, e.target, {
      label: e.label || '',
      size: 1,
      color: '#333'
    });
  }
});

// Render
const container = document.getElementById('graph');
const renderer = new Sigma(graph, container, {
  renderLabels: true,
  labelSize: 11,
  labelColor: { color: '#C8A951' },
  defaultEdgeColor: '#333',
  defaultNodeColor: '#C8A951',
  minCameraRatio: 0.1,
  maxCameraRatio: 10,
});

// Legend
const legendEl = document.getElementById('legend');
const colors = %s;
let legendHTML = '';
for (const [type, color] of Object.entries(colors)) {
  const count = typeCount[type] || 0;
  if (count > 0) {
    legendHTML += '<div class="legend-item"><span class="legend-dot" style="background:' + color + '"></span>' + type + ' (' + count + ')</div>';
  }
}
legendEl.innerHTML = legendHTML;

// Stats
document.getElementById('stats').innerHTML =
  '<div style="color:#C8A951;font-size:14px;font-weight:600">Stats</div>' +
  '<div>Nodes: ' + data.nodes.length + '</div>' +
  '<div>Edges: ' + data.edges.length + '</div>' +
  '<div>Platform: ' + data.platform + '</div>';

// Tooltip on hover
const tooltip = document.getElementById('tooltip');
renderer.on('enterNode', ({node}) => {
  const attrs = graph.getNodeAttributes(node);
  tooltip.style.display = 'block';
  tooltip.innerHTML = '<div style="color:#C8A951;font-weight:600">' + attrs.label + '</div>' +
    '<div style="color:#888">Type: ' + (attrs.type || 'unknown') + '</div>';
});
renderer.on('leaveNode', () => { tooltip.style.display = 'none'; });
renderer.getMouseCaptor().on('mousemove', (e) => {
  tooltip.style.left = e.original.clientX + 12 + 'px';
  tooltip.style.top = e.original.clientY + 12 + 'px';
});
</script>
</body>
</html>`, g.Hostname, g.Hostname, g.Platform, g.GeneratedAt,
		string(graphJSON),
		mustJSON(NodeColors))

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	return os.WriteFile(outputPath, []byte(html), 0644)
}

func mustJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
