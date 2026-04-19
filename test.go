// Package test provides a lightweight, but useful extension to the std lib's testing package with
// a friendlier and more intuitive API.
//
// Simple tests become trivial and test provides mechanisms for adding useful context to test failures.
package test // import "go.followtheprocess.codes/test"

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"

	"go.followtheprocess.codes/diff"
	"go.followtheprocess.codes/diff/render"
	"go.followtheprocess.codes/hue"
)

// errAny is a sentinel rendered in failure output when the caller expected
// an error but got nil — it reads more naturally than the previous
// placeholder errors.New("error") (which printed "Wanted: error").
var errAny = errors.New("<any error>")

// ColorEnabled sets whether the output from this package is colourised.
//
// test defaults to automatic detection based on a number of attributes:
//   - The value of $NO_COLOR and/or $FORCE_COLOR
//   - The value of $TERM
//   - Whether [os.Stdout] is pointing to a terminal
//
// This means that test should do a reasonable job of auto-detecting when to colourise output
// and should not write escape sequences when piping between processes or when writing to files etc.
//
// ColorEnabled may be called safely from concurrently executing goroutines.
func ColorEnabled(v bool) {
	hue.Enabled(v)
}

// Equal fails if got != want.
//
//	test.Equal(t, "apples", "apples") // Passes
//	test.Equal(t, "apples", "oranges") // Fails
func Equal[T comparable](tb testing.TB, got, want T, options ...Option) {
	tb.Helper()

	cfg := defaultConfig()
	cfg.title = "Not Equal"

	for _, option := range options {
		if err := option.apply(&cfg); err != nil {
			tb.Fatalf("Equal: could not apply options: %v", err)

			return
		}
	}

	if got != want {
		fail := failure[T]{
			got:  got,
			want: want,
			cfg:  cfg,
		}
		tb.Fatal(fail.String())
	}
}

// NotEqual is the opposite of [Equal], it fails if got == want.
//
//	test.NotEqual(t, 10, 42) // Passes
//	test.NotEqual(t, 42, 42) // Fails
func NotEqual[T comparable](tb testing.TB, got, want T, options ...Option) {
	tb.Helper()

	cfg := defaultConfig()
	cfg.title = "Equal"

	for _, option := range options {
		if err := option.apply(&cfg); err != nil {
			tb.Fatalf("NotEqual: could not apply options: %v", err)

			return
		}
	}

	if got == want {
		fail := failure[T]{
			got:  got,
			want: want,
			cfg:  cfg,
		}
		tb.Fatal(fail.String())
	}
}

// EqualFunc is like [Equal] but accepts a custom comparator function, useful
// when the items to be compared do not implement the comparable generic constraint.
//
// The signature of the comparator is such that standard library functions such as
// [slices.Equal] or [maps.Equal] can be used.
//
// The comparator should return true if the two items should be considered equal.
//
//	test.EqualFunc(t, []int{1, 2, 3}, []int{1, 2, 3}, slices.Equal) // Passes
//	test.EqualFunc(t, []int{1, 2, 3}, []int{4, 5, 6}, slices.Equal) // Fails
func EqualFunc[T any](tb testing.TB, got, want T, equal func(a, b T) bool, options ...Option) {
	tb.Helper()

	cfg := defaultConfig()
	cfg.title = "Not Equal"

	for _, option := range options {
		if err := option.apply(&cfg); err != nil {
			tb.Fatalf("EqualFunc: could not apply options: %v", err)

			return
		}
	}

	if !equal(got, want) {
		cfg.reason = "equal(got, want) returned false"
		fail := failure[T]{
			got:  got,
			want: want,
			cfg:  cfg,
		}
		tb.Fatal(fail.String())
	}
}

// NotEqualFunc is like [NotEqual] but accepts a custom comparator function, useful
// when the items to be compared do not implement the comparable generic constraint.
//
// The signature of the comparator is such that standard library functions such as
// [slices.Equal] or [maps.Equal] can be used.
//
// The comparator should return true if the two items should be considered equal.
//
//	test.NotEqualFunc(t, []int{1, 2, 3}, []int{1, 2, 3}, slices.Equal) // Fails
//	test.NotEqualFunc(t, []int{1, 2, 3}, []int{4, 5, 6}, slices.Equal) // Passes
func NotEqualFunc[T any](tb testing.TB, got, want T, equal func(a, b T) bool, options ...Option) {
	tb.Helper()

	cfg := defaultConfig()
	cfg.title = "Equal"

	for _, option := range options {
		if err := option.apply(&cfg); err != nil {
			tb.Fatalf("NotEqualFunc: could not apply options: %v", err)

			return
		}
	}

	if equal(got, want) {
		cfg.reason = "equal(got, want) returned true"
		fail := failure[T]{
			got:  got,
			want: want,
			cfg:  cfg,
		}
		tb.Fatal(fail.String())
	}
}

