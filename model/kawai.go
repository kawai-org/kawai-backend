package model

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 1. users: Data induk (Admin & User User WA gabung disini, beda di 'Role')
type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	PhoneNumber string             `bson:"phone_number" json:"phone_number"`
	Name        string             `bson:"name" json:"name"`
	Role        string             `bson:"role,omitempty" json:"role,omitempty"`     // "user" atau "admin"
	Password    string             `bson:"password,omitempty" json:"password,omitempty"` // Cuma diisi kalau admin
	CreatedAt   time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`
}

// 2. bot_profiles: Konfigurasi Bot
type BotProfile struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Token       string             `bson:"token" json:"token"`
	Phonenumber string             `bson:"phonenumber" json:"phonenumber"`
	URLApi      string             `bson:"urlapi" json:"urlapi"`
	BotName     string             `bson:"botname" json:"botname"`
}

// 3. message_logs: Riwayat Chat
type MessageLog struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	From       string             `bson:"from" json:"from"`
	Message    string             `bson:"message" json:"message"`
	ReceivedAt time.Time          `bson:"received_at" json:"received_at"`
}

// 4. notes: Catatan
type Note struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserPhone string             `bson:"user_phone" json:"user_phone"`
	Original  string             `bson:"original" json:"original"`
	Content   string             `bson:"content" json:"content"`
	Type      string             `bson:"type" json:"type"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// 5. links: URL Extracted
type Link struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	NoteID    primitive.ObjectID `bson:"note_id" json:"note_id"`
	UserPhone string             `bson:"user_phone" json:"user_phone"`
	URL       string             `bson:"url" json:"url"`
	Title     string             `bson:"title,omitempty" json:"title,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// 6. tags: Hashtag (#) - Pertahankan untuk fitur Search
type Tag struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	NoteID    primitive.ObjectID `bson:"note_id" json:"note_id"`
	TagName   string             `bson:"tag_name" json:"tag_name"`
	UserPhone string             `bson:"user_phone" json:"user_phone"`
}

// 7. activity_logs: PENGGANTI CATEGORIES (Syarat Audit Trail)
type ActivityLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserPhone string             `bson:"user_phone" json:"user_phone"` // Siapa pelakunya
	Action    string             `bson:"action" json:"action"`         // "LOGIN_ADMIN", "DELETE_NOTE", "GENERATE_MAGIC_LINK"
	IPAddress string             `bson:"ip_address,omitempty" json:"ip_address,omitempty"`
	Details   string             `bson:"details" json:"details"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// 8. reminders: Pengingat
type Reminder struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserPhone     string             `bson:"user_phone" json:"user_phone"`
	Title         string             `bson:"title" json:"title"`
	ScheduledTime time.Time          `bson:"scheduled_time" json:"scheduled_time"`
	Status        string             `bson:"status" json:"status"`
}

// 9. google_tokens: Auth Google
type GoogleToken struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	UserPhone    string             `bson:"user_phone"`
	ClientID     string             `bson:"client_id"`
	ClientSecret string             `bson:"client_secret"`
	RefreshToken string             `bson:"refresh_token"`
	CreatedAt    time.Time          `bson:"created_at"`
}

// 10. drive_files: History Upload
type DriveFile struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	UserPhone    string             `bson:"user_phone"`
	GoogleFileID string             `bson:"google_file_id"`
	FileName     string             `bson:"file_name"`
	MimeType     string             `bson:"mime_type"`
	DriveLink    string             `bson:"drive_link"`
	UploadedAt   time.Time          `bson:"uploaded_at"`
}