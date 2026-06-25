package middleware

import (
	"crypto/subtle"
	"net/http"

	"website-of-methodological-materials/internal/handlers"
)

// AdminAuth проверяет заголовок X-Admin-Token
func AdminAuth(expectedToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("X-Admin-Token")
			if token == "" || !secureCompare(token, expectedToken) {
				handlers.RespondError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func secureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
