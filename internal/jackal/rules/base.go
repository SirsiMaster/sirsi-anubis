package rules

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/SirsiMaster/sirsi-anubis/internal/cleaner"
	"github.com/SirsiMaster/sirsi-anubis/internal/jackal"
)

// baseScanRule provides shared functionality for simple path-based rules.
type baseScanRule struct {
	name        string
	displayName string
	category    jackal.Category
	description string
	platforms   []string
	paths       []string // Paths to scan (supports ~ expansion)
	excludes    []string // Paths to exclude
	minAgeDays  int      // Minimum age in days (0 = no minimum)
}

func (r *baseScanRule) Name() string              { return r.name }
func (r *baseScanRule) DisplayName() string       { return r.displayName }
func (r *baseScanRule) Category() jackal.Category { return r.category }
func (r *baseScanRule) Description() string       { return r.description }
func (r *baseScanRule) Platforms() []string       { return r.platforms }

func (r *baseScanRule) Scan(ctx context.Context, opts jackal.ScanOptions) ([]jackal.Finding, error) {
	var findings []jackal.Finding
	homeDir := opts.HomeDir
	if homeDir == "" {
		homeDir, _ = os.UserHomeDir()
	}

	minAge := r.minAgeDays
	if opts.MinAgeDays > 0 {
		minAge = opts.MinAgeDays
	}

	for _, pattern := range r.paths {
		expanded := jackal.ExpandPath(pattern, homeDir)

		// Handle glob patterns
		matches, err := filepath.Glob(expanded)
		if err != nil {
			continue
		}

		for _, match := range matches {
			// Check excludes
			if r.isExcluded(match, homeDir) {
				continue
			}

			info, err := os.Lstat(match)
			if err != nil {
				continue
			}

			// Check minimum age
			if minAge > 0 {
				cutoff := time.Now().AddDate(0, 0, -minAge)
				if info.ModTime().After(cutoff) {
					continue
				}
			}

			size := info.Size()
			isDir := info.IsDir()
			fileCount := 1
			if isDir {
				size = cleaner.DirSize(match)
				fileCount = countFiles(match)
			}

			// Skip empty directories/files
			if size == 0 {
				continue
			}

			findings = append(findings, jackal.Finding{
				RuleName:     r.name,
				Category:     r.category,
				Description:  r.displayName,
				Path:         match,
				SizeBytes:    size,
				FileCount:    fileCount,
				Severity:     jackal.SeveritySafe,
				LastModified: info.ModTime(),
				IsDir:        isDir,
			})
		}
	}

	return findings, nil
}

func (r *baseScanRule) Clean(ctx context.Context, findings []jackal.Finding, opts jackal.CleanOptions) (*jackal.CleanResult, error) {
	result := &jackal.CleanResult{}

	for _, f := range findings {
		freed, err := cleaner.DeleteFile(f.Path, opts.DryRun, opts.UseTrash)
		if err != nil {
			result.Skipped++
			result.Errors = append(result.Errors, err)
			continue
		}
		result.Cleaned++
		result.BytesFreed += freed
	}

	return result, nil
}

func (r *baseScanRule) isExcluded(path string, homeDir string) bool {
	for _, exclude := range r.excludes {
		expanded := jackal.ExpandPath(exclude, homeDir)
		matched, _ := filepath.Match(expanded, path)
		if matched {
			return true
		}
		// Also check prefix match for directories
		if len(expanded) > 0 && expanded[len(expanded)-1] != '*' {
			expandedDir := expanded
			if filepath.IsAbs(path) && filepath.IsAbs(expandedDir) {
				rel, err := filepath.Rel(expandedDir, path)
				if err == nil && rel != ".." && !filepath.IsAbs(rel) {
					return true
				}
			}
		}
	}
	return false
}

func countFiles(dir string) int {
	count := 0
	_ = filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	return count
}
