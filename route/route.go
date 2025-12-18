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
	case method == "POST" && strings.HasPrefix(path, "/webhook/nomor/"):
		controller.PostInboxNomor(w, r)
	case method == "GET" && path == "/":
		controller.GetHome(w, r)
	default:
		controller.NotFound(w, r)
	}
}