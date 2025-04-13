package handlers

import (
	"forum/backend/auth"
	"forum/backend/functions"
	"forum/backend/models"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func CommentHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasPrefix(path, "/comment/") {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}
	parts := strings.Split(path[len("/comment/"):], "/")
	if len(parts) != 2 {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}
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

	comment, err := functions.GetCommentByID(id)
	if err != nil {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}

	user, err := functions.GetUserByID(session.UserID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	if session.UserID != comment.UserID && user.Role != "owner" {
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}

	if parts[1] == "edit" {
		if r.Method == "GET" {
			t, err := template.ParseFiles("frontend/templates/editcomment.html")
			if err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError)
				return
			}
			post, err := functions.GetPostByID(comment.PostID)
			if err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError)
				return
			}
			unreadCount, err := functions.GetUnreadNotificationCount(session.UserID)
			if err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError)
				return
			}
			data := struct {
				User        *models.User
				Post        *models.Post
				Comment     *models.Comment
				Error       string
				UnreadCount int
			}{
				User:        user,
				Post:        post,
				Comment:     comment,
				Error:       "",
				UnreadCount: unreadCount,
			}
			t.Execute(w, data)
		} else if r.Method == "POST" {
			err := r.ParseMultipartForm(maxUploadSize)
			if err != nil {
				http.Error(w, "Fichier trop volumineux", http.StatusBadRequest)
				return
			}
			content := r.FormValue("content")
			file, handler, err := r.FormFile("image")
			var imagePath string
			if err == nil {
				defer file.Close()
				if handler.Size > maxUploadSize {
					t, _ := template.ParseFiles("frontend/templates/editcomment.html")
					post, _ := functions.GetPostByID(comment.PostID)
					unreadCount, _ := functions.GetUnreadNotificationCount(session.UserID)
					data := struct {
						User        *models.User
						Post        *models.Post
						Comment     *models.Comment
						Error       string
						UnreadCount int
					}{
						User:        user,
						Post:        post,
						Comment:     comment,
						Error:       "L'image est trop volumineuse (max 20 Mo).",
						UnreadCount: unreadCount,
					}
					t.Execute(w, data)
					return
				}
				ext := strings.ToLower(filepath.Ext(handler.Filename))
				allowedExts := []string{".jpg", ".jpeg", ".png", ".gif"}
				validExt := false
				for _, allowed := range allowedExts {
					if ext == allowed {
						validExt = true
						break
					}
				}
				if !validExt {
					t, _ := template.ParseFiles("frontend/templates/editcomment.html")
					post, _ := functions.GetPostByID(comment.PostID)
					unreadCount, _ := functions.GetUnreadNotificationCount(session.UserID)
					data := struct {
						User        *models.User
						Post        *models.Post
						Comment     *models.Comment
						Error       string
						UnreadCount int
					}{
						User:        user,
						Post:        post,
						Comment:     comment,
						Error:       "Extension d'image non supportée (JPEG, PNG, GIF uniquement).",
						UnreadCount: unreadCount,
					}
					t.Execute(w, data)
					return
				}
				if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
					ErrorHandler(w, r, http.StatusInternalServerError)
					return
				}
				imagePath = uploadPath + strconv.FormatInt(time.Now().UnixNano(), 10) + ext
				f, err := os.OpenFile(imagePath, os.O_WRONLY|os.O_CREATE, 0666)
				if err != nil {
					ErrorHandler(w, r, http.StatusInternalServerError)
					return
				}
				defer f.Close()
				io.Copy(f, file)
				imagePath = "/static/images/comments/" + filepath.Base(imagePath)
			} else {
				imagePath = comment.ImagePath
			}

			err = functions.UpdateComment(comment.ID, content, imagePath)
			if err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/post/"+strconv.Itoa(comment.PostID), http.StatusSeeOther)
		} else {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		}
	} else if parts[1] == "delete" && r.Method == "POST" {
		err = functions.DeleteComment(id)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/post/"+strconv.Itoa(comment.PostID), http.StatusSeeOther)
	} else {
		ErrorHandler(w, r, http.StatusNotFound)
	}
}
