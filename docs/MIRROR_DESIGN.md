# 𓂀 Mirror — Semantic File Deduplication & Importance Ranking

> **Status**: Design Document — v0.1  
> **Filed**: March 21, 2026  
> **Module**: `internal/mirror/`  
> **Command**: `anubis mirror`  
> **Theme**: Egyptian copper mirrors — tools of truth that reveal what is real and what is reflection

---

## The Problem

People accumulate thousands of duplicate files — photos, music, documents — across
Downloads, Desktop, iCloud, external drives, and app exports. Existing dedup tools
are dumb: they find exact matches and ask you to pick one. But **which one matters?**

The photo in your Camera Roll that's tagged with faces, GPS, and referenced by 3 albums
is not the same as the WhatsApp-compressed copy sitting in Downloads. They're byte-different
but semantically identical — and the user needs to keep the *right* one.

Nobody solves this well because it requires:
1. Understanding file content (not just hashes)  
2. Understanding file context (where it lives, what references it)  
3. Presenting relationships visually so users can make informed decisions

## The Insight

Apple's Neural Engine (ANE) on M-series Macs and A-series iPhones does
15+ trillion operations per second. CoreML models run on-device, privately,
with zero cloud dependency.

Anubis already has:
- **Brain module** — downloads/manages CoreML + ONNX models from GitHub Releases
- **Classifier interface** — `Classify()` and `ClassifyBatch()` with worker pool
- **Seba graph** — kinetic infrastructure visualization with force-directed layout
- **Ka scanner** — detects ghost app remnants (similar pattern: scan → classify → report)
- **Rule A11** — no telemetry, everything stays on-device

We just need to connect the dots.

---

## Product Model

### Core Principle: One Engine, Two Interfaces

The **interface determines who uses which features, not what features exist.**
GUI and CLI are equal citizens. Every free-tier feature is accessible from both.
Every pro-tier feature is accessible from both. The user picks their comfort zone.

```
┌──────────────────────────────────────────────────────┐
│                    anubis mirror                      │
│                                                       │
│   ┌───────────────────────────────────────────────┐   │
│   │             Shared Engine (Go)                 │   │
│   │  scanner • hasher • ranker • dedup • cleaner  │   │
│   └───────────────────┬───────────────────────────┘   │
│                       │                               │
│         ┌─────────────┼─────────────┐                 │
│         ▼                           ▼                 │
│   ┌───────────┐               ┌───────────┐           │
│   │    GUI    │               │    CLI    │           │
│   │  Browser  │               │ Terminal  │           │
│   │           │               │           │           │
│   │ Your      │               │ Devs &    │           │
│   │ friend    │               │ sysadmins │           │
│   └───────────┘               └───────────┘           │
└──────────────────────────────────────────────────────┘
```

### 🆓 Ankh (Free) — Full Deduplication Engine

**One engine. Same features. Pick your interface.**

| Feature | GUI | CLI |
|:--------|:---:|:---:|
| **Exact match (SHA-256)** | ✅ | ✅ |
| **Perceptual hash (images)** | ✅ | ✅ |
| **Audio fingerprint (music)** | ✅ | ✅ |
| **Size pre-filter** | ✅ | ✅ |
| **Media type filters** | ✅ Filter chips | ✅ `--photos`, `--music` |
| **Min/max size filter** | ✅ Slider | ✅ `--min-size`, `--max-size` |
| **Protected directories** | ✅ Drag to protect | ✅ `--protect ~/Originals` |
| **Smart recommendations** | ✅ ✓ Keep / ✗ Remove | ✅ Same in terminal |
| **Dry run (default)** | ✅ Preview only | ✅ `--dry-run` |
| **JSON export** | ✅ Download button | ✅ `--json` |

**GUI** (for your friend):
```
anubis mirror                    # Opens browser → drag folders → done
```

**CLI** (for devs/automation):
```
anubis mirror ~/Photos ~/Downloads        # Scan specific dirs
anubis mirror --photos ~/Pictures          # Photo-specific scan
anubis mirror --music ~/Music              # Music-specific scan
anubis mirror --min-size 1MB               # Skip small files
anubis mirror --protect ~/Originals        # Lock important dirs
```

