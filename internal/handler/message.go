package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/Napat/golang-testcontainers-demo/pkg/model"
	"github.com/Napat/golang-testcontainers-demo/pkg/response"
	"github.com/Napat/golang-testcontainers-demo/pkg/routes"
)

type MessageProducer interface {
	SendMessage(key string, value interface{}) error
}

type MessageHandler struct {
	producer MessageProducer
	routes   []routes.Route
}

func NewMessageHandler(producer MessageProducer) *MessageHandler {
	h := &MessageHandler{
		producer: producer,
	}

	// Prepare routes
	h.routes = []routes.Route{
		{
			Method:  http.MethodPost,
			Pattern: "/messages",
			Handler: h.sendMessage,
		},
	}

	return h
}

// GetRoutes returns all routes for this handler
func (h *MessageHandler) GetRoutes() []routes.Route {
	return h.routes
}

// @Summary Send a message
// @Description Send a message to the message broker
// @Tags messages
// @Accept json
// @Produce json
// @Param message body map[string]string true "Message content"
// @Success 202 {object} map[string]string
// @Router /api/v1/messages [post]
func (h *MessageHandler) sendMessage(w http.ResponseWriter, r *http.Request) {
	var req model.MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "Invalid request payload", "sendMessage")
		return
	}

	if err := h.producer.SendMessage("message", req); err != nil {
		log.Printf("Error sending message: %v", err)
		response.RespondWithError(w, http.StatusInternalServerError, "Failed to send message", "sendMessage")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Message sent successfully",
	})
}

func (h *MessageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1")

	if path == "/messages" {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.sendMessage(w, r)
		return
	}

	http.NotFound(w, r)
}
