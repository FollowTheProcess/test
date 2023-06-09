# test

[![License](https://img.shields.io/github/license/FollowTheProcess/test)](https://github.com/FollowTheProcess/test)
[![Go Reference](https://pkg.go.dev/badge/github.com/FollowTheProcess/test.svg)](https://pkg.go.dev/github.com/FollowTheProcess/test)
[![Go Report Card](https://goreportcard.com/badge/github.com/FollowTheProcess/test)](https://goreportcard.com/report/github.com/FollowTheProcess/test)
[![GitHub](https://img.shields.io/github/v/release/FollowTheProcess/test?logo=github&sort=semver)](https://github.com/FollowTheProcess/test)
[![CI](https://github.com/FollowTheProcess/test/workflows/CI/badge.svg)](https://github.com/FollowTheProcess/test/actions?query=workflow%3ACI)
[![codecov](https://codecov.io/gh/FollowTheProcess/test/branch/main/graph/badge.svg)](https://codecov.io/gh/FollowTheProcess/test)

***A lightweight test helper package*** 🧪

## Project Description

`test` is my take on a handy, lightweight Go test helper package. Inspired by [matryer/is], [carlmjohnson/be] and others.

It provides a lightweight, but useful, extension to the std lib testing package with a friendlier and hopefully intuitive API. You definitely don't need it,
but might find it useful anyway 🙂

## Installation

```shell
go get github.com/FollowTheProcess/test@latest
```

## Usage

`test` is as easy as...

```go
func TestSomething(t *testing.T) {
    test.Equal(t, "hello", "hello") // Obviously fine
    test.Equal(t, "hello", "there") // Fails

    test.NotEqual(t, 42, 27) // Passes, these are not equal
    test.NotEqual(t, 42, 42) // Fails

    err := doSomething()
    test.Ok(t, err) // Fails if err != nil

    test.True(t, true) // Passes
    test.False(t, true) // Fails

    // Just like the good old reflect.DeepEqual, but with a nicer format
    test.DeepEqual(t, []string{"hello"}, []string{"world"}) // Fails
}
```

### Non Comparable Types

`test` uses Go 1.18+ generics under the hood for most of the comparison, which is great, but what if your types don't satisfy `comparable`. We also provide
`test.EqualFunc` and `test.NotEqualFunc` for those exact situations!

These allow you to pass in a custom comparator function for your type, if your comparator function returns true, the types are considered equal.

```go
func TestNonComparableTypes(t *testing.T) {
    // Slices do not satisfy comparable
    a := []string{"hello", "there"}
    b := []string{"hello", "there"}
    c := []string{"general", "kenobi"}

    // Custom function, returns true if things should be considered equal
    sliceEqual := func(a, b, []string) { return true } // Cheating

    test.EqualFunc(t, a, b, sliceEqual) // Passes

    // Can also use e.g. golang.org/x/exp/slices
    test.EqualFunc(t, a, b, slices.Equal[string]) // Also passes :)

    test.EqualFunc(t, a, c, slices.Equal[string]) // Fails
}
```

You can also use this same pattern for custom user defined types, structs etc.

### Rich Comparison

Large structs or long slices can often be difficult to compare using `reflect.DeepEqual`, you have to scan for the difference yourself. `test` provides a
`test.Diff` function that produces a rich text diff for you on failure:

```go
func TestDiff(t *testing.T) {
    // Pretent these are very long, or are large structs
    a := []string{"hello", "world"}
    b := []string{"hello", "there"}

    test.Diff(t, a, b)
}
```

Will give you:

```plain
--- FAIL: TestDiff (0.00s)
    main_test.go:14: Mismatch (-want, +got):
          []string{
                "hello",
        -       "there",
        +       "world",
          }
```

### Credits

This package was created with [copier] and the [FollowTheProcess/go_copier] project template.

[copier]: https://copier.readthedocs.io/en/stable/
[FollowTheProcess/go_copier]: https://github.com/FollowTheProcess/go_copier
[matryer/is]: https://github.com/matryer/is
[carlmjohnson/be]: https://github.com/carlmjohnson/be
