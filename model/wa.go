package model

// 1. Pesan Masuk dari PushWa
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

// 3. Response Standar API Kita
type Response struct {
	Response string `json:"response"`
}