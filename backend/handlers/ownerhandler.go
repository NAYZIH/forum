package handlers

import (
	"forum/backend/auth"
	"forum/backend/functions"
	"forum/backend/models"
	"html/template"
	"net/http"
	"strconv"
)

func OwnerHandler(w http.ResponseWriter, r *http.Request) {
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
	ownerUser, err := functions.GetUserByID(session.UserID)
	if err != nil {
		http.Error(w, "Utilisateur introuvable", http.StatusInternalServerError)
		return
	}
	if ownerUser.Role != "owner" {
		http.Error(w, "Accès refusé", http.StatusForbidden)
		return
	}

	if r.Method == "POST" {
		action := r.FormValue("action")
		switch action {
		case "force_logout":
			userIDStr := r.FormValue("user_id")
			userID, err := strconv.Atoi(userIDStr)
			if err != nil {
				http.Error(w, "ID de l'utilisateur invalide", http.StatusBadRequest)
				return
			}
			err = functions.DeleteUserSessions(userID)
			if err != nil {
				http.Error(w, "Erreur lors de la déconnexion de l'utilisateur", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/owner", http.StatusSeeOther)
			return
		case "update_user_role":
			userIDStr := r.FormValue("user_id")
			role := r.FormValue("role")
			userID, err := strconv.Atoi(userIDStr)
			if err != nil {
				http.Error(w, "ID de l'utilisateur invalide", http.StatusBadRequest)
				return
			}
			err = functions.UpdateUserRole(userID, role)
			if err != nil {
				http.Error(w, "Erreur lors de la mise à jour du rôle", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/owner", http.StatusSeeOther)
			return
		case "delete_user":
			userIDStr := r.FormValue("user_id")
			userID, err := strconv.Atoi(userIDStr)
			if err != nil {
				http.Error(w, "ID de l'utilisateur invalide", http.StatusBadRequest)
				return
			}
			err = functions.DeleteUser(userID)
			if err != nil {
				http.Error(w, "Erreur lors de la suppression de l'utilisateur", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/owner", http.StatusSeeOther)
			return
		case "delete_post":
			postIDStr := r.FormValue("post_id")
			postID, err := strconv.Atoi(postIDStr)
			if err != nil {
				http.Error(w, "ID du post invalide", http.StatusBadRequest)
				return
			}
			err = functions.DeletePost(postID)
			if err != nil {
				http.Error(w, "Erreur lors de la suppression du post", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/owner", http.StatusSeeOther)
			return
		}
	}

	users, err := functions.GetAllUsers()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des utilisateurs", http.StatusInternalServerError)
		return
	}
	posts, err := functions.GetPosts("", "", 0)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des posts", http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles("frontend/templates/owner.html")
	if err != nil {
		http.Error(w, "Erreur de template", http.StatusInternalServerError)
		return
	}
	data := struct {
		Owner *models.User
		Users []models.User
		Posts []models.Post
	}{
		Owner: ownerUser,
		Users: users,
		Posts: posts,
	}
	t.Execute(w, data)
}
