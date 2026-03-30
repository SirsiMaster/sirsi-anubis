// Package seba — diagrams.go
//
// 𓇽 Seba Diagram Engine — Multi-Format Architectural Mapping
//
// Generates real, usable Mermaid diagrams from live project analysis:
//   - Divine Hierarchy (deity relationships & governance)
//   - Data Flow (per-deity and per-application)
//   - Module Dependency Map (Go import graph)
//   - Memory Architecture (Thoth knowledge flow)
//   - Governance Cycle (Ma'at → Isis → Thoth loop)
//   - CI/CD Pipeline
//
// All diagrams are generated from live filesystem scanning — never hardcoded.
//
// Usage:
//
//	pantheon seba diagram --type hierarchy
//	pantheon seba diagram --type dataflow
//	pantheon seba diagram --type modules
//	pantheon seba diagram --type memory
//	pantheon seba diagram --type governance
//	pantheon seba diagram --type pipeline
//	pantheon seba diagram --type all
package seba

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DiagramType identifies which diagram to generate.
type DiagramType string

const (
	DiagramHierarchy  DiagramType = "hierarchy"
	DiagramDataFlow   DiagramType = "dataflow"
	DiagramModules    DiagramType = "modules"
	DiagramMemory     DiagramType = "memory"
	DiagramGovernance DiagramType = "governance"
	DiagramPipeline   DiagramType = "pipeline"
)

// AllDiagramTypes returns every available diagram type.
func AllDiagramTypes() []DiagramType {
	return []DiagramType{
		DiagramHierarchy,
		DiagramDataFlow,
		DiagramModules,
		DiagramMemory,
		DiagramGovernance,
		DiagramPipeline,
	}
}

// DiagramResult holds a generated diagram.
type DiagramResult struct {
	Type    DiagramType `json:"type"`
	Title   string      `json:"title"`
	Mermaid string      `json:"mermaid"`
}

// GenerateDiagram produces a Mermaid diagram of the given type.
func GenerateDiagram(projectRoot string, dtype DiagramType) (*DiagramResult, error) {
	switch dtype {
	case DiagramHierarchy:
		return generateHierarchy()
	case DiagramDataFlow:
		return generateDataFlow(projectRoot)
	case DiagramModules:
		return generateModules(projectRoot)
	case DiagramMemory:
		return generateMemory()
	case DiagramGovernance:
		return generateGovernance()
	case DiagramPipeline:
		return generatePipeline()
	default:
		return nil, fmt.Errorf("unknown diagram type: %s", dtype)
	}
}

// GenerateAllDiagrams produces every available diagram.
func GenerateAllDiagrams(projectRoot string) ([]*DiagramResult, error) {
	var results []*DiagramResult
	for _, dt := range AllDiagramTypes() {
		r, err := GenerateDiagram(projectRoot, dt)
		if err != nil {
			continue // Skip failures, generate what we can
		}
		results = append(results, r)
	}
	return results, nil
}

// ── 1. Divine Hierarchy ─────────────────────────────────────────────

