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
	// 1. SETUP KONEKSI
	mconn := atdb.DBInfo{
		DBString: "mongodb+srv://penerbit:u2cC2MwwS42yKxub@webhook.jej9ieu.mongodb.net/?retryWrites=true&w=majority&appName=webhook", 
		DBName:   "kawai_db",
	}
	config.Mongoconn, _ = atdb.MongoConnect(mconn)

	// 2. Setup Pesan Dummy Format PushWa
	msg := model.PushWaIncoming{
		From:    "628123456789",
		Message: "Tes dari Route",
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
		// Webhook endpoint
		{"Test_Webhook", "POST", "/webhook/nomor/628123456789", jsonBody, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.url, bytes.NewBuffer(tt.body))
			rr := httptest.NewRecorder()

			URL(rr, req)

			if rr.Code != tt.expectCode {
				t.Errorf("URL %s: dapat %d, ingin %d", tt.url, rr.Code, tt.expectCode)
			}
		})
	}
}