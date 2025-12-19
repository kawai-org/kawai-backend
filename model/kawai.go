package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 1. Profil Pengguna
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Phone     string             `bson:"phone" json:"phone"`
	Name      string             `bson:"name" json:"name"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// 2. Identitas Auth
type Identity struct {
	Phone    string `bson:"phone" json:"phone"`
	Password string `bson:"password" json:"-"`
}

// 3. Catatan (Notes)
type Note struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserPhone string             `bson:"user_phone" json:"user_phone"`
	Title     string             `bson:"title" json:"title"`     // Bisa diisi otomatis (misal: 10 huruf pertama)
	Content   string             `bson:"content" json:"content"` // Isinya RAW message (teks + link campur)
	Type      string             `bson:"type" json:"type"`       // "text" atau "link" (opsional, hasil deteksi bot)
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// 4. Link Tersimpan
type Link struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserPhone   string             `bson:"user_phone" json:"user_phone"`
	URL         string             `bson:"url" json:"url"`
	Description string             `bson:"description" json:"description"`
}

// 5. Pengingat (Reminders)
type Reminder struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserPhone string             `bson:"user_phone" json:"user_phone"`
	Title     string             `bson:"title" json:"title"`
	Time      time.Time          `bson:"time" json:"time"`
}

// 6. Log Sinkronisasi Google
type SyncLog struct {
	UserPhone string    `bson:"user_phone" json:"user_phone"`
	Service   string    `bson:"service" json:"service"`
	LastSync  time.Time `bson:"last_sync" json:"last_sync"`
}

// 7. Token Session
type Token struct {
	Phone string `bson:"phone" json:"phone"`
	Value string `bson:"value" json:"value"`
}

// 8. Google Creds
type GoogleCred struct {
	Phone        string `bson:"phone" json:"phone"`
	RefreshToken string `bson:"refresh_token" json:"-"`
}

// 9. Audit Log
type AuditLog struct {
	Phone    string    `bson:"phone" json:"phone"`
	Action   string    `bson:"action" json:"action"`
	DateTime time.Time `bson:"datetime" json:"datetime"`
}

// 10. File Metadata
type FileMeta struct {
	UserPhone string `bson:"user_phone" json:"user_phone"`
	FileName  string `bson:"file_name" json:"file_name"`
	FileID    string `bson:"file_id" json:"file_id"`
	FileUrl   string `bson:"file_url" json:"file_url"`
	MimeType  string `bson:"mime_type" json:"mime_type"`
	Title     string `bson:"title" json:"title"`
}

type IteungMessage struct {
	Phone   string `json:"phone" bson:"phone"`
	Alias   string `json:"alias" bson:"alias"`
	Message string `json:"message" bson:"message"`
}

type Response struct {
	Response string `json:"response"`
}

type WebHook struct {
	URL    string `json:"url" bson:"url"`
	Secret string `json:"secret" bson:"secret"`
}

// SimpleEvent disesuaikan dengan gcalendar.go
type SimpleEvent struct {
	Summary     string     `json:"summary" bson:"summary"`
	Location    string     `json:"location" bson:"location"`
	Description string     `json:"description" bson:"description"`
	Date        string     `json:"date" bson:"date"`           // Tambahkan ini
	TimeStart   string     `json:"timestart" bson:"timestart"` // Tambahkan ini
	TimeEnd     string     `json:"timeend" bson:"timeend"`     // Tambahkan ini
	Attendees   []string   `json:"attendees" bson:"attendees"`
	Attachments []FileMeta `json:"attachments" bson:"attachments"`
}

// CredentialRecord disesuaikan dengan gettoken.go
type CredentialRecord struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	ClientID     string             `bson:"client_id"`
	ClientSecret string             `bson:"client_secret"`
	Scopes       []string           `bson:"scopes"`
	AuthURI      string             `bson:"auth_uri"`
	TokenURI     string             `bson:"token_uri"`
	RedirectURIs []string           `bson:"redirect_uris"`
	Token        string             `bson:"token"`
	RefreshToken string             `bson:"refresh_token"`
	Expiry       time.Time          `bson:"expiry"`
}