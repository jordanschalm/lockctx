package example

import (
	"fmt"

	"github.com/jordanschalm/lockctx"
)

const LockIDX = "X"
const LockIDY = "Y"

// LowLevelOperationX is an operation that must be called while a specific lock is held.
func LowLevelOperationX(ctx lockctx.Context) error {
	if !ctx.HoldsLock(LockIDX) {
		return fmt.Errorf("caller must hold lock %s", LockIDX)
	}
	// do operation X
	return nil
}

// LowLevelOperationY is an operation that must be called while a specific lock is held.
func LowLevelOperationY(ctx lockctx.Context) error {
	if !ctx.HoldsLock(LockIDY) {
		return fmt.Errorf("caller must hold lock %s", LockIDY)
	}
	// do operation Y
	return nil
}

// HighLevelComponent is a component which performs complex operations requiring multiple locks.
type HighLevelComponent struct {
	mgr lockctx.Manager
}

func (c *HighLevelComponent) DoSomething() error {
	// Create a context to acquire all necessary locks
	ctx := c.mgr.NewContext()
	// Deferring Release guarantees that all our held locks will be released when exiting this scope
	defer ctx.Release()
	err := ctx.AcquireLock(LockIDX)
	if err != nil {
		return fmt.Errorf("could not acquire lock %s", LockIDX)
	}
	err = ctx.AcquireLock(LockIDY)
	if err != nil {
		return fmt.Errorf("could not acquire lock %s", LockIDY)
	}

	// Perform lock-requiring operations
	// If we later add an operation requiring lock Z that we have not acquired, the function will always error.
	err = LowLevelOperationX(ctx)
	if err != nil {
		return err
	}
	err = LowLevelOperationY(ctx)
	if err != nil {
		return err
	}

	return nil
}
