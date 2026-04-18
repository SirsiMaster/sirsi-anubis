package workstream

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func tempStore(t *testing.T, initial []Workstream) (*Store, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "workstreams.json")
	if initial != nil {
		data, err := json.MarshalIndent(initial, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	store, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	return store, path
}

func TestNewStore_CreatesFileIfMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "workstreams.json")

	store, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(store.All()) != 0 {
		t.Fatalf("expected empty store, got %d items", len(store.All()))
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected file to be created")
	}
}

func TestNewStore_LoadsExisting(t *testing.T) {
	items := []Workstream{
		{Name: "Alpha", Dir: "~/alpha", Status: StatusActive},
		{Name: "Beta", Dir: "~/beta", Status: StatusRetired},
	}
	store, _ := tempStore(t, items)

	if len(store.All()) != 2 {
		t.Fatalf("expected 2 items, got %d", len(store.All()))
	}
	if store.All()[0].Name != "Alpha" {
		t.Fatalf("expected Alpha, got %s", store.All()[0].Name)
	}
}

func TestActive_FiltersRetired(t *testing.T) {
	items := []Workstream{
		{Name: "A", Dir: "~/a", Status: StatusActive},
		{Name: "B", Dir: "~/b", Status: StatusRetired},
		{Name: "C", Dir: "~/c", Status: StatusActive},
	}
	store, _ := tempStore(t, items)

	active := store.Active()
	if len(active) != 2 {
		t.Fatalf("expected 2 active, got %d", len(active))
	}
	if active[0].Name != "A" || active[1].Name != "C" {
		t.Fatalf("expected A and C, got %s and %s", active[0].Name, active[1].Name)
	}
}

func TestGetActive_MapsDisplayNumber(t *testing.T) {
	items := []Workstream{
		{Name: "A", Dir: "~/a", Status: StatusRetired},
		{Name: "B", Dir: "~/b", Status: StatusActive},
		{Name: "C", Dir: "~/c", Status: StatusActive},
	}
	store, _ := tempStore(t, items)

	tests := []struct {
		displayNum int
		wantName   string
		wantIdx    int
		wantErr    bool
	}{
		{1, "B", 1, false},
		{2, "C", 2, false},
		{0, "", -1, true},
		{3, "", -1, true},
	}

	for _, tt := range tests {
		ws, idx, err := store.GetActive(tt.displayNum)
		if tt.wantErr {
			if err == nil {
				t.Errorf("GetActive(%d): expected error", tt.displayNum)
			}
			continue
		}
		if err != nil {
			t.Errorf("GetActive(%d): %v", tt.displayNum, err)
			continue
		}
		if ws.Name != tt.wantName {
			t.Errorf("GetActive(%d): name = %s, want %s", tt.displayNum, ws.Name, tt.wantName)
		}
		if idx != tt.wantIdx {
			t.Errorf("GetActive(%d): idx = %d, want %d", tt.displayNum, idx, tt.wantIdx)
		}
	}
}

func TestAdd_PersistsToFile(t *testing.T) {
	store, path := tempStore(t, nil)

	if err := store.Add("Test", "~/test"); err != nil {
		t.Fatal(err)
	}

	// Reload from disk
	store2, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(store2.All()) != 1 {
		t.Fatalf("expected 1, got %d", len(store2.All()))
	}
	ws := store2.All()[0]
	if ws.Name != "Test" || ws.Dir != "~/test" || ws.Status != StatusActive {
		t.Fatalf("unexpected workstream: %+v", ws)
	}
	if ws.CreatedAt == "" {
		t.Fatal("expected CreatedAt to be set")
	}
}

func TestAdd_DuplicateNameErrors(t *testing.T) {
	store, _ := tempStore(t, []Workstream{
		{Name: "Existing", Dir: "~/x", Status: StatusActive},
	})
	err := store.Add("Existing", "~/y")
	if err == nil {
		t.Fatal("expected error for duplicate name")
	}
}

func TestRename(t *testing.T) {
	store, _ := tempStore(t, []Workstream{
		{Name: "Old", Dir: "~/x", Status: StatusActive},
	})
	if err := store.Rename(0, "New"); err != nil {
		t.Fatal(err)
	}
	if store.All()[0].Name != "New" {
		t.Fatalf("expected New, got %s", store.All()[0].Name)
	}
}

func TestRetireAndActivate(t *testing.T) {
	store, _ := tempStore(t, []Workstream{
		{Name: "A", Dir: "~/a", Status: StatusActive},
	})

	if err := store.Retire(0); err != nil {
		t.Fatal(err)
	}
	if store.All()[0].Status != StatusRetired {
		t.Fatal("expected retired")
	}
	if len(store.Active()) != 0 {
		t.Fatal("expected 0 active")
	}

	if err := store.Activate(0); err != nil {
		t.Fatal(err)
	}
	if store.All()[0].Status != StatusActive {
		t.Fatal("expected active")
	}
}

