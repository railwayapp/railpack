package mise

import (
	"errors"
	"fmt"
)

// TemporaryError marks a failure that is worth retrying: network transport
// errors, rate limits, and server-side errors. Deterministic failures (a 404
// for a version that does not exist, a malformed archive) are never wrapped in
// this, since retrying them cannot succeed.
type TemporaryError struct {
	URL string
	Err error
}

func (e *TemporaryError) Error() string {
	return fmt.Sprintf("temporary failure fetching %s: %s", e.URL, e.Err)
}

func (e *TemporaryError) Unwrap() error {
	return e.Err
}

// reports whether err (or anything it wraps) is a transient failure. Callers
// use this to decide whether retrying the whole operation makes sense.
func IsTemporary(err error) bool {
	var temporary *TemporaryError
	return errors.As(err, &temporary)
}
