package middleware

import (
	"github.com/72sevenzy2/http-router/core"
	"github.com/72sevenzy2/json-parser"
	"log/slog"
	"net/http"
)

// recoverer middleware (for preventing server crashes)

func Recoverer() core.Middleware {
	return func(hf core.HandlerFunc) core.HandlerFunc {
		return func(w http.ResponseWriter, r *core.Request) {
			defer func() { // catches any crashses and recovers the request, while printing the err in return.
				if err := recover(); err != nil {
					slog.Error("Recoverer", "ERR", err)
					response.JSON(w, http.StatusInternalServerError, nil, http.StatusText(http.StatusInternalServerError))
				}
			}()

			hf(w, r) // next handler
		}
	}

}
