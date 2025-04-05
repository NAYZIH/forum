package handlers

import (
	"forum/backend/auth"
	"forum/backend/functions"
	"forum/backend/models"
	"html/template"
	"log"
	"net/http"
)

func ErrorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	t, err := template.ParseFiles("frontend/templates/error.html")
	if err != nil {
		log.Println("Erreur de template :", err)
		http.Error(w, "Erreur interne", http.StatusInternalServerError)
		return
	}
	var user *models.User
	var unreadCount int
	sessionID, err := r.Cookie("session_id")
	if err == nil {
		session, err := auth.GetSession(sessionID.Value)
		if err == nil && session != nil {
			user, _ = functions.GetUserByID(session.UserID)
			unreadCount, _ = functions.GetUnreadNotificationCount(session.UserID)
		}
	}
	data := struct {
		User        *models.User
		UnreadCount int
		Status      int
	}{
		User:        user,
		UnreadCount: unreadCount,
		Status:      status,
	}
	t.Execute(w, data)
}
