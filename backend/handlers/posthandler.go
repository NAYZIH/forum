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

const maxUploadSize = 20 * 1024 * 1024
const uploadPath = "./frontend/static/images/posts/"

func NewPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/post/new" {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}

	sessionID, err := r.Cookie("session_id")
	var user *models.User
	var unreadCount int
	if err == nil {
		session, err := auth.GetSession(sessionID.Value)
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

	if r.Method == "GET" {
		t, err := template.ParseFiles("frontend/templates/newpost.html")
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		categories, err := functions.GetAllCategories()
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

		err = r.ParseMultipartForm(maxUploadSize)
		if err != nil {
			http.Error(w, "Fichier trop volumineux", http.StatusBadRequest)
			return
		}
		title := r.FormValue("title")
		content := r.FormValue("content")
		categories := r.Form["categories[]"]
		newCategories := r.FormValue("new_categories")
		if newCategories != "" {
			newCats := strings.Split(newCategories, ",")
			for i, cat := range newCats {
				newCats[i] = strings.TrimSpace(cat)
			}
			categories = append(categories, newCats...)
		}

		file, handler, err := r.FormFile("image")
		var imagePath string
		if err == nil {
			defer file.Close()
			if handler.Size > maxUploadSize {
				t, _ := template.ParseFiles("frontend/templates/newpost.html")
				data := struct {
					User       *models.User
					Categories []string
					Error      string
				}{
					Categories: categories,
					Error:      "L'image est trop volumineuse (max 20 Mo).",
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
				t, _ := template.ParseFiles("frontend/templates/newpost.html")
				data := struct {
					User       *models.User
					Categories []string
					Error      string
				}{
					Categories: categories,
					Error:      "Extension d'image non supportée (JPEG, PNG, GIF uniquement).",
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
			imagePath = "/static/images/posts/" + filepath.Base(imagePath)
		}

		err = functions.CreatePost(session.UserID, title, content, categories, imagePath)
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
			Post        *models.Post
			Comments    []models.Comment
			User        *models.User
			UnreadCount int
		}{
			Post:     post,
			Comments: comments,
		}
		if cookie, err := r.Cookie("session_id"); err == nil {
			if session, err := auth.GetSession(cookie.Value); err == nil && session != nil {
				data.User, _ = functions.GetUserByID(session.UserID)
				data.UnreadCount, _ = functions.GetUnreadNotificationCount(session.UserID)
			}
		}
		t.Execute(w, data)
	} else if len(parts) == 2 && parts[1] == "edit" {
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
		sessionID, err := r.Cookie("session_id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		session, err := auth.GetSession(sessionID.Value)
		if err != nil || session == nil || session.UserID != post.UserID {
			http.Error(w, "Non autorisé", http.StatusUnauthorized)
			return
		}
		if r.Method == "GET" {
			categories, err := functions.GetAllCategories()
			if err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError)
				return
			}
			t, err := template.New("editpost.html").Funcs(template.FuncMap{
				"in": func(slice []string, val string) bool {
					for _, item := range slice {
						if item == val {
							return true
						}
					}
					return false
				},
			}).ParseFiles("frontend/templates/editpost.html")
			if err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError)
				return
			}
			user, _ := functions.GetUserByID(session.UserID)
			unreadCount, _ := functions.GetUnreadNotificationCount(session.UserID)

			data := struct {
				Post        *models.Post
				Categories  []string
				Error       string
				User        *models.User
				UnreadCount int
			}{
				Post:        post,
				Categories:  categories,
				Error:       "",
				User:        user,
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
				categories, err := functions.GetAllCategories()
				if err != nil {
					ErrorHandler(w, r, http.StatusInternalServerError)
					return
				}
				t, err := template.New("editpost.html").Funcs(template.FuncMap{
					"in": func(slice []string, val string) bool {
						for _, item := range slice {
							if item == val {
								return true
							}
						}
						return false
					},
				}).ParseFiles("frontend/templates/editpost.html")
				if err != nil {
					ErrorHandler(w, r, http.StatusInternalServerError)
					return
				}
				user, _ := functions.GetUserByID(session.UserID)
				unreadCount, _ := functions.GetUnreadNotificationCount(session.UserID)

				data := struct {
					Post        *models.Post
					Categories  []string
					Error       string
					User        *models.User
					UnreadCount int
				}{
					Post:        post,
					Categories:  categories,
					Error:       "",
					User:        user,
					UnreadCount: unreadCount,
				}
				t.Execute(w, data)
			}
			file, handler, err := r.FormFile("image")
			var imagePath string
			if err == nil {
				defer file.Close()
				if handler.Size > maxUploadSize {
					categories, err := functions.GetAllCategories()
					if err != nil {
						ErrorHandler(w, r, http.StatusInternalServerError)
						return
					}
					t, err := template.New("editpost.html").Funcs(template.FuncMap{
						"in": func(slice []string, val string) bool {
							for _, item := range slice {
								if item == val {
									return true
								}
							}
							return false
						},
					}).ParseFiles("frontend/templates/editpost.html")
					if err != nil {
						ErrorHandler(w, r, http.StatusInternalServerError)
						return
					}
					user, _ := functions.GetUserByID(session.UserID)
					unreadCount, _ := functions.GetUnreadNotificationCount(session.UserID)

					data := struct {
						Post        *models.Post
						Categories  []string
						Error       string
						User        *models.User
						UnreadCount int
					}{
						Post:        post,
						Categories:  categories,
						Error:       "",
						User:        user,
						UnreadCount: unreadCount,
					}
					t.Execute(w, data)
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
					categories, err := functions.GetAllCategories()
					if err != nil {
						ErrorHandler(w, r, http.StatusInternalServerError)
						return
					}
					t, err := template.New("editpost.html").Funcs(template.FuncMap{
						"in": func(slice []string, val string) bool {
							for _, item := range slice {
								if item == val {
									return true
								}
							}
							return false
						},
					}).ParseFiles("frontend/templates/editpost.html")
					if err != nil {
						ErrorHandler(w, r, http.StatusInternalServerError)
						return
					}
					user, _ := functions.GetUserByID(session.UserID)
					unreadCount, _ := functions.GetUnreadNotificationCount(session.UserID)

					data := struct {
						Post        *models.Post
						Categories  []string
						Error       string
						User        *models.User
						UnreadCount int
					}{
						Post:        post,
						Categories:  categories,
						Error:       "",
						User:        user,
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
				imagePath = "/static/images/posts/" + filepath.Base(imagePath)
			} else {
				imagePath = post.ImagePath
			}
			err = functions.UpdatePost(post.ID, title, content, allCategories, imagePath)
			if err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/post/"+strconv.Itoa(post.ID), http.StatusSeeOther)
		} else {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		}
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
		err = r.ParseMultipartForm(maxUploadSize)
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
				http.Error(w, "L'image est trop volumineuse (max 20 Mo).", http.StatusBadRequest)
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
				http.Error(w, "Extension d'image non supportée (JPEG, PNG, GIF uniquement).", http.StatusBadRequest)
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
			imagePath = "/static/images/posts/" + filepath.Base(imagePath)
		}
		err = functions.CreateComment(id, session.UserID, content, imagePath)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/post/"+parts[0], http.StatusSeeOther)
	} else if len(parts) == 2 && parts[1] == "delete" && r.Method == "POST" {
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
		post, err := functions.GetPostByID(id)
		if err != nil {
			ErrorHandler(w, r, http.StatusNotFound)
			return
		}
		if session.UserID != post.UserID {
			http.Error(w, "Non autorisé", http.StatusUnauthorized)
			return
		}
		err = functions.DeletePost(id)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		ErrorHandler(w, r, http.StatusNotFound)
	}
}
