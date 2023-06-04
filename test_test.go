package test_test

import (
	"testing"

	"github.com/FollowTheProcess/test"
)

func TestHello(t *testing.T) {
	got := test.Hello()
	want := "Hello test"

	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}
}