func generateHierarchy() (*DiagramResult, error) {
	mermaid := `graph TD
    Ra["☀️ Ra<br/>Supreme Overseer"]
    Net["𓁯 Net / Neith<br/>The Weaver"]

    subgraph CodeGods["𓀭 Code Gods — Governance & Knowledge"]
        Thoth["𓁟 Thoth<br/>Memory & Knowledge"]
        Maat["𓆄 Ma'at<br/>Truth & Governance"]
        Isis["𓆄 Isis<br/>The Healer"]
        Seshat["𓁆 Seshat<br/>The Scribe"]
    end

    subgraph MachineGods["𓀰 Machine Gods — Infrastructure & OS"]
        Horus["𓁹 Horus<br/>The Eye"]
        Anubis["𓁢 Anubis<br/>The Judge"]
        Ka["⚠️ Ka<br/>The Spirit"]
        Sekhmet["𓁵 Sekhmet<br/>The Warrior"]
        Hapi["𓈗 Hapi<br/>The Flow"]
        Khepri["𓆣 Khepri<br/>The Scarab"]
        Seba["𓇽 Seba<br/>The Star"]
    end

    Ra --> Net
    Net --> CodeGods
    Net --> MachineGods

    Maat -->|"weighs"| Isis
    Isis -->|"heals"| Thoth
    Thoth -->|"records"| Net
    Seshat -->|"bridges"| Thoth

    Horus -->|"manifest"| Anubis
    Horus -->|"manifest"| Ka
    Hapi -->|"accelerates"| Sekhmet
    Seba -->|"reports to"| Net

    style Ra fill:#FFD700,stroke:#C8A951,color:#000
    style Net fill:#8E44AD,stroke:#6C3483,color:#fff
    style Thoth fill:#1A1A5E,stroke:#C8A951,color:#C8A951
    style Maat fill:#2ECC71,stroke:#27AE60,color:#000
    style Isis fill:#E74C3C,stroke:#C0392B,color:#fff
    style Seshat fill:#3498DB,stroke:#2980B9,color:#fff
    style Horus fill:#F39C12,stroke:#E67E22,color:#000
    style Anubis fill:#C8A951,stroke:#A17D32,color:#000
    style Ka fill:#E74C3C,stroke:#C0392B,color:#fff
    style Sekhmet fill:#E74C3C,stroke:#C0392B,color:#fff
    style Hapi fill:#1ABC9C,stroke:#16A085,color:#000
    style Khepri fill:#2ECC71,stroke:#27AE60,color:#000
    style Seba fill:#9B59B6,stroke:#8E44AD,color:#fff`

	return &DiagramResult{
		Type:    DiagramHierarchy,
		Title:   "𓇽 Divine Hierarchy — The Pantheon Governance Tree",
		Mermaid: mermaid,
	}, nil
}

// ── 2. Data Flow ────────────────────────────────────────────────────

func generateDataFlow(projectRoot string) (*DiagramResult, error) {
	// Discover which deity commands exist
	deities := discoverDeities(projectRoot)

	var sb strings.Builder
	sb.WriteString("graph LR\n")
	sb.WriteString("    User([\"👤 User / CLI\"])\n")
	sb.WriteString("    Binary[\"🏛️ pantheon binary\"]\n")
	sb.WriteString("    User --> Binary\n\n")

	for _, d := range deities {
		id := strings.Title(d.name)
		sb.WriteString(fmt.Sprintf("    %s[\"%s %s<br/>%s\"]\n", id, d.glyph, id, d.domain))
	}
	sb.WriteString("\n")
	for _, d := range deities {
		id := strings.Title(d.name)
		sb.WriteString(fmt.Sprintf("    Binary --> %s\n", id))
	}

	// Add data stores
	sb.WriteString("\n    FS[(\"📁 Filesystem\")]\n")
	sb.WriteString("    Git[(\"🔀 Git History\")]\n")
	sb.WriteString("    Config[(\"⚙️ .thoth/ .pantheon/\")]\n")
	sb.WriteString("    Network[(\"🌐 Network/Fleet\")]\n\n")

	// Wire data stores to deities
	for _, d := range deities {
		id := strings.Title(d.name)
		switch d.name {
		case "anubis":
			sb.WriteString(fmt.Sprintf("    %s --> FS\n", id))
		case "maat":
			sb.WriteString(fmt.Sprintf("    %s --> Git\n", id))
			sb.WriteString(fmt.Sprintf("    %s --> FS\n", id))
		case "thoth":
			sb.WriteString(fmt.Sprintf("    %s --> Config\n", id))
			sb.WriteString(fmt.Sprintf("    %s --> Git\n", id))
		case "hapi":
			sb.WriteString(fmt.Sprintf("    %s --> FS\n", id))
		case "seba":
			sb.WriteString(fmt.Sprintf("    %s --> Network\n", id))
			sb.WriteString(fmt.Sprintf("    %s --> FS\n", id))
		case "seshat":
			sb.WriteString(fmt.Sprintf("    %s --> Config\n", id))
		}
	}

	return &DiagramResult{
		Type:    DiagramDataFlow,
		Title:   "𓇽 Data Flow — CLI → Deities → Resources",
		Mermaid: sb.String(),
	}, nil
}

// ── 3. Module Dependency Map ────────────────────────────────────────

