package model

import "time"

// MessageRequest represents a message sending request
// @Description Message sending request body
type MessageRequest struct {
	Content string `json:"content" binding:"required"`
}

// Message represents a message in the system
// @Description Message information
type Message struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
}
