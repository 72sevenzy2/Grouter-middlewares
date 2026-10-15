package middleware

import (
	"github.com/72sevenzy2/http-router/core"
	"github.com/72sevenzy2/json-parser/response"
	"net/http"
	"strings"
)

// bearer auth middleware (this includes having a bearer token which will then be compared to the authkey )

func BearerAuth(AuthKey string) core.Middleware {
	return func(hf core.HandlerFunc) core.HandlerFunc {
		return func(w http.ResponseWriter, r *core.Request) {
			authLab := r.Header.Get("Authorization") // grabbing the token

			scheme, token, ok := strings.Cut(authLab, " ")
			// case-insensitive auth-header value matching (RFC 9110) with UTF-8 comparison.
			if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
				response.JSON(w, http.StatusUnauthorized, nil, http.StatusText(http.StatusUnauthorized))
				return
			}

			// continuing if "Bearer " doesnt include in the authkey.
			if AuthKey != authLab { // check if the authkey is matching
				response.JSON(w, http.StatusForbidden, nil, http.StatusText(http.StatusForbidden))
				return // exit the request
			}

			hf(w, r) // continue to next handler
		}
	}
}
