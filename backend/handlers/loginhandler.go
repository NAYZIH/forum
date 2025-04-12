package handlers

import (
	"forum/backend/auth"
	"forum/backend/functions"
	"html/template"
	"net/http"
	"time"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, err := template.ParseFiles("frontend/templates/register.html")
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		t.Execute(w, nil)
	} else if r.Method == "POST" {
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")
		bio := r.FormValue("bio")
		if email == "" || username == "" || password == "" {
			http.Error(w, "Tous les champs sont requis", http.StatusBadRequest)
			return
		}
		if _, err := functions.GetUserByEmail(email); err == nil {
			http.Error(w, "Email déjà pris", http.StatusConflict)
			return
		}
		err := functions.CreateUser(email, username, password, bio)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, err := template.ParseFiles("frontend/templates/login.html")
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		t.Execute(w, nil)
	} else if r.Method == "POST" {
		identifier := r.FormValue("identifier")
		password := r.FormValue("password")
		user, err := functions.Authenticate(identifier, password)
		if err != nil {
			http.Error(w, "Identifiants invalides", http.StatusUnauthorized)
			return
		}
		session, err := auth.CreateSession(user.ID)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:    "session_id",
			Value:   session.SessionID,
			Expires: session.ExpiresAt,
			Path:    "/",
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		cookie, err := r.Cookie("session_id")
		if err == nil {
			auth.DeleteSession(cookie.Value)
		}
		http.SetCookie(w, &http.Cookie{
			Name:    "session_id",
			Value:   "",
			Expires: time.Now().Add(-1 * time.Hour),
			Path:    "/",
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}
