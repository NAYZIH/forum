package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		log.Fatal("Erreur lors de l'ouverture de la base de données :", err)
	}

	schema, err := os.ReadFile("backend/database/schema.sql")
	if err != nil {
		log.Fatal("Erreur lors de la lecture de schema.sql :", err)
	}
	_, err = DB.Exec(string(schema))
	if err != nil {
		log.Fatal("Erreur lors de l'exécution du schéma SQL :", err)
	}
}
