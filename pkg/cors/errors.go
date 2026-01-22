package cors

import "errors"

var (
	// ErrNoCORSConfig is returned when no CORS configuration exists.
	ErrNoCORSConfig = errors.New("cors: bucket has no CORS configuration")
)
