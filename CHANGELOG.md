# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.0.0] - 2025-01-XX

### Added

#### Core Features
- ✅ **Full S3 API compatibility** – supports all standard S3 operations
- ✅ **Modular design** – bucket and object services with clearer APIs
- ✅ **AWS Signature V4/V2** – full signing support, including streaming
- ✅ **Health checks** – built-in health checks with retries
- ✅ **HTTP tracing** – request performance tracing and debugging support
- ✅ **Transport tuning** – configurable HTTP transport with pooling, timeouts, TLS

#### Bucket Operations
- Create/delete buckets (region, object locking, force delete, etc.)
- List buckets
- Check if a bucket exists
- Get bucket location

#### Object Operations
- Upload/download objects (metadata, tags, storage class, etc.)
- Get object info and metadata
- Delete objects
- List objects (prefix, recursive, max-keys filters)
- Copy objects (metadata replacement, conditional copy, etc.)

#### Multipart Upload
- Initiate multipart uploads
- Upload parts
- Complete multipart uploads
- Abort multipart uploads

#### Advanced
- Streaming signature support (AWS Signature V4 chunked upload)
- Location cache optimization (fewer GetBucketLocation calls)
- Smart retry policy (network errors and selected HTTP status codes)
- Automatic path-style selection for IP endpoints

### Technical Details

#### New Modules
- `internal/signer/` – AWS signing (V4/V2/streaming)
- `internal/transport/` – HTTP transport and tracing
- `internal/core/` – core executor and health checks
- `bucket/` – bucket service
- `object/` – object service

#### Test Coverage
- Unit test coverage > 60%
- 150+ total test cases
- Core functionality fully tested

#### Examples
- Bucket operations (`examples/rustfs/bucketops.go`)
- Object operations (`examples/rustfs/objectops.go`)
- Multipart upload (`examples/rustfs/multipart.go`)
- Health check (`examples/rustfs/health.go`)
- HTTP tracing (`examples/rustfs/trace.go`)

### Changed
- New modular API design
- Option-function pattern for flexible configuration
- Improved error handling and type definitions

### Dependencies
- Go 1.25+
- github.com/google/uuid v1.6.0
- golang.org/x/net v0.25.0

### Documentation
- English-first README with Chinese companion
- GoDoc comments for public APIs
- Detailed usage examples
- OpenSpec specification documents

---

## Roadmap

### v1.1.3 (Planned)
- [x] Presigned URL support (GET/PUT, header overrides, SSE signing)
- [x] Object tagging (set/get/delete, tagging count parsing)
- [x] Bucket policy management (set/get/delete)
- [x] Lifecycle management (set/get/delete)
- [x] Server-side encryption options (SSE-S3/SSE-C for upload/download)
- [x] File helpers (fput/fget for path-based upload/download)

### v1.1.4 (Planned)
- [x] Object versioning
- [x] Cross-region replication
- [x] Event notifications
- [x] Access logging

---

## Contributing

Issues and Pull Requests are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

Apache License 2.0 – see [LICENSE](LICENSE) for details.
