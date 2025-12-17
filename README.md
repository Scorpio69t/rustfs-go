# 🚀 RustFS Go SDK

<div align="center">

[![Go Reference](https://pkg.go.dev/badge/github.com/Scorpio69t/rustfs-go.svg)](https://pkg.go.dev/github.com/Scorpio69t/rustfs-go)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8?logo=go)](https://go.dev/)
[![GitHub stars](https://img.shields.io/github/stars/Scorpio69t/rustfs-go?style=social)](https://github.com/Scorpio69t/rustfs-go)

**A high-performance Go client library for RustFS object storage system**

[English](#english) | [中文](#中文)

</div>

---

<!-- English Section -->
<div id="english"></div>

## 📖 Overview

RustFS Go SDK is a comprehensive Go client library for interacting with RustFS object storage system. It is fully compatible with S3 API, providing a clean and intuitive interface that supports all standard S3 operations.

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
// 获取 Bucket 服务
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
// 获取 Object 服务
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
// 获取 Object 服务并类型断言为支持分片上传的接口
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

> 📖 **完整示例**: 查看 [examples/rustfs/multipart.go](examples/rustfs/multipart.go)

### 🔐 Presigned URLs

> **⏳ 待实现**: 预签名 URL 功能计划在后续版本中提供。


### 🏷️ Object Tagging

> **⏳ 待实现**: 对象标签功能计划在后续版本中提供。

### 🏥 Health Check

```go
// 基本健康检查
result := client.HealthCheck(nil)
if result.Healthy {
    fmt.Printf("✅ 服务健康，响应时间: %v\n", result.ResponseTime)
} else {
    fmt.Printf("❌ 服务不健康: %v\n", result.Error)
}

// 带超时的健康检查
opts := &core.HealthCheckOptions{
    Timeout: 5 * time.Second,
    Context: context.Background(),
}
result := client.HealthCheck(opts)

// 带重试的健康检查
result := client.HealthCheckWithRetry(opts, 3)
```

> 📖 **完整示例**: 查看 [examples/rustfs/health.go](examples/rustfs/health.go)

### 📊 HTTP Request Tracing

```go
import "github.com/Scorpio69t/rustfs-go/internal/transport"

// 创建追踪 hook
var traceInfo *transport.TraceInfo
hook := func(info transport.TraceInfo) {
    traceCopy := info
    traceInfo = &traceCopy
}

// 创建带追踪的 context
traceCtx := transport.NewTraceContext(ctx, hook)

// 执行请求
bucketSvc := client.Bucket()
exists, err := bucketSvc.Exists(traceCtx, "my-bucket")

// 分析追踪信息
if traceInfo != nil {
    fmt.Printf("连接复用: %v\n", traceInfo.ConnReused)
    fmt.Printf("总耗时: %v\n", traceInfo.TotalDuration())

    // 各阶段耗时
    timings := traceInfo.GetTimings()
    for stage, duration := range timings {
        fmt.Printf("%s: %v\n", stage, duration)
    }
}
```

> 📖 **完整示例**: 查看 [examples/rustfs/trace.go](examples/rustfs/trace.go)

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

- [Bucket Operations](examples/rustfs/bucketops.go) - 存储桶操作示例
- [Object Operations](examples/rustfs/objectops.go) - 对象操作示例
- [Multipart Upload](examples/rustfs/multipart.go) - 分片上传示例
- [Health Check](examples/rustfs/health.go) - 健康检查示例
- [HTTP Tracing](examples/rustfs/trace.go) - HTTP 请求追踪示例

### 运行示例

```bash
cd examples/rustfs

# 运行示例
go run -tags example bucketops.go
go run -tags example objectops.go
go run -tags example multipart.go
go run -tags example health.go
go run -tags example trace.go
```

> **💡 提示**: 运行示例前，请确保：
> - RustFS 服务器正在运行（默认 `127.0.0.1:9000`）
> - 更新示例代码中的访问密钥
> - 创建示例中使用的存储桶

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

<!-- Chinese Section -->
<div id="中文"></div>

## 📖 概述

RustFS Go SDK 是一个用于与 RustFS 对象存储系统交互的 Go 语言客户端库。它完全兼容 S3 API，提供了简洁易用的接口，支持所有标准的 S3 操作。

### ✨ 特性

- ✅ **完全兼容 S3 API** - 支持所有 S3 兼容操作
- ✅ **简洁的 API 设计** - 直观易用的接口
- ✅ **完整的操作支持** - 存储桶管理、对象操作、多部分上传等
- ✅ **流式签名** - 支持 AWS Signature V4 分块上传流式签名
- ✅ **健康检查** - 内置健康检查机制，支持重试
- ✅ **HTTP 追踪** - 请求追踪功能，便于性能监控和调试
- ✅ **错误处理** - 完善的错误处理和重试机制
- ✅ **流式支持** - 高效的大文件流式上传/下载
- ✅ **生产就绪** - 经过充分测试，提供完整示例

## 🚀 安装

```bash
go get github.com/Scorpio69t/rustfs-go
```

## 📚 快速开始

### 初始化客户端

```go
package main

import (
    "context"
    "log"

    "github.com/Scorpio69t/rustfs-go"
    "github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
    // 初始化客户端
    client, err := rustfs.New("127.0.0.1:9000", &rustfs.Options{
        Creds:  credentials.NewStaticV4("your-access-key", "your-secret-key", ""),
        Secure: false, // 设置为 true 使用 HTTPS
    })
    if err != nil {
        log.Fatalln(err)
    }

    ctx := context.Background()
    // 使用客户端进行操作...
}
```

### 📦 存储桶操作

```go
// 创建存储桶
err := client.MakeBucket(ctx, "my-bucket", rustfs.MakeBucketOptions{
    Region: "us-east-1",
})

// 列出所有存储桶
buckets, err := client.ListBuckets(ctx)
for _, bucket := range buckets {
    fmt.Println(bucket.Name)
}

// 检查存储桶是否存在
exists, err := client.BucketExists(ctx, "my-bucket")

// 列出存储桶中的对象
objectsCh := client.ListObjects(ctx, "my-bucket", rustfs.ListObjectsOptions{
    Prefix:  "prefix/",
    MaxKeys: 100,
})
for obj := range objectsCh {
    fmt.Println(obj.Key, obj.Size)
}

// 删除存储桶
err = client.RemoveBucket(ctx, "my-bucket", rustfs.RemoveBucketOptions{})
```

### 📄 对象操作

```go
// 从 reader 上传对象
data := strings.NewReader("Hello, RustFS!")
uploadInfo, err := client.PutObject(ctx, "my-bucket", "my-object.txt",
    data, data.Size(), rustfs.PutObjectOptions{
        ContentType: "text/plain",
        UserMetadata: map[string]string{
            "author": "rustfs-go",
        },
    })

// 从文件上传对象
uploadInfo, err := client.FPutObject(ctx, "my-bucket", "file.txt",
    "/path/to/local/file.txt", rustfs.PutObjectOptions{
        ContentType: "text/plain",
    })

// 下载对象
obj, err := client.GetObject(ctx, "my-bucket", "my-object.txt",
    rustfs.GetObjectOptions{})
defer obj.Close()

buf := make([]byte, 1024)
n, _ := obj.Read(buf)
fmt.Println(string(buf[:n]))

// 下载对象到文件
err = client.FGetObject(ctx, "my-bucket", "my-object.txt",
    "/path/to/local/download.txt", rustfs.GetObjectOptions{})

// 获取对象信息
objInfo, err := client.StatObject(ctx, "my-bucket", "my-object.txt",
    rustfs.StatObjectOptions{})

// 删除对象
err = client.RemoveObject(ctx, "my-bucket", "my-object.txt",
    rustfs.RemoveObjectOptions{})
```

### 🔄 多部分上传

```go
// 初始化多部分上传
uploadID, err := client.InitiateMultipartUpload(ctx, "my-bucket",
    "large-file.txt", rustfs.PutObjectOptions{
        ContentType: "text/plain",
    })

// 上传分片
part1, err := client.UploadPart(ctx, "my-bucket", "large-file.txt",
    uploadID, 1, part1Data, partSize, rustfs.PutObjectPartOptions{})
part2, err := client.UploadPart(ctx, "my-bucket", "large-file.txt",
    uploadID, 2, part2Data, partSize, rustfs.PutObjectPartOptions{})

// 完成多部分上传
parts := []rustfs.CompletePart{
    {PartNumber: part1.PartNumber, ETag: part1.ETag},
    {PartNumber: part2.PartNumber, ETag: part2.ETag},
}
uploadInfo, err := client.CompleteMultipartUpload(ctx, "my-bucket",
    "large-file.txt", uploadID, parts, rustfs.PutObjectOptions{})

// 取消多部分上传
err = client.AbortMultipartUpload(ctx, "my-bucket", "large-file.txt",
    uploadID, rustfs.AbortMultipartUploadOptions{})
```

### 🔐 预签名 URL

```go
// 生成预签名 GET URL（1小时有效）
presignedURL, err := client.PresignedGetObject(ctx, "my-bucket",
    "my-object.txt", time.Hour, url.Values{})

// 生成预签名 PUT URL
presignedPutURL, err := client.PresignedPutObject(ctx, "my-bucket",
    "upload.txt", time.Hour)

// 生成预签名 POST URL
policy := rustfs.NewPostPolicy()
policy.SetExpires(time.Now().Add(time.Hour))
policy.SetCondition("$eq", "bucket", "my-bucket")
policy.SetCondition("$eq", "key", "post-object.txt")
policy.SetCondition("$eq", "Content-Type", "text/plain")

postURL, formData, err := client.PresignedPostPolicy(ctx, policy)
```

### 🔄 对象复制

```go
// 复制对象
copyInfo, err := client.CopyObject(ctx, "source-bucket", "source-object.txt",
    "dest-bucket", "dest-object.txt", rustfs.CopyObjectOptions{
        ContentType: "text/plain",
    })
```

### 🏷️ 对象标签

> **⏳ 待实现**: 对象标签功能计划在后续版本中提供。

### 🏥 健康检查

```go
// 基本健康检查
result := client.HealthCheck(nil)
if result.Healthy {
    fmt.Printf("✅ 服务健康，响应时间: %v\n", result.ResponseTime)
} else {
    fmt.Printf("❌ 服务不健康: %v\n", result.Error)
}

// 带超时的健康检查
opts := &core.HealthCheckOptions{
    Timeout: 5 * time.Second,
    Context: context.Background(),
}
result := client.HealthCheck(opts)

// 检查特定存储桶
opts := &core.HealthCheckOptions{
    Timeout:    3 * time.Second,
    BucketName: "my-bucket",
    Context:    ctx,
}
result := client.HealthCheck(opts)

// 带重试的健康检查
result := client.HealthCheckWithRetry(opts, 3)

// 定期健康检查
ticker := time.NewTicker(30 * time.Second)
defer ticker.Stop()

for range ticker.C {
    result := client.HealthCheck(nil)
    if !result.Healthy {
        log.Printf("健康检查失败: %v", result.Error)
    }
}
```

> 📖 **完整示例**: 查看 [examples/rustfs/health.go](examples/rustfs/health.go)

### 📊 HTTP 请求追踪

```go
import "github.com/Scorpio69t/rustfs-go/internal/transport"

// 创建追踪回调
var traceInfo *transport.TraceInfo
hook := func(info transport.TraceInfo) {
    traceCopy := info
    traceInfo = &traceCopy
}

// 创建带追踪的 context
traceCtx := transport.NewTraceContext(ctx, hook)

// 执行请求
bucketSvc := client.Bucket()
exists, err := bucketSvc.Exists(traceCtx, "my-bucket")

// 分析追踪信息
if traceInfo != nil {
    fmt.Printf("连接复用: %v\n", traceInfo.ConnReused)
    fmt.Printf("总耗时: %v\n", traceInfo.TotalDuration())

    // 各阶段耗时
    timings := traceInfo.GetTimings()
    for stage, duration := range timings {
        fmt.Printf("%s: %v\n", stage, duration)
    }
}
```

> 📖 **完整示例**: 查看 [examples/rustfs/trace.go](examples/rustfs/trace.go)

## 🔑 凭证管理

### 静态凭证

```go
creds := credentials.NewStaticV4("access-key", "secret-key", "")
```

### 环境变量

```go
creds := credentials.NewEnvAWS()
// 从环境变量读取:
// AWS_ACCESS_KEY_ID
// AWS_SECRET_ACCESS_KEY
// AWS_SESSION_TOKEN
```

## ⚙️ 配置选项

```go
client, err := rustfs.New("rustfs.example.com", &rustfs.Options{
    Creds:        credentials.NewStaticV4("access-key", "secret-key", ""),
    Secure:       true,              // 使用 HTTPS
    Region:       "us-east-1",       // 区域
    BucketLookup: rustfs.BucketLookupDNS, // 存储桶查找方式
    Transport:    nil,               // 自定义 HTTP Transport
    MaxRetries:   10,                // 最大重试次数
})
```

## 📝 示例代码

更多示例代码请查看 [examples/rustfs](examples/rustfs/) 目录：

- [存储桶操作示例](examples/rustfs/bucketops.go)
- [对象操作示例](examples/rustfs/objectops.go)
- [分片上传示例](examples/rustfs/multipart.go)
- [健康检查示例](examples/rustfs/health.go)
- [HTTP 追踪示例](examples/rustfs/trace.go)

### 运行示例

```bash
cd examples/rustfs

# 运行示例
go run -tags example bucketops.go
go run -tags example objectops.go
go run -tags example multipart.go
go run -tags example health.go
go run -tags example trace.go
```

> **💡 提示**: 运行示例前，请确保：
> - RustFS 服务器正在运行（默认 `127.0.0.1:9000`）
> - 更新示例代码中的访问密钥
> - 创建示例中使用的存储桶

## 📖 API 文档

完整的 API 文档请访问: https://pkg.go.dev/github.com/Scorpio69t/rustfs-go

## 📄 许可证

本项目采用 Apache License 2.0 许可证。详情请查看 [LICENSE](LICENSE) 文件。

## 🤝 贡献

欢迎贡献代码！请查看 [CONTRIBUTING.md](CONTRIBUTING.md) 了解贡献指南。

## 🔗 参考资源

- [AWS S3 API 文档](https://docs.aws.amazon.com/AmazonS3/latest/API/Welcome.html) - API 规范
- [AWS Signature Version 4](https://docs.aws.amazon.com/general/latest/gr/signature-version-4.html) - 签名算法

## 💬 支持

如有问题或建议，请提交 [Issue](https://github.com/Scorpio69t/rustfs-go/issues)。

---

<div align="center">

**Made with ❤️ by the RustFS Go SDK community**

[⬆ Back to Top](#-rustfs-go-sdk)

</div>
