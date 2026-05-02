# API

## GET /files/{id}/info

Returns file metadata without file content.

**Responses:**

* 200 OK — metadata returned
* 400 Bad Request — invalid ID format
* 404 Not Found — file does not exist

---

## GET /files/{id}/content

Returns file content.

Optional image transformation parameters may be provided.

### Query parameters

* `width` — optional, target width
* `height` — optional, target height
* `format` — optional, output image format

**Responses:**

* 200 OK — file content returned
* 400 Bad Request — invalid ID format or parameters
* 404 Not Found — file does not exist
* 422 Unprocessable Entity — stored file cannot be processed (e.g. corrupted image)

---

## POST /files/upload

Creates a new file or updates an existing one.

**Behavior:**

* if ID is not provided → a new file is created
* if ID exists:

  * different hash → file is replaced
  * same hash → binary is preserved, metadata is updated

**Request body:**

```json
{
  "id": "optional-file-id",
  "data": "<base64-encoded-binary>",
  "hash": "<sha256-hash>",
  "isImage": true,
  "metadata": {
    "any": "custom fields"
  }
}
```

**Responses:**

* 200 OK — file created or updated
* 400 Bad Request — invalid JSON or base64
* 413 Payload Too Large — request exceeds size limit
* 415 Unsupported Media Type — `isImage=true`, but file is not a valid image
* 422 Unprocessable Entity — hash mismatch 

---

## DELETE /files/{id}/delete

Deletes a file.

The operation is idempotent.

**Responses:**

* 204 No Content — file deleted or did not exist
* 400 Bad Request — invalid ID format

---

# File model

A file consists of:

* ID (36-character string)
* binary data
* metadata (JSON, user-defined)
* hash
* image flag (`isImage`)

---

# Upload rules

* `id` is optional
* upload is idempotent only when an explicit ID is provided

### Hash

* ensures data integrity
* prevents unnecessary rewrites

### Image handling

* `isImage` is optional:

  * omitted — service detects automatically
  * `true` — service validates image and converts it to configured format
  * invalid image → 415 error

---

# Middleware

## Request ID

* reads `X-Request-ID` header if provided
* generates a new ID if missing
* propagates request ID via context
* returns request ID in response headers

## Recovery

* catches panics
* returns HTTP 500

## Access log

* logs request metadata: method, path, status, duration, request ID

## Timeout

* sets request timeout via context

## Size limit

* validates request body size

## Authorization

* validates access token
* extracts identity into context
* does not enforce business-level access rules

Final access decisions are made in business logic.

---

# Examples

## Get file content (public)

```bash
curl -X GET \
  "http://localhost:8080/files/{id}/content"
```

---

## Get file content (authorized)

```bash
curl -X GET \
  "http://localhost:8080/files/{id}/content" \
  -H "Authorization: Bearer <token>"
```

---

## Upload file

```bash
curl -X POST \
  "http://localhost:8080/files/upload" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "optional-file-id",
    "data": "<base64-encoded-binary>",
    "hash": "<sha256-hash>",
    "isImage": true,
    "metadata": {
      "any": "custom fields"
    }
  }'
```

---

## Delete file

```bash
curl -X DELETE \
  "http://localhost:8080/files/{id}/delete" \
  -H "Authorization: Bearer <token>"
```
