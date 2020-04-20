package util

import "regexp"

var (
	r = regexp.MustCompile("[a-zA-Z0-9_.+-]+@([a-zA-Z0-9-]+\\.[a-zA-Z0-9-.])")
)

// EmailMask masks the username part of the email address.
func EmailMask(s string) string {
	return r.ReplaceAllString(s, "***@$1")
}
