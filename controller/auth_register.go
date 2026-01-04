package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// Input struct untuk Register
type RegisterInput struct {
	PhoneNumber string `json:"phone_number"`
	Name        string `json:"name"`
	Password    string `json:"password"`
	SecretCode  string `json:"secret_code"` // Opsional: Untuk jadi Admin
}

func Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var input RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Invalid Body"})
		return
	}

	// 1. Validasi Input
	if input.PhoneNumber == "" || input.Password == "" || input.Name == "" {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Data tidak lengkap"})
		return
	}

	// 2. Cek apakah User sudah ada?
	var existingUser model.User
	filter := bson.M{"phone_number": input.PhoneNumber}
	err := config.Mongoconn.Collection("users").FindOne(context.TODO(), filter).Decode(&existingUser)
	if err == nil {
		// Jika tidak error, berarti user ketemu (Duplicate)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Nomor HP sudah terdaftar"})
		return
	}

	// 3. Hash Password (Wajib!)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Gagal hash password"})
		return
	}

	// 4. Tentukan Role
	role := "user" // Default
	// Cek environment variable untuk kode rahasia admin
	adminSecret := os.Getenv("ADMIN_SECRET_KEY") 
	if adminSecret != "" && input.SecretCode == adminSecret {
		role = "admin" // Auto jadi admin jika kode benar
	}

	// 5. Siapkan Data
	newUser := model.User{
		ID:          primitive.NewObjectID(),
		PhoneNumber: input.PhoneNumber,
		Name:        input.Name,
		Role:        role,
		Password:    string(hashedPassword), // Simpan yang sudah di-hash
		CreatedAt:   time.Now(),
	}

	// 6. Simpan ke Database
	_, err = config.Mongoconn.Collection("users").InsertOne(context.TODO(), newUser)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Gagal simpan ke database"})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": "success", 
		"msg": "Registrasi Berhasil sebagai " + role,
	})
}