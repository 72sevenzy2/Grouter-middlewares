package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/72sevenzy2/http-router/core"
)

// custom responseWriter type to capture status code and request byte size.
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

// custom limited writer function for Logger() to limit body size reading.
type LimitedBuffer struct {
	buf   *bytes.Buffer
	limit uint32
}

// plain structs to work with default values.
type bodySize struct {
	size uint32
}

var LoggerBuf = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
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
			buf := LoggerBuf.Get().(*bytes.Buffer)
			buf.Reset()

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
				buf:   buf,
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
				"request_body_size_bytes", len(body),
				"body_truncated", truncated,
			)

			slog.Debug("request body",
				"body", string(body),
			)
			headers := RedactSensitiveHeaders(r.Header.Clone())
			slog.Debug("request headers",
				"headers", headers,
			)
			LoggerBuf.Put(buf)
		}
	}
}

// WriteHeader() implements the http.ResponseWriter interface.
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write() implements the http.ResponseWriter interface.
func (rw *responseWriter) Write(b []byte) (int, error) {
	v, err := rw.ResponseWriter.Write(b)
	rw.size += v // tracking the size in bytes
	return v, err
}

// Write() implements the io.Writer interface (for LimitedBuffer{})
func (l *LimitedBuffer) Write(p []byte) (int, error) {
	remaining := int(l.limit) - l.buf.Len()
	if remaining <= 0 {
		return len(p), nil
	}

	if len(p) > int(remaining) { // check if []byte that is being written to l.buf exceeds remaining.
		p = p[:remaining] // truncate
	}

	return l.buf.Write(p)
}

// for redacting sensitive request headers before displaying.
func RedactSensitiveHeaders(headers http.Header) http.Header {
	// headers are excluded case-insensitively.
	headers.Del("authorization")
	headers.Del("proxy-authorization")
	headers.Del("cookie")
	headers.Del("set-cookie")
	headers.Del("x-api-key")
	headers.Del("x-auth-token")
	headers.Del("x-access-token")
	headers.Del("x-refresh-token")
	headers.Del("x-csrf-token")
	headers.Del("x-amz-security-token")
	headers.Del("x-goog-api-key")
	headers.Del("authentication-info")
	headers.Del("proxy-authenticate")

	return headers
}
