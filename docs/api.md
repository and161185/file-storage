# API

## Authorization

The service uses static read and write tokens from `app.security`.

Authorization header format:

```http
Authorization: Bearer <token>
```

Resolved access is stored in request context as access flags:

* `Read`
* `Write`

The middleware only resolves request-level access flags. Final business access decisions are made in the business layer.

---

## GET /files/{id}/info

Returns file metadata without file content.

Requires read authorization.

### Path parameters

* `id` — 36-character file ID

### Response body

```json
{
  "id": "file-id",
  "hash_source": "sha256-of-original-input",
  "hash_stored": "sha256-of-stored-content",
  "public": false,
  "file_size": 12345,
  "is_image": true,
  "format": "jpeg",
  "width": 1920,
  "height": 1080,
  "metadata": {
    "title": "example",
    "published": true,
    "score": 4.5
  },
  "created_at": "2026-05-03T10:00:00Z",
  "updated_at": "2026-05-03T10:00:00Z"
}
```

### Responses

* `200 OK` — metadata returned
* `400 Bad Request` — invalid ID format
* `403 Forbidden` — missing or insufficient read access
* `404 Not Found` — file does not exist
* `500 Internal Server Error` — internal error

---

## GET /files/{id}/content

Returns file content.

Public files are available without authorization. Private files require read authorization.

Optional image transformation parameters may be provided.

### Path parameters

* `id` — 36-character file ID

### Query parameters

* `width` — optional target width, from `10` to `10000`
* `height` — optional target height, from `10` to `10000`
* `format` — optional output image format

### Responses

* `200 OK` — file content returned
* `400 Bad Request` — invalid ID format or invalid query parameters
* `403 Forbidden` — private file requested without read access
* `404 Not Found` — file does not exist
* `415 Unsupported Media Type` — unsupported requested output format
* `422 Unprocessable Entity` — stored file cannot be processed
* `500 Internal Server Error` — internal error

---

## POST /files/upload

Creates a new file or updates an existing one.

Requires write authorization.

### Behavior

* if `id` is not provided, a new file ID is generated
* if `id` is provided, upload is idempotent for the same ID and data
* if source hash is unchanged, binary data may be preserved and metadata updated
* if source hash is changed, file content is replaced
* if `public` is omitted, the file is private by default

### Request body

```json
{
  "id": "optional-file-id",
  "data": "<base64-encoded-binary>",
  "hash": "<sha256-hash>",
  "public": false,
  "is_image": true,
  "metadata": {
    "title": "example",
    "published": true,
    "score": 4.5
  }
}
```

### Metadata constraints

Allowed metadata value types:

* string
* boolean
* number

Unsupported metadata value types:

* object
* array
* null

### Response body

```json
{
  "id": "file-id"
}
```

### Responses

* `200 OK` — file created or updated
* `400 Bad Request` — invalid JSON, invalid base64, unknown field, invalid ID
* `403 Forbidden` — missing or insufficient write access
* `413 Payload Too Large` — request exceeds configured size limit
* `415 Unsupported Media Type` — unsupported image type or output format
* `422 Unprocessable Entity` — hash mismatch, invalid image, unsupported metadata value
* `500 Internal Server Error` — internal error

---

## DELETE /files/{id}/delete

Deletes a file.

Requires write authorization.

The operation is idempotent.

### Path parameters

* `id` — 36-character file ID

### Responses

* `204 No Content` — file deleted or did not exist
* `400 Bad Request` — invalid ID format
* `403 Forbidden` — missing or insufficient write access
* `500 Internal Server Error` — internal error

---

## GET /files/metrics

Returns Prometheus metrics.

This endpoint is intended for operational monitoring.

Current behavior:

* no authorization middleware
* handler timeout is applied
* response format is Prometheus text exposition format

### Responses

* `200 OK` — metrics returned

---

# File model

A file consists of:

* ID
* binary data
* system metadata
* user metadata
* source hash
* stored hash
* public flag
* image flag
* optional image format and dimensions

System metadata may affect service behavior. User metadata is stored and returned, but is not interpreted by business logic.

---

# Image handling

The service determines whether a file is an image based on:

* explicit `is_image` flag in the request, or
* automatic detection from binary data if the flag is omitted

If the file is an image:

* it is validated
* it may be resized
* it may be re-encoded to configured storage format
* requested content format may be applied during content retrieval

Non-image files are stored as-is.

---

# Middleware

## Request ID

* reads `X-Request-ID` header if provided
* generates a new ID if missing
* propagates request ID via context
* returns request ID in response headers

## Recovery

* catches panics
* returns `500 Internal Server Error`

## Access log

* logs method, path, status, duration and request ID

## Metrics

* records HTTP request count
* records request duration
* records in-flight requests

## Rate limiter

* limits request rate using token bucket settings

## Timeout

* sets request timeout via context

## Size limit

* limits request body size

## Authorization

* validates bearer token
* resolves read/write access flags
* stores access flags in context
* does not make final file-level access decisions

---

# Examples

## Get public file content

```bash
curl -X GET \
  "http://localhost:8080/files/{id}/content"
```

## Get private file content

```bash
curl -X GET \
  "http://localhost:8080/files/{id}/content" \
  -H "Authorization: Bearer <read-token>"
```

## Get file metadata

```bash
curl -X GET \
  "http://localhost:8080/files/{id}/info" \
  -H "Authorization: Bearer <read-token>"
```

## Upload file

```bash
curl -X POST \
  "http://localhost:8080/files/upload" \
  -H "Authorization: Bearer <write-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "optional-file-id",
    "data": "<base64-encoded-binary>",
    "hash": "<sha256-hash>",
    "public": false,
    "is_image": true,
    "metadata": {
      "title": "example",
      "published": true,
      "score": 4.5
    }
  }'
```

## Delete file

```bash
curl -X DELETE \
  "http://localhost:8080/files/{id}/delete" \
  -H "Authorization: Bearer <write-token>"
```

## Metrics

```bash
curl -X GET \
  "http://localhost:8080/files/metrics"
```
