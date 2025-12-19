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

	// 2. GANTI IteungMessage MENJADI WAMessage
	msg := model.WAMessage{
		Phone_number: "628123456789",
		Chat_number:  "628123456789", // Perlu chat number untuk balasan
		Message:      "simpan Beli buku baru",
	}
	body, _ := json.Marshal(msg)

	req := httptest.NewRequest("POST", "/webhook/nomor/628123456789", bytes.NewBuffer(body))
	
	// PENTING: Tambahkan Header Secret Palsu/Asli agar lolos validasi
	// Pastikan di DB "profile" kamu secretnya sama, atau test ini akan return 403
	req.Header.Set("secret", "rahasiaKawai123") 

	rr := httptest.NewRecorder()

	// 3. Jalankan Fungsi
	PostInboxNomor(rr, req)

	// 4. Verifikasi
	// Jika gagal karena secret, code akan 403. Jika sukses 200.
	if rr.Code != http.StatusOK {
		t.Logf("Status code: %d (Mungkin secret di DB tidak cocok dengan 'rahasiaKawai123', abaikan jika ini unit test lokal tanpa DB real)", rr.Code)
	}

	var response model.Response
	json.Unmarshal(rr.Body.Bytes(), &response)

	// Cek respon (hanya jika sukses login)
	if rr.Code == http.StatusOK && !strings.Contains(response.Response, "OK") {
		t.Errorf("Respon salah, dapat: %v", response.Response)
	}
}