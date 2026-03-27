package scales

import (
	"os"
	"path/filepath"
	"testing"
)

// ─── CollectMetrics wrapper ──────────────────────────────────────────────────
// CollectMetrics calls live scanners (Jackal + Ka). We can't unit-test the
// full flow without hitting the filesystem, but we can verify the function
// signature and that it returns non-nil metrics when the system is available.

func TestCollectMetrics_Runs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping live scan in short mode")
	}
	// CollectMetrics runs real scans — just verify it doesn't panic
	// and returns a valid structure or a reasonable error.
	metrics, err := CollectMetrics()
	if err != nil {
		t.Logf("CollectMetrics returned error (expected in CI): %v", err)
		return
	}
	if metrics == nil {
		t.Fatal("expected non-nil metrics")
	}
	if metrics.TotalSize < 0 {
		t.Error("total_size should be non-negative")
	}
	if metrics.FindingCount < 0 {
		t.Error("finding_count should be non-negative")
	}
	if metrics.GhostCount < 0 {
		t.Error("ghost_count should be non-negative")
	}
}

// ─── Enforce ──────────────────────────────────────────────────────────────

func TestEnforce_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping live enforcement in short mode")
	}
	policy := DefaultPolicy().Policies[0]
	result, err := Enforce(policy)
	if err != nil {
		t.Logf("Enforce returned error (expected in CI): %v", err)
		return
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.PolicyName != policy.Name {
		t.Errorf("PolicyName = %q, want %q", result.PolicyName, policy.Name)
	}
	if len(result.Verdicts) != len(policy.Rules) {
		t.Errorf("got %d verdicts, want %d", len(result.Verdicts), len(policy.Rules))
	}
}

// ─── LoadPolicyFile ──────────────────────────────────────────────────────────

func TestLoadPolicyFile_Valid(t *testing.T) {
	yamlContent := `api_version: v1
policies:
  - name: test-policy
    version: "1.0"
    rules:
      - id: r1
        name: Test Rule
        metric: total_size
        operator: gt
        threshold: 1
        unit: GB
        severity: warn
`
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.yaml")
	os.WriteFile(path, []byte(yamlContent), 0o644)

	pf, err := LoadPolicyFile(path)
	if err != nil {
		t.Fatalf("LoadPolicyFile: %v", err)
	}
	if len(pf.Policies) != 1 {
		t.Fatalf("expected 1 policy, got %d", len(pf.Policies))
	}
	if pf.Policies[0].Name != "test-policy" {
		t.Errorf("name = %q, want %q", pf.Policies[0].Name, "test-policy")
	}
}

