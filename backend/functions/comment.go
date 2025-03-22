package functions

import (
	"forum/backend/database"
	"forum/backend/models"
)

func CreateComment(postID, userID int, content, imagePath string) error {
	_, err := database.DB.Exec("INSERT INTO comments (post_id, user_id, content, image_path) VALUES (?, ?, ?, ?)", postID, userID, content, imagePath)
	return err
}

func GetCommentsByPostID(postID int) ([]models.Comment, error) {
	rows, err := database.DB.Query(`
        SELECT c.id, c.post_id, c.user_id, u.username, c.content, c.image_path, c.created_at
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
		err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Username, &c.Content, &c.ImagePath, &c.CreatedAt)
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
        SELECT c.id, c.post_id, c.user_id, u.username, c.content, c.image_path, c.created_at, p.title
        FROM comments c
        JOIN users u ON c.user_id = u.id
        JOIN posts p ON c.post_id = p.id
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
		var postTitle string
		err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Username, &c.Content, &c.ImagePath, &c.CreatedAt, &postTitle)
		if err != nil {
			return nil, err
		}
		c.PostTitle = postTitle
		comments = append(comments, c)
	}
	return comments, nil
}

func GetLikedCommentsByUserID(userID int) ([]models.Comment, error) {
	rows, err := database.DB.Query(`
        SELECT c.id, c.post_id, c.user_id, u.username, c.content, c.image_path, c.created_at, p.title
        FROM comments c
        JOIN users u ON c.user_id = u.id
        JOIN posts p ON c.post_id = p.id
        JOIN comment_likes cl ON c.id = cl.comment_id
        WHERE cl.user_id = ? AND cl.value = 1
        ORDER BY cl.created_at DESC
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []models.Comment
	for rows.Next() {
		var c models.Comment
		err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Username, &c.Content, &c.ImagePath, &c.CreatedAt, &c.PostTitle)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func GetDislikedCommentsByUserID(userID int) ([]models.Comment, error) {
	rows, err := database.DB.Query(`
        SELECT c.id, c.post_id, c.user_id, u.username, c.content, c.image_path, c.created_at, p.title
        FROM comments c
        JOIN users u ON c.user_id = u.id
        JOIN posts p ON c.post_id = p.id
        JOIN comment_likes cl ON c.id = cl.comment_id
        WHERE cl.user_id = ? AND cl.value = -1
        ORDER BY cl.created_at DESC
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []models.Comment
	for rows.Next() {
		var c models.Comment
		err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Username, &c.Content, &c.ImagePath, &c.CreatedAt, &c.PostTitle)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}
