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
	"github.com/kawai-org/kawai-backend/helper/timeparse"
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

// Helper: Cek apakah text diawali salah satu keyword
func hasPrefixAny(text string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.HasPrefix(text, kw) {
			return true
		}
	}
	return false
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
	sender := msg.From
	if strings.Contains(sender, "@") {
		sender = strings.Split(sender, "@")[0]
	}

	// ==========================================
	// 🔥 PERBAIKAN 2: AUTO-SWITCH IDENTITAS
	// ==========================================
	if sender == "233332956778603" { // ID Laptop
		sender = "6285793766959"     // ID HP Utama
	}

	// Debugging
	fmt.Printf("Raw From: %s | Final User: %s | Pesan: %s\n", msg.From, sender, msg.Message)

	// 2. Audit Log
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

	// Cek apakah pesan HANYA ANGKA?
	targetNo, errNum := strconv.Atoi(pesan)
	isNumberOnly := errNum == nil && targetNo > 0

	// ==========================================
	// A. FITUR SIMPAN / CATAT
	// Keyword: simpan, catat, save
	// ==========================================
	if hasPrefixAny(pesanLower, []string{"simpan", "catat", "save"}) {
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
				UserPhone: sender,
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
	// Keyword: list, menu, tampilkan
	// ==========================================
	} else if hasPrefixAny(pesanLower, []string{"list", "menu", "tampilkan"}) {
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
					
					// 🔥 UPDATE IKON (Termasuk Reminder) 🔥
					icon := "📝"
					if n.Type == "link" { icon = "🔗" }
					if n.Type == "mixed" { icon = "📑" }
					if n.Type == "reminder" { icon = "⏰" } // Ikon baru!

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
	// D. FITUR PENGINGAT (REMINDER + NOTE HYBRID)
	// Keyword: ingatkan, remind
	// ==========================================
	} else if hasPrefixAny(pesanLower, []string{"ingatkan", "remind", "ingat", "ing", "ingt", "rmd", "reminder"}) {
		// Panggil Helper Parsing Waktu
		scheduledTime, title := timeparse.ParseNaturalTime(pesan)
		
		if scheduledTime.IsZero() {
			replyMsg = `🤔 *Waduh, saya kurang paham waktunya.*
Coba ketik waktu yang jelas ya, contohnya:
- _Besok jam 9_ (atau _Bsk jm 9_)
- _5 menit lagi_ (atau _5 mnt lg_)
- _Tgl 17 Agustus_
- _Hari Senin_`
		} else if scheduledTime.Before(time.Now().Add(-1 * time.Minute)) {
			// Validasi Waktu
			replyMsg = "⚠️ Waktu sudah terlewat bos. Coba waktu yang akan datang."
		} else {
			// 1. SIMPAN KE REMINDERS (Untuk Alarm Cron Job)
			newReminder := model.Reminder{
				ID:            primitive.NewObjectID(),
				UserPhone:     sender,
				Title:         title,
				ScheduledTime: scheduledTime,
				Status:        "pending",
			}
			atdb.InsertOneDoc(config.Mongoconn, "reminders", newReminder)

			// 2. SIMPAN KE NOTES (Agar masuk list catatan juga - HYBRID FEATURE)
			noteID := primitive.NewObjectID()
			newNote := model.Note{
				ID:        noteID,
				UserPhone: sender,
				Original:  pesan,
				Content:   title,      // Isi catatan = Judul yang sudah dibersihkan
				Type:      "reminder", // Tipe khusus biar ikonnya beda
				CreatedAt: time.Now(),
			}
			atdb.InsertOneDoc(config.Mongoconn, "notes", newNote)

			// Feedback Lengkap
			timeStr := scheduledTime.Format("Monday, 02 Jan • 15:04 WIB")
			replyMsg = fmt.Sprintf("⏰ *Pengingat & Catatan Diset!*\n\n📌 Topik: %s\n⏳ Waktu: %s\n\n_Data ini juga sudah masuk ke menu List._", title, timeStr)
		}

	// ==========================================
	// E. FITUR BANTUAN (HELP)
	// Keyword: help, info, halo, dll
	// ==========================================
	} else if hasPrefixAny(pesanLower, []string{"help", "bantuan", "halo", "hai", "hi", "p", "info"}) {
		replyMsg = `🤖 *Kawai Assistant Menu*
_Asisten pribadi untuk Catat & Ingat._

1️⃣ *SIMPAN CATATAN*
   Keyword: _Catat, Simpan, Save_
   👉 _Catat ide skripsi bab 1_
   👉 _Simpan Link Zoom https://zoom.us_

2️⃣ *LIHAT DATA*
   Keyword: _List, Menu_
   👉 _List_ (Lihat semua)
   👉 _List Link_ (Khusus link)

3️⃣ *BACA DETAIL*
   Keyword: _(Ketik Nomornya Saja)_
   👉 _1_ (Untuk baca no 1)

4️⃣ *PENGINGAT JADWAL*
   Keyword: _Ingatkan, Ingat, Remind_
   ✅ _Ingatkan Rapat Besok jam 10_
   ✅ _Ingat bayar UKT Lusa_
   ✅ _Ingatkan tgl 17 Agustus_
   ✅ _Ingatkan masak mie 5 menit lagi_
   _(Bisa disingkat: bsk, jm, tgl, lg)_

Selamat mencoba! 🚀`
	}

	// Kirim Balasan
	if replyMsg != "" && profile.Token != "" {
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