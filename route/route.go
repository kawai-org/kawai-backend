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

func MiddlewareAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "user", claims)
		next(w, r.WithContext(ctx))
	}
}

func URL(w http.ResponseWriter, r *http.Request) {
	if config.SetAccessControlHeaders(w, r) {
		return
	}
	config.SetEnv()

	path := r.URL.Path
	method := r.Method

	switch {

	// --- PUBLIC ROUTES ---
	case method == "POST" && strings.HasPrefix(path, "/webhook/nomor/"):
		controller.PostInboxNomor(w, r)
	case method == "GET" && path == "/api/cron":
		controller.HandleCron(w, r)
	case method == "POST" && path == "/api/admin/login": // Login Admin
		controller.LoginAdmin(w, r)
	case method == "POST" && path == "/api/register":
        controller.Register(w, r)

	// --- GOOGLE AUTH ---
	case method == "GET" && path == "/api/auth/google/login":
		controller.GoogleLogin(w, r)
	case method == "GET" && path == "/api/auth/google/callback":
		controller.GoogleCallback(w, r)

	// --- ADMIN DASHBOARD (PROTECTED) ---
	case method == "GET" && path == "/api/admin/users": // Lihat User
		MiddlewareAuth(controller.GetAllUsers)(w, r)
	case method == "GET" && path == "/api/admin/stats": // Lihat Angka2
		MiddlewareAuth(controller.GetSystemStats)(w, r)
	case method == "POST" && path == "/api/admin/ban": // Blokir User
		MiddlewareAuth(controller.BanUser)(w, r)

	// --- USER DASHBOARD (PROTECTED) ---
	
	// Notes (List, Edit, Delete)
	case method == "GET" && path == "/api/dashboard/notes":
		MiddlewareAuth(controller.GetMyNotes)(w, r)
	case method == "GET" && path == "/api/dashboard/notes/:id/detail":
		MiddlewareAuth(controller.GetNoteDetailWithTags)(w, r)
	case method == "PUT" && strings.HasPrefix(path, "/api/notes/"):
		MiddlewareAuth(controller.UpdateNote)(w, r)
	case method == "DELETE" && strings.HasPrefix(path, "/api/notes/"):
		MiddlewareAuth(controller.DeleteNote)(w, r)

	// Reminders (List, Edit, Delete)
	case method == "GET" && path == "/api/dashboard/reminders":
		MiddlewareAuth(controller.GetReminders)(w, r)
	case method == "PUT" && strings.HasPrefix(path, "/api/reminders/"):
		MiddlewareAuth(controller.UpdateReminder)(w, r)
	case method == "DELETE" && strings.HasPrefix(path, "/api/reminders/"): // BARU
		MiddlewareAuth(controller.DeleteReminder)(w, r)

	// Default
	case method == "GET" && path == "/":
		controller.GetHome(w, r)
	default:
		controller.NotFound(w, r)
	}
}