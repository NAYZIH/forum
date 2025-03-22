package functions

import (
	"forum/backend/database"
	"forum/backend/models"

	"golang.org/x/crypto/bcrypt"
)

func CreateUser(email, username, password, bio string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = database.DB.Exec("INSERT INTO users (email, username, password, bio) VALUES (?, ?, ?, ?)",
		email, username, string(hashedPassword), bio)
	return err
}

func GetUserByEmail(email string) (*models.User, error) {
	row := database.DB.QueryRow("SELECT id, email, username, password, bio, avatar_path, created_at FROM users WHERE email = ?", email)
	user := &models.User{}
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.Bio, &user.AvatarPath, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByID(id int) (*models.User, error) {
	row := database.DB.QueryRow("SELECT id, email, username, password, bio, avatar_path, created_at FROM users WHERE id = ?", id)
	user := &models.User{}
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.Bio, &user.AvatarPath, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByEmailOrUsername(identifier string) (*models.User, error) {
	row := database.DB.QueryRow("SELECT id, email, username, password, bio, avatar_path, created_at FROM users WHERE email = ? OR username = ?", identifier, identifier)
	user := &models.User{}
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.Bio, &user.AvatarPath, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func Authenticate(identifier, password string) (*models.User, error) {
	user, err := GetUserByEmailOrUsername(identifier)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, err
	}
	return user, nil
}

func UpdateUser(id int, username, email, bio, avatarPath string) error {
	_, err := database.DB.Exec("UPDATE users SET username = ?, email = ?, bio = ?, avatar_path = ? WHERE id = ?", username, email, bio, avatarPath, id)
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
