<div align="center">
  <h1>Kawai Assistant - Backend Services</h1>
  <p><i>A Go-based robust backend engine for handling notes, reminders, and WhatsApp Bot integration.</i></p>

  <p>
    <img alt="Go Version" src="https://img.shields.io/badge/Go-1.24.4-00ADD8?style=flat-square&logo=go" />
    <img alt="MongoDB" src="https://img.shields.io/badge/MongoDB-Native%20Driver-47A248?style=flat-square&logo=mongodb" />
    <img alt="JWT Auth" src="https://img.shields.io/badge/Security-JWT-000000?style=flat-square&logo=jsonwebtokens" />
    <img alt="Google OAuth" src="https://img.shields.io/badge/Auth-Google%20OAuth2-4285F4?style=flat-square&logo=google" />
    <img alt="Deployment" src="https://img.shields.io/badge/Deployed%20on-Vercel-black?style=flat-square&logo=vercel" />
    <img alt="Swagger API Docs" src="https://img.shields.io/badge/API%20Docs-Swagger-85EA2D?style=flat-square&logo=swagger" />
  </p>
</div>

<hr>

## Deskripsi Proyek
**Kawai Assistant Backend** adalah inti layanan komputasi untuk mengelola catatan pengguna (notes) dan pengingat (reminders) secara efisien dan cepat, dengan integrasi langsung melalui *WhatsApp Bot*. 

Proyek ini dibangun murni menggunakan **Go (Golang)** yang disajikan secara efisien, mengandalkan *standard library* `net/http` yang dioptimalkan. Layanan ini memastikan setiap interaksi API, mulai dari pengambilan data hingga manajemen token dan keamanan, berjalan dengan kecepatan tinggi yang biasa disediakan oleh kompilasi arsitektur native Go. Proyek ini sangat cocok untuk di-*deploy* sebagai arsitektur *Serverless* dalam ekosistem Vercel.

## Fitur Utama (Key Features)
- **Sistem Autentikasi Solid**: Kombinasi antara integrasi *Google Sign-In* (OAuth2) dan *JSON Web Token (JWT)* yang menciptakan perlindungan lapis ganda di tingkat API.
- **Dashboard Data (Notes & Reminders)**: RESTful API yang melayani CRUD catatan dan sistem alarm pengingat. Dilindungi penuh oleh lapisan *Middleware Authentication*.
- **WhatsApp Bot Webhook**: Sistem yang terhubung secara realtime untuk memproses trigger dan interaksi data yang terkirim dari infrastruktur bot WhatsApp.
- **⏱Automated Cron Jobs**: Integrasi *background jobs* melalui pemanggilan *Cron* guna mengeksekusi operasi otomatis dalam sistem (seperti pengecekan notifikasi rutin).
- **Self-Generated API Docs**: Menyediakan dokumen interaktif berbasis [Swagger Configuration (OpenAPI)](https://swagger.io/) untuk memudahkan developer Frontend / Mobile dalam memahami endpoints.

## Stack Teknologi & Dependensi (Tech Stack)
- **Core Language:** [Go (Golang)](https://go.dev/) 1.24+
- **Database:** [MongoDB Native Driver](https://github.com/mongodb/mongo-go-driver)
- **Autentikasi & Keamanan:** `golang-jwt/jwt/v5`, `golang.org/x/oauth2`
- **Environment Management:** `joho/godotenv`
- **API Documentation Generation:** `swaggo/swag`, `swaggo/http-swagger/v2`

## 📁 Struktur Proyek Pendekatan Moduler
Proyek ini mengadaptasi pola arsitektur bersih yang dipisahkan ke beberapa modular folder, menjadikannya rapi untuk standar industri:
- `/api`: Menyediakan Main Handler Endpoint (berfungsi sebagai Entry-point untuk Vercel Serverless).
- `/controller`: Berisi serangkaian *Business Logic* yang mengolah request dan mengakses model Database.
- `/route`: Konfigurasi alur *Routing* endpoint serta *Middleware*.
- `/helper` & `/config`: Kode pemecah beban berisi konfigurasi koneksi MongoDB dan set of helper.
- `/docs`: Output JSON dan YAML *auto-generated* oleh sistem Swagger.

## Setup & Instalasi Lokal
Untuk menjalankan sistem API ini pada *machine* Anda, ikuti langkah-langkah di bawah ini:

1. **Clone repositori ini**
   ```bash
   git clone <url-repo-ini>
   cd kawai-backend
   ```
2. **Setup variasi *Environment***
   Salin `.env.example` ke file `.env` dan konfigurasikan secara mandiri (termasuk URI MongoDB, JWT Secret, ID Google OAuth, dan Port).
   ```bash
   PORT=3000
   MONGOSTRING=mongodb+srv://...
   JWT_SECRET=your_super_secret_jwt
   # GOOGLE_CLIENT_ID, dll
   ```
3. **Install Dependensi & Perbarui Swagger**
   ```bash
   go mod tidy
   swag init -g api/index.go
   ```
4. **Jalankan Aplikasi**
   ```bash
   go run main.go
   ```
Layanan akan berjalan dan dapat diakses di portal Swagger UI melalui `http://localhost:3000/swagger/index.html`.

## 🛡️ License
> Proyek ini dapat dimodifikasi atau digunakan hanya untuk tujuan *showcase* dan non-komersil, diset dalam lisensi tersendiri bagi pembuat repositori ini.
