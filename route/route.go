package route

import (
	"net/http"
	"strings"

	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/controller"
)

func URL(w http.ResponseWriter, r *http.Request) {
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

	// Route Home / Cek Status
	case method == "GET" && path == "/":
		controller.GetHome(w, r)

	// Default 404
	default:
		controller.NotFound(w, r)
	}
}