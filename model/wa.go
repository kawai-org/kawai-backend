package model

// 1. Pesan Masuk dari PushWa / WhatsApp
type PushWaIncoming struct {
	DeviceNumber string `json:"deviceNumber"`
	Message      string `json:"message"`
	PushName string `json:"pushname"`
	From         string `json:"from"` 
	FileUrl      string `json:"file_url,omitempty"`  // Versi 1
	Url          string `json:"url,omitempty"`       // Versi 2 (Sering dipakai)
	MimeType     string `json:"mimetype,omitempty"`  // Versi 1
	MimeType2    string `json:"mime_type,omitempty"` // Versi 2
}

// 2. Struktur Kirim Pesan ke PushWa
type PushWaSend struct {
	Token   string `json:"token"`
	Target  string `json:"target"`
	Type    string `json:"type"`
	Delay   string `json:"delay"`
	Message string `json:"message"`
}

// 3. Response Standar API
type Response struct {
	Response string `json:"response"`
}