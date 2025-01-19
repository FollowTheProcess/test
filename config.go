package test

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

const (
	defaultFloatEqualityThreshold = 1e-8
)

// config holds test-specific configuration including additional context
// and how the caller wants this library to behave.
type config struct {
	title                  string  // Title of the test, shown as a header in the failure log
	context                string  // Additional context passed by the caller
	reason                 string  // Concise reason why the test has failed, only used sparingly and not in a user option
	floatEqualityThreshold float64 // The difference threshold below which two floats are considered equal
}

// defaultConfig returns a default configuration.
func defaultConfig() config {
	return config{
		floatEqualityThreshold: defaultFloatEqualityThreshold,
	}
}

// failure represents a test failure, including any set config.
type failure[T any] struct {
	got  T      // The actual value
	want T      // Expected value
	cfg  config // Test config
}

// String implements [fmt.Stringer] for failure, allowing it to print itself in the test log.
func (f failure[T]) String() string {
	s := &strings.Builder{}
	s.WriteByte('\n')

	s.WriteString(f.cfg.title)
	s.WriteByte('\n')
	s.WriteString(strings.Repeat("-", len(f.cfg.title)))
	s.WriteString("\n\n")

	fmt.Fprintf(s, "Got:\t%+v\n", f.got)
	fmt.Fprintf(s, "Wanted:\t%+v\n", f.want)

	if f.cfg.context != "" {
		fmt.Fprintf(s, "\n(%s)\n", f.cfg.context)
	}

	if f.cfg.reason != "" {
		fmt.Fprintf(s, "\nBecause: %s\n", f.cfg.reason)
	}

	return s.String()
}

// Option is a configuration option for a test.
type Option interface {
	// Apply the option to the test config, returning an error if the option
	// cannot be applied for whatever reason.
	apply(cfg *config) error
}

// option is a function adapter implementing the Option interface, analogous
// to how http.HandlerFunc implements the Handler interface.
type option func(cfg *config) error

// apply applies the option, implementing the Option interface for the option
// function adapter.
func (o option) apply(cfg *config) error {
	return o(cfg)
}

// FloatEqualityThreshold is an [Option] to set the maximum difference allowed between
// two floating point numbers before they are considered equal. This setting is only
// used in [NearlyEqual] and [NotNearlyEqual].
//
// Setting threshold to ±math.Inf is an error and will fail the test.
//
// The default is 1e-8, a sensible default for most cases.
func FloatEqualityThreshold(threshold float64) Option {
	f := func(cfg *config) error {
		if math.IsInf(threshold, 0) {
			return errors.New("cannot set floating point equality threshold to ±infinity")
		}

		cfg.floatEqualityThreshold = threshold

		return nil
	}

	return option(f)
}

// Title is an [Option] that sets the title of the test in the test failure log.
//
// The title is shown as an underlined header in the test failure, below which the
// actual and expected values will be shown.
//
// By default this will be named sensibly after the test function being called, for
// example [Equal] has a default title "Not Equal".
//
// Setting title explicitly to the empty string "" is an error and will fail the test.
//
//	test.Equal(t, "apples", "oranges", test.Title("Wrong fruits!"))
func Title(title string) Option {
	f := func(cfg *config) error {
		if title == "" {
			return errors.New("cannot set title to an empty string")
		}

		cfg.title = strings.TrimSpace(title)

		return nil
	}

	return option(f)
}

// Context is an [Option] that allows the caller to inject useful contextual information
// as to why the test failed. This can be a useful addition to the test failure output log.
//
// The signature of context allows the use of fmt print verbs to format the message in the
// same way one might use [fmt.Sprintf].
//
// It is not necessary to include a newline character at the end of format.
//
// Setting context explicitly to the empty string "" is an error and will fail the test.
//
// For example:
//
//	err := doSomethingComplicated()
//	test.Ok(t, err, test.Context("something complicated failed"))
func Context(format string, args ...any) Option {
	f := func(cfg *config) error {
		if format == "" {
			return errors.New("cannot set context to an empty string")
		}

		context := fmt.Sprintf(format, args...)
		cfg.context = strings.TrimSpace(context)

		return nil
	}

	return option(f)
}
