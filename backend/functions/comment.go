package functions

import (
	"forum/backend/database"
	"time"
)

type Comment struct {
	ID        int
	PostID    int
	UserID    int
	Content   string
	CreatedAt time.Time
	Likes     int
	Dislikes  int
}

func CreateComment(postID, userID int, content string) error {
	_, err := database.DB.Exec("INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)", postID, userID, content)
	return err
}

func GetCommentsByPostID(postID int) ([]Comment, error) {
	rows, err := database.DB.Query("SELECT id, post_id, user_id, content, created_at FROM comments WHERE post_id = ? ORDER BY created_at ASC", postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []Comment
	for rows.Next() {
		var c Comment
		err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Content, &c.CreatedAt)
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
