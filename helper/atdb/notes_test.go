package atdb

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive" // Wajib import ini
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

// Kita test fungsi wrapper database kamu menggunakan Mock
func TestGetOneDoc(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Success GetOneDoc", func(mt *mtest.T) {
		collectionName := "test_col"
		
		// Mock Data yang akan dikembalikan oleh DB
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "db."+collectionName, mtest.FirstBatch, bson.D{
			{Key: "name", Value: "Kawai"},
			{Key: "status", Value: "Active"},
		}))

		// Panggil fungsi asli GetOneDoc
		res, err := GetOneDoc[struct{ Name string `bson:"name"` }](mt.Client.Database("db"), collectionName, bson.M{"name": "Kawai"})

		if err != nil {
			t.Errorf("Gagal ambil doc: %v", err)
		}
		
		if res.Name != "Kawai" {
			t.Errorf("Expected Kawai, got %s", res.Name)
		}
	})

	mt.Run("Not Found GetOneDoc", func(mt *mtest.T) {
		// Mock Data Kosong (0 document)
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "db.test_col", mtest.FirstBatch, bson.D{}))

		_, err := GetOneDoc[struct{ Name string }](mt.Client.Database("db"), "test_col", bson.M{"name": "Ghost"})

		// Harusnya error karena tidak ditemukan
		if err == nil {
			t.Errorf("Harusnya error not found, tapi malah sukses")
		}
	})
}

func TestInsertOneDoc(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Success Insert", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		// InsertOneDoc mengembalikan (id, err)
		id, err := InsertOneDoc(mt.Client.Database("db"), "test_col", bson.M{"name": "New Data"})
		
		if err != nil {
			t.Errorf("Gagal insert: %v", err)
		}
		
		// PERBAIKAN UTAMA DISINI:
		// Jangan pakai 'id == nil', tapi 'id == primitive.NilObjectID'
		// Karena id tipe-nya primitive.ObjectID (Array), bukan Pointer.
		if id == primitive.NilObjectID {
			t.Errorf("ID hasil insert kosong (NilObjectID)")
		}
	})
}