package auth

import (
	"forum/backend/database"
	"forum/backend/models"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func CreateSession(userID int) (*models.Session, error) {
	sessionID := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err := database.DB.Exec("INSERT INTO sessions (session_id, user_id, expires_at) VALUES (?, ?, ?)",
		sessionID, userID, expiresAt)
	if err != nil {
		return nil, err
	}
	return &models.Session{
		SessionID: sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}, nil
}

func GetSession(sessionID string) (*models.Session, error) {
	row := database.DB.QueryRow("SELECT session_id, user_id, expires_at FROM sessions WHERE session_id = ?", sessionID)
	s := &models.Session{}
	err := row.Scan(&s.SessionID, &s.UserID, &s.ExpiresAt)
	if err != nil {
		return nil, err
	}
	if time.Now().After(s.ExpiresAt) {
		return nil, nil
	}
	return s, nil
}

func DeleteSession(sessionID string) error {
	_, err := database.DB.Exec("DELETE FROM sessions WHERE session_id = ?", sessionID)
	return err
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		session, err := GetSession(cookie.Value)
		if err != nil || session == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}
