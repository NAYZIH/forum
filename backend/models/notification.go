package models

import "time"

type Notification struct {
	ID         int
	UserID     int
	Type       string
	PostID     *int
	CommentID  *int
	FromUserID int
	FromUser   string
	PostTitle  string
	CreatedAt  time.Time
	IsRead     int
}
