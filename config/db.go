package config

import (
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
)

var MongoString string = os.Getenv("MONGOSTRING")

// Jangan panggil atdb di sini untuk menghindari circular import
var Mongoconn *mongo.Database
var ErrorMongoconn error

// Tambahkan fungsi yang dibutuhkan route agar tidak error
func SetAccessControlHeaders(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Secret")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return true
	}
	return false
}

func SetEnv() {
	// Kosongkan sementara untuk keperluan inisialisasi environment jika ada
}