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

func (t *TB) Fatalf(format string, args ...any) {
	t.failed = true
	fmt.Fprintf(t.out, format, args...)
}

// TODO: Refactor all the tests below to fit into this table

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
				test.Equal(tb, "apples", "apples") // These obviously are equal
			},
			wantFail: false, // Should pass
			wantOut:  "",    // And write no output
		},
		{
			name: "equal string fail",
			testFunc: func(tb testing.TB) {
				test.Equal(tb, "apples", "oranges")
			},
			wantFail: true,
			wantOut:  "\nGot:\tapples\nWanted:\toranges\n",
		},
		{
			name: "equal int pass",
			testFunc: func(tb testing.TB) {
				test.Equal(tb, 1, 1)
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "equal int fail",
			testFunc: func(tb testing.TB) {
				test.Equal(tb, 1, 42)
			},
			wantFail: true,
			wantOut:  "\nGot:\t1\nWanted:\t42\n",
		},
		{
			name: "nearly equal pass",
			testFunc: func(tb testing.TB) {
				test.NearlyEqual(tb, 3.0000000001, 3.0)
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "nearly equal fail",
			testFunc: func(tb testing.TB) {
				test.NearlyEqual(tb, 3.0000001, 3.0)
			},
			wantFail: true,
			wantOut:  "\nGot:\t3.0000001\nWanted:\t3\n",
		},
		{
			name: "not equal string pass",
			testFunc: func(tb testing.TB) {
				test.NotEqual(tb, "apples", "oranges") // Should pass, these aren't equal
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "not equal string fail",
			testFunc: func(tb testing.TB) {
				test.NotEqual(tb, "apples", "apples")
			},
			wantFail: true,
			wantOut:  "\nValues were equal:\tapples\n",
		},
		{
			name: "not equal int pass",
			testFunc: func(tb testing.TB) {
				test.NotEqual(tb, 1, 42)
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "not equal int fail",
			testFunc: func(tb testing.TB) {
				test.NotEqual(tb, 1, 1)
			},
			wantFail: true,
			wantOut:  "\nValues were equal:\t1\n",
		},
		{
			name: "ok pass",
			testFunc: func(tb testing.TB) {
				test.Ok(tb, nil)
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "ok fail",
			testFunc: func(tb testing.TB) {
				test.Ok(tb, errors.New("uh oh"))
			},
			wantFail: true,
			wantOut:  "\nGot error:\tuh oh\nWanted:\tnil\n",
		},
		{
			name: "err pass",
			testFunc: func(tb testing.TB) {
				test.Err(tb, errors.New("uh oh"))
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "err fail",
			testFunc: func(tb testing.TB) {
				test.Err(tb, nil)
			},
			wantFail: true,
			wantOut:  "Error was nil\n",
		},
		{
			name: "true pass",
			testFunc: func(tb testing.TB) {
				test.True(tb, true)
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "true fail",
			testFunc: func(tb testing.TB) {
				test.True(tb, false)
			},
			wantFail: true,
			wantOut:  "\nGot:\tfalse\nWanted:\ttrue",
		},
		{
			name: "false pass",
			testFunc: func(tb testing.TB) {
				test.False(tb, false)
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "false fail",
			testFunc: func(tb testing.TB) {
				test.False(tb, true)
			},
			wantFail: true,
			wantOut:  "\nGot:\ttrue\nWanted:\tfalse",
		},
		{
			name: "equal func pass",
			testFunc: func(tb testing.TB) {
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
				rubbishEqual := func(a, b string) bool {
					return false // Never equal
				}
				test.EqualFunc(tb, "word", "word", rubbishEqual)
			},
			wantFail: true,
			wantOut:  "\nGot:\tword\nWanted:\tword\n",
		},
		{
			name: "not equal func pass",
			testFunc: func(tb testing.TB) {
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
				rubbishNotEqual := func(a, b string) bool {
					return true // Always equal
				}
				test.NotEqualFunc(tb, "word", "different word", rubbishNotEqual)
			},
			wantFail: true,
			wantOut:  "\nValues were equal:\tword\n",
		},
		{
			name: "deep equal pass",
			testFunc: func(tb testing.TB) {
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
				a := []string{"a", "b", "c"}
				b := []string{"d", "e", "f"}

				test.DeepEqual(tb, a, b)
			},
			wantFail: true,
			wantOut:  "\nGot:\t[a b c]\nWanted:\t[d e f]\n",
		},
		{
			name: "want err pass when got and wanted",
			testFunc: func(tb testing.TB) {
				test.WantErr(tb, errors.New("uh oh"), true) // We wanted an error and got one
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "want err fail when got and not wanted",
			testFunc: func(tb testing.TB) {
				test.WantErr(tb, errors.New("uh oh"), false) // Didn't want an error but got one
			},
			wantFail: true,
			wantOut:  "\nGot error:\tuh oh\nWanted error:\tfalse\n",
		},
		{
			name: "want err pass when not got and not wanted",
			testFunc: func(tb testing.TB) {
				test.WantErr(tb, nil, false) // Didn't want an error and didn't get one
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "want err fail when not got but wanted",
			testFunc: func(tb testing.TB) {
				test.WantErr(tb, nil, true) // Wanted an error but didn't get one
			},
			wantFail: true,
			wantOut:  "\nGot error:\t<nil>\nWanted error:\ttrue\n",
		},
		{
			name: "file pass",
			testFunc: func(tb testing.TB) {
				test.File(tb, "hello\n", filepath.Join(test.Data(t), "file.txt"))
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "diff pass string",
			testFunc: func(tb testing.TB) {
				test.Diff(tb, "hello", "hello")
			},
			wantFail: false,
			wantOut:  "",
		},
		{
			name: "diff fail string",
			testFunc: func(tb testing.TB) {
				test.Diff(tb, "hello", "hello there")
			},
			wantFail: true,
			wantOut: fmt.Sprintf(
				"Mismatch (-want, +got):\n%s",
				cmp.Diff("hello there", "hello"),
			), // Output equivalent to diff
		},
		{
			name: "diff pass string slice",
			testFunc: func(tb testing.TB) {
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
				got := []string{"hello", "there"}
				want := []string{"not", "me"}
				test.Diff(tb, got, want)
			},
			wantFail: true,
			wantOut: fmt.Sprintf(
				"Mismatch (-want, +got):\n%s",
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
				t.Errorf(
					"%s failure mismatch. failed: %v, wanted failure: %v",
					tt.name,
					tb.failed,
					tt.wantFail,
				)
			}

			if got := buf.String(); got != tt.wantOut {
				t.Errorf("%s output mismatch. got: %s, wanted %s", tt.name, got, tt.wantOut)
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