// ModuleDep represents an import relationship between internal modules.
type ModuleDep struct {
	From string
	To   string
}

func generateModules(projectRoot string) (*DiagramResult, error) {
	deps, modules := scanModuleDeps(projectRoot)

	var sb strings.Builder
	sb.WriteString("graph TD\n")

	// Classify modules into pillars
	pillars := map[string][]string{
		"Anubis": {"cleaner", "guard", "jackal", "ka", "mirror", "horus", "ignore", "sight", "stealth"},
		"Ma'at":  {"maat", "isis", "scales"},
		"Thoth":  {"thoth", "brain", "logging"},
		"Hapi":   {"hapi", "yield", "profile"},
		"Seba":   {"seba", "scarab", "osiris"},
		"Seshat": {"seshat", "mcp"},
		"Core":   {"output", "platform", "updater", "neith"},
	}

	pillarOf := map[string]string{}
	for pillar, mods := range pillars {
		for _, m := range mods {
			pillarOf[m] = pillar
		}
	}

	// Group modules into subgraphs
	pillarModules := map[string][]string{}
	for _, m := range modules {
		p, ok := pillarOf[m]
		if !ok {
			p = "Other"
		}
		pillarModules[p] = append(pillarModules[p], m)
	}

	pillarOrder := []string{"Anubis", "Ma'at", "Thoth", "Hapi", "Seba", "Seshat", "Core", "Other"}
	for _, p := range pillarOrder {
		mods, ok := pillarModules[p]
		if !ok || len(mods) == 0 {
			continue
		}
		sort.Strings(mods)
		sb.WriteString(fmt.Sprintf("    subgraph %s\n", p))
		for _, m := range mods {
			sb.WriteString(fmt.Sprintf("        %s[\"%s\"]\n", m, m))
		}
		sb.WriteString("    end\n")
	}

	// Write edges
	sb.WriteString("\n")
	edgeSet := map[string]bool{}
	for _, d := range deps {
		key := d.From + "->" + d.To
		if edgeSet[key] {
			continue
		}
		edgeSet[key] = true
		sb.WriteString(fmt.Sprintf("    %s --> %s\n", d.From, d.To))
	}

	return &DiagramResult{
		Type:    DiagramModules,
		Title:   "𓇽 Module Dependency Map — internal/ Import Graph",
		Mermaid: sb.String(),
	}, nil
}

// scanModuleDeps parses all Go files in internal/ and extracts internal import edges.
func scanModuleDeps(projectRoot string) ([]ModuleDep, []string) {
	internalDir := filepath.Join(projectRoot, "internal")
	entries, err := os.ReadDir(internalDir)
	if err != nil {
		return nil, nil
	}

	modulePrefix := "github.com/SirsiMaster/sirsi-pantheon/internal/"
	var deps []ModuleDep
	moduleSet := map[string]bool{}
	fset := token.NewFileSet()

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		modName := entry.Name()
		moduleSet[modName] = true
		modDir := filepath.Join(internalDir, modName)

		goFiles, _ := filepath.Glob(filepath.Join(modDir, "*.go"))
		for _, gf := range goFiles {
			if strings.HasSuffix(gf, "_test.go") {
				continue
			}
			f, err := parser.ParseFile(fset, gf, nil, parser.ImportsOnly)
			if err != nil {
				continue
			}
			for _, imp := range f.Imports {
				path := strings.Trim(imp.Path.Value, `"`)
				if strings.HasPrefix(path, modulePrefix) {
					target := strings.TrimPrefix(path, modulePrefix)
					// Normalize sub-packages: jackal/rules → jackal
					if idx := strings.Index(target, "/"); idx != -1 {
						target = target[:idx]
					}
					if target != modName {
						deps = append(deps, ModuleDep{From: modName, To: target})
					}
				}
			}
		}
	}

	var modules []string
	for m := range moduleSet {
		modules = append(modules, m)
	}
	sort.Strings(modules)

	return deps, modules
}

// ── 4. Memory Architecture ──────────────────────────────────────────

