package handlers

import (
	"forum/backend/auth"
	"forum/backend/functions"
	"forum/backend/models"
	"html/template"
	"log"
	"net/http"
	"strings"
)

func FilterHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasPrefix(path, "/filter/") {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}
	filter := strings.TrimPrefix(path, "/filter/")
	var userID int
	sessionID, err := r.Cookie("session_id")
	if err == nil {
		session, err := auth.GetSession(sessionID.Value)
		if err == nil && session != nil {
			userID = session.UserID
		}
	}
	if (filter == "created" || filter == "liked") && userID == 0 {
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}
	var posts []models.Post
	if filter == "category" {
		category := r.URL.Query().Get("category")
		posts, err = functions.GetPosts("category", category, userID)
	} else {
		posts, err = functions.GetPosts(filter, "", userID)
	}
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	t, err := template.ParseFiles("frontend/templates/index.html")
	if err != nil {
		log.Println("Erreur de template :", err)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	data := struct {
		User       *models.User
		Posts      []models.Post
		Categories []string
	}{
		Posts: posts,
	}
	if userID != 0 {
		data.User, _ = functions.GetUserByID(userID)
	}
	data.Categories, _ = functions.GetAllCategories()
	t.Execute(w, data)
}
