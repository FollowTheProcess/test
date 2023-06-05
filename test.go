// Package test provides a lightweight, but useful extension to the std lib's testing package
// for a friendlier and more intuitive API
package test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Equal fails if got != want.
func Equal[T comparable](t testing.TB, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("\nGot:\t%+v\nWanted:\t%+v\n", got, want)
	}
}

// EqualFunc is like Equal but allows the user to pass a custom comparator, useful
// when the items to be compared do not implement the comparable generic constraint
//
// The comparator should return true if the two items should be considered equal.
func EqualFunc[T any](t testing.TB, got, want T, equal func(a, b T) bool) {
	t.Helper()
	if !equal(got, want) {
		t.Fatalf("\nGot:\t%+v\nWanted:\t%+v\n", got, want)
	}
}

// NotEqual fails if got == want.
func NotEqual[T comparable](t testing.TB, got, want T) {
	t.Helper()
	if got == want {
		t.Fatalf("\nValues were equal:\t%+v\n", got)
	}
}

// NotEqualFunc is like NotEqual but allows the user to pass a custom comparator, useful
// when the items to be compared do not implement the comparable generic constraint
//
// The comparator should return true if the two items should be considered equal.
func NotEqualFunc[T any](t testing.TB, got, want T, equal func(a, b T) bool) {
	t.Helper()
	if equal(got, want) {
		t.Fatalf("\nValues were equal:\t%+v\n", got)
	}
}

// Ok fails if err != nil.
func Ok(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("\nGot error:\t%v\nWanted:\tnil\n", err)
	}
}

// True fails if v is false.
func True(t testing.TB, v bool) {
	t.Helper()
	if !v {
		t.Fatalf("\nGot:\t%+v\nWanted:\t%+v", v, true)
	}
}

// False fails if v is true.
func False(t testing.TB, v bool) {
	t.Helper()
	if v {
		t.Fatalf("\nGot:\t%+v\nWanted:\t%+v", v, false)
	}
}

// Diff fails if got != want and provides a rich diff.
func Diff[T any](t testing.TB, got, want T) {
	t.Helper()
	if diff := cmp.Diff(want, got, cmp.AllowUnexported(got, want)); diff != "" {
		t.Fatalf("Mismatch (-want, +got):\n%s", diff)
	}
}