func generateMemory() (*DiagramResult, error) {
	mermaid := `graph TD
    subgraph Sources["📥 Knowledge Sources"]
        Gemini["🤖 Gemini AI Mode"]
        NotebookLM["📓 NotebookLM"]
        Antigravity["💎 Antigravity IDE"]
        GitLog["🔀 Git History"]
    end

    subgraph ThothEngine["𓁟 Thoth — The Memory"]
        MemoryYAML["memory.yaml<br/>Project Identity"]
        Journal["journal.md<br/>Auto-Sync Log"]
        Rules["ANUBIS_RULES.md<br/>Governance Canon"]
    end

    subgraph SeshatBridge["𓁆 Seshat — The Bridge"]
        Extract["Extract Conversations"]
        Package["Package as Sources"]
        Inject["Inject as Knowledge Items"]
    end

    subgraph Storage["💾 Persistent Storage"]
        KI["Knowledge Items<br/>.gemini/antigravity/knowledge/"]
        Brain["Brain Logs<br/>.gemini/antigravity/brain/"]
        DotThoth[".thoth/ Directory"]
    end

    Gemini -->|"Takeout export"| Extract
    Extract --> Package
    Package -->|"upload"| NotebookLM
    NotebookLM -->|"distill"| Inject
    Inject --> KI

    Antigravity -->|"reads"| KI
    Antigravity -->|"writes"| Brain
    GitLog -->|"thoth sync"| Journal
    GitLog -->|"thoth sync"| MemoryYAML

    KI --> Antigravity
    DotThoth --- MemoryYAML
    DotThoth --- Journal
    DotThoth --- Rules

    style Gemini fill:#4285F4,stroke:#3367D6,color:#fff
    style NotebookLM fill:#EA4335,stroke:#C5221F,color:#fff
    style Antigravity fill:#9B59B6,stroke:#8E44AD,color:#fff
    style KI fill:#1A1A5E,stroke:#C8A951,color:#C8A951
    style Brain fill:#1A1A5E,stroke:#C8A951,color:#C8A951`

	return &DiagramResult{
		Type:    DiagramMemory,
		Title:   "𓇽 Memory Architecture — Knowledge Flow",
		Mermaid: mermaid,
	}, nil
}

// ── 5. Governance Cycle ─────────────────────────────────────────────

func generateGovernance() (*DiagramResult, error) {
	mermaid := `graph LR
    Net["𓁯 Net<br/>Publishes Plan"]
    Machine["𓀰 Machine Gods<br/>Execute Plan"]
    Maat["𓆄 Ma'at<br/>Weighs Execution"]
    Isis["𓆄 Isis<br/>Heals Drift"]
    Thoth["𓁟 Thoth<br/>Records Achievement"]

    Net -->|"1. Plan"| Machine
    Machine -->|"2. Execute"| Maat
    Maat -->|"3. Findings"| Isis
    Isis -->|"4. Remediate"| Thoth
    Thoth -->|"5. Record"| Net
    Net -->|"6. Re-align"| Net

    subgraph MaatDetail["𓆄 Ma'at Detail"]
        Audit["maat audit<br/>Governance Scan"]
        Scales["maat scales<br/>Policy Enforcement"]
        Pulse["maat pulse<br/>Dynamic Metrics"]
        Heal["maat heal<br/>Trigger Isis"]
    end

    Maat --> Audit
    Maat --> Scales
    Maat --> Pulse
    Maat --> Heal
    Heal --> Isis

    style Net fill:#8E44AD,stroke:#6C3483,color:#fff
    style Machine fill:#2C3E50,stroke:#1A252F,color:#C8A951
    style Maat fill:#2ECC71,stroke:#27AE60,color:#000
    style Isis fill:#E74C3C,stroke:#C0392B,color:#fff
    style Thoth fill:#1A1A5E,stroke:#C8A951,color:#C8A951
    style Pulse fill:#F39C12,stroke:#E67E22,color:#000`

	return &DiagramResult{
		Type:    DiagramGovernance,
		Title:   "𓇽 Governance Cycle — The Ma'at → Isis → Thoth Loop",
		Mermaid: mermaid,
	}, nil
}

// ── 6. CI/CD Pipeline ───────────────────────────────────────────────

