package client

import (
	"fmt"
	"net/http"
)

// IsNotFound returns true if the error represents a NotFound response from an
// upstream service.
func IsNotFound(err error) bool {
	e, ok := err.(scmError)
	return ok && e.Status == http.StatusNotFound
}

type scmError struct {
	msg    string
	Status int
}

func (s scmError) Error() string {
	return fmt.Sprintf("%s: (%d)", s.msg, s.Status)
}
