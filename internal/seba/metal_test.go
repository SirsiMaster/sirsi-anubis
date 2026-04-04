package seba

import (
	"crypto/sha256"
	"testing"
)

// ── Metal GPU Hashing Tests ────────────────────────────────────────
// Tests MetalHashBatch which dispatches to either the real Metal GPU
// compute shader (on macOS with CGO) or the pure-Go fallback.

func TestMetalHashBatch_Empty(t *testing.T) {
	t.Parallel()

	result, err := MetalHashBatch(nil)
	if err != nil {
		t.Fatalf("MetalHashBatch(nil) error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for nil input, got %d hashes", len(result))
	}

	result, err = MetalHashBatch([][]byte{})
	if err != nil {
		t.Fatalf("MetalHashBatch([]) error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for empty input, got %d hashes", len(result))
	}
}

func TestMetalHashBatch_SingleBlock(t *testing.T) {
	t.Parallel()

	data := []byte("hello, pantheon")
	expected := sha256.Sum256(data)

	hashes, err := MetalHashBatch([][]byte{data})
	if err != nil {
		t.Fatalf("MetalHashBatch error: %v", err)
	}
	if len(hashes) != 1 {
		t.Fatalf("expected 1 hash, got %d", len(hashes))
	}
	if hashes[0] != expected {
		t.Errorf("hash mismatch:\n  got:  %x\n  want: %x", hashes[0], expected)
	}
}

func TestMetalHashBatch_MultipleBlocks(t *testing.T) {
	t.Parallel()

	blocks := [][]byte{
		[]byte("block zero"),
		[]byte("block one"),
		[]byte("block two"),
		[]byte("block three"),
		[]byte("block four"),
	}

	hashes, err := MetalHashBatch(blocks)
	if err != nil {
		t.Fatalf("MetalHashBatch error: %v", err)
	}
	if len(hashes) != len(blocks) {
		t.Fatalf("expected %d hashes, got %d", len(blocks), len(hashes))
	}

	// Verify each hash matches Go's stdlib SHA-256
	for i, block := range blocks {
		expected := sha256.Sum256(block)
		if hashes[i] != expected {
			t.Errorf("block %d hash mismatch:\n  got:  %x\n  want: %x", i, hashes[i], expected)
		}
	}
}

func TestMetalHashBatch_EmptyBlock(t *testing.T) {
	t.Parallel()

	// SHA-256 of empty input is a well-known constant
	blocks := [][]byte{{}}
	expected := sha256.Sum256([]byte{})

	hashes, err := MetalHashBatch(blocks)
	if err != nil {
		t.Fatalf("MetalHashBatch error: %v", err)
	}
	if len(hashes) != 1 {
		t.Fatalf("expected 1 hash, got %d", len(hashes))
	}
	if hashes[0] != expected {
		t.Errorf("empty block hash mismatch:\n  got:  %x\n  want: %x", hashes[0], expected)
	}
}

func TestMetalHashBatch_MixedSizes(t *testing.T) {
	t.Parallel()

	blocks := [][]byte{
		{},                 // 0 bytes
		{0x42},             // 1 byte
		make([]byte, 64),   // exactly one SHA-256 block
		make([]byte, 65),   // one block + 1 byte
		make([]byte, 4096), // 4 KB (common I/O buffer size)
	}

	// Fill non-empty blocks with deterministic data
	for i := range blocks[2] {
		blocks[2][i] = byte(i % 256)
	}
	for i := range blocks[3] {
		blocks[3][i] = byte((i * 7) % 256)
	}
	for i := range blocks[4] {
		blocks[4][i] = byte((i * 13) % 256)
	}

	hashes, err := MetalHashBatch(blocks)
	if err != nil {
		t.Fatalf("MetalHashBatch error: %v", err)
	}
	if len(hashes) != len(blocks) {
		t.Fatalf("expected %d hashes, got %d", len(blocks), len(hashes))
	}

	for i, block := range blocks {
		expected := sha256.Sum256(block)
		if hashes[i] != expected {
			t.Errorf("block %d (size %d) hash mismatch", i, len(block))
		}
	}
}

func TestMetalHashBatch_LargeBlock(t *testing.T) {
	t.Parallel()

	// 1 MB block — tests that the Metal shader handles large inputs
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	expected := sha256.Sum256(data)

	hashes, err := MetalHashBatch([][]byte{data})
	if err != nil {
		t.Fatalf("MetalHashBatch error: %v", err)
	}
	if hashes[0] != expected {
		t.Error("1 MB block hash mismatch")
	}
}

func TestMetalHashBatch_ManyBlocks(t *testing.T) {
	t.Parallel()

	// 100 blocks — tests parallel dispatch
	n := 100
	blocks := make([][]byte, n)
	for i := range blocks {
		blocks[i] = []byte{byte(i), byte(i >> 8)}
	}

	hashes, err := MetalHashBatch(blocks)
	if err != nil {
		t.Fatalf("MetalHashBatch error: %v", err)
	}
	if len(hashes) != n {
		t.Fatalf("expected %d hashes, got %d", n, len(hashes))
	}

	for i, block := range blocks {
		expected := sha256.Sum256(block)
		if hashes[i] != expected {
			t.Errorf("block %d hash mismatch", i)
		}
	}
}

func TestMetalHashBatch_DeterministicOutput(t *testing.T) {
	t.Parallel()

	data := []byte("deterministic test data for Metal GPU hashing")
	blocks := [][]byte{data, data, data}

	hashes, err := MetalHashBatch(blocks)
	if err != nil {
		t.Fatalf("MetalHashBatch error: %v", err)
	}

	// All three blocks are identical — hashes must be identical
	if hashes[0] != hashes[1] || hashes[1] != hashes[2] {
		t.Error("identical blocks should produce identical hashes")
	}
}

// ── Metal availability and GPU info ────────────────────────────────

func TestMetalAvailable(t *testing.T) {
	t.Parallel()
	// Just verify it doesn't panic — result depends on platform
	_ = metalAvailable()
}

func TestMetalGPUName(t *testing.T) {
	t.Parallel()
	name := metalGPUName()
	// On non-Metal platforms, returns "unavailable"
	if name == "" {
		t.Error("metalGPUName() should not return empty string")
	}
}

func TestMetalGPUCores(t *testing.T) {
	t.Parallel()
	cores := metalGPUCores()
	// On non-Metal platforms, returns 0
	if cores < 0 {
		t.Error("metalGPUCores() should not return negative")
	}
}

// ── Benchmark ──────────────────────────────────────────────────────

func BenchmarkMetalHashBatch_1000x4KB(b *testing.B) {
	blocks := make([][]byte, 1000)
	for i := range blocks {
		blocks[i] = make([]byte, 4096)
		for j := range blocks[i] {
			blocks[i][j] = byte((i*4096 + j) % 256)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MetalHashBatch(blocks)
	}
}
