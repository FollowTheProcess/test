package test_test

import (
	"bytes"
	"fmt"
	"io"
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
	shouldPass := func(fn func(tb testing.TB)) {
		t.Helper()
		buf := &bytes.Buffer{}
		tb := &TB{out: buf}

		if tb.failed {
			t.Fatal("Initial failed state should be false")
		}

		// Call our test function
		fn(tb)

		if tb.failed {
			t.Fatal("Should have passed")
		}

		if buf.String() != "" {
			t.Fatalf("Shouldn't have written anything on success\nGot:\t%+v\n", buf.String())
		}
	}

	shouldPass(func(tb testing.TB) { test.Equal(tb, "hello", "hello") })
	shouldPass(func(tb testing.TB) { test.Equal(tb, 42, 42) })
	shouldPass(func(tb testing.TB) { test.Equal(tb, true, true) })
	shouldPass(func(tb testing.TB) { test.Equal(tb, 3.14, 3.14) })

	shouldPass(func(tb testing.TB) {
		test.EqualFunc(tb, "something", "equal", func(got, want string) bool { return true })
	})

	shouldPass(func(tb testing.TB) {
		test.EqualFunc(tb, 42, 42, func(got, want int) bool { return true })
	})
}

func TestFail(t *testing.T) {
	shouldFail := func(fn func(tb testing.TB)) {
		t.Helper()
		buf := &bytes.Buffer{}
		tb := &TB{out: buf}

		if tb.failed {
			t.Fatal("Initial failed state should be false")
		}

		// Call our test function
		fn(tb)

		if !tb.failed {
			t.Fatal("Should have failed")
		}

		if buf.String() == "" {
			t.Fatal("Should have written on failure")
		}
	}

	shouldFail(func(tb testing.TB) { test.Equal(tb, "something", "else") })
	shouldFail(func(tb testing.TB) { test.Equal(tb, 42, 27) })
	shouldFail(func(tb testing.TB) { test.Equal(tb, true, false) })
	shouldFail(func(tb testing.TB) { test.Equal(tb, 3.14, 8.96) })

	shouldFail(func(tb testing.TB) {
		test.EqualFunc(tb, "something", "different", func(got, want string) bool { return false })
	})

	shouldFail(func(tb testing.TB) {
		test.EqualFunc(tb, 42, 127, func(got, want int) bool { return false })
	})
}
