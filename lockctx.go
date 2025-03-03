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
	// CanAcquire returns true if a goroutine already holding the given locks is
	// allowed to also acquire the next lock.
	//
	// Implementations must be safe for concurrent use by multiple goroutines.
	// Implementations must be non-blocking.
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
	HoldsLock(lockID string) bool

	// Release releases all currently held locks and permanently marks this Context as "used".
	// This method is non-blocking.
	//
	// Panics if Release has ever been called on this Context.
	Release()
}

type manager struct {
	policy Policy
	locks  map[string]*sync.Mutex
}

func NewManager(lockIDs []string, policy Policy) Manager {
	mgr := &manager{
		policy: policy,
		locks:  make(map[string]*sync.Mutex, len(lockIDs)),
	}
	for _, lockID := range lockIDs {
		mgr.locks[lockID] = new(sync.Mutex)
	}
	return mgr
}

func (m *manager) NewContext() Context {
	return &context{
		mgr:  m,
		used: false,
	}
}

type context struct {
	mgr     *manager
	holding []string
	used    bool
}

func (ctx *context) AcquireLock(lockID string) error {
	if ctx.used {
		panic("lockctx: context has been released")
	}
	if !ctx.mgr.policy.CanAcquire(ctx.holding, lockID) {
		return ErrPolicyViolation
	}
	ctx.mgr.locks[lockID].Lock()
	ctx.holding = append(ctx.holding, lockID)
	return nil
}

func (ctx *context) HoldsLock(lockID string) bool {
	if ctx.used {
		return false
	}
	for _, heldLock := range ctx.holding {
		if heldLock == lockID {
			return true
		}
	}
	return false
}

func (ctx *context) Release() {
	if ctx.used {
		panic("lockctx: context has been released")
	}
	for _, lockID := range ctx.holding {
		ctx.mgr.locks[lockID].Unlock()
	}
	ctx.used = true
}

type statelessPolicy func([]string, string) bool

func (policy statelessPolicy) CanAcquire(holding []string, next string) bool {
	return policy(holding, next)
}

// StringOrderPolicy enforces that locks are acquired in lexicographic sort order.
// Guarantees no deadlock.
var StringOrderPolicy statelessPolicy = func(holding []string, next string) bool {
	if len(holding) == 0 {
		return true
	}
	last := holding[len(holding)-1]
	// next lock ID must sort after last acquired lock
	return last < next
}

// NoPolicy enforces no constraints on lock ordering.
var NoPolicy statelessPolicy = func(holding []string, next string) bool {
	return true
}
