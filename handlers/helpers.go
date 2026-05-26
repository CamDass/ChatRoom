package handlers

import (
	"html/template"
	"log"
	"net/http"
)

var templates map[string]*template.Template

func InitTemplates() {
	templates = make(map[string]*template.Template)

	pages := []string{
		"index.html",
		"topic.html",
		"login.html",
		"register.html",
		"create_topic.html",
	}

	for _, page := range pages {
		t := template.Must(template.ParseFiles(
			"templates/base.html",
			"templates/"+page,
		))
		templates[page] = t
	}
}

func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	t, ok := templates[name]
	if !ok {
		http.Error(w, "Template introuvable", http.StatusInternalServerError)
		return
	}
	if err := t.ExecuteTemplate(w, "base.html", data); err != nil {
		log.Println("Template error:", err)
		http.Error(w, "Erreur interne", http.StatusInternalServerError)
	}
}
