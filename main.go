package kawai

import (
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/helper/atdb"
	"github.com/kawai-org/kawai-backend/route"
)

func init() {
	// 1. Siapkan Info Database
	mconn := atdb.DBInfo{
		DBString: config.MongoString,
		DBName:   "kawai_db", // Ganti sesuai nama DB kamu
	}

	// 2. Koneksi ke MongoDB
	config.Mongoconn, config.ErrorMongoconn = atdb.MongoConnect(mconn)

	// 3. Daftarkan ke Functions Framework (GCP)
	functions.HTTP("WebHook", route.URL)
}