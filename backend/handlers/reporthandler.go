package handlers

import (
	"forum/backend/auth"
	"forum/backend/functions"
	"forum/backend/models"
	"html/template"
	"net/http"
	"strconv"
)

func ReportHandler(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Utilisateur introuvable", http.StatusInternalServerError)
		return
	}
	if user.Role != "modérateur" && user.Role != "administrateur" && user.Role != "owner" {
		http.Error(w, "Accès refusé", http.StatusForbidden)
		return
	}

	if r.Method == "POST" {
		postIDStr := r.FormValue("post_id")
		commentIDStr := r.FormValue("comment_id")
		reason := r.FormValue("reason")

		var postID, commentID *int
		if postIDStr != "" {
			pID, err := strconv.Atoi(postIDStr)
			if err != nil {
				http.Error(w, "ID invalide", http.StatusBadRequest)
				return
			}
			postID = &pID
		}
		if commentIDStr != "" {
			cID, err := strconv.Atoi(commentIDStr)
			if err != nil {
				http.Error(w, "ID invalide", http.StatusBadRequest)
				return
			}
			commentID = &cID
		}

		err = functions.CreateReport(user.ID, postID, commentID, reason)
		if err != nil {
			http.Error(w, "Erreur lors du signalement", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

func AdminReportHandler(w http.ResponseWriter, r *http.Request) {
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
	if adminUser.Role != "administrateur" && adminUser.Role != "owner" {
		http.Error(w, "Accès refusé", http.StatusForbidden)
		return
	}

	if r.Method == "GET" {
		reports, err := functions.GetPendingReports()
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des signalements", http.StatusInternalServerError)
			return
		}
		t, err := template.ParseFiles("frontend/templates/adminreport.html")
		if err != nil {
			http.Error(w, "Erreur de template", http.StatusInternalServerError)
			return
		}
		data := struct {
			Admin   *models.User
			Reports []models.Report
		}{
			Admin:   adminUser,
			Reports: reports,
		}
		t.Execute(w, data)
	} else if r.Method == "POST" {
		reportIDStr := r.FormValue("report_id")
		action := r.FormValue("action")
		reportID, err := strconv.Atoi(reportIDStr)
		if err != nil {
			http.Error(w, "ID de signalement invalide", http.StatusBadRequest)
			return
		}
		if action == "review" {
			err = functions.UpdateReportStatus(reportID, "reviewed")
			if err != nil {
				http.Error(w, "Erreur lors de la mise à jour du signalement", http.StatusInternalServerError)
				return
			}
		} else if action == "delete_post" {
			postIDStr := r.FormValue("post_id")
			postID, err := strconv.Atoi(postIDStr)
			if err != nil {
				http.Error(w, "ID de post invalide", http.StatusBadRequest)
				return
			}
			err = functions.DeletePost(postID)
			if err != nil {
				http.Error(w, "Erreur lors de la suppression du post", http.StatusInternalServerError)
				return
			}
			err = functions.UpdateReportStatus(reportID, "deleted")
			if err != nil {
				http.Error(w, "Erreur lors de la mise à jour du signalement", http.StatusInternalServerError)
				return
			}
		} else if action == "delete_comment" {
			commentIDStr := r.FormValue("comment_id")
			commentID, err := strconv.Atoi(commentIDStr)
			if err != nil {
				http.Error(w, "ID de commentaire invalide", http.StatusBadRequest)
				return
			}
			err = functions.DeleteComment(commentID)
			if err != nil {
				http.Error(w, "Erreur lors de la suppression du commentaire", http.StatusInternalServerError)
				return
			}
			err = functions.UpdateReportStatus(reportID, "deleted")
			if err != nil {
				http.Error(w, "Erreur lors de la mise à jour du signalement", http.StatusInternalServerError)
				return
			}
		}
		http.Redirect(w, r, "/admin/report", http.StatusSeeOther)
	} else {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}
