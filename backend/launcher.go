package backend

import (
	"fmt"
	"forum/backend/handlers"
	"log"
	"net/http"
)

func Launcher() {
	http.HandleFunc("/", handlers.IndexHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/post/new", handlers.NewPostHandler)
	http.HandleFunc("/post/", handlers.PostHandler)     // /post/{id} et /post/{id}/comment
	http.HandleFunc("/like/", handlers.LikeHandler)     // /like/post/{id} et /like/comment/{id}
	http.HandleFunc("/filter/", handlers.FilterHandler) // /filter/...

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("frontend/static"))))

	port := ":1945"
	fmt.Printf("Serveur démarré sur http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