// NearlyEqual is like [Equal] but for floating point numbers where absolute equality often fails.
//
// If the difference between got and want is sufficiently small, they are considered equal. This threshold
// defaults to 1e-8 but can be configured with the [FloatEqualityThreshold] option.
//
//	test.NearlyEqual(t, 3.0000000001, 3.0) // Passes, close enough to be considered equal
//	test.NearlyEqual(t, 3.0000001, 3.0) // Fails, too different
func NearlyEqual[T ~float32 | ~float64](tb testing.TB, got, want T, options ...Option) {
	tb.Helper()

	cfg := defaultConfig()
	cfg.title = "Not NearlyEqual"

	for _, option := range options {
		if err := option.apply(&cfg); err != nil {
			tb.Fatalf("NearlyEqual: could not apply options: %v", err)

			return
		}
	}

	delta := math.Abs(float64(got - want))
	if delta > cfg.floatEqualityThreshold {
		cfg.reason = fmt.Sprintf(
			"Difference %v - %v = %v exceeds maximum tolerance of %v",
			got,
			want,
			delta,
			cfg.floatEqualityThreshold,
		)
		fail := failure[T]{
			got:  got,
			want: want,
			cfg:  cfg,
		}
		tb.Fatal(fail.String())
	}
}

// NotNearlyEqual is the opposite of [NearlyEqual]. It fails when got and want
// are within the float equality threshold of each other.
//
// The threshold defaults to 1e-8 and can be configured with the
// [FloatEqualityThreshold] option.
//
//	test.NotNearlyEqual(t, 3.0000001, 3.0) // Passes, different enough
//	test.NotNearlyEqual(t, 3.0000000001, 3.0) // Fails, too close to be considered different
func NotNearlyEqual[T ~float32 | ~float64](tb testing.TB, got, want T, options ...Option) {
	tb.Helper()

	cfg := defaultConfig()
	cfg.title = "NearlyEqual"

	for _, option := range options {
		if err := option.apply(&cfg); err != nil {
			tb.Fatalf("NotNearlyEqual: could not apply options: %v", err)

			return
		}
	}

	delta := math.Abs(float64(got - want))
	if delta <= cfg.floatEqualityThreshold {
		cfg.reason = fmt.Sprintf(
			"Difference %v - %v = %v is within tolerance of %v",
			got,
			want,
			delta,
			cfg.floatEqualityThreshold,
		)
		fail := failure[T]{
			got:  got,
			want: want,
			cfg:  cfg,
		}
		tb.Fatal(fail.String())
	}
}

// Ok fails if err != nil.
//
//	err := doSomething()
//	test.Ok(t, err)
func Ok(tb testing.TB, err error, options ...Option) {
	tb.Helper()

	cfg := defaultConfig()
	cfg.title = "Not Ok"

	for _, option := range options {
		if optionErr := option.apply(&cfg); optionErr != nil {
			tb.Fatalf("Ok: could not apply options: %v", optionErr)

			return
		}
	}

	if err != nil {
		fail := failure[error]{
			got:  err,
			want: nil,
			cfg:  cfg,
		}
		tb.Fatal(fail.String())
	}
}

// Err fails if err == nil.
//
//	err := shouldFail()
//	test.Err(t, err)
func Err(tb testing.TB, err error, options ...Option) {
	tb.Helper()

	cfg := defaultConfig()
	cfg.title = "Not Err"

	for _, option := range options {
		if optionErr := option.apply(&cfg); optionErr != nil {
			tb.Fatalf("Err: could not apply options: %v", optionErr)

			return
		}
	}

	if err == nil {
		fail := failure[error]{
			got:  nil,
			want: errAny,
			cfg:  cfg,
		}
		tb.Fatal(fail.String())
	}
}

// ErrorIs fails if err does not match target as reported by [errors.Is].
//
//	var ErrMadeUp = errors.New("made up error")
//	test.ErrorIs(t, err, ErrMadeUp)
func ErrorIs(tb testing.TB, err, target error, options ...Option) {
	tb.Helper()

	cfg := defaultConfig()
	cfg.title = "Wrong Error"

	for _, option := range options {
		if optionErr := option.apply(&cfg); optionErr != nil {
			tb.Fatalf("ErrorIs: could not apply options: %v", optionErr)

			return
		}
	}

	if !errors.Is(err, target) {
		fail := failure[error]{
			got:  err,
			want: target,
			cfg:  cfg,
		}
		tb.Fatal(fail.String())
	}
}

