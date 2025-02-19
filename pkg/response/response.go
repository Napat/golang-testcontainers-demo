package response

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error  string `json:"error"`
	Source string `json:"source"`
}

// RespondWithError sends an error response with the specified status code and message
func RespondWithError(w http.ResponseWriter, code int, message string, source string) {
	response := ErrorResponse{
		Error:  message,
		Source: source,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// RespondWithJSON sends a JSON response with the specified status code and payload
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
