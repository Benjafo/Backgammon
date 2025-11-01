package util

import (
	"encoding/json"
	"log"
	"net/http"
)

// ErrorResponse sends a JSON error response
func ErrorResponse(w http.ResponseWriter, status int, message string) {
	JSONResponse(w, status, map[string]string{"error": message})
}

func JSONResponse(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("json encode error: %v", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
	}
}
