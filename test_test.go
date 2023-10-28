package test_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/FollowTheProcess/test"
)

type TB struct {
	testing.TB
	out    io.Writer
	failed bool
}

func (t *TB) Helper() {}

func (t *TB) Fatalf(format string, args ...any) {
	t.failed = true
	fmt.Fprintf(t.out, format, args...)
}

func TestPass(t *testing.T) {
	shouldPass := func(fn func(tb testing.TB)) {
		t.Helper()
		buf := &bytes.Buffer{}
		tb := &TB{out: buf}

		if tb.failed {
			t.Fatal("Initial failed state should be false")
		}

		// Call our test function
		fn(tb)

		if tb.failed {
			t.Fatal("Should have passed")
		}

		if buf.String() != "" {
			t.Fatalf("Shouldn't have written anything on success\nGot:\t%+v\n", buf.String())
		}
	}

	// All functions that should not fail their test TB
	passFns := []func(tb testing.TB){
		func(tb testing.TB) { test.Equal(tb, "hello", "hello") },
		func(tb testing.TB) { test.Equal(tb, "hello", "hello") },
		func(tb testing.TB) { test.Equal(tb, 42, 42) },
		func(tb testing.TB) { test.Equal(tb, true, true) },
		func(tb testing.TB) { test.Equal(tb, 3.14, 3.14) },
		func(tb testing.TB) { test.NotEqual(tb, "hello", "there") },
		func(tb testing.TB) { test.NotEqual(tb, 42, 27) },
		func(tb testing.TB) { test.NotEqual(tb, true, false) },
		func(tb testing.TB) { test.NotEqual(tb, 3.14, 8.67) },
		func(tb testing.TB) { test.Ok(tb, nil) },
		func(tb testing.TB) { test.Ok(tb, nil, "Something") },
		func(tb testing.TB) { test.Err(tb, errors.New("uh oh")) },
		func(tb testing.TB) { test.Err(tb, errors.New("uh oh"), "Something") },
		func(tb testing.TB) { test.True(tb, true) },
		func(tb testing.TB) { test.False(tb, false) },
		func(tb testing.TB) { test.Diff(tb, 42, 42) },
		func(tb testing.TB) { test.Diff(tb, true, true) },
		func(tb testing.TB) { test.Diff(tb, "hello", "hello") },
		func(tb testing.TB) { test.Diff(tb, 3.14, 3.14) },
		func(tb testing.TB) { test.Diff(tb, []string{"hello"}, []string{"hello"}) },
		func(tb testing.TB) {
			test.EqualFunc(tb, "something", "equal", func(got, want string) bool { return true })
		},
		func(tb testing.TB) { test.EqualFunc(tb, 42, 42, func(got, want int) bool { return true }) },
		func(tb testing.TB) {
			test.EqualFunc(tb, []string{"hello"}, []string{"hello"}, func(got, want []string) bool { return true })
		},
		func(tb testing.TB) {
			test.NotEqualFunc(tb, "something", "different", func(got, want string) bool { return false })
		},
		func(tb testing.TB) { test.NotEqualFunc(tb, 42, 12, func(got, want int) bool { return false }) },
		func(tb testing.TB) {
			test.NotEqualFunc(tb, []string{"hello"}, []string{"something", "else"}, func(got, want []string) bool { return false })
		},
		func(tb testing.TB) {
			test.Diff(tb, struct{ name string }{name: "dave"}, struct{ name string }{name: "dave"})
		},
		func(tb testing.TB) {
			test.Diff(tb, struct{ Name string }{Name: "dave"}, struct{ Name string }{Name: "dave"})
		},
		func(tb testing.TB) { test.DeepEqual(tb, []string{"hello"}, []string{"hello"}) },
		func(tb testing.TB) { test.ErrIsWanted(tb, errors.New("uh oh"), true) },
		func(tb testing.TB) { test.ErrIsWanted(tb, nilErr(), false) },
	}

	for _, fn := range passFns {
		shouldPass(fn)
	}
}

func TestFail(t *testing.T) {
	shouldFail := func(fn func(tb testing.TB)) {
		t.Helper()
		buf := &bytes.Buffer{}
		tb := &TB{out: buf}

		if tb.failed {
			t.Fatal("Initial failed state should be false")
		}

		// Call our test function
		fn(tb)

		if !tb.failed {
			t.Fatal("Should have failed")
		}

		if buf.String() == "" {
			t.Fatal("Should have written on failure")
		}
	}

	// All functions that should fail their test TB
	failFns := []func(tb testing.TB){
		func(tb testing.TB) { test.Equal(tb, "something", "else") },
		func(tb testing.TB) { test.Equal(tb, 42, 27) },
		func(tb testing.TB) { test.Equal(tb, true, false) },
		func(tb testing.TB) { test.Equal(tb, 3.14, 8.96) },
		func(tb testing.TB) { test.NotEqual(tb, "something", "something") },
		func(tb testing.TB) { test.NotEqual(tb, 42, 42) },
		func(tb testing.TB) { test.NotEqual(tb, true, true) },
		func(tb testing.TB) { test.NotEqual(tb, 3.14, 3.14) },
		func(tb testing.TB) { test.Ok(tb, errors.New("uh oh")) },
		func(tb testing.TB) { test.Ok(tb, errors.New("uh oh"), "Something") },
		func(tb testing.TB) { test.Err(tb, nilErr()) },
		func(tb testing.TB) { test.Err(tb, nilErr(), "Something") },
		func(tb testing.TB) { test.True(tb, false) },
		func(tb testing.TB) { test.False(tb, true) },
		func(tb testing.TB) { test.Diff(tb, "hello", "there") },
		func(tb testing.TB) { test.Diff(tb, 42, 27) },
		func(tb testing.TB) { test.Diff(tb, true, false) },
		func(tb testing.TB) { test.Diff(tb, 3.14, 8.69) },
		func(tb testing.TB) { test.Diff(tb, []string{"hello"}, []string{"there"}) },
		func(tb testing.TB) {
			test.EqualFunc(tb, "something", "different", func(got, want string) bool { return false })
		},
		func(tb testing.TB) { test.EqualFunc(tb, 42, 127, func(got, want int) bool { return false }) },
		func(tb testing.TB) {
			test.EqualFunc(tb, []int{42}, []int{27}, func(got, want []int) bool { return false })
		},
		func(tb testing.TB) {
			test.NotEqualFunc(tb, "something", "something", func(got, want string) bool { return true })
		},
		func(tb testing.TB) { test.NotEqualFunc(tb, 42, 42, func(got, want int) bool { return true }) },
		func(tb testing.TB) {
			test.NotEqualFunc(tb, []int{42}, []int{42}, func(got, want []int) bool { return true })
		},
		func(tb testing.TB) {
			test.Diff(tb, struct{ name string }{name: "dave"}, struct{ name string }{name: "john"})
		},
		func(tb testing.TB) {
			test.Diff(tb, struct{ Name string }{Name: "dave"}, struct{ Name string }{Name: "john"})
		},
		func(tb testing.TB) { test.DeepEqual(tb, []string{"hello"}, []string{"world"}) },
		func(tb testing.TB) { test.ErrIsWanted(tb, errors.New("uh oh"), false) },
		func(tb testing.TB) { test.ErrIsWanted(tb, nilErr(), true) },
	}

	for _, fn := range failFns {
		shouldFail(fn)
	}
}

// Always returns a nil error, needed because manually constructing
// nil means it's not an error type but here it is.
func nilErr() error {
	return nil
}
