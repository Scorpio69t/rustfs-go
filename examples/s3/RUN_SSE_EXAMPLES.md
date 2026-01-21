# SSE (Server-Side Encryption) 示例运行指南

本目录包含 3 个服务端加密示例程序，演示如何使用 RustFS Go SDK 进行加密对象操作。

## 示例文件

1. **encryption-sse-s3-put.go** - SSE-S3 加密示例（服务器管理密钥）
````markdown
# SSE (Server-Side Encryption) Examples — Run Guide

This folder contains three server-side encryption example programs that demonstrate how to use the RustFS Go SDK to perform encrypted object operations.

## Example files

1. **encryption-sse-s3-put.go** — SSE‑S3 example (server-managed keys)
2. **encryption-sse-c-put.go** — SSE‑C example (customer-provided key)
3. **encryption-bucket-config.go** — Bucket-level default encryption configuration

## Prerequisites

You need a running S3‑compatible server such as:
- MinIO
- RustFS
- AWS S3
- Any other S3 compatible storage

### Quick start MinIO (Docker)

```bash
docker run -d \
  -p 9000:9000 \
  -p 9001:9001 \
  --name minio \
  -e "MINIO_ROOT_USER=minioadmin" \
  -e "MINIO_ROOT_PASSWORD=minioadmin" \
  minio/minio server /data --console-address ":9001"
```

## Running the examples

### Method 1 — use default configuration

```bash
# 1. SSE-S3 example
go run encryption-sse-s3-put.go

# 2. SSE-C example
go run encryption-sse-c-put.go

# 3. Bucket encryption configuration example
go run encryption-bucket-config.go
```

Default config used by examples:
- Endpoint: `localhost:9000`
- Access Key: `minioadmin`
- Secret Key: `minioadmin`

### Method 2 — use custom configuration

```bash
# Windows PowerShell
$env:S3_ENDPOINT="your-endpoint:9000"
$env:S3_ACCESS_KEY="your-access-key"
$env:S3_SECRET_KEY="your-secret-key"
go run encryption-sse-s3-put.go

# Linux/macOS Bash
export S3_ENDPOINT="your-endpoint:9000"
export S3_ACCESS_KEY="your-access-key"
export S3_SECRET_KEY="your-secret-key"
go run encryption-sse-s3-put.go
```

### Method 3 — build then run

```bash
# Build
go build encryption-sse-s3-put.go
go build encryption-sse-c-put.go
go build encryption-bucket-config.go

# Run
./encryption-sse-s3-put      # Linux/macOS
.\encryption-sse-s3-put.exe  # Windows
```

## Example notes

### 1. encryption-sse-s3-put.go

Shows the simplest server-side encryption mode:

```go
// Enable SSE-S3 on upload
uploadInfo, err := objectSvc.Put(ctx, bucketName, objectName, reader, size,
    object.WithSSES3(), // server-managed keys
)

// Server decrypts automatically on download
downloadReader, info, err := objectSvc.Get(ctx, bucketName, objectName)
```

Key points:
- ✅ No key management required
- ✅ Server performs encryption/decryption automatically
- ✅ Suitable for most use cases

### 2. encryption-sse-c-put.go

Shows customer-provided key encryption (SSE‑C):

```go
// Generate 256-bit key
key := make([]byte, 32)
rand.Read(key)

// Provide key on upload
uploadInfo, err := objectSvc.Put(ctx, bucketName, objectName, reader, size,
    object.WithSSEC(key), // client-provided key
)

// Must provide same key on download
downloadReader, info, err := objectSvc.Get(ctx, bucketName, objectName,
    object.WithGetSSEC(key),
)
```

Key points:
- ✅ Full control of the encryption key
- ✅ Key is not stored on the server
- ⚠️ If the key is lost, data cannot be decrypted
- ⚠️ The same key must be provided on every access

### 3. encryption-bucket-config.go

Shows bucket default encryption configuration:

```go
// Set bucket default encryption to SSE-S3
config := sse.NewConfiguration()
err = bucketSvc.SetEncryption(ctx, bucketName, config)

// Get bucket encryption configuration
config, err := bucketSvc.GetEncryption(ctx, bucketName)

// Delete bucket encryption configuration
err = bucketSvc.DeleteEncryption(ctx, bucketName)
```

Key points:
- ✅ New objects are automatically encrypted
- ✅ No need to pass encryption options on every upload
- ✅ Supports SSE‑S3 and SSE‑KMS

## Expected output (examples)

### SSE-S3 example output

```
✓ Bucket created: test-encryption
✓ Uploaded with SSE-S3 encryption
  Bucket: test-encryption
  Object: encrypted-object.txt
  ETag: "d41d8cd98f00b204e9800998ecf8427e"
  Size: 45 bytes

✓ Download successful (server-decrypted)
  Content: This is sensitive data encrypted by SSE-S3

✓ Object metadata
  Size: 45 bytes
  ETag: "d41d8cd98f00b204e9800998ecf8427e"
  LastModified: 2026-01-21 09:30:00 +0000 UTC

Notes:
  - SSE-S3 uses server-managed keys
  - Server encrypts/decrypts automatically
  - Client does not need to manage keys
```

### SSE-C example output

```
✓ Bucket created: test-encryption-customer
✓ Generated 256-bit encryption key
  Key (hex prefix): abc123...

✓ Uploaded with SSE-C encryption
  Bucket: test-encryption-customer
  Object: customer-encrypted-object.txt
  ETag: "d41d8cd98f00b204e9800998ecf8427e"

✓ Download with correct key succeeded
  Content: This is sensitive data encrypted with a customer key

✗ Download with wrong key failed (expected)
  Error: Access Denied

Notes:
  - SSE-C uses a 256-bit key supplied by the client
  - Key is not stored on the server
  - Keep the key safe; losing it means permanent data loss
```

### Bucket encryption configuration example output

```
✓ Bucket created: test-bucket-encryption

✓ Set bucket default encryption (SSE-S3)
✓ Retrieved bucket encryption config
  Algorithm: AES256

✓ Uploaded object (auto-encrypted)
  Object: auto-encrypted-object.txt

✓ Updated bucket encryption to SSE-KMS
  KMS Key ID: arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012

✓ Deleted bucket encryption configuration
✓ Verified deletion: no encryption config present (expected)

Notes:
  - Bucket default encryption applies to all new objects
  - Supports SSE-S3 and SSE-KMS
  - Existing objects are unaffected
```

## Troubleshooting

### Connection failure

```
Error: InternalError: Bad Gateway
```

Fixes:
1. Ensure your S3 server is running
2. Check the endpoint and port
3. Verify network connectivity

### Authentication failure

```
Error: Access Denied
```

Fixes:
1. Verify Access Key and Secret Key
2. Check IAM permissions (for AWS S3)
3. Verify bucket policy

### Build errors

```
Error: cannot find package
```

Fixes:
```bash
# Update dependencies
go mod tidy
go mod download
```

## References

- [AWS S3 Server-Side Encryption](https://docs.aws.amazon.com/AmazonS3/latest/userguide/serv-side-encryption.html)
- [MinIO Encryption Guide](https://min.io/docs/minio/linux/administration/server-side-encryption.html)
- [RustFS Go SDK Documentation](../../README.md)

## Support

If you need help:
1. Check [CONTRIBUTING.md](../../CONTRIBUTING.md)
2. Open a GitHub Issue: https://github.com/Scorpio69t/rustfs-go/issues
3. Consult project documentation

````
