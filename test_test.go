package test_test

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"testing"

	"go.followtheprocess.codes/snapshot"
	"go.followtheprocess.codes/test"
)

var (
	update = flag.Bool("update", false, "Update snapshots")
	clean  = flag.Bool("clean", false, "Erase all snapshots and recreate them from scratch")
)

// TB is a fake implementation of [testing.TB] that simply records in internal
// state whether or not it would have failed and what it would have written.
type TB struct {
	testing.TB

	out    io.Writer
	failed bool
}

func (t *TB) Helper() {}

func (t *TB) Fatal(args ...any) {
	t.failed = true
	fmt.Fprint(t.out, args...)
}

func (t *TB) Fatalf(format string, args ...any) {
	t.failed = true
	fmt.Fprintf(t.out, format, args...)
}

func TestTest(t *testing.T) {
	tests := []struct {
		fn       func(tb testing.TB) // The test function we're... testing?
		name     string              // Name of the test case
		wantFail bool                // Whether it should fail
	}{
		{
			name: "Equal/pass",
			fn: func(tb testing.TB) {
				test.Equal(tb, "apples", "apples")
			},
			wantFail: false,
		},
		{
			name: "Equal/fail",
			fn: func(tb testing.TB) {
				test.Equal(tb, "apples", "oranges")
			},
			wantFail: true,
		},
		{
			name: "Equal/fail with context",
			fn: func(tb testing.TB) {
				test.Equal(tb, "apples", "oranges", test.Context("Apples are not oranges!"))
			},
			wantFail: true,
		},
		{
			name: "Equal/fail context format",
			fn: func(tb testing.TB) {
				test.Equal(tb, "apples", "oranges", test.Context("Apples == Oranges: %v", false))
			},
			wantFail: true,
		},
		{
			name: "Equal/fail with title",
			fn: func(tb testing.TB) {
				test.Equal(tb, "apples", "oranges", test.Title("My fruit test"))
			},
			wantFail: true,
		},
		{
			name: "NotEqual/pass",
			fn: func(tb testing.TB) {
				test.NotEqual(tb, "apples", "oranges")
			},
			wantFail: false,
		},
		{
			name: "NotEqual/fail",
			fn: func(tb testing.TB) {
				test.NotEqual(tb, "apples", "apples")
			},
			wantFail: true,
		},
		{
			name: "NotEqual/fail with context",
			fn: func(tb testing.TB) {
				test.NotEqual(tb, 42, 42, test.Context("42 is the meaning of life"))
			},
			wantFail: true,
		},
		{
			name: "NotEqual/fail context format",
			fn: func(tb testing.TB) {
				test.NotEqual(tb, 42, 42, test.Context("42 == meaning of life: %v", true))
			},
			wantFail: true,
		},
		{
			name: "NotEqual/fail with title",
			fn: func(tb testing.TB) {
				test.NotEqual(tb, "apples", "apples", test.Title("My fruit test"))
			},
			wantFail: true,
		},
		{
			name: "EqualFunc/pass",
			fn: func(tb testing.TB) {
				test.EqualFunc(tb, []int{1, 2, 3, 4}, []int{1, 2, 3, 4}, slices.Equal)
			},
			wantFail: false,
		},
		{
			name: "EqualFunc/fail",
			fn: func(tb testing.TB) {
				cmp := func(_, _ []string) bool { return false } // Cheating
				test.EqualFunc(tb, []string{"hello"}, []string{"there"}, cmp)
			},
			wantFail: true,
		},
		{
			name: "EqualFunc/fail with context",
			fn: func(tb testing.TB) {
				test.EqualFunc(
					tb,
					[]string{"hello"},
					[]string{"there"},
					slices.Equal,
					test.Context("some context here"),
				)
			},
			wantFail: true,
		},
		{
			name: "EqualFunc/fail context format",
			fn: func(tb testing.TB) {
				test.EqualFunc(
					tb,
					[]string{"hello"},
					[]string{"there"},
					slices.Equal,
					test.Context("who's bad at testing... %s", "you"),
				)
			},
			wantFail: true,
		},
		{
			name: "EqualFunc/fail with title",
			fn: func(tb testing.TB) {
				test.EqualFunc(tb, []string{"hello"}, []string{"there"}, slices.Equal, test.Title("Hello!"))
			},
			wantFail: true,
		},
		{
			name: "NotEqualFunc/pass",
			fn: func(tb testing.TB) {
				test.NotEqualFunc(tb, []int{1, 2, 3, 4}, []int{5, 6, 7, 8}, slices.Equal)
			},
			wantFail: false,
		},
		{
			name: "NotEqualFunc/fail",
			fn: func(tb testing.TB) {
				cmp := func(_, _ []string) bool { return true } // Cheating
				test.NotEqualFunc(tb, []string{"hello"}, []string{"there"}, cmp)
			},
			wantFail: true,
		},
		{
			name: "NotEqualFunc/fail with context",
			fn: func(tb testing.TB) {
				test.NotEqualFunc(
					tb,
					[]string{"hello"},
					[]string{"hello"},
					slices.Equal,
					test.Context("some context here"),
				)
			},
			wantFail: true,
		},
		{
			name: "NotEqualFunc/fail context format",
			fn: func(tb testing.TB) {
				test.NotEqualFunc(
					tb,
					[]string{"hello"},
					[]string{"hello"},
					slices.Equal,
					test.Context("who's bad at testing... %s", "you"),
				)
			},
			wantFail: true,
		},
		{
			name: "NotEqualFunc/fail with title",
			fn: func(tb testing.TB) {
				test.NotEqualFunc(tb, []string{"hello"}, []string{"hello"}, slices.Equal, test.Title("Hello!"))
			},
			wantFail: true,
		},
		{
			name: "NearlyEqual/pass",
			fn: func(tb testing.TB) {
				test.NearlyEqual(tb, 3.0000000001, 3.0)
			},
			wantFail: false,
		},
		{
			name: "NearlyEqual/fail",
			fn: func(tb testing.TB) {
				test.NearlyEqual(tb, 3.0000001, 3.0)
			},
			wantFail: true,
		},
		{
			name: "NearlyEqual/fail custom tolerance",
			fn: func(tb testing.TB) {
				test.NearlyEqual(tb, 3.2, 3.0, test.FloatEqualityThreshold(0.1))
			},
			wantFail: true,
		},
		{
			name: "NearlyEqual/fail with context",
			fn: func(tb testing.TB) {
				test.NearlyEqual(tb, 3.0000001, 3.0, test.Context("Numbers don't work that way"))
			},
			wantFail: true,
		},
		{
			name: "Ok/pass",
			fn: func(tb testing.TB) {
				test.Ok(tb, nil)
			},
			wantFail: false,
		},
		{
			name: "Ok/fail",
			fn: func(tb testing.TB) {
				test.Ok(tb, errors.New("uh oh"))
			},
			wantFail: true,
		},
		{
			name: "Ok/fail with context",
			fn: func(tb testing.TB) {
				test.Ok(tb, errors.New("uh oh"), test.Context("Could not frobnicate the baz"))
			},
			wantFail: true,
		},
		{
			name: "Ok/fail with title",
			fn: func(tb testing.TB) {
				test.Ok(tb, errors.New("uh oh"), test.Title("Bang!"))
			},
			wantFail: true,
		},
		{
			name: "Err/pass",
			fn: func(tb testing.TB) {
				test.Err(tb, errors.New("bang"))
			},
			wantFail: false,
		},
		{
			name: "Err/fail",
			fn: func(tb testing.TB) {
				test.Err(tb, nil)
			},
			wantFail: true,
		},
		{
			name: "Err/fail with context",
			fn: func(tb testing.TB) {
				test.Err(tb, nil, test.Context("Frobnicated the baz when it should have failed"))
			},
			wantFail: true,
		},
		{
			name: "Err/fail with title",
			fn: func(tb testing.TB) {
				test.Err(tb, nil, test.Title("Everything is fine?"))
			},
			wantFail: true,
		},
		{
			name: "WantErr/pass error",
			fn: func(tb testing.TB) {
				test.WantErr(tb, errors.New("bang"), true) // Wanted an error and got one - should pass
			},
			wantFail: false,
		},
		{
			name: "WantErr/pass nil",
			fn: func(tb testing.TB) {
				test.WantErr(tb, nil, false) // Didn't want an error and got nil - should pass
			},
			wantFail: false,
		},
		{
			name: "WantErr/fail error",
			fn: func(tb testing.TB) {
				test.WantErr(tb, errors.New("bang"), false) // Got an error but didn't want one - should fail
			},
			wantFail: true,
		},
		{
			name: "WantErr/fail nil",
			fn: func(tb testing.TB) {
				test.WantErr(tb, nil, true) // Didn't get an error but wanted one - should fail
			},
			wantFail: true,
		},
		{
			name: "WantErr/fail with context",
			fn: func(tb testing.TB) {
				test.WantErr(tb, errors.New("bang"), false, test.Context("Errors are bad!"))
			},
			wantFail: true,
		},
		{
			name: "WantErr/fail with title",
			fn: func(tb testing.TB) {
				test.WantErr(tb, errors.New("bang"), false, test.Title("A very bad test"))
			},
			wantFail: true,
		},
		{
			name: "True/pass",
			fn: func(tb testing.TB) {
				test.True(tb, true)
			},
			wantFail: false,
		},
		{
			name: "True/fail",
			fn: func(tb testing.TB) {
				test.True(tb, false)
			},
			wantFail: true,
		},
		{
			name: "True/fail with context",
			fn: func(tb testing.TB) {
				test.True(tb, false, test.Context("must always be true"))
			},
			wantFail: true,
		},
		{
			name: "True/fail with title",
			fn: func(tb testing.TB) {
				test.True(tb, false, test.Title("Argh!"))
			},
			wantFail: true,
		},
		{
			name: "False/pass",
			fn: func(tb testing.TB) {
				test.False(tb, false)
			},
			wantFail: false,
		},
		{
			name: "False/fail",
			fn: func(tb testing.TB) {
				test.False(tb, true)
			},
			wantFail: true,
		},
		{
			name: "False/fail with context",
			fn: func(tb testing.TB) {
				test.False(tb, true, test.Context("must always be false"))
			},
			wantFail: true,
		},
		{
			name: "False/fail with title",
			fn: func(tb testing.TB) {
				test.False(tb, true, test.Title("Argh!"))
			},
			wantFail: true,
		},
		{
			name: "Diff/pass",
			fn: func(tb testing.TB) {
				got := "Some\nstuff here in this file\nlines as well wow\nsome more stuff\n"
				want := "Some\nstuff here in this file\nlines as well wow\nsome more stuff\n"

				test.Diff(tb, got, want)
			},
			wantFail: false,
		},
		{
			name: "Diff/pass no trailing newline",
			fn: func(tb testing.TB) {
				got := "Some\nstuff here in this file\nlines as well wow\nsome more stuff"
				want := "Some\nstuff here in this file\nlines as well wow\nsome more stuff"

				test.Diff(tb, got, want)
			},
			wantFail: false,
		},
		{
			name: "Diff/fail",
			fn: func(tb testing.TB) {
				got := "Some\nstuff here in this file\nlines as well wow\nsome more stuff\n"
				want := "Some\ndifferent stuff here in this file\nthis line is different\nsome more stuff\n"
				test.Diff(tb, got, want)
			},
			wantFail: true,
		},
		{
			name: "Diff/fail no trailing newline",
			fn: func(tb testing.TB) {
				got := "Some\nstuff here in this file\nlines as well wow\nsome more stuff"
				want := "Some\ndifferent stuff here in this file\nthis line is different\nsome more stuff"
				test.Diff(tb, got, want)
			},
			wantFail: true,
		},
		{
			name: "DiffBytes/pass",
			fn: func(tb testing.TB) {
				got := []byte("Some\nstuff here in this file\nlines as well wow\nsome more stuff\n")
				want := []byte("Some\nstuff here in this file\nlines as well wow\nsome more stuff\n")

				test.DiffBytes(tb, got, want)
			},
			wantFail: false,
		},
		{
			name: "DiffBytes/fail",
			fn: func(tb testing.TB) {
				got := []byte("Some\nstuff here in this file\nlines as well wow\nsome more stuff\n")
				want := []byte("Some\ndifferent stuff here in this file\nthis line is different\nsome more stuff\n")
				test.DiffBytes(tb, got, want)
			},
			wantFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			tb := &TB{out: buf}
			snap := snapshot.New(t, snapshot.Update(*update), snapshot.Clean(*clean), snapshot.Color(false))

			if tb.failed {
				t.Fatalf("%s initial failed state should be false", tt.name)
			}

			// Call the test function, passing in the mock TB that just records
			// what a "real" TB would have done
			tt.fn(tb)

			if tb.failed != tt.wantFail {
				t.Fatalf("\nIncorrect Failure\n\ntb.failed:\t%v\nwanted:\t%v\n", tb.failed, tt.wantFail)
			}

			// Test the output matches our snapshot file, only for failed tests
			// as there should be no output for passed tests
			if !tb.failed {
				if buf.Len() != 0 {
					t.Fatalf("\nIncorrect Output\n\nA passed test should have no output, got: %s\n", buf.String())
				}
			} else {
				snap.Snap(buf.String())
			}
		})
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
