package backend

import (
	"fmt"
	"forum/backend/handlers"
	"net/http"
)

func Launcher() {
	http.HandleFunc("/", handlers.IndexHandler)

	port := ":1945"
	fmt.Printf("http://localhost%s\n", port)

	http.ListenAndServe(port, nil)
}
