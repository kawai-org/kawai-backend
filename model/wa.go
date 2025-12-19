package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// 1. Struktur Pesan Masuk dari PushWa (Webhook)
// PushWa biasanya mengirim data dengan format ini
type PushWaIncoming struct {
	From      string `json:"from"`      // Nomor pengirim (e.g., 628xxx)
	Message   string `json:"message"`   // Isi pesan
	PushName  string `json:"pushname"`  // Nama kontak WA
	Type      string `json:"type"`      // Teks, image, dll
	Timestamp int64  `json:"timestamp"` // Waktu kirim
	IsGroup   bool   `json:"is_group,omitempty"` // Opsional, tergantung versi PushWa
}

// 2. Struktur untuk Mengirim Pesan (API Send)
// Dokumentasi PushWa mewajibkan Token ada di dalam Body
type PushWaSend struct {
	Token   string `json:"token"`
	Target  string `json:"target"`  // Nomor tujuan
	Type    string `json:"type"`    // "text"
	Delay   string `json:"delay"`   // "1" (detik)
	Message string `json:"message"` // Isi pesan
}

// 3. Response sederhana
type Response struct {
	Response string `json:"response"`
}

// 4. Profile Database (Update struktur agar sesuai kebutuhan)
type Profile struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Token       string             `bson:"token" json:"token"`
	Phonenumber string             `bson:"phonenumber" json:"phonenumber"`
	URLApi      string             `bson:"urlapi" json:"urlapi"` // https://dash.pushwa.com/api/kirimPesan
	Secret      string             `bson:"secret" json:"secret"` // (Opsional jika PushWa support webhook secret)
}