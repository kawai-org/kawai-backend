package controller

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Helper internal untuk file ini
func getUserClaims(r *http.Request) jwt.MapClaims {
	claims, _ := r.Context().Value("user").(jwt.MapClaims)
	return claims
}

// 1. API UNTUK USER: Ambil Catatan Sendiri
func GetMyNotes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	claims := getUserClaims(r)
	// Pastikan user_phone ada di token, convert ke string
	userPhone, ok := claims["user_phone"].(string)
	if !ok {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Token invalid"})
		return
	}

	filter := bson.M{"user_phone": userPhone}
	opts := options.Find().SetSort(bson.M{"created_at": -1})

	cursor, err := config.Mongoconn.Collection("notes").Find(context.TODO(), filter, opts)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Database Error"})
		return
	}

	// Inisialisasi slice kosong agar return [] bukan null jika kosong
	notes := []model.Note{}
	if cursor != nil {
		cursor.All(context.TODO(), &notes)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   notes,
	})
}

// 2. API UNTUK ADMIN: Lihat Semua User
func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims := getUserClaims(r)
	role, _ := claims["role"].(string)

	// Proteksi Sederhana: Hanya admin yang boleh
	if role != "admin" && role != "super_admin" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Akses Ditolak"})
		return
	}

	cursor, err := config.Mongoconn.Collection("users").Find(context.TODO(), bson.M{})
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Database Error"})
		return
	}
	
	users := []model.User{}
	if cursor != nil {
		cursor.All(context.TODO(), &users)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"total":  len(users),
		"data":   users,
	})
}