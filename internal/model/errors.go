package model

import "errors"

var (
	ErrTaskNotFound    = errors.New("task not found")
	ErrInvalidTaskName = errors.New("name must not be empty")
)
