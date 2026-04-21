#!/usr/bin/env python3
"""
generate_training_data.py — Brain classifier training data generator.

Walks common macOS directories, extracts file metadata features, and labels
each file using the same heuristic rules as classifyByHeuristic() in
internal/brain/inference.go.

Output: training_data.csv with columns:
  file_extension, path_depth, dir_node_modules, dir_pycache, dir_cache,
  dir_build, dir_dist, dir_vendor, file_size_bucket, basename_pattern, class

Usage:
  python generate_training_data.py [--output training_data.csv] [--max-files 200000]
"""

import argparse
import csv
import os
import sys
from pathlib import Path

# ── FileClass labels (mirrors internal/brain/inference.go) ──────────────
CLASSES = [
    "junk", "essential", "project", "model", "data",
    "media", "archive", "config", "unknown",
]

# ── Top 50 extensions for one-hot encoding ──────────────────────────────
TOP_EXTENSIONS = [
    ".go", ".py", ".js", ".ts", ".rs", ".c", ".cpp", ".h", ".java", ".rb",
    ".swift", ".kt", ".scala", ".zig", ".json", ".yaml", ".yml", ".toml",
    ".ini", ".cfg", ".conf", ".xml", ".plist", ".env", ".log", ".tmp",
    ".bak", ".swp", ".pyc", ".cache", ".onnx", ".pt", ".pth", ".safetensors",
    ".mlmodel", ".tflite", ".csv", ".tsv", ".parquet", ".sqlite", ".db",
    ".sql", ".jpg", ".png", ".gif", ".mp4", ".mp3", ".zip", ".tar", ".gz",
]

# ── Extension-to-class mapping (from inference.go classifyByHeuristic) ──
EXT_CLASS_MAP = {
    # Junk
    ".log": "junk", ".tmp": "junk", ".bak": "junk", ".swp": "junk",
    ".swo": "junk", ".DS_Store": "junk", ".pyc": "junk", ".cache": "junk",
    # Model weights
    ".onnx": "model", ".pt": "model", ".pth": "model", ".safetensors": "model",
    ".ckpt": "model", ".h5": "model", ".pb": "model", ".mlmodel": "model",
    ".mlmodelc": "model", ".tflite": "model", ".bin": "model",
    # Source / Project
    ".go": "project", ".py": "project", ".js": "project", ".ts": "project",
    ".rs": "project", ".c": "project", ".cpp": "project", ".h": "project",
    ".java": "project", ".rb": "project", ".swift": "project", ".kt": "project",
    ".scala": "project", ".zig": "project",
    # Config
    ".yaml": "config", ".yml": "config", ".toml": "config", ".ini": "config",
    ".cfg": "config", ".conf": "config", ".json": "config", ".xml": "config",
    ".plist": "config", ".env": "config",
    # Media
    ".jpg": "media", ".jpeg": "media", ".png": "media", ".gif": "media",
    ".webp": "media", ".svg": "media", ".mp4": "media", ".mov": "media",
    ".avi": "media", ".mkv": "media", ".mp3": "media", ".wav": "media",
    ".flac": "media", ".aac": "media",
    # Archive
    ".zip": "archive", ".tar": "archive", ".gz": "archive", ".bz2": "archive",
    ".xz": "archive", ".7z": "archive", ".rar": "archive", ".dmg": "archive",
    # Data
    ".csv": "data", ".tsv": "data", ".parquet": "data", ".sqlite": "data",
    ".db": "data", ".sql": "data",
}

# ── Basename-to-class mapping ──
BASENAME_CLASS_MAP = {
    "Thumbs.db": "junk", ".DS_Store": "junk",
    "Dockerfile": "project", "Makefile": "project", "Taskfile.yml": "project",
    "LICENSE": "project", "README.md": "project", "CHANGELOG.md": "project",
}

# ── Basename pattern categories ──
BASENAME_PATTERNS = {
    "dotfile": lambda b: b.startswith(".") and not b.startswith(".."),
    "uppercase": lambda b: b == b.upper() and b.isalpha(),
    "lockfile": lambda b: b.endswith(".lock") or b.endswith("-lock.json") or b == "yarn.lock",
    "readme": lambda b: b.lower().startswith("readme"),
    "license": lambda b: b.lower().startswith("license") or b.lower().startswith("licence"),
    "test": lambda b: "test" in b.lower() or "spec" in b.lower(),
    "normal": lambda b: True,  # fallback
}

# ── Directory segment indicators ──
DIR_SEGMENTS = ["node_modules", "__pycache__", ".cache", "build", "dist", "vendor"]


def path_depth(filepath: str) -> int:
    """Count the number of path components."""
    return len(Path(filepath).parts)


def file_size_bucket(size_bytes: int) -> str:
    """Categorize file size into buckets."""
    if size_bytes < 1024:            # < 1KB
        return "tiny"
    elif size_bytes < 100 * 1024:    # < 100KB
        return "small"
    elif size_bytes < 10 * 1024 * 1024:  # < 10MB
        return "medium"
    elif size_bytes < 100 * 1024 * 1024:  # < 100MB
        return "large"
    else:
        return "huge"


def contains_segment(dirpath: str, segment: str) -> bool:
    """Check if directory path contains a specific path segment."""
    parts = Path(dirpath).parts
    return segment in parts


def get_basename_pattern(basename: str) -> str:
    """Classify basename into a pattern category."""
    for pattern_name, check_fn in BASENAME_PATTERNS.items():
        if pattern_name != "normal" and check_fn(basename):
            return pattern_name
    return "normal"


