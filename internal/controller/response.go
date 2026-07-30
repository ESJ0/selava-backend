package controller

import (
	"encoding/json"
	"log"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("error escribiendo respuesta JSON: %v", err)
	}
}

func respondError(w http.ResponseWriter, status int, mensaje string) {
	respondJSON(w, status, errorResponse{Error: mensaje})
}
