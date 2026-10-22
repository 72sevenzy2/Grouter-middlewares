package middleware

import (
	"bytes"
	"github.com/72sevenzy2/http-router/core"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// custom responseWriter type to capture status code and request byte size.
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

// override WriteHeader with custom response writer struct
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code // saving status code in struct
	rw.ResponseWriter.WriteHeader(code)
}

// overriding Write to capture byte size
func (rw *responseWriter) Write(b []byte) (int, error) {
	v, err := rw.ResponseWriter.Write(b)
	rw.size += v // tracking the size in bytes
	return v, err
}

// custom limited writer function for Logger() to limit body size reading.
type LimitedBuffer struct {
	buf   *bytes.Buffer
	limit uint32
}

// custom write func for LimitedBuffer, (allocates new slice based on truncated size on original slice)
func (l *LimitedBuffer) Write(p []byte) (int, error) {
	remaining := l.limit - uint32(l.buf.Len())

	if remaining <= 0 {
		return len(p), nil
	}

	if len(p) > int(remaining) { // check if []byte that is being written to l.buf exceeds remaining.
		p = p[:remaining] // truncate
	}

	return l.buf.Write(p)
}

// plain structs to work with default values.
type bodySize struct {
	size uint32
}

func Logger(confSize uint32) core.Middleware {
	return func(hf core.HandlerFunc) core.HandlerFunc {
		return func(w http.ResponseWriter, r *core.Request) {
			start := time.Now()

			slog.Info("request started",
				"timestamp", start,
				"method", r.Method,
				"path", r.URL,
			)

			// buffer for comparison in limited writer Write().
			buf := bytes.Buffer{}

			// setting default value first for request body size
			var opt bodySize
			opt = bodySize{
				size: 1024,
			}

			if confSize != 0 {
				opt = bodySize{
					size: confSize,
				}
			}

			// limit size
			lm := &LimitedBuffer{
				buf:   &buf,
				limit: opt.size + 1,
			}

			r.Body = io.NopCloser(io.TeeReader(r.Body, lm))

			rw := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			hf(rw, r)

			// truncating if over 1 kb
			body := buf.Bytes()
			truncated := uint32(len(body)) > opt.size

			if truncated {
				body = body[:opt.size]
			}

			endTime := time.Since(start)

			slog.Info("request ended",
				"method", r.Method,
				"path", r.URL,
				"duration", endTime,
				"status_code", rw.status,
				"response_body_size_bytes", rw.size,
				"request_body_size_bytes", len(buf.Bytes()),
				"body_truncated", truncated,
			)

			slog.Debug("request body",
				"body", string(body),
			)

			// redacting sensitive header before printing
			header := r.Header.Clone()
			header.Del("Authorization")

			slog.Debug("request headers",
				"headers", header,
			)
		}
	}
}
