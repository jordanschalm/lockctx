package assert

import (
	"errors"
	"testing"
	"time"
)

func ErrorIs(t *testing.T, err, target error) {
	if err == nil {
		t.Logf("expected error %s but got nil", target.Error())
		t.Fail()
	} else if !errors.Is(err, target) {
		t.Logf("expected error type %T but got %s", target, err.Error())
		t.Fail()
	}
}

func NoError(t *testing.T, err error) {
	if err != nil {
		t.Logf("expected no error but got: %s", err.Error())
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

func DoesNotReturnAfter(t *testing.T, d time.Duration, f func()) {
	returned := make(chan struct{})
	go func() {
		f()
		close(returned)
	}()
	select {
	case <-time.After(d):
		return
	case <-returned:
		t.Logf("function returned within %s", d)
		t.Fail()
	}
}
