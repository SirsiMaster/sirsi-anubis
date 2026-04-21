#!/usr/bin/env python3
"""
train_model.py — Train the Brain file classifier and export to CoreML.

Reads training_data.csv, trains a Random Forest classifier using scikit-learn,
evaluates with a train/test split, and exports to CoreML format (.mlmodel)
then compiles to .mlmodelc.

The model is intentionally simple (Random Forest on tabular features):
- No deep learning, no GPU required
- Fast training (<30s on any machine)
- Small output (<5MB)
- Interpretable feature importances

Usage:
  python train_model.py [--input training_data.csv] [--output classifier]
"""

import argparse
import os
import subprocess
import sys

import coremltools as ct
import numpy as np
import pandas as pd
from sklearn.ensemble import RandomForestClassifier
from sklearn.metrics import classification_report
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import LabelEncoder, OrdinalEncoder

# ── Feature columns ─────────────────────────────────────────────────────
CATEGORICAL_FEATURES = ["file_extension", "file_size_bucket", "basename_pattern"]
BINARY_FEATURES = [
    "dir_node_modules", "dir___pycache__", "dir_.cache",
    "dir_build", "dir_dist", "dir_vendor",
]
NUMERIC_FEATURES = ["path_depth"]
TARGET = "class"

# All 9 FileClass labels from inference.go
ALL_CLASSES = [
    "junk", "essential", "project", "model", "data",
    "media", "archive", "config", "unknown",
]


def load_data(csv_path: str) -> pd.DataFrame:
    """Load and validate the training CSV."""
    df = pd.read_csv(csv_path)
    required = CATEGORICAL_FEATURES + BINARY_FEATURES + NUMERIC_FEATURES + [TARGET]
    missing = [c for c in required if c not in df.columns]
    if missing:
        print(f"ERROR: Missing columns in CSV: {missing}", file=sys.stderr)
        sys.exit(1)
    print(f"Loaded {len(df)} samples from {csv_path}", file=sys.stderr)
    return df


def encode_features(df: pd.DataFrame):
    """
    Encode categorical features using ordinal encoding.
    Returns (X array, y array, feature_names, encoders dict, label_encoder).
    """
    encoders = {}

    # Encode categorical features with OrdinalEncoder (handles unseen as -1)
    for col in CATEGORICAL_FEATURES:
        enc = OrdinalEncoder(handle_unknown="use_encoded_value", unknown_value=-1)
        df[col + "_enc"] = enc.fit_transform(df[[col]]).astype(int).ravel()
        encoders[col] = enc

    # Encode target labels
    label_enc = LabelEncoder()
    label_enc.classes_ = np.array(ALL_CLASSES)
    df[TARGET + "_enc"] = label_enc.transform(df[TARGET])

    # Build feature matrix
    feature_cols = (
        [c + "_enc" for c in CATEGORICAL_FEATURES]
        + BINARY_FEATURES
        + NUMERIC_FEATURES
    )
    feature_names = (
        [c + "_encoded" for c in CATEGORICAL_FEATURES]
        + BINARY_FEATURES
        + NUMERIC_FEATURES
    )

    X = df[feature_cols].values.astype(np.float32)
    y = df[TARGET + "_enc"].values

    return X, y, feature_names, encoders, label_enc


def train_model(X_train, y_train, label_enc):
    """Train a Random Forest classifier."""
    clf = RandomForestClassifier(
        n_estimators=100,
        max_depth=20,
        min_samples_split=5,
        min_samples_leaf=2,
        class_weight="balanced",  # Handle class imbalance
        random_state=42,
        n_jobs=-1,  # Use all CPU cores
    )
    clf.fit(X_train, y_train)
    return clf


def evaluate_model(clf, X_test, y_test, label_enc, feature_names):
    """Print classification report and feature importances."""
    y_pred = clf.predict(X_test)

    # Only use labels that appear in test set
    present_labels = sorted(set(y_test) | set(y_pred))
    target_names = [ALL_CLASSES[i] for i in present_labels]

    print("\n--- Classification Report ---", file=sys.stderr)
    print(
        classification_report(
            y_test, y_pred,
            labels=present_labels,
            target_names=target_names,
            zero_division=0,
        ),
        file=sys.stderr,
    )

    # Feature importances
    importances = clf.feature_importances_
    sorted_idx = np.argsort(importances)[::-1]
    print("--- Feature Importances ---", file=sys.stderr)
    for i in sorted_idx[:10]:
        print(f"  {feature_names[i]:30s}  {importances[i]:.4f}", file=sys.stderr)
    print(file=sys.stderr)


