package lockctx_test

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/jordanschalm/lockctx"
	"github.com/jordanschalm/lockctx/internal/assert"
)

func TestNoPolicy(t *testing.T) {
	lockIDs := lockIDsFixture(5)

	t.Run("can acquire in any order", func(t *testing.T) {
		mgr := lockctx.NewManager(lockIDs, lockctx.NoPolicy)
		ctx := mgr.NewContext()
		rand.Shuffle(len(lockIDs), func(i, j int) {
			lockIDs[i], lockIDs[j] = lockIDs[j], lockIDs[i]
		})
		for _, id := range lockIDs {
			err := ctx.AcquireLock(id)
			assert.NoError(t, err)
		}
	})
}

func TestStringOrderPolicy(t *testing.T) {
	lockIDs := lockIDsFixture(5)
	slices.Sort(lockIDs)
	t.Run("can acquire in string order", func(t *testing.T) {
		mgr := lockctx.NewManager(lockIDs, lockctx.StringOrderPolicy)
		ctx := mgr.NewContext()
		for _, id := range lockIDs {
			err := ctx.AcquireLock(id)
			assert.NoError(t, err)
		}
	})
	t.Run("can acquire in string order with skipped indexes", func(t *testing.T) {
		mgr := lockctx.NewManager(lockIDs, lockctx.StringOrderPolicy)
		ctx := mgr.NewContext()
		for i, id := range lockIDs {
			if i%2 == 0 {
				continue // skip every second lock
			}
			err := ctx.AcquireLock(id)
			assert.NoError(t, err)
		}
	})
	t.Run("cannot acquire out of order", func(t *testing.T) {
		mgr := lockctx.NewManager(lockIDs, lockctx.StringOrderPolicy)
		ctx := mgr.NewContext()
		// first acquire a random lock (except the first one)
		lock1Index := rand.IntN(len(lockIDs)-1) + 1
		lock1 := lockIDs[lock1Index]
		err := ctx.AcquireLock(lock1)
		assert.NoError(t, err)
		// then acquire a random lock that sorts before lock1
		lock2Index := rand.IntN(lock1Index)
		lock2 := lockIDs[lock2Index]
		err = ctx.AcquireLock(lock2)
		assert.ErrorIs(t, err, lockctx.ErrPolicyViolation)
	})
}

func TestDAGPolicy(t *testing.T) {
	lockIDs := lockIDsFixture(5)
	t.Run("empty dag", func(t *testing.T) {
		mgr := lockctx.NewManager(lockIDs, lockctx.NewDAGPolicyBuilder().Build())

		t.Run("can acquire any individual lock", func(t *testing.T) {
			for _, id := range lockIDs {
				ctx := mgr.NewContext()
				err := ctx.AcquireLock(id)
				assert.NoError(t, err)
				ctx.Release()
			}
		})
		t.Run("cannot acquire any two locks", func(t *testing.T) {
			for _, id1 := range lockIDs {
				ctx := mgr.NewContext()
				err := ctx.AcquireLock(id1)
				assert.NoError(t, err)
				for _, id2 := range lockIDs {
					err = ctx.AcquireLock(id2)
					assert.ErrorIs(t, err, lockctx.ErrPolicyViolation)
				}
			}
		})
	})

	t.Run("linear dag", func(t *testing.T) {
		policy := lockctx.NewDAGPolicyBuilder().
			Add(lockIDs[0], lockIDs[1]).
			Add(lockIDs[1], lockIDs[2]).
			Add(lockIDs[2], lockIDs[3]).
			Add(lockIDs[3], lockIDs[4]).
			Build()
		mgr := lockctx.NewManager(lockIDs, policy)

		t.Run("can acquire any individual lock", func(t *testing.T) {
			for _, id := range lockIDs {
				ctx := mgr.NewContext()
				err := ctx.AcquireLock(id)
				assert.NoError(t, err)
				ctx.Release()
			}
		})
		t.Run("can only acquire locks in dag order", func(t *testing.T) {
			for i, id := range lockIDs {
				ctx := mgr.NewContext()
				err := ctx.AcquireLock(id)
				assert.NoError(t, err)
				for j, id2 := range lockIDs {
					if j > i {
						err = ctx.AcquireLock(id2)
						assert.NoError(t, err)
					} else {
						err = ctx.AcquireLock(id2)
						assert.ErrorIs(t, err, lockctx.ErrPolicyViolation)
					}
				}
				ctx.Release()
			}
		})
	})
}
