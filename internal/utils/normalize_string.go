package utils

import "strings"

// [NormalizeUsernameOrEmail] converts a username to its canonical representation.
// Usernames are case-insensitive in this application, so storing them in
// lowercase prevents values such as "Alice" and "alice" from behaving as
// different usernames on different API paths.
func NormalizeUsernameOrEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