### 👁️ Eye of Horus (Pro) — Neural Importance Ranking

**Same principle: both interfaces get the full pro feature set.**

| Feature | GUI | CLI |
|:--------|:---:|:---:|
| **Face detection (ANE)** | ✅ Face badges on photos | ✅ `--protect-faces` |
| **Scene classification** | ✅ Scene labels in UI | ✅ Scene tags in output |
| **Importance scoring** | ✅ Visual importance bar | ✅ `--rank` |
| **Metadata analysis** | ✅ EXIF/GPS indicators | ✅ In JSON output |
| **Knowledge graph** | ✅ Interactive Seba view | ✅ `--graph` |
| **Smart auto-select** | ✅ One-click cleanup | ✅ `--clean --confirm` |

**GUI** (pro features appear as visual upgrades):
- Photos show face badges and scene labels
- Importance bar next to each file (amber → gold gradient)
- "View Relationships" button opens Seba knowledge graph
- One-click "Smart Clean" selects lowest-importance duplicates

**CLI** (pro features appear as flags):
```
anubis mirror --rank ~/Photos              # Add importance scoring
anubis mirror --graph ~/Photos             # Generate knowledge graph
anubis mirror --clean --confirm            # Auto-remove lowest-ranked
anubis mirror --protect-faces              # Never delete photos with faces
```

**ANE/CoreML Models** (downloaded via `anubis install-brain`):
- `mirror-vision-v1.mlmodelc` — face detection + scene classification
- `mirror-audio-v1.mlmodelc` — audio fingerprinting + genre classification
- `mirror-embeddings-v1.onnx` — file embedding model for semantic similarity

---

## Architecture

```
┌─────────────────────────────────────────────────┐
│                anubis mirror                     │
│                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────────┐  │
│  │ Scanner  │  │ Hasher   │  │ Classifier   │  │
│  │          │  │          │  │ (Brain)      │  │
│  │ Walk dirs│→ │ SHA-256  │→ │ CoreML/ONNX  │  │
│  │ Filter   │  │ pHash    │  │ Face detect  │  │
│  │ Group    │  │ AudioFP  │  │ Scene class  │  │
│  └──────────┘  └──────────┘  └──────────────┘  │
│       │              │              │            │
│       └──────────────┼──────────────┘            │
│                      ▼                           │
│              ┌──────────────┐                    │
│              │  Ranker      │                    │
│              │  Importance  │                    │
│              │  scoring     │                    │
│              └──────┬───────┘                    │
│                     │                            │
│         ┌───────────┼───────────┐                │
│         ▼           ▼           ▼                │
│    ┌─────────┐ ┌─────────┐ ┌─────────┐          │
│    │ Report  │ │ Graph   │ │ Cleaner │          │
│    │ (CLI)   │ │ (Seba)  │ │ (safe)  │          │
│    └─────────┘ └─────────┘ └─────────┘          │
└─────────────────────────────────────────────────┘
```

### Module: `internal/mirror/`

```
internal/mirror/
├── scanner.go       # Directory walker, file grouping, size pre-filter
├── hasher.go        # SHA-256, perceptual hash (pHash), audio fingerprint
├── ranker.go        # Importance scoring engine
├── dedup.go         # Duplicate group management, selection logic
├── types.go         # Core types: DuplicateGroup, FileEntry, ImportanceScore
└── mirror_test.go   # Tests
```

### Key Types

```go
type FileEntry struct {
    Path         string
    Size         int64
    ModTime      time.Time
    SHA256       string
    PHash        string   // perceptual hash (images)
    AudioFP      string   // audio fingerprint (music)
    Importance   float64  // 0.0-1.0 (pro tier)
    HasFaces     bool     // CoreML face detection (pro)
    SceneType    string   // "landscape", "portrait", "document" (pro)
    IsProtected  bool     // In a safe-list directory
    References   int      // Number of apps/libraries referencing this file
    MediaType    string   // "photo", "music", "video", "document", "other"
}

type DuplicateGroup struct {
    ID           string
    Files        []FileEntry
    MatchType    string   // "exact", "perceptual", "audio", "semantic"
    Recommended  int      // Index of file to keep
    Confidence   float64  // How confident is the recommendation
    TotalWaste   int64    // Bytes recoverable by removing duplicates
}

type MirrorResult struct {
    Groups         []DuplicateGroup
    TotalFiles     int
    TotalDuplicates int
    TotalWasteBytes int64
    ScanDuration   time.Duration
    ModelUsed      string  // "hash-only" or brain model name
}
```

