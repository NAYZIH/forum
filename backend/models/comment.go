package models

import "time"

type Comment struct {
	ID         int
	PostID     int
	UserID     int
	Username   string
	AvatarPath string
	Content    string
	ImagePath  string
	CreatedAt  time.Time
	Likes      int
	Dislikes   int
	PostTitle  string
}