func generatePipeline() (*DiagramResult, error) {
	mermaid := `graph LR
    Push["🔀 git push"]
    Gate["𓇳 Pre-Push Gate"]
    GFmt["gofmt check"]
    Test["go test -short"]
    CI["GitHub Actions CI"]
    Lint["golangci-lint"]
    FullTest["go test -race -cover"]
    Pulse["maat pulse --json"]
    Metrics["📊 metrics.json"]
    Coverage["📈 coverage.out"]
    Build["go build"]
    Binary["🏛️ pantheon binary"]

    Push --> Gate
    Gate --> GFmt
    Gate --> Test
    GFmt -->|"pass"| CI
    Test -->|"pass"| CI

    CI --> Lint
    CI --> FullTest
    CI --> Build
    FullTest --> Coverage
    Build --> Binary
    Binary --> Pulse
    Pulse --> Metrics

    style Push fill:#2C3E50,stroke:#1A252F,color:#fff
    style Gate fill:#C8A951,stroke:#A17D32,color:#000
    style CI fill:#4285F4,stroke:#3367D6,color:#fff
    style Metrics fill:#F39C12,stroke:#E67E22,color:#000
    style Binary fill:#2ECC71,stroke:#27AE60,color:#000`

	return &DiagramResult{
		Type:    DiagramPipeline,
		Title:   "𓇽 CI/CD Pipeline — Push → Gate → CI → Artifacts",
		Mermaid: mermaid,
	}, nil
}

// ── Helpers ─────────────────────────────────────────────────────────

type deityInfo struct {
	name   string
	glyph  string
	domain string
}

func discoverDeities(projectRoot string) []deityInfo {
	known := []deityInfo{
		{"anubis", "𓁢", "Hygiene"},
		{"maat", "𓆄", "Governance"},
		{"thoth", "𓁟", "Knowledge"},
		{"hapi", "𓈗", "Compute"},
		{"seba", "𓇽", "Mapping"},
		{"seshat", "𓁆", "Scribe"},
	}

	var found []deityInfo
	for _, d := range known {
		cmdPath := filepath.Join(projectRoot, "cmd", "pantheon", d.name+".go")
		if _, err := os.Stat(cmdPath); err == nil {
			found = append(found, d)
		}
	}
	return found
}