func TestDelete(t *testing.T) {
	store, _ := tempStore(t, []Workstream{
		{Name: "A", Dir: "~/a", Status: StatusActive},
		{Name: "B", Dir: "~/b", Status: StatusActive},
	})

	if err := store.Delete(0); err != nil {
		t.Fatal(err)
	}
	if len(store.All()) != 1 {
		t.Fatalf("expected 1, got %d", len(store.All()))
	}
	if store.All()[0].Name != "B" {
		t.Fatalf("expected B, got %s", store.All()[0].Name)
	}
}

func TestBoundsChecks(t *testing.T) {
	store, _ := tempStore(t, []Workstream{
		{Name: "A", Dir: "~/a", Status: StatusActive},
	})

	if err := store.Rename(-1, "x"); err == nil {
		t.Error("expected error for negative index")
	}
	if err := store.Rename(5, "x"); err == nil {
		t.Error("expected error for out of bounds")
	}
	if err := store.Retire(-1); err == nil {
		t.Error("expected error for negative index")
	}
	if err := store.Delete(5); err == nil {
		t.Error("expected error for out of bounds")
	}
	if err := store.Activate(-1); err == nil {
		t.Error("expected error for negative index")
	}
	if err := store.TouchLastUsed(5); err == nil {
		t.Error("expected error for out of bounds")
	}
}

func TestBackwardsCompatibility(t *testing.T) {
	// The existing sw script writes JSON without new fields.
	// Verify we can load it.
	legacy := `[
		{"name": "FinalWishes", "dir": "~/Development/FinalWishes", "memory": "project_finalwishes_overhaul.md", "status": "active"},
		{"name": "Assiduous", "dir": "~/Development/assiduous", "memory": "", "status": "active"}
	]`
	dir := t.TempDir()
	path := filepath.Join(dir, "workstreams.json")
	if err := os.WriteFile(path, []byte(legacy), 0644); err != nil {
		t.Fatal(err)
	}

	store, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(store.All()) != 2 {
		t.Fatalf("expected 2, got %d", len(store.All()))
	}
	ws := store.All()[0]
	if ws.Name != "FinalWishes" {
		t.Fatalf("expected FinalWishes, got %s", ws.Name)
	}
	if ws.Memory != "project_finalwishes_overhaul.md" {
		t.Fatalf("expected memory file, got %s", ws.Memory)
	}
	// New fields should be zero-valued
	if ws.AI != "" || ws.IDE != "" || ws.CreatedAt != "" {
		t.Fatalf("new fields should be empty for legacy data: AI=%s IDE=%s CreatedAt=%s", ws.AI, ws.IDE, ws.CreatedAt)
	}
}

func TestExpandDir(t *testing.T) {
	home, _ := os.UserHomeDir()
	tests := []struct {
		input string
		want  string
	}{
		{"~/foo", filepath.Join(home, "foo")},
		{"~", home},
		{"/absolute/path", "/absolute/path"},
		{"relative", "relative"},
	}
	for _, tt := range tests {
		got := ExpandDir(tt.input)
		if got != tt.want {
			t.Errorf("ExpandDir(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCompressDir(t *testing.T) {
	home, _ := os.UserHomeDir()
	tests := []struct {
		input string
		want  string
	}{
		{filepath.Join(home, "foo"), "~/foo"},
		{"/other/path", "/other/path"},
	}
	for _, tt := range tests {
		got := CompressDir(tt.input)
		if got != tt.want {
			t.Errorf("CompressDir(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFindLauncher(t *testing.T) {
	tests := []struct {
		id   string
		want string
	}{
		{"claude", "Claude Code"},
		{"codex", "Codex"},
		{"gemini", "Gemini CLI"},
		{"antigravity", "Antigravity"},
		{"vscode", "VS Code"},
		{"cursor", "Cursor"},
		{"windsurf", "Windsurf"},
		{"zed", "Zed"},
		{"nonexistent", ""},
	}
	for _, tt := range tests {
		l := FindLauncher(tt.id)
		if tt.want == "" {
			if l != nil {
				t.Errorf("FindLauncher(%q) should be nil", tt.id)
			}
			continue
		}
		if l == nil {
			t.Errorf("FindLauncher(%q) returned nil, want %s", tt.id, tt.want)
			continue
		}
		if l.Name() != tt.want {
			t.Errorf("FindLauncher(%q).Name() = %s, want %s", tt.id, l.Name(), tt.want)
		}
	}
}

func TestAllLaunchers_Count(t *testing.T) {
	all := AllLaunchers()
	if len(all) != 8 {
		t.Fatalf("expected 8 launchers, got %d", len(all))
	}

	// Verify kinds
	aiCount, ideCount := 0, 0
	for _, l := range all {
		switch l.Kind() {
		case "ai":
			aiCount++
		case "ide":
			ideCount++
		default:
			t.Errorf("unexpected kind: %s", l.Kind())
		}
	}
	if aiCount != 3 {
		t.Errorf("expected 3 AI launchers, got %d", aiCount)
	}
	if ideCount != 5 {
		t.Errorf("expected 5 IDE launchers, got %d", ideCount)
	}
}
