# Architecture

## Overview
The service is a single-binary file storage designed for use in a corporate network.
It is optimized for correctness and predictable behavior rather than scalability or cloud-native deployment.

The primary client is an internal system (1C).

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
- idempotent upsert when an explicit file ID is provided
- deciding whether file data should be rewritten
- image processing (resize and format selection based on request and configuration)
- access decisions based on file metadata (public / private)

Storage never decides how data should be handled.

---

## Storage

Filesystem-based storage is used.

Key properties:
- deterministic directory layout
- atomic writes (tmp file + rename)
- filesystem-level locking to ensure consistency
- metadata stored alongside binary data
- no database or external dependencies

Storage does not interpret metadata and does not enforce access rules.

---

## Metadata model

File metadata is split conceptually:
- system fields (size, hash, format, image dimensions, public flag)
- user-defined metadata (opaque, not interpreted by the service)

System fields may affect business behavior.
User metadata never does.

---

## Access model

Access control is enforced at two levels:
- endpoint-level checks in HTTP handlers
- file-level access decisions in business logic

Public files may be accessed without authorization.
Private files require read authorization.
All write operations always require authorization.

Public visibility is an explicit business property of a file.

---

## Image handling

Images may be resized so that the longest side does not exceed the configured limit.

The output image format is determined by:
- service configuration, and
- request parameters when an explicit format is requested

Image re-encoding is performed only when required by resizing or format change.
Non-image files are never modified.

---

## Non-goals

The following features are intentionally not implemented:
- CDN integration
- object storage (S3 / MinIO)
- streaming or range requests
- garbage collection of orphaned files
- container orchestration (Docker / Kubernetes)

These decisions are deliberate to keep the service simple and predictable.
