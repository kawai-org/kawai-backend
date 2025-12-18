package gcf

import (
	"net/http"

	"github.com/kawai-org/kawai-backend/route"
)

// KawaiBackend adalah entry point untuk Google Cloud Function
func KawaiBackend(w http.ResponseWriter, r *http.Request) {
	route.URL(w, r)
}