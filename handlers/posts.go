package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"forum/database"
	"forum/middleware"
	"forum/models"
)

func CreatePost(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value(middleware.UserKey).(*models.User)

	topicID, err := strconv.Atoi(r.FormValue("topic_id"))
	content := strings.TrimSpace(r.FormValue("content"))
	imageURL := strings.TrimSpace(r.FormValue("image_url"))

	var parentID *int
	if pid, err := strconv.Atoi(r.FormValue("parent_id")); err == nil && pid > 0 {
		parentID = &pid
	}

	if err != nil || content == "" {
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	database.CreatePost(topicID, user.ID, content, imageURL, parentID)
	http.Redirect(w, r, "/topic/"+strconv.Itoa(topicID), http.StatusSeeOther)
}

func React(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value(middleware.UserKey).(*models.User)

	postID, err := strconv.Atoi(r.FormValue("post_id"))
	topicID := r.FormValue("topic_id")
	reactionType := r.FormValue("type")

	if err != nil || (reactionType != "like" && reactionType != "dislike") {
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	database.ToggleReaction(user.ID, postID, reactionType)
	http.Redirect(w, r, "/topic/"+topicID, http.StatusSeeOther)
}
