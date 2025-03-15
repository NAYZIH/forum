package handlers

import (
	"html/template"
	"log"
	"net/http"
)

func ErrorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	t, err := template.ParseFiles("frontend/templates/error.html")
	if err != nil {
		log.Println("Erreur de template :", err)
		http.Error(w, "Erreur interne", http.StatusInternalServerError)
		return
	}
	data := struct {
		Status int
	}{
		Status: status,
	}
	t.Execute(w, data)
}
