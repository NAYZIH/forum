package models

type EditProfileData struct {
	User    *User
	Avatars []string
	Error   string
}
