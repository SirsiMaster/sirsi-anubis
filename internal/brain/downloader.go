// Package brain provides on-demand neural model management for Anubis Pro.
// Downloads, verifies, and manages CoreML/ONNX weights in ~/.anubis/weights/.
// The base Anubis binary ships without models — "install-brain" adds neural
// capabilities for semantic file classification and context sanitization.
//
// Rule A11: No telemetry. Downloads are from GitHub Releases only.
// Rule A1: Weights are self-deletable via --remove or --stealth.
package brain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	// DefaultManifestURL points to the brain manifest in the latest GitHub Release.
	DefaultManifestURL = "https://raw.githubusercontent.com/SirsiMaster/sirsi-pantheon/main/brain-manifest.json"

	// DefaultWeightsDir is where models are stored.
	DefaultWeightsDir = ".anubis/weights"

	// ManifestFile is the local manifest filename.
	ManifestFile = "manifest.json"

	// httpTimeout for downloads (generous for large models).
	httpTimeout = 5 * time.Minute

	// manifestTimeout for manifest fetches.
	manifestTimeout = 10 * time.Second
)

// ModelInfo describes a downloadable model in the remote manifest.
type ModelInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Format      string `json:"format"` // "onnx", "coreml", "both"
	SizeBytes   int64  `json:"size_bytes"`
	SHA256      string `json:"sha256"`
	DownloadURL string `json:"download_url"`
	MinVersion  string `json:"min_version"` // Minimum Anubis version required
}

// RemoteManifest is the structure of brain-manifest.json in the repo.
type RemoteManifest struct {
	SchemaVersion int         `json:"schema_version"`
	Updated       string      `json:"updated"`
	Models        []ModelInfo `json:"models"`
	DefaultModel  string      `json:"default_model"`
}

// LocalManifest tracks what's installed locally in ~/.anubis/weights/.
type LocalManifest struct {
	InstalledModel string    `json:"installed_model"`
	Version        string    `json:"version"`
	Format         string    `json:"format"`
	SHA256         string    `json:"sha256"`
	SizeBytes      int64     `json:"size_bytes"`
	InstalledAt    time.Time `json:"installed_at"`
	ModelFile      string    `json:"model_file"`
}

// ProgressFunc is called during download with bytes downloaded and total size.
type ProgressFunc func(downloaded, total int64)

// Status represents the result of a brain operation.
type Status struct {
	Installed    bool           `json:"installed"`
	Model        *LocalManifest `json:"model,omitempty"`
	Available    []ModelInfo    `json:"available,omitempty"`
	UpdateReady  bool           `json:"update_ready"`
	LatestRemote *ModelInfo     `json:"latest_remote,omitempty"`
	WeightsDir   string         `json:"weights_dir"`
	Error        string         `json:"error,omitempty"`
}

// WeightsDir returns the absolute path to the weights directory.
func WeightsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, DefaultWeightsDir), nil
}

// GetStatus checks the current brain installation status.
func GetStatus() (*Status, error) {
	dir, err := WeightsDir()
	if err != nil {
		return nil, err
	}

	status := &Status{
		WeightsDir: dir,
	}

	// Check local manifest
	local, err := readLocalManifest(dir)
	if err == nil && local != nil {
		status.Installed = true
		status.Model = local
	}

	// Check remote manifest for updates
	remote, err := FetchRemoteManifest()
	if err == nil && remote != nil {
		status.Available = remote.Models

		// Find if update is available
		if status.Installed && local != nil {
			for _, m := range remote.Models {
				if m.Name == local.InstalledModel && m.Version != local.Version {
					status.UpdateReady = true
					info := m
					status.LatestRemote = &info
					break
				}
			}
		}
	}

	return status, nil
}

// FetchRemoteManifest downloads the brain manifest from GitHub.
func FetchRemoteManifest() (*RemoteManifest, error) {
	client := &http.Client{Timeout: manifestTimeout}

	resp, err := client.Get(DefaultManifestURL)
	if err != nil {
		return nil, fmt.Errorf("fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifest fetch returned HTTP %d", resp.StatusCode)
	}

	var manifest RemoteManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("decode manifest: %w", err)
	}

	return &manifest, nil
}

// Install downloads and installs the default brain model.
// The onProgress callback is called during download with (bytesDownloaded, totalBytes).
func Install(onProgress ProgressFunc) (*LocalManifest, error) {
	// Fetch remote manifest
	remote, err := FetchRemoteManifest()
	if err != nil {
		return nil, fmt.Errorf("fetch remote manifest: %w", err)
	}

	if len(remote.Models) == 0 {
		return nil, fmt.Errorf("no models available in remote manifest")
	}

	// Find default model (or use first available)
	var model *ModelInfo
	for _, m := range remote.Models {
		if m.Name == remote.DefaultModel {
			info := m
			model = &info
			break
		}
	}
	if model == nil {
		info := remote.Models[0]
		model = &info
	}

	// Select platform-appropriate model
	model = selectPlatformModel(model, remote)

	return installModel(model, onProgress)
}

// InstallFromManifest installs a specific model by name from the remote manifest.
func InstallFromManifest(modelName string, onProgress ProgressFunc) (*LocalManifest, error) {
	remote, err := FetchRemoteManifest()
	if err != nil {
		return nil, fmt.Errorf("fetch remote manifest: %w", err)
	}

	for _, m := range remote.Models {
		if m.Name == modelName {
			return installModel(&m, onProgress)
		}
	}

	return nil, fmt.Errorf("model %q not found in remote manifest", modelName)
}

