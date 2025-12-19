package handler

import (
	"net/http"
	"os"

	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/helper/atdb"
	"github.com/kawai-org/kawai-backend/route"
)

// Handler adalah fungsi utama yang akan dipanggil oleh Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	// Inisialisasi Database (dilakukan setiap ada request atau bisa dioptimasi nanti)
	mconn := atdb.DBInfo{
		DBString: os.Getenv("MONGOSTRING"),
		DBName:   "kawai_db",
	}
	
	// Pastikan koneksi db diisi ke config
	config.Mongoconn, _ = atdb.MongoConnect(mconn)

	// Teruskan request ke fungsi URL (Webhook) kamu
	route.URL(w, r)
}