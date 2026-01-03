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
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("Login Fail Mock", func(mt *mtest.T) {
		config.Mongoconn = mt.Client.Database("kawai_db")
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "kawai_db.users", mtest.FirstBatch, bson.D{}))

		reqBody, _ := json.Marshal(map[string]string{
			"username": "08123456789",
			"password": "wrongpassword",
		})
		req, _ := http.NewRequest("POST", "/api/admin/login", bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()
		URL(w, req)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["status"] == "success" {
			t.Errorf("Harusnya gagal login, tapi malah sukses")
		}
	})
}

func TestGetMyNotes_Success(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("Get Notes Route", func(mt *mtest.T) {
		config.Mongoconn = mt.Client.Database("kawai_db")
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

func TestUnauthorized_Access(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/dashboard/notes", nil)
	w := httptest.NewRecorder()
	URL(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Harusnya 401 Unauthorized, dapet %v", w.Code)
	}
}

// PERBAIKAN: TestURL dengan Mocking
func TestURL(t *testing.T) {
    // Kita bungkus TestURL dalam mtest agar bisa handle koneksi DB "siluman"
    mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

    mt.Run("Test URL Routes", func(mt *mtest.T) {
        // Inject Mock DB
        config.Mongoconn = mt.Client.Database("kawai_db")
        
        // Mock Response Default (Agar InsertOneDoc tidak error Timeout)
        // Kita kasih response sukses untuk operasi apapun yang mungkin dipanggil
        mt.AddMockResponses(mtest.CreateSuccessResponse())
        
        t.Run("Test_NotFound", func(t *testing.T) {
            req, _ := http.NewRequest("GET", "/ngasal", nil)
            w := httptest.NewRecorder()
            URL(w, req)
            if w.Code != http.StatusNotFound {
                t.Errorf("URL /ngasal: dapat %v, ingin %v", w.Code, http.StatusNotFound)
            }
        })
        
        // Anda bisa tambahkan sub-test lain disini jika ada
    })
}