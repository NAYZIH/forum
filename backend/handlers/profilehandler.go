package handlers

import (
	"forum/backend/auth"
	"forum/backend/functions"
	"forum/backend/models"
	"html/template"
	"net/http"
	"os"
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
	var unreadCount int
	sessionID, _ := r.Cookie("session_id")
	if sessionID != nil {
		session, err := auth.GetSession(sessionID.Value)
		if err == nil && session != nil {
			currentUserID = session.UserID
			unreadCount, _ = functions.GetUnreadNotificationCount(currentUserID)
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
		UnreadCount      int
	}{
		User:             targetUser,
		Posts:            posts,
		Comments:         comments,
		LikedPosts:       likedPosts,
		DislikedPosts:    dislikedPosts,
		LikedComments:    likedComments,
		DislikedComments: dislikedComments,
		IsOwnProfile:     currentUserID == targetUserID,
		UnreadCount:      unreadCount,
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
		avatars, err := getAvailableAvatars()
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		t, err := template.ParseFiles("frontend/templates/editprofile.html")
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		data := models.EditProfileData{
			User:    user,
			Avatars: avatars,
		}
		t.Execute(w, data)
	} else if r.Method == "POST" {
		currentPassword := r.FormValue("current_password")
		newUsername := r.FormValue("username")
		newEmail := r.FormValue("email")
		newBio := r.FormValue("bio")
		newAvatar := r.FormValue("avatar")
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword))
		if err != nil {
			avatars, _ := getAvailableAvatars()
			t, _ := template.ParseFiles("frontend/templates/editprofile.html")
			data := models.EditProfileData{
				User:    user,
				Avatars: avatars,
				Error:   "Mot de passe actuel incorrect",
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
			avatars, _ := getAvailableAvatars()
			t, _ := template.ParseFiles("frontend/templates/editprofile.html")
			data := models.EditProfileData{
				User:    user,
				Avatars: avatars,
				Error:   "Email déjà pris",
			}
			t.Execute(w, data)
			return
		}
		err = functions.UpdateUser(user.ID, newUsername, newEmail, newBio, newAvatar)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
	} else {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

func getAvailableAvatars() ([]string, error) {
	dir := "./frontend/static/images/profile"
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var avatars []string
	for _, file := range files {
		if !file.IsDir() {
			avatars = append(avatars, "/static/images/profile/"+file.Name())
		}
	}
	return avatars, nil
}
