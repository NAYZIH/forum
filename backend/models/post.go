package models

import "time"

type Post struct {
	ID         int
	UserID     int
	Title      string
	Content    string
	CreatedAt  time.Time
	Categories []string
	Likes      int
	Dislikes   int
}
