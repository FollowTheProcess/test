// Package test provides a lightweight, but useful extension to the std lib's testing package
// with a friendlier and more intuitive API.
package test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/aymanbagabas/go-udiff"
	"github.com/google/go-cmp/cmp"
)

// floatEqualityThreshold allows us to do near-equality checks for floats.
const floatEqualityThreshold = 1e-8

// failure represents a test failure, including context and reason.
type failure[T any] struct {
	got     T      // What we got
	want    T      // Expected value
	title   string // Title of the failure, used as a header
	reason  string // Optional reason for additional context
	comment string // Optional line comment for context
}

// String prints a failure.
func (f failure[T]) String() string {
	var msg string
	if f.comment != "" {
		msg = fmt.Sprintf(
			"\n%s  // %s\n%s\nGot:\t%+v\nWanted:\t%+v\n",
			f.title,
			f.comment,
			strings.Repeat("-", len(f.title)),
			f.got,
			f.want,
		)
	} else {
		msg = fmt.Sprintf(
			"\n%s\n%s\nGot:\t%+v\nWanted:\t%+v\n",
			f.title,
			strings.Repeat("-", len(f.title)),
			f.got,
			f.want,
		)
	}

	if f.reason != "" {
		// Bolt the reason on the end
		msg = fmt.Sprintf("%s\n%s\n", msg, f.reason)
	}

	return msg
}

// Equal fails if got != want.
//
//	test.Equal(t, "apples", "apples") // Passes
//	test.Equal(t, "apples", "oranges") // Fails
func Equal[T comparable](tb testing.TB, got, want T) {
	tb.Helper()
	if got != want {
		fail := failure[T]{
			got:     got,
			want:    want,
			title:   "Not Equal",
			comment: getComment(),
		}
		tb.Fatal(fail.String())
	}
}

// NearlyEqual is like Equal but for floating point numbers where typically equality often fails.
//
// If the difference between got and want is sufficiently small, they are considered equal.
//
//	test.NearlyEqual(t, 3.0000000001, 3.0) // Passes, close enough to be considered equal
//	test.NearlyEqual(t, 3.0000001, 3.0) // Fails, too different
func NearlyEqual[T ~float32 | ~float64](tb testing.TB, got, want T) {
	tb.Helper()
	diff := math.Abs(float64(got - want))
	if diff >= floatEqualityThreshold {
		fail := failure[T]{
			got:   got,
			want:  want,
			title: "Not NearlyEqual",
			reason: fmt.Sprintf(
				"Difference %v exceeds maximum tolerance of %v",
				diff,
				floatEqualityThreshold,
			),
			comment: getComment(),
		}
		tb.Fatal(fail.String())
	}
}

// EqualFunc is like Equal but allows the user to pass a custom comparator, useful
// when the items to be compared do not implement the comparable generic constraint
//
// The comparator should return true if the two items should be considered equal.
func EqualFunc[T any](tb testing.TB, got, want T, equal func(a, b T) bool) {
	tb.Helper()
	if !equal(got, want) {
		fail := failure[T]{
			got:     got,
			want:    want,
			title:   "Not Equal",
			reason:  "equal(got, want) returned false",
			comment: getComment(),
		}
		tb.Fatal(fail.String())
	}
}

// NotEqual fails if got == want.
//
//	test.NotEqual(t, "apples", "oranges") // Passes
//	test.NotEqual(t, "apples", "apples") // Fails
func NotEqual[T comparable](tb testing.TB, got, want T) {
	tb.Helper()
	if got == want {
		if comment := getComment(); comment != "" {
			tb.Fatalf(
				"\nEqual  // %s\n%s\nGot:\t%+v\n\nExpected values to be different\n",
				comment,
				strings.Repeat("-", len("Equal")),
				got,
			)
		} else {
			tb.Fatalf("\nEqual\n%s\nGot:\t%+v\n\nExpected values to be different\n", strings.Repeat("-", len("Equal")), got)
		}
	}
}

