package middlewares

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

// responseRecorder records the HTTP status code while preserving the wrapped
// ResponseWriter. The Unwrap method also allows http.ResponseController to
// reach optional capabilities provided by the underlying writer.
type responseRecorder struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

// Unwrap exposes the underlying ResponseWriter to http.ResponseController.
func (r *responseRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	// net/http only uses the first WriteHeader call. Mirroring that behavior
	// prevents the logger from recording a status that was never sent.
	if r.wroteHeader {
		return
	}

	r.statusCode = statusCode
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(body []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}

	return r.ResponseWriter.Write(body)
}

// Flush preserves streaming responses such as Server-Sent Events.
func (r *responseRecorder) Flush() {
	r.WriteHeader(http.StatusOK)

	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack preserves support for protocols such as WebSockets when the
// underlying ResponseWriter supports connection hijacking.
func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}

	return hijacker.Hijack()
}

// Push preserves HTTP/2 server push support for writers that implement it.
// HTTP/2 server push is deprecated by browsers, but forwarding the interface
// keeps this wrapper transparent to the underlying writer.
func (r *responseRecorder) Push(target string, opts *http.PushOptions) error {
	pusher, ok := r.ResponseWriter.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}

	return pusher.Push(target, opts)
}

// ReadFrom preserves io.ReaderFrom so io.Copy can use the optimized path of
// the underlying ResponseWriter when available.
func (r *responseRecorder) ReadFrom(src io.Reader) (int64, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}

	if readerFrom, ok := r.ResponseWriter.(io.ReaderFrom); ok {
		return readerFrom.ReadFrom(src)
	}

	return io.Copy(r.ResponseWriter, src)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		recorder := &responseRecorder{
			ResponseWriter: w,
		}

		next.ServeHTTP(recorder, r)

		statusCode := recorder.statusCode
		if statusCode == 0 {
			statusCode = http.StatusOK
		}

		log.Printf(
			"request_id=%s method=%s path=%s status=%d duration=%s remote=%s",
			GetRequestID(r.Context()),
			r.Method,
			r.URL.Path,
			statusCode,
			time.Since(start),
			r.RemoteAddr,
		)
	})
}