package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// Header untuk validasi secret dari WhatsAuth//Header yang dikirim ke webhook dan whatsauth
type Header struct {
	Secret string `reqHeader:"secret,omitempty"` //whatsauth ke webhook
}

//Body Message yang dikirim ke webhook
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
	EntryPoint         string  `json:"entrypoint,omitempty" bson:"entrypoint,omitempty"`
	From_link          bool    `json:"from_link,omitempty" bson:"from_link,omitempty"`
	From_link_delay    uint32  `json:"from_link_delay,omitempty" bson:"from_link_delay,omitempty"`
	Is_group           bool    `json:"is_group,omitempty" bson:"is_group,omitempty"`
	Filename           string  `json:"filename,omitempty" bson:"filename,omitempty"`
	Filedata           string  `json:"filedata,omitempty" bson:"filedata,omitempty"`
	Latitude           float64 `json:"latitude,omitempty" bson:"latitude,omitempty"`
	Longitude          float64 `json:"longitude,omitempty" bson:"longitude,omitempty"`
	LiveLoc            bool    `json:"liveloc,omitempty" bson:"liveloc,omitempty"`
}


type QRStatus struct {
	PhoneNumber string `json:"phonenumber"`
	Status      bool   `json:"status"`
	Code        string `json:"code"`
	Message     string `json:"message"`
}

type SendText struct {
	To       string `json:"to"`
	IsGroup  bool   `json:"isgroup"`
	Messages string `json:"messages"`
}

type Prefill struct {
	Key   string `bson:"key"`
	Value string `bson:"value"`
}

type Webhook struct {
	URL    string `json:"url"`
	Secret string `json:"secret"`
}

// 3. Format Pesan Balasan (Reply) ke WhatsApp
type TextMessage struct {
	To       string `json:"to"`
	IsGroup  bool   `json:"isgroup"`
	Messages string `json:"messages"`
}

type DocumentMessage struct {
	To        string `json:"to"`
	Base64Doc string `json:"base64doc"` // File Drive harus diubah jadi Base64 dulu
	Filename  string `json:"filename,omitempty"`
	Caption   string `json:"caption,omitempty"`
	IsGroup   bool   `json:"isgroup,omitempty"`
}

// 4. Pesan Keluar: GAMBAR
type ImageMessage struct {
	To          string `json:"to"`
	Base64Image string `json:"base64image"`
	Caption     string `json:"caption,omitempty"`
	IsGroup     bool   `json:"isgroup,omitempty"`
}


// 5. Struktur Profile untuk Database (PENTING: Ini untuk simpan Token & Secret)
type Profile struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Token       string             `bson:"token" json:"token"`
	Phonenumber string             `bson:"phonenumber" json:"phonenumber"`
	Secret      string             `bson:"secret" json:"secret"`
	URL         string             `bson:"url" json:"url"`
	URLApiText  string             `bson:"urlapitext" json:"urlapitext"`
	BotName     string             `bson:"botname" json:"botname"`
}