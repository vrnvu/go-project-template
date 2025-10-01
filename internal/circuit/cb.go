package circuit

import (
	"errors"
	"testing"
)

type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

type Result int

const (
	Rejected Result = iota
	Failed
	Succeeded
)

func asserts(condition bool) {
	if !condition {
		panic("assertion failed")
	}
}

func Ok(t *testing.T) func() error {
	t.Helper()
	return func() error {
		return nil
	}
}

func Error(t *testing.T) func() error {
	t.Helper()
	return func() error {
		return errors.New("error")
	}
}
