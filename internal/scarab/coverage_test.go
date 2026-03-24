package scarab

import (
	"testing"
)

func TestAuditContainers_Coverage(t *testing.T) {
	audit, err := AuditContainers()
	if err != nil {
		t.Fatalf("AuditContainers: %v", err)
	}
	if audit == nil {
		t.Fatal("expected non-nil audit")
	}
	// Docker may or may not be running — both cases are valid.
	if audit.DockerRunning {
		t.Logf("Docker running: %d containers (%d running, %d stopped), %d dangling images, %d unused volumes",
			len(audit.Containers), audit.RunningCount, audit.StoppedCount, audit.DanglingImages, audit.UnusedVolumes)
	} else {
		t.Log("Docker is not running (expected in CI)")
	}
}
