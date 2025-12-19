package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- KELOMPOK 1: CORE & USER MANAGEMENT ---

// 1. users: Data pengguna yang berinteraksi
type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	PhoneNumber string             `bson:"phone_number" json:"phone_number"`
	Name        string             `bson:"name" json:"name"`
	Role        string             `bson:"role" json:"role"` // "admin", "user"
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

// 2. bot_profiles: Konfigurasi Bot (Token PushWa, dll)
type BotProfile struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Token       string             `bson:"token" json:"token"`
	Phonenumber string             `bson:"phonenumber" json:"phonenumber"`
	URLApi      string             `bson:"urlapi" json:"urlapi"`
	BotName     string             `bson:"botname" json:"botname"`
}

// 3. message_logs: Audit Trail pesan masuk (Pengganti Inbox)
type MessageLog struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	From       string             `bson:"from" json:"from"`
	Message    string             `bson:"message" json:"message"`
	ReceivedAt time.Time          `bson:"received_at" json:"received_at"`
}

// 4. error_logs: Menyimpan error sistem untuk debugging
type ErrorLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Context   string             `bson:"context" json:"context"` // misal: "ParsingNote"
	ErrorMsg  string             `bson:"error_msg" json:"error_msg"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// --- KELOMPOK 2: THE BRAIN (CATATAN & RELASI) ---

// 5. notes: Tabel utama catatan
type Note struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserPhone string             `bson:"user_phone" json:"user_phone"`
	Original  string             `bson:"original" json:"original"` // Pesan asli user
	Content   string             `bson:"content" json:"content"`   // Pesan bersih tanpa keyword
	Type      string             `bson:"type" json:"type"`         // "text", "link", "mixed"
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// 6. links: URL yang diekstrak dari catatan
type Link struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	NoteID    primitive.ObjectID `bson:"note_id" json:"note_id"` // Relasi ke Note
	UserPhone string             `bson:"user_phone" json:"user_phone"`
	URL       string             `bson:"url" json:"url"`
	Title     string             `bson:"title,omitempty" json:"title,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// 7. tags: Hashtag yang diekstrak (misal: #penting)
type Tag struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	NoteID    primitive.ObjectID `bson:"note_id" json:"note_id"` // Relasi ke Note
	TagName   string             `bson:"tag_name" json:"tag_name"`
	UserPhone string             `bson:"user_phone" json:"user_phone"`
}

// 8. categories: Lookup kategori (Opsional, untuk pengembangan dashboard)
type Category struct {
	ID   primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name string             `bson:"name" json:"name"` // "Kuliah", "Kerjaan", "Pribadi"
}

// --- KELOMPOK 3: GOOGLE CALENDAR PREP ---

// 9. reminders: Data pengingat lokal sebelum push ke Google
type Reminder struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserPhone     string             `bson:"user_phone" json:"user_phone"`
	Title         string             `bson:"title" json:"title"`
	ScheduledTime time.Time          `bson:"scheduled_time" json:"scheduled_time"`
	Status        string             `bson:"status" json:"status"` // "pending", "synced"
}

// 10. google_creds: Token OAuth User
type CredentialRecord struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	UserPhone    string             `bson:"user_phone"` // Relasi ke User
	AccessToken  string             `bson:"access_token"`
	RefreshToken string             `bson:"refresh_token"`
	Expiry       time.Time          `bson:"expiry"`
}