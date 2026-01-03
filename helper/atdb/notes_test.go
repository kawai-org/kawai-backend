package atdb

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

// Kita test fungsi wrapper database kamu menggunakan Mock
func TestGetOneDoc(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Success GetOneDoc", func(mt *mtest.T) {
		collectionName := "test_col"
		
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "db."+collectionName, mtest.FirstBatch, bson.D{
			{Key: "name", Value: "Kawai"},
			{Key: "status", Value: "Active"},
		}))

		res, err := GetOneDoc[struct{ Name string `bson:"name"` }](mt.Client.Database("db"), collectionName, bson.M{"name": "Kawai"})

		if err != nil {
			t.Errorf("Gagal ambil doc: %v", err)
		}
		
		if res.Name != "Kawai" {
			t.Errorf("Expected Kawai, got %s", res.Name)
		}
	})

	mt.Run("Not Found GetOneDoc", func(mt *mtest.T) {
		// PERBAIKAN DISINI: Hapus 'bson.D{}' agar batch benar-benar kosong
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "db.test_col", mtest.FirstBatch))

		_, err := GetOneDoc[struct{ Name string }](mt.Client.Database("db"), "test_col", bson.M{"name": "Ghost"})

		if err == nil {
			t.Errorf("Harusnya error not found, tapi malah sukses")
		}
	})
}

func TestInsertOneDoc(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Success Insert", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		id, err := InsertOneDoc(mt.Client.Database("db"), "test_col", bson.M{"name": "New Data"})
		
		if err != nil {
			t.Errorf("Gagal insert: %v", err)
		}
		
		if id == primitive.NilObjectID {
			t.Errorf("ID hasil insert kosong")
		}
	})
}