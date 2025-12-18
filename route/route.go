package route

import (
	"net/http"
	"strings" // Pakai library standar go

	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/controller"
)

func URL(w http.ResponseWriter, r *http.Request) {
	if config.SetAccessControlHeaders(w, r) {
		return
	}
	config.SetEnv()

	var method, path string = r.Method, r.URL.Path
	
	// Gunakan strings.HasPrefix sebagai pengganti at.URLParam untuk sementara
	switch {
	case method == "POST" && strings.HasPrefix(path, "/webhook/nomor/"):
		controller.PostInboxNomor(w, r)
	case method == "GET" && path == "/":
		controller.GetHome(w, r)
	case method == "GET" && path == "/refresh/token":
		controller.GetNewToken(w, r)
	default:
		controller.NotFound(w, r)
	}
}