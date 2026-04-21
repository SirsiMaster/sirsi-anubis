package mobile

import (
	"encoding/json"

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
