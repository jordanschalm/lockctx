package lockctx_test

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"testing"

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

func TestAcquireLock(t *testing.T) {
	// acquire a lock twice? -> currently deadlocks, should error instead
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
