package main

import (
	"log"
	"net/http"
	"os"

	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/helper/atdb"
	"github.com/kawai-org/kawai-backend/route"
)

func main() {
	// Inisialisasi Database
	mconn := atdb.DBInfo{
		DBString: os.Getenv("MONGOSTRING"), // Mengambil dari Environment Variable Railway
		DBName:   "kawai_db",
	}
	config.Mongoconn, _ = atdb.MongoConnect(mconn)


	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server berjalan di port %s", port)
	if err := http.ListenAndServe(":"+port, http.HandlerFunc(route.URL)); err != nil {
		log.Fatal(err)
	}
}