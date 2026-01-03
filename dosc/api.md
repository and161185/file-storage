# Handlers
## GET /info
    200 - Ok
    404 - NotFound
    422 - FormatUnsupported

## GET /content
    200 - Ok
    404 - NotFound
    422 - FormatUnsupported

## POST /upload
    201 Created
    - new file created (no id OR id not found)

    200 OK
    - file existed:
    - hash different → file replaced
    - hash same → binary preserved, metadata updated

    400 Bad Request
    - invalid JSON / invalid base64

    413 Payload Too Large

    415 Unsupported Media Type
    - isImage=true but file is not recognized as image

    422 Unprocessable Entity
    - provided hash doesn’t match calculated one
    - no data to upload

## DELETE /delete
    204 - No Content
    (delete is idempotent)

    400 - Bad Request
    - ID incorrect format 

# Logic
## File
File consist of 
1. ID 36 symbols
2. Binary data
3. Metadata (JSON, any fields you want)
4. Hash
5. IsImage

## UPLOAD
    requires: Binary, Metadata, Hash

    ID optional
    ID absent OR not found - create (201)
    ID exists:
       - hash differs - replace file (200)
       - hash equals - skip storing, update     metadata only (200)

    hash:
       - ensures integrity
       - avoids pointless rewrites

    isImage optional:
       - omitted - service detects
       - true - service validates image format, converts to storage format
       - invalid image - 415

## CONTENT
    raw binary (file only), format and size can be requested

## INFO
    structure with metadata without binary data


# Middlewares

## 1.Request ID
check if X-Request-ID header exists
- if exists forward it to request_id
- if not, creates new request_id for query
request_id stored in context, return in responce header

put logger into context

## 2.Recover
catch panic
answer 500

## 3.Access Log
log request/response metadata: method, path, status, duration, request_id

## 4.Timeout
set timeout in Context

## 5.Size limit
check request body size 

## 6.Authorization
access control through Token
identity derived from Token stored in context

