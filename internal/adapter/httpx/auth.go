package httpx

import (
	"crypto/subtle"
	"net/http"
)

// RequireBearerAuth accepts exactly one bearer credential matching the configured fake key.
func RequireBearerAuth(r *http.Request, want string) bool {
	values := r.Header.Values("Authorization")
	if want == "" || len(values) != 1 {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(values[0]), []byte("Bearer "+want)) == 1
}
