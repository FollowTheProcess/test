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
	"strings"
	"sync"
	"testing"

	"go.followtheprocess.codes/hue"
	"go.followtheprocess.codes/test/internal/diff"
)

const (
	header = hue.Cyan | hue.Bold
	green  = hue.Green
	red    = hue.Red
)

// ColorEnabled sets whether the output from this package is colourised.
//
// test defaults to automatic detection based on a number of attributes:
//   - The value of $NO_COLOR and/or $FORCE_COLOR
//   - The value of $TERM (xterm enables colour)
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

// NotEqualFunc is like [Equal] but accepts a custom comparator function, useful
// when the items to be compared do not implement the comparable generic constraint.
//
// The signature of the comparator is such that standard library functions such as
// [slices.Equal] or [maps.Equal] can be used.
//
// The comparator should return true if the two items should be considered equal.
//
//	test.EqualFunc(t, []int{1, 2, 3}, []int{1, 2, 3}, slices.Equal) // Fails
//	test.EqualFunc(t, []int{1, 2, 3}, []int{4, 5, 6}, slices.Equal) // Passes
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

	diff := math.Abs(float64(got - want))
	if diff > cfg.floatEqualityThreshold {
		cfg.reason = fmt.Sprintf(
			"Difference %v - %v = %v exceeds maximum tolerance of %v",
			got,
			want,
			diff,
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
			want: errors.New("error"),
			cfg:  cfg,
		}
		tb.Fatal(fail.String())
	}
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
			wanted = errors.New("error")
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
func Diff(tb testing.TB, got, want string) {
	tb.Helper()

	// TODO(@FollowTheProcess): If either got or want don't end in a newline, add one
	if diff := diff.Diff("want", []byte(want), "got", []byte(got)); diff != nil {
		tb.Fatalf("\nDiff\n----\n%s\n", prettyDiff(string(diff)))
	}
}

// DiffBytes fails if the two []byte got and want are not equal and provides a rich
// unified diff of the two for easy comparison.
func DiffBytes(tb testing.TB, got, want []byte) {
	tb.Helper()

	if diff := diff.Diff("want", want, "got", got); diff != nil {
		tb.Fatalf("\nDiff\n----\n%s\n", prettyDiff(string(diff)))
	}
}

// CaptureOutput captures and returns data printed to [os.Stdout] and [os.Stderr] by the provided function fn, allowing
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

	wg.Add(2) //nolint: mnd // 2 because stdout and stderr

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
	stdoutCloseErr := stdoutWriter.Close()
	if stdoutCloseErr != nil {
		tb.Fatalf("CaptureOutput: could not close stdout pipe: %v", stdoutCloseErr)
	}

	stderrCloseErr := stderrWriter.Close()
	if stderrCloseErr != nil {
		tb.Fatalf("CaptueOutput: could not close stderr pipe: %v", stderrCloseErr)
	}

	capturedStdout := <-stdoutCapture
	capturedStderr := <-stderrCapture

	wg.Wait()

	return capturedStdout, capturedStderr
}

// prettyDiff takes a string diff in unified diff format and colourises it for easier viewing.
func prettyDiff(diff string) string {
	lines := strings.Split(diff, "\n")
	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "---") || strings.HasPrefix(trimmed, "- ") {
			lines[i] = red.Sprint(lines[i])
		}

		if strings.HasPrefix(trimmed, "@@") {
			lines[i] = header.Sprint(lines[i])
		}

		if strings.HasPrefix(trimmed, "+++") || strings.HasPrefix(trimmed, "+ ") {
			lines[i] = green.Sprint(lines[i])
		}
	}

	return strings.Join(lines, "\n")
}
