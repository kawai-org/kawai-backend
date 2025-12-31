package atdb

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv" // Import ini
	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Helper untuk load env dan koneksi agar tidak berulang
func setupTestDB(t *testing.T) {
	if config.Mongoconn == nil {
		// Asumsi file ini di helper/atdb (2 level dari root)
		_ = godotenv.Load("../../.env") 
		
		mongoString := os.Getenv("MONGOSTRING")
		if mongoString == "" {
			t.Fatal("MONGOSTRING tidak ditemukan di .env")
		}

		mconn := DBInfo{
			DBString: mongoString, // AMAN
			DBName:   "kawai_db",
		}
		var err error
		config.Mongoconn, err = MongoConnect(mconn)
		if err != nil {
			t.Fatalf("Gagal koneksi database: %v", err)
		}
	}
}

func TestInsertNote(t *testing.T) {
	// 1. SETUP KONEKSI
	setupTestDB(t)

	// 2. DATA DUMMY (Sesuai Struct Baru)
	data := model.Note{
		ID:        primitive.NewObjectID(),
		UserPhone: "628123456789",
		Original:  "Catat Test Code Coverage",
		Content:   "Test Code Coverage",
		Type:      "text",
		CreatedAt: time.Now(),
	}

	// 3. EKSEKUSI
	res, err := InsertNote(data)

	// 4. VALIDASI
	if err != nil {
		t.Errorf("Gagal menyimpan catatan: %v", err)
	}

	if res == nil {
		t.Error("Hasil insert kosong (nil)")
	}
}

func TestGetAndDeleteNote(t *testing.T) {
	// Pastikan koneksi ready
	setupTestDB(t)

	// Gunakan filter berdasarkan Content
	filter := bson.M{"content": "Test Code Coverage"}

	// 1. Test Get
	note, err := GetOneDoc[model.Note](config.Mongoconn, "notes", filter)
	if err != nil {
		t.Errorf("Gagal ambil catatan: %v", err)
	}

	// VALIDASI
	if note.Content != "Test Code Coverage" {
		t.Errorf("Data salah: ingin 'Test Code Coverage', dapat '%s'", note.Content)
	}

	// 2. Test DeleteMany
	collection := config.Mongoconn.Collection("notes")
	_, err = collection.DeleteMany(context.TODO(), filter)

	if err != nil {
		t.Errorf("Gagal membersihkan database: %v", err)
	}
}