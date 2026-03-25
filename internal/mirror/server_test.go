package mirror

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SirsiMaster/sirsi-pantheon/internal/platform"
)

// ─── NewServer ────────────────────────────────────────────────────────────

func TestNewServer(t *testing.T) {
	srv, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	if srv == nil {
		t.Fatal("NewServer returned nil")
	}
	if srv.port == 0 {
		t.Error("port should be non-zero")
	}
	url := srv.URL()
	if !strings.HasPrefix(url, "http://127.0.0.1:") {
		t.Errorf("URL = %q, want http://127.0.0.1:*", url)
	}
	// Clean up listener
	srv.listener.Close()
}

func TestServer_URL(t *testing.T) {
	srv := &Server{port: 12345}
	if got := srv.URL(); got != "http://127.0.0.1:12345" {
		t.Errorf("URL() = %q, want http://127.0.0.1:12345", got)
	}
}

// ─── handleBrowse ──────────────────────────────────────────────────────────

func TestHandleBrowse_DefaultDir(t *testing.T) {
	m := &platform.Mock{HomeDir: "/users/mock"}
	srv := &Server{platform: m}
	req := httptest.NewRequest("GET", "/api/browse", nil)
	w := httptest.NewRecorder()

	srv.handleBrowse(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	if body["current"] != "/users/mock" {
		t.Errorf("current = %v, want /users/mock", body["current"])
	}
}

func TestHandleBrowse_WithPath(t *testing.T) {
	m := &platform.Mock{}
	srv := &Server{platform: m}
	req := httptest.NewRequest("GET", "/api/browse?path=/tmp", nil)
	w := httptest.NewRecorder()

	srv.handleBrowse(w, req)

	var body map[string]interface{}
	json.NewDecoder(w.Result().Body).Decode(&body)
	if body["current"] != "/tmp" {
		t.Errorf("current = %v, want /tmp", body["current"])
	}
}

// ─── handlePickFolder ──────────────────────────────────────────────────────

func TestHandlePickFolder(t *testing.T) {
	m := &platform.Mock{PickFolderPath: "/picked/folder"}
	srv := &Server{platform: m}
	req := httptest.NewRequest("GET", "/api/pick-folder", nil)
	w := httptest.NewRecorder()

	srv.handlePickFolder(w, req)

	var body map[string]string
	json.NewDecoder(w.Result().Body).Decode(&body)
	if body["path"] != "/picked/folder" {
		t.Errorf("path = %q, want /picked/folder", body["path"])
	}
}

// ─── handleScan ─────────────────────────────────────────────────────────────

func TestHandleScan_MethodNotAllowed(t *testing.T) {
	srv := &Server{}
	req := httptest.NewRequest("GET", "/api/scan", nil)
	w := httptest.NewRecorder()

	srv.handleScan(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

// ─── OpenBrowser ────────────────────────────────────────────────────────

func TestServer_OpenBrowser(t *testing.T) {
	m := &platform.Mock{}
	srv := &Server{port: 1234, platform: m}
	err := srv.OpenBrowser()
	if err != nil {
		t.Errorf("OpenBrowser: %v", err)
	}
	if m.OpenBrowserURL != "http://127.0.0.1:1234" {
		t.Errorf("OpenBrowserURL = %q, want http://127.0.0.1:1234", m.OpenBrowserURL)
	}
}
