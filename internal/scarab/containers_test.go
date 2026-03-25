package scarab

import (
	"fmt"
	"testing"

	"github.com/SirsiMaster/sirsi-pantheon/internal/platform"
)

func TestAuditContainers_DockerNotRunning(t *testing.T) {
	m := &platform.Mock{
		CommandError: fmt.Errorf("docker not found"),
	}

	audit, err := AuditContainersWith(m)
	if err != nil {
		t.Fatalf("AuditContainersWith failed: %v", err)
	}
	if audit.DockerRunning {
		t.Error("Expected DockerRunning to be false")
	}
}

func TestAuditContainers_Success(t *testing.T) {
	m := &platform.Mock{
		CommandResults: map[string]string{
			"docker info": "OK",
			"docker ps -a --format {{.ID}}\t{{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}": "id1\tname1\timg1\tUp 2 hours\t80:80\nid2\tname2\timg2\tExited (0)\t",
			"docker images -f dangling=true -q":                                              "img1\nimg2\n",
			"docker volume ls -f dangling=true -q":                                           "vol1\n",
		},
	}

	audit, err := AuditContainersWith(m)
	if err != nil {
		t.Fatalf("AuditContainersWith failed: %v", err)
	}

	if !audit.DockerRunning {
		t.Error("Expected DockerRunning to be true")
	}
	if audit.RunningCount != 1 {
		t.Errorf("RunningCount = %d, want 1", audit.RunningCount)
	}
	if audit.StoppedCount != 1 {
		t.Errorf("StoppedCount = %d, want 1", audit.StoppedCount)
	}
	if audit.DanglingImages != 2 {
		t.Errorf("DanglingImages = %d, want 2", audit.DanglingImages)
	}
	if audit.UnusedVolumes != 1 {
		t.Errorf("UnusedVolumes = %d, want 1", audit.UnusedVolumes)
	}
	if len(audit.Containers) != 2 {
		t.Errorf("Containers count = %d, want 2", len(audit.Containers))
	}
}
