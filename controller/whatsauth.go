package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/helper/atapi"
	"github.com/kawai-org/kawai-backend/helper/atdb"
	"github.com/kawai-org/kawai-backend/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func GetHome(respw http.ResponseWriter, req *http.Request) {
	resp := model.Response{Response: "It works! Kawai Assistant is Online."}
	WriteJSON(respw, http.StatusOK, resp)
}

func PostInboxNomor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Decode Pesan Masuk (Sesuai Format Baru PushWa)
	var msg model.PushWaIncoming
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		fmt.Println("Error Decode JSON:", err)
		json.NewEncoder(w).Encode(model.Response{Response: "Bad Request"})
		return
	}

	// Debugging Log: Cek apakah data masuk
	fmt.Printf("Pesan Masuk: Dari=%s, Isi=%s\n", msg.From, msg.Message)

	// Validasi: Pastikan pengirim dan pesan ada
	if msg.From == "" || msg.Message == "" {
		json.NewEncoder(w).Encode(model.Response{Response: "Data Kosong (From/Message)"})
		return
	}

	// 2. Ambil Token API dari Database
	profile, err := atdb.GetOneDoc[model.Profile](config.Mongoconn, "profile", bson.M{})
	if err != nil {
		fmt.Println("Error DB Profile:", err)
		json.NewEncoder(w).Encode(model.Response{Response: "Error DB"})
		return
	}

	// 3. Logika Bot
	pesan := strings.TrimSpace(msg.Message)
	pesanLower := strings.ToLower(pesan)
	var replyMsg string

	// --- A. FITUR SIMPAN ---
	if strings.HasPrefix(pesanLower, "simpan") || strings.HasPrefix(pesanLower, "catat") {
		keyword := "simpan"
		if strings.HasPrefix(pesanLower, "catat") { keyword = "catat" }
		
		// Ambil isi setelah kata kunci
		content := strings.TrimSpace(pesan[len(keyword):])
		
		if content == "" {
			replyMsg = "Isi catatannya kosong bos. Coba: 'simpan beli beras'"
		} else {
			// Simpan ke DB
			newNote := model.Note{
				ID:        primitive.NewObjectID(),
				UserPhone: msg.From, // Pakai msg.From dari PushWa
				Title:     "Catatan",
				Content:   content,
				UpdatedAt: time.Now(),
			}
			_, err := atdb.InsertNote(newNote)
			if err != nil {
				replyMsg = "Gagal menyimpan ke database 😢"
				fmt.Println("Error InsertNote:", err)
			} else {
				replyMsg = "✅ Tersimpan!"
			}
		}

	// --- B. FITUR LIST ---
	} else if pesanLower == "list" || pesanLower == "menu" {
		filter := bson.M{"user_phone": msg.From}
		notes, err := atdb.GetAllDoc[[]model.Note](config.Mongoconn, "notes", filter)
		
		if err != nil || len(notes) == 0 {
			replyMsg = "📭 Belum ada catatan. Yuk ketik 'simpan [sesuatu]'"
		} else {
			var sb strings.Builder
			sb.WriteString("📂 *Daftar Catatan Kamu:*\n\n")
			for i, note := range notes {
				item := fmt.Sprintf("%d. %s\n", i+1, note.Content)
				sb.WriteString(item)
			}
			replyMsg = sb.String()
		}
	} else {
        // Balasan default (opsional, untuk memastikan bot hidup)
        // replyMsg = "Halo! Ketik 'simpan [isi]' atau 'list'." 
    }

	// 4. Kirim Balasan ke API PushWa (Tanpa 'go' agar Vercel menunggu)
	if replyMsg != "" {
		fmt.Println("Mengirim balasan ke:", msg.From)
		kirim := model.PushWaSend{
			Token:   profile.Token,
			Target:  msg.From, // Balas ke nomor pengirim
			Type:    "text",
			Delay:   "1",
			Message: replyMsg,
		}
		// Kirim POST ke PushWa
		atapi.PostJSON[interface{}](kirim, profile.URLApi)
	}

	json.NewEncoder(w).Encode(model.Response{Response: "OK"})
}

func GetNewToken(respw http.ResponseWriter, req *http.Request) {
	resp := model.Response{Response: "Feature Refresh Token coming soon"}
	WriteJSON(respw, http.StatusOK, resp)
}

func NotFound(respw http.ResponseWriter, req *http.Request) {
	resp := model.Response{Response: "Error 404: Path Not Found"}
	WriteJSON(respw, http.StatusNotFound, resp)
}