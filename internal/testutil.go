package internal

import (
	"errors"
	"testing"
)

func AssertErrorIs(t *testing.T, err, target error) {
	if err == nil {
		t.Fail()
	}
	if !errors.Is(err, target) {
		t.Fail()
	}
}

func AssertNoError(t *testing.T, err error) {
	if err != nil {
		t.Fail()
	}
}

func AssertTrue(t *testing.T, b bool) {
	if !b {
		t.Fail()
	}
}

func AssertFalse(t *testing.T, b bool) {
	AssertTrue(t, !b)
}
