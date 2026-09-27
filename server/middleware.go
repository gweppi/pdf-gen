package server

import (
	"crypto/subtle"
	"net/http"
)

func APIKeyMiddleware(expectedKey string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientKey := r.Header.Get("X-API-Key")

		// Constant-time comparison prevents timing attacks
		if clientKey == "" || subtle.ConstantTimeCompare([]byte(clientKey), []byte(expectedKey)) != 1 {
			http.Error(w, `{"error": "Unauthorized: Invalid or missing API key"}`, http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
