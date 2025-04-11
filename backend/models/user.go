package models

import "time"

type User struct {
	ID         int       `json:"id"`
	Email      string    `json:"email"`
	Username   string    `json:"username"`
	Password   string    `json:"password,omitempty"`
	Bio        string    `json:"bio"`
	AvatarPath string    `json:"avatar_path"`
	Role       string    `json:"role"`
	CreatedAt  time.Time `json:"created_at"`
}
