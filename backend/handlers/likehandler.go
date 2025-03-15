package handlers

import (
	"forum/backend/auth"
	"forum/backend/functions"
	"net/http"
	"strconv"
	"strings"
)

func LikeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
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
	path := r.URL.Path
	if !strings.HasPrefix(path, "/like/") {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}
	parts := strings.Split(path[len("/like/"):], "/")
	if len(parts) != 2 {
		ErrorHandler(w, r, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}
	action := r.FormValue("action")
	value := 1
	if action == "dislike" {
		value = -1
	}
	if parts[0] == "post" {
		err = functions.LikePost(session.UserID, id, value)
	} else if parts[0] == "comment" {
		err = functions.LikeComment(session.UserID, id, value)
	} else {
		ErrorHandler(w, r, http.StatusBadRequest)
		return
	}
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
}
