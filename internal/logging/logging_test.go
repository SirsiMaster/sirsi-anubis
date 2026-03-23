package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestInit_DefaultLevel(t *testing.T) {
	Init(false, false, false)
	// Default level should be warn — debug and info should be suppressed
	var buf bytes.Buffer
	SetOutput(&buf)
	Debug("should not appear")
	Info("should not appear")
	if buf.Len() > 0 {
		t.Errorf("expected no output at warn level, got: %q", buf.String())
	}
}

func TestInit_VerboseMode(t *testing.T) {
	Init(true, false, false)
	var buf bytes.Buffer
	SetOutput(&buf)
	Debug("debug message", "key", "value")
	if !strings.Contains(buf.String(), "debug message") {
		t.Errorf("expected debug output in verbose mode, got: %q", buf.String())
	}
}

func TestInit_QuietMode(t *testing.T) {
	Init(false, true, false)
	var buf bytes.Buffer
	SetOutput(&buf)
	Warn("should not appear")
	Info("should not appear")
	Debug("should not appear")
	if buf.Len() > 0 {
		t.Errorf("expected no output in quiet mode for warn/info/debug, got: %q", buf.String())
	}
}

func TestInit_QuietModeAllowsError(t *testing.T) {
	Init(false, true, false)
	var buf bytes.Buffer
	SetOutput(&buf)
	Error("critical error", "code", 500)
	if !strings.Contains(buf.String(), "critical error") {
		t.Errorf("expected error output in quiet mode, got: %q", buf.String())
	}
}

func TestWith_AddsContext(t *testing.T) {
	Init(true, false, false)
	var buf bytes.Buffer
	SetOutput(&buf)
	contextLogger := With("module", "mirror")
	contextLogger.Debug("scan started")
	output := buf.String()
	if !strings.Contains(output, "module") || !strings.Contains(output, "mirror") {
		t.Errorf("expected context fields in output, got: %q", output)
	}
}

func TestL_ReturnsLogger(t *testing.T) {
	Init(false, false, false)
	l := L()
	if l == nil {
		t.Error("L() returned nil")
	}
}
