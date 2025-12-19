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

	// --- 1. LOGIKA URL PARAMETER ---
	pathNomor := strings.TrimPrefix(r.URL.Path, "/webhook/nomor/")
	if pathNomor == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(model.Response{Response: "URL Salah: Nomor tidak ditemukan"})
		return
	}

	// --- 2. CEK SECRET ---
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

	// --- 3. DECODE PESAN ---
	var msg model.WAMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// --- 4. LOGIKA BOT ---
	if !msg.Is_group {
		pesan := strings.TrimSpace(msg.Message)
		pesanLower := strings.ToLower(pesan)
		var replyMsg string

		// A. FITUR SIMPAN CATATAN
		if strings.HasPrefix(pesanLower, "simpan") || strings.HasPrefix(pesanLower, "catat") {
			// Hapus kata perintah di depan
			keyword := "simpan"
			if strings.HasPrefix(pesanLower, "catat") {
				keyword = "catat"
			}
			// Ambil isi konten (substring setelah keyword)
			// Menggunakan len(keyword) agar akurat memotongnya
			if len(pesan) <= len(keyword) {
				replyMsg = "Isi catatannya kosong. Coba: 'simpan beli beras'"
			} else {
				content := strings.TrimSpace(pesan[len(keyword):])
				
				if content == "" {
					replyMsg = "Isi catatannya kosong bos."
				} else {
					// Buat Judul Otomatis (20 Karakter pertama + "...")
					title := "Catatan"
					if len(content) > 20 {
						title = content[:20] + "..."
					} else {
						title = content
					}

					// Cek Tipe (Text atau Link)
					noteType := "text"
					if strings.Contains(content, "http") {
						noteType = "link"
					}

					newNote := model.Note{
						ID:        primitive.NewObjectID(),
						UserPhone: msg.Phone_number,
						Title:     title,
						Content:   content,
						Type:      noteType,
						UpdatedAt: time.Now(),
					}
					
					_, err := atdb.InsertNote(newNote)
					if err != nil {
						replyMsg = "Gagal menyimpan ke database 😢"
					} else {
						replyMsg = "✅ Tersimpan! (" + noteType + ")"
					}
				}
			}

		// B. FITUR TAMPILKAN LIST
		} else if pesanLower == "list" || pesanLower == "tampilkan" || pesanLower == "catatan" {
			filter := bson.M{"user_phone": msg.Phone_number}
			// Mengambil semua catatan user
			notes, err := atdb.GetAllDoc[[]model.Note](config.Mongoconn, "notes", filter)

			if err != nil || len(notes) == 0 {
				replyMsg = "📭 Kamu belum punya catatan. Yuk ketik 'simpan [sesuatu]'"
			} else {
				var sb strings.Builder
				sb.WriteString("📂 *Daftar Catatan Kamu:*\n\n")
				
				for i, note := range notes {
					// Format Tampilan: 1. [Judul] (Type)
					icon := "📝"
					if note.Type == "link" {
						icon = "🔗"
					}
					// Ambil cuplikan isi
					preview := note.Content
					if len(preview) > 30 {
						preview = preview[:30] + "..."
					}
					
					item := fmt.Sprintf("%d. *%s* %s\n   _%s_\n", i+1, note.Title, icon, preview)
					sb.WriteString(item)
				}
				sb.WriteString("\nKetik 'simpan' untuk menambah lagi.")
				replyMsg = sb.String()
			}

		// C. MENU DEFAULT
		} else if pesanLower == "halo" || pesanLower == "menu" {
			replyMsg = "Halo! Saya Kawai Bot 🤖.\n\nPerintah:\n- *simpan [isi]* : Mencatat sesuatu\n- *list* : Lihat semua catatan"
		} else {
			// Diam saja jika perintah tidak dikenal, atau balas default (opsional)
			return
		}

		// Kirim Balasan ke API WhatsAuth
		if replyMsg != "" {
			dt := model.TextMessage{
				To:       msg.Chat_number,
				IsGroup:  false,
				Messages: replyMsg,
			}
			// Pastikan pakai Goroutine agar tidak blocking (opsional tapi disarankan)
			go atapi.PostStructWithToken[model.Response]("Token", profile.Token, dt, profile.URLApiText)
		}
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