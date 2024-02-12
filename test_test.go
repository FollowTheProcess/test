package test_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	shouldPass := func(name string, fn func(tb testing.TB)) {
		t.Helper()
		buf := &bytes.Buffer{}
		tb := &TB{out: buf}

		if tb.failed {
			t.Fatalf("%s initial failed state should be false", name)
		}

		// Call our test function
		fn(tb)

		if tb.failed {
			t.Fatalf("%s should have passed", name)
		}

		if buf.String() != "" {
			t.Fatalf("%s houldn't have written anything on success\nGot:\t%+v\n", name, buf.String())
		}
	}

	// All functions that should not fail their test TB
	passFns := map[string]func(tb testing.TB){
		"Equal string":      func(tb testing.TB) { test.Equal(tb, "hello", "hello") },
		"Equal int":         func(tb testing.TB) { test.Equal(tb, 42, 42) },
		"Equal bool":        func(tb testing.TB) { test.Equal(tb, true, true) },
		"Equal float":       func(tb testing.TB) { test.Equal(tb, 3.14, 3.14) },
		"NearlyEqual":       func(tb testing.TB) { test.NearlyEqual(tb, 3.0000000001, 3.0) },
		"NotEqual string":   func(tb testing.TB) { test.NotEqual(tb, "hello", "there") },
		"NotEqual int":      func(tb testing.TB) { test.NotEqual(tb, 42, 27) },
		"NotEqual bool":     func(tb testing.TB) { test.NotEqual(tb, true, false) },
		"NotEqual float":    func(tb testing.TB) { test.NotEqual(tb, 3.14, 8.67) },
		"Ok nil":            func(tb testing.TB) { test.Ok(tb, nil) },
		"Ok with context":   func(tb testing.TB) { test.Ok(tb, nil, "Something") },
		"Err":               func(tb testing.TB) { test.Err(tb, errors.New("uh oh")) },
		"Err with context":  func(tb testing.TB) { test.Err(tb, errors.New("uh oh"), "Something") },
		"True":              func(tb testing.TB) { test.True(tb, true) },
		"False":             func(tb testing.TB) { test.False(tb, false) },
		"Diff int":          func(tb testing.TB) { test.Diff(tb, 42, 42) },
		"Diff bool":         func(tb testing.TB) { test.Diff(tb, true, true) },
		"Diff string":       func(tb testing.TB) { test.Diff(tb, "hello", "hello") },
		"Diff float":        func(tb testing.TB) { test.Diff(tb, 3.14, 3.14) },
		"Diff string slice": func(tb testing.TB) { test.Diff(tb, []string{"hello"}, []string{"hello"}) },
		"EqualFunc string": func(tb testing.TB) {
			test.EqualFunc(tb, "something", "equal", func(_, _ string) bool { return true })
		},
		"EqualFunc int": func(tb testing.TB) { test.EqualFunc(tb, 42, 42, func(_, _ int) bool { return true }) },
		"EqualFunc string slice": func(tb testing.TB) {
			test.EqualFunc(tb, []string{"hello"}, []string{"hello"}, func(_, _ []string) bool { return true })
		},
		"NotEqualFunc string": func(tb testing.TB) {
			test.NotEqualFunc(tb, "something", "different", func(_, _ string) bool { return false })
		},
		"NotEqualFunc int": func(tb testing.TB) { test.NotEqualFunc(tb, 42, 12, func(_, _ int) bool { return false }) },
		"NotEqualFunc string slice": func(tb testing.TB) {
			test.NotEqualFunc(tb, []string{"hello"}, []string{"something", "else"}, func(_, _ []string) bool { return false })
		},
		"Diff unexported struct": func(tb testing.TB) {
			test.Diff(tb, struct{ name string }{name: "dave"}, struct{ name string }{name: "dave"})
		},
		"Diff exported struct": func(tb testing.TB) {
			test.Diff(tb, struct{ Name string }{Name: "dave"}, struct{ Name string }{Name: "dave"})
		},
		"DeepEqual string slice": func(tb testing.TB) { test.DeepEqual(tb, []string{"hello"}, []string{"hello"}) },
		"WantErr true":           func(tb testing.TB) { test.WantErr(tb, errors.New("uh oh"), true) },
		"WantErr false":          func(tb testing.TB) { test.WantErr(tb, nilErr(), false) },
		"File":                   func(tb testing.TB) { test.File(tb, "file.txt", "hello\n") },
	}

	for name, fn := range passFns {
		shouldPass(name, fn)
	}
}

