// Package test provides a lightweight, but useful extension to the std lib's testing package
// with a friendlier and more intuitive API.
package test

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// floatEqualityThreshold allows us to do near-equality checks for floats.
const floatEqualityThreshold = 1e-8

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

// NearlyEqual is like Equal but for floating point numbers where typically equality often fails.
//
// If the difference between got and want is sufficiently small, they are considered equal.
//
//	test.NearlyEqual(t, 3.0000000001, 3.0) // Passes, close enough to be considered equal
//	test.NearlyEqual(t, 3.0000001, 3.0) // Fails, too different
func NearlyEqual[T ~float32 | ~float64](t testing.TB, got, want T) {
	t.Helper()
	if math.Abs(float64(got-want)) >= floatEqualityThreshold {
		t.Fatalf("\nGot:\t%v\nWanted:\t%v\n", got, want)
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

// File fails if the contents of the given file do not match want.
//
// It takes the name of a file (relative to $CWD/testdata) and the contents to compare.
//
// If the contents differ, the test will fail with output equivalent to [Diff].
//
// Files with differing line endings (e.g windows CR LF \r\n vs unix LF \n) will be normalised to
// \n prior to comparison so this function will behave identically across multiple platforms.
//
//	test.File(t, "expected.txt", "hello\n")
func File(t testing.TB, file, want string) {
	t.Helper()
	file = filepath.Join(Data(t), file)
	contents, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("could not read %s: %v", file, err)
	}

	contents = bytes.ReplaceAll(contents, []byte("\r\n"), []byte("\n"))

	Diff(t, string(contents), want)
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
func CaptureOutput(t testing.TB, fn func() error) (stdout, stderr string) {
	t.Helper()

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
		t.Fatalf("CaptureOutput: could not construct an os.Pipe(): %v", err)
	}

	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("CaptureOutput: could not construct an os.Pipe(): %v", err)
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
			t.Fatalf("CaptureOutput: failed to copy from stdout reader: %v", err)
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
			t.Fatalf("CaptureOutput: failed to copy from stderr reader: %v", err)
		}
		stderrCapture <- buf.String()
	}(&wg)

	// Call the test function that produces the output
	if err := fn(); err != nil {
		t.Fatalf("CaptureOutput: user function returned an error: %v", err)
	}

	// Close the writers
	stdoutWriter.Close()
	stderrWriter.Close()

	capturedStdout := <-stdoutCapture
	capturedStderr := <-stderrCapture

	wg.Wait()

	return capturedStdout, capturedStderr
}
