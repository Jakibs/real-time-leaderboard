package models

type User struct {
	UserId       int    `json:"userId"`
	Username     string `json:"username"`
	PasswordHash string `json:"password"`
}
