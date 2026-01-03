package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ==========================================
// KELOMPOK 1: USER & SYSTEM (3 Tabel)
// ==========================================

// 1. users: Data induk pengguna
type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	PhoneNumber string             `bson:"phone_number" json:"phone_number"`
	Name        string             `bson:"name" json:"name"`
	Role        string             `bson:"role,omitempty" json:"role,omitempty"`         // Field Baru
	Password    string             `bson:"password,omitempty" json:"password,omitempty"` // Field Baru
	CreatedAt   time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`
}

// 2. bot_profiles: Konfigurasi Bot (Token WA, URL API)
type BotProfile struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Token       string             `bson:"token" json:"token"`
	Phonenumber string             `bson:"phonenumber" json:"phonenumber"`
	URLApi      string             `bson:"urlapi" json:"urlapi"`
	BotName     string             `bson:"botname" json:"botname"`
}

// 3. message_logs: Audit Trail / Riwayat Chat
type MessageLog struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	From       string             `bson:"from" json:"from"`
	Message    string             `bson:"message" json:"message"`
	ReceivedAt time.Time          `bson:"received_at" json:"received_at"`
}

// ==========================================
// KELOMPOK 2: CORE FEATURES (Catatan) (4 Tabel)
// ==========================================

// 4. notes: Tabel utama catatan & hybrid reminder
type Note struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserPhone string             `bson:"user_phone" json:"user_phone"`
	Original  string             `bson:"original" json:"original"`
	Content   string             `bson:"content" json:"content"`
	Type      string             `bson:"type" json:"type"` // "text", "link", "mixed", "reminder"
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// 5. links: URL yang diekstrak dari catatan
type Link struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	NoteID    primitive.ObjectID `bson:"note_id" json:"note_id"`
	UserPhone string             `bson:"user_phone" json:"user_phone"`
	URL       string             `bson:"url" json:"url"`
	Title     string             `bson:"title,omitempty" json:"title,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// 6. tags: Hashtag (#)
type Tag struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	NoteID    primitive.ObjectID `bson:"note_id" json:"note_id"`
	TagName   string             `bson:"tag_name" json:"tag_name"`
	UserPhone string             `bson:"user_phone" json:"user_phone"`
}

// 7. categories: Pengelompokan (Syarat Relasi Tambahan)
type Category struct {
	ID   primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name string             `bson:"name" json:"name"` // "Kuliah", "Pribadi"
}

// ==========================================
// KELOMPOK 3: REMINDER & GOOGLE DRIVE (3 Tabel)
// ==========================================

// 8. reminders: Jadwal Pengingat (Cron Job)
type Reminder struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserPhone     string             `bson:"user_phone" json:"user_phone"`
	Title         string             `bson:"title" json:"title"`
	ScheduledTime time.Time          `bson:"scheduled_time" json:"scheduled_time"`
	Status        string             `bson:"status" json:"status"` // "pending", "sent"
}

// 9. google_tokens: Kunci Akses ke Google Drive User
// (Pengganti CredentialRecord yang lama)
type GoogleToken struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	UserPhone    string             `bson:"user_phone"`
	ClientID     string             `bson:"client_id"`
	ClientSecret string             `bson:"client_secret"`
	RefreshToken string             `bson:"refresh_token"`
	CreatedAt    time.Time          `bson:"created_at"`
}

// 10. drive_files: Log File yang sukses diupload ke Drive
type DriveFile struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	UserPhone    string             `bson:"user_phone"`
	GoogleFileID string             `bson:"google_file_id"`
	FileName     string             `bson:"file_name"`
	MimeType     string             `bson:"mime_type"` // pdf, image/jpeg
	DriveLink    string             `bson:"drive_link"`
	UploadedAt   time.Time          `bson:"uploaded_at"`
}

type FAQ struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Question  string             `bson:"question" json:"question"` // Keyword, misal: "biaya kuliah"
	Answer    string             `bson:"answer" json:"answer"`     // Jawaban bot
	CreatedBy string             `bson:"created_by" json:"created_by"` // "admin"
}

// Struct untuk Data Dashboard (Gabungan/Join)
type DashboardStats struct {
    TotalUsers    int64 `json:"total_users"`
    TotalNotes    int64 `json:"total_notes"`
    TotalMessages int64 `json:"total_messages"`
}