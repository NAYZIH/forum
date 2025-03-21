package models

import "time"

type Session struct {
	SessionID string
	UserID    int
	ExpiresAt time.Time
}