func TestFail(t *testing.T) {
	shouldFail := func(name string, fn func(tb testing.TB)) {
		t.Helper()
		buf := &bytes.Buffer{}
		tb := &TB{out: buf}

		if tb.failed {
			t.Fatalf("%s initial failed state should be false", name)
		}

		// Call our test function
		fn(tb)

		if !tb.failed {
			t.Fatalf("%s should have failed", name)
		}

		if buf.String() == "" {
			t.Fatalf("%s should have written on failure", name)
		}
	}

	// All functions that should fail their test TB
	failFns := map[string]func(tb testing.TB){
		"Equal string":      func(tb testing.TB) { test.Equal(tb, "something", "else") },
		"Equal int":         func(tb testing.TB) { test.Equal(tb, 42, 27) },
		"Equal bool":        func(tb testing.TB) { test.Equal(tb, true, false) },
		"Equal float":       func(tb testing.TB) { test.Equal(tb, 3.14, 8.96) },
		"NearlyEqual":       func(tb testing.TB) { test.NearlyEqual(tb, 3.0000001, 3.0) },
		"NotEqual string":   func(tb testing.TB) { test.NotEqual(tb, "something", "something") },
		"NotEqual int":      func(tb testing.TB) { test.NotEqual(tb, 42, 42) },
		"NotEqual bool":     func(tb testing.TB) { test.NotEqual(tb, true, true) },
		"NotEqual float":    func(tb testing.TB) { test.NotEqual(tb, 3.14, 3.14) },
		"Ok":                func(tb testing.TB) { test.Ok(tb, errors.New("uh oh")) },
		"Ok with context":   func(tb testing.TB) { test.Ok(tb, errors.New("uh oh"), "Something") },
		"Err":               func(tb testing.TB) { test.Err(tb, nilErr()) },
		"Err with context":  func(tb testing.TB) { test.Err(tb, nilErr(), "Something") },
		"True":              func(tb testing.TB) { test.True(tb, false) },
		"False":             func(tb testing.TB) { test.False(tb, true) },
		"Diff string":       func(tb testing.TB) { test.Diff(tb, "hello", "there") },
		"Diff int":          func(tb testing.TB) { test.Diff(tb, 42, 27) },
		"Diff bool":         func(tb testing.TB) { test.Diff(tb, true, false) },
		"Diff float":        func(tb testing.TB) { test.Diff(tb, 3.14, 8.69) },
		"Diff string slice": func(tb testing.TB) { test.Diff(tb, []string{"hello"}, []string{"there"}) },
		"EqualFunc string": func(tb testing.TB) {
			test.EqualFunc(tb, "something", "different", func(_, _ string) bool { return false })
		},
		"EqualFunc int": func(tb testing.TB) { test.EqualFunc(tb, 42, 127, func(_, _ int) bool { return false }) },
		"EqualFunc string slice": func(tb testing.TB) {
			test.EqualFunc(tb, []int{42}, []int{27}, func(_, _ []int) bool { return false })
		},
		"NotEqualFunc string": func(tb testing.TB) {
			test.NotEqualFunc(tb, "something", "something", func(_, _ string) bool { return true })
		},
		"NotEqualFunc int": func(tb testing.TB) { test.NotEqualFunc(tb, 42, 42, func(_, _ int) bool { return true }) },
		"NotEqualFunc int slice": func(tb testing.TB) {
			test.NotEqualFunc(tb, []int{42}, []int{42}, func(_, _ []int) bool { return true })
		},
		"Diff unexported struct": func(tb testing.TB) {
			test.Diff(tb, struct{ name string }{name: "dave"}, struct{ name string }{name: "john"})
		},
		"Diff exported struct": func(tb testing.TB) {
			test.Diff(tb, struct{ Name string }{Name: "dave"}, struct{ Name string }{Name: "john"})
		},
		"DeepEqual string slice": func(tb testing.TB) { test.DeepEqual(tb, []string{"hello"}, []string{"world"}) },
		"WantErr true":           func(tb testing.TB) { test.WantErr(tb, errors.New("uh oh"), false) },
		"WantErr false":          func(tb testing.TB) { test.WantErr(tb, nilErr(), true) },
		"File wrong":             func(tb testing.TB) { test.File(tb, "file.txt", "wrong\n") },
		"File missing":           func(tb testing.TB) { test.File(tb, "missing.txt", "wrong\n") },
	}

	for name, fn := range failFns {
		shouldFail(name, fn)
	}
}

func TestData(t *testing.T) {
	got := test.Data(t)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Test for Data could not get cwd: %v", err)
	}

	want := filepath.Join(cwd, "testdata")

	if got != want {
		t.Errorf("\nGot:\t%s\nWanted:\t%s\n", got, want)
	}
}

func TestCapture(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		// Some fake user function that writes to stdout and stderr
		fn := func() error {
			fmt.Fprintln(os.Stdout, "hello stdout")
			fmt.Fprintln(os.Stderr, "hello stderr")

			return nil
		}

		stdout, stderr := test.CaptureOutput(t, fn)

		test.Equal(t, stdout, "hello stdout\n")
		test.Equal(t, stderr, "hello stderr\n")
	})

	t.Run("sad", func(t *testing.T) {
		// This time the user function returns an error
		fn := func() error {
			return errors.New("it broke")
		}

		buf := &bytes.Buffer{}
		testTB := &TB{out: buf}

		stdout, stderr := test.CaptureOutput(testTB, fn)

		// Test should have failed
		test.True(t, testTB.failed)

		// stdout and stderr should be empty
		test.Equal(t, stdout, "")
		test.Equal(t, stderr, "")
	})
}

// Always returns a nil error, needed because manually constructing
// nil means it's not an error type but here it is.
func nilErr() error {
	return nil
}
