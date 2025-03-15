package backend

import (
	"fmt"
	"forum/backend/database"
	"forum/backend/handlers"
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
	http.HandleFunc("/profile/edit", handlers.EditProfileHandler)
	http.HandleFunc("/post/new", handlers.NewPostHandler)
	http.HandleFunc("/post/", handlers.PostHandler)
	http.HandleFunc("/like/", handlers.LikeHandler)
	http.HandleFunc("/filter/", handlers.FilterHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("frontend/static"))))

	port := ":1945"
	fmt.Printf("Serveur démarré sur http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
