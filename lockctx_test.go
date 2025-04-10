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

func TestHoldsLock(t *testing.T) {
	ids := lockIDsFixture(5)
	mgr := lockctx.NewManager(ids, lockctx.NoPolicy)

	t.Run("at construction should hold no lock", func(t *testing.T) {
		ctx := mgr.NewContext()
		defer ctx.Release()
		for _, id := range ids {
			assert.False(t, ctx.HoldsLock(id))
		}
	})
	t.Run("holding a lock", func(t *testing.T) {
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

	})
	t.Run("after release", func(t *testing.T) {

	})
}

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
