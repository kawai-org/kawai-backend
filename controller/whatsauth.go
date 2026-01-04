package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
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

// Helper: Fix Invalid UTF-8 (Penyebab error di database)
func fixUTF8(s string) string {
	if !utf8.ValidString(s) {
		v := make([]rune, 0, len(s))
		for i, r := range s {
			if r == utf8.RuneError {
				_, size := utf8.DecodeRuneInString(s[i:])
				if size == 1 {
					continue
				}
			}
			v = append(v, r)
		}
		return string(v)
	}
	return s
}

// Helper: Upsert User (Otomatis simpan data user baru)
func EnsureUserExists(phone, name string) {
    // Filter berdasarkan nomor HP
	filter := bson.M{"phone_number": phone}
    
    // Data yang mau disimpan/diupdate
	update := bson.M{
		"$set": bson.M{
			"phone_number": phone,
			"name":         name,
			"role":         "user", // Default role
            // "$setOnInsert": bson.M{"created_at": time.Now()}, // jika mau created_at tidak berubah
		},
	}
    // Opsi: Upsert = True (Kalau gak ada, buat baru. Kalau ada, update)
	opts := options.Update().SetUpsert(true)
    
	config.Mongoconn.Collection("users").UpdateOne(context.TODO(), filter, update, opts)
}

