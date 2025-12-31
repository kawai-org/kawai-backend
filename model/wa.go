package model

// 1. Pesan Masuk dari PushWa / WhatsApp
type PushWaIncoming struct {
	DeviceNumber string `json:"deviceNumber"`
	Message      string `json:"message"`
	From         string `json:"from"` // Nomor Pengirim
	
	//  FITUR UPLOAD DRIVE
	FileUrl      string `json:"file_url,omitempty"`  // Link download dari WA
	MimeType     string `json:"mimetype,omitempty"`  // Tipe file (pdf/jpg)
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