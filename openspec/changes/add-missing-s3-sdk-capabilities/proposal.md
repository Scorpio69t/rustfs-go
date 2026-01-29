# Change: Add missing S3 SDK capabilities

## Why
The old S3 example set exposed several S3 features that are still missing or not surfaced in the current SDK. This gap blocks full parity for users migrating from the old SDK and makes some S3 workflows (presign HEAD, checksum uploads, multipart listings, S3 Accelerate, client-side encryption, response header overrides) impossible or awkward to implement.

## What Changes
- Add presigned HEAD URL generation.
- Add response header override support for GetObject (response-content-type, disposition, etc.).
- Add multipart listing APIs (list uploads, list parts) and matching options.
- Add checksum mode support for uploads and copy.
- Add S3 Accelerate endpoint support for object uploads/downloads.
- Add client-side encryption (CSE) helpers and object ops.
- Add documentation and examples for the new APIs.

## Impact
- Impacted specs: api
- Impacted code:
  - object/ (presign, get options, multipart, checksum)
  - internal/core/ (request metadata and signing for new query/header options)
  - pkg/sse or new pkg/cse (client-side encryption helpers)
  - examples/s3 (new samples)
  - README.md / README.zh.md
