package controller

import (
	"context" // Tambahan untuk fungsi Find DB
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv" // Tambahan untuk parsing halaman (List 1, List 2)
	"strings"
	"time"

	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/helper/atapi"
	"github.com/kawai-org/kawai-backend/helper/atdb"
	"github.com/kawai-org/kawai-backend/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options" // Tambahan untuk pagination
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
	logMsg := model.MessageLog{
		ID:         primitive.NewObjectID(),
		From:       msg.From,
		Message:    msg.Message,
		ReceivedAt: time.Now(),
	}
	// InsertLog
	_, errLog := atdb.InsertOneDoc(config.Mongoconn, "message_logs", logMsg)
	if errLog != nil {
		fmt.Println("GAGAL SIMPAN MESSAGE LOG:", errLog)
	}

	// 3. Ambil Profile Bot
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

			// Simpan ke Notes
			noteID := primitive.NewObjectID()
			newNote := model.Note{
				ID:        noteID,
				UserPhone: msg.From,
				Original:  pesan,
				Content:   content,
				Type:      noteType,
				CreatedAt: time.Now(),
			}
			atdb.InsertOneDoc(config.Mongoconn, "notes", newNote)

			// Simpan Relasi Link
			if foundURL != "" {
				newLink := model.Link{
					ID:        primitive.NewObjectID(),
					NoteID:    noteID, // Relational Key
					UserPhone: msg.From,
					URL:       foundURL,
					CreatedAt: time.Now(),
				}
				atdb.InsertOneDoc(config.Mongoconn, "links", newLink)
			}

			// Simpan Relasi Tags
			if len(foundTags) > 0 {
				for _, t := range foundTags {
					newTag := model.Tag{
						ID:        primitive.NewObjectID(),
						NoteID:    noteID, // Relational Key
						TagName:   t,
						UserPhone: msg.From,
					}
					atdb.InsertOneDoc(config.Mongoconn, "tags", newTag)
				}
			}

			replyMsg = fmt.Sprintf("✅ Tersimpan!\nType: %s\nTags: %v", noteType, len(foundTags))

		} else {
			replyMsg = "Ketik: *Simpan [isi catatan]*"
		}

	// --- B. FITUR LIST (RETRIEVE dengan PAGINATION & FORMATTING) ---
	} else if strings.HasPrefix(pesanLower, "list") || strings.HasPrefix(pesanLower, "tampilkan") || strings.HasPrefix(pesanLower, "menu") {
		filter := bson.M{"user_phone": msg.From}

		// 1. Logika Pagination
		page := 1
		limit := 10 // Menampilkan 10 item per halaman
		
		parts := strings.Fields(pesan)
		// Cek apakah ada angka di akhir pesan (misal: "List 2")
		if len(parts) > 1 {
			if p, err := strconv.Atoi(parts[len(parts)-1]); err == nil && p > 0 {
				page = p
			}
		}
		
		skip := int64((page - 1) * limit)
		// Opsi Query: Limit 10, Skip sesuai halaman, Urutkan dari yang terbaru
		findOptions := options.Find().SetLimit(int64(limit)).SetSkip(skip).SetSort(bson.M{"created_at": -1})

		// 2. Mode Tampilkan Link
		if strings.Contains(pesanLower, "link") {
			cursor, err := config.Mongoconn.Collection("links").Find(context.TODO(), filter, findOptions)
			var links []model.Link
			if err == nil {
				cursor.All(context.TODO(), &links)
			}

			if len(links) > 0 {
				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("🔗 *Link Kamu (Hal %d)*\n", page))
				sb.WriteString("------------------\n")
				
				for i, l := range links {
					nomor := (page-1)*limit + i + 1
					sb.WriteString(fmt.Sprintf("%d. %s\n", nomor, l.URL))
				}
				sb.WriteString("------------------\n")
				sb.WriteString(fmt.Sprintf("Ketik *List Link %d* untuk next.", page+1))
				replyMsg = sb.String()
			} else {
				if page == 1 {
					replyMsg = "Belum ada link tersimpan."
				} else {
					replyMsg = "Halaman ini kosong."
				}
			}

		// 3. Mode Tampilkan Semua Catatan (Default)
		} else {
			cursor, err := config.Mongoconn.Collection("notes").Find(context.TODO(), filter, findOptions)
			var notes []model.Note
			if err == nil {
				cursor.All(context.TODO(), &notes)
			}

			if len(notes) > 0 {
				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("📂 *Catatan (Hal %d)*\n", page))
				sb.WriteString("------------------\n")

				for i, n := range notes {
					nomor := (page-1)*limit + i + 1
					
					// JUDUL DINAMIS: Ambil konten, potong jika kepanjangan
					displayTitle := n.Content
					if len(displayTitle) > 40 {
						displayTitle = displayTitle[:40] + "..."
					}
					
					// ICON SESUAI TIPE
					icon := "📝"
					if n.Type == "link" { icon = "🔗" }
					if n.Type == "mixed" { icon = "📑" }

					// FORMAT TANGGAL: 19 Des 15:00
					dateStr := n.CreatedAt.Format("02 Jan 15:04")

					// Format Pesan:
					// 1. 📝 Judul Singkat...
					//    (Tanggal)
					sb.WriteString(fmt.Sprintf("%d. %s *%s*\n   📅 %s\n\n", nomor, icon, displayTitle, dateStr))
				}
				
				sb.WriteString("------------------\n")
				sb.WriteString(fmt.Sprintf("Ketik *List %d* untuk next.", page+1))
				replyMsg = sb.String()
			} else {
				if page == 1 {
					replyMsg = "Belum ada catatan. Yuk *Catat* sesuatu!"
				} else {
					replyMsg = "Sudah tidak ada data di halaman ini."
				}
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