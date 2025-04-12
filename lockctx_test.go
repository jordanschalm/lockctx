package lockctx_test

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/jordanschalm/lockctx"
	"github.com/jordanschalm/lockctx/internal/assert"
)

func holdsAll(ctx lockctx.Context, lockIds []string) bool {
	for _, id := range lockIds {
		if !ctx.HoldsLock(id) {
			return false
		}
	}
	return true
}

func holdsAny(ctx lockctx.Context, lockIds []string) bool {
	for _, id := range lockIds {
		if ctx.HoldsLock(id) {
			return true
		}
	}
	return false
}

func lockIDsFixture(n int) []string {
	ids := make([]string, n)
	for i := range ids {
		ids[i] = fmt.Sprintf("%d", i)
	}
	return ids
}

func TestErrors(t *testing.T) {
	t.Run("ErrPolicyViolation", func(t *testing.T) {
		err := fmt.Errorf("something bad happened: %w", lockctx.ErrPolicyViolation)
		assert.ErrorIs(t, err, lockctx.ErrPolicyViolation)
	})
	t.Run("UnknownLockError", func(t *testing.T) {
		err := lockctx.NewUnknownLockError("lockid")
		assert.True(t, lockctx.IsUnknownLockError(err))
		wrapped := fmt.Errorf("something bad happened: %w", err)
		assert.True(t, lockctx.IsUnknownLockError(wrapped))
	})
}

func TestAcquireLock(t *testing.T) {
	ids := lockIDsFixture(2)
	existentID := ids[0]
	nonexistentID := ids[1]
	t.Run("can acquire existent lock", func(t *testing.T) {
		t.Parallel()

		mgr := lockctx.NewManager([]string{existentID}, lockctx.NoPolicy)
		ctx := mgr.NewContext()
		err := ctx.AcquireLock(existentID)
		assert.NoError(t, err)
	})
	t.Run("cannot acquire nonexistent lock", func(t *testing.T) {
		t.Parallel()

		mgr := lockctx.NewManager([]string{existentID}, lockctx.NoPolicy)
		ctx := mgr.NewContext()
		err := ctx.AcquireLock(nonexistentID)
		assert.True(t, lockctx.IsUnknownLockError(err))
	})
	t.Run("if policy allows, can acquire same lock twice, resulting in deadlock", func(t *testing.T) {
		t.Parallel()

		mgr := lockctx.NewManager([]string{existentID}, lockctx.NoPolicy)
		ctx := mgr.NewContext()
		err := ctx.AcquireLock(existentID)
		assert.NoError(t, err)
		assert.DoesNotReturnAfter(t, time.Millisecond*10, func() {
			_ = ctx.AcquireLock(existentID) // blocks forever
		})
	})
}

// TestHoldsLock tests the HoldsLock function under various circumstances.
func TestHoldsLock(t *testing.T) {
	ids := lockIDsFixture(5)

	t.Run("at construction should hold no lock", func(t *testing.T) {
		mgr := lockctx.NewManager(ids, lockctx.NoPolicy)
		ctx := mgr.NewContext()
		defer ctx.Release()
		for _, id := range ids {
			assert.False(t, ctx.HoldsLock(id))
		}
	})
	t.Run("holding a lock", func(t *testing.T) {
		mgr := lockctx.NewManager(ids, lockctx.NoPolicy)
		ctx := mgr.NewContext()
		defer ctx.Release()

		toAcquire := ids[rand.IntN(len(ids))]
		err := ctx.AcquireLock(toAcquire)
		assert.NoError(t, err)

		for _, id := range ids {
			isHolding := ctx.HoldsLock(id)
			assert.True(t, isHolding == (toAcquire == id))
		}
	})
	t.Run("holding multiple locks", func(t *testing.T) {
		mgr := lockctx.NewManager(ids, lockctx.NoPolicy)
		ctx := mgr.NewContext()
		defer ctx.Release()

		toAcquire := ids[:rand.IntN(len(ids))]
		acquired := make(map[string]bool, len(toAcquire))
		for _, id := range toAcquire {
			acquired[id] = true
			err := ctx.AcquireLock(id)
			assert.NoError(t, err)
		}

		for _, id := range ids {
			isHolding := ctx.HoldsLock(id)
			assert.True(t, isHolding == acquired[id])
		}
	})
	t.Run("after release", func(t *testing.T) {
		mgr := lockctx.NewManager(ids, lockctx.NoPolicy)
		ctx := mgr.NewContext()

		toAcquire := ids[rand.IntN(len(ids))]
		err := ctx.AcquireLock(toAcquire)
		assert.NoError(t, err)
		ctx.Release()

		for _, id := range ids {
			assert.False(t, ctx.HoldsLock(id))
		}
	})
}

// TestConcurrentContexts tests concurrent goroutines using Context to acquire multiple locks.
// Each goroutine attempts to acquire all locks in order and the test verifies that at each
// step the Context reports the correct set of locks are currently held.
func TestConcurrentContexts(t *testing.T) {
	lockIDs := lockIDsFixture(5)
	mgr := lockctx.NewManager(lockIDs, lockctx.NoPolicy)
	wg := new(sync.WaitGroup)
	wg.Add(len(lockIDs))
	for i := 0; i < len(lockIDs); i++ {
		go func() {
			defer wg.Done()

			ctx := mgr.NewContext()
			defer ctx.Release()

			assert.False(t, holdsAny(ctx, lockIDs))
			for j := 0; j < len(lockIDs); j++ {
				err := ctx.AcquireLock(lockIDs[j])
				assert.NoError(t, err)
				assert.True(t, holdsAll(ctx, lockIDs[:j+1]))
				assert.False(t, holdsAny(ctx, lockIDs[j+1:]))
			}
		}()
	}
	wg.Wait()
}
