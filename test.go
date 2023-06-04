// Package test provides a lightweight, but useful extension to the std lib's testing package
// for a friendlier and more intuitive API, as well some enhanced functionality like rich
// comparison and colour output.
package test

import "testing"

// Equal fails if got != want.
func Equal[T comparable](t testing.TB, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("\nGot:\t%+v\nWanted:\t%+v\n", got, want)
	}
}

// EqualFunc is like Equal but allows the user to pass a custom comparator, useful
// when the items to be compared to not implement the comparable generic constraint
//
// The comparator should return true if got and want should be considered equal.
func EqualFunc[T any](t testing.TB, got, want T, comparator func(got, want T) bool) {
	t.Helper()
	if !comparator(got, want) {
		t.Fatalf("\nGot:\t%+v\nWanted:\t%+v\n", got, want)
	}
}
