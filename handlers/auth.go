package handlers

import (
	"net/http"
	"time"

	"forum/database"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func RegisterGET(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "register.html", nil)
}

func RegisterPOST(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if username == "" || email == "" || password == "" {
		renderTemplate(w, "register.html", map[string]string{"Error": "Tous les champs sont requis"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		renderTemplate(w, "register.html", map[string]string{"Error": "Erreur interne"})
		return
	}

	if err := database.CreateUser(username, email, string(hash)); err != nil {
		renderTemplate(w, "register.html", map[string]string{"Error": "Nom d'utilisateur ou email déjà utilisé"})
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func LoginGET(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "login.html", nil)
}

func LoginPOST(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := database.GetUserByEmail(email)
	if err != nil {
		renderTemplate(w, "login.html", map[string]string{"Error": "Email ou mot de passe incorrect"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		renderTemplate(w, "login.html", map[string]string{"Error": "Email ou mot de passe incorrect"})
		return
	}

	sessionID := uuid.New().String()
	expires := time.Now().Add(24 * time.Hour)

	if err := database.CreateSession(sessionID, user.ID, expires); err != nil {
		renderTemplate(w, "login.html", map[string]string{"Error": "Erreur interne"})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionID,
		Expires:  expires,
		HttpOnly: true,
		Path:     "/",
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		database.DeleteSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Path:     "/",
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
