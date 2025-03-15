package functions

import "forum/backend/database"

func LikePost(userID, postID int, value int) error {
	_, err := database.DB.Exec(`
		INSERT INTO post_likes (user_id, post_id, value) VALUES (?, ?, ?)
		ON CONFLICT(user_id, post_id) DO UPDATE SET value = ?`,
		userID, postID, value, value)
	return err
}

func LikeComment(userID, commentID int, value int) error {
	_, err := database.DB.Exec(`
		INSERT INTO comment_likes (user_id, comment_id, value) VALUES (?, ?, ?)
		ON CONFLICT(user_id, comment_id) DO UPDATE SET value = ?`,
		userID, commentID, value, value)
	return err
}
