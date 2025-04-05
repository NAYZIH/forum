package handlers

import (
	"forum/backend/auth"
	"forum/backend/functions"
	"forum/backend/models"
	"html/template"
	"net/http"
)

func NotificationHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/notification" {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}

	sessionID, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	session, err := auth.GetSession(sessionID.Value)
	if err != nil || session == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := functions.GetUserByID(session.UserID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	notifications, err := functions.GetNotificationsByUserID(session.UserID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	unreadCount, err := functions.GetUnreadNotificationCount(session.UserID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		t, err := template.ParseFiles("frontend/templates/notification.html")
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		data := struct {
			User          *models.User
			Notifications []models.Notification
			UnreadCount   int
		}{
			User:          user,
			Notifications: notifications,
			UnreadCount:   unreadCount,
		}
		err = t.Execute(w, data)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}

		err = functions.MarkNotificationsAsRead(session.UserID)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}
