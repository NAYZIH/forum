package main

import (
	"forum/backend"
	"forum/backend/database"
)

func main() {
	database.InitDB()  // Initialise la base de données
	backend.Launcher() // Lance le serveur
}
