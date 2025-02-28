package lockctx

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrDuplicate = errors.New("duplicate lock")

var defaultManager = NewManager()

type Manager interface {
	New(id string) (Lock, error)
}

type managerBuilder struct {
}

func (m *managerBuilder) AddLock(id string) *managerBuilder {
	return nil
}

func (m *managerBuilder) Build() (Manager, error) {
	return nil, nil
}

func NewManager() Manager {
	return &manager{
		created: make(map[string]struct{}),
	}
}

type manager struct {
	mu      sync.Mutex
	created map[string]struct{}
}

type Lock interface {
	Acquire(id string) error
	Release()
}

type LockProof interface {
	Valid(id string) bool
}


func (m *manager) New(id string) (Lock, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.created[id]; ok {
		return nil, ErrDuplicate
	}
	m.created[id] = struct{}{}
	return &lock{
		mu:   sync.Mutex{},
		id:   id,
		held: atomic.Bool{},
	}, nil
}

type Lock interface {
	sync.Locker
	ID() string
	Proof() Proof
}

func NewLock(id string) (Lock, error) {
	return defaultManager.New(id)
}

type lock struct {
	mu   sync.Mutex
	id   string
	held atomic.Bool
}

func (l *lock) Lock() {
	l.mu.Lock()
	l.held.Store(true)
}

func (l *lock) Unlock() {
	l.held.Store(false)
	l.mu.Unlock()
}

func (l *lock) ID() string {
	return l.id
}

func (l *lock) valid(id string) bool {
	return l.held.Load() && l.id == id
}

func (l *lock) Proof() Proof {
	return &proof{lock: l}
}

// Proof represents the capability to operate on a shared resource protected by a lock.
type Proof interface {
	// Valid returns true if the proof remains valid
	Valid(id string) bool
}

type proof struct {
	lock *lock
}

func (c *proof) Valid(id string) bool {
	return c.lock.valid(id)
}
