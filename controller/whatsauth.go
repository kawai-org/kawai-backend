package controller

import (
	"encoding/json"
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

	// --- 1. LOGIKA URL PARAMETER (PENGGANTI :nomorwa) ---
	// URL: /webhook/nomor/628123456789
	// Kita potong prefix "/webhook/nomor/" untuk dapat nomornya
	pathNomor := strings.TrimPrefix(r.URL.Path, "/webhook/nomor/")
	
	// Validasi sederhana: kalau kosong, berarti URL salah
	if pathNomor == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(model.Response{Response: "URL Salah: Nomor tidak ditemukan di URL"})
		return
	}

	// --- 2. CEK SECRET (KEAMANAN) ---
	secret := r.Header.Get("secret")
	if secret == "" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(model.Response{Response: "Secret Kosong"})
		return
	}

	profile, err := atdb.GetOneDoc[model.Profile](config.Mongoconn, "profile", bson.M{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(model.Response{Response: "Gagal ambil profile DB"})
		return
	}

	if secret != profile.Secret {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(model.Response{Response: "Secret Salah"})
		return
	}

	// --- 3. DECODE PESAN (STRUCT LENGKAP) ---
	var msg model.WAMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	// Validasi tambahan: Pastikan nomor di URL sama dengan nomor tujuan chat (opsional tapi bagus)
	// msg.Chat_server biasanya berisi nomor WA server kita
	
	// --- 4. LOGIKA BOT ---
	if !msg.Is_group {
		pesan := strings.ToLower(strings.TrimSpace(msg.Message))
		var replyMsg string

		if strings.HasPrefix(pesan, "simpan") || strings.HasPrefix(pesan, "catat") {
			content := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(pesan, "simpan"), "catat"))
			if content == "" {
				replyMsg = "Isi catatannya kosong bos."
			} else {
				// Simpan ke DB
				newNote := model.Note{
					ID:        primitive.NewObjectID(),
					UserPhone: msg.Phone_number, // Nomor pengirim
					Title:     "Catatan WhatsApp",
					Content:   content,
					UpdatedAt: time.Now(),
				}
				atdb.InsertNote(newNote)
				replyMsg = "Oke, catatan berhasil disimpan! ✨"
			}
		} else {
            // Logic default
            return 
        }

		// Kirim Balasan
		dt := model.TextMessage{
			To:       msg.Chat_number,
			IsGroup:  false,
			Messages: replyMsg,
		}
		atapi.PostStructWithToken[model.Response]("Token", profile.Token, dt, profile.URLApiText)
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