package utils

import (
	"errors"
	"net/http"
)

const MaxJSONBodySize = 1 << 20 // 1 MB

func LimitJSONBody(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxJSONBodySize)
}

func IsJSONBodyTooLarge(err error) bool {
	var maxBytesErr *http.MaxBytesError

	return errors.As(err, &maxBytesErr)
}