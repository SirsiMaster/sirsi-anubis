package mirror

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
	srv := &Server{}
	req := httptest.NewRequest("GET", "/api/browse", nil)
	w := httptest.NewRecorder()

	srv.handleBrowse(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	if body["current"] == nil {
		t.Error("expected 'current' field in response")
	}
}

func TestHandleBrowse_WithPath(t *testing.T) {
	srv := &Server{}
	req := httptest.NewRequest("GET", "/api/browse?path=/tmp", nil)
	w := httptest.NewRecorder()

	srv.handleBrowse(w, req)

	var body map[string]interface{}
	json.NewDecoder(w.Result().Body).Decode(&body)
	if body["current"] != "/tmp" {
		t.Errorf("current = %v, want /tmp", body["current"])
	}
}

func TestHandleBrowse_InvalidPath(t *testing.T) {
	srv := &Server{}
	req := httptest.NewRequest("GET", "/api/browse?path=/nonexistent_dir_xyz", nil)
	w := httptest.NewRecorder()

	srv.handleBrowse(w, req)

	var body map[string]interface{}
	json.NewDecoder(w.Result().Body).Decode(&body)
	if body["error"] == nil {
		t.Error("expected 'error' field for invalid path")
	}
}

// ─── handlePickFolder ──────────────────────────────────────────────────────

func TestHandlePickFolder(t *testing.T) {
	t.Skip("PickFolder launches interactive Finder dialog — skip in automated tests")
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

func TestHandleScan_BadJSON(t *testing.T) {
	srv := &Server{}
	req := httptest.NewRequest("POST", "/api/scan", strings.NewReader("not json"))
	w := httptest.NewRecorder()

	srv.handleScan(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandleScan_AlreadyScanning(t *testing.T) {
	srv := &Server{scanning: true}
	body := `{"paths":["/tmp"]}`
	req := httptest.NewRequest("POST", "/api/scan", strings.NewReader(body))
	w := httptest.NewRecorder()

	srv.handleScan(w, req)

	var resp map[string]string
	json.NewDecoder(w.Result().Body).Decode(&resp)
	if resp["status"] != "already_scanning" {
		t.Errorf("status = %q, want already_scanning", resp["status"])
	}
}

func TestHandleScan_StartsScan(t *testing.T) {
	srv := &Server{}
	body := `{"paths":["/tmp"],"min_size":0}`
	req := httptest.NewRequest("POST", "/api/scan", strings.NewReader(body))
	w := httptest.NewRecorder()

	srv.handleScan(w, req)

	var resp map[string]string
	json.NewDecoder(w.Result().Body).Decode(&resp)
	if resp["status"] != "started" {
		t.Errorf("status = %q, want started", resp["status"])
	}
}

// ─── handleStatus ───────────────────────────────────────────────────────────

func TestHandleStatus_NotScanning(t *testing.T) {
	srv := &Server{}
	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()

	srv.handleStatus(w, req)

	var body map[string]interface{}
	json.NewDecoder(w.Result().Body).Decode(&body)
	if body["scanning"] != false {
		t.Error("expected scanning=false")
	}
	if body["has_result"] != false {
		t.Error("expected has_result=false")
	}
}

func TestHandleStatus_WithResult(t *testing.T) {
	srv := &Server{result: &MirrorResult{}}
	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()

	srv.handleStatus(w, req)

	var body map[string]interface{}
	json.NewDecoder(w.Result().Body).Decode(&body)
	if body["has_result"] != true {
		t.Error("expected has_result=true")
	}
}

// ─── handleResult ──────────────────────────────────────────────────────────

func TestHandleResult_NoResult(t *testing.T) {
	srv := &Server{}
	req := httptest.NewRequest("GET", "/api/result", nil)
	w := httptest.NewRecorder()

	srv.handleResult(w, req)

	var body map[string]string
	json.NewDecoder(w.Result().Body).Decode(&body)
	if body["status"] != "no_result" {
		t.Errorf("status = %q, want no_result", body["status"])
	}
}

func TestHandleResult_WithResult(t *testing.T) {
	srv := &Server{result: &MirrorResult{
		TotalScanned:    100,
		TotalDuplicates: 5,
	}}
	req := httptest.NewRequest("GET", "/api/result", nil)
	w := httptest.NewRecorder()

	srv.handleResult(w, req)

	var body MirrorResult
	json.NewDecoder(w.Result().Body).Decode(&body)
	if body.TotalScanned != 100 {
		t.Errorf("TotalScanned = %d, want 100", body.TotalScanned)
	}
}

// ─── handleUI ──────────────────────────────────────────────────────────────

func TestHandleUI(t *testing.T) {
	srv := &Server{}
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	srv.handleUI(w, req)

	if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/html", w.Header().Get("Content-Type"))
	}
	if !strings.Contains(w.Body.String(), "Mirror") {
		t.Error("expected HTML to contain 'Mirror'")
	}
}

// ─── mirrorHTML ──────────────────────────────────────────────────────────

func TestMirrorHTML(t *testing.T) {
	html := mirrorHTML()
	if html == "" {
		t.Fatal("mirrorHTML returned empty string")
	}
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("expected valid HTML document")
	}
	if !strings.Contains(html, "Mirror") {
		t.Error("expected HTML to reference Mirror")
	}
}

// ─── OpenBrowser ────────────────────────────────────────────────────────

func TestServer_OpenBrowser(t *testing.T) {
	t.Skip("OpenBrowser launches real browser — skip in automated tests")
}
