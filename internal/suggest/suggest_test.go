package suggest

import (
	"errors"
	"testing"
)

func TestAfter_AllDeities(t *testing.T) {
	t.Parallel()

	deities := []struct {
		deity string
		subs  []string
	}{
		{"anubis", []string{"weigh", "judge", "ka", "mirror", "apps", ""}},
		{"isis", []string{"network", ""}},
		{"maat", []string{"audit", "pulse", "heal", ""}},
		{"ra", []string{"deploy", "status", "health", "test", "lint", ""}},
		{"net", []string{"align", ""}},
		{"thoth", []string{"sync", "compact", "init", ""}},
		{"seshat", []string{"ingest", "notebooklm", ""}},
		{"seba", []string{"hardware", "diagram", "fleet", ""}},
		{"osiris", []string{"assess", ""}},
		{"horus", []string{"scan", ""}},
	}

	for _, d := range deities {
		for _, sub := range d.subs {
			ctx := Context{Deity: d.deity, Subcommand: sub}
			actions := After(ctx)
			if len(actions) == 0 {
				t.Errorf("After(%s/%s) returned no actions", d.deity, sub)
			}
			for _, a := range actions {
				if a.Short == "" {
					t.Errorf("After(%s/%s): action %q has empty Short", d.deity, sub, a.Command)
				}
				if a.Description == "" {
					t.Errorf("After(%s/%s): action %q has empty Description", d.deity, sub, a.Command)
				}
			}
		}
	}
}

func TestAfter_UnknownDeity(t *testing.T) {
	t.Parallel()
	actions := After(Context{Deity: "unknown"})
	if actions != nil {
		t.Errorf("After(unknown) should return nil, got %d actions", len(actions))
	}
}

func TestAfter_AnubisWeighWithFindings(t *testing.T) {
	t.Parallel()
	actions := After(Context{Deity: "anubis", Subcommand: "weigh", FindingsCount: 42})
	if len(actions) == 0 {
		t.Fatal("expected actions")
	}
	if actions[0].Command != "findings" {
		t.Errorf("expected first action to be 'findings', got %q", actions[0].Command)
	}
}

func TestAfter_AnubisWeighNoFindings(t *testing.T) {
	t.Parallel()
	actions := After(Context{Deity: "anubis", Subcommand: "weigh", FindingsCount: 0})
	for _, a := range actions {
		if a.Command == "findings" {
			t.Error("should not suggest 'findings' when FindingsCount is 0")
		}
	}
}

func TestOnError_PermissionDenied(t *testing.T) {
	t.Parallel()
	actions := OnError(Context{Err: errors.New("permission denied")})
	if len(actions) == 0 {
		t.Fatal("expected error guidance")
	}
	if actions[0].Short != "Check permissions" {
		t.Errorf("expected permission guidance, got %q", actions[0].Short)
	}
}

func TestOnError_NotFound(t *testing.T) {
	t.Parallel()
	actions := OnError(Context{Err: errors.New("file not found")})
	if len(actions) == 0 {
		t.Fatal("expected error guidance")
	}
	if actions[0].Command != "sirsi doctor" {
		t.Errorf("expected doctor suggestion, got %q", actions[0].Command)
	}
}

func TestOnError_Timeout(t *testing.T) {
	t.Parallel()
	actions := OnError(Context{Err: errors.New("context deadline exceeded")})
	if len(actions) == 0 {
		t.Fatal("expected error guidance")
	}
}

func TestOnError_ConnectionRefused(t *testing.T) {
	t.Parallel()
	actions := OnError(Context{Err: errors.New("connection refused")})
	if len(actions) == 0 {
		t.Fatal("expected error guidance")
	}
	if actions[0].Command != "sirsi isis network" {
		t.Errorf("expected isis network suggestion, got %q", actions[0].Command)
	}
}

func TestOnError_DeitySpecific(t *testing.T) {
	t.Parallel()
	deities := []string{"anubis", "ra", "thoth", "maat", "seshat", "horus", "unknown"}
	for _, d := range deities {
		actions := OnError(Context{Deity: d, Err: errors.New("some error")})
		if len(actions) == 0 {
			t.Errorf("OnError(%s) returned no actions", d)
		}
	}
}

func TestPlaceholder_AllDeities(t *testing.T) {
	t.Parallel()

	cases := []struct {
		deity, sub string
	}{
		{"anubis", "weigh"},
		{"anubis", "judge"},
		{"anubis", "ka"},
		{"isis", "network"},
		{"isis", ""},
		{"maat", "audit"},
		{"maat", "pulse"},
		{"ra", "deploy"},
		{"ra", "status"},
		{"net", "align"},
		{"thoth", "sync"},
		{"thoth", "compact"},
		{"seshat", "ingest"},
		{"seba", "hardware"},
		{"seba", "diagram"},
		{"osiris", "assess"},
		{"horus", "scan"},
	}

	for _, c := range cases {
		p := Placeholder(Context{Deity: c.deity, Subcommand: c.sub})
		if p == "" || p == "What next?" {
			t.Errorf("Placeholder(%s/%s) = %q — should be contextual", c.deity, c.sub, p)
		}
	}
}

func TestPlaceholder_Error(t *testing.T) {
	t.Parallel()
	p := Placeholder(Context{Err: errors.New("fail")})
	if p == "" || p == "What next?" {
		t.Errorf("error placeholder should guide to doctor, got %q", p)
	}
}

func TestCommands(t *testing.T) {
	t.Parallel()
	cmds := Commands(Context{Deity: "anubis", Subcommand: "weigh", FindingsCount: 10})
	if len(cmds) == 0 {
		t.Fatal("expected commands")
	}
	for _, c := range cmds {
		if c == "" {
			t.Error("empty command in list")
		}
		if c[0:5] == "sirsi" {
			t.Errorf("Commands should strip 'sirsi ' prefix, got %q", c)
		}
	}
}

func TestCommands_StripsSirsiPrefix(t *testing.T) {
	t.Parallel()
	cmds := Commands(Context{Deity: "maat", Subcommand: "audit"})
	for _, c := range cmds {
		if len(c) > 5 && c[:6] == "sirsi " {
			t.Errorf("command %q still has 'sirsi ' prefix", c)
		}
	}
}
