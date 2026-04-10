package mobile

import (
	"encoding/json"
	"testing"
)

func TestSebaDetectHardware(t *testing.T) {
	result := SebaDetectHardware()

	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected ok=true, got error: %s", resp.Error)
	}

	var hw struct {
		CPUCores int    `json:"cpu_cores"`
		CPUModel string `json:"cpu_model"`
		CPUArch  string `json:"cpu_arch"`
		TotalRAM int64  `json:"total_ram"`
	}
	if err := json.Unmarshal(resp.Data, &hw); err != nil {
		t.Fatalf("failed to parse hardware profile: %v", err)
	}

	if hw.CPUCores == 0 {
		t.Error("expected non-zero CPU cores")
	}
	if hw.CPUArch == "" {
		t.Error("expected non-empty CPU arch")
	}
	if hw.TotalRAM == 0 {
		t.Error("expected non-zero RAM")
	}
}

func TestSebaDetectAccelerators(t *testing.T) {
	result := SebaDetectAccelerators()

	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected ok=true, got error: %s", resp.Error)
	}

	var accel struct {
		CPUCores int  `json:"cpu_cores"`
		HasMetal bool `json:"has_metal"`
	}
	if err := json.Unmarshal(resp.Data, &accel); err != nil {
		t.Fatalf("failed to parse accelerator profile: %v", err)
	}

	if accel.CPUCores == 0 {
		t.Error("expected non-zero CPU cores")
	}
}
