# Lock Context

An idea for synchronization of database operations, inspired by Flow's transition from Badger to Pebble.

Badger has snapshot isolation, and all our low-level storage functions were written under the assumption that
they would be executed in the context of an isolated transaction. Pebble has only read-committed isolation.

Cadence has the notion of capabilities, which are non-copiable resources that represent authorization to
perform some action.

To safely migrate our Badger code to Pebble, we conceptually want something similar. 
There is a set of database operations that now need to be synchronized at the application level.
We would like to:
- not rewrite all our database code
- clearly communicate which operations require such synchronization
- have strong and comprehensible guarantees of exclusivity for relevant database operations
