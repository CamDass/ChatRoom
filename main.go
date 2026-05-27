package main

import (
	"log"
	"net/http"
	"time"

	"forum/database"
	"forum/handlers"
	"forum/middleware"
)

func main() {
	database.InitDB()
	handlers.InitTemplates()

	go func() {
		for {
			time.Sleep(1 * time.Hour)
			database.CleanExpiredSessions()
		}
	}()

	http.HandleFunc("/login", middleware.Auth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.LoginPOST(w, r)
		} else {
			handlers.LoginGET(w, r)
		}
	}))
	http.HandleFunc("/register", middleware.Auth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.RegisterPOST(w, r)
		} else {
			handlers.RegisterGET(w, r)
		}
	}))
	http.HandleFunc("/logout", middleware.RequireAuth(handlers.Logout))

	http.HandleFunc("/topic/create", middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.CreateTopicPOST(w, r)
		} else {
			handlers.CreateTopicGET(w, r)
		}
	}))
	http.HandleFunc("/post/create", middleware.RequireAuth(handlers.CreatePost))
	http.HandleFunc("/post/react", middleware.RequireAuth(handlers.React))

	http.HandleFunc("/topic/", middleware.Auth(handlers.TopicView))
	http.HandleFunc("/category/", middleware.Auth(handlers.CategoryView))
	http.HandleFunc("/search", middleware.Auth(handlers.Search))

	http.HandleFunc("/user/", middleware.Auth(handlers.ProfileView))
	http.HandleFunc("/profile/update", middleware.RequireAuth(handlers.ProfileUpdate))
	http.HandleFunc("/api/users", middleware.Auth(handlers.SearchUsers))

	http.HandleFunc("/", middleware.Auth(handlers.Home))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Serveur démarré sur http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
