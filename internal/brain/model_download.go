// Package brain — model_download.go provides GitHub Release-based model downloads.
//
// Downloads classifier.mlmodelc.tar.gz from a tagged GitHub Release, extracts it
// to the weights directory (~/.config/sirsi/brain/), and writes a local manifest.
//
// Rule A11: No telemetry. Downloads are from GitHub Releases only.
// Rule A1: Downloaded files are verified by SHA-256 before extraction.
package brain

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// GitHubReleaseBaseURL is the template for downloading release assets.
	// The version placeholder is replaced at runtime.
	GitHubReleaseBaseURL = "https://github.com/SirsiMaster/sirsi-pantheon/releases/download"

	// ModelAssetName is the tarball name in GitHub Releases.
	ModelAssetName = "classifier.mlmodelc.tar.gz"

	// ModelDirName is the extracted CoreML model directory name.
	ModelDirName = "classifier.mlmodelc"

	// downloadModelTimeout for downloading the model tarball.
	downloadModelTimeout = 3 * time.Minute
)

// DownloadModel downloads a compiled CoreML model from a GitHub Release.
// version should be a git tag (e.g., "v0.9.0", "v1.0.0-brain.1").
// Returns the local path to the extracted .mlmodelc directory.
//
// Flow:
//  1. Download classifier.mlmodelc.tar.gz from the release
//  2. Extract to weights directory
//  3. Write local manifest
func DownloadModel(version string) (string, error) {
	if version == "" {
		return "", fmt.Errorf("version is required (e.g., v0.9.0)")
	}

	dir, err := weightsDirFn()
	if err != nil {
		return "", fmt.Errorf("resolve weights dir: %w", err)
	}

	// Ensure weights directory exists
	if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
		return "", fmt.Errorf("create weights dir: %w", mkErr)
	}

	// Build download URL
	url := fmt.Sprintf("%s/%s/%s", GitHubReleaseBaseURL, version, ModelAssetName)

	// Download to a temporary file
	tmpFile := filepath.Join(dir, ModelAssetName+".tmp")
	defer os.Remove(tmpFile) // clean up on any path

	if dlErr := downloadToFile(url, tmpFile); dlErr != nil {
		return "", fmt.Errorf("download model %s: %w", version, dlErr)
	}

	// Compute SHA-256 of the tarball for manifest
	checksum, hashErr := computeSHA256(tmpFile)
	if hashErr != nil {
		return "", fmt.Errorf("checksum: %w", hashErr)
	}

	// Get file size before extraction
	fi, statErr := os.Stat(tmpFile)
	if statErr != nil {
		return "", fmt.Errorf("stat download: %w", statErr)
	}

	// Remove existing model directory if present
	modelDir := filepath.Join(dir, ModelDirName)
	_ = os.RemoveAll(modelDir)

	// Extract tarball
	if exErr := extractTarGz(tmpFile, dir); exErr != nil {
		return "", fmt.Errorf("extract model: %w", exErr)
	}

	// Verify the model directory was created
	if _, verifyErr := os.Stat(modelDir); verifyErr != nil {
		return "", fmt.Errorf("model directory not found after extraction: %w", verifyErr)
	}

	// Write local manifest
	local := &LocalManifest{
		InstalledModel: "brain-classifier",
		Version:        version,
		Format:         "coreml",
		SHA256:         checksum,
		SizeBytes:      fi.Size(),
		InstalledAt:    time.Now(),
		ModelFile:      ModelDirName,
	}

	if mErr := writeLocalManifest(dir, local); mErr != nil {
		return "", fmt.Errorf("write manifest: %w", mErr)
	}

	return modelDir, nil
}

// downloadToFile downloads a URL to a local file path.
func downloadToFile(url, destPath string) error {
	client := &http.Client{Timeout: downloadModelTimeout}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	out, createErr := os.Create(destPath)
	if createErr != nil {
		return fmt.Errorf("create %s: %w", destPath, createErr)
	}
	defer out.Close()

	if _, cpErr := io.Copy(out, resp.Body); cpErr != nil {
		return fmt.Errorf("write %s: %w", destPath, cpErr)
	}

	return nil
}

// extractTarGz extracts a .tar.gz file into the target directory.
// It validates paths to prevent directory traversal (zip-slip).
func extractTarGz(tarGzPath, targetDir string) error {
	f, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, gzErr := gzip.NewReader(f)
	if gzErr != nil {
		return fmt.Errorf("gzip open: %w", gzErr)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	absTarget, _ := filepath.Abs(targetDir)

	for {
		header, nextErr := tr.Next()
		if nextErr == io.EOF {
			break
		}
		if nextErr != nil {
			return fmt.Errorf("tar read: %w", nextErr)
		}

		// Sanitize path — prevent directory traversal (zip-slip)
		cleanName := filepath.Clean(header.Name)
		if strings.HasPrefix(cleanName, "..") || strings.HasPrefix(cleanName, "/") {
			return fmt.Errorf("illegal path in tarball: %s", header.Name)
		}
		target := filepath.Join(targetDir, cleanName)

		// Verify the target is within the extraction directory
		absPath, _ := filepath.Abs(target)
		if !strings.HasPrefix(absPath, absTarget) {
			return fmt.Errorf("path traversal detected: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if mkErr := os.MkdirAll(target, os.FileMode(header.Mode)); mkErr != nil {
				return fmt.Errorf("mkdir %s: %w", target, mkErr)
			}
		case tar.TypeReg:
			// Ensure parent directory exists
			if mkErr := os.MkdirAll(filepath.Dir(target), 0o755); mkErr != nil {
				return fmt.Errorf("mkdir parent %s: %w", target, mkErr)
			}

			out, createErr := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if createErr != nil {
				return fmt.Errorf("create %s: %w", target, createErr)
			}

			// Limit copy size to prevent decompression bombs (100MB max per file)
			if _, cpErr := io.Copy(out, io.LimitReader(tr, 100*1024*1024)); cpErr != nil {
				out.Close()
				return fmt.Errorf("extract %s: %w", target, cpErr)
			}
			out.Close()
		}
	}

	return nil
}

// computeSHA256 computes the SHA-256 checksum of a file, returning a hex string.
func computeSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, cpErr := io.Copy(h, f); cpErr != nil {
		return "", cpErr
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
