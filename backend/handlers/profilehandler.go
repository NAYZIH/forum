package handlers

import (
	"forum/backend/auth"
	"forum/backend/functions"
	"forum/backend/models"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasPrefix(path, "/profile") {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}
	parts := strings.Split(strings.TrimPrefix(path, "/profile"), "/")
	var targetUserID int
	var err error
	if len(parts) > 1 && parts[1] != "" {
		targetUserID, err = strconv.Atoi(parts[1])
		if err != nil {
			http.Error(w, "ID invalide", http.StatusBadRequest)
			return
		}
	} else {
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
		targetUserID = session.UserID
	}
	targetUser, err := functions.GetUserByID(targetUserID)
	if err != nil {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}
	posts, err := functions.GetPosts("created", "", targetUserID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	comments, err := functions.GetCommentsByUserID(targetUserID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	likedPosts, err := functions.GetLikedPostsByUserID(targetUserID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	dislikedPosts, err := functions.GetDislikedPostsByUserID(targetUserID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	likedComments, err := functions.GetLikedCommentsByUserID(targetUserID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	dislikedComments, err := functions.GetDislikedCommentsByUserID(targetUserID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	var currentUserID int
	sessionID, _ := r.Cookie("session_id")
	if sessionID != nil {
		session, err := auth.GetSession(sessionID.Value)
		if err == nil && session != nil {
			currentUserID = session.UserID
		}
	}
	t, err := template.ParseFiles("frontend/templates/profile.html")
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	data := struct {
		User             *models.User
		Posts            []models.Post
		Comments         []models.Comment
		LikedPosts       []models.Post
		DislikedPosts    []models.Post
		LikedComments    []models.Comment
		DislikedComments []models.Comment
		IsOwnProfile     bool
	}{
		User:             targetUser,
		Posts:            posts,
		Comments:         comments,
		LikedPosts:       likedPosts,
		DislikedPosts:    dislikedPosts,
		LikedComments:    likedComments,
		DislikedComments: dislikedComments,
		IsOwnProfile:     currentUserID == targetUserID,
	}
	t.Execute(w, data)
}

func EditProfileHandler(w http.ResponseWriter, r *http.Request) {
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
	if r.Method == "GET" {
		t, err := template.ParseFiles("frontend/templates/editprofile.html")
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		data := models.EditProfileData{
			User: user,
		}
		t.Execute(w, data)
	} else if r.Method == "POST" {
		currentPassword := r.FormValue("current_password")
		newUsername := r.FormValue("username")
		newEmail := r.FormValue("email")
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword))
		if err != nil {
			t, _ := template.ParseFiles("frontend/templates/editprofile.html")
			data := models.EditProfileData{
				User:  user,
				Error: "Mot de passe actuel incorrect",
			}
			t.Execute(w, data)
			return
		}
		exists, err := functions.EmailExists(newEmail, user.ID)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		if exists {
			t, _ := template.ParseFiles("frontend/templates/editprofile.html")
			data := models.EditProfileData{
				User:  user,
				Error: "Email déjà pris",
			}
			t.Execute(w, data)
			return
		}
		err = functions.UpdateUser(user.ID, newUsername, newEmail)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
	} else {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}
