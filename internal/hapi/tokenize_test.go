package hapi

import (
	"testing"
)

// ── FastTokenize Tests ──────────────────────────────────────────────────
// These cover the BPE-style tokenizer via the hapi backward-compat wrapper.

func TestFastTokenize_Empty(t *testing.T) {
	t.Parallel()
	tokens := FastTokenize("")
	if len(tokens) != 0 {
		t.Errorf("FastTokenize('') = %d tokens, want 0", len(tokens))
	}
}

func TestFastTokenize_SingleWord(t *testing.T) {
	t.Parallel()
	tokens := FastTokenize("hello")
	if len(tokens) != 1 {
		t.Errorf("FastTokenize('hello') = %d tokens, want 1", len(tokens))
	}
	if tokens[0] == 0 {
		t.Error("token hash should not be zero for non-empty input")
	}
}

func TestFastTokenize_WordsAndSpaces(t *testing.T) {
	t.Parallel()
	tokens := FastTokenize("hello world")
	// Expect: "hello" (word), " " (whitespace), "world" (word) = 3 tokens
	if len(tokens) != 3 {
		t.Errorf("FastTokenize('hello world') = %d tokens, want 3", len(tokens))
	}
}

func TestFastTokenize_SpecialChars(t *testing.T) {
	t.Parallel()
	tokens := FastTokenize("a+b=c")
	// "a" (word), "+" (special), "b" (word), "=" (special), "c" (word) = 5 tokens
	if len(tokens) != 5 {
		t.Errorf("FastTokenize('a+b=c') = %d tokens, want 5", len(tokens))
	}
}

func TestFastTokenize_Deterministic(t *testing.T) {
	t.Parallel()
	text := "The quick brown fox jumps over the lazy dog."
	t1 := FastTokenize(text)
	t2 := FastTokenize(text)
	if len(t1) != len(t2) {
		t.Fatalf("Tokenize not deterministic: %d != %d", len(t1), len(t2))
	}
	for i := range t1 {
		if t1[i] != t2[i] {
			t.Errorf("token[%d] differs: %d != %d", i, t1[i], t2[i])
		}
	}
}

func TestFastTokenize_MixedContent(t *testing.T) {
	t.Parallel()
	// Multi-line with tabs, special chars, words
	text := "func main() {\n\tfmt.Println(\"hello\")\n}"
	tokens := FastTokenize(text)
	if len(tokens) == 0 {
		t.Error("should produce tokens for Go code")
	}
	// Verify all tokens are positive (bit-masked to 0x7FFFFFFF)
	for i, tok := range tokens {
		if tok < 0 {
			t.Errorf("token[%d] = %d, should be non-negative", i, tok)
		}
	}
}

func TestFastTokenize_OnlyWhitespace(t *testing.T) {
	t.Parallel()
	tokens := FastTokenize("   \t\n  ")
	// All whitespace clusters into 1 token
	if len(tokens) != 1 {
		t.Errorf("FastTokenize(whitespace) = %d tokens, want 1", len(tokens))
	}
}

func TestFastTokenize_OnlySpecial(t *testing.T) {
	t.Parallel()
	tokens := FastTokenize("+-*/")
	// Each special char is its own token
	if len(tokens) != 4 {
		t.Errorf("FastTokenize('+-*/') = %d tokens, want 4", len(tokens))
	}
}

func TestFastTokenize_LargeInput(t *testing.T) {
	t.Parallel()
	// Should not panic or allocate excessively
	large := make([]byte, 10000)
	for i := range large {
		large[i] = 'a' + byte(i%26)
	}
	tokens := FastTokenize(string(large))
	if len(tokens) == 0 {
		t.Error("should produce tokens for large input")
	}
}

// ── AcceleratorProfile struct ────────────────────────────────────────────

func TestAcceleratorProfile_ZeroValue(t *testing.T) {
	t.Parallel()
	p := AcceleratorProfile{}
	if p.HasGPU || p.HasANE || p.HasMetal || p.HasCUDA || p.HasROCm || p.HasOneAPI {
		t.Error("zero-value profile should have no hardware flags set")
	}
	if p.CPUCores != 0 {
		t.Errorf("CPUCores = %d, want 0", p.CPUCores)
	}
}
