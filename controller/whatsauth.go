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

	// 1. Ambil Profile & Token dari Database
	profile, err := atdb.GetOneDoc[model.Profile](config.Mongoconn, "profile", bson.M{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(model.Response{Response: "Gagal ambil profile DB"})
		return
	}

	// 2. Decode Pesan Masuk (Format PushWa)
	var msg model.PushWaIncoming
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validasi sederhana: Pastikan pesan tidak kosong
	if msg.Message == "" {
		json.NewEncoder(w).Encode(model.Response{Response: "Pesan kosong / Format salah"})
		return
	}

	// 3. Logika Bot
	// (PushWa kadang kirim pesan group juga, kita filter jika perlu)
	if !msg.IsGroup { 
		pesan := strings.TrimSpace(msg.Message)
		pesanLower := strings.ToLower(pesan)
		var replyMsg string

		// A. FITUR SIMPAN
		if strings.HasPrefix(pesanLower, "simpan") || strings.HasPrefix(pesanLower, "catat") {
			keyword := "simpan"
			if strings.HasPrefix(pesanLower, "catat") {
				keyword = "catat"
			}
			
			if len(pesan) <= len(keyword) {
				replyMsg = "Isi catatannya kosong. Coba: 'simpan beli beras'"
			} else {
				content := strings.TrimSpace(pesan[len(keyword):])
				if content == "" {
					replyMsg = "Isi kosong."
				} else {
					// Judul Otomatis
					title := "Catatan"
					if len(content) > 20 {
						title = content[:20] + "..."
					} else {
						title = content
					}

					noteType := "text"
					if strings.Contains(content, "http") {
						noteType = "link"
					}

					newNote := model.Note{
						ID:        primitive.NewObjectID(),
						UserPhone: msg.From, // PushWa pakai "from"
						Title:     title,
						Content:   content,
						Type:      noteType,
						UpdatedAt: time.Now(),
					}
					
					_, err := atdb.InsertNote(newNote)
					if err != nil {
						replyMsg = "Gagal menyimpan database 😢"
					} else {
						replyMsg = "✅ Tersimpan! (" + noteType + ")"
					}
				}
			}

		// B. FITUR LIST
		} else if pesanLower == "list" || pesanLower == "menu" {
			filter := bson.M{"user_phone": msg.From}
			notes, err := atdb.GetAllDoc[[]model.Note](config.Mongoconn, "notes", filter)

			if err != nil || len(notes) == 0 {
				replyMsg = "📭 Belum ada catatan. Ketik 'simpan [isi]'."
			} else {
				var sb strings.Builder
				sb.WriteString("📂 *Catatan Kamu:*\n\n")
				for i, note := range notes {
					icon := "📝"
					if note.Type == "link" { icon = "🔗" }
					preview := note.Content
					if len(preview) > 30 { preview = preview[:30] + "..." }
					item := fmt.Sprintf("%d. *%s* %s\n   _%s_\n", i+1, note.Title, icon, preview)
					sb.WriteString(item)
				}
				replyMsg = sb.String()
			}
		}

		// 4. Kirim Balasan ke PushWa (API Kirim Pesan)
		if replyMsg != "" {
			dataKirim := model.PushWaSend{
				Token:   profile.Token,
				Target:  msg.From,
				Type:    "text",
				Delay:   "1",
				Message: replyMsg,
			}
			// Gunakan fungsi PostJSON yang baru kita buat
			go atapi.PostJSON[interface{}](dataKirim, profile.URLApi)
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