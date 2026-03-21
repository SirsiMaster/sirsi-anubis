# 🧠 Brain Module — Neural Classification Engine

**Package:** `internal/brain`
**Tier:** Anubis Pro (Neural Edition)
**Status:** Scaffold complete, awaiting trained model

---

## Architecture

```
brain/
├── downloader.go       — Model fetcher, manifest, checksum, version management
├── downloader_test.go  — 9 tests (manifest, hash, bytes, state)
├── inference.go        — Classifier interface + StubClassifier
└── inference_test.go   — 13 tests (35+ file type classifications, batch)
```

### Classifier Interface

```go
type Classifier interface {
    Name() string
    Load(weightsDir string) error
    Classify(filePath string) (*Classification, error)
    ClassifyBatch(filePaths []string, workers int) (*BatchResult, error)
    Close() error
}
```

### Backends

| Backend | Status | Platform | Notes |
|:--------|:-------|:---------|:------|
| `StubClassifier` | ✅ Shipping | All | File extension/path heuristics |
| `ONNXClassifier` | 📋 Planned | All (CPU) | `ort-go` bindings |
| `CoreMLClassifier` | 📋 Planned | macOS | CGO bridge, Neural Engine |

### File Classification Categories

| Class | Description | Example Extensions |
|:------|:-----------|:------------------|
| `junk` | Temporary, cache, build artifacts | `.log`, `.tmp`, `.bak`, `.pyc` |
| `essential` | System/app critical | — (ML-only) |
| `project` | Source code, documentation | `.go`, `.py`, `.js`, `.rs` |
| `model` | ML model weights | `.onnx`, `.pt`, `.safetensors` |
| `data` | Datasets, databases | `.csv`, `.sqlite`, `.parquet` |
| `media` | Images, video, audio | `.jpg`, `.mp4`, `.mp3` |
| `archive` | Compressed archives | `.zip`, `.tar`, `.dmg` |
| `config` | Configuration files | `.yaml`, `.json`, `.toml` |
| `unknown` | Unclassified | — |

---

## Downloader

### Manifest-Driven Architecture

The brain module uses a two-manifest system:

1. **Remote Manifest** (`brain-manifest.json` in repo root)
   - Lists available models with versions, checksums, download URLs
   - Fetched from GitHub on `install-brain` / `--update`

2. **Local Manifest** (`~/.anubis/weights/manifest.json`)
   - Tracks installed model, version, format, install date
   - Used for update detection and status display

### Download Flow

```
1. Fetch remote manifest from GitHub
2. Select platform-appropriate model (CoreML on Apple Silicon, ONNX otherwise)
3. Download with streaming progress (32KB buffer)
4. Verify SHA-256 checksum
5. Write local manifest
6. Report success
```

### Safety

- **Checksum verification**: SHA-256 post-download
- **Partial download cleanup**: Removes file on download failure
- **Timeout**: 5 min download, 10s manifest fetch
- **No telemetry**: Only contacts public GitHub raw content API
- **Self-deletable**: `--remove` and `--stealth` clean up completely

---

## CLI Commands

```bash
anubis install-brain             # Install default model
anubis install-brain --update    # Check for and install latest
anubis install-brain --remove    # Delete weights
anubis uninstall-brain           # Alias for --remove
```

All commands support `--json` and `--quiet` global flags.

---

## Adding a New Backend

1. Implement the `Classifier` interface
2. Add platform detection in `GetClassifier()`
3. Update `selectPlatformModel()` for format preference
4. Add build tags if CGO required (e.g., `//go:build darwin`)

---

## Size Budget

| Component | Budget | Notes |
|:----------|:-------|:------|
| Brain module code | < 100 KB | Pure Go, no external deps |
| Model weights | < 100 MB | Downloaded on demand, not in binary |
| Binary impact | ~0 | No weight increase to base `anubis` |
