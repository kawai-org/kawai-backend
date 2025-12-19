package atdb

import (
	"context"
	"testing"
	"time"

	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestInsertNote(t *testing.T) {
	// 1. SETUP KONEKSI
	mconn := DBInfo{
		DBString: "mongodb+srv://penerbit:u2cC2MwwS42yKxub@webhook.jej9ieu.mongodb.net/?retryWrites=true&w=majority&appName=webhook",
		DBName:   "kawai_db",
	}

	var err error
	config.Mongoconn, err = MongoConnect(mconn)
	// Jika koneksi gagal, test akan berhenti di sini
	if err != nil {
		t.Fatalf("Gagal koneksi database saat testing: %v", err)
	}

	// 2. DATA DUMMY (Sesuai Struct Baru)
	data := model.Note{
		ID:        primitive.NewObjectID(),
		UserPhone: "628123456789",
		Original:  "Catat Test Code Coverage", // Field baru menggantikan Title
		Content:   "Test Code Coverage",
		Type:      "text",
		CreatedAt: time.Now(), // Field baru menggantikan UpdatedAt
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
	// Pastikan koneksi tidak nil
	if config.Mongoconn == nil {
		mconn := DBInfo{
			DBString: "mongodb+srv://penerbit:u2cC2MwwS42yKxub@webhook.jej9ieu.mongodb.net/?retryWrites=true&w=majority&appName=webhook",
			DBName:   "kawai_db",
		}
		config.Mongoconn, _ = MongoConnect(mconn)
	}

	// Gunakan filter berdasarkan Content atau Original, karena Title sudah tidak ada
	filter := bson.M{"content": "Test Code Coverage"}

	// 1. Test Get
	note, err := GetOneDoc[model.Note](config.Mongoconn, "notes", filter)
	if err != nil {
		t.Errorf("Gagal ambil catatan: %v", err)
	}

	// VALIDASI: Cek Content bukan Title
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