def classify_by_heuristic(filepath: str, size_bytes: int) -> str:
    """
    Replicate classifyByHeuristic() from inference.go.
    Returns the FileClass label string.
    """
    ext = os.path.splitext(filepath)[1]
    basename = os.path.basename(filepath)
    dirpath = os.path.dirname(filepath)

    # Path-based heuristics (most specific first)
    if contains_segment(dirpath, "node_modules"):
        return "junk"
    if contains_segment(dirpath, "__pycache__"):
        return "junk"
    if contains_segment(dirpath, ".cache"):
        return "junk"

    # Basename match
    if basename in BASENAME_CLASS_MAP:
        return BASENAME_CLASS_MAP[basename]

    # Extension match
    if ext in EXT_CLASS_MAP:
        return EXT_CLASS_MAP[ext]

    # Low-confidence path-based
    if contains_segment(dirpath, "build") or contains_segment(dirpath, "dist"):
        return "junk"
    if contains_segment(dirpath, "vendor"):
        return "project"

    return "unknown"


def extract_features(filepath: str, size_bytes: int) -> dict:
    """Extract feature dict for a single file."""
    ext = os.path.splitext(filepath)[1].lower()
    basename = os.path.basename(filepath)
    dirpath = os.path.dirname(filepath)
    depth = path_depth(filepath)

    features = {
        "file_extension": ext if ext else "(none)",
        "path_depth": depth,
    }

    # Directory segment binary features
    for seg in DIR_SEGMENTS:
        features[f"dir_{seg}"] = 1 if contains_segment(dirpath, seg) else 0

    features["file_size_bucket"] = file_size_bucket(size_bytes)
    features["basename_pattern"] = get_basename_pattern(basename)
    features["class"] = classify_by_heuristic(filepath, size_bytes)

    return features


def walk_directories(roots: list[str], max_files: int) -> list[dict]:
    """Walk directory trees and generate feature rows."""
    rows = []
    seen = set()

    for root in roots:
        root = os.path.expanduser(root)
        if not os.path.isdir(root):
            print(f"  Skipping {root} (not a directory)", file=sys.stderr)
            continue

        print(f"  Scanning {root}...", file=sys.stderr)
        scanned = 0

        for dirpath, dirnames, filenames in os.walk(root, followlinks=False):
            # Skip hidden directories we don't care about (performance)
            dirnames[:] = [
                d for d in dirnames
                if not d.startswith(".Trash")
                and d != ".git"
                and d != ".svn"
            ]

            for fname in filenames:
                if len(rows) >= max_files:
                    print(f"    Reached max files ({max_files}), stopping scan.", file=sys.stderr)
                    return rows

                fpath = os.path.join(dirpath, fname)

                # Deduplicate by resolved path
                try:
                    real = os.path.realpath(fpath)
                except OSError:
                    continue
                if real in seen:
                    continue
                seen.add(real)

                # Get file size (skip unreadable files)
                try:
                    stat = os.lstat(fpath)
                    if not stat.st_mode & 0o100000:  # not a regular file
                        continue
                    size = stat.st_size
                except OSError:
                    continue

                features = extract_features(fpath, size)
                rows.append(features)
                scanned += 1

        print(f"    Found {scanned} files in {root}", file=sys.stderr)

    return rows


def main():
    parser = argparse.ArgumentParser(
        description="Generate training data for the Brain file classifier."
    )
    parser.add_argument(
        "--output", "-o",
        default="training_data.csv",
        help="Output CSV path (default: training_data.csv)",
    )
    parser.add_argument(
        "--max-files", "-m",
        type=int,
        default=200_000,
        help="Maximum number of files to scan (default: 200000)",
    )
    parser.add_argument(
        "--dirs",
        nargs="+",
        default=None,
        help="Directories to scan (default: common macOS paths)",
    )
    args = parser.parse_args()

    # Default scan directories for macOS
    scan_dirs = args.dirs or [
        "/usr",
        "/opt",
        "/Applications",
        "~/Library",
        "~/Development",
    ]

    print(f"Brain Training Data Generator", file=sys.stderr)
    print(f"  Max files: {args.max_files}", file=sys.stderr)
    print(f"  Directories: {scan_dirs}", file=sys.stderr)
    print(f"  Output: {args.output}", file=sys.stderr)
    print(file=sys.stderr)

    rows = walk_directories(scan_dirs, args.max_files)

    if not rows:
        print("ERROR: No files found. Check scan directories.", file=sys.stderr)
        sys.exit(1)

    # Write CSV
    fieldnames = [
        "file_extension", "path_depth",
        *[f"dir_{seg}" for seg in DIR_SEGMENTS],
        "file_size_bucket", "basename_pattern", "class",
    ]

    with open(args.output, "w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(rows)

    # Print class distribution
    from collections import Counter
    dist = Counter(r["class"] for r in rows)
    print(f"\nGenerated {len(rows)} samples -> {args.output}", file=sys.stderr)
    print(f"\nClass distribution:", file=sys.stderr)
    for cls in CLASSES:
        count = dist.get(cls, 0)
        pct = 100 * count / len(rows) if rows else 0
        bar = "#" * int(pct / 2)
        print(f"  {cls:12s}  {count:7d}  ({pct:5.1f}%)  {bar}", file=sys.stderr)


if __name__ == "__main__":
    main()
