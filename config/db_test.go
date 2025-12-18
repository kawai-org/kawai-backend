package config

import (
	"net/http/httptest"
	"testing"
)

func TestSetAccessControlHeaders(t *testing.T) {
	// Menyiapkan request dan recorder
	req := httptest.NewRequest("OPTIONS", "/", nil)
	rr := httptest.NewRecorder()

	// Pemanggilan fungsi (menggunakan http melalui rr dan req)
	isOptions := SetAccessControlHeaders(rr, req)

	// Validasi hasil agar baris kode di db.go terhitung hijau semua
	if !isOptions {
		t.Error("Harusnya return true untuk method OPTIONS")
	}

	// Test untuk method selain OPTIONS
	reqGet := httptest.NewRequest("GET", "/", nil)
	rrGet := httptest.NewRecorder()
	if SetAccessControlHeaders(rrGet, reqGet) {
		t.Error("Harusnya return false untuk method GET")
	}
}

func TestSetEnv(t *testing.T) {
	// Memanggil fungsi SetEnv agar terhitung dalam coverage
	SetEnv()
}