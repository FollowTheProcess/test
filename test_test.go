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
	"github.com/google/go-cmp/cmp"
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

func TestPassFail(t *testing.T) {
	tests := []struct {
		testFunc func(tb testing.TB) // The test function we're... testing
		wantOut  string              // What we wanted the TB to print
		name     string              // Name of the test case
		wantFail bool                // Whether we wanted the testFunc to fail it's TB
	}{
		{
			name: "equal string pass",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.Equal(tb, "apples", "apples") // These obviously are equal
			},
			wantFail: false, // Should pass
			wantOut:  "",    // And write no output
		},
		{
			name: "equal string fail",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.Equal(tb, "apples", "oranges")
			},
			wantFail: true,
			wantOut:  "\nNot Equal\n---------\nGot:\tapples\nWanted:\toranges\n",
		},
		{
			name: "equal string fail with comment",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.Equal(tb, "apples", "oranges") // apples are not oranges
			},
			wantFail: true,
			wantOut:  "\nNot Equal  // apples are not oranges\n---------\nGot:\tapples\nWanted:\toranges\n",
		},
		{
			name: "equal int pass",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.Equal(tb, 1, 1)
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "equal int fail",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.Equal(tb, 1, 42)
			},
			wantFail: true,
			wantOut:  "\nNot Equal\n---------\nGot:\t1\nWanted:\t42\n",
		},
		{
			name: "nearly equal pass",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.NearlyEqual(tb, 3.0000000001, 3.0)
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "nearly equal fail",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.NearlyEqual(tb, 3.0000001, 3.0)
			},
			wantFail: true,
			wantOut:  "\nNot NearlyEqual\n---------------\nGot:\t3.0000001\nWanted:\t3\n\nDifference 9.999999983634211e-08 exceeds maximum tolerance of 1e-08\n",
		},
		{
			name: "nearly equal fail with comment",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.NearlyEqual(tb, 3.0000001, 3.0) // Ooof so close
			},
			wantFail: true,
			wantOut:  "\nNot NearlyEqual  // Ooof so close\n---------------\nGot:\t3.0000001\nWanted:\t3\n\nDifference 9.999999983634211e-08 exceeds maximum tolerance of 1e-08\n",
		},
		{
			name: "not equal string pass",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.NotEqual(tb, "apples", "oranges") // Should pass, these aren't equal
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "not equal string fail",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.NotEqual(tb, "apples", "apples")
			},
			wantFail: true,
			wantOut:  "\nValues were equal:\tapples\n",
		},
		{
			name: "not equal int pass",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.NotEqual(tb, 1, 42)
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "not equal int fail",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.NotEqual(tb, 1, 1)
			},
			wantFail: true,
			wantOut:  "\nValues were equal:\t1\n",
		},
		{
			name: "ok pass",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.Ok(tb, nil)
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "ok fail",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.Ok(tb, errors.New("uh oh"))
			},
			wantFail: true,
			wantOut:  "\nNot Ok\n------\nGot:\tuh oh\nWanted:\t<nil>\n",
		},
		{
			name: "ok fail with comment",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.Ok(tb, errors.New("uh oh")) // Calling some function
			},
			wantFail: true,
			wantOut:  "\nNot Ok  // Calling some function\n------\nGot:\tuh oh\nWanted:\t<nil>\n",
		},
		{
			name: "err pass",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.Err(tb, errors.New("uh oh"))
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "err fail",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.Err(tb, nil)
			},
			wantFail: true,
			wantOut:  "\nNot Err\n-------\nGot:\t<nil>\nWanted:\terror\n",
		},
		{
			name: "err fail with comment",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.Err(tb, nil) // Should have failed
			},
			wantFail: true,
			wantOut:  "\nNot Err  // Should have failed\n-------\nGot:\t<nil>\nWanted:\terror\n",
		},
		{
			name: "true pass",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.True(tb, true)
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "true fail",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.True(tb, false)
			},
			wantFail: true,
			wantOut:  "\nNot True\n--------\nGot:\tfalse\nWanted:\ttrue\n",
		},
		{
			name: "true fail with comment",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.True(tb, false) // Comment here
			},
			wantFail: true,
			wantOut:  "\nNot True  // Comment here\n--------\nGot:\tfalse\nWanted:\ttrue\n",
		},
		{
			name: "false pass",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.False(tb, false)
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "false fail",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.False(tb, true)
			},
			wantFail: true,
			wantOut:  "\nNot False\n---------\nGot:\ttrue\nWanted:\tfalse\n",
		},
		{
			name: "false fail with comment",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.False(tb, true) // Should always be false
			},
			wantFail: true,
			wantOut:  "\nNot False  // Should always be false\n---------\nGot:\ttrue\nWanted:\tfalse\n",
		},
		{
			name: "equal func pass",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				rubbishEqual := func(a, b string) bool {
					return true // Always equal
				}
				test.EqualFunc(tb, "word", "different word", rubbishEqual)
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "equal func fail",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				rubbishEqual := func(a, b string) bool {
					return false // Never equal
				}
				test.EqualFunc(tb, "word", "word", rubbishEqual)
			},
			wantFail: true,
			wantOut:  "\nNot Equal\n---------\nGot:\tword\nWanted:\tword\n\nequal(got, want) returned false\n",
		},
		{
			name: "equal func fail with comment",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				rubbishEqual := func(a, b string) bool {
					return false // Never equal
				}
				test.EqualFunc(tb, "word", "word", rubbishEqual) // Uh oh
			},
			wantFail: true,
			wantOut:  "\nNot Equal  // Uh oh\n---------\nGot:\tword\nWanted:\tword\n\nequal(got, want) returned false\n",
		},
		{
			name: "not equal func pass",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				rubbishNotEqual := func(a, b string) bool {
					return false // Never equal
				}
				test.NotEqualFunc(tb, "word", "word", rubbishNotEqual)
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "not equal func fail",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				rubbishNotEqual := func(a, b string) bool {
					return true // Always equal
				}
				test.NotEqualFunc(tb, "word", "different word", rubbishNotEqual)
			},
			wantFail: true,
			wantOut:  "\nValues were equal:\tword\n\nequal(got, want) returned true\n",
		},
		{
			name: "deep equal pass",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				a := []string{"a", "b", "c"}
				b := []string{"a", "b", "c"}

				test.DeepEqual(tb, a, b)
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "deep equal fail",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				a := []string{"a", "b", "c"}
				b := []string{"d", "e", "f"}

				test.DeepEqual(tb, a, b)
			},
			wantFail: true,
			wantOut:  "\nNot Equal\n---------\nGot:\t[a b c]\nWanted:\t[d e f]\n\nreflect.DeepEqual(got, want) returned false\n",
		},
		{
			name: "deep equal fail with comment",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				a := []string{"a", "b", "c"}
				b := []string{"d", "e", "f"}

				test.DeepEqual(tb, a, b) // Oh no!
			},
			wantFail: true,
			wantOut:  "\nNot Equal  // Oh no!\n---------\nGot:\t[a b c]\nWanted:\t[d e f]\n\nreflect.DeepEqual(got, want) returned false\n",
		},
		{
			name: "want err pass when got and wanted",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.WantErr(tb, errors.New("uh oh"), true) // We wanted an error and got one
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "want err fail when got and not wanted",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.WantErr(tb, errors.New("uh oh"), false) // Didn't want an error but got one
			},
			wantFail: true,
			wantOut:  "\nWantErr\n-------\nGot error:\tuh oh\nWanted error:\tfalse\n",
		},
		{
			name: "want err pass when not got and not wanted",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.WantErr(tb, nil, false) // Didn't want an error and didn't get one
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "want err fail when not got but wanted",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.WantErr(tb, nil, true) // Wanted an error but didn't get one
			},
			wantFail: true,
			wantOut:  "\nWantErr\n-------\nGot error:\t<nil>\nWanted error:\ttrue\n",
		},
		{
			name: "file pass",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.File(tb, "hello\n", filepath.Join(test.Data(t), "file.txt"))
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "diff pass string",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.Diff(tb, "hello", "hello")
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "diff fail string",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				test.Diff(tb, "hello", "hello there")
			},
			wantFail: true,
			wantOut: fmt.Sprintf(
				"\nMismatch (-want, +got):\n%s\n",
				cmp.Diff("hello there", "hello"),
			), // Output equivalent to diff
		},
		{
			name: "diff pass string slice",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				got := []string{"hello", "there"}
				want := []string{"hello", "there"}
				test.Diff(tb, got, want)
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "diff fail string slice",
			testFunc: func(tb testing.TB) {
				tb.Helper()
				got := []string{"hello", "there"}
				want := []string{"not", "me"}
				test.Diff(tb, got, want)
			},
			wantFail: true,
			wantOut: fmt.Sprintf(
				"\nMismatch (-want, +got):\n%s\n",
				cmp.Diff([]string{"not", "me"}, []string{"hello", "there"}),
			), // Output equivalent to diff
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			tb := &TB{out: buf}

			if tb.failed {
				t.Fatalf("%s initial failed state should be false", tt.name)
			}

			// Call the test function, passing in our mock TB that simply
			// records whether or not it would have failed and what it would
			// have written
			tt.testFunc(tb)

			if tb.failed != tt.wantFail {
				t.Fatalf(
					"\n%s failure mismatch\n--------------\nfailed:\t%v\nwanted failure:\t%v\n",
					tt.name,
					tb.failed,
					tt.wantFail,
				)
			}

			if got := buf.String(); got != tt.wantOut {
				t.Errorf(
					"\n%s output mismatch\n---------------\nGot:\t%s\nWanted:\t%s\n",
					tt.name,
					got,
					tt.wantOut,
				)
			}
		})
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
