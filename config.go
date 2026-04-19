package test

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
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
	f.cfg.writeHeader(s)

	fmt.Fprintf(s, "Got:\t%+v\n", f.got)
	fmt.Fprintf(s, "Wanted:\t%+v\n", f.want)

	f.cfg.writeFooter(s)

	return s.String()
}

// writeHeader writes the title block (leading blank line, title, underline, blank line)
// to s. The underline is sized by rune count so multi-byte titles align correctly.
func (c config) writeHeader(s *strings.Builder) {
	s.WriteByte('\n')
	s.WriteString(c.title)
	s.WriteByte('\n')
	s.WriteString(strings.Repeat("-", utf8.RuneCountInString(c.title)))
	s.WriteString("\n\n")
}

// writeFooter writes any optional context and reason lines to s.
func (c config) writeFooter(s *strings.Builder) {
	if c.context != "" {
		fmt.Fprintf(s, "\n(%s)\n", c.context)
	}

	if c.reason != "" {
		fmt.Fprintf(s, "\nBecause: %s\n", c.reason)
	}
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
// Setting threshold to ±[math.Inf] is an error and will fail the test.
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
		context := strings.TrimSpace(fmt.Sprintf(format, args...))
		if context == "" {
			return errors.New("cannot set context to an empty string")
		}

		cfg.context = context

		return nil
	}

	return option(f)
}
