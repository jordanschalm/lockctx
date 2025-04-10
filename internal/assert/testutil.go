package assert

import (
	"errors"
	"testing"
)

func ErrorIs(t *testing.T, err, target error) {
	if err == nil {
		t.Fail()
	}
	if !errors.Is(err, target) {
		t.Fail()
	}
}

func NoError(t *testing.T, err error) {
	if err != nil {
		t.Fail()
	}
}

func True(t *testing.T, b bool) {
	if !b {
		t.Fail()
	}
}

func False(t *testing.T, b bool) {
	True(t, !b)
}
