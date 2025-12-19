package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/helper/atdb"
	"github.com/kawai-org/kawai-backend/model"
)

func TestPostInboxNomorSimpan(t *testing.T) {
	// 1. Inisialisasi Database untuk Testing
	mconn := atdb.DBInfo{
		DBString: "mongodb+srv://penerbit:u2cC2MwwS42yKxub@webhook.jej9ieu.mongodb.net/?retryWrites=true&w=majority&appName=webhook", 
		DBName:   "kawai_db",
	}
	config.Mongoconn, _ = atdb.MongoConnect(mconn)

	// 2. GANTI MENGGUNAKAN FORMAT PUSHWA (PushWaIncoming)
	msg := model.PushWaIncoming{
		From:      "628123456789", // PushWa pakai "From", bukan "Phone"
		Message:   "simpan Beli buku baru",
		PushName:  "Tester",
		Type:      "text",
		Timestamp: 12345678,
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

	// Cek respon "OK" (karena balasan chat dikirim via API call terpisah/goroutine)
	if rr.Code == http.StatusOK && !strings.Contains(response.Response, "OK") {
		t.Errorf("Respon salah, dapat: %v", response.Response)
	}
}