package models

type Config struct {
	GoogleClientID     string `json:"google_client_id"`
	GoogleClientSecret string `json:"google_client_secret"`
	GithubClientID     string `json:"github_client_id"`
	GithubClientSecret string `json:"github_client_secret"`
}
