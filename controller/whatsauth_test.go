package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv" // Import ini
	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/helper/atdb"
	"github.com/kawai-org/kawai-backend/model"
)

func TestPostInboxNomorSimpan(t *testing.T) {
	// Load .env (Asumsi file ini di folder controller, mundur 1 langkah)
	_ = godotenv.Load("../.env")

	mongoString := os.Getenv("MONGOSTRING")
	if mongoString == "" {
		t.Fatal("MONGOSTRING tidak ditemukan di .env")
	}

	// 1. Inisialisasi Database untuk Testing
	mconn := atdb.DBInfo{
		DBString: mongoString, // AMAN
		DBName:   "kawai_db",
	}
	config.Mongoconn, _ = atdb.MongoConnect(mconn)

	// 2. DATA DUMMY
	msg := model.PushWaIncoming{
		From:    "628123456789",
		Message: "simpan Beli buku baru",
	}
	body, _ := json.Marshal(msg)

	req := httptest.NewRequest("POST", "/webhook/nomor/628123456789", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	// 3. Jalankan Fungsi
	PostInboxNomor(rr, req)

	// 4. Verifikasi
	if rr.Code != http.StatusOK {
		t.Logf("Status code: %d", rr.Code)
	}

	var response model.Response
	json.Unmarshal(rr.Body.Bytes(), &response)

	// Cek respon "OK"
	if rr.Code == http.StatusOK && !strings.Contains(response.Response, "OK") {
		t.Errorf("Respon salah, dapat: %v", response.Response)
	}
}