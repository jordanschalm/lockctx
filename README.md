# Lockctx

Lockctx is a tool for organizing modules which use locks.
It can be useful when:
- the module uses a relatively large number of locks
- the module is large
- the module spans multiple layers, where locks are acquired in one layer and expected to have been acquired in another

## Benefits

Lockctx is used in large modules where some functions acquire locks and some functions require locks.

- Lock-requiring functions can enforce that a lock they require to have been acquired was acquired by the caller.
- Lock-acquiring functions release all held locks together, so using defer statements is easy and safe
- Policies can guarantee deadlock-free operation

## Usage Rules

There are some usage requirements which must be satisfied for Lockctx to provide these benefits:
- Lock-requiring functions must check their `lockctx.Proof` holds the expected lock and exit otherwise
- `lockctx.Context` instances must not be shared between goroutines
