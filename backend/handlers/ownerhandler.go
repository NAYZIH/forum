package handlers

import (
	"database/sql"
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

	switch r.Method {
	case "GET":
		users, err := functions.GetAllUsers()
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des utilisateurs", http.StatusInternalServerError)
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
		}{
			Owner: ownerUser,
			Users: users,
		}
		t.Execute(w, data)

	case "POST":
		action := r.FormValue("action")
		userIDStr := r.FormValue("user_id")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "ID de l'utilisateur invalide", http.StatusBadRequest)
			return
		}

		switch action {
		case "update":
			email := r.FormValue("email")
			username := r.FormValue("username")
			bio := r.FormValue("bio")
			avatarPath := r.FormValue("avatar_path")
			if avatarPath == "" {
				avatarPath = "/static/images/profile/default.png"
			}
			err = functions.UpdateUser(userID, username, email, bio, avatarPath)
			if err != nil {
				http.Error(w, "Erreur lors de la mise à jour de l'utilisateur", http.StatusInternalServerError)
				return
			}

		case "delete":
			if userID == ownerUser.ID {
				http.Error(w, "L'owner ne peut pas se supprimer lui-même", http.StatusForbidden)
				return
			}
			err = functions.DeleteUser(userID)
			if err != nil {
				http.Error(w, "Erreur lors de la suppression de l'utilisateur", http.StatusInternalServerError)
				return
			}

		case "update_role":
			role := r.FormValue("role")
			if userID == ownerUser.ID && role != "owner" {
				http.Error(w, "L'owner ne peut pas modifier son propre rôle", http.StatusForbidden)
				return
			}
			err = functions.UpdateUserRole(userID, role)
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "Un seul owner est autorisé", http.StatusBadRequest)
				} else {
					http.Error(w, "Erreur lors de la mise à jour du rôle", http.StatusInternalServerError)
				}
				return
			}

		case "force_logout":
			err = functions.ForceLogout(userID)
			if err != nil {
				http.Error(w, "Erreur lors de la déconnexion forcée", http.StatusInternalServerError)
				return
			}
		}
		http.Redirect(w, r, "/owner", http.StatusSeeOther)

	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}
