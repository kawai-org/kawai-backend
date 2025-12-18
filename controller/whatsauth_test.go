package controller

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

func TestPostInboxNomor(t *testing.T) {
	// 1. SETUP: Pastikan koneksi DB ada (sama seperti di atdb_test)
	mconn := atdb.DBInfo{
		DBString: "mongodb+srv://penerbit:u2cC2MwwS42yKxub@webhook.jej9ieu.mongodb.net/?retryWrites=true&w=majority&appName=webhook", 
		DBName:   "kawai_db",
	}
	config.Mongoconn, _ = atdb.MongoConnect(mconn)

	// 2. DATA: Buat simulasi pesan dari WhatsApp/WhatsAuth
	msg := model.IteungMessage{
		Phone:   "628123456789",
		Alias:   "Kawai User",
		Message: "Tes kirim pesan ke bot",
	}
	body, _ := json.Marshal(msg)

	// 3. REQUEST: Buat request palsu
	req, _ := http.NewRequest("POST", "/webhook/nomor/628123456789", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	// 4. EKSEKUSI: Panggil fungsi controller
	handler := http.HandlerFunc(PostInboxNomor)
	handler.ServeHTTP(rr, req)

	// 5. VALIDASI: Cek responnya
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code salah: dapat %v ingin %v", status, http.StatusOK)
	}

	var response model.Response
	json.Unmarshal(rr.Body.Bytes(), &response)
	if response.Response != "Pesan diterima oleh Kawai" {
		t.Errorf("Respon pesan salah: %v", response.Response)
	}
}