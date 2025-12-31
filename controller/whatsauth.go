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
	"github.com/kawai-org/kawai-backend/helper/gdrive" 
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

// Helper: Cek keyword
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

// --- LOGIKA UTAMA BOT ---
func PostInboxNomor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Decode Payload JSON dari WhatsApp
	var msg model.PushWaIncoming
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		json.NewEncoder(w).Encode(model.Response{Response: "Bad Request"})
		return
	}

	if msg.From == "" {
		json.NewEncoder(w).Encode(model.Response{Response: "Empty Data"})
		return
	}

	// 2. Sanitasi Nomor HP & Switch Identitas
	sender := msg.From
	if strings.Contains(sender, "@") {
		sender = strings.Split(sender, "@")[0]
	}
	if sender == "233332956778603" { // ID Laptop -> ID HP
		sender = "6285793766959"
	}

	// 3. Simpan Log Chat (Audit Trail)
	atdb.InsertOneDoc(config.Mongoconn, "message_logs", model.MessageLog{
		ID:         primitive.NewObjectID(),
		From:       sender,
		Message:    msg.Message,
		ReceivedAt: time.Now(),
	})

	profile, _ := atdb.GetOneDoc[model.BotProfile](config.Mongoconn, "profile", bson.M{})

	// Persiapan Variabel
	pesan := strings.TrimSpace(msg.Message)
	pesanLower := strings.ToLower(pesan)
	var replyMsg string

	// Cek apakah pesan HANYA ANGKA? (Untuk shortcut baca detail)
	targetNo, errNum := strconv.Atoi(pesan)
	isNumberOnly := errNum == nil && targetNo > 0

	// ==========================================
	// F. FITUR UPLOAD FILE KE GOOGLE DRIVE (BARU!) 📁
	// Logika: Jika ada FileUrl, berarti user kirim file
	// ==========================================
	if msg.FileUrl != "" {
		// Kirim notifikasi loading biar user gak bingung
		kirimLoading := model.PushWaSend{
			Token:   profile.Token,
			Target:  msg.From,
			Type:    "text",
			Delay:   "0",
			Message: "⏳ Sedang mengupload file ke Drive...",
		}
		atapi.PostJSON[interface{}](kirimLoading, profile.URLApi)

		// 1. Tentukan Nama File
		fileName := pesan // Gunakan caption sebagai nama file
		if fileName == "" {
			fileName = fmt.Sprintf("WA-Upload-%d", time.Now().Unix())
		}
		
		// Pastikan ekstensi file ada (penting biar bisa dibuka di Drive)
		if !strings.Contains(fileName, ".") {
			ext := ".file"
			if msg.MimeType == "application/pdf" { ext = ".pdf" }
			if strings.Contains(msg.MimeType, "image") { ext = ".jpg" }
			if strings.Contains(msg.MimeType, "word") { ext = ".docx" }
			if strings.Contains(msg.MimeType, "sheet") { ext = ".xlsx" }
			fileName += ext
		}

		// 2. Download File dari WhatsApp Server
		respFile, errDown := http.Get(msg.FileUrl)
		if errDown != nil {
			replyMsg = "❌ Gagal mendownload file dari WhatsApp."
		} else {
			defer respFile.Body.Close()

			// 3. Upload ke Google Drive via Helper gdrive
			fileID, webLink, errUp := gdrive.UploadToDrive(sender, fileName, respFile.Body)
			
			if errUp != nil {
				// Cek terminal Vercel/Lokal untuk detail error
				fmt.Printf("Error Upload GDrive: %v\n", errUp)
				replyMsg = "❌ Gagal upload ke Drive. Cek apakah Token Google sudah diset di database?"
			} else {
				// 4. Sukses! Simpan metadata ke tabel 'drive_files'
				newFile := model.DriveFile{
					ID:           primitive.NewObjectID(),
					UserPhone:    sender,
					GoogleFileID: fileID,
					FileName:     fileName,
					MimeType:     msg.MimeType,
					DriveLink:    webLink,
					UploadedAt:   time.Now(),
				}
				atdb.InsertOneDoc(config.Mongoconn, "drive_files", newFile)

				replyMsg = fmt.Sprintf("✅ *File Tersimpan di Drive!*\n\n📂 Nama: %s\n🔗 Link: %s\n\n_File ini aman di folder KAWAI_FILES_", fileName, webLink)
			}
		}

	// ==========================================
	// A. FITUR SIMPAN / CATAT
	// ==========================================
	} else if hasPrefixAny(pesanLower, []string{"simpan", "catat", "save"}) {
		parts := strings.Fields(pesan)
		if len(parts) > 1 {
			content := strings.TrimSpace(strings.Join(parts[1:], " "))
			foundURL := extractURL(content)
			foundTags := extractTags(content)

			noteType := "text"
			if foundURL != "" {
				if content == foundURL { noteType = "link" } else { noteType = "mixed" }
			}

			noteID := primitive.NewObjectID()
			atdb.InsertOneDoc(config.Mongoconn, "notes", model.Note{
				ID: noteID, UserPhone: sender, Original: pesan, Content: content, Type: noteType, CreatedAt: time.Now(),
			})

			// Insert Link terpisah
			if foundURL != "" {
				linkTitle := content
				if linkTitle == "" || linkTitle == foundURL { linkTitle = foundURL }
				if len(linkTitle) > 50 { linkTitle = linkTitle[:50] + "..." }
				atdb.InsertOneDoc(config.Mongoconn, "links", model.Link{
					ID: primitive.NewObjectID(), NoteID: noteID, UserPhone: sender, URL: foundURL, Title: linkTitle, CreatedAt: time.Now(),
				})
			}
			// Insert Tags
			for _, t := range foundTags {
				atdb.InsertOneDoc(config.Mongoconn, "tags", model.Tag{
					ID: primitive.NewObjectID(), NoteID: noteID, TagName: t, UserPhone: sender,
				})
			}
			replyMsg = "✅ Tersimpan! Ketik *List* untuk melihat."
		} else {
			replyMsg = "Format salah bos.\nKetik: *Catat [isi catatan]*"
		}

	// ==========================================
	// B. FITUR LIST (Dengan Ikon Baru)
	// ==========================================
	} else if hasPrefixAny(pesanLower, []string{"list", "menu", "tampilkan"}) {
		filter := bson.M{"user_phone": sender}
		page := 1
		limit := 10
		
		parts := strings.Fields(pesan)
		if len(parts) > 1 {
			if p, err := strconv.Atoi(parts[len(parts)-1]); err == nil && p > 0 { page = p }
		}
		opts := options.Find().SetLimit(int64(limit)).SetSkip(int64((page - 1) * limit)).SetSort(bson.M{"created_at": -1})

		// Cek mode List Link
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
		} else {
			// List Catatan Biasa
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
					
					// Ikon Cerdas
					icon := "📝"
					if n.Type == "link" { icon = "🔗" }
					if n.Type == "mixed" { icon = "📑" }
					if n.Type == "reminder" { icon = "⏰" }

					sb.WriteString(fmt.Sprintf("\n%d. %s %s", nomor, icon, display))
				}
				sb.WriteString(fmt.Sprintf("\n\n👉 *Ketik nomornya* untuk detail.  (Contoh: ketik *1*)\n👉 Ketik *List %d* untuk halaman berikutnya.", page+1))
				replyMsg = sb.String()
			} else {
				replyMsg = "Belum ada catatan. Yuk ketik *Catat [isi]*"
			}
		}

	// ==========================================
	// C. FITUR BACA DETAIL
	// ==========================================
	} else if isNumberOnly {
		skip := int64(targetNo - 1)
		opts := options.FindOne().SetSkip(skip).SetSort(bson.M{"created_at": -1})
		var note model.Note
		errDB := config.Mongoconn.Collection("notes").FindOne(context.TODO(), bson.M{"user_phone": sender}, opts).Decode(&note)
		
		if errDB == nil {
			dateStr := note.CreatedAt.Format("02 Jan 15:04")
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("📂 *DETAIL NO. %d*\n📅 %s\n----------------------\n%s\n----------------------", targetNo, dateStr, note.Content))
			// Tampilkan link jika ada
			if note.Type == "link" || note.Type == "mixed" {
				url := extractURL(note.Content)
				if url != "" { sb.WriteString(fmt.Sprintf("\n🔗 *Link:* %s", url)) }
			}
			replyMsg = sb.String()
		} else {
			replyMsg = fmt.Sprintf("❌ Data nomor %d tidak ditemukan.", targetNo)
		}

	// ==========================================
	// D. FITUR PENGINGAT (Hybrid + Timeparse Baru)
	// ==========================================
	} else if hasPrefixAny(pesanLower, []string{"ingatkan", "remind", "ingat", "ing", "ingt"}) {
		scheduledTime, title := timeparse.ParseNaturalTime(pesan)
		
		if scheduledTime.IsZero() {
			replyMsg = "🤔 *Waduh, saya kurang paham waktunya.*
Coba ketik waktu yang jelas ya, contohnya:
- _Besok jam 9_ (atau _Bsk jm 9_)
- _5 menit lagi_ (atau _5 mnt lg_)
- _Tgl 17 Agustus_
- _Hari Senin_"
		} else if scheduledTime.Before(time.Now()) {
			replyMsg = "⚠️ Waktu sudah lewat."
		} else {
			// Simpan ke Reminders (Alarm)
			atdb.InsertOneDoc(config.Mongoconn, "reminders", model.Reminder{
				ID: primitive.NewObjectID(), UserPhone: sender, Title: title, ScheduledTime: scheduledTime, Status: "pending",
			})
			// Simpan ke Notes (Arsip)
			atdb.InsertOneDoc(config.Mongoconn, "notes", model.Note{
				ID: primitive.NewObjectID(), UserPhone: sender, Original: pesan, Content: title, Type: "reminder", CreatedAt: time.Now(),
			})
			
			timeStr := scheduledTime.Format("02 Jan • 15:04 WIB")
			replyMsg = fmt.Sprintf("⏰ *Pengingat Diset!*\n\n📌 Topik: %s\n⏳ Waktu: %s", title, timeStr)
		}

	// ==========================================
	// E. FITUR BANTUAN
	// ==========================================
	} else if hasPrefixAny(pesanLower, []string{"help", "bantuan", "halo", "menu", "p", "hi", "hai", "info"}) {
		replyMsg = `🤖 *Kawai Assistant Menu*

1️⃣ *UPLOAD KE DRIVE* 📂
   _Kirim File (PDF/Foto) + Caption_
   👉 Otomatis masuk folder KAWAI_FILES!

2️⃣ *SIMPAN CATATAN* 📝
   Keyword: _Catat, Simpan_
   👉 _Catat ide skripsi bab 1_
   👉 _Simpan Link Zoom https://zoom.us_

3️⃣ *LIHAT DATA* 📂
   Keyword: _List, Menu_
   👉 _List_ (Lihat semua)
   👉 _List Link_ (Khusus link)

4️⃣ *BACA DETAIL*
   Keyword: _(Ketik Nomornya Saja)_
   👉 _1_ (Untuk baca no 1)

5️⃣ *PENGINGAT* ⏰
   Keyword: _Ingatkan, Ingat, Remind_
   ✅ _Ingatkan Rapat Besok jam 10_
   ✅ _Ingat bayar UKT Lusa 09.30_
   ✅ _Ingatkan tgl 17 Agustus 13:00_
   ✅ _Ingatkan masak mie 5 menit lagi_

Selamat mencoba! 😊`
	}
	

	// Kirim Balasan Final
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