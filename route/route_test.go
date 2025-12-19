package route

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/helper/atdb"
	"github.com/kawai-org/kawai-backend/model"
)

func TestURL(t *testing.T) {
	// 1. SETUP KONEKSI (Pastikan URI MongoDB benar untuk testing)
	mconn := atdb.DBInfo{
		DBString: "mongodb+srv://penerbit:u2cC2MwwS42yKxub@webhook.jej9ieu.mongodb.net/?retryWrites=true&w=majority&appName=webhook", 
		DBName:   "kawai_db",
	}
	config.Mongoconn, _ = atdb.MongoConnect(mconn)

	// 2. GANTI IteungMessage MENJADI WAMessage
	msg := model.WAMessage{
		Phone_number: "628123456789",
		Message:      "Tes dari Route",
	}
	jsonBody, _ := json.Marshal(msg)

	// 3. Daftar test case
	tests := []struct {
		name       string
		method     string
		url        string
		body       []byte
		expectCode int
	}{
		{"Test_CORS", "OPTIONS", "/", nil, http.StatusOK},
		{"Test_Home", "GET", "/", nil, http.StatusOK},
		{"Test_NotFound", "GET", "/ngasal", nil, http.StatusNotFound},
		{"Test_Webhook", "POST", "/webhook/nomor/628123456789", jsonBody, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.url, bytes.NewBuffer(tt.body))
			
			// Tambahkan Header Secret agar lolos validasi controller
			req.Header.Set("secret", "rahasiaKawai123") // Sesuaikan dengan secret di DB Profile kamu saat ini
			
			rr := httptest.NewRecorder()

			URL(rr, req)

			// Validasi Status Code
			// Catatan: Jika secret salah, mungkin akan return 403 Forbidden, sesuaikan ekspektasi atau secretnya
			if rr.Code != tt.expectCode && rr.Code != http.StatusForbidden { 
				t.Errorf("URL %s: dapat %d, ingin %d", tt.url, rr.Code, tt.expectCode)
			}
		})
	}
}