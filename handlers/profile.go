package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"forum/database"
	"forum/middleware"
	"forum/models"
)

func ProfileView(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/user/")
	if username == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	currentUser, _ := r.Context().Value(middleware.UserKey).(*models.User)

	profileUser, err := database.GetUserByUsername(username)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	favs, _ := database.GetFavoriteTopics(profileUser.ID)
	cats, _ := database.GetCategories()
	isOwn := currentUser != nil && currentUser.ID == profileUser.ID

	renderTemplate(w, "profile.html", models.PageData{
		User:           currentUser,
		Categories:     cats,
		ProfileUser:    profileUser,
		FavoriteTopics: favs,
		IsOwnProfile:   isOwn,
	})
}

func ProfileUpdate(w http.ResponseWriter, r *http.Request) {
	currentUser, _ := r.Context().Value(middleware.UserKey).(*models.User)

	r.ParseMultipartForm(5 << 20)
	bio := strings.TrimSpace(r.FormValue("bio"))
	avatarURL := currentUser.AvatarURL

	file, header, err := r.FormFile("avatar")
	if err == nil {
		defer file.Close()
		ext := strings.ToLower(filepath.Ext(header.Filename))
		allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
		if allowed[ext] {
			os.MkdirAll("static/uploads", 0755)
			filename := "avatar_" + strings.ReplaceAll(currentUser.Username, " ", "_") + ext
			dst, err := os.Create("static/uploads/" + filename)
			if err == nil {
				defer dst.Close()
				io.Copy(dst, file)
				avatarURL = "/static/uploads/" + filename
			}
		}
	}

	database.UpdateUserProfile(currentUser.ID, bio, avatarURL)
	http.Redirect(w, r, "/user/"+currentUser.Username, http.StatusSeeOther)
}
