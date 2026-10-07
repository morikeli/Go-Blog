package middlewares

import "net/http"

// Middleware defines the standard signature for Go HTTP middleware components.
// It accepts a target handler and returns a new handler that wraps it.
type Middleware func(http.Handler) http.Handler

// Chain constructs an HTTP handler chain by executing middlewares in left-to-right order.
// For example, Chain(finalHandler, m1, m2) will execute m1 first, then m2, then finalHandler.
func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	// Loop backward so the first middleware passed in becomes the outermost wrapper.
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}