# Handlers
## GET /get
    200 - Ok
    404 - NotFound
    422 - FormatUnsupported

## POST /upload
    201 - Created
    413 - TooBig
    422 - FormatUnsupported

## DELETE /delete
    204 - No Content
    (delete is idempotent)

# Logic
## File
File consist of 
1. ID
2. Binary data
3. Metadata (JSON, any fields you want)

## UPLOAD
- needs Binary data and Matadata
- ID is optional
- if ID is provided and file exists - file is replaced
- if ID is provided and file does not exist - file is created
- if ID is not provided - file with new id is created
- always returns ID

## GET
supports two modes:
1. raw binary (file only)
2. structure with binary data and metadata
format and size can be requested

# Middlewares

## 1.Recover
catch panic
answer 500

## 2.Request ID
check if X-Request-ID header exists
- if exists forward it to request_id
- if not, creates new request_id for query
request_id stored in context, return in responce header

## 3.Access Log
log request/response metadata: method, path, status, duration, request_id

## 4.Timeout
set timeout in Context

## 5.Size limit
check request body size 

## 6.Authorization
access control through Token
identity derived from Token stored in context

