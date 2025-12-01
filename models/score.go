package models

type Score struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Best     int    `json:"best"`
}
