package httpx

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
)

func APIMiddleware(next http.Handler, apiToken, frontendOrigin string) http.Handler {
	return corsMiddleware(frontendOrigin, authenticationMiddleware(next, apiToken))
}

func authenticationMiddleware(next http.Handler, apiToken string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if apiToken == "" || !isProtectedPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		const prefix = "Bearer "
		authorization := r.Header.Get("Authorization")
		if len(authorization) != len(prefix)+len(apiToken) ||
			!strings.HasPrefix(authorization, prefix) ||
			subtle.ConstantTimeCompare(
				[]byte(authorization[len(prefix):]),
				[]byte(apiToken),
			) != 1 {
			writeAccessError(w, http.StatusUnauthorized, "unauthorized", "valid API authentication is required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isProtectedPath(path string) bool {
	return path == "/stats" || path == "/jobs" || strings.HasPrefix(path, "/jobs/")
}

func corsMiddleware(frontendOrigin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if frontendOrigin != "" && origin != "" {
			if origin != frontendOrigin {
				if r.Method == http.MethodOptions {
					writeAccessError(w, http.StatusForbidden, "forbidden_origin", "the request origin is not allowed")
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", frontendOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Idempotency-Key")
			w.Header().Add("Vary", "Origin")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeAccessError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]map[string]string{
		"error": {
			"code":    code,
			"message": message,
		},
	})
}
