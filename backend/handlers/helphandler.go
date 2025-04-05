package handlers

import (
	"forum/backend/auth"
	"forum/backend/functions"
	"forum/backend/models"
	"html/template"
	"net/http"
)

func HelpHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/help" {
		ErrorHandler(w, r, http.StatusNotFound)
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

	t, err := template.ParseFiles("frontend/templates/help.html")
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	data := struct {
		User        *models.User
		UnreadCount int
	}{
		User:        user,
		UnreadCount: unreadCount,
	}
	t.Execute(w, data)
}
