package functions

import (
	"forum/backend/database"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int
	Email     string
	Username  string
	Password  string // Haché
	CreatedAt string
}

func CreateUser(email, username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = database.DB.Exec("INSERT INTO users (email, username, password) VALUES (?, ?, ?)",
		email, username, string(hashedPassword))
	return err
}

func GetUserByEmail(email string) (*User, error) {
	row := database.DB.QueryRow("SELECT id, email, username, password, created_at FROM users WHERE email = ?", email)
	user := &User{}
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByID(id int) (*User, error) {
	row := database.DB.QueryRow("SELECT id, email, username, password, created_at FROM users WHERE id = ?", id)
	user := &User{}
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func Authenticate(email, password string) (*User, error) {
	user, err := GetUserByEmail(email)
	if err != nil {
		return nil, err // Ne pas préciser "email non trouvé" pour la sécurité
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, err // "Mot de passe incorrect" masqué sous une erreur générique
	}
	return user, nil
}
