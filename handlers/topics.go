package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"forum/database"
	"forum/middleware"
	"forum/models"
)

func Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	user, _ := r.Context().Value(middleware.UserKey).(*models.User)
	filter := r.URL.Query().Get("filter")

	userID := 0
	if user != nil {
		userID = user.ID
	}

	topics, err := database.GetTopics(filter, userID)
	if err != nil {
		http.Error(w, "Erreur base de données", http.StatusInternalServerError)
		return
	}

	cats, _ := database.GetCategories()

	renderTemplate(w, "index.html", models.PageData{
		User:       user,
		Topics:     topics,
		Categories: cats,
		Filter:     filter,
	})
}

func CategoryView(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/category/")
	user, _ := r.Context().Value(middleware.UserKey).(*models.User)

	topics, err := database.GetTopicsByCategory(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	cats, _ := database.GetCategories()
	cat, _ := database.GetCategoryBySlug(slug)

	renderTemplate(w, "index.html", models.PageData{
		User:       user,
		Topics:     topics,
		Categories: cats,
		Category:   cat,
	})
}

func TopicView(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/topic/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	user, _ := r.Context().Value(middleware.UserKey).(*models.User)
	userID := 0
	if user != nil {
		userID = user.ID
	}

	topic, err := database.GetTopicByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	posts, err := database.GetPostsByTopic(id, userID)
	if err != nil {
		http.Error(w, "Erreur base de données", http.StatusInternalServerError)
		return
	}

	cats, _ := database.GetCategories()

	renderTemplate(w, "topic.html", models.PageData{
		User:       user,
		Topic:      topic,
		Posts:      posts,
		Categories: cats,
	})
}

func CreateTopicGET(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value(middleware.UserKey).(*models.User)
	cats, _ := database.GetCategories()
	renderTemplate(w, "create_topic.html", models.PageData{User: user, Categories: cats})
}

func CreateTopicPOST(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value(middleware.UserKey).(*models.User)

	title := strings.TrimSpace(r.FormValue("title"))
	catID, err := strconv.Atoi(r.FormValue("category_id"))
	if err != nil || title == "" {
		cats, _ := database.GetCategories()
		renderTemplate(w, "create_topic.html", models.PageData{
			User: user, Categories: cats, Error: "Titre et catégorie requis",
		})
		return
	}

	topicID, err := database.CreateTopic(user.ID, catID, title)
	if err != nil {
		http.Error(w, "Erreur création topic", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/topic/"+strconv.FormatInt(topicID, 10), http.StatusSeeOther)
}

func Search(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value(middleware.UserKey).(*models.User)
	q := strings.TrimSpace(r.URL.Query().Get("q"))

	var topics []models.Topic
	if q != "" {
		topics, _ = database.SearchTopics(q)
	}

	cats, _ := database.GetCategories()

	renderTemplate(w, "index.html", models.PageData{
		User:       user,
		Topics:     topics,
		Categories: cats,
		Filter:     "search",
		Search:     q,
	})
}

func SearchUsers(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	users, _ := database.SearchUsernames(q)
	w.Header().Set("Content-Type", "application/json")

	var result []string
	for _, u := range users {
		result = append(result, u)
	}

	json.NewEncoder(w).Encode(result)
}
