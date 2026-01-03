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

// Helper: Ambil Claims dari Context
func getUserClaims(r *http.Request) jwt.MapClaims {
	claims, _ := r.Context().Value("user").(jwt.MapClaims)
	return claims
}

// Helper: Ambil ID dari URL
func getIDFromURL(r *http.Request, prefix string) (primitive.ObjectID, error) {
	idStr := strings.TrimPrefix(r.URL.Path, prefix)
	return primitive.ObjectIDFromHex(idStr)
}

// ==========================================
// BAGIAN 1: CRUD USER (NOTES & REMINDERS)
// ==========================================

// --- NOTES ---

func GetMyNotes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	claims := getUserClaims(r)
	userPhone, _ := claims["user_phone"].(string)

	// Ambil parameter ?search=... dari URL
	keyword := r.URL.Query().Get("search")

	// Filter Dasar: Harus milik user ini
	filter := bson.M{"user_phone": userPhone}

	// Jika ada keyword search, tambahkan filter regex (case insensitive)
	if keyword != "" {
		filter["content"] = bson.M{
			"$regex":   keyword,
			"$options": "i", // "i" artinya tidak peduli huruf besar/kecil
		}
	}

	opts := options.Find().SetSort(bson.M{"created_at": -1})

	cursor, err := config.Mongoconn.Collection("notes").Find(context.TODO(), filter, opts)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "msg": "DB Error"})
		return
	}

	notes := []model.Note{}
	if cursor != nil {
		cursor.All(context.TODO(), &notes)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   notes,
	})
}

func UpdateNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	claims := getUserClaims(r)
	userPhone, _ := claims["user_phone"].(string)

	noteID, err := getIDFromURL(r, "/api/notes/")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "ID URL Salah"})
		return
	}

	var input struct { Content string `json:"content"` }
	json.NewDecoder(r.Body).Decode(&input)

	filter := bson.M{"_id": noteID, "user_phone": userPhone}
	update := bson.M{"$set": bson.M{"content": input.Content}}

	res, _ := config.Mongoconn.Collection("notes").UpdateOne(context.TODO(), filter, update)
	if res.MatchedCount == 0 {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Gagal update"})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "msg": "Updated"})
}

func DeleteNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	claims := getUserClaims(r)
	userPhone, _ := claims["user_phone"].(string)

	noteID, err := getIDFromURL(r, "/api/notes/")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "ID URL Salah"})
		return
	}

	filter := bson.M{"_id": noteID, "user_phone": userPhone}
	res, _ := config.Mongoconn.Collection("notes").DeleteOne(context.TODO(), filter)
	
	if res.DeletedCount == 0 {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Gagal hapus"})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "msg": "Deleted"})
}

// --- REMINDERS ---

func GetReminders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	claims := getUserClaims(r)
	userPhone, _ := claims["user_phone"].(string)

	keyword := r.URL.Query().Get("search")

	filter := bson.M{"user_phone": userPhone}

	// Logika Search untuk Reminder (Cari berdasarkan Judul)
	if keyword != "" {
		filter["title"] = bson.M{
			"$regex":   keyword,
			"$options": "i",
		}
	}

	opts := options.Find().SetSort(bson.M{"scheduled_time": 1})

	cursor, _ := config.Mongoconn.Collection("reminders").Find(context.TODO(), filter, opts)
	
	reminders := []model.Reminder{}
	if cursor != nil {
		cursor.All(context.TODO(), &reminders)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": reminders})
}

func UpdateReminder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	claims := getUserClaims(r)
	userPhone, _ := claims["user_phone"].(string)

	id, err := getIDFromURL(r, "/api/reminders/")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "ID Salah"})
		return
	}

	var input struct {
		Title string    `json:"title"`
		Time  time.Time `json:"scheduled_time"`
	}
	json.NewDecoder(r.Body).Decode(&input)

	updateFields := bson.M{}
	if input.Title != "" { updateFields["title"] = input.Title }
	if !input.Time.IsZero() { updateFields["scheduled_time"] = input.Time }

	filter := bson.M{"_id": id, "user_phone": userPhone}
	res, _ := config.Mongoconn.Collection("reminders").UpdateOne(context.TODO(), filter, bson.M{"$set": updateFields})

	if res.MatchedCount == 0 {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Gagal update"})
	} else {
		json.NewEncoder(w).Encode(map[string]string{"status": "success", "msg": "Updated"})
	}
}

// [BARU] DELETE REMINDER
func DeleteReminder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	claims := getUserClaims(r)
	userPhone, _ := claims["user_phone"].(string)

	id, err := getIDFromURL(r, "/api/reminders/")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "ID Salah"})
		return
	}

	filter := bson.M{"_id": id, "user_phone": userPhone}
	res, _ := config.Mongoconn.Collection("reminders").DeleteOne(context.TODO(), filter)

	if res.DeletedCount == 0 {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Gagal hapus"})
	} else {
		json.NewEncoder(w).Encode(map[string]string{"status": "success", "msg": "Deleted"})
	}
}

// ==========================================
// BAGIAN 2: ADMIN SIMPLE (Tukang Intip & Satpam)
// ==========================================

// Helper Cek Admin
func isAdmin(r *http.Request) bool {
	claims := getUserClaims(r)
	role, ok := claims["role"].(string)
	return ok && role == "admin" // Hapus super_admin, simple aja
}

// 1. LIHAT SEMUA USER (Untuk Audit/Monitoring)
func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !isAdmin(r) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "Khusus Admin"})
		return
	}

	cursor, _ := config.Mongoconn.Collection("users").Find(context.TODO(), bson.M{})
	users := []model.User{}
	if cursor != nil { cursor.All(context.TODO(), &users) }

	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "total": len(users), "data": users})
}

// 2. STATISTIK SIMPLE (Total User, Total Notes)
func GetSystemStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !isAdmin(r) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	userCount, _ := config.Mongoconn.Collection("users").CountDocuments(context.TODO(), bson.M{})
	noteCount, _ := config.Mongoconn.Collection("notes").CountDocuments(context.TODO(), bson.M{})
	reminderCount, _ := config.Mongoconn.Collection("reminders").CountDocuments(context.TODO(), bson.M{})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"stats": map[string]int64{
			"total_users":     userCount,
			"total_notes":     noteCount,
			"total_reminders": reminderCount,
		},
	})
}

// 3. BAN USER (Blokir User Iseng)
func BanUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !isAdmin(r) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var input struct {
		Phone  string `json:"phone_number"`
		Action string `json:"action"` // "ban" atau "unban"
	}
	json.NewDecoder(r.Body).Decode(&input)

	status := "active"
	if input.Action == "ban" {
		status = "banned"
	}

	filter := bson.M{"phone_number": input.Phone}
	update := bson.M{"$set": bson.M{"status": status}}
	
	res, _ := config.Mongoconn.Collection("users").UpdateOne(context.TODO(), filter, update)
	
	if res.MatchedCount == 0 {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "msg": "User tidak ditemukan"})
	} else {
		json.NewEncoder(w).Encode(map[string]string{"status": "success", "msg": "Status user diupdate: " + status})
	}
}