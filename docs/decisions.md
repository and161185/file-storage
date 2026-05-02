# Architectural decisions

This document describes key design decisions made in the file storage service.

---

## Filesystem storage instead of database

**Decision**

The service uses a local filesystem for storing file data and metadata.

**Why**

- the number of files is relatively small (tens of thousands)
- the primary use case is storing one file per business entity
- filesystem storage requires no additional infrastructure or maintenance
- the system is designed for moderate workloads where a simple solution is sufficient
- the storage layer is abstracted, allowing alternative backends without changing business logic

**Alternatives considered**

- relational database
- object storage (S3 / MinIO)

These were rejected because they add complexity without real benefit.

**Trade-offs**

As a result of using a filesystem:

- no built-in indexing or querying capabilities
- limited scalability
- no replication or redundancy (data protection is handled externally)

---

## Two-slot storage with active slot file

**Decision**

Each file is stored using two slots (`A` and `B`) and an active slot file `[id].versions`
that defines the active slot for content and metadata.

**Why**

- initially, direct overwrite caused inconsistencies between data and metadata
- the two-slot design provides a simple and reliable way to guarantee consistency
- a single active slot file acts as the source of truth for reads

**Alternatives considered**

- in-place overwrite with locking

These approaches were rejected due to higher complexity and less predictable behavior.

**Trade-offs**

- additional storage overhead due to inactive data
- requires background cleanup (garbage collector)

---

## Per-ID locking with flock

**Decision**

Write operations for a given file ID are synchronized using a filesystem lock (`flock`).

**Why**

- file ID represents a complete unit of data (content + metadata)
- write operations on a file must be serialized
- filesystem locking works across processes and components (including GC)

**Alternatives considered**

- in-memory mutexes
- distributed locking

These approaches were rejected because in-memory mutexes do not work across processes,
and distributed locking introduces additional infrastructure and complexity
that are not required for a single-node system.

**Trade-offs**

- limited to single-node execution
- dependent on filesystem semantics

---

## Single-node storage model

**Decision**

The service operates as a single-node system with local filesystem storage.

**Why**

- current data size and load are low
- available storage capacity is sufficient
- simplicity and predictability are prioritized over scalability

**Alternatives considered**

- distributed storage
- replicated systems

These approaches were rejected as premature optimization.

**Trade-offs**

- no horizontal scaling
- no redundancy
- limited fault tolerance

---

## No streaming and no range requests

**Decision**

The service does not support streaming uploads/downloads or range requests.

**Why**

- the primary use case is storing images, which are relatively small
- files are served as full images to external systems
- streaming would add unnecessary complexity

**Alternatives considered**

- streaming uploads/downloads
- partial content support

These were rejected because they are not required for the current use cases.

**Trade-offs**

- inefficient for large file transfers
- no support for partial reads

---

## Background garbage collector

**Decision**

Cleanup of obsolete and incomplete files is performed by a background garbage collector.

**Why**

- write path should remain simple and fast
- cleanup is not required for correctness during request handling
- garbage collector can also reconstruct active state if needed

**Alternatives considered**

- synchronous cleanup during writes

This approach was rejected due to increased latency and complexity.

**Trade-offs**

- cleanup is performed asynchronously
- storage may temporarily contain obsolete data
- the garbage collector adds background I/O load