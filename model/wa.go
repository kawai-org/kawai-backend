package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// 1. Pesan Masuk dari PushWa (Sesuai Info Support)
type PushWaIncoming struct {
	DeviceNumber string `json:"deviceNumber"`
	Message      string `json:"message"`
	From         string `json:"from"` // Nomor Pengirim
}

// 2. Struktur Kirim Pesan ke PushWa
type PushWaSend struct {
	Token   string `json:"token"`
	Target  string `json:"target"`
	Type    string `json:"type"`
	Delay   string `json:"delay"`
	Message string `json:"message"`
}

type Response struct {
	Response string `json:"response"`
}

// 3. Profile Database
type Profile struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Token       string             `bson:"token" json:"token"`
	Phonenumber string             `bson:"phonenumber" json:"phonenumber"`
	URLApi      string             `bson:"urlapi" json:"urlapi"` // https://dash.pushwa.com/api/kirimPesan
}