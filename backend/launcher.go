package backend

import (
	"fmt"
	"forum/backend/database"
	"forum/backend/handlers"
	"forum/backend/websocket"
	"log"
	"net/http"
)

func Launcher() {
	database.InitDB()

	http.HandleFunc("/", handlers.IndexHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/profile", handlers.ProfileHandler)
	http.HandleFunc("/profile/", handlers.ProfileHandler)
	http.HandleFunc("/profile/edit", handlers.EditProfileHandler)
	http.HandleFunc("/post/new", handlers.NewPostHandler)
	http.HandleFunc("/post/", handlers.PostHandler)
	http.HandleFunc("/like/", handlers.LikeHandler)
	http.HandleFunc("/filter/", handlers.FilterHandler)
	http.HandleFunc("/comment/", handlers.CommentHandler)
	http.HandleFunc("/help", handlers.HelpHandler)
	http.HandleFunc("/notification", handlers.NotificationHandler)
	http.HandleFunc("/login/google", handlers.GoogleLoginHandler)
	http.HandleFunc("/callback/google", handlers.GoogleCallbackHandler)
	http.HandleFunc("/login/github", handlers.GithubLoginHandler)
	http.HandleFunc("/callback/github", handlers.GithubCallbackHandler)

	http.HandleFunc("/moderation", handlers.ModerationHandler)
	http.HandleFunc("/admin", handlers.AdminHandler)

	http.HandleFunc("/report", handlers.ReportHandler)
	http.HandleFunc("/admin/report", handlers.AdminReportHandler)

	http.HandleFunc("/ws", websocket.HandleConnections)
	go websocket.HandleMessages()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("frontend/static"))))

	port := ":1945"
	fmt.Printf("Serveur démarré sur http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
