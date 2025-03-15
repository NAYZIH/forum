package functions

import (
	"forum/backend/database"
	"forum/backend/models"
)

func CreateComment(postID, userID int, content string) error {
	_, err := database.DB.Exec("INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)", postID, userID, content)
	return err
}

func GetCommentsByPostID(postID int) ([]models.Comment, error) {
	rows, err := database.DB.Query(`
        SELECT c.id, c.post_id, c.user_id, u.username, c.content, c.created_at
        FROM comments c
        JOIN users u ON c.user_id = u.id
        WHERE c.post_id = ?
        ORDER BY c.created_at ASC
    `, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []models.Comment
	for rows.Next() {
		var c models.Comment
		err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Username, &c.Content, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		err = database.DB.QueryRow("SELECT COUNT(*) FROM comment_likes WHERE comment_id = ? AND value = 1", c.ID).Scan(&c.Likes)
		if err != nil {
			return nil, err
		}
		err = database.DB.QueryRow("SELECT COUNT(*) FROM comment_likes WHERE comment_id = ? AND value = -1", c.ID).Scan(&c.Dislikes)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func GetCommentsByUserID(userID int) ([]models.Comment, error) {
	rows, err := database.DB.Query(`
        SELECT c.id, c.post_id, c.user_id, u.username, c.content, c.created_at
        FROM comments c
        JOIN users u ON c.user_id = u.id
        WHERE c.user_id = ?
        ORDER BY c.created_at DESC
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []models.Comment
	for rows.Next() {
		var c models.Comment
		err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Username, &c.Content, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}