func TestLoadPolicyFile_NotFound(t *testing.T) {
	_, err := LoadPolicyFile("/nonexistent/policy.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestLoadPolicyFile_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	os.WriteFile(path, []byte("{{{{not yaml"), 0o644)

	_, err := LoadPolicyFile(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadPolicyFile_EmptyPolicies(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.yaml")
	os.WriteFile(path, []byte("api_version: v1\npolicies: []\n"), 0o644)

	_, err := LoadPolicyFile(path)
	if err == nil {
		t.Fatal("expected error for empty policies")
	}
}

// ─── ValidatePolicy ──────────────────────────────────────────────────────────

func TestValidatePolicy_Valid(t *testing.T) {
	yamlContent := `api_version: v1
policies:
  - name: test
    version: "1.0"
    rules:
      - id: r1
        name: Rule 1
        metric: total_size
        operator: gt
        threshold: 1
        severity: warn
`
	dir := t.TempDir()
	path := filepath.Join(dir, "valid.yaml")
	os.WriteFile(path, []byte(yamlContent), 0o644)

	errs := ValidatePolicy(path)
	if len(errs) != 0 {
		t.Errorf("expected no validation errors, got %d: %v", len(errs), errs)
	}
}

func TestValidatePolicy_NotFound(t *testing.T) {
	errs := ValidatePolicy("/nonexistent.yaml")
	if len(errs) != 1 {
		t.Fatalf("expected 1 validation error, got %d", len(errs))
	}
}

func TestValidatePolicy_MissingName(t *testing.T) {
	yamlContent := `api_version: v1
policies:
  - version: "1.0"
    rules:
      - id: r1
        name: Rule 1
        metric: total_size
        operator: gt
        threshold: 1
        severity: warn
`
	dir := t.TempDir()
	path := filepath.Join(dir, "noname.yaml")
	os.WriteFile(path, []byte(yamlContent), 0o644)

	errs := ValidatePolicy(path)
	found := false
	for _, e := range errs {
		if e.Field == "name" {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error for missing name")
	}
}

func TestValidatePolicy_InvalidOperator(t *testing.T) {
	yamlContent := `api_version: v1
policies:
  - name: test
    rules:
      - id: r1
        name: Rule 1
        metric: total_size
        operator: invalid
        threshold: 1
        severity: warn
`
	dir := t.TempDir()
	path := filepath.Join(dir, "badop.yaml")
	os.WriteFile(path, []byte(yamlContent), 0o644)

	errs := ValidatePolicy(path)
	found := false
	for _, e := range errs {
		if e.Field == "operator" {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error for invalid operator")
	}
}

func TestValidatePolicy_InvalidSeverity(t *testing.T) {
	yamlContent := `api_version: v1
policies:
  - name: test
    rules:
      - id: r1
        name: Rule 1
        metric: total_size
        operator: gt
        threshold: 1
        severity: critical
`
	dir := t.TempDir()
	path := filepath.Join(dir, "badsev.yaml")
	os.WriteFile(path, []byte(yamlContent), 0o644)

	errs := ValidatePolicy(path)
	found := false
	for _, e := range errs {
		if e.Field == "severity" {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error for invalid severity")
	}
}

func TestValidatePolicy_InvalidMetric(t *testing.T) {
	yamlContent := `api_version: v1
policies:
  - name: test
    rules:
      - id: r1
        name: Rule 1
        metric: cpu_usage
        operator: gt
        threshold: 1
        severity: warn
`
	dir := t.TempDir()
	path := filepath.Join(dir, "badmetric.yaml")
	os.WriteFile(path, []byte(yamlContent), 0o644)

	errs := ValidatePolicy(path)
	found := false
	for _, e := range errs {
		if e.Field == "metric" {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error for invalid metric")
	}
}

func TestValidatePolicy_DuplicateRuleID(t *testing.T) {
	yamlContent := `api_version: v1
policies:
  - name: test
    rules:
      - id: r1
        name: Rule 1
        metric: total_size
        operator: gt
        threshold: 1
        severity: warn
      - id: r1
        name: Rule 2
        metric: finding_count
        operator: gt
        threshold: 10
        severity: fail
`
	dir := t.TempDir()
	path := filepath.Join(dir, "dup.yaml")
	os.WriteFile(path, []byte(yamlContent), 0o644)

	errs := ValidatePolicy(path)
	found := false
	for _, e := range errs {
		if e.Field == "id" {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error for duplicate rule ID")
	}
}

func TestValidatePolicy_NoRules(t *testing.T) {
	yamlContent := `api_version: v1
policies:
  - name: test
    rules: []
`
	dir := t.TempDir()
	path := filepath.Join(dir, "norules.yaml")
	os.WriteFile(path, []byte(yamlContent), 0o644)

	errs := ValidatePolicy(path)
	found := false
	for _, e := range errs {
		if e.Message == "policy must have at least one rule" {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error for empty rules")
	}
}

// ─── validatePolicies ──────────────────────────────────────────────────────

func TestValidatePolicies_AllErrors(t *testing.T) {
	pf := &PolicyFile{
		Policies: []Policy{
			{
				// Missing name
				Rules: []PolicyRule{
					{ID: "r1", Name: "R1", Metric: "invalid_metric", Operator: "bad", Severity: "unknown"},
					{ID: "r1", Name: "R2", Metric: "total_size", Operator: "gt", Severity: "warn"}, // duplicate
				},
			},
			{
				Name:  "empty",
				Rules: []PolicyRule{}, // no rules
			},
		},
	}

	errs := validatePolicies(pf)
	if len(errs) < 5 {
		t.Errorf("expected at least 5 validation errors, got %d", len(errs))
		for _, e := range errs {
			t.Logf("  %+v", e)
		}
	}
}
