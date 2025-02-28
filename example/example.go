package example

import (
	"github.com/jordanschalm/lockctx"
)

const LockIDX = "X"

type DB interface {
	Get(key string) (string, error)
	Put(key, value string) error
}

// LowLevelOperation is an operation that must be called while a specific lock is held.
func LowLevelOperationX(proof lockctx.Proof) {
	if !proof.Valid(LockIDX) {
		panic("caller must acquire lock")
	}
}

type HighLevelComponent struct {
	mu lockctx.Lock
}

func (c *HighLevelComponent) DoThing() {
	c.mu.Lock()

}

func example() {
	mgr := lockctx.NewManager()
	lock, err := mgr.New(LockIDX)
	if err != nil {
		panic(err)
	}

	c := HighLevelComponent{lock}
}
