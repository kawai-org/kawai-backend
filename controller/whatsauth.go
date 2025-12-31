package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/helper/atapi"
	"github.com/kawai-org/kawai-backend/helper/atdb"
	"github.com/kawai-org/kawai-backend/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Helper: Ekstrak URL
func extractURL(text string) string {
	re := regexp.MustCompile(`https?://[^\s]+`)
	return re.FindString(text)
}

// Helper: Ekstrak Hashtags
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
	resp := model.Response{Response: "Kawai Assistant Online."}
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

	if msg.From == "" || msg.Message == "" {
		json.NewEncoder(w).Encode(model.Response{Response: "Empty Data"})
		return
	}

	// ==========================================
	// 🔥 PERBAIKAN 1: SANITASI NOMOR HP
	// ==========================================
	// Ambil nomor HP bersih (buang @s.whatsapp.net / @lid jika ada)
	sender := msg.From
	if strings.Contains(sender, "@") {
		sender = strings.Split(sender, "@")[0]
	}

	// ==========================================
	// 🔥 PERBAIKAN 2: AUTO-SWITCH IDENTITAS (SOLUSI FINAL)
	// ==========================================
	// Jika pesan datang dari ID Laptop (2333...), paksa ubah jadi ID HP Utama (628...).
	// Tujuannya agar database tetap satu dan sinkron.
	if sender == "233332956778603" { // ID Laptop
		sender = "6285793766959"     // ID HP Utama
	}

	// Debugging: Pastikan 'Final User' sekarang selalu 628...
	fmt.Printf("Raw From: %s | Final User: %s | Pesan: %s\n", msg.From, sender, msg.Message)

	// 2. Audit Log (Tetap simpan history dengan ID yang sudah disatukan)
	atdb.InsertOneDoc(config.Mongoconn, "message_logs", model.MessageLog{
		ID:         primitive.NewObjectID(),
		From:       sender,
		Message:    msg.Message,
		ReceivedAt: time.Now(),
	})

	profile, _ := atdb.GetOneDoc[model.BotProfile](config.Mongoconn, "profile", bson.M{})

	// 3. Logic Processing
	pesan := strings.TrimSpace(msg.Message)
	pesanLower := strings.ToLower(pesan)
	var replyMsg string

	// Cek apakah pesan HANYA ANGKA? (Fitur Shortcut Detail)
	targetNo, errNum := strconv.Atoi(pesan)
	isNumberOnly := errNum == nil && targetNo > 0

	// ==========================================
	// A. FITUR SIMPAN / CATAT
	// ==========================================
	if strings.HasPrefix(pesanLower, "simpan") || strings.HasPrefix(pesanLower, "catat") || strings.HasPrefix(pesanLower, "save") {
		parts := strings.Fields(pesan)
		
		if len(parts) > 1 {
			content := strings.TrimSpace(strings.Join(parts[1:], " "))
			foundURL := extractURL(content)
			foundTags := extractTags(content)

			// Tentukan Tipe
			noteType := "text"
			if foundURL != "" {
				if content == foundURL {
					noteType = "link"
				} else {
					noteType = "mixed"
				}
			}

			// Insert Note
			noteID := primitive.NewObjectID()
			atdb.InsertOneDoc(config.Mongoconn, "notes", model.Note{
				ID:        noteID,
				UserPhone: sender, // Sudah otomatis jadi 628...
				Original:  pesan,
				Content:   content,
				Type:      noteType,
				CreatedAt: time.Now(),
			})

			// Insert Link
			if foundURL != "" {
				linkTitle := content
				if linkTitle == "" || linkTitle == foundURL {
					linkTitle = foundURL
				}
				if len(linkTitle) > 50 { linkTitle = linkTitle[:50] + "..." }

				atdb.InsertOneDoc(config.Mongoconn, "links", model.Link{
					ID:        primitive.NewObjectID(),
					NoteID:    noteID,
					UserPhone: sender,
					URL:       foundURL,
					Title:     linkTitle,
					CreatedAt: time.Now(),
				})
			}

			// Insert Tags
			if len(foundTags) > 0 {
				for _, t := range foundTags {
					atdb.InsertOneDoc(config.Mongoconn, "tags", model.Tag{
						ID:        primitive.NewObjectID(),
						NoteID:    noteID,
						TagName:   t,
						UserPhone: sender,
					})
				}
			}

			replyMsg = fmt.Sprintf("✅ Tersimpan!\n\n💡 *Tips:* Ketik *List* untuk melihat catatanmu.")
			if noteType == "link" {
				replyMsg += "\n(Lain kali, tambahkan judul biar gak lupa ya. Contoh: *Catat Link Zoom https://...*)"
			}

		} else {
			replyMsg = "Format salah bos.\nKetik: *Catat [isi catatan]*"
		}

	// ==========================================
	// B. FITUR LIST (List & List Link)
	// ==========================================
	} else if strings.HasPrefix(pesanLower, "list") || strings.HasPrefix(pesanLower, "menu") {
		filter := bson.M{"user_phone": sender}
		
		page := 1
		limit := 10
		parts := strings.Fields(pesan)
		if len(parts) > 1 {
			if p, err := strconv.Atoi(parts[len(parts)-1]); err == nil && p > 0 {
				page = p
			}
		}
		skip := int64((page - 1) * limit)
		opts := options.Find().SetLimit(int64(limit)).SetSkip(skip).SetSort(bson.M{"created_at": -1})

		// --- Mode List Link ---
		if strings.Contains(pesanLower, "link") {
			cursor, _ := config.Mongoconn.Collection("links").Find(context.TODO(), filter, opts)
			var links []model.Link
			if cursor != nil { cursor.All(context.TODO(), &links) }

			if len(links) > 0 {
				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("🔗 *Koleksi Link (Hal %d)*\n", page))
				for i, l := range links {
					nomor := (page-1)*limit + i + 1
					judul := l.Title
					if judul == "" { judul = l.URL }
					
					sb.WriteString(fmt.Sprintf("\n%d. *%s*\n   %s", nomor, judul, l.URL))
				}
				sb.WriteString(fmt.Sprintf("\n\n_Ketik *List Link %d* untuk halaman berikutnya._", page+1))
				replyMsg = sb.String()
			} else {
				replyMsg = "Belum ada link tersimpan."
			}

		// --- Mode List Catatan (Default) ---
		} else {
			cursor, _ := config.Mongoconn.Collection("notes").Find(context.TODO(), filter, opts)
			var notes []model.Note
			if cursor != nil { cursor.All(context.TODO(), &notes) }

			if len(notes) > 0 {
				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("📂 *Catatan (Hal %d)*\n", page))
				for i, n := range notes {
					nomor := (page-1)*limit + i + 1
					
					display := n.Content
					if len(display) > 35 { display = display[:35] + "..." }
					
					icon := "📝"
					if n.Type == "link" { icon = "🔗" } else if n.Type == "mixed" { icon = "📑" }

					sb.WriteString(fmt.Sprintf("\n%d. %s %s", nomor, icon, display))
				}
				sb.WriteString("\n\n👉 *Ketik nomornya saja* untuk baca detail. (Contoh: ketik *1*)")
				sb.WriteString(fmt.Sprintf("\n👉 Ketik *List %d* untuk halaman berikutnya.", page+1))
				replyMsg = sb.String()
			} else {
				replyMsg = "Belum ada catatan. Yuk ketik *Catat [isi]*"
			}
		}

	// ==========================================
	// C. FITUR BACA DETAIL (Shortcut Angka)
	// ==========================================
	} else if isNumberOnly {
		skip := int64(targetNo - 1)
		opts := options.FindOne().SetSkip(skip).SetSort(bson.M{"created_at": -1})
		filter := bson.M{"user_phone": sender}

		var note model.Note
		errDB := config.Mongoconn.Collection("notes").FindOne(context.TODO(), filter, opts).Decode(&note)
		
		if errDB == nil {
			dateStr := note.CreatedAt.Format("02 Jan 15:04")
			
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("📂 *DETAIL NO. %d*\n", targetNo))
			sb.WriteString(fmt.Sprintf("📅 %s | Tipe: %s\n", dateStr, note.Type))
			sb.WriteString("----------------------\n")
			sb.WriteString(note.Content) 
			sb.WriteString("\n----------------------")
			
			if note.Type == "link" || note.Type == "mixed" {
				url := extractURL(note.Content)
				if url != "" {
					sb.WriteString(fmt.Sprintf("\n🔗 *Link:* %s", url))
				}
			}
			replyMsg = sb.String()
		} else {
			replyMsg = fmt.Sprintf("❌ Data nomor %d tidak ditemukan.", targetNo)
		}

	// ==========================================
	// D. FITUR BANTUAN
	// ==========================================
	} else if strings.HasPrefix(pesanLower, "help") || strings.HasPrefix(pesanLower, "bantuan") || strings.HasPrefix(pesanLower, "halo") || strings.HasPrefix(pesanLower, "info") {
		replyMsg = `🤖 *Halo! Saya Kawai Assistant.*
Saya bantu kamu catat hal penting & simpan link biar gak hilang ditelan chat.

*Cara Pakai:*
1️⃣ *Simpan Catatan*
   Ketik: _Catat beli gorengan 5 biji_
2️⃣ *Simpan Link*
   Ketik: _Catat Materi Kuliah https://google.com_
3️⃣ *Lihat Data*
   Ketik: _List_ atau _List Link_
4️⃣ *Baca Detail*
   Cukup ketik *Nomor*-nya saja (misal: _1_)

Selamat mencoba! 🚀`
	}

	// Kirim Balasan
	if replyMsg != "" && profile.Token != "" {
		kirim := model.PushWaSend{
			Token:   profile.Token,
			Target:  msg.From, // Target tetap msg.From asli agar sampai ke device yang benar
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