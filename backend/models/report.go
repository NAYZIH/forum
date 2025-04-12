package models

import "time"

type Report struct {
	ID          int       `json:"id"`
	ModeratorID int       `json:"moderator_id"`
	PostID      *int      `json:"post_id"`
	CommentID   *int      `json:"comment_id"`
	Reason      string    `json:"reason"`
	CreatedAt   time.Time `json:"created_at"`
	Status      string    `json:"status"`
}
