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

const maxUploadSize = 20 * 1024 * 1024
const uploadPath = "./frontend/static/images/posts/"

func NewPostHandler(w http.ResponseWriter, r *http.Request) {
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
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	unreadCount, err := functions.GetUnreadNotificationCount(session.UserID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		categories, err := functions.GetAllCategories()
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		t, err := template.ParseFiles("frontend/templates/newpost.html")
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		data := struct {
			User        *models.User
			Categories  []string
			Error       string
			UnreadCount int
		}{
			User:        user,
			Categories:  categories,
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
		title := r.FormValue("title")
		content := r.FormValue("content")
		categories := r.Form["categories[]"]
		newCategories := strings.Split(r.FormValue("new_categories"), ",")
		for _, cat := range newCategories {
			if strings.TrimSpace(cat) != "" {
				categories = append(categories, strings.TrimSpace(cat))
			}
		}
		file, handler, err := r.FormFile("image")
		var imagePath string
		if err == nil {
			defer file.Close()
			if handler.Size > maxUploadSize {
				t, _ := template.ParseFiles("frontend/templates/newpost.html")
				cats, _ := functions.GetAllCategories()
				data := struct {
					User        *models.User
					Categories  []string
					Error       string
					UnreadCount int
				}{
					User:        user,
					Categories:  cats,
					Error:       "Fichier trop volumineux (max 20 Mo)",
					UnreadCount: unreadCount,
				}
				t.Execute(w, data)
				return
			}
			ext := strings.ToLower(filepath.Ext(handler.Filename))
			if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
				t, _ := template.ParseFiles("frontend/templates/newpost.html")
				cats, _ := functions.GetAllCategories()
				data := struct {
					User        *models.User
					Categories  []string
					Error       string
					UnreadCount int
				}{
					User:        user,
					Categories:  cats,
					Error:       "Type de fichier non autorisé",
					UnreadCount: unreadCount,
				}
				t.Execute(w, data)
				return
			}
			err = os.MkdirAll(uploadPath, os.ModePerm)
			if err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError)
				return
			}
			filename := strconv.FormatInt(time.Now().UnixNano(), 10) + ext
			imagePath = uploadPath + filename
			f, err := os.OpenFile(imagePath, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError)
				return
			}
			defer f.Close()
			io.Copy(f, file)
			imagePath = "/static/images/posts/" + filename
		}
		err = functions.CreatePost(user.ID, title, content, categories, imagePath)
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
	if len(parts) == 0 {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}
	idStr := parts[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	sessionCookie, _ := r.Cookie("session_id")
	var user *models.User
	var unreadCount int
	if sessionCookie != nil {
		session, err := auth.GetSession(sessionCookie.Value)
		if err == nil && session != nil {
			user, err = functions.GetUserByID(session.UserID)
			if err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError)
				return
			}
			unreadCount, err = functions.GetUnreadNotificationCount(session.UserID)
			if err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError)
				return
			}
		}
	}

	post, err := functions.GetPostByID(id)
	if err != nil {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}

	isOwner := user != nil && user.Role == "owner"

	if len(parts) == 1 {
		if r.Method == "GET" {
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
				User        *models.User
				Post        *models.Post
				Comments    []models.Comment
				UnreadCount int
			}{
				User:        user,
				Post:        post,
				Comments:    comments,
				UnreadCount: unreadCount,
			}
			t.Execute(w, data)
		} else if r.Method == "POST" && r.URL.Path == "/post/"+idStr+"/comment" {
			if user == nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
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
					http.Error(w, "Fichier trop volumineux (max 20 Mo)", http.StatusBadRequest)
					return
				}
				ext := strings.ToLower(filepath.Ext(handler.Filename))
				if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
					http.Error(w, "Type de fichier non autorisé", http.StatusBadRequest)
					return
				}
				err = os.MkdirAll(uploadPath, os.ModePerm)
				if err != nil {
					ErrorHandler(w, r, http.StatusInternalServerError)
					return
				}
				filename := strconv.FormatInt(time.Now().UnixNano(), 10) + ext
				imagePath = uploadPath + filename
				f, err := os.OpenFile(imagePath, os.O_WRONLY|os.O_CREATE, 0666)
				if err != nil {
					ErrorHandler(w, r, http.StatusInternalServerError)
					return
				}
				defer f.Close()
				io.Copy(f, file)
				imagePath = "/static/images/posts/" + filename
			}
			err = functions.CreateComment(id, user.ID, content, imagePath)
			if err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/post/"+idStr, http.StatusSeeOther)
		}
	} else if len(parts) == 2 {
		if parts[1] == "edit" {
			if user == nil || (!isOwner && user.ID != post.UserID) {
				http.Error(w, "Non autorisé", http.StatusUnauthorized)
				return
			}
			if r.Method == "GET" {
				categories, err := functions.GetAllCategories()
				if err != nil {
					ErrorHandler(w, r, http.StatusInternalServerError)
					return
				}
				t, err := template.ParseFiles("frontend/templates/editpost.html")
				if err != nil {
					ErrorHandler(w, r, http.StatusInternalServerError)
					return
				}
				data := struct {
					User        *models.User
					Post        *models.Post
					Categories  []string
					Error       string
					UnreadCount int
				}{
					User:        user,
					Post:        post,
					Categories:  categories,
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
				title := r.FormValue("title")
				content := r.FormValue("content")
				categories := r.Form["categories[]"]
				newCategories := strings.Split(r.FormValue("new_categories"), ",")
				for _, cat := range newCategories {
					if strings.TrimSpace(cat) != "" {
						categories = append(categories, strings.TrimSpace(cat))
					}
				}
				file, handler, err := r.FormFile("image")
				var imagePath string
				if err == nil {
					defer file.Close()
					if handler.Size > maxUploadSize {
						t, _ := template.ParseFiles("frontend/templates/editpost.html")
						cats, _ := functions.GetAllCategories()
						data := struct {
							User        *models.User
							Post        *models.Post
							Categories  []string
							Error       string
							UnreadCount int
						}{
							User:        user,
							Post:        post,
							Categories:  cats,
							Error:       "Fichier trop volumineux (max 20 Mo)",
							UnreadCount: unreadCount,
						}
						t.Execute(w, data)
						return
					}
					ext := strings.ToLower(filepath.Ext(handler.Filename))
					if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
						t, _ := template.ParseFiles("frontend/templates/editpost.html")
						cats, _ := functions.GetAllCategories()
						data := struct {
							User        *models.User
							Post        *models.Post
							Categories  []string
							Error       string
							UnreadCount int
						}{
							User:        user,
							Post:        post,
							Categories:  cats,
							Error:       "Type de fichier non autorisé",
							UnreadCount: unreadCount,
						}
						t.Execute(w, data)
						return
					}
					err = os.MkdirAll(uploadPath, os.ModePerm)
					if err != nil {
						ErrorHandler(w, r, http.StatusInternalServerError)
						return
					}
					filename := strconv.FormatInt(time.Now().UnixNano(), 10) + ext
					imagePath = uploadPath + filename
					f, err := os.OpenFile(imagePath, os.O_WRONLY|os.O_CREATE, 0666)
					if err != nil {
						ErrorHandler(w, r, http.StatusInternalServerError)
						return
					}
					defer f.Close()
					io.Copy(f, file)
					imagePath = "/static/images/posts/" + filename
				} else {
					imagePath = post.ImagePath
				}
				err = functions.UpdatePost(id, title, content, categories, imagePath)
				if err != nil {
					ErrorHandler(w, r, http.StatusInternalServerError)
					return
				}
				http.Redirect(w, r, "/post/"+idStr, http.StatusSeeOther)
			}
		} else if parts[1] == "delete" && r.Method == "POST" {
			if user == nil || (!isOwner && user.ID != post.UserID) {
				http.Error(w, "Non autorisé", http.StatusUnauthorized)
				return
			}
			err = functions.DeletePost(id)
			if err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}
}
