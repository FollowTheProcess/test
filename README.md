# test

[![License](https://img.shields.io/github/license/FollowTheProcess/test)](https://github.com/FollowTheProcess/test)
[![Go Reference](https://pkg.go.dev/badge/go.followtheprocess.codes/test.svg)](https://pkg.go.dev/go.followtheprocess.codes/test)
[![Go Report Card](https://goreportcard.com/badge/github.com/FollowTheProcess/test)](https://goreportcard.com/report/github.com/FollowTheProcess/test)
[![GitHub](https://img.shields.io/github/v/release/FollowTheProcess/test?logo=github&sort=semver)](https://github.com/FollowTheProcess/test)
[![CI](https://github.com/FollowTheProcess/test/workflows/CI/badge.svg)](https://github.com/FollowTheProcess/test/actions?query=workflow%3ACI)
[![codecov](https://codecov.io/gh/FollowTheProcess/test/branch/main/graph/badge.svg)](https://codecov.io/gh/FollowTheProcess/test)

***A lightweight test helper package*** 🧪

## Project Description

`test` is my take on a handy, lightweight Go test helper package. Inspired by [matryer/is], [earthboundkid/be] and others.

It provides a lightweight, but useful, extension to the std lib testing package with a friendlier and hopefully intuitive API. You definitely don't need it,
but might find it useful anyway 🙂

## Installation

```shell
go get go.followtheprocess.codes/test@latest
```

## Usage

`test` is as easy as...

```go
func TestSomething(t *testing.T) {
    test.Equal(t, "hello", "hello") // Obviously fine
    test.Equal(t, "hello", "there") // Fails

    test.NotEqual(t, 42, 27) // Passes, these are not equal
    test.NotEqual(t, 42, 42) // Fails

    test.NearlyEqual(t, 3.0000000001, 3.0) // Look, floats handled easily!

    err := doSomething()
    test.Ok(t, err) // Fails if err != nil
    test.Err(t, err) // Fails if err == nil

    test.True(t, true) // Passes
    test.False(t, true) // Fails
}
```

### Add Additional Context

`test` provides a number of options to decorate your test log with useful context:

```go
func TestDetail(t *testing.T) {
    test.Equal(t, "apples", "oranges", test.Title("Fruit scramble!"), test.Context("Apples are not oranges!"))
}
```

Will get you an error log in the test that looks like this...

```plaintext
--- FAIL: TestDemo (0.00s)
    test_test.go:501:
        Fruit scramble!
        ---------------

        Got:    apples
        Wanted: oranges

        (Apples are not oranges!)

FAIL
```

### Non Comparable Types

`test` uses generics under the hood for most of the comparison, which is great, but what if your types don't satisfy `comparable`. We also provide
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

    // Can also use any function here
    test.EqualFunc(t, a, b, slices.Equal) // Also passes :)

    test.EqualFunc(t, a, c, slices.Equal) // Fails
}
```

You can also use this same pattern for custom user defined types, structs etc.

### Table Driven Tests

Table driven tests are great! But when you test errors too it can get a bit awkward, you have to do the `if (err != nil) != tt.wantErr` thing and I personally
*always* have to do the boolean logic in my head to make sure I got that right. Enter `test.WantErr`:

```go
func TestTableThings(t *testing.T) {
    tests := []struct {
        name    string
        want    int
        wantErr bool
    }{
        {
            name:    "no error",
            want:    4,
            wantErr: false,
        },
        {
            name:    "yes error",
            want:    4,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := SomeFunction()

            test.WantErr(t, err, tt.wantErr)
            test.Equal(t, got, tt.want)
        })
    }
}
```

Which is basically semantically equivalent to:

```go
func TestTableThings(t *testing.T) {
    tests := []struct {
        name    string
        want    int
        wantErr bool
    }{
        {
            name:    "no error",
            want:    4,
            wantErr: false,
        },
        {
            name:    "yes error",
            want:    4,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := SomeFunction()

            if tt.wantErr {
                test.Err(t, err)
            } else {
                test.Ok(t, err)
            }
            test.Equal(t, got, tt.want)
        })
    }
}
```

### Capturing Stdout and Stderr

We've all been there, trying to test a function that prints but doesn't accept an `io.Writer` as a destination 🙄.

That's where `test.CaptureOutput` comes in!

```go
func TestOutput(t *testing.T) {
    // Function that prints to stdout and stderr, but imagine this is defined somewhere else
    // maybe a 3rd party library that you don't control, it just prints and you can't tell it where
    fn := func() error {
        fmt.Fprintln(os.Stdout, "hello stdout")
        fmt.Fprintln(os.Stderr, "hello stderr")

        return nil
    }

    // CaptureOutput to the rescue!
    stdout, stderr := test.CaptureOutput(t, fn)

    test.Equal(t, stdout, "hello stdout\n")
    test.Equal(t, stderr, "hello stderr\n")
}
```

Under the hood `CaptureOutput` temporarily captures both streams, copies the data to a buffer and returns the output back to you, before cleaning everything back up again.

### See Also

- [FollowTheProcess/snapshot] for golden file/snapshot testing 📸

### Credits

This package was created with [copier] and the [FollowTheProcess/go_copier] project template.

[copier]: https://copier.readthedocs.io/en/stable/
[FollowTheProcess/go_copier]: https://github.com/FollowTheProcess/go_copier
[matryer/is]: https://github.com/matryer/is
[earthboundkid/be]: https://github.com/earthboundkid/be
[FollowTheProcess/snapshot]: https://github.com/FollowTheProcess/snapshot
