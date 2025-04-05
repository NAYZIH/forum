package handlers

import (
	"forum/backend/auth"
	"forum/backend/functions"
	"forum/backend/models"
	"html/template"
	"log"
	"net/http"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}
	sessionID, _ := r.Cookie("session_id")
	var user *models.User
	var userID int
	var unreadCount int
	if sessionID != nil {
		session, err := auth.GetSession(sessionID.Value)
		if err == nil && session != nil {
			user, _ = functions.GetUserByID(session.UserID)
			userID = session.UserID
			unreadCount, _ = functions.GetUnreadNotificationCount(userID)
		}
	}
	posts, err := functions.GetPosts("", "", userID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	categories, err := functions.GetAllCategories()
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
		User        *models.User
		Posts       []models.Post
		Categories  []string
		UnreadCount int
	}{
		User:        user,
		Posts:       posts,
		Categories:  categories,
		UnreadCount: unreadCount,
	}
	t.Execute(w, data)
}
