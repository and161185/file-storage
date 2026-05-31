# Architecture

## Overview
The service is a single-binary file storage system designed for single-node deployments. It focuses on correctness and predictable behavior, with a simple operational model and no external dependencies.

---

## Layered design

The service is structured into three layers:

- HTTP layer — request handling, routing, authorization checks at endpoint level
- Business layer — all domain logic and decision making
- Storage layer — dumb persistence, no business logic

Each layer depends only on the layer below.

---

## Business logic

The business layer is responsible for:
- idempotent delete
- idempotent upsert when a file ID is provided
- deciding whether data should be rewritten
- image processing (resize and format selection based on request and configuration)
- access decisions based on file metadata (public / private)

Storage never decides how data should be handled.

---

## Storage model

Filesystem-based storage is used.

Each file has a unique ID and is stored in a directory structure based on it.
The storage keeps content and metadata in two slots (A/B) and uses an active slot file to indicate the active slot for each.

For each file ID, the storage maintains:
- active slot file — active slot information for content and metadata
- a content file per slot (A/B)
- a metadata file per slot (A/B)
- lock file for per-ID synchronization

For each file, content and metadata may point to different active slots. The active slot file stores active slots for content and metadata independently.
The active slot state is stored in the active slot file, which is updated atomically using the same tmp + rename approach as regular file writes.

If the active slot file is missing, the storage falls back to a non-versioned layout.

---

## Write path

Writes are performed under a per-ID lock.

The storage writes new content and metadata to the inactive slots first.
Only after all new files are fully written and synced does it atomically replace the active slot file.

This ensures that readers continue to observe the previous active state until the new active slot state is fully ready.

Old versions are not removed during the write path and are cleaned up later by the garbage collector.

---

## Read path

Reads use the active slot defined in the active slot file.

The storage reads the active slot and opens the corresponding content and metadata files.

If the active slot file is missing, the storage uses a non-versioned layout.

---

## Garbage collection and recovery

The garbage collector scans the storage tree and removes files that are not part of the active slot state.

If the active slot file cannot be read, the garbage collector reconstructs the active slot state from the most recent files based on modification time.

The garbage collector uses the same per-ID lock as HTTP operations, so cleanup and request processing are synchronized through a single locking mechanism.

---

## Consistency model

The storage provides the following guarantees:

- readers resolve both content and metadata through the active slot file
- slot switch is atomic
- readers observe either the old or the new state, never a partial update
- writes for the same file ID are serialized using a per-ID lock file
- inactive data may remain temporarily and are removed asynchronously by the garbage collector
- readers never access files outside of the active slot state

A crash during write does not expose partially written data to readers, because the active slot file is replaced only after all new files are ready.

A crash during write may leave incomplete files on disk, but they are never referenced by the active state and therefore are not visible to readers.

---

## Metadata model

File metadata is split conceptually:
- system fields: size, hashes, format, image dimensions, public flag, timestamps
- user-defined metadata: flat custom fields provided by clients

System fields may affect business behavior.
User-defined metadata is stored and returned, but is not interpreted by the service.
Allowed user metadata value types:
- string
- boolean
- number

Nested objects, arrays and null values are not supported.

---

## Access model

Access control is enforced at two levels:
- HTTP middleware resolves request-level access flags from static tokens
- business logic performs file-level access decisions

Public files may be accessed without authorization through the content endpoint.
Private files require read access.
Metadata access always requires read access.
Upload and delete operations always require write access.

Public visibility is an explicit business property of a file.

---

## Image handling

Images may be resized so that the longest side does not exceed the configured limit.

The output image format is determined by:
- service configuration, and
- request parameters when a format is explicitly specified

Image re-encoding is performed only when required by resizing or format change.
Non-image files are never modified.

---

## Non-goals

The following features are intentionally not implemented:
- distributed storage or distributed locking
- object storage (S3 / MinIO)
- streaming or range requests
- multi-node write coordination
- container orchestration (Docker / Kubernetes)

These decisions are deliberate to keep the service simple and predictable.
