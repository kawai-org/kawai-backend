package controller

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/helper/atdb"
	"github.com/kawai-org/kawai-backend/model"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

var googleOauthConfig = &oauth2.Config{
    ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
    ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	RedirectURL:  "https://kawai-be.vercel.app/api/auth/google/callback", 
	Scopes:       []string{drive.DriveScope},
	Endpoint:     google.Endpoint,
}

// 1. HANDLER LOGIN: 

func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	// Ambil nomor HP dari parameter URL
	userPhone := r.URL.Query().Get("phone")
	if userPhone == "" {
		http.Error(w, "Nomor HP diperlukan", http.StatusBadRequest)
		return
	}

	// "State" adalah cara kita menitipkan Data (Nomor HP) ke Google
	// Nanti Google akan mengembalikan data ini setelah user login
	state := userPhone 

	// Buat URL Login Google
	url := googleOauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	
	// Lempar user ke halaman Google
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// 2. HANDLER CALLBACK: Menerima Laporan dari Google
// URL: /api/auth/google/callback
func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Ambil data yang dikembalikan Google
	code := r.URL.Query().Get("code")
	userPhone := r.URL.Query().Get("state") // Ini nomor HP yang kita titip tadi

	if code == "" || userPhone == "" {
		http.Error(w, "Gagal login Google (Code/State missing)", http.StatusBadRequest)
		return
	}

	// Tukar "Code" menjadi "Token"
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Gagal menukar token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// SIMPAN TOKEN KE DATABASE MONGODB
	// Kita simpan/update data di collection 'google_tokens'
	// Jadi nanti setiap nomor HP punya token sendiri-sendiri!
	newDoc := model.GoogleToken{
		UserPhone:    userPhone,
		ClientID:     googleOauthConfig.ClientID,
		ClientSecret: googleOauthConfig.ClientSecret,
		RefreshToken: token.RefreshToken, // Ini yang paling penting
	}

	// Upsert (Update kalau ada, Insert kalau belum ada)
	// Catatan: Helper atdb.InsertOneDoc itu insert biasa. 
	// Untuk kasus ini kita hapus dulu yang lama biar gampang, lalu insert baru.
	// (Idealnya pake ReplaceOne, tapi biar codingan kamu ga berubah banyak)
	atdb.DeleteOneDoc(config.Mongoconn, "google_tokens", bson.M{"user_phone": userPhone})
	atdb.InsertOneDoc(config.Mongoconn, "google_tokens", newDoc)

	// Tampilkan pesan sukses di Browser User
	fmt.Fprintf(w, "<h1>BERHASIL! 🎉</h1><p>Akun Google berhasil terhubung dengan WhatsApp %s.</p><p>Silakan tutup tab ini dan kembali ke WhatsApp.</p>", userPhone)
}