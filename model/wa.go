package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// 1. Pesan Masuk dari PushWa (Sesuai Info Support)
type PushWaIncoming struct {
	DeviceNumber string `json:"deviceNumber"` // Nomor bot yang menerima
	Message      string `json:"message"`      // Isi pesan
	From         string `json:"from"`         // Nomor pengirim (User)
}

// 2. Struktur Kirim Pesan ke PushWa
type PushWaSend struct {
	Token   string `json:"token"`
	Target  string `json:"target"`
	Type    string `json:"type"`
	Delay   string `json:"delay"`
	Message string `json:"message"`
}

// 3. Response Standar
type Response struct {
	Response string `json:"response"`
}

// 4. Profile Database
type Profile struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Token       string             `bson:"token" json:"token"`
	Phonenumber string             `bson:"phonenumber" json:"phonenumber"`
	URLApi      string             `bson:"urlapi" json:"urlapi"` // https://dash.pushwa.com/api/kirimPesan
}