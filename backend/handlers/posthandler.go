package handlers

import (
	"forum/backend/auth"
	"forum/backend/functions"
	"forum/backend/models"
	"html/template"
	"net/http"
	"strconv"
	"strings"
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
		if email == "" || username == "" || password == "" {
			http.Error(w, "Tous les champs sont requis", http.StatusBadRequest)
			return
		}
		if _, err := functions.GetUserByEmail(email); err == nil {
			http.Error(w, "Email déjà pris", http.StatusConflict)
			return
		}
		err := functions.CreateUser(email, username, password)
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
		email := r.FormValue("email")
		password := r.FormValue("password")
		user, err := functions.Authenticate(email, password)
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

func NewPostHandler(w http.ResponseWriter, r *http.Request) {
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

	categories, err := functions.GetAllCategories()
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		t, err := template.ParseFiles("frontend/templates/newpost.html")
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		data := struct {
			Categories []string
			Error      string
		}{
			Categories: categories,
			Error:      "",
		}
		t.Execute(w, data)
	} else if r.Method == "POST" {
		title := r.FormValue("title")
		content := r.FormValue("content")
		selectedCategories := r.Form["categories[]"]
		newCategoriesStr := r.FormValue("new_categories")

		newCategories := strings.Split(newCategoriesStr, ",")
		for i, cat := range newCategories {
			newCategories[i] = strings.TrimSpace(cat)
		}

		var allCategories []string
		for _, cat := range append(selectedCategories, newCategories...) {
			if cat != "" {
				allCategories = append(allCategories, cat)
			}
		}

		if len(allCategories) == 0 {
			t, err := template.ParseFiles("frontend/templates/newpost.html")
			if err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError)
				return
			}
			data := struct {
				Categories []string
				Error      string
			}{
				Categories: categories,
				Error:      "Veuillez sélectionner ou créer au moins une catégorie.",
			}
			t.Execute(w, data)
			return
		}

		err = functions.CreatePost(session.UserID, title, content, allCategories)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasPrefix(path, "/post/") {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}
	parts := strings.Split(path[len("/post/"):], "/")
	if len(parts) == 1 && r.Method == "GET" {
		id, err := strconv.Atoi(parts[0])
		if err != nil {
			http.Error(w, "ID invalide", http.StatusBadRequest)
			return
		}
		post, err := functions.GetPostByID(id)
		if err != nil {
			ErrorHandler(w, r, http.StatusNotFound)
			return
		}
		comments, err := functions.GetCommentsByPostID(id)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		t, err := template.ParseFiles("frontend/templates/post.html")
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		data := struct {
			Post     *models.Post
			Comments []models.Comment
			User     *models.User
		}{
			Post:     post,
			Comments: comments,
		}
		if cookie, err := r.Cookie("session_id"); err == nil {
			if session, err := auth.GetSession(cookie.Value); err == nil && session != nil {
				data.User, _ = functions.GetUserByID(session.UserID)
			}
		}
		t.Execute(w, data)
	} else if len(parts) == 2 && parts[1] == "comment" && r.Method == "POST" {
		id, err := strconv.Atoi(parts[0])
		if err != nil {
			http.Error(w, "ID invalide", http.StatusBadRequest)
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
		content := r.FormValue("content")
		err = functions.CreateComment(id, session.UserID, content)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/post/"+parts[0], http.StatusSeeOther)
	} else {
		ErrorHandler(w, r, http.StatusNotFound)
	}
}
