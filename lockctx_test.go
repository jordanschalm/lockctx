package lockctx_test

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"
	"testing"

	"github.com/jordanschalm/lockctx"
)

func assertErrorIs(t *testing.T, err, target error) {
	if err == nil {
		t.Fail()
	}
	if !errors.Is(err, target) {
		t.Fail()
	}
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Fail()
	}
}

func assertTrue(t *testing.T, b bool) {
	if !b {
		t.Fail()
	}
}

func assertFalse(t *testing.T, b bool) {
	assertTrue(t, !b)
}

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
			assertFalse(t, ctx.HoldsLock(id))
		}
	})
	t.Run("holding a lock", func(t *testing.T) {
		ctx := mgr.NewContext()
		defer ctx.Release()

		toAcquire := ids[rand.IntN(len(ids))]
		err := ctx.AcquireLock(toAcquire)
		assertNoError(t, err)

		for _, id := range ids {
			isHolding := ctx.HoldsLock(id)
			assertTrue(t, isHolding == (toAcquire == id))
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

			assertFalse(t, holdsAny(ctx, lockIDs))
			for j := 0; j < len(lockIDs); j++ {
				err := ctx.AcquireLock(lockIDs[j])
				assertNoError(t, err)
				assertTrue(t, holdsAll(ctx, lockIDs[:j+1]))
				assertFalse(t, holdsAny(ctx, lockIDs[j+1:]))
			}
		}()
	}
	wg.Wait()
}
