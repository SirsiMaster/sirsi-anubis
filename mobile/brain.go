package mobile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SirsiMaster/sirsi-pantheon/internal/brain"
)

// ModelInfoResponse is the JSON envelope for model backend info.
type ModelInfoResponse struct {
	Name   string `json:"name"`
	Loaded bool   `json:"loaded"`
	Type   string `json:"type"` // "stub", "spotlight", "coreml", "onnx"
}

// BrainClassify classifies a single file and returns Classification JSON.
// The response envelope contains a Classification object with path, class,
// confidence, and model_used fields.
func BrainClassify(filePath string) string {
	if filePath == "" {
		return errorJSON("file path is required")
	}

	classifier, err := brain.GetClassifier()
	if err != nil {
		return errorJSON("failed to get classifier: " + err.Error())
	}
	defer classifier.Close()

	result, err := classifier.Classify(filePath)
	if err != nil {
		return errorJSON("classify failed: " + err.Error())
	}

	return successJSON(result)
}

// BrainClassifyBatch classifies multiple files concurrently.
// pathsJSON is a JSON array of file path strings.
// workers controls concurrency (0 defaults to 4).
// Returns Response JSON with BatchResult data.
func BrainClassifyBatch(pathsJSON string, workers int) string {
	if pathsJSON == "" {
		return errorJSON("paths JSON is required")
	}

	var paths []string
	if err := json.Unmarshal([]byte(pathsJSON), &paths); err != nil {
		return errorJSON("invalid paths JSON: " + err.Error())
	}

	if len(paths) == 0 {
		return errorJSON("paths array is empty")
	}

	classifier, err := brain.GetClassifier()
	if err != nil {
		return errorJSON("failed to get classifier: " + err.Error())
	}
	defer classifier.Close()

	result, err := classifier.ClassifyBatch(paths, workers)
	if err != nil {
		return errorJSON("batch classify failed: " + err.Error())
	}

	return successJSON(result)
}

// BrainModelInfo returns information about the current model backend.
// Returns Response JSON with ModelInfoResponse data.
func BrainModelInfo() string {
	classifier, err := brain.GetClassifier()
	if err != nil {
		return errorJSON("failed to get classifier: " + err.Error())
	}
	defer classifier.Close()

	name := classifier.Name()
	backendType := classifierType(name)

	info := ModelInfoResponse{
		Name:   name,
		Loaded: true,
		Type:   backendType,
	}

	return successJSON(info)
}

// BrainInstallModel copies a CoreML model (.mlmodelc directory) to the weights directory.
// modelPath is the absolute path to the source .mlmodelc directory.
// Returns Response JSON indicating success or failure.
func BrainInstallModel(modelPath string) string {
	if modelPath == "" {
		return errorJSON("model path is required")
	}

	// Validate source exists and is a directory (mlmodelc is a directory)
	info, err := os.Stat(modelPath)
	if err != nil {
		return errorJSON(fmt.Sprintf("model not found at %s: %s", modelPath, err.Error()))
	}
	if !info.IsDir() {
		return errorJSON("model path must be a .mlmodelc directory, not a file")
	}

	// Get the weights directory
	weightsDir, err := brain.WeightsDir()
	if err != nil {
		return errorJSON("failed to resolve weights dir: " + err.Error())
	}

	// Ensure weights directory exists
	if mkErr := os.MkdirAll(weightsDir, 0o755); mkErr != nil {
		return errorJSON("failed to create weights dir: " + mkErr.Error())
	}

	destPath := filepath.Join(weightsDir, "classifier.mlmodelc")

	// Remove existing model if present
	_ = os.RemoveAll(destPath)

	// Copy the mlmodelc directory recursively
	if cpErr := copyDir(modelPath, destPath); cpErr != nil {
		return errorJSON("failed to copy model: " + cpErr.Error())
	}

	result := map[string]string{
		"installed_at": destPath,
		"source":       modelPath,
	}
	return successJSON(result)
}

// copyDir recursively copies a directory tree from src to dst.
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if mkErr := os.MkdirAll(dst, srcInfo.Mode()); mkErr != nil {
		return mkErr
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, readErr := os.ReadFile(srcPath)
			if readErr != nil {
				return readErr
			}
			info, _ := entry.Info()
			mode := os.FileMode(0o644)
			if info != nil {
				mode = info.Mode()
			}
			if writeErr := os.WriteFile(dstPath, data, mode); writeErr != nil {
				return writeErr
			}
		}
	}

	return nil
}

// classifierType maps a classifier name to a human-readable backend type.
func classifierType(name string) string {
	switch {
	case name == "stub-heuristic-v1":
		return "stub"
	case name == "spotlight-mdls":
		return "spotlight"
	case name == "coreml-ane":
		return "coreml"
	default:
		return "unknown"
	}
}
