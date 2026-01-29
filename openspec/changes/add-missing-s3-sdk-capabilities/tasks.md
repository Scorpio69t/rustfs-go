## 1. Proposal
- [ ] 1.1 Confirm scope of missing capabilities
- [ ] 1.2 Review existing API/spec coverage

## 2. Presigned HEAD
- [ ] 2.1 Add PresignHead API and options
- [ ] 2.2 Update signer/request metadata for HEAD presign
- [ ] 2.3 Add tests and example

## 3. GetObject Response Header Overrides
- [ ] 3.1 Add Get options for response header overrides
- [ ] 3.2 Wire query params into GET requests
- [ ] 3.3 Add tests and example

## 4. Multipart Listing APIs
- [ ] 4.1 Add ListMultipartUploads API
- [ ] 4.2 Add ListObjectParts API
- [ ] 4.3 Add pagination options
- [ ] 4.4 Add tests and example

## 5. Checksum Mode
- [ ] 5.1 Add checksum mode options for Put/Multipart
- [ ] 5.2 Wire checksum headers/query params
- [ ] 5.3 Add tests and example

## 6. S3 Accelerate
- [ ] 6.1 Add accelerate endpoint option
- [ ] 6.2 Apply to object operations
- [ ] 6.3 Add tests and example

## 7. Client-Side Encryption (CSE)
- [ ] 7.1 Introduce CSE helpers (encrypt/decrypt streams)
- [ ] 7.2 Add Put/Get integration options
- [ ] 7.3 Add tests and example

## 8. Docs & Examples
- [ ] 8.1 Update README.md and README.zh.md
- [ ] 8.2 Update examples/s3/README.md
- [ ] 8.3 Add or update changelog

## 9. Validation
- [ ] 9.1 go test ./...
- [ ] 9.2 golangci-lint run ./...
