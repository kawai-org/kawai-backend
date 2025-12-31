package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

	
	//  DEBUGGING MODE: LIHAT RAW JSON
	
	bodyBytes, _ := io.ReadAll(r.Body)

	fmt.Printf("RAW JSON: %s\n", string(bodyBytes)) 
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// 1. Decode Payload
	var msg model.PushWaIncoming
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		json.NewEncoder(w).Encode(model.Response{Response: "Bad Request"})
		return
	}

	if msg.From == "" {
		json.NewEncoder(w).Encode(model.Response{Response: "Empty Data"})
		return
	}

	// 2. Sanitasi Nomor HP
	sender := msg.From
	if strings.Contains(sender, "@") {
		sender = strings.Split(sender, "@")[0]
	}
	if sender == "233332956778603" { 
		sender = "6285793766959"
	}

	// 3. Simpan Log Chat
	atdb.InsertOneDoc(config.Mongoconn, "message_logs", model.MessageLog{
		ID:         primitive.NewObjectID(),
		From:       sender,
		Message:    msg.Message,
		ReceivedAt: time.Now(),
	})

	profile, _ := atdb.GetOneDoc[model.BotProfile](config.Mongoconn, "profile", bson.M{})

	pesan := strings.TrimSpace(msg.Message)
	pesanLower := strings.ToLower(pesan)
	var replyMsg string

	targetNo, errNum := strconv.Atoi(pesan)
	isNumberOnly := errNum == nil && targetNo > 0


	// FITUR F: UPLOAD FILE 

	finalFileUrl := msg.FileUrl
	if finalFileUrl == "" { finalFileUrl = msg.Url }
	

	if finalFileUrl != "" {
		// --- LOGIKA UPLOAD ---
		kirimLoading := model.PushWaSend{
			Token:   profile.Token, Target:  msg.From, Type:    "text", Delay:   "0",
			Message: "⏳ Sedang mendownload file...",
		}
		atapi.PostJSON[interface{}](kirimLoading, profile.URLApi)

		fileName := pesan 
		if fileName == "" { fileName = fmt.Sprintf("WA-Upload-%d", time.Now().Unix()) }
		
		if !strings.Contains(fileName, ".") {
			fileName += ".file" 
		}

		respFile, errDown := http.Get(finalFileUrl)
		if errDown != nil {
			replyMsg = "❌ Gagal mendownload file."
		} else {
			defer respFile.Body.Close()
			fileID, webLink, errUp := gdrive.UploadToDrive(sender, fileName, respFile.Body)
			
			if errUp != nil {
				fmt.Printf("Error Upload GDrive: %v\n", errUp)
				replyMsg = "❌ Gagal upload ke Drive."
			} else {
				newFile := model.DriveFile{
					ID:           primitive.NewObjectID(), UserPhone:    sender,
					GoogleFileID: fileID, FileName:     fileName,
					MimeType:     "unknown", DriveLink:    webLink,
					UploadedAt:   time.Now(),
				}
				atdb.InsertOneDoc(config.Mongoconn, "drive_files", newFile)
				replyMsg = fmt.Sprintf("✅ *File Tersimpan!*\n\n📂 %s\n🔗 %s", fileName, webLink)
			}
		}

	
	// FITUR : BACKUP / EXPORT

	} else if hasPrefixAny(pesanLower, []string{"backup", "export", "unduh"}) {
		
		// 1. Ambil Semua Catatan User
		filter := bson.M{"user_phone": sender}
		opts := options.Find().SetSort(bson.M{"created_at": -1})
		
		cursor, err := config.Mongoconn.Collection("notes").Find(context.TODO(), filter, opts)
		if err != nil {
			replyMsg = "❌ Gagal mengambil data database."
		} else {
			var notes []model.Note
			if cursor != nil { cursor.All(context.TODO(), &notes) }

			if len(notes) > 0 {
				// 2. Susun Isi File Text
				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("DATA BACKUP KAWAI - %s\n", sender))
				sb.WriteString(fmt.Sprintf("Generated: %s\n", time.Now().Format("02 Jan 2006 15:04")))
				sb.WriteString("=========================================\n\n")

				for i, n := range notes {
					sb.WriteString(fmt.Sprintf("[%d] %s (%s)\n", i+1, n.CreatedAt.Format("02/01/2006"), n.Type))
					sb.WriteString(fmt.Sprintf("Isi: %s\n", n.Content))
					sb.WriteString("-----------------------------------------\n")
				}

				// 3. Upload ke Drive
				fileContent := sb.String()
				fileReader := strings.NewReader(fileContent)
				fileName := fmt.Sprintf("Backup_Kawai_%s.txt", time.Now().Format("20060102_1504"))

				kirimLoading := model.PushWaSend{
					Token:   profile.Token, Target:  msg.From, Type:    "text", Delay:   "0",
					Message: "⏳ Membuat backup & upload ke Drive...",
				}
				atapi.PostJSON[interface{}](kirimLoading, profile.URLApi)

				fileID, webLink, errUp := gdrive.UploadToDrive(sender, fileName, fileReader)

				if errUp != nil {
					fmt.Printf("Error Backup: %v\n", errUp)
					replyMsg = "❌ Gagal backup ke Drive. Cek token Google."
				} else {
					// Simpan Log
					newFile := model.DriveFile{
						ID:           primitive.NewObjectID(), UserPhone:    sender,
						GoogleFileID: fileID, FileName:     fileName,
						MimeType:     "text/plain", DriveLink:    webLink,
						UploadedAt:   time.Now(),
					}
					atdb.InsertOneDoc(config.Mongoconn, "drive_files", newFile)

					replyMsg = fmt.Sprintf("✅ *Backup Berhasil!*\n\n📂 Nama: %s\n🔗 Link: %s", fileName, webLink)
				}

			} else {
				replyMsg = "Belum ada catatan untuk dibackup."
			}
		}

	// FITUR A: SIMPAN / CATAT
	
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

			if foundURL != "" {
				linkTitle := content
				if linkTitle == "" || linkTitle == foundURL { linkTitle = foundURL }
				atdb.InsertOneDoc(config.Mongoconn, "links", model.Link{
					ID: primitive.NewObjectID(), NoteID: noteID, UserPhone: sender, URL: foundURL, Title: linkTitle, CreatedAt: time.Now(),
				})
			}
			for _, t := range foundTags {
				atdb.InsertOneDoc(config.Mongoconn, "tags", model.Tag{
					ID: primitive.NewObjectID(), NoteID: noteID, TagName: t, UserPhone: sender,
				})
			}
			replyMsg = "✅ Tersimpan! Ketik *List* untuk melihat."
		} else {
			replyMsg = "Format salah. Ketik: *Catat [isi]*"
		}

	
	// FITUR B: LIST
	
	} else if hasPrefixAny(pesanLower, []string{"list", "menu", "tampilkan"}) {
		filter := bson.M{"user_phone": sender}
		opts := options.Find().SetLimit(10).SetSort(bson.M{"created_at": -1})

		if strings.Contains(pesanLower, "link") {
			cursor, _ := config.Mongoconn.Collection("links").Find(context.TODO(), filter, opts)
			var links []model.Link
			if cursor != nil { cursor.All(context.TODO(), &links) }
			if len(links) > 0 {
				var sb strings.Builder
				sb.WriteString("🔗 *Koleksi Link*\n")
				for i, l := range links {
					sb.WriteString(fmt.Sprintf("\n%d. %s\n   %s", i+1, l.Title, l.URL))
				}
				replyMsg = sb.String()
			} else {
				replyMsg = "Belum ada link."
			}
		} else {
			cursor, _ := config.Mongoconn.Collection("notes").Find(context.TODO(), filter, opts)
			var notes []model.Note
			if cursor != nil { cursor.All(context.TODO(), &notes) }
			if len(notes) > 0 {
				var sb strings.Builder
				sb.WriteString("📂 *Catatan Terkini*\n")
				for i, n := range notes {
					display := n.Content
					if len(display) > 35 { display = display[:35] + "..." }
					sb.WriteString(fmt.Sprintf("\n%d. %s", i+1, display))
				}
				sb.WriteString("\n\n💡 Ketik *Backup* untuk simpan semua ke Drive!")
				replyMsg = sb.String()
			} else {
				replyMsg = "Belum ada catatan."
			}
		}

	// FITUR C: BACA DETAIL 

	} else if isNumberOnly {
		skip := int64(targetNo - 1)
		opts := options.FindOne().SetSkip(skip).SetSort(bson.M{"created_at": -1})
		var note model.Note
		errDB := config.Mongoconn.Collection("notes").FindOne(context.TODO(), bson.M{"user_phone": sender}, opts).Decode(&note)
		if errDB == nil {
			replyMsg = fmt.Sprintf("📂 *DETAIL NO. %d*\n\n%s", targetNo, note.Content)
		} else {
			replyMsg = "❌ Data tidak ditemukan."
		}

	
	// FITUR D: REMINDER
	
	} else if hasPrefixAny(pesanLower, []string{"ingatkan", "remind", "ingat", "ing", "ingt"}) {
		scheduledTime, title := timeparse.ParseNaturalTime(pesan)
		
		if scheduledTime.IsZero() {
			replyMsg = `🤔 *Waduh, saya kurang paham waktunya.*
Coba ketik waktu yang jelas ya, contohnya:
- _Besok jam 9_ (atau _Bsk jm 9_)
- _5 menit lagi_ (atau _5 mnt lg_)
- _Tgl 17 Agustus_
- _Hari Senin_`
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

	
	// FITUR E: BANTUAN

	} else if hasPrefixAny(pesanLower, []string{"help", "bantuan", "halo", "menu", "p", "hi", "hai", "info"}) {
		replyMsg = `🤖 *Kawai Assistant Menu*

1️⃣ *BACKUP KE DRIVE* (Andalan! 🌟)
   Ketik: _Backup_ atau _Export_
   👉 Upload rekap catatan ke Google Drive.

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
			Token:   profile.Token, Target:  msg.From, Type:    "text", Delay:   "1", Message: replyMsg,
		}
		atapi.PostJSON[interface{}](kirim, profile.URLApi)
	}

	json.NewEncoder(w).Encode(model.Response{Response: "OK"})
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(model.Response{Response: "404 Not Found"})
}