// ErrorAs asserts that err or some error in its chain matches the concrete
// type T as reported by [errors.AsType], and returns the matched error so
// the caller can make further assertions on its fields without having to
// unwrap it a second time. The return value may be ignored.
//
// Discard the return when you only care about the type check:
//
//	test.ErrorAs[*os.PathError](t, err)
//
// Or bind it to drill into the matched error:
//
//	got := test.ErrorAs[*os.PathError](t, err)
//	test.Equal(t, got.Op, "open")
//	test.Equal(t, got.Path, "/does/not/exist")
func ErrorAs[T error](tb testing.TB, err error, options ...Option) T {
	tb.Helper()

	cfg := defaultConfig()
	cfg.title = "Wrong Error Type"

	for _, option := range options {
		if optionErr := option.apply(&cfg); optionErr != nil {
			tb.Fatalf("ErrorAs: could not apply options: %v", optionErr)

			var zero T

			return zero
		}
	}

	if target, ok := errors.AsType[T](err); ok {
		return target
	}

	got := "<nil>"
	if err != nil {
		got = fmt.Sprintf("%T: %s", err, err.Error())
	}

	fail := failure[string]{
		got:  got,
		want: fmt.Sprintf("error matching %s", reflect.TypeFor[T]()),
		cfg:  cfg,
	}
	tb.Fatal(fail.String())

	var zero T

	return zero
}

// WantErr fails if you got an error and didn't want it, or if you didn't
// get an error but wanted one.
//
// It greatly simplifies checking for errors in table driven tests where an error
// may or may not be nil on any given test case.
//
//	test.WantErr(t, errors.New("uh oh"), true) // Passes, got error when we wanted one
//	test.WantErr(t, errors.New("uh oh"), false) // Fails, got error but didn't want one
//	test.WantErr(t, nil, true) // Fails, wanted an error but didn't get one
//	test.WantErr(t, nil, false) // Passes, didn't want an error and didn't get one
func WantErr(tb testing.TB, err error, want bool, options ...Option) {
	tb.Helper()

	cfg := defaultConfig()
	cfg.title = "WantErr"

	for _, option := range options {
		if optionErr := option.apply(&cfg); optionErr != nil {
			tb.Fatalf("WantErr: could not apply options: %v", optionErr)

			return
		}
	}

	if (err != nil) != want {
		var (
			reason string
			wanted error
		)

		if want {
			reason = fmt.Sprintf("Wanted an error but got %v", err)
			wanted = errAny
		} else {
			reason = fmt.Sprintf("Got an unexpected error: %v", err)
			wanted = nil
		}

		cfg.reason = reason
		fail := failure[error]{
			got:  err,
			want: wanted,
			cfg:  cfg,
		}
		tb.Fatal(fail.String())
	}
}

// True fails if got is false.
//
//	test.True(t, true) // Passes
//	test.True(t, false) // Fails
func True(tb testing.TB, got bool, options ...Option) {
	tb.Helper()

	cfg := defaultConfig()
	cfg.title = "Not True"

	for _, option := range options {
		if err := option.apply(&cfg); err != nil {
			tb.Fatalf("True: could not apply options: %v", err)

			return
		}
	}

	if !got {
		fail := failure[bool]{
			got:  got,
			want: true,
			cfg:  cfg,
		}
		tb.Fatal(fail.String())
	}
}

// False fails if got is true.
//
//	test.False(t, false) // Passes
//	test.False(t, true) // Fails
func False(tb testing.TB, got bool, options ...Option) {
	tb.Helper()

	cfg := defaultConfig()
	cfg.title = "Not False"

	for _, option := range options {
		if err := option.apply(&cfg); err != nil {
			tb.Fatalf("False: could not apply options: %v", err)

			return
		}
	}

	if got {
		fail := failure[bool]{
			got:  got,
			want: false,
			cfg:  cfg,
		}
		tb.Fatal(fail.String())
	}
}

// Diff fails if the two strings got and want are not equal and provides a rich
// unified diff of the two for easy comparison.
//
// If either got or want do not end in a newline, one is added to avoid a
// "No newline at end of file" warning in the diff which is visually distracting.
func Diff(tb testing.TB, got, want string, options ...Option) {
	tb.Helper()
	DiffBytes(tb, []byte(got), []byte(want), options...)
}

