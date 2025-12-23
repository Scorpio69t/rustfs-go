# 🚀 RustFS Go SDK

<div align="center">

[![Go Reference](https://pkg.go.dev/badge/github.com/Scorpio69t/rustfs-go.svg)](https://pkg.go.dev/github.com/Scorpio69t/rustfs-go)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8?logo=go)](https://go.dev/)
[![GitHub stars](https://img.shields.io/github/stars/Scorpio69t/rustfs-go?style=social)](https://github.com/Scorpio69t/rustfs-go)

**A high-performance Go client library for RustFS object storage system**

[English](README.md) | [中文](README.zh.md)

</div>

---

## 📖 Overview

RustFS Go SDK is a comprehensive Go client library for interacting with the RustFS object storage system. It is fully compatible with the S3 API, providing a clean and intuitive interface that supports all standard S3 operations.

### ✨ Features

- ✅ **Full S3 API Compatibility** - Complete support for all S3-compatible operations
- ✅ **Clean API Design** - Intuitive and easy-to-use interface
- ✅ **Comprehensive Operations** - Bucket management, object operations, multipart uploads, and more
- ✅ **Streaming Signature** - AWS Signature V4 streaming support for chunked uploads
- ✅ **Health Check** - Built-in health check with retry mechanism
- ✅ **HTTP Tracing** - Request tracing for performance monitoring and debugging
- ✅ **Error Handling** - Robust error handling and retry mechanisms
- ✅ **Streaming Support** - Efficient streaming upload/download for large files
- ✅ **Production Ready** - Well-tested with comprehensive examples

## 🚀 Installation

```bash
go get github.com/Scorpio69t/rustfs-go
```

## 📚 Quick Start

### Initialize Client

```go
package main

import (
    "context"
    "log"

    "github.com/Scorpio69t/rustfs-go"
    "github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
    // Initialize client
    client, err := rustfs.New("127.0.0.1:9000", &rustfs.Options{
        Credentials: credentials.NewStaticV4("your-access-key", "your-secret-key", ""),
        Secure:      false, // Set to true for HTTPS
    })
    if err != nil {
        log.Fatalln(err)
    }

    ctx := context.Background()
    // Use client for operations...
}
```

### 📦 Bucket Operations

```go
// Obtain the Bucket service
bucketSvc := client.Bucket()

// Create bucket
err := bucketSvc.Create(ctx, "my-bucket",
    bucket.WithRegion("us-east-1"),
    bucket.WithObjectLocking(false),
)

// List all buckets
buckets, err := bucketSvc.List(ctx)
for _, bucket := range buckets {
    fmt.Println(bucket.Name)
}

// Check if bucket exists
exists, err := bucketSvc.Exists(ctx, "my-bucket")

// Get bucket location
location, err := bucketSvc.GetLocation(ctx, "my-bucket")

// Delete bucket
err = bucketSvc.Delete(ctx, "my-bucket")
// Or force delete (RustFS extension, deletes all objects)
err = bucketSvc.Delete(ctx, "my-bucket", bucket.WithForceDelete(true))
```

### 📄 Object Operations

```go
// Obtain the Object service
objectSvc := client.Object()

// Upload object from reader
data := strings.NewReader("Hello, RustFS!")
uploadInfo, err := objectSvc.Put(ctx, "my-bucket", "my-object.txt",
    data, int64(data.Len()),
    object.WithContentType("text/plain"),
    object.WithUserMetadata(map[string]string{
        "author": "rustfs-go",
    }),
    object.WithUserTags(map[string]string{
        "category": "example",
    }),
)

// Download object
reader, objInfo, err := objectSvc.Get(ctx, "my-bucket", "my-object.txt")
defer reader.Close()

buf := make([]byte, 1024)
n, _ := reader.Read(buf)
fmt.Println(string(buf[:n]))

// Download with range
reader, _, err := objectSvc.Get(ctx, "my-bucket", "my-object.txt",
    object.WithGetRange(0, 99), // First 100 bytes
)

// Get object information
objInfo, err := objectSvc.Stat(ctx, "my-bucket", "my-object.txt")

// List objects
objectsCh := objectSvc.List(ctx, "my-bucket")
for obj := range objectsCh {
    if obj.Err != nil {
        log.Println(obj.Err)
        break
    }
    fmt.Println(obj.Key, obj.Size)
}

// Copy object
copyInfo, err := objectSvc.Copy(ctx,
    "my-bucket", "copy.txt",     // destination
    "my-bucket", "my-object.txt", // source
)

// Delete object
err = objectSvc.Delete(ctx, "my-bucket", "my-object.txt")
```

### 🔄 Multipart Upload

```go
// Obtain the Object service and assert to the multipart-capable interface
objectSvc := client.Object()
type MultipartService interface {
    InitiateMultipartUpload(ctx context.Context, bucketName, objectName string,
        opts ...object.PutOption) (string, error)
    UploadPart(ctx context.Context, bucketName, objectName, uploadID string,
        partNumber int, reader io.Reader, partSize int64,
        opts ...object.PutOption) (types.ObjectPart, error)
    CompleteMultipartUpload(ctx context.Context, bucketName, objectName, uploadID string,
        parts []types.ObjectPart, opts ...object.PutOption) (types.UploadInfo, error)
    AbortMultipartUpload(ctx context.Context, bucketName, objectName, uploadID string) error
}
multipartSvc := objectSvc.(MultipartService)

// 1. Initialize multipart upload
uploadID, err := multipartSvc.InitiateMultipartUpload(ctx, "my-bucket", "large-file.txt",
    object.WithContentType("text/plain"),
)

// 2. Upload parts
var parts []types.ObjectPart
part1, err := multipartSvc.UploadPart(ctx, "my-bucket", "large-file.txt",
    uploadID, 1, part1Data, partSize)
parts = append(parts, part1)

part2, err := multipartSvc.UploadPart(ctx, "my-bucket", "large-file.txt",
    uploadID, 2, part2Data, partSize)
parts = append(parts, part2)

// 3. Complete multipart upload
uploadInfo, err := multipartSvc.CompleteMultipartUpload(ctx, "my-bucket",
    "large-file.txt", uploadID, parts)

// 4. Abort multipart upload (if needed)
err = multipartSvc.AbortMultipartUpload(ctx, "my-bucket", "large-file.txt", uploadID)
```

> 📖 **Full example**: see [examples/rustfs/multipart.go](examples/rustfs/multipart.go)

### 🔐 Presigned URLs

```go
// Generate presigned GET URL with response header override
url, signedHeaders, err := client.Object().PresignGet(
    ctx,
    "my-bucket",
    "photo.jpg",
    15*time.Minute,
    url.Values{"response-content-type": []string{"image/jpeg"}},
)

// Generate presigned PUT URL signing SSE-S3 header
putURL, putSignedHeaders, err := client.Object().PresignPut(
    ctx,
    "my-bucket",
    "uploads/photo.jpg",
    15*time.Minute,
    nil,
    object.WithPresignSSES3(),
)
```

> 📖 **Full example**: see [examples/rustfs/presigned.go](examples/rustfs/presigned.go)

### 🏷️ Object Tagging & File Helpers

```go
// Upload from file with tags (add object.WithSSES3() if SSE is enabled on your server)
uploadInfo, err := client.Object().FPut(
    ctx,
    "my-bucket",
    "demo/hello.txt",
    "/path/to/file.txt",
    object.WithUserTags(map[string]string{"env": "dev"}),
)

// Read and delete tags
tags, _ := client.Object().GetTagging(ctx, "my-bucket", uploadInfo.Key)
_ = client.Object().DeleteTagging(ctx, "my-bucket", uploadInfo.Key)
```

> 📖 **Full example**: see [examples/rustfs/object_tagging.go](examples/rustfs/object_tagging.go)

### 📜 Bucket Policy & Lifecycle

```go
policyJSON := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::my-bucket/public/*"]}]}`
_ = client.Bucket().SetPolicy(ctx, "my-bucket", policyJSON)

lifecycleXML := []byte(`<LifecycleConfiguration><Rule><ID>expire-temp</ID><Status>Enabled</Status><Filter><Prefix>temp/</Prefix></Filter><Expiration><Days>30</Days></Expiration></Rule></LifecycleConfiguration>`)
_ = client.Bucket().SetLifecycle(ctx, "my-bucket", lifecycleXML)
```

> 📖 **Full example**: see [examples/rustfs/bucket_policy_lifecycle.go](examples/rustfs/bucket_policy_lifecycle.go)

### 🏥 Health Check

```go
// Basic health check
result := client.HealthCheck(nil)
if result.Healthy {
    fmt.Printf("✅ Service is healthy, response time: %v\n", result.ResponseTime)
} else {
    fmt.Printf("❌ Service is unhealthy: %v\n", result.Error)
}

// Health check with timeout
opts := &core.HealthCheckOptions{
    Timeout: 5 * time.Second,
    Context: context.Background(),
}
result := client.HealthCheck(opts)

// Health check with retries
result := client.HealthCheckWithRetry(opts, 3)
```

> 📖 **Full example**: see [examples/rustfs/health.go](examples/rustfs/health.go)

### 📊 HTTP Request Tracing

```go
import "github.com/Scorpio69t/rustfs-go/internal/transport"

// Build a trace hook
var traceInfo *transport.TraceInfo
hook := func(info transport.TraceInfo) {
    traceCopy := info
    traceInfo = &traceCopy
}

// Create a trace-enabled context
traceCtx := transport.NewTraceContext(ctx, hook)

// Execute a request
bucketSvc := client.Bucket()
exists, err := bucketSvc.Exists(traceCtx, "my-bucket")

// Inspect trace results
if traceInfo != nil {
    fmt.Printf("Connection reused: %v\n", traceInfo.ConnReused)
    fmt.Printf("Total duration: %v\n", traceInfo.TotalDuration())

    // Stage timings
    timings := traceInfo.GetTimings()
    for stage, duration := range timings {
        fmt.Printf("%s: %v\n", stage, duration)
    }
}
```

> 📖 **Full example**: see [examples/rustfs/trace.go](examples/rustfs/trace.go)

## 🔑 Credentials Management

### Static Credentials

```go
creds := credentials.NewStaticV4("access-key", "secret-key", "")
```

### Environment Variables

```go
creds := credentials.NewEnvAWS()
// Reads from environment variables:
// AWS_ACCESS_KEY_ID
// AWS_SECRET_ACCESS_KEY
// AWS_SESSION_TOKEN
```

## ⚙️ Configuration Options

```go
client, err := rustfs.New("rustfs.example.com", &rustfs.Options{
    Creds:        credentials.NewStaticV4("access-key", "secret-key", ""),
    Secure:       true,              // Use HTTPS
    Region:       "us-east-1",       // Region
    BucketLookup: rustfs.BucketLookupDNS, // Bucket lookup style
    Transport:    nil,               // Custom HTTP Transport
    MaxRetries:   10,                // Max retry attempts
})
```

## 📝 Examples

More example code can be found in the [examples/rustfs](examples/rustfs/) directory:

- [Bucket Operations](examples/rustfs/bucketops.go)
- [Object Operations](examples/rustfs/objectops.go)
- [Multipart Upload](examples/rustfs/multipart.go)
- [Health Check](examples/rustfs/health.go)
- [HTTP Tracing](examples/rustfs/trace.go)

### Run the examples

```bash
cd examples/rustfs

# Run examples
go run -tags example bucketops.go
go run -tags example objectops.go
go run -tags example multipart.go
go run -tags example health.go
go run -tags example trace.go
```

> **💡 Tip**: Before running examples, make sure:
> - A RustFS server is running (default `127.0.0.1:9000`)
> - Access keys in the sample code are updated
> - The buckets referenced in the samples exist

## 📖 API Documentation

Full API documentation is available at: https://pkg.go.dev/github.com/Scorpio69t/rustfs-go

## 📄 License

This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 🔗 References

- [AWS S3 API Documentation](https://docs.aws.amazon.com/AmazonS3/latest/API/Welcome.html) - API specification
- [AWS Signature Version 4](https://docs.aws.amazon.com/general/latest/gr/signature-version-4.html) - Signature algorithm

## 💬 Support

For issues or suggestions, please submit an [Issue](https://github.com/Scorpio69t/rustfs-go/issues).

---

<div align="center">

**Made with ❤️ by the RustFS Go SDK community**

[⬆ Back to Top](#-rustfs-go-sdk)

</div>
