package models

import "time"

type Post struct {
	ID         int
	UserID     int
	Username   string
	Title      string
	Content    string
	ImagePath  string
	CreatedAt  time.Time
	Categories []string
	Likes      int
	Dislikes   int
}
