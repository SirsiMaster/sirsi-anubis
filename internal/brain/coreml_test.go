package brain

import (
	"os"
	"path/filepath"
	"testing"
)

// ── CoreMLClassifier tests ─────────────────────────────────────────
// These test the CoreMLClassifier struct and its fallback behavior.
// On non-macOS or non-CGO builds, coremlAvailable() returns false,
// so the classifier falls back to SpotlightClassifier/StubClassifier.

func TestNewCoreMLClassifier(t *testing.T) {
	t.Parallel()
	c := NewCoreMLClassifier()
	if c == nil {
		t.Fatal("NewCoreMLClassifier() returned nil")
	}
	if c.fallback == nil {
		t.Fatal("fallback should be initialized")
	}
}

func TestCoreMLClassifier_Name_NoModel(t *testing.T) {
	t.Parallel()
	c := NewCoreMLClassifier()
	_ = c.Load("")
	// Without a loaded CoreML model, should fall back
	name := c.Name()
	if name == "" {
		t.Error("Name() should not be empty")
	}
	// If CoreML is unavailable, it should return the fallback name
	if !coremlAvailable() && name == "coreml-ane" {
		t.Error("should not report coreml-ane when CoreML is unavailable")
	}
}

func TestCoreMLClassifier_Load_EmptyDir(t *testing.T) {
	t.Parallel()
	c := NewCoreMLClassifier()
	err := c.Load("")
	if err != nil {
		t.Fatalf("Load('') should not error: %v", err)
	}
	// Without a model directory, loaded should remain false
	if !coremlAvailable() && c.loaded {
		t.Error("loaded should be false without CoreML")
	}
}

func TestCoreMLClassifier_Load_NoModelFile(t *testing.T) {
	t.Parallel()
	c := NewCoreMLClassifier()
	dir := t.TempDir()
	err := c.Load(dir)
	if err != nil {
		t.Fatalf("Load should not error: %v", err)
	}
	// No classifier.mlmodelc in dir, so loaded should be false
	if c.loaded && !coremlAvailable() {
		t.Error("loaded should be false without CoreML availability")
	}
}

func TestCoreMLClassifier_Load_WithModelFile(t *testing.T) {
	t.Parallel()
	c := NewCoreMLClassifier()
	dir := t.TempDir()

	// Create a fake .mlmodelc directory
	modelDir := filepath.Join(dir, "classifier.mlmodelc")
	if err := os.MkdirAll(modelDir, 0o755); err != nil {
		t.Fatal(err)
	}

	err := c.Load(dir)
	if err != nil {
		t.Fatalf("Load should not error: %v", err)
	}

	// On macOS with CGO, this would load; on other platforms, it silently falls back
	if coremlAvailable() {
		if !c.loaded {
			t.Error("loaded should be true when CoreML is available and model exists")
		}
		if c.modelPath != modelDir {
			t.Errorf("modelPath = %q, want %q", c.modelPath, modelDir)
		}
	}
}

func TestCoreMLClassifier_Classify_Fallback(t *testing.T) {
	t.Parallel()
	c := NewCoreMLClassifier()
	_ = c.Load("")

	// Without a loaded model, should fall back to the spotlight/stub classifier
	result, err := c.Classify("/test/main.go")
	if err != nil {
		t.Fatalf("Classify error: %v", err)
	}
	if result.Class != ClassProject {
		t.Errorf("Class = %q, want %q (fallback should classify .go as project)", result.Class, ClassProject)
	}
}

func TestCoreMLClassifier_ClassifyBatch_Fallback(t *testing.T) {
	t.Parallel()
	c := NewCoreMLClassifier()
	_ = c.Load("")

	files := []string{"/test/main.go", "/tmp/debug.log"}
	result, err := c.ClassifyBatch(files, 2)
	if err != nil {
		t.Fatalf("ClassifyBatch error: %v", err)
	}
	if result.FilesProcessed != 2 {
		t.Errorf("FilesProcessed = %d, want 2", result.FilesProcessed)
	}
}

func TestCoreMLClassifier_Close(t *testing.T) {
	t.Parallel()
	c := NewCoreMLClassifier()
	_ = c.Load("")
	err := c.Close()
	if err != nil {
		t.Fatalf("Close error: %v", err)
	}
	if c.loaded {
		t.Error("loaded should be false after Close")
	}
}

// ── mapCoreMLLabel tests ───────────────────────────────────────────

