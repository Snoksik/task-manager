package model

type Task struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	UserID int64  `json:"user_id"`
}
