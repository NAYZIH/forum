package handlers

import (
	"forum/backend/auth"
	"forum/backend/functions"
	"forum/backend/models"
	"html/template"
	"net/http"
	"strconv"
)

func ModerationHandler(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	session, err := auth.GetSession(sessionCookie.Value)
	if err != nil || session == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	user, err := functions.GetUserByID(session.UserID)
	if err != nil {
		http.Error(w, "Utilisateur introuvable", http.StatusInternalServerError)
		return
	}
	if user.Role != "modérateur" && user.Role != "administrateur" {
		http.Error(w, "Accès refusé", http.StatusForbidden)
		return
	}

	if r.Method == "GET" {
		posts, err := functions.GetPendingPosts()
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des postes", http.StatusInternalServerError)
			return
		}
		t, err := template.ParseFiles("frontend/templates/moderation.html")
		if err != nil {
			http.Error(w, "Erreur de template", http.StatusInternalServerError)
			return
		}
		data := struct {
			User  *models.User
			Posts []models.Post
		}{
			User:  user,
			Posts: posts,
		}
		t.Execute(w, data)
	} else if r.Method == "POST" {
		action := r.FormValue("action")
		postIDStr := r.FormValue("post_id")
		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			http.Error(w, "ID de post invalide", http.StatusBadRequest)
			return
		}
		if action == "approve" {
			err = functions.ApprovePost(postID)
		} else if action == "reject" {
			flag := r.FormValue("flag")
			err = functions.RejectPost(postID, flag)
		} else {
			http.Error(w, "Action invalide", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, "Erreur lors de la modération", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/moderation", http.StatusSeeOther)
	} else {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}
