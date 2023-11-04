// Package test provides a lightweight, but useful extension to the std lib's testing package
// with a friendlier and more intuitive API.
package test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Equal fails if got != want.
//
//	test.Equal(t, "apples", "apples") // Passes
//	test.Equal(t, "apples", "oranges") // Fails
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
//
//	test.NotEqual(t, "apples", "oranges") // Passes
//	test.NotEqual(t, "apples", "apples") // Fails
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

// Ok fails if err != nil, optionally adding context to the output.
//
//	err := doSomething()
//	test.Ok(t, err, "Doing something")
func Ok(t testing.TB, err error, context ...string) {
	t.Helper()
	var msg string
	if len(context) == 0 {
		msg = fmt.Sprintf("\nGot error:\t%v\nWanted:\tnil\n", err)
	} else {
		msg = fmt.Sprintf("\nGot error:\t%v\nWanted:\tnil\nContext:\t%s\n", err, context[0])
	}
	if err != nil {
		t.Fatalf(msg, err)
	}
}

// Err fails if err == nil.
//
//	err := shouldReturnErr()
//	test.Err(t, err, "shouldReturnErr")
func Err(t testing.TB, err error, context ...string) {
	t.Helper()
	var msg string
	if len(context) == 0 {
		msg = fmt.Sprintf("Error was not nil:\t%v\n", err)
	} else {
		msg = fmt.Sprintf("Error was not nil:\t%v\nContext:\t%s", err, context[0])
	}
	if err == nil {
		t.Fatalf(msg, err)
	}
}

// WantErr fails if you got an error and didn't want it, or if you
// didn't get an error but wanted one.
//
// It simplifies checking for errors in table driven tests where on any
// iteration err may or may not be nil.
//
//	test.WantErr(t, errors.New("uh oh"), true) // Passes, got error when we wanted one
//	test.WantErr(t, errors.New("uh oh"), false) // Fails, got error but didn't want one
//	test.WantErr(t, nil, true) // Fails, wanted an error but didn't get one
//	test.WantErr(t, nil, false) // Passes, didn't want an error and didn't get one
func WantErr(t testing.TB, err error, want bool) {
	t.Helper()
	if (err != nil) != want {
		t.Fatalf("\nGot error:\t%v\nWanted error:\t%v\n", err, want)
	}
}

// True fails if v is false.
//
//	test.True(t, true) // Passes
//	test.True(t, false) // Fails
func True(t testing.TB, v bool) {
	t.Helper()
	if !v {
		t.Fatalf("\nGot:\t%v\nWanted:\t%v", v, true)
	}
}

// False fails if v is true.
//
//	test.False(t, false) // Passes
//	test.False(t, true) // Fails
func False(t testing.TB, v bool) {
	t.Helper()
	if v {
		t.Fatalf("\nGot:\t%v\nWanted:\t%v", v, false)
	}
}

// Diff fails if got != want and provides a rich diff.
//
// If got and want are structs, unexported fields will be included in the comparison.
func Diff(t testing.TB, got, want any) {
	t.Helper()
	if reflect.TypeOf(got).Kind() == reflect.Struct {
		if diff := cmp.Diff(want, got, cmp.AllowUnexported(got, want)); diff != "" {
			t.Fatalf("Mismatch (-want, +got):\n%s", diff)
		}
	} else {
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("Mismatch (-want, +got):\n%s", diff)
		}
	}
}

// DeepEqual fails if reflect.DeepEqual(got, want) == false.
func DeepEqual(t testing.TB, got, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("\nGot:\t%+v\nWanted:\t%+v\n", got, want)
	}
}

// Data returns the filepath to the testdata directory for the current package.
//
// When running tests, Go will change the cwd to the directory of the package under test. This means
// that reference data stored in $CWD/testdata can be easily retrieved in the same way for any package.
//
// The $CWD/testdata directory is a Go idiom, common practice, and is completely ignored by the go tool.
//
// Data makes no guarantee that $CWD/testdata exists, it simply returns it's path.
//
//	file := filepath.Join(test.Data(t), "test.txt")
func Data(t testing.TB) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get $CWD: %v", err)
	}

	return filepath.Join(cwd, "testdata")
}
