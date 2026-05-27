package handlers

import (
	"html/template"
	"log"
	"net/http"
	"regexp"
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
		"profile.html",
	}

	funcMap := template.FuncMap{
		"parseMentions": parseMentions,
	}

	for _, page := range pages {
		t := template.Must(template.New("").Funcs(funcMap).ParseFiles(
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

var mentionRegex = regexp.MustCompile(`@(\w+)`)

func parseMentions(content string) template.HTML {
	escaped := template.HTMLEscapeString(content)
	result := mentionRegex.ReplaceAllStringFunc(escaped, func(match string) string {
		username := match[1:]
		return `<span class="mention-link" data-username="` + username + `">` + match + `</span>`
	})
	return template.HTML(result)
}
