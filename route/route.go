package route

import (
	"net/http"
	"strings"

	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/controller"
)

func URL(w http.ResponseWriter, r *http.Request) {
	// CORS Config
	if config.SetAccessControlHeaders(w, r) {
		return
	}
	config.SetEnv()

	path := r.URL.Path
	method := r.Method

	switch {
	// Route untuk Webhook WhatsApp (Menerima Pesan)
	case method == "POST" && strings.HasPrefix(path, "/webhook/nomor/"):
		controller.PostInboxNomor(w, r)

	// Route untuk Cron Job (Pengingat Otomatis)
	case method == "GET" && path == "/api/cron":
		controller.HandleCron(w, r)

	// Login Google: User diarahkan ke Google
	case method == "GET" && path == "/api/auth/google/login":
		controller.GoogleLogin(w, r)

	// Callback Google: Menerima kode dari Google
	case method == "GET" && path == "/api/auth/google/callback":
		controller.GoogleCallback(w, r)
	

	// case method == "POST" && path == "/api/register":
	// 	controller.Register(w, r)
	// case method == "POST" && path == "/api/login":
	// 	controller.Login(w, r)

	// Route Home / Cek Status
	case method == "GET" && path == "/":
		controller.GetHome(w, r)

	// Default 404
	default:
		controller.NotFound(w, r)
	}
}