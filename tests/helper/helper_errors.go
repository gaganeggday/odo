package helper

import (
	"regexp"
	"testing"
)

// MatchErrorString takes a string and matches on the error and returns true if the
// string matches the error.
//
// If the string can't be compiled as an regexp, then this will fail with a
// Fatal error.
func MatchErrorString(t *testing.T, s string, e error) bool {
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
