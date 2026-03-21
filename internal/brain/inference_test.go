package brain

import (
	"testing"
)

func TestStubClassifier_Name(t *testing.T) {
	c := NewStubClassifier()
	if c.Name() != "stub-heuristic-v1" {
		t.Errorf("Name() = %q, want %q", c.Name(), "stub-heuristic-v1")
	}
}

func TestStubClassifier_LoadAndClose(t *testing.T) {
	c := NewStubClassifier()

	// Should error before load
	_, err := c.Classify("/some/file.go")
	if err == nil {
		t.Error("Classify before Load should error")
	}

	// Load
	if err := c.Load(""); err != nil {
		t.Fatalf("Load error: %v", err)
	}

	// Close
	if err := c.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}

	// Should error after close
	_, err = c.Classify("/some/file.go")
	if err == nil {
		t.Error("Classify after Close should error")
	}
}

func TestStubClassifier_Classify(t *testing.T) {
	c := NewStubClassifier()
	_ = c.Load("")
	defer c.Close()

	tests := []struct {
		path     string
		expected FileClass
	}{
		// Junk
		{"/tmp/debug.log", ClassJunk},
		{"/var/temp.tmp", ClassJunk},
		{"/home/user/backup.bak", ClassJunk},
		{"/path/.DS_Store", ClassJunk},

		// Source/Project
		{"/project/main.go", ClassProject},
		{"/src/app.py", ClassProject},
		{"/web/index.js", ClassProject},
		{"/lib/utils.rs", ClassProject},
		{"/code/server.ts", ClassProject},

		// Config
		{"/etc/config.yaml", ClassConfig},
		{"/app/settings.json", ClassConfig},
		{"/home/.bashrc.toml", ClassConfig},

		// Media
		{"/photos/vacation.jpg", ClassMedia},
		{"/video/demo.mp4", ClassMedia},
		{"/music/track.mp3", ClassMedia},

		// Archives
		{"/downloads/package.zip", ClassArchive},
		{"/backup/full.tar", ClassArchive},
		{"/installer/app.dmg", ClassArchive},

		// Data
		{"/data/export.csv", ClassData},
		{"/db/main.sqlite", ClassData},

		// Model weights
		{"/models/bert.onnx", ClassModel},
		{"/weights/model.pt", ClassModel},
		{"/ml/checkpoint.safetensors", ClassModel},

		// Path-based junk
		{"/project/node_modules/lodash/index.js", ClassJunk},
		{"/src/__pycache__/module.pyc", ClassJunk},

		// Filename-based project
		{"/repo/README.md", ClassProject},
		{"/repo/LICENSE", ClassProject},
		{"/repo/Dockerfile", ClassProject},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result, err := c.Classify(tt.path)
			if err != nil {
				t.Fatalf("Classify(%q) error: %v", tt.path, err)
			}
			if result.Class != tt.expected {
				t.Errorf("Classify(%q) = %q, want %q (confidence: %.2f)",
					tt.path, result.Class, tt.expected, result.Confidence)
			}
			if result.Confidence < 0 || result.Confidence > 1 {
				t.Errorf("Confidence out of range [0,1]: %f", result.Confidence)
			}
			if result.ModelUsed != c.Name() {
				t.Errorf("ModelUsed = %q, want %q", result.ModelUsed, c.Name())
			}
		})
	}
}

func TestStubClassifier_ClassifyBatch(t *testing.T) {
	c := NewStubClassifier()
	_ = c.Load("")
	defer c.Close()

	files := []string{
		"/project/main.go",
		"/tmp/debug.log",
		"/data/export.csv",
		"/photos/cat.jpg",
		"/models/bert.onnx",
	}

	result, err := c.ClassifyBatch(files, 2)
	if err != nil {
		t.Fatalf("ClassifyBatch error: %v", err)
	}

	if result.FilesProcessed != len(files) {
		t.Errorf("FilesProcessed = %d, want %d", result.FilesProcessed, len(files))
	}
	if result.FilesSkipped != 0 {
		t.Errorf("FilesSkipped = %d, want 0", result.FilesSkipped)
	}
	if len(result.Classifications) != len(files) {
		t.Errorf("Classifications count = %d, want %d", len(result.Classifications), len(files))
	}
	if result.ModelUsed != c.Name() {
		t.Errorf("ModelUsed = %q, want %q", result.ModelUsed, c.Name())
	}
}

func TestStubClassifier_ClassifyBatch_BeforeLoad(t *testing.T) {
	c := NewStubClassifier()
	_, err := c.ClassifyBatch([]string{"/file.go"}, 1)
	if err == nil {
		t.Error("ClassifyBatch before Load should error")
	}
}

func TestStubClassifier_ClassifyBatch_DefaultWorkers(t *testing.T) {
	c := NewStubClassifier()
	_ = c.Load("")
	defer c.Close()

	// workers <= 0 should default to 4
	result, err := c.ClassifyBatch([]string{"/test.go"}, 0)
	if err != nil {
		t.Fatalf("ClassifyBatch error: %v", err)
	}
	if result.FilesProcessed != 1 {
		t.Errorf("FilesProcessed = %d, want 1", result.FilesProcessed)
	}
}

func TestClassifyByHeuristic_Unknown(t *testing.T) {
	class, confidence := classifyByHeuristic("/some/random/file")
	if class != ClassUnknown {
		t.Errorf("Expected ClassUnknown for extensionless file, got %q", class)
	}
	if confidence != 0.0 {
		t.Errorf("Expected 0.0 confidence for unknown, got %f", confidence)
	}
}

func TestGetClassifier(t *testing.T) {
	// Should return a working classifier (stub) even without model installed
	classifier, err := GetClassifier()
	if err != nil {
		t.Fatalf("GetClassifier() error: %v", err)
	}
	if classifier == nil {
		t.Fatal("GetClassifier() returned nil")
	}
	if classifier.Name() == "" {
		t.Error("Classifier name should not be empty")
	}
	defer classifier.Close()

	// Should be able to classify a file
	result, err := classifier.Classify("/test/main.go")
	if err != nil {
		t.Fatalf("Classify error: %v", err)
	}
	if result.Class != ClassProject {
		t.Errorf("Expected ClassProject for .go file, got %q", result.Class)
	}
}

func TestSplitPath(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"/Users/test/project/main.go", []string{"Users", "test", "project", "main.go"}},
		{"/a/b/c", []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := splitPath(tt.input)
			if len(got) != len(tt.expected) {
				t.Errorf("splitPath(%q) = %v (len %d), want %v (len %d)",
					tt.input, got, len(got), tt.expected, len(tt.expected))
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("splitPath(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.expected[i])
				}
			}
		})
	}
}
