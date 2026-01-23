package atapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test untuk fungsi PostJSON
func TestPostJSON(t *testing.T) {
	// 1. Buat Server Palsu
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	payload := map[string]string{"foo": "bar"}

	// PERBAIKAN 1: Argumen DITUKAR (payload dulu, baru URL) sesuai api.go
	// PERBAIKAN 2: Tangkap 3 nilai kembalian (code, result, err)
	code, result, err := PostJSON[map[string]interface{}](payload, ts.URL)
	
	if err != nil {
		t.Errorf("PostJSON error: %v", err)
	}

	// PERBAIKAN 3: Cek statusCode langsung dari variabel 'code'
	if code != http.StatusOK {
		t.Errorf("Expected 200, got %d", code)
	}
	
	// Cek isi balasan
	if result["status"] != "ok" {
		t.Errorf("Expected result status 'ok', got %v", result["status"])
	}
}

// Test untuk fungsi Get
func TestGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	// PERBAIKAN: Tangkap 3 nilai kembalian
	code, _, err := Get[map[string]interface{}](ts.URL)
	
	if err != nil {
		t.Errorf("Get error: %v", err)
	}
	
	if code != http.StatusOK {
		t.Errorf("Expected 200, got %d", code)
	}
}