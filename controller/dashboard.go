package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Helper internal untuk file ini
func getUserClaims(r *http.Request) jwt.MapClaims {
	claims, _ := r.Context().Value("user").(jwt.MapClaims)
	return claims
}

// @Summary Get My Notes
// @Description Ambil semua catatan user yang sedang login
// @Tags User
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/dashboard/notes [get]
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
	if role != "admin"  {
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


// Helper: Ambil ID dari URL (Contoh: /api/notes/123 -> return 123)
func getIDFromURL(r *http.Request, prefix string) (primitive.ObjectID, error) {
	idStr := strings.TrimPrefix(r.URL.Path, prefix)
	return primitive.ObjectIDFromHex(idStr)
}

// 3. EDIT CATATAN (Fix Typo)
func UpdateNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	claims := getUserClaims(r)
	userPhone := claims["user_phone"].(string)

	// Ambil ID dari URL
	noteID, err := getIDFromURL(r, "/api/notes/")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "ID URL salah"})
		return
	}

	// Decode Body (Isi baru)
	var input struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Body JSON salah"})
		return
	}

	// QUERY PENTING: Filter by ID DAN UserPhone (Supaya gak edit punya orang lain)
	filter := bson.M{"_id": noteID, "user_phone": userPhone}
	update := bson.M{"$set": bson.M{"content": input.Content}}

	res, err := config.Mongoconn.Collection("notes").UpdateOne(context.TODO(), filter, update)
	if err != nil || res.MatchedCount == 0 {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Gagal update (Catatan tidak ditemukan/bukan milikmu)"})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success", "msg": "Catatan diperbarui"})
}

// 4. HAPUS CATATAN
func DeleteNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	claims := getUserClaims(r)
	userPhone := claims["user_phone"].(string)

	noteID, err := getIDFromURL(r, "/api/notes/")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "ID URL salah"})
		return
	}

	// Filter by ID & Phone (Security Check)
	filter := bson.M{"_id": noteID, "user_phone": userPhone}
	
	res, err := config.Mongoconn.Collection("notes").DeleteOne(context.TODO(), filter)
	if err != nil || res.DeletedCount == 0 {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Gagal hapus (Data tidak ditemukan)"})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success", "msg": "Catatan dihapus"})
}

// 5. LIST REMINDERS (PENGINGAT)
func GetReminders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	claims := getUserClaims(r)
	userPhone := claims["user_phone"].(string)

	filter := bson.M{"user_phone": userPhone}
	opts := options.Find().SetSort(bson.M{"scheduled_time": 1}) // Urut dari yang terdekat

	cursor, _ := config.Mongoconn.Collection("reminders").Find(context.TODO(), filter, opts)
	
	reminders := []model.Reminder{} // Pakai slice kosong biar gak null
	if cursor != nil {
		cursor.All(context.TODO(), &reminders)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": reminders})
}

// 6. UPDATE REMINDER (Ganti Jam/Tanggal)
func UpdateReminder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	claims := getUserClaims(r)
	userPhone := claims["user_phone"].(string)

	id, err := getIDFromURL(r, "/api/reminders/")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "ID salah"})
		return
	}

	// Input bisa ganti Judul atau Waktu
	var input struct {
		Title string    `json:"title"`
		Time  time.Time `json:"scheduled_time"`
	}
	json.NewDecoder(r.Body).Decode(&input)

	// Siapkan update query dinamis
	updateFields := bson.M{}
	if input.Title != "" { updateFields["title"] = input.Title }
	if !input.Time.IsZero() { updateFields["scheduled_time"] = input.Time }

	if len(updateFields) == 0 {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Tidak ada data yang diubah"})
		return
	}

	filter := bson.M{"_id": id, "user_phone": userPhone}
	res, _ := config.Mongoconn.Collection("reminders").UpdateOne(context.TODO(), filter, bson.M{"$set": updateFields})

	if res.MatchedCount == 0 {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Gagal update"})
	} else {
		json.NewEncoder(w).Encode(map[string]string{"status": "success", "msg": "Pengingat diupdate"})
	}
}