// DiffBytes fails if the two []byte got and want are not equal and provides a rich
// unified diff of the two for easy comparison.
//
// If either got or want do not end in a newline, one is added to avoid a
// "No newline at end of file" warning in the diff which is visually distracting.
func DiffBytes(tb testing.TB, got, want []byte, options ...Option) {
	tb.Helper()

	cfg := defaultConfig()
	cfg.title = "Diff"

	for _, option := range options {
		if err := option.apply(&cfg); err != nil {
			tb.Fatalf("DiffBytes: could not apply options: %v", err)

			return
		}
	}

	got = fixNL(got)
	want = fixNL(want)

	d := diff.New("want", want, "got", got)

	if !d.Equal() {
		s := &strings.Builder{}
		cfg.writeHeader(s)
		s.Write(render.Render(d))
		cfg.writeFooter(s)
		tb.Fatal(s.String())
	}
}

// DiffReader reads data from both got and want [io.Reader] and provides
// a rich unified diff of the two for easy comparison.
//
// If either got or want do not end in a newline, one is added to avoid
// a "No newline at end of file" warning in the diff which is visually distracting.
func DiffReader(tb testing.TB, got, want io.Reader, options ...Option) {
	tb.Helper()

	gotData, err := io.ReadAll(got)
	if err != nil {
		tb.Fatalf("DiffReader: could not read from got: %v", err)
	}

	wantData, err := io.ReadAll(want)
	if err != nil {
		tb.Fatalf("DiffReader: could not read from want: %v", err)
	}

	DiffBytes(tb, gotData, wantData, options...)
}

// CaptureOutput captures and returns data printed to [os.Stdout] and [os.Stderr] by the provided function fn, allowing
// you to test functions that write to those streams and do not have an option to pass in an [io.Writer].
//
// If the provided function returns a non nil error, the test is failed with the error logged as the reason.
//
// If any error occurs capturing stdout or stderr, the test will also be failed with a descriptive log.
//
// CaptureOutput replaces the process-wide [os.Stdout] and [os.Stderr] for the duration of the call,
// so it is NOT safe to use from tests marked with [testing.T.Parallel].
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

	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		tb.Fatalf("CaptureOutput: could not construct an os.Pipe(): %v", err)

		return "", ""
	}

	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		stdoutReader.Close()
		stdoutWriter.Close()
		tb.Fatalf("CaptureOutput: could not construct an os.Pipe(): %v", err)

		return "", ""
	}

	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	// Buffered so the copy goroutines can always deliver their result, even if
	// the main goroutine exits early via Fatalf / runtime.Goexit / panic.
	stdoutCapture := make(chan string, 1)
	stderrCapture := make(chan string, 1)

	// Goroutines use Errorf rather than Fatalf — the testing contract says
	// FailNow/Fatal* must only be called from the main test goroutine.
	go copyInto(tb, "stdout", stdoutReader, stdoutCapture)
	go copyInto(tb, "stderr", stderrReader, stderrCapture)

	// Ensure the real streams are restored and the pipe writers are closed on
	// every exit path (including panic / Goexit). Closing the writers lets the
	// copy goroutines see EOF and deliver their buffers to the channels.
	writersClosed := false
	closeWriters := func() {
		if writersClosed {
			return
		}

		writersClosed = true

		stdoutWriter.Close()
		stderrWriter.Close()
	}

	defer func() {
		closeWriters()

		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	if fnErr := fn(); fnErr != nil {
		tb.Fatalf("CaptureOutput: user function returned an error: %v", fnErr)

		return "", ""
	}

	// Happy path: close writers now so we can receive the captured data before
	// the defer runs (the defer's close is then a no-op).
	closeWriters()

	return <-stdoutCapture, <-stderrCapture
}

// copyInto reads from r into a buffer and sends the result on out. Any copy
// error is reported against tb via Errorf — Fatal* is unsafe from non-main
// goroutines. The send uses a buffered channel so it never blocks.
func copyInto(tb testing.TB, name string, r io.Reader, out chan<- string) {
	tb.Helper()

	buf := &bytes.Buffer{}

	defer func() {
		out <- buf.String()
	}()

	if _, err := io.Copy(buf, r); err != nil {
		tb.Errorf("CaptureOutput: failed to copy from %s reader: %v", name, err)
	}
}

// If data is empty or ends in \n, fixNL returns data.
// Otherwise fixNL returns a new slice consisting of data with a final \n added.
func fixNL(data []byte) []byte {
	if len(data) == 0 || data[len(data)-1] == '\n' {
		return data
	}

	d := make([]byte, len(data)+1)
	copy(d, data)
	d[len(data)] = '\n'

	return d
}