def export_coreml(clf, feature_names, label_enc, output_name: str):
    """
    Convert sklearn RandomForest to CoreML format.
    Outputs both .mlmodel (source) and .mlmodelc (compiled).
    """
    mlmodel_path = output_name + ".mlmodel"
    mlmodelc_path = output_name + ".mlmodelc"

    # Convert to CoreML using coremltools
    coreml_model = ct.converters.sklearn.convert(
        clf,
        input_features=feature_names,
        output_feature_names="class_label",
    )

    # Set model metadata
    coreml_model.author = "Sirsi Technologies"
    coreml_model.short_description = (
        "Brain file classifier — classifies files into 9 semantic categories "
        "(junk, essential, project, model, data, media, archive, config, unknown)"
    )
    coreml_model.version = "1.0.0"
    coreml_model.license = "Apache-2.0"

    # Add class label mapping as user-defined metadata
    spec = coreml_model.get_spec()
    for i, label in enumerate(ALL_CLASSES):
        spec.description.metadata.userDefined[f"class_{i}"] = label

    # Save .mlmodel
    ct.utils.save_spec(spec, mlmodel_path)
    size_mb = os.path.getsize(mlmodel_path) / (1024 * 1024)
    print(f"Saved {mlmodel_path} ({size_mb:.1f} MB)", file=sys.stderr)

    # Compile to .mlmodelc using xcrun coremlcompiler (macOS only)
    if sys.platform == "darwin":
        print(f"Compiling to {mlmodelc_path}...", file=sys.stderr)
        try:
            # coremlcompiler outputs a directory
            subprocess.run(
                ["xcrun", "coremlcompiler", "compile", mlmodel_path, "."],
                check=True,
                capture_output=True,
                text=True,
            )
            if os.path.isdir(mlmodelc_path):
                size_total = sum(
                    os.path.getsize(os.path.join(dp, f))
                    for dp, _, fnames in os.walk(mlmodelc_path)
                    for f in fnames
                )
                print(
                    f"Compiled {mlmodelc_path} ({size_total / (1024 * 1024):.1f} MB)",
                    file=sys.stderr,
                )
            else:
                print(
                    f"WARNING: coremlcompiler did not produce {mlmodelc_path}. "
                    "The .mlmodel file is still usable.",
                    file=sys.stderr,
                )
        except FileNotFoundError:
            print(
                "WARNING: xcrun coremlcompiler not found. "
                "Install Xcode Command Line Tools to compile .mlmodelc. "
                "The .mlmodel file is still usable for conversion.",
                file=sys.stderr,
            )
        except subprocess.CalledProcessError as e:
            print(
                f"WARNING: coremlcompiler failed: {e.stderr}. "
                "The .mlmodel file is still usable.",
                file=sys.stderr,
            )
    else:
        print(
            "NOTE: .mlmodelc compilation requires macOS. "
            f"Transfer {mlmodel_path} to a Mac and run:\n"
            f"  xcrun coremlcompiler compile {mlmodel_path} .",
            file=sys.stderr,
        )

    return mlmodel_path, mlmodelc_path


def main():
    parser = argparse.ArgumentParser(
        description="Train the Brain file classifier and export to CoreML."
    )
    parser.add_argument(
        "--input", "-i",
        default="training_data.csv",
        help="Input CSV from generate_training_data.py (default: training_data.csv)",
    )
    parser.add_argument(
        "--output", "-o",
        default="classifier",
        help="Output model name without extension (default: classifier)",
    )
    parser.add_argument(
        "--test-size",
        type=float,
        default=0.2,
        help="Fraction of data for test split (default: 0.2)",
    )
    args = parser.parse_args()

    print("Brain Model Trainer", file=sys.stderr)
    print(f"  Input:     {args.input}", file=sys.stderr)
    print(f"  Output:    {args.output}.mlmodel / {args.output}.mlmodelc", file=sys.stderr)
    print(f"  Test size: {args.test_size}", file=sys.stderr)
    print(file=sys.stderr)

    # Load data
    df = load_data(args.input)

    # Filter out classes with too few samples for stratified split
    class_counts = df[TARGET].value_counts()
    min_samples = max(2, int(1 / args.test_size) + 1)  # need at least 1 per split
    valid_classes = class_counts[class_counts >= min_samples].index.tolist()
    dropped = class_counts[class_counts < min_samples]
    if len(dropped) > 0:
        print(f"NOTE: Dropping classes with <{min_samples} samples: "
              f"{dict(dropped)}", file=sys.stderr)
        df = df[df[TARGET].isin(valid_classes)]

    # Encode features
    X, y, feature_names, encoders, label_enc = encode_features(df)
    print(f"Feature matrix: {X.shape[0]} samples x {X.shape[1]} features", file=sys.stderr)

    # Train/test split
    X_train, X_test, y_train, y_test = train_test_split(
        X, y,
        test_size=args.test_size,
        random_state=42,
        stratify=y,
    )
    print(f"Train: {len(X_train)}, Test: {len(X_test)}", file=sys.stderr)

    # Train
    print("\nTraining Random Forest...", file=sys.stderr)
    clf = train_model(X_train, y_train, label_enc)
    print(f"Training complete. Trees: {clf.n_estimators}, "
          f"Max depth: {clf.max_depth}", file=sys.stderr)

    # Evaluate
    evaluate_model(clf, X_test, y_test, label_enc, feature_names)

    # Export to CoreML
    print("Exporting to CoreML...", file=sys.stderr)
    mlmodel_path, mlmodelc_path = export_coreml(
        clf, feature_names, label_enc, args.output,
    )

    # Summary
    print("\n--- Summary ---", file=sys.stderr)
    print(f"  Training samples:  {len(X_train)}", file=sys.stderr)
    print(f"  Test samples:      {len(X_test)}", file=sys.stderr)
    print(f"  Features:          {len(feature_names)}", file=sys.stderr)
    print(f"  Classes:           {len(valid_classes)}", file=sys.stderr)
    print(f"  Model (source):    {mlmodel_path}", file=sys.stderr)
    print(f"  Model (compiled):  {mlmodelc_path}", file=sys.stderr)
    print(f"\nTo install into Pantheon:", file=sys.stderr)
    print(f"  cp -R {mlmodelc_path} ~/.config/sirsi/brain/classifier.mlmodelc",
          file=sys.stderr)


if __name__ == "__main__":
    main()
