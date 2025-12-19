package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/helper/atapi"
	"github.com/kawai-org/kawai-backend/helper/atdb"
	"github.com/kawai-org/kawai-backend/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Helper: Ekstrak URL dari text
func extractURL(text string) string {
	re := regexp.MustCompile(`https?://[^\s]+`)
	return re.FindString(text)
}

// Helper: Ekstrak Hashtags (contoh: #penting)
func extractTags(text string) []string {
	re := regexp.MustCompile(`#\w+`)
	return re.FindAllString(text, -1)
}

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

	// 1. Decode Payload
	var msg model.PushWaIncoming
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		json.NewEncoder(w).Encode(model.Response{Response: "Bad Request"})
		return
	}

	// Validasi Data
	if msg.From == "" || msg.Message == "" {
		json.NewEncoder(w).Encode(model.Response{Response: "Data Kosong"})
		return
	}

	fmt.Printf("Pesan Masuk: %s | Isi: %s\n", msg.From, msg.Message)

	// 2. AUDIT TRAIL: Simpan ke 'message_logs'
	// PENTING: Kita hapus 'go func()' sementara agar kita bisa lihat errornya langsung di log Vercel jika gagal
	logMsg := model.MessageLog{
		ID:         primitive.NewObjectID(),
		From:       msg.From,
		Message:    msg.Message,
		ReceivedAt: time.Now(),
	}
	_, errLog := atdb.InsertOneDoc(config.Mongoconn, "message_logs", logMsg)
	if errLog != nil {
		fmt.Println("GAGAL SIMPAN MESSAGE LOG:", errLog) // Ini akan muncul di log Vercel
	}

	// 3. Ambil Profile Bot
	// Pastikan nama collection di DB kamu 'profile' atau 'bot_profiles'. 
	// Kita pakai 'profile' dulu karena itu yang sudah ada datanya.
	profile, err := atdb.GetOneDoc[model.BotProfile](config.Mongoconn, "profile", bson.M{})
	if err != nil {
		fmt.Println("Error DB Profile:", err)
	}

	// 4. SMART LOGIC PROCESSING
	pesan := strings.TrimSpace(msg.Message)
	pesanLower := strings.ToLower(pesan)
	var replyMsg string

	// --- A. FITUR SIMPAN (SMART PARSER) ---
	if strings.HasPrefix(pesanLower, "simpan") || strings.HasPrefix(pesanLower, "catat") || strings.HasPrefix(pesanLower, "save") {
		// Hapus keyword di depan
		parts := strings.Fields(pesan)
		if len(parts) > 1 {
			// Gabungkan kembali sisa pesan
			content := strings.TrimSpace(strings.Join(parts[1:], " "))
			
			// Analisa Konten
			foundURL := extractURL(content)
			foundTags := extractTags(content)
			
			noteType := "text"
			if foundURL != "" {
				if content == foundURL {
					noteType = "link" // Isinya cuma link
				} else {
					noteType = "mixed" // Teks campur link
				}
			}

			// Simpan ke Notes (Tabel 5)
			noteID := primitive.NewObjectID()
			newNote := model.Note{
				ID:        noteID,
				UserPhone: msg.From,
				Original:  pesan,
				Content:   content,
				Type:      noteType,
				CreatedAt: time.Now(),
			}
			// Insert ke collection 'notes'
			atdb.InsertOneDoc(config.Mongoconn, "notes", newNote)

			// Simpan Relasi Link (Tabel 6)
			if foundURL != "" {
				newLink := model.Link{
					ID:        primitive.NewObjectID(),
					NoteID:    noteID, // Relational Key
					UserPhone: msg.From,
					URL:       foundURL,
					CreatedAt: time.Now(),
				}
				// Insert ke collection 'links' - Otomatis terbuat jika belum ada
				atdb.InsertOneDoc(config.Mongoconn, "links", newLink)
			}

			// Simpan Relasi Tags (Tabel 7)
			if len(foundTags) > 0 {
				for _, t := range foundTags {
					newTag := model.Tag{
						ID:        primitive.NewObjectID(),
						NoteID:    noteID, // Relational Key
						TagName:   t,
						UserPhone: msg.From,
					}
					// Insert ke collection 'tags'
					atdb.InsertOneDoc(config.Mongoconn, "tags", newTag)
				}
			}

			replyMsg = fmt.Sprintf("✅ Tersimpan!\nType: %s\nTags: %v", noteType, len(foundTags))

		} else {
			replyMsg = "Ketik: *Simpan [isi catatan]*"
		}

	// --- B. FITUR LIST (RETRIEVE) ---
	} else if strings.HasPrefix(pesanLower, "list") || strings.HasPrefix(pesanLower, "menu") {
		filter := bson.M{"user_phone": msg.From}
		
		// Jika user minta "list link", ambil dari tabel links
		if strings.Contains(pesanLower, "link") {
			links, _ := atdb.GetAllDoc[[]model.Link](config.Mongoconn, "links", filter)
			if len(links) > 0 {
				var sb strings.Builder
				sb.WriteString("🔗 *List Link Kamu:*\n")
				for i, l := range links {
					sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, l.URL))
				}
				replyMsg = sb.String()
			} else {
				replyMsg = "Belum ada link tersimpan."
			}
		} else {
			// Default ambil dari notes
			notes, _ := atdb.GetAllDoc[[]model.Note](config.Mongoconn, "notes", filter)
			if len(notes) > 0 {
				var sb strings.Builder
				sb.WriteString("📝 *Catatan Kamu:*\n")
				for i, n := range notes {
					preview := n.Content
					if len(preview) > 30 { preview = preview[:30] + "..." }
					sb.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, n.Type, preview))
				}
				replyMsg = sb.String()
			} else {
				replyMsg = "Belum ada catatan."
			}
		}
	}

	// 5. Kirim Balasan (Tanpa 'go' routine untuk Vercel)
	if replyMsg != "" {
		fmt.Println("Mengirim balasan ke:", msg.From)
		kirim := model.PushWaSend{
			Token:   profile.Token,
			Target:  msg.From,
			Type:    "text",
			Delay:   "1",
			Message: replyMsg,
		}
		atapi.PostJSON[interface{}](kirim, profile.URLApi)
	}

	json.NewEncoder(w).Encode(model.Response{Response: "OK"})
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(model.Response{Response: "404 Not Found"})
}