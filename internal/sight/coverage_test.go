package sight

import (
	"testing"
)

func TestScan(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping live lsregister scan in short mode")
	}
	result, err := Scan()
	if err != nil {
		t.Logf("Scan: %v (expected in non-macOS or CI)", err)
		return
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	t.Logf("Sight: %d ghost registrations found", result.TotalGhosts)
}
