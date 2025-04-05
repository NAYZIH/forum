package functions

import (
	"database/sql"
	"forum/backend/database"
	"forum/backend/models"
)

func GetNotificationsByUserID(userID int) ([]models.Notification, error) {
	rows, err := database.DB.Query(`
		SELECT n.id, n.user_id, n.type, n.post_id, n.comment_id, n.from_user_id, u.username, p.title, n.created_at, n.is_read
		FROM notifications n
		JOIN users u ON n.from_user_id = u.id
		LEFT JOIN posts p ON n.post_id = p.id
		WHERE n.user_id = ?
		ORDER BY n.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var n models.Notification
		var postID, commentID sql.NullInt64
		var postTitle sql.NullString
		err := rows.Scan(&n.ID, &n.UserID, &n.Type, &postID, &commentID, &n.FromUserID, &n.FromUser, &postTitle, &n.CreatedAt, &n.IsRead)
		if err != nil {
			return nil, err
		}
		if postID.Valid {
			pID := int(postID.Int64)
			n.PostID = &pID
		}
		if commentID.Valid {
			cID := int(commentID.Int64)
			n.CommentID = &cID
		}
		if postTitle.Valid {
			n.PostTitle = postTitle.String
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}

func GetUnreadNotificationCount(userID int) (int, error) {
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = 0", userID).Scan(&count)
	return count, err
}

func MarkNotificationsAsRead(userID int) error {
	_, err := database.DB.Exec("UPDATE notifications SET is_read = 1 WHERE user_id = ?", userID)
	return err
}
