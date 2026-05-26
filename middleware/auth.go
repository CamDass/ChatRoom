package middleware

import (
	"context"
	"net/http"

	"forum/database"
	"forum/models"
)

type contextKey string

const UserKey contextKey = "user"

// GetUserFromRequest retourne l'utilisateur connecté ou nil
func GetUserFromRequest(r *http.Request) *models.User {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil
	}

	session, err := database.GetSession(cookie.Value)
	if err != nil {
		return nil
	}

	user, err := database.GetUserByID(session.UserID)
	if err != nil {
		return nil
	}

	return user
}

// Auth injecte l'utilisateur dans le contexte à chaque requête
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromRequest(r)
		ctx := context.WithValue(r.Context(), UserKey, user)
		next(w, r.WithContext(ctx))
	}
}

// RequireAuth redirige vers /login si non connecté
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromRequest(r)
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		ctx := context.WithValue(r.Context(), UserKey, user)
		next(w, r.WithContext(ctx))
	}
}
