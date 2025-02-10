package handler

import (
	"encoding/json"
	"net/http"

	repository_event "github.com/Napat/golang-testcontainers-demo/internal/repository/repository_event"
	"github.com/Napat/golang-testcontainers-demo/pkg/errors"
)

type MessageHandler struct {
	producer *repository_event.ProducerRepository
}

// MessageRequest represents the message request body
type MessageRequest struct {
	Content string `json:"content"`
}

// @Summary Send a message
// @Description Send a message to Kafka
// @Tags messages
// @Accept json
// @Produce json
// @Param message body MessageRequest true "Message content"
// @Success 202 "Accepted"
// @Failure 400 {object} errors.Error "Bad Request"
// @Failure 500 {object} errors.Error "Internal Server Error"
// @Router /messages [post]
func NewMessageHandler(producer *repository_event.ProducerRepository) *MessageHandler {
	return &MessageHandler{producer: producer}
}

func (h *MessageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var message MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		json.NewEncoder(w).Encode(errors.NewBadRequest("createMessage", "Invalid request body"))
		return
	}

	if err := h.producer.SendMessage("message", message); err != nil {
		json.NewEncoder(w).Encode(errors.NewInternalError("createMessage", err))
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
