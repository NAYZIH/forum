package handlers

import (
	"forum/backend/auth"
	"forum/backend/functions"
	"forum/backend/models"
	"html/template"
	"net/http"
	"strconv"
)

func AdminHandler(w http.ResponseWriter, r *http.Request) {
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
	adminUser, err := functions.GetUserByID(session.UserID)
	if err != nil {
		http.Error(w, "Utilisateur introuvable", http.StatusInternalServerError)
		return
	}
	if adminUser.Role != "administrateur" {
		http.Error(w, "Accès refusé", http.StatusForbidden)
		return
	}

	if r.Method == "GET" {
		users, err := functions.GetAllUsers()
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des utilisateurs", http.StatusInternalServerError)
			return
		}
		t, err := template.ParseFiles("frontend/templates/admin.html")
		if err != nil {
			http.Error(w, "Erreur de template", http.StatusInternalServerError)
			return
		}
		data := struct {
			Admin *models.User
			Users []models.User
		}{
			Admin: adminUser,
			Users: users,
		}
		t.Execute(w, data)
	} else if r.Method == "POST" {
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
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}
