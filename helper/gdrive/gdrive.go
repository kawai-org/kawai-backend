package gdrive

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/helper/atdb"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Struktur Data Token (Sesuai Database Kamu)
type GoogleToken struct {
	UserPhone    string    `bson:"user_phone"`
	ClientID     string    `bson:"client_id"`
	ClientSecret string    `bson:"client_secret"`
	RefreshToken string    `bson:"refresh_token"`
}

// 1. Fungsi Koneksi ke Google Drive (Otomatis Refresh Token)
func GetService(userPhone string) (*drive.Service, error) {
	// A. Ambil Token dari Database
	filter := bson.M{"user_phone": userPhone}
	tokenData, err := atdb.GetOneDoc[GoogleToken](config.Mongoconn, "google_tokens", filter)
	if err != nil {
		return nil, fmt.Errorf("token google tidak ditemukan untuk user %s", userPhone)
	}

	// B. Setup Konfigurasi OAuth2
	conf := &oauth2.Config{
		ClientID:     tokenData.ClientID,
		ClientSecret: tokenData.ClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "https://kawai-be.vercel.app/api/auth/google/callback",
		Scopes:       []string{drive.DriveScope},
	}

	// C. Buat Token Source
	// Kita set expiry ke masa lalu biar dia dipaksa refresh token pakai RefreshToken dari DB
	token := &oauth2.Token{
		RefreshToken: tokenData.RefreshToken,
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(-1 * time.Hour), 
	}

	tokenSource := conf.TokenSource(context.Background(), token)

	// D. Buat Client Service
	srv, err := drive.NewService(context.Background(), option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, err
	}

	return srv, nil
}

// 2. Fungsi Cari/Buat Folder "KAWAI_FILES"
func GetOrCreateFolder(srv *drive.Service) (string, error) {
	folderName := "KAWAI_FILES"
	
	// Cek apakah folder sudah ada?
	q := fmt.Sprintf("name = '%s' and mimeType = 'application/vnd.google-apps.folder' and trashed = false", folderName)
	r, err := srv.Files.List().Q(q).Fields("files(id)").Do()
	if err != nil {
		return "", err
	}

	// Kalau ada, kembalikan ID folder lama
	if len(r.Files) > 0 {
		return r.Files[0].Id, nil
	}

	// Kalau tidak ada, buat folder baru
	f := &drive.File{
		Name:     folderName,
		MimeType: "application/vnd.google-apps.folder",
	}
	folder, err := srv.Files.Create(f).Fields("id").Do()
	if err != nil {
		return "", err
	}
	return folder.Id, nil
}

// 3. Fungsi Inti: Upload File
func UploadToDrive(userPhone string, fileName string, fileContent io.Reader) (string, string, error) {
	// A. Connect ke Drive
	srv, err := GetService(userPhone)
	if err != nil {
		return "", "", err
	}

	// B. Dapatkan ID Folder Tujuan
	folderID, err := GetOrCreateFolder(srv)
	if err != nil {
		return "", "", fmt.Errorf("gagal akses folder: %v", err)
	}

	// C. Siapkan File
	f := &drive.File{
		Name:    fileName,
		Parents: []string{folderID}, // Masukkan ke dalam folder KAWAI_FILES
	}

	// D. Eksekusi Upload
	res, err := srv.Files.Create(f).Media(fileContent).Fields("id", "webViewLink").Do()
	if err != nil {
		return "", "", err
	}

	// Return ID File dan Link Web (Bisa diklik user)
	return res.Id, res.WebViewLink, nil
}