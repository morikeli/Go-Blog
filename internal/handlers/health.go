package handlers

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) CheckServerHealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]string{"status": "OK!", "message": "App ran successfully!"}
	json.NewEncoder(w).Encode(response)

}
