package util

import (
	"errors"
	"fmt"

	"golang.org/x/exp/constraints"
)

// Tern returns a if cond is true, otherwise b.
func Tern[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

// Map applies f to each element of s and returns the result.
func Map[T1, T2 any](s []T1, f func(v T1) T2) []T2 {
	if s == nil {
		return nil
	}

	res := make([]T2, len(s))
	for i := range s {
		res[i] = f(s[i])
	}
	return res
}

// Must1 panics if err is not nil, otherwise returns v.
func Must1[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// Must1B panics if ok is false, otherwise returns v.
func Must1B[T any](v T, ok bool) T {
	if !ok {
		panic("expected ok")
	}
	return v
}

// Assert panics if cond is false.
// Optional arguments are treated as a format string and arguments for fmt.Sprintf.
func Assert(cond bool, msgAndArgs ...any) {
	if !cond {
		msg := "assertion failed"
		if len(msgAndArgs) > 0 {
			if format, ok := msgAndArgs[0].(string); ok {
				msg += ": " + fmt.Sprintf(format, msgAndArgs[1:]...)
			} else {
				msg += ": " + fmt.Sprint(msgAndArgs...)
			}
		}
		panic(errors.New(msg))
	}
}

// Min returns the smaller of a and b.
func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Max returns the larger of a and b.
func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}
