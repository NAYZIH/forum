package functions

import (
	"forum/backend/database"
	"forum/backend/models"

	"golang.org/x/crypto/bcrypt"
)

func CreateUser(email, username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = database.DB.Exec("INSERT INTO users (email, username, password) VALUES (?, ?, ?)",
		email, username, string(hashedPassword))
	return err
}

func GetUserByEmail(email string) (*models.User, error) {
	row := database.DB.QueryRow("SELECT id, email, username, password, created_at FROM users WHERE email = ?", email)
	user := &models.User{}
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByID(id int) (*models.User, error) {
	row := database.DB.QueryRow("SELECT id, email, username, password, created_at FROM users WHERE id = ?", id)
	user := &models.User{}
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func Authenticate(email, password string) (*models.User, error) {
	user, err := GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, err
	}
	return user, nil
}

func UpdateUser(id int, username, email string) error {
	_, err := database.DB.Exec("UPDATE users SET username = ?, email = ? WHERE id = ?", username, email, id)
	return err
}

func EmailExists(email string, excludeUserID int) (bool, error) {
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? AND id != ?", email, excludeUserID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
