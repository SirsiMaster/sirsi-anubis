package mobile

import (
	"encoding/json"
	"testing"

	"github.com/SirsiMaster/sirsi-pantheon/internal/brain"
)

func TestBrainClassify(t *testing.T) {
	tests := []struct {
		name      string
		filePath  string
		wantOK    bool
		wantClass brain.FileClass
	}{
		{
			name:      "go source file",
			filePath:  "/tmp/main.go",
			wantOK:    true,
			wantClass: brain.ClassProject,
		},
		{
			name:      "log file — junk",
			filePath:  "/tmp/app.log",
			wantOK:    true,
			wantClass: brain.ClassJunk,
		},
		{
			name:      "yaml config",
			filePath:  "/etc/app/config.yaml",
			wantOK:    true,
			wantClass: brain.ClassConfig,
		},
		{
			name:      "png media",
			filePath:  "/Users/test/photo.png",
			wantOK:    true,
			wantClass: brain.ClassMedia,
		},
		{
			name:      "zip archive",
			filePath:  "/tmp/backup.zip",
			wantOK:    true,
			wantClass: brain.ClassArchive,
		},
		{
			name:      "csv data",
			filePath:  "/data/report.csv",
			wantOK:    true,
			wantClass: brain.ClassData,
		},
		{
			name:      "onnx model",
			filePath:  "/models/classifier.onnx",
			wantOK:    true,
			wantClass: brain.ClassModel,
		},
		{
			name:      "unknown extension",
			filePath:  "/tmp/mystery.xyz123",
			wantOK:    true,
			wantClass: brain.ClassUnknown,
		},
		{
			name:     "empty path — error",
			filePath: "",
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BrainClassify(tt.filePath)

			var resp Response
			if err := json.Unmarshal([]byte(result), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			if resp.OK != tt.wantOK {
				t.Fatalf("expected ok=%v, got ok=%v (error: %s)", tt.wantOK, resp.OK, resp.Error)
			}

			if !tt.wantOK {
				if resp.Error == "" {
					t.Error("expected non-empty error for failure case")
				}
				return
			}

			var classification brain.Classification
			if err := json.Unmarshal(resp.Data, &classification); err != nil {
				t.Fatalf("failed to parse classification: %v", err)
			}

			if classification.Class != tt.wantClass {
				t.Errorf("expected class %q, got %q", tt.wantClass, classification.Class)
			}

			if classification.Path != tt.filePath {
				t.Errorf("expected path %q, got %q", tt.filePath, classification.Path)
			}

			if classification.ModelUsed == "" {
				t.Error("expected non-empty model_used")
			}
		})
	}
}

func TestBrainClassifyBatch(t *testing.T) {
	tests := []struct {
		name         string
		pathsJSON    string
		workers      int
		wantOK       bool
		wantMinFiles int
	}{
		{
			name:         "valid batch of 3 files",
			pathsJSON:    `["/tmp/main.go", "/tmp/app.log", "/data/report.csv"]`,
			workers:      2,
			wantOK:       true,
			wantMinFiles: 3,
		},
		{
			name:         "single file batch",
			pathsJSON:    `["/tmp/test.py"]`,
			workers:      1,
			wantOK:       true,
			wantMinFiles: 1,
		},
		{
			name:         "default workers (0)",
			pathsJSON:    `["/tmp/main.go"]`,
			workers:      0,
			wantOK:       true,
			wantMinFiles: 1,
		},
		{
			name:      "empty string — error",
			pathsJSON: "",
			workers:   4,
			wantOK:    false,
		},
		{
			name:      "invalid JSON — error",
			pathsJSON: "{not an array}",
			workers:   4,
			wantOK:    false,
		},
		{
			name:      "empty array — error",
			pathsJSON: `[]`,
			workers:   4,
			wantOK:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BrainClassifyBatch(tt.pathsJSON, tt.workers)

			var resp Response
			if err := json.Unmarshal([]byte(result), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			if resp.OK != tt.wantOK {
				t.Fatalf("expected ok=%v, got ok=%v (error: %s)", tt.wantOK, resp.OK, resp.Error)
			}

			if !tt.wantOK {
				if resp.Error == "" {
					t.Error("expected non-empty error for failure case")
				}
				return
			}

			var batch brain.BatchResult
			if err := json.Unmarshal(resp.Data, &batch); err != nil {
				t.Fatalf("failed to parse batch result: %v", err)
			}

			if batch.FilesProcessed < tt.wantMinFiles {
				t.Errorf("expected at least %d files processed, got %d", tt.wantMinFiles, batch.FilesProcessed)
			}

			if batch.ModelUsed == "" {
				t.Error("expected non-empty model_used")
			}

			// Each classification should have valid data
			for _, c := range batch.Classifications {
				if c.Path == "" {
					t.Error("classification has empty path")
				}
				if c.Class == "" {
					t.Error("classification has empty class")
				}
			}
		})
	}
}

func TestBrainModelInfo(t *testing.T) {
	result := BrainModelInfo()

	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("expected ok=true, got error: %s", resp.Error)
	}

	var info ModelInfoResponse
	if err := json.Unmarshal(resp.Data, &info); err != nil {
		t.Fatalf("failed to parse model info: %v", err)
	}

	if info.Name == "" {
		t.Error("expected non-empty model name")
	}

	if !info.Loaded {
		t.Error("expected loaded=true")
	}

	// Type must be one of the known backends
	validTypes := map[string]bool{"stub": true, "spotlight": true, "coreml": true, "unknown": true}
	if !validTypes[info.Type] {
		t.Errorf("unexpected backend type %q", info.Type)
	}
}

func TestClassifierType(t *testing.T) {
	tests := []struct {
		name     string
		wantType string
	}{
		{"stub-heuristic-v1", "stub"},
		{"spotlight-mdls", "spotlight"},
		{"coreml-ane", "coreml"},
		{"something-else", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifierType(tt.name)
			if got != tt.wantType {
				t.Errorf("classifierType(%q) = %q, want %q", tt.name, got, tt.wantType)
			}
		})
	}
}
