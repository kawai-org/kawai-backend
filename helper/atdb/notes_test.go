package atdb

import (
	"context"
	"testing"
	"time"

	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/model"
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

    // 2. DATA DUMMY
    data := model.Note{
        ID:        primitive.NewObjectID(),
        UserPhone: "628123456789",
        Title:     "Test Catatan",
        Content:   "Ini adalah isi catatan testing code coverage.",
		UpdatedAt: time.Now(),
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
    // Pastikan koneksi tidak nil (jika running test ini sendirian)
    if config.Mongoconn == nil {
        mconn := DBInfo{
            DBString: "URI_MONGODB_KAMU", 
            DBName:   "kawai_db",
        }
        config.Mongoconn, _ = MongoConnect(mconn)
    }

    filter := primitive.M{"title": "Test Catatan"}
    
    // 1. Test Get (Gunakan variabel 'note' agar tidak error)
    note, err := GetOneDoc[model.Note](config.Mongoconn, "notes", filter)
    if err != nil {
        t.Errorf("Gagal ambil catatan: %v", err)
    }
    
    // GUNAKAN VARIABEL 'note' DI SINI (Contoh: Print atau bandingkan)
    if note.Title != "Test Catatan" {
        t.Errorf("Data salah: ingin 'Test Catatan', dapat '%s'", note.Title)
    }

    // 2. Test DeleteMany (Pakai context karena DeleteMany butuh context)
    // Tambahkan import "context" lagi di atas jika menghilang
    collection := config.Mongoconn.Collection("notes")
    _, err = collection.DeleteMany(context.TODO(), filter) 
    
    if err != nil {
        t.Errorf("Gagal membersihkan database: %v", err)
    }
}