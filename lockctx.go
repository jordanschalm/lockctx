package lockctx

import (
	"errors"
	"sync"
)

var ErrPolicyViolation = errors.New("policy violation")

// Manager controls access to a set of locks.
// The set of locks and Policy (if any) is defined at construction time and is constant for the lifecycle of the Manager.
type Manager interface {
	// NewContext returns a new Context which is able to acquire locks managed by this Manager.
	NewContext() Context
}

// Policy defines whether a goroutine is allowed acquire a new lock based on locks it already holds.
// Policies exist to prevent deadlock by defining a canonical ordering for the set of locks in a Manager.
type Policy interface {
	// CanAcquire returns true if a goroutine holding the given locks are allowed to subsequently
	// acquire the next lock.
	CanAcquire(holding []string, next string) bool
}

// Context represents a goroutine's access to one or more locks managed by a Manager.
// A new Context must be created every time a goroutine first acquires a lock.
// A Context is not safe for concurrent access by multiple goroutines.
type Context interface {
	// AcquireLock acquires the lock with the given ID, unless doing so violates the configured Policy.
	// This function will block if the lock is held by another goroutine.
	//
	// Returns ErrPolicyViolation if acquiring the lock would violate the configured Policy.
	// Panics if no lock with the given ID exists.
	// Panics if Release has ever been called on this Context.
	AcquireLock(lockID string) error

	// HoldsLock returns true if this goroutine currently holds the lock with the given ID.
	// This method is non-blocking.
	//
	// Panics if no lock with the given ID exists.
	// Panics if Release has ever been called on this Context.
	HoldsLock(lockID string) bool

	// Release releases all currently held locks and permanently marks this Context as "used".
	// This method is non-blocking.
	//
	// Panics if Release has ever been called on this Context.
	Release()
}

type manager struct {
	policy Policy
	locks  map[string]sync.Mutex
}
