package test

import (
	"regexp"
	"testing"
)

// MatchError checks errors against a regexp.
//
// Returns true if the string is empty and the error is nil.
// Returns false if the string is not empty and the error is nil.
// Otherwise returns the result of a regexp match against the string.
func MatchError(t *testing.T, s string, e error) bool {
	t.Helper()
	if s == "" && e == nil {
		return true
	}
	if s != "" && e == nil {
		return false
	}
	match, err := regexp.MatchString(s, e.Error())
	if err != nil {
		t.Fatal(err)
	}
	return match
}
