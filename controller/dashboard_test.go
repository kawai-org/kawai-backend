package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kawai-org/kawai-backend/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestGetMyNotes(t *testing.T) {
	// Setup mtest untuk Mock Database
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Success Get Notes", func(mt *mtest.T) {
		// 1. Sambungkan Mock DB ke Config
		config.Mongoconn = mt.Client.Database("kawai_db")

		// 2. Siapkan Data Palsu (Mock Response dari DB)
		// Kita pakai bson.M agar tidak error "unkeyed fields"
		mt.AddMockResponses(mtest.CreateCursorResponse(
			1,
			"kawai_db.notes",
			mtest.FirstBatch,
			bson.D{
				{Key: "user_phone", Value: "62812345"},
				{Key: "content", Value: "Test Catatan 1"},
				{Key: "type", Value: "text"},
			},
		))

		// 3. Buat Request Palsu
		req := httptest.NewRequest("GET", "/api/dashboard/notes", nil)
		w := httptest.NewRecorder()

		// 4. INJECT CONTEXT (Penting!)
		// Kita memanipulasi request seolah-olah sudah lolos dari MiddlewareAuth
		// dan membawa data user (claims)
		claims := jwt.MapClaims{
			"user_phone": "62812345",
			"role":       "user",
		}
		// "user" adalah key yang sama dengan yang ada di route.go/MiddlewareAuth
		ctx := context.WithValue(req.Context(), "user", claims)
		req = req.WithContext(ctx)

		// 5. PANGGIL CONTROLLER (Ini yang tadi kurang)
		GetMyNotes(w, req)

		// 6. Cek Hasil
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}