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

// 3. Catatan (Notes) - INI FITUR UTAMA KITA SEKARANG
type Note struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserPhone string             `bson:"user_phone" json:"user_phone"`
	Title     string             `bson:"title" json:"title"`
	Content   string             `bson:"content" json:"content"` // Isi mentah
	Type      string             `bson:"type" json:"type"`       // "text" atau "link"
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// 4. Link Tersimpan (Opsional, karena sudah dicover Note, tapi boleh disimpan)
type Link struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserPhone   string             `bson:"user_phone" json:"user_phone"`
	URL         string             `bson:"url" json:"url"`
	Description string             `bson:"description" json:"description"`
}

// 5. Struktur Pendukung Google Calendar (SimpleEvent & CredentialRecord)
// Kita simpan di sini agar helper/gcallapi tidak error
type SimpleEvent struct {
	Summary     string     `json:"summary" bson:"summary"`
	Location    string     `json:"location" bson:"location"`
	Description string     `json:"description" bson:"description"`
	Date        string     `json:"date" bson:"date"`
	TimeStart   string     `json:"timestart" bson:"timestart"`
	TimeEnd     string     `json:"timeend" bson:"timeend"`
	Attendees   []string   `json:"attendees" bson:"attendees"`
	Attachments []FileMeta `json:"attachments" bson:"attachments"`
}

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

type FileMeta struct {
	UserPhone string `bson:"user_phone" json:"user_phone"`
	FileName  string `bson:"file_name" json:"file_name"`
	FileID    string `bson:"file_id" json:"file_id"`
	FileUrl   string `bson:"file_url" json:"file_url"`
	MimeType  string `bson:"mime_type" json:"mime_type"`
	Title     string `bson:"title" json:"title"`
}