// Update checks for a newer version and re-downloads if available.
func Update(onProgress ProgressFunc) (*LocalManifest, bool, error) {
	dir, err := WeightsDir()
	if err != nil {
		return nil, false, err
	}

	local, err := readLocalManifest(dir)
	if err != nil || local == nil {
		// Not installed — do a fresh install
		manifest, installErr := Install(onProgress)
		return manifest, true, installErr
	}

	// Fetch remote
	remote, err := FetchRemoteManifest()
	if err != nil {
		return nil, false, fmt.Errorf("fetch remote manifest: %w", err)
	}

	// Check if update is available
	for _, m := range remote.Models {
		if m.Name == local.InstalledModel && m.Version != local.Version {
			// Different version available — update
			manifest, installErr := installModel(&m, onProgress)
			return manifest, true, installErr
		}
	}

	return local, false, nil // Already up to date
}

// Remove deletes all installed brain weights and the local manifest.
func Remove() error {
	dir, err := WeightsDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil // Nothing to remove
	}

	return os.RemoveAll(dir)
}

// IsInstalled returns true if a brain model is currently installed.
func IsInstalled() bool {
	dir, err := WeightsDir()
	if err != nil {
		return false
	}

	local, err := readLocalManifest(dir)
	return err == nil && local != nil
}

// selectPlatformModel picks the best model variant for the current platform.
func selectPlatformModel(defaultModel *ModelInfo, remote *RemoteManifest) *ModelInfo {
	// On macOS with Apple Silicon, prefer CoreML if available
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		for _, m := range remote.Models {
			if m.Format == "coreml" && strings.HasPrefix(m.Name, strings.TrimSuffix(defaultModel.Name, "-onnx")) {
				info := m
				return &info
			}
		}
	}
	return defaultModel
}

// installModel handles the actual download, verification, and manifest write.
func installModel(model *ModelInfo, onProgress ProgressFunc) (*LocalManifest, error) {
	dir, err := WeightsDir()
	if err != nil {
		return nil, err
	}

	// Create weights directory
	if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
		return nil, fmt.Errorf("create weights dir: %w", mkErr)
	}

	// Determine local filename
	filename := model.Name
	switch model.Format {
	case "onnx":
		if !strings.HasSuffix(filename, ".onnx") {
			filename += ".onnx"
		}
	case "coreml":
		if !strings.HasSuffix(filename, ".mlmodelc") && !strings.HasSuffix(filename, ".mlpackage") {
			filename += ".mlmodelc"
		}
	default:
		filename += ".bin"
	}
	localPath := filepath.Join(dir, filename)

	// Download with progress
	if dlErr := downloadFile(model.DownloadURL, localPath, model.SizeBytes, onProgress); dlErr != nil {
		// Clean up partial download
		_ = os.Remove(localPath)
		return nil, fmt.Errorf("download model: %w", dlErr)
	}

	// Verify checksum
	if model.SHA256 != "" {
		actualHash, hashErr := hashFile(localPath)
		if hashErr != nil {
			_ = os.Remove(localPath)
			return nil, fmt.Errorf("checksum verification: %w", hashErr)
		}
		if !strings.EqualFold(actualHash, model.SHA256) {
			_ = os.Remove(localPath)
			return nil, fmt.Errorf("checksum mismatch: expected %s, got %s", model.SHA256, actualHash)
		}
	}

	// Get actual file size
	fi, err := os.Stat(localPath)
	if err != nil {
		return nil, fmt.Errorf("stat model file: %w", err)
	}

	// Write local manifest
	local := &LocalManifest{
		InstalledModel: model.Name,
		Version:        model.Version,
		Format:         model.Format,
		SHA256:         model.SHA256,
		SizeBytes:      fi.Size(),
		InstalledAt:    time.Now(),
		ModelFile:      filename,
	}

	if err := writeLocalManifest(dir, local); err != nil {
		return nil, fmt.Errorf("write local manifest: %w", err)
	}

	return local, nil
}

// downloadFile downloads a URL to a local path with progress reporting.
func downloadFile(url, destPath string, expectedSize int64, onProgress ProgressFunc) error {
	client := &http.Client{Timeout: httpTimeout}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP GET: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned HTTP %d", resp.StatusCode)
	}

	// Use Content-Length if available, fall back to expected size
	total := resp.ContentLength
	if total <= 0 {
		total = expectedSize
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	// Copy with progress reporting
	var downloaded int64
	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := out.Write(buf[:n])
			if writeErr != nil {
				return fmt.Errorf("write: %w", writeErr)
			}
			downloaded += int64(n)
			if onProgress != nil {
				onProgress(downloaded, total)
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return fmt.Errorf("read: %w", readErr)
		}
	}

	return nil
}

// hashFile computes the SHA-256 hash of a file.
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// readLocalManifest reads the local manifest from the weights directory.
func readLocalManifest(dir string) (*LocalManifest, error) {
	path := filepath.Join(dir, ManifestFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest LocalManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parse local manifest: %w", err)
	}

	return &manifest, nil
}

// writeLocalManifest writes the local manifest to the weights directory.
func writeLocalManifest(dir string, manifest *LocalManifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}

	path := filepath.Join(dir, ManifestFile)
	return os.WriteFile(path, data, 0o644)
}

// FormatBytes formats bytes into a human-readable string.
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
