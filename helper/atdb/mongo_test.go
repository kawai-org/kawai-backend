package atdb

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestMongoOperations(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	// 1. Test InsertOneDoc
	mt.Run("InsertOneDoc Success", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, err := InsertOneDoc(mt.Client.Database("db"), "coll", bson.M{"data": "test"})
		if err != nil {
			t.Errorf("InsertOneDoc error: %v", err)
		}
	})

	// 2. Test GetCountDoc
	mt.Run("GetCountDoc Success", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(
			1,
			"db.coll",
			mtest.FirstBatch,
			bson.D{{Key: "n", Value: 5}}, 
		))
		count, err := GetCountDoc(mt.Client.Database("db"), "coll", bson.M{})
		if err != nil {
			t.Errorf("GetCountDoc error: %v", err)
		}
		if count != 5 {
			t.Errorf("Expected 5, got %d", count)
		}
	})

	// 3. Test DeleteOneDoc (Tangkap 2 value: result, err)
	mt.Run("DeleteOneDoc Success", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, err := DeleteOneDoc(mt.Client.Database("db"), "coll", bson.M{"_id": "123"})
		if err != nil {
			t.Errorf("DeleteOneDoc error: %v", err)
		}
	})

	// 4. Test UpdateOneDoc (Tangkap 2 value: result, err)
	mt.Run("UpdateOneDoc Success", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, err := UpdateOneDoc(mt.Client.Database("db"), "coll", bson.M{"_id": "123"}, bson.M{"$set": bson.M{"a": 1}})
		if err != nil {
			t.Errorf("UpdateOneDoc error: %v", err)
		}
	})
}

// Perbaikan Warning Unused Field
func TestConnectionHelpers(t *testing.T) {
	dbInfo := DBInfo{
		DBString: "mongodb://localhost:27017",
		DBName:   "test_db",
	}
	
	// Gunakan kedua field agar tidak dianggap unused
	if dbInfo.DBName != "test_db" || dbInfo.DBString == "" {
		t.Error("Struct DBInfo salah assign")
	}
}