func TestMapCoreMLLabel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		label    string
		expected FileClass
	}{
		// Junk variants
		{"junk", ClassJunk},
		{"cache", ClassJunk},
		{"temp", ClassJunk},
		{"build_artifact", ClassJunk},

		// Essential variants
		{"essential", ClassEssential},
		{"system", ClassEssential},
		{"critical", ClassEssential},

		// Project variants
		{"project", ClassProject},
		{"source", ClassProject},
		{"documentation", ClassProject},

		// Model variants
		{"model", ClassModel},
		{"weights", ClassModel},
		{"checkpoint", ClassModel},

		// Data variants
		{"data", ClassData},
		{"dataset", ClassData},
		{"database", ClassData},

		// Media variants
		{"media", ClassMedia},
		{"image", ClassMedia},
		{"video", ClassMedia},
		{"audio", ClassMedia},

		// Archive variants
		{"archive", ClassArchive},
		{"compressed", ClassArchive},

		// Config variants
		{"config", ClassConfig},
		{"configuration", ClassConfig},
		{"settings", ClassConfig},

		// Unknown
		{"something_else", ClassUnknown},
		{"", ClassUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			got := mapCoreMLLabel(tt.label)
			if got != tt.expected {
				t.Errorf("mapCoreMLLabel(%q) = %q, want %q", tt.label, got, tt.expected)
			}
		})
	}
}

// ── mapContentType tests ───────────────────────────────────────────

func TestMapContentType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		uti      string
		expected FileClass
	}{
		// Source code UTIs
		{"public.source-code", ClassProject},
		{"public.c-source", ClassProject},
		{"public.swift-source", ClassProject},
		{"public.python-script", ClassProject},
		{"public.shell-script", ClassProject},
		{"com.sun.java-source", ClassProject},

		// Text/docs
		{"public.plain-text", ClassProject},
		{"public.html", ClassProject},
		{"net.daringfireball.markdown", ClassProject},
		{"public.json", ClassProject},

		// Config
		{"public.yaml", ClassConfig},
		{"com.apple.property-list", ClassConfig},

		// Images
		{"public.image", ClassMedia},
		{"public.jpeg", ClassMedia},
		{"public.png", ClassMedia},
		{"public.heic", ClassMedia},

		// Audio/Video
		{"public.audio", ClassMedia},
		{"public.video", ClassMedia},
		{"public.mp3", ClassMedia},
		{"public.mpeg-4", ClassMedia},

		// Archives
		{"public.zip-archive", ClassArchive},
		{"org.gnu.gnu-tar-archive", ClassArchive},
		{"com.apple.disk-image", ClassArchive},

		// ML models
		{"com.apple.coreml.model", ClassModel},

		// Data
		{"public.database", ClassData},
		{"public.comma-separated-values-text", ClassData},

		// Executables
		{"public.executable", ClassEssential},
		{"public.unix-executable", ClassEssential},
		{"com.apple.mach-o-binary", ClassEssential},

		// Junk
		{"public.cache", ClassJunk},
		{"public.log", ClassJunk},

		// Unknown
		{"com.unknown.type", ClassUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.uti, func(t *testing.T) {
			class, confidence := mapContentType(tt.uti)
			if class != tt.expected {
				t.Errorf("mapContentType(%q) = %q, want %q", tt.uti, class, tt.expected)
			}
			if class != ClassUnknown && confidence <= 0 {
				t.Errorf("confidence should be > 0 for known type %q", tt.uti)
			}
		})
	}
}

// ── SpotlightClassifier tests ──────────────────────────────────────

func TestNewSpotlightClassifier(t *testing.T) {
	t.Parallel()
	c := NewSpotlightClassifier()
	if c == nil {
		t.Fatal("NewSpotlightClassifier() returned nil")
	}
	if c.stub == nil {
		t.Fatal("stub should be initialized")
	}
}

func TestSpotlightClassifier_LoadAndClose(t *testing.T) {
	t.Parallel()
	c := NewSpotlightClassifier()
	if err := c.Load(""); err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if err := c.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
}

func TestSpotlightClassifier_Classify(t *testing.T) {
	t.Parallel()
	c := NewSpotlightClassifier()
	_ = c.Load("")
	defer c.Close()

	result, err := c.Classify("/test/main.go")
	if err != nil {
		t.Fatalf("Classify error: %v", err)
	}
	if result.Class != ClassProject {
		t.Errorf("Class = %q, want %q", result.Class, ClassProject)
	}
}

func TestSpotlightClassifier_ClassifyBatch(t *testing.T) {
	t.Parallel()
	c := NewSpotlightClassifier()
	_ = c.Load("")
	defer c.Close()

	result, err := c.ClassifyBatch([]string{"/test.go", "/tmp/file.log"}, 2)
	if err != nil {
		t.Fatalf("ClassifyBatch error: %v", err)
	}
	if result.FilesProcessed < 2 {
		t.Errorf("FilesProcessed = %d, want >= 2", result.FilesProcessed)
	}
}
