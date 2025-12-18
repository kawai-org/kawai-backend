package gcf

import (
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/helper/atdb"
	"github.com/kawai-org/kawai-backend/route"
)

func init() {
	// Info Database
	mconn := atdb.DBInfo{
		DBString: config.MongoString,
		DBName:   "kawai_db",
	}

	// Koneksi ke MongoDB
	config.Mongoconn, config.ErrorMongoconn = atdb.MongoConnect(mconn)

	// Daftarkan ke Functions Framework (GCP)
	functions.HTTP("WebHook", route.URL)
}