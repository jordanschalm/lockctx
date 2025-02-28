package lockctx_test

import (
	"errors"
	"slices"
	"sort"
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

func TestLockCreation(t *testing.T) {

	t.Run("default manager should produce exactly one lock with a given id", func(t *testing.T) {
		_, err := lockctx.NewLock("a")
		assertNoError(t, err)
		_, err = lockctx.NewLock("a")
		assertErrorIs(t, err, lockctx.ErrDuplicate)
	})

	t.Run("distinct managers should each produce exactly one lock with a given id", func(t *testing.T) {
		mgr1 := lockctx.NewManager()
		mgr2 := lockctx.NewManager()
		_, err := mgr1.New("a")
		assertNoError(t, err)
		_, err = mgr1.New("a")
		assertErrorIs(t, err, lockctx.ErrDuplicate)
		_, err = mgr2.New("a")
		assertNoError(t, err)
		_, err = mgr2.New("a")
		assertErrorIs(t, err, lockctx.ErrDuplicate)
	})
}

func TestConcurrentLockCreation(t *testing.T) {
	mgr := lockctx.NewManager()
	ids := []string{"a", "b", "c", "d", "e"}
	locks := make(chan lockctx.Lock, len(ids))
	wg := new(sync.WaitGroup)
	wg.Add(len(ids))
	for i := 0; i < len(ids); i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < len(ids); j++ {
				lock, err := mgr.New(ids[j])
				if err != nil {
					assertErrorIs(t, err, lockctx.ErrDuplicate)
					continue
				}
				locks <- lock
			}
		}()
	}

	wg.Wait()
	close(locks)
	createdLockIDs := make([]string, 0, len(locks))
	for lock := range locks {
		createdLockIDs = append(createdLockIDs, lock.ID())
	}
	sort.Strings(createdLockIDs)
	if !slices.Equal(ids, createdLockIDs) {
		t.Fail()
	}
}
