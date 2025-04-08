package functions

import (
	"database/sql"
	"forum/backend/database"
)

func GetPostLikeValue(userID, postID int) (int, error) {
	var value int
	err := database.DB.QueryRow("SELECT value FROM post_likes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&value)
	if err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	return value, nil
}

func GetCommentLikeValue(userID, commentID int) (int, error) {
	var value int
	err := database.DB.QueryRow("SELECT value FROM comment_likes WHERE user_id = ? AND comment_id = ?", userID, commentID).Scan(&value)
	if err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	return value, nil
}

func LikePost(userID, postID int, action string) error {
	currentValue, err := GetPostLikeValue(userID, postID)
	if err != nil {
		return err
	}

	var postOwnerID int
	err = database.DB.QueryRow("SELECT user_id FROM posts WHERE id = ?", postID).Scan(&postOwnerID)
	if err != nil {
		return err
	}

	if userID != postOwnerID {
		var notificationType string
		if action == "like" {
			notificationType = "like_post"
		} else if action == "dislike" {
			notificationType = "dislike_post"
		}
		if currentValue == 0 {
			err = createNotification(postOwnerID, userID, notificationType, &postID, nil)
			if err != nil {
				return err
			}
		}
	}

	if action == "like" {
		if currentValue == 1 {
			_, err = database.DB.Exec("DELETE FROM post_likes WHERE user_id = ? AND post_id = ?", userID, postID)
		} else {
			_, err = database.DB.Exec(`
				INSERT INTO post_likes (user_id, post_id, value) VALUES (?, ?, 1)
				ON CONFLICT(user_id, post_id) DO UPDATE SET value = 1`, userID, postID)
		}
	} else if action == "dislike" {
		if currentValue == -1 {
			_, err = database.DB.Exec("DELETE FROM post_likes WHERE user_id = ? AND post_id = ?", userID, postID)
		} else {
			_, err = database.DB.Exec(`
				INSERT INTO post_likes (user_id, post_id, value) VALUES (?, ?, -1)
				ON CONFLICT(user_id, post_id) DO UPDATE SET value = -1`, userID, postID)
		}
	}
	return err
}

func LikeComment(userID, commentID int, action string) error {
	currentValue, err := GetCommentLikeValue(userID, commentID)
	if err != nil {
		return err
	}

	var commentOwnerID int
	err = database.DB.QueryRow("SELECT user_id FROM comments WHERE id = ?", commentID).Scan(&commentOwnerID)
	if err != nil {
		return err
	}

	if userID != commentOwnerID {
		var notificationType string
		if action == "like" {
			notificationType = "like_comment"
		} else if action == "dislike" {
			notificationType = "dislike_comment"
		}
		if currentValue == 0 {
			err = createNotification(commentOwnerID, userID, notificationType, nil, &commentID)
			if err != nil {
				return err
			}
		}
	}

	if action == "like" {
		if currentValue == 1 {
			_, err = database.DB.Exec("DELETE FROM comment_likes WHERE user_id = ? AND comment_id = ?", userID, commentID)
		} else {
			_, err = database.DB.Exec(`
				INSERT INTO comment_likes (user_id, comment_id, value) VALUES (?, ?, 1)
				ON CONFLICT(user_id, comment_id) DO UPDATE SET value = 1`, userID, commentID)
		}
	} else if action == "dislike" {
		if currentValue == -1 {
			_, err = database.DB.Exec("DELETE FROM comment_likes WHERE user_id = ? AND comment_id = ?", userID, commentID)
		} else {
			_, err = database.DB.Exec(`
				INSERT INTO comment_likes (user_id, comment_id, value) VALUES (?, ?, -1)
				ON CONFLICT(user_id, comment_id) DO UPDATE SET value = -1`, userID, commentID)
		}
	}
	return err
}
