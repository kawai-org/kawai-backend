package model

// 1. Pesan Masuk dari PushWa / WhatsApp / Fonnte
type PushWaIncoming struct {
	DeviceNumber string `json:"deviceNumber"`
	Device       string `json:"device,omitempty"` // Milik Fonnte

	Message      string `json:"message"`
	Text         string `json:"text,omitempty"`   // Milik Fonnte

	PushName     string `json:"pushname"`
	Name         string `json:"name,omitempty"`   // Milik Fonnte

	From         string `json:"from"`
	Sender       string `json:"sender,omitempty"` // Milik Fonnte

	ChatID       string `json:"chat_id,omitempty"` 
	RemoteJID    string `json:"remote_jid,omitempty"`
	Participant  string `json:"participant,omitempty"`
	FileUrl      string `json:"file_url,omitempty"`  
	Url          string `json:"url,omitempty"`       
	MimeType     string `json:"mimetype,omitempty"`  
	MimeType2    string `json:"mime_type,omitempty"` 
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

// 4. Struktur Respon Khusus untuk menangkap balasan dari API di Cron
type APIResponse struct {
    Status  interface{} `json:"status"`
    Message string      `json:"message"`
}

// 5. Struktur Kirim Pesan ke Fonnte
type FonnteSend struct {
	Target string `json:"target"`
	Typing string `json:"typing,omitempty"` // Opsional: False
	Delay  string `json:"delay"`
	Message string `json:"message"`
}