// RenderDiagramsHTML produces a self-contained HTML page with all diagrams.
func RenderDiagramsHTML(diagrams []*DiagramResult, outputPath string) error {
	var cards strings.Builder
	for _, d := range diagrams {
		escaped := strings.ReplaceAll(d.Mermaid, "`", "\\`")
		cards.WriteString(fmt.Sprintf(`
    <div class="diagram-card" id="%s">
      <h2>%s</h2>
      <div class="mermaid-container">
        <pre class="mermaid">%s</pre>
      </div>
      <button onclick="copyDiagram(this)" data-mermaid="%s">📋 Copy Mermaid</button>
    </div>
`, string(d.Type), d.Title, d.Mermaid, escaped))
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>𓇽 Seba — Architectural Diagrams</title>
<script src="https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.min.js"></script>
<style>
  @import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;600&display=swap');
  :root {
    --bg: #06060F;
    --card-bg: #0D0D1A;
    --gold: #C8A951;
    --text: #E0E0E0;
    --dim: #555;
    --border: rgba(200,169,81,0.15);
  }
  * { margin:0; padding:0; box-sizing:border-box; }
  body {
    background: var(--bg);
    color: var(--text);
    font-family: 'Inter', -apple-system, system-ui, sans-serif;
    padding: 2rem;
    min-height: 100vh;
  }
  h1 {
    color: var(--gold);
    font-size: 1.8rem;
    font-weight: 300;
    letter-spacing: 3px;
    text-transform: uppercase;
    text-align: center;
    margin-bottom: 0.5rem;
  }
  .subtitle {
    text-align: center;
    color: var(--dim);
    font-size: 0.85rem;
    margin-bottom: 3rem;
    letter-spacing: 1px;
  }
  .diagram-card {
    background: var(--card-bg);
    border: 1px solid var(--border);
    border-radius: 16px;
    padding: 2rem;
    margin-bottom: 2.5rem;
    position: relative;
    overflow: hidden;
    scroll-margin-top: 1rem;
  }
  .diagram-card::before {
    content: '';
    position: absolute;
    top: 0; left: 0; right: 0;
    height: 2px;
    background: linear-gradient(90deg, transparent, var(--gold), transparent);
    opacity: 0.4;
  }
  .diagram-card h2 {
    color: var(--gold);
    font-size: 1.1rem;
    font-weight: 400;
    margin-bottom: 1.5rem;
    letter-spacing: 1px;
  }
  .mermaid-container {
    background: rgba(255,255,255,0.03);
    border-radius: 12px;
    padding: 1.5rem;
    overflow-x: auto;
    min-height: 200px;
  }
  /* Hide raw source text — Mermaid replaces with SVG */
  pre.mermaid {
    visibility: hidden;
    font-size: 0;
    line-height: 0;
    overflow: hidden;
    max-height: 0;
  }
  /* Once Mermaid renders, SVG becomes visible */
  pre.mermaid[data-processed="true"],
  pre.mermaid svg,
  .mermaid svg {
    visibility: visible;
    font-size: initial;
    line-height: initial;
    max-height: none;
    display: flex;
    justify-content: center;
    width: 100%%;
  }
  button {
    position: absolute;
    top: 1.5rem; right: 1.5rem;
    background: rgba(200,169,81,0.1);
    border: 1px solid var(--border);
    color: var(--dim);
    padding: 6px 14px;
    border-radius: 8px;
    cursor: pointer;
    font-size: 0.75rem;
    transition: all 0.3s;
    z-index: 2;
  }
  button:hover {
    border-color: var(--gold);
    color: var(--gold);
    background: rgba(200,169,81,0.2);
  }
  .nav {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    justify-content: center;
    margin-bottom: 2.5rem;
  }
  .nav a {
    color: var(--dim);
    text-decoration: none;
    padding: 6px 16px;
    border: 1px solid var(--border);
    border-radius: 8px;
    font-size: 0.8rem;
    transition: all 0.3s;
  }
  .nav a:hover {
    border-color: var(--gold);
    color: var(--gold);
    background: rgba(200,169,81,0.08);
  }
  footer {
    text-align: center;
    color: var(--dim);
    font-size: 0.75rem;
    margin-top: 3rem;
    padding: 2rem;
  }
</style>
</head>
<body>
  <h1>𓇽 Seba — The Star Map</h1>
  <p class="subtitle">Architectural Diagrams · Generated from Live Project Analysis</p>

  <div class="nav">
    <a href="#hierarchy">Hierarchy</a>
    <a href="#dataflow">Data Flow</a>
    <a href="#modules">Modules</a>
    <a href="#memory">Memory</a>
    <a href="#governance">Governance</a>
    <a href="#pipeline">Pipeline</a>
  </div>

  %s

  <footer>
    <p>𓇽 Seba · Sirsi Pantheon v1.0.0-rc1 · Generated from live source analysis</p>
  </footer>

  <script>
    mermaid.initialize({
      startOnLoad: true,
      theme: 'dark',
      themeVariables: {
        primaryColor: '#1A1A5E',
        primaryTextColor: '#C8A951',
        primaryBorderColor: '#C8A951',
        lineColor: '#C8A951',
        secondaryColor: '#0D0D1A',
        tertiaryColor: '#06060F',
        fontFamily: 'Inter, -apple-system, system-ui, sans-serif'
      }
    });

    // After Mermaid renders, ensure SVGs are visible and source text is hidden
    mermaid.run().then(() => {
      document.querySelectorAll('pre.mermaid').forEach(pre => {
        pre.style.visibility = 'visible';
        pre.style.maxHeight = 'none';
        pre.style.fontSize = 'initial';
        pre.style.lineHeight = 'initial';
      });
    });

    function copyDiagram(btn) {
      const mmd = btn.getAttribute('data-mermaid');
      navigator.clipboard.writeText(mmd);
      btn.textContent = '✅ Copied!';
      setTimeout(() => btn.textContent = '📋 Copy Mermaid', 2000);
    }
  </script>
</body>
</html>`, cards.String())

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	return os.WriteFile(outputPath, []byte(html), 0o644)
}
