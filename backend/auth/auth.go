package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"forum/backend/database"
	"forum/backend/models"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

var GoogleOauthConfig *oauth2.Config
var GithubOauthConfig *oauth2.Config

func init() {
	log.Println("Chargement de la configuration depuis config.json")
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Erreur lors de l'ouverture de config.json : %v", err)
	}
	defer file.Close()

	var config models.Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Fatalf("Erreur lors du décodage de config.json : %v", err)
	}

	if config.GoogleClientID == "" || config.GoogleClientSecret == "" || config.GithubClientID == "" || config.GithubClientSecret == "" {
		log.Fatal("Erreur : un ou plusieurs identifiants manquent dans config.json")
	}

	GoogleOauthConfig = &oauth2.Config{
		ClientID:     config.GoogleClientID,
		ClientSecret: config.GoogleClientSecret,
		RedirectURL:  "http://localhost:3945/callback/google",
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	GithubOauthConfig = &oauth2.Config{
		ClientID:     config.GithubClientID,
		ClientSecret: config.GithubClientSecret,
		RedirectURL:  "http://localhost:3945/callback/github",
		Scopes:       []string{"user"},
		Endpoint:     github.Endpoint,
	}

	log.Println("Configuration OAuth initialisée avec succès")
}

func GenerateRandomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func GenerateRandomPassword() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return base64.URLEncoding.EncodeToString([]byte("defaultpassword123"))
	}
	return base64.URLEncoding.EncodeToString(b)[:16]
}

func CreateSession(userID int) (*models.Session, error) {
	log.Printf("Création de session pour userID=%d", userID)
	sessionID := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	session := &models.Session{
		SessionID: sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	_, err := database.DB.Exec(
		"INSERT INTO sessions (session_id, user_id, expires_at) VALUES (?, ?, ?)",
		session.SessionID, session.UserID, session.ExpiresAt,
	)
	if err != nil {
		log.Printf("Erreur lors de l'insertion de la session dans la base de données : %v", err)
		return nil, err
	}

	log.Println("Session insérée dans la base de données")
	return session, nil
}

func GetSession(sessionID string) (*models.Session, error) {
	log.Printf("Récupération de la session : session_id=%s", sessionID)
	var session models.Session
	err := database.DB.QueryRow(
		"SELECT session_id, user_id, expires_at FROM sessions WHERE session_id = ? AND expires_at > ?",
		sessionID, time.Now(),
	).Scan(&session.SessionID, &session.UserID, &session.ExpiresAt)
	if err != nil {
		log.Printf("Erreur lors de la récupération de la session : %v", err)
		return nil, err
	}
	log.Println("Session récupérée avec succès")
	return &session, nil
}

func DeleteSession(sessionID string) error {
	log.Printf("Suppression de la session : session_id=%s", sessionID)
	_, err := database.DB.Exec("DELETE FROM sessions WHERE session_id = ?", sessionID)
	if err != nil {
		log.Printf("Erreur lors de la suppression de la session : %v", err)
		return err
	}
	log.Println("Session supprimée avec succès")
	return nil
}