// --- LOGIKA UTAMA BOT ---
func PostInboxNomor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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

	// 2. Sanitasi & Simpan User
	msg.Message = fixUTF8(msg.Message)
	sender := msg.From
	if strings.Contains(sender, "@") {
		sender = strings.Split(sender, "@")[0]
	}
	
	// Antisipasi jika PushName kosong dari WA
	userName := msg.PushName
	if userName == "" {
		userName = "Kawai User"
	}
	go EnsureUserExists(sender, userName) // Simpan user baru ke DB

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
	if finalFileUrl == "" {
		finalFileUrl = msg.Url
	}

	if finalFileUrl != "" {
		// --- LOGIKA UPLOAD ---
		kirimLoading := model.PushWaSend{
			Token:   profile.Token, Target: msg.From, Type: "text", Delay: "0",
			Message: "⏳ Sedang mendownload file...",
		}
		atapi.PostJSON[interface{}](kirimLoading, profile.URLApi)

		fileName := pesan
		if fileName == "" {
			fileName = fmt.Sprintf("WA-Upload-%d", time.Now().Unix())
		}

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
					ID:           primitive.NewObjectID(), UserPhone: sender,
					GoogleFileID: fileID, FileName: fileName,
					MimeType: "unknown", DriveLink: webLink,
					UploadedAt: time.Now(),
				}
				atdb.InsertOneDoc(config.Mongoconn, "drive_files", newFile)
				replyMsg = fmt.Sprintf("✅ *File Tersimpan!*\n\n📂 %s\n🔗 %s", fileName, webLink)
			}
		}

		} else if hasPrefixAny(pesanLower, []string{"dashboard", "admin", "panel", "login"}) {
    
    // 1. Buat Token JWT (Berlaku 15 menit)
    expirationTime := time.Now().Add(15 * time.Minute)
    claims := &jwt.MapClaims{
        "user_phone": sender,
        "role":       "user",
        "exp":        expirationTime.Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    
    // 2. Ambil Secret Key
    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        fmt.Println("CRITICAL ERROR: JWT_SECRET belum diset di .env!") 
        jwtSecret = "temporary_secret_jangan_dipakai_prod"
    }
    tokenString, _ := token.SignedString([]byte(jwtSecret))

	fmt.Printf("\n[MAGIC LINK]\nUser: %s\nToken: %s\nLink: https://kawai-frontend.vercel.app/auth/magic?token=%s\n\n", sender, tokenString, tokenString)
	

    // 3. Buat Magic Link
    // Pastikan domain ini sesuai dengan frontend kamu nanti
    magicLink := fmt.Sprintf("https://kawai-frontend.vercel.app/auth/magic?token=%s", tokenString)
    
    // 4. Siapkan Pesan Balasan
    replyMsg = fmt.Sprintf("🎛 *DASHBOARD USER*\n\nKlik link di bawah ini untuk mengelola catatan & pengingat (Edit/Hapus/Rapikan):\n\n👉 %s\n\n_Link ini kedaluwarsa dalam 15 menit._", magicLink)

    // 5. [IMPLEMENTASI STORED PROCEDURE / AUDIT LOG]
    // Catat aktivitas ini ke database secara background (Async)
    go func() {
        logData := model.ActivityLog{
            ID:        primitive.NewObjectID(),
            UserPhone: sender,
            Action:    "REQUEST_DASHBOARD",
            Details:   "User meminta magic link dashboard",
            CreatedAt: time.Now(),
        }
        atdb.InsertOneDoc(config.Mongoconn, "activity_logs", logData)
    }()

	// FITUR : BACKUP / EXPORT 
	} else if hasPrefixAny(pesanLower, []string{"backup", "export", "unduh"}) {

		// 1. CEK APAKAH USER SUDAH PUNYA TOKEN?
		filterToken := bson.M{"user_phone": sender}
		checkToken, _ := config.Mongoconn.Collection("google_tokens").CountDocuments(context.TODO(), filterToken)

		if checkToken == 0 {
			// KASUS A: Belum Login Google -> Kirim Link Login
			linkLogin := fmt.Sprintf("https://kawai-be.vercel.app/api/auth/google/login?phone=%s", sender)

			replyMsg = fmt.Sprintf("⚠️ *Akses Drive Belum Terhubung*\n\nMaaf bos, saya belum punya izin akses ke Google Drive kamu.\n\n👉 *Klik link ini untuk menghubungkan:*\n%s\n\nSetelah sukses login, coba ketik *Backup* lagi ya!", linkLogin)

		} else {
			// KASUS B: Sudah Login -> Lanjutkan Logika Backup Lama
			
			// Ambil Semua Catatan User
			filter := bson.M{"user_phone": sender}
			opts := options.Find().SetSort(bson.M{"created_at": -1})

			cursor, err := config.Mongoconn.Collection("notes").Find(context.TODO(), filter, opts)
			if err != nil {
				replyMsg = "❌ Gagal mengambil data database."
			} else {
				var notes []model.Note
				if cursor != nil {
					cursor.All(context.TODO(), &notes)
				}

				if len(notes) > 0 {
					// Susun Isi File Text
					var sb strings.Builder
					sb.WriteString(fmt.Sprintf("DATA BACKUP KAWAI - %s\n", sender))
					sb.WriteString(fmt.Sprintf("Generated: %s\n", time.Now().Format("02 Jan 2006 15:04")))
					sb.WriteString("=========================================\n\n")

					for i, n := range notes {
						sb.WriteString(fmt.Sprintf("[%d] %s (%s)\n", i+1, n.CreatedAt.Format("02/01/2006"), n.Type))
						sb.WriteString(fmt.Sprintf("Isi: %s\n", n.Content))
						sb.WriteString("-----------------------------------------\n")
					}

					// Upload ke Drive
					fileContent := sb.String()
					fileReader := strings.NewReader(fileContent)
					fileName := fmt.Sprintf("Backup_Kawai_%s.txt", time.Now().Format("20060102_1504"))

					kirimLoading := model.PushWaSend{
						Token:   profile.Token, Target: msg.From, Type: "text", Delay: "0",
						Message: "⏳ Membuat backup & upload ke Drive...",
					}
					atapi.PostJSON[interface{}](kirimLoading, profile.URLApi)

					fileID, webLink, errUp := gdrive.UploadToDrive(sender, fileName, fileReader)

					if errUp != nil {
						// 1. Log error asli ke console supaya bisa dicek di Vercel
						fmt.Printf("Error Backup: %v\n", errUp)

						// 2. Cek apakah errornya karena kuota penuh
						errString := fmt.Sprintf("%v", errUp)
						if strings.Contains(errString, "storageQuotaExceeded") {
							replyMsg = "❌ Gagal Backup: Penyimpanan Google Drive penuh! Mohon hapus beberapa file dulu."
						} else {
							replyMsg = "❌ Gagal backup ke Drive. Token mungkin expired, coba login ulang."
						}
					} else {
						// Simpan Log jika sukses
						newFile := model.DriveFile{
							ID:           primitive.NewObjectID(), UserPhone: sender,
							GoogleFileID: fileID, FileName: fileName,
							MimeType:     "text/plain", DriveLink: webLink,
							UploadedAt:   time.Now(),
						}
						atdb.InsertOneDoc(config.Mongoconn, "drive_files", newFile)

						replyMsg = fmt.Sprintf("✅ *Backup Berhasil!*\n\n📂 Nama: %s\n🔗 Link: %s", fileName, webLink)
					}

				} else {
					replyMsg = "Belum ada catatan untuk dibackup."
				}
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
				if content == foundURL {
					noteType = "link"
				} else {
					noteType = "mixed"
				}
			}

			noteID := primitive.NewObjectID()
			atdb.InsertOneDoc(config.Mongoconn, "notes", model.Note{
				ID: noteID, UserPhone: sender, Original: pesan, Content: content, Type: noteType, CreatedAt: time.Now(),
			})

			if foundURL != "" {
				linkTitle := content
				if linkTitle == "" || linkTitle == foundURL {
					linkTitle = foundURL
				}
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
			if cursor != nil {
				cursor.All(context.TODO(), &links)
			}
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
			if cursor != nil {
				cursor.All(context.TODO(), &notes)
			}
			if len(notes) > 0 {
				var sb strings.Builder
				sb.WriteString("📂 *Catatan Terkini*\n")
				for i, n := range notes {
					display := n.Content
					if len(display) > 35 {
						display = display[:35] + "..."
					}
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

1️⃣ *SIMPAN CATATAN* 📝
   Keyword: _Catat, Simpan_
   _Catat [Hashtag] [Isi]_
   👉 _Catat #kuliah ide skripsi _
   👉 _Catat beli telor #belanja_
   👉 _Simpan Link Zoom https://zoom.us_
   _(Nanti di Dashboard bisa dicari per label!)_

2️⃣ *LIHAT DATA* 📂
   Keyword: _List, Menu_
   👉 _List_ (Lihat semua)
   👉 _List Link_ (Khusus link)
3️⃣ *BACA CATATAN* 📖
   Keyword: _Baca, Show_
   👉 _Baca 1_ (Untuk baca no 1)

4️⃣ *PENGINGAT* ⏰
   Keyword: _Ingatkan, Ingat, Remind_
   ✅ _Ingatkan Rapat Besok jam 10_
   ✅ _Ingat bayar UKT Lusa 09.30_
   ✅ _Ingatkan tgl 17 Agustus 13:00_
   ✅ _Ingatkan masak mie 5 menit lagi_

5️⃣ *DASHBOARD* 
   Ketik: _Dashboard_
   👉 Edit catatan & atur alarm lewat web.

6️⃣ *BACKUP DATA* 💾
   Keyword: _Backup, Export_
   👉 _Backup_ (Simpan semua catatan ke Google Drive)

Selamat mencoba! 😊`
	}

	// Kirim Balasan Final
	if replyMsg != "" && profile.Token != "" {
		kirim := model.PushWaSend{
			Token:   profile.Token, Target: msg.From, Type: "text", Delay: "1", Message: replyMsg,
		}
		atapi.PostJSON[interface{}](kirim, profile.URLApi)
	}

	json.NewEncoder(w).Encode(model.Response{Response: "OK"})
}

func NotFound(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json") // Tambahan bagus
	w.WriteHeader(http.StatusNotFound)                
	json.NewEncoder(w).Encode(model.Response{Response: "404 Not Found"})
}