// NotEqualFunc is like NotEqual but allows the user to pass a custom comparator, useful
// when the items to be compared do not implement the comparable generic constraint
//
// The comparator should return true if the two items should be considered equal.
func NotEqualFunc[T any](tb testing.TB, got, want T, equal func(a, b T) bool) {
	tb.Helper()
	if equal(got, want) {
		if comment := getComment(); comment != "" {
			tb.Fatalf(
				"\nEqual  // %s\n%s\nGot:\t%+v\n\nequal(got, want) returned true\n",
				comment,
				strings.Repeat("-", len("Equal")),
				got,
			)
		} else {
			tb.Fatalf("\nEqual\n%s\nGot:\t%+v\n\nequal(got, want) returned true\n", strings.Repeat("-", len("Equal")), got)
		}
	}
}

// Ok fails if err != nil.
//
//	err := doSomething()
//	test.Ok(t, err)
func Ok(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		fail := failure[error]{
			got:     err,
			want:    nil,
			title:   "Not Ok",
			comment: getComment(),
		}
		tb.Fatal(fail.String())
	}
}

// Err fails if err == nil.
//
//	err := shouldReturnErr()
//	test.Err(t, err)
func Err(tb testing.TB, err error) {
	tb.Helper()
	if err == nil {
		fail := failure[error]{
			got:     nil,
			want:    errors.New("error"),
			title:   "Not Err",
			comment: getComment(),
		}
		tb.Fatal(fail.String())
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
func WantErr(tb testing.TB, err error, want bool) {
	tb.Helper()
	if (err != nil) != want {
		var reason string
		var wanted error
		if want {
			reason = fmt.Sprintf("Wanted an error but got %v", err)
			wanted = errors.New("error")
		} else {
			reason = fmt.Sprintf("Got an unexpected error: %v", err)
			wanted = nil
		}
		fail := failure[any]{
			got:     err,
			want:    wanted,
			title:   "WantErr",
			reason:  reason,
			comment: getComment(),
		}
		tb.Fatal(fail.String())
	}
}

// True fails if v is false.
//
//	test.True(t, true) // Passes
//	test.True(t, false) // Fails
func True(tb testing.TB, v bool) {
	tb.Helper()
	if !v {
		fail := failure[bool]{
			got:     v,
			want:    true,
			title:   "Not True",
			comment: getComment(),
		}
		tb.Fatal(fail.String())
	}
}

// False fails if v is true.
//
//	test.False(t, false) // Passes
//	test.False(t, true) // Fails
func False(tb testing.TB, v bool) {
	tb.Helper()
	if v {
		fail := failure[bool]{
			got:     v,
			want:    false,
			title:   "Not False",
			comment: getComment(),
		}
		tb.Fatal(fail.String())
	}
}

// Diff fails if got != want and provides a rich diff.
func Diff(tb testing.TB, got, want any) {
	// TODO: Nicer output for diff, don't like the +got -want thing
	tb.Helper()
	if diff := cmp.Diff(want, got); diff != "" {
		tb.Fatalf("\nMismatch (-want, +got):\n%s\n", diff)
	}
}

// DeepEqual fails if reflect.DeepEqual(got, want) == false.
func DeepEqual(tb testing.TB, got, want any) {
	tb.Helper()
	if !reflect.DeepEqual(got, want) {
		fail := failure[any]{
			got:     got,
			want:    want,
			title:   "Not Equal",
			reason:  "reflect.DeepEqual(got, want) returned false",
			comment: getComment(),
		}
		tb.Fatal(fail.String())
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
func Data(tb testing.TB) string {
	tb.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		tb.Fatalf("could not get $CWD: %v", err)
	}

	return filepath.Join(cwd, "testdata")
}

// File fails if got does not match the contents of the given file.
//
// It takes a string and the path of a file to compare, use [Data] to obtain
// the path to the current packages testdata directory.
//
// If the contents differ, the test will fail with output similar to executing git diff
// on the contents.
//
// Files with differing line endings (e.g windows CR LF \r\n vs unix LF \n) will be normalised to
// \n prior to comparison so this function will behave identically across multiple platforms.
//
//	test.File(t, "hello\n", "expected.txt")
func File(tb testing.TB, got, file string) {
	tb.Helper()
	f, err := filepath.Abs(file)
	if err != nil {
		tb.Fatalf("could not make %s absolute: %v", file, err)
	}
	contents, err := os.ReadFile(f)
	if err != nil {
		tb.Fatalf("could not read %s: %v", f, err)
	}

	contents = bytes.ReplaceAll(contents, []byte("\r\n"), []byte("\n"))

	if diff := udiff.Unified("want", "got", string(contents), got); diff != "" {
		tb.Fatalf("\nMismatch\n--------\n%s\n", diff)
	}
}

// CaptureOutput captures and returns data printed to stdout and stderr by the provided function fn, allowing
// you to test functions that write to those streams and do not have an option to pass in an [io.Writer].
//
// If the provided function returns a non nil error, the test is failed with the error logged as the reason.
//
// If any error occurs capturing stdout or stderr, the test will also be failed with a descriptive log.
//
//	fn := func() error {
//		fmt.Println("hello stdout")
//		return nil
//	}
//
//	stdout, stderr := test.CaptureOutput(t, fn)
//	fmt.Print(stdout) // "hello stdout\n"
//	fmt.Print(stderr) // ""
func CaptureOutput(tb testing.TB, fn func() error) (stdout, stderr string) {
	tb.Helper()

	// Take copies of the original streams
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	defer func() {
		// Restore everything back to normal
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		tb.Fatalf("CaptureOutput: could not construct an os.Pipe(): %v", err)
	}

	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		tb.Fatalf("CaptureOutput: could not construct an os.Pipe(): %v", err)
	}

	// Set stdout and stderr streams to the pipe writers
	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	stdoutCapture := make(chan string)
	stderrCapture := make(chan string)

	var wg sync.WaitGroup
	wg.Add(2) //nolint: mnd

	// Copy in goroutines to avoid blocking
	go func(wg *sync.WaitGroup) {
		defer func() {
			close(stdoutCapture)
			wg.Done()
		}()
		buf := &bytes.Buffer{}
		if _, err := io.Copy(buf, stdoutReader); err != nil {
			tb.Fatalf("CaptureOutput: failed to copy from stdout reader: %v", err)
		}
		stdoutCapture <- buf.String()
	}(&wg)

	go func(wg *sync.WaitGroup) {
		defer func() {
			close(stderrCapture)
			wg.Done()
		}()
		buf := &bytes.Buffer{}
		if _, err := io.Copy(buf, stderrReader); err != nil {
			tb.Fatalf("CaptureOutput: failed to copy from stderr reader: %v", err)
		}
		stderrCapture <- buf.String()
	}(&wg)

	// Call the test function that produces the output
	if err := fn(); err != nil {
		tb.Fatalf("CaptureOutput: user function returned an error: %v", err)
	}

	// Close the writers
	stdoutWriter.Close()
	stderrWriter.Close()

	capturedStdout := <-stdoutCapture
	capturedStderr := <-stderrCapture

	wg.Wait()

	return capturedStdout, capturedStderr
}

// getComment loads a Go line comment from a line where a test function has been called.
//
// If any error happens or there is no comment, an empty string is returned so as not
// to influence the test with an unrelated error.
func getComment() string {
	skip := 2 // Skip 2 frames, one for this function, the other for the calling test function
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}

	f, err := os.Open(file)
	if err != nil {
		return ""
	}
	defer f.Close()

	currentLine := 1 // Line numbers in source files start from 1
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// Skip through until we get to the line returned from runtime.Caller
		if currentLine != line {
			currentLine++
			continue
		}

		_, comment, ok := strings.Cut(scanner.Text(), "//")
		if !ok {
			// There was no comment on this line
			return ""
		}

		// Now comment will be everything from the "//" until the end of the line
		return strings.TrimSpace(comment)
	}

	// Didn't find one
	return ""
}
