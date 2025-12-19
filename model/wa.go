package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// 1. Header untuk validasi secret
type Header struct {
	Secret string `reqHeader:"secret,omitempty"`
}

// 2. Body Pesan Masuk (Webhook)
type WAMessage struct {
	Phone_number       string  `json:"phone_number,omitempty" bson:"phone_number,omitempty"`
	Reply_phone_number string  `json:"reply_phone_number,omitempty" bson:"reply_phone_number,omitempty"`
	Chat_number        string  `json:"chat_number,omitempty" bson:"chat_number,omitempty"`
	Chat_server        string  `json:"chat_server,omitempty" bson:"chat_server,omitempty"`
	Group_name         string  `json:"group_name,omitempty" bson:"group_name,omitempty"`
	Group_id           string  `json:"group_id,omitempty" bson:"group_id,omitempty"`
	Group              string  `json:"group,omitempty" bson:"group,omitempty"`
	Alias_name         string  `json:"alias_name,omitempty" bson:"alias_name,omitempty"`
	Message            string  `json:"messages,omitempty" bson:"messages,omitempty"`
	Is_group           bool    `json:"is_group,omitempty" bson:"is_group,omitempty"`
	Filename           string  `json:"filename,omitempty" bson:"filename,omitempty"`
	Filedata           string  `json:"filedata,omitempty" bson:"filedata,omitempty"`
}

// 3. Response API (Standar Balasan ke Server WhatsAuth)
type Response struct {
	Response string `json:"response"`
}

// 4. Pesan Keluar: TEKS
type TextMessage struct {
	To       string `json:"to"`
	IsGroup  bool   `json:"isgroup"`
	Messages string `json:"messages"`
}

// 5. Pesan Keluar: DOKUMEN (Disiapkan untuk nanti)
type DocumentMessage struct {
	To        string `json:"to"`
	Base64Doc string `json:"base64doc"`
	Filename  string `json:"filename,omitempty"`
	Caption   string `json:"caption,omitempty"`
	IsGroup   bool   `json:"isgroup,omitempty"`
}

// 6. Pesan Keluar: GAMBAR (Disiapkan untuk nanti)
type ImageMessage struct {
	To          string `json:"to"`
	Base64Image string `json:"base64image"`
	Caption     string `json:"caption,omitempty"`
	IsGroup     bool   `json:"isgroup,omitempty"`
}

// 7. Profile Database (Penting untuk Controller ambil Token)
type Profile struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Token       string             `bson:"token" json:"token"`
	Phonenumber string             `bson:"phonenumber" json:"phonenumber"`
	Secret      string             `bson:"secret" json:"secret"`
	URL         string             `bson:"url" json:"url"`
	URLApiText  string             `bson:"urlapitext" json:"urlapitext"`
	BotName     string             `bson:"botname" json:"botname"`
}