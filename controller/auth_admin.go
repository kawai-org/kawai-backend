package controller

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/helper/atdb"
	"github.com/kawai-org/kawai-backend/model"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)


type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoginAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
    
	var input LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Invalid Body"})
		return
	}

	// 1. Cari Admin di Database
	var admin model.User
	filter := bson.M{"phone_number": input.Username, "role": "admin"} // Username pakai NoHP/Email
	
    // Gunakan helper get doc kamu
    admin, err := atdb.GetOneDoc[model.User](config.Mongoconn, "users", filter)
    if err != nil {
        json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Admin tidak ditemukan / Bukan Admin"})
        return
    }

	// 2. Cek Password (Hash vs Plain)
    // Di database field password harusnya: "$2a$10$X7..." (Hasil Hash)
	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(input.Password))
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Password Salah"})
		return
	}

	// 3. Generate JWT Token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.MapClaims{
		"id":   admin.ID,
		"role": "admin",
		"exp":  expirationTime.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	// 4. Return Token ke Frontend
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"token":  tokenString,
        "role":   "admin", 
	})
}