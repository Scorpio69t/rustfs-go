## 1. Proposal
- [x] 1.1 Confirm scope of missing capabilities
- [x] 1.2 Review existing API/spec coverage

## 2. Presigned HEAD
- [x] 2.1 Add PresignHead API and options
- [x] 2.2 Update signer/request metadata for HEAD presign
- [x] 2.3 Add tests and example

## 3. GetObject Response Header Overrides
- [x] 3.1 Add Get options for response header overrides
- [x] 3.2 Wire query params into GET requests
- [x] 3.3 Add tests and example

## 4. Multipart Listing APIs
- [x] 4.1 Add ListMultipartUploads API
- [x] 4.2 Add ListObjectParts API
- [x] 4.3 Add pagination options
- [x] 4.4 Add tests and example

## 5. Checksum Mode
- [x] 5.1 Add checksum mode options for Put/Multipart
- [x] 5.2 Wire checksum headers/query params
- [x] 5.3 Add tests and example

## 6. S3 Accelerate
- [x] 6.1 Add accelerate endpoint option
- [x] 6.2 Apply to object operations
- [x] 6.3 Add tests and example

## 7. Client-Side Encryption (CSE)
- [x] 7.1 Introduce CSE helpers (encrypt/decrypt streams)
- [x] 7.2 Add Put/Get integration options
- [x] 7.3 Add tests and example

## 8. Docs & Examples
- [x] 8.1 Update README.md and README.zh.md
- [x] 8.2 Update examples/s3/README.md
- [x] 8.3 Add or update changelog

## 9. Validation
- [x] 9.1 go test ./...
- [x] 9.2 golangci-lint run ./...
