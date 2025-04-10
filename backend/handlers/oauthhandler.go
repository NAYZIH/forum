package handlers

import (
	"context"
	"encoding/json"
	"forum/backend/auth"
	"forum/backend/functions"
	"forum/backend/models"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Début de GoogleLoginHandler")
	state := auth.GenerateRandomState()
	log.Printf("État OAuth généré : %s", state)
	cookie := http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(5 * time.Minute),
	}
	http.SetCookie(w, &cookie)
	log.Println("Cookie oauth_state défini")
	url := auth.GoogleOauthConfig.AuthCodeURL(state)
	log.Printf("Redirection vers Google : %s", url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Début de GoogleCallbackHandler")
	cookie, err := r.Cookie("oauth_state")
	if err != nil {
		log.Printf("Erreur lors de la récupération du cookie oauth_state : %v", err)
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}
	if r.FormValue("state") != cookie.Value {
		log.Printf("État OAuth invalide : attendu=%s, reçu=%s", cookie.Value, r.FormValue("state"))
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}
	log.Println("État OAuth valide")

	code := r.FormValue("code")
	log.Printf("Code d'autorisation reçu : %s", code)
	token, err := auth.GoogleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Erreur lors de l'échange du token : %v", err)
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}
	log.Println("Token échangé avec succès")

	client := auth.GoogleOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Printf("Erreur lors de la récupération des infos utilisateur : %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	log.Println("Infos utilisateur récupérées")

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.Printf("Erreur lors du décodage des infos utilisateur : %v", err)
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}
	log.Printf("Infos utilisateur décodées : %v", userInfo)

	email, ok := userInfo["email"].(string)
	if !ok {
		log.Println("Erreur : Email non trouvé dans les infos utilisateur")
		http.Error(w, "Email not found", http.StatusInternalServerError)
		return
	}
	username, ok := userInfo["name"].(string)
	if !ok {
		log.Println("Avertissement : Nom non trouvé, utilisation de GoogleUser")
		username = "GoogleUser"
	}
	log.Printf("Utilisateur OAuth : email=%s, username=%s", email, username)

	if err := handleOAuthUser(w, r, email, username); err != nil {
		log.Printf("Erreur lors de la gestion de l'utilisateur OAuth : %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("Utilisateur OAuth géré avec succès")
}

func GithubLoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Début de GithubLoginHandler")
	state := auth.GenerateRandomState()
	log.Printf("État OAuth généré : %s", state)
	cookie := http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(5 * time.Minute),
	}
	http.SetCookie(w, &cookie)
	log.Println("Cookie oauth_state défini")
	url := auth.GithubOauthConfig.AuthCodeURL(state)
	log.Printf("Redirection vers GitHub : %s", url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GithubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Début de GithubCallbackHandler")
	cookie, err := r.Cookie("oauth_state")
	if err != nil {
		log.Printf("Erreur lors de la récupération du cookie oauth_state : %v", err)
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}
	if r.FormValue("state") != cookie.Value {
		log.Printf("État OAuth invalide : attendu=%s, reçu=%s", cookie.Value, r.FormValue("state"))
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}
	log.Println("État OAuth valide")

	code := r.FormValue("code")
	log.Printf("Code d'autorisation reçu : %s", code)
	token, err := auth.GithubOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Erreur lors de l'échange du token : %v", err)
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}
	log.Println("Token échangé avec succès")

	client := auth.GithubOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		log.Printf("Erreur lors de la récupération des infos utilisateur : %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	log.Println("Infos utilisateur récupérées")

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.Printf("Erreur lors du décodage des infos utilisateur : %v", err)
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}
	log.Printf("Infos utilisateur décodées : %v", userInfo)

	email, ok := userInfo["email"].(string)
	if !ok {
		log.Println("Email non trouvé dans les infos utilisateur, récupération des emails")
		resp, err := client.Get("https://api.github.com/user/emails")
		if err != nil {
			log.Printf("Erreur lors de la récupération des emails : %v", err)
			http.Error(w, "Failed to fetch user emails", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		var emails []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
			log.Printf("Erreur lors du décodage des emails : %v", err)
			http.Error(w, "Failed to decode emails", http.StatusInternalServerError)
			return
		}
		for _, e := range emails {
			if primary, _ := e["primary"].(bool); primary {
				if verified, _ := e["verified"].(bool); verified {
					email = e["email"].(string)
					break
				}
			}
		}
		if email == "" {
			log.Println("Erreur : Aucun email primaire vérifié trouvé")
			http.Error(w, "No verified primary email found", http.StatusBadRequest)
			return
		}
	}

	username, ok := userInfo["login"].(string)
	if !ok {
		log.Println("Avertissement : Login non trouvé, utilisation de GithubUser")
		username = "GithubUser"
	}
	log.Printf("Utilisateur OAuth : email=%s, username=%s", email, username)

	if err := handleOAuthUser(w, r, email, username); err != nil {
		log.Printf("Erreur lors de la gestion de l'utilisateur OAuth : %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("Utilisateur OAuth géré avec succès")
}

func handleOAuthUser(w http.ResponseWriter, r *http.Request, email, username string) error {
	log.Printf("Début de handleOAuthUser : email=%s, username=%s", email, username)
	user, err := functions.GetUserByEmail(email)
	if err != nil && err.Error() != "sql: no rows in result set" {
		log.Printf("Erreur lors de la recherche de l'utilisateur par email : %v", err)
		return err
	}

	if user != nil {
		log.Printf("Utilisateur existant trouvé : ID=%d", user.ID)
		session, err := auth.CreateSession(user.ID)
		if err != nil {
			log.Printf("Erreur lors de la création de la session : %v", err)
			return err
		}
		log.Printf("Session créée : SessionID=%s", session.SessionID)
		setSessionCookie(w, session)
		log.Println("Cookie de session défini")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		log.Println("Redirection vers /")
		return nil
	}

	log.Println("Utilisateur non trouvé, création d'un nouvel utilisateur")
	password := auth.GenerateRandomPassword()
	log.Printf("Mot de passe généré : %s", password)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Erreur lors du hachage du mot de passe : %v", err)
		return err
	}

	newUser := &models.User{
		Email:    email,
		Username: username,
		Password: string(hashedPassword),
	}
	if err := functions.CreateUser(newUser.Email, newUser.Username, newUser.Password, ""); err != nil {
		log.Printf("Erreur lors de la création de l'utilisateur : %v", err)
		return err
	}
	log.Println("Nouvel utilisateur créé")

	user, err = functions.GetUserByEmail(email)
	if err != nil {
		log.Printf("Erreur lors de la récupération du nouvel utilisateur : %v", err)
		return err
	}
	log.Printf("Nouvel utilisateur récupéré : ID=%d", user.ID)

	session, err := auth.CreateSession(user.ID)
	if err != nil {
		log.Printf("Erreur lors de la création de la session pour le nouvel utilisateur : %v", err)
		return err
	}
	log.Printf("Session créée : SessionID=%s", session.SessionID)
	setSessionCookie(w, session)
	log.Println("Cookie de session défini")
	http.Redirect(w, r, "/", http.StatusSeeOther)
	log.Println("Redirection vers /")
	return nil
}

func setSessionCookie(w http.ResponseWriter, session *models.Session) {
	log.Printf("Définition du cookie : session_id=%s, Expires=%v", session.SessionID, session.ExpiresAt)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.SessionID,
		Path:     "/",
		HttpOnly: true,
		Expires:  session.ExpiresAt,
	})
}
