package route

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kawai-org/kawai-backend/config"
	"github.com/kawai-org/kawai-backend/controller"
)

// --- MIDDLEWARE AUTHENTICATION ---
// Fungsi ini mengecek apakah User membawa "Karcis" (Token) yang valid
func MiddlewareAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Ambil Header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: Token tidak ditemukan", http.StatusUnauthorized)
			return
		}

		// 2. Format harus "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Unauthorized: Format token salah", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]

		// 3. Validasi Token dengan Secret Key
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized: Token tidak valid atau expired", http.StatusUnauthorized)
			return
		}

		// 4. Simpan Data User (No HP / Role) ke Context agar bisa dibaca Controller
		ctx := context.WithValue(r.Context(), "user", claims)
		next(w, r.WithContext(ctx))
	}
}

// --- ROUTING URL ---
func URL(w http.ResponseWriter, r *http.Request) {
	// CORS Config (Agar frontend bisa akses)
	if config.SetAccessControlHeaders(w, r) {
		return
	}
	config.SetEnv()

	path := r.URL.Path
	method := r.Method

	switch {
	
	// 1. WEBHOOK
	case method == "POST" && strings.HasPrefix(path, "/webhook/nomor/"):
		controller.PostInboxNomor(w, r)

	// 2. ADMIN LOGIN
	case method == "POST" && path == "/api/admin/login":
		controller.LoginAdmin(w, r)
	case method == "GET" && path == "/api/admin/users":
		MiddlewareAuth(controller.GetAllUsers)(w, r)

	// 3. DASHBOARD USER - NOTES (CRUD)
	case method == "GET" && path == "/api/dashboard/notes":
		MiddlewareAuth(controller.GetMyNotes)(w, r)
	
	// Handle Update & Delete Note (Pakai HasPrefix karena ada ID dibelakangnya)
	case method == "PUT" && strings.HasPrefix(path, "/api/notes/"):
		MiddlewareAuth(controller.UpdateNote)(w, r)
	case method == "DELETE" && strings.HasPrefix(path, "/api/notes/"):
		MiddlewareAuth(controller.DeleteNote)(w, r)

	// 4. DASHBOARD USER - REMINDERS (CRUD)
	case method == "GET" && path == "/api/dashboard/reminders":
		MiddlewareAuth(controller.GetReminders)(w, r)
	case method == "PUT" && strings.HasPrefix(path, "/api/reminders/"):
		MiddlewareAuth(controller.UpdateReminder)(w, r)

	// 5. CRON JOB & GOOGLE AUTH (Existing)
	case method == "GET" && path == "/api/cron":
		controller.HandleCron(w, r)
	case method == "GET" && path == "/api/auth/google/login":
		controller.GoogleLogin(w, r)
	case method == "GET" && path == "/api/auth/google/callback":
		controller.GoogleCallback(w, r)

	// 6. HOME
	case method == "GET" && path == "/":
		controller.GetHome(w, r)

	default:
		controller.NotFound(w, r)
	}
}