---

## Importance Scoring (Pro)

The importance score is a weighted composite:

| Signal | Weight | Description |
|:-------|:-------|:------------|
| **Has faces** | 0.25 | Photos with detected faces are important |
| **Has GPS/EXIF** | 0.10 | Rich metadata = original source |
| **Album membership** | 0.15 | Referenced by Photos.app albums |
| **File age** | 0.05 | Older = more likely original |
| **File size** | 0.10 | Larger = less compressed = higher quality |
| **Directory depth** | 0.05 | Shallow = intentionally placed |
| **Protected dir** | 0.15 | In user's safe list |
| **Reference count** | 0.10 | Other files/apps point to this |
| **Finder tags** | 0.05 | User has manually tagged this file |

Score = Σ(signal × weight), normalized to [0.0, 1.0]

The file with the **highest importance score** in a duplicate group is the keeper.

---

## Implementation Phases

### Phase 1: Hash Scanner (Ship Now — Free Tier)
- [ ] `internal/mirror/scanner.go` — parallel directory walker
- [ ] `internal/mirror/hasher.go` — SHA-256 + size-based grouping
- [ ] `internal/mirror/types.go` — core types
- [ ] `internal/mirror/dedup.go` — duplicate group management
- [ ] `cmd/anubis/mirror.go` — CLI command
- [ ] Tests

### Phase 2: Perceptual Hashing (Free Tier Enhancement)
- [ ] pHash implementation for images (DCT-based)
- [ ] Hamming distance threshold for "similar enough"
- [ ] Audio fingerprinting (Chromaprint-compatible)
- [ ] `--photos` and `--music` flags

### Phase 3: Importance Ranking (Pro Tier)
- [ ] `internal/mirror/ranker.go` — scoring engine
- [ ] EXIF/metadata extraction
- [ ] Directory protection / safe lists
- [ ] Reference counting (Spotlight metadata)
- [ ] `--rank` flag

### Phase 4: Neural Classification (Pro Tier — ANE)
- [ ] CoreML face detection model
- [ ] Scene classification model
- [ ] Integration with Brain module download pipeline
- [ ] `--protect-faces` flag

### Phase 5: Knowledge Graph (Pro Tier — Seba)
- [ ] File relationship graph generation
- [ ] "Why are these duplicates?" edge labels
- [ ] Importance visualization (node size = importance)
- [ ] Interactive deletion from graph UI
- [ ] `--graph` flag

---

## Competitive Landscape

| Tool | Exact Dedup | Perceptual | Importance | On-Device ML | Knowledge Graph |
|:-----|:----------:|:----------:|:----------:|:------------:|:---------------:|
| **fdupes** | ✅ | ❌ | ❌ | ❌ | ❌ |
| **rdfind** | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Gemini 2** | ✅ | ✅ | ❌ | ❌ | ❌ |
| **dupeGuru** | ✅ | ✅ | ❌ | ❌ | ❌ |
| **CleanMyMac** | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Anubis Mirror** | ✅ | ✅ | ✅ | ✅ (ANE) | ✅ (Seba) |

**The moat**: Nobody else combines deduplication + on-device neural importance ranking +
knowledge graph visualization. And it's open source (free tier) with a premium upgrade path.

---

## Revenue Model

- **Free (Ankh)**: Full dedup engine via GUI + CLI — open source, forever free.
  Both interfaces have identical feature parity. The interface determines
  the user, not the capability.
- **Pro ($9/mo or $79/yr)**: ANE neural features (face detection, importance
  ranking, knowledge graph) — accessible from both GUI and CLI.
  GUI shows visual upgrades (face badges, importance bars, Seba graph view).
  CLI shows equivalent data via flags and JSON output.
- **Enterprise**: Fleet dedup across teams, policy enforcement via Scales.
