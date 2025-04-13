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
	row := database.DB.QueryRow("SELECT id, email, username, password, bio, avatar_path, role, created_at FROM users WHERE email = ?", email)
	user := &models.User{}
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.Bio, &user.AvatarPath, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByID(id int) (*models.User, error) {
	row := database.DB.QueryRow("SELECT id, email, username, password, bio, avatar_path, role, created_at FROM users WHERE id = ?", id)
	user := &models.User{}
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.Bio, &user.AvatarPath, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByEmailOrUsername(identifier string) (*models.User, error) {
	row := database.DB.QueryRow("SELECT id, email, username, password, bio, avatar_path, role, created_at FROM users WHERE email = ? OR username = ?", identifier, identifier)
	user := &models.User{}
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.Bio, &user.AvatarPath, &user.Role, &user.CreatedAt)
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

func GetAllUsers() ([]models.User, error) {
	rows, err := database.DB.Query("SELECT id, email, username, bio, avatar_path, role, created_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.Bio, &u.AvatarPath, &u.Role, &u.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func UpdateUserRole(userID int, role string) error {
	_, err := database.DB.Exec("UPDATE users SET role = ? WHERE id = ?", role, userID)
	return err
}

func DeleteUser(userID int) error {
	_, err := database.DB.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		return err
	}
	_, err = database.DB.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	return err
}

func DeleteUserSessions(userID int) error {
	_, err := database.DB.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	return err
}
