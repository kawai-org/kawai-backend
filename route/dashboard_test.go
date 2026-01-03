package route

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kawai-org/kawai-backend/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

// Helper: Buat Token Palsu
func createTestToken(role, phone string) string {
	os.Setenv("JWT_SECRET", "testsecret")
	claims := jwt.MapClaims{
		"user_phone": phone,
		"role":       role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, _ := token.SignedString([]byte("testsecret"))
	return t
}

func TestAdminLogin_Fail(t *testing.T) {
	// 1. Setup Mock DB
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Login Fail Mock", func(mt *mtest.T) {
		// PENTING: Inject Mock DB ke Global Config supaya tidak PANIC
		config.Mongoconn = mt.Client.Database("kawai_db")

		// Mock: Database tidak menemukan user (return kosong)
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "kawai_db.users", mtest.FirstBatch, bson.D{}))

		reqBody, _ := json.Marshal(map[string]string{
			"username": "08123456789",
			"password": "wrongpassword",
		})

		req, _ := http.NewRequest("POST", "/api/admin/login", bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()

		// Sekarang aman dipanggil karena config.Mongoconn sudah ada isinya (Mock)
		URL(w, req)

		// Cek: Harusnya statusnya OK (200) tapi isinya pesan error, atau status error
		// Tergantung implementasi controller kamu. Biasanya kita cek body-nya.
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)

		// Pastikan responsenya bukan sukses login
		if resp["status"] == "success" {
			t.Errorf("Harusnya gagal login, tapi malah sukses")
		}
	})
}

func TestGetMyNotes_Success(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Get Notes Route", func(mt *mtest.T) {
		config.Mongoconn = mt.Client.Database("kawai_db")

		// Mock: Database mengembalikan 1 catatan
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "kawai_db.notes", mtest.FirstBatch, bson.D{
			{Key: "user_phone", Value: "62812345"},
			{Key: "content", Value: "Cek Route"},
		}))

		token := createTestToken("user", "62812345")
		req, _ := http.NewRequest("GET", "/api/dashboard/notes", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		URL(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %v", w.Code)
		}
	})
}

// Tambahkan test coverage untuk middleware (akses tanpa token)
func TestUnauthorized_Access(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/dashboard/notes", nil)
	w := httptest.NewRecorder()
	
	// MiddlewareAuth dipanggil di dalam URL switch case
	URL(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Harusnya 401 Unauthorized, dapet %v", w.Code)
	}
}