# 🚀 RustFS Go SDK

<div align="center">

[![Go Reference](https://pkg.go.dev/badge/github.com/Scorpio69t/rustfs-go.svg)](https://pkg.go.dev/github.com/Scorpio69t/rustfs-go)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8?logo=go)](https://go.dev/)
[![GitHub stars](https://img.shields.io/github/stars/Scorpio69t/rustfs-go?style=social)](https://github.com/Scorpio69t/rustfs-go)

**面向 RustFS 对象存储的高性能 Go 客户端 SDK**

[English](README.md) | [中文](README.zh.md)

</div>

---

## 📖 概述

RustFS Go SDK 是一个用于与 RustFS 对象存储系统交互的 Go 语言客户端库。它完全兼容 S3 API，提供简洁易用的接口，支持所有标准的 S3 操作。

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
- ✅ **数据保护** - 桶版本控制、跨区复制、事件通知、访问日志（示例见 `examples/rustfs/data_protection.go`）

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
// 获取 Bucket 服务
bucketSvc := client.Bucket()

// 创建存储桶
err := bucketSvc.Create(ctx, "my-bucket",
    bucket.WithRegion("us-east-1"),
    bucket.WithObjectLocking(false),
)

// 启用版本控制与保护配置
_ = bucketSvc.SetVersioning(ctx, "my-bucket", types.VersioningConfig{Status: "Enabled"})
_ = bucketSvc.SetReplication(ctx, "my-bucket", []byte(`<ReplicationConfiguration>...</ReplicationConfiguration>`))
_ = bucketSvc.SetNotification(ctx, "my-bucket", []byte(`<NotificationConfiguration>...</NotificationConfiguration>`))
_ = bucketSvc.SetLogging(ctx, "my-bucket", []byte(`<BucketLoggingStatus>...</BucketLoggingStatus>`))

// 列出所有存储桶
buckets, err := bucketSvc.List(ctx)
for _, bucket := range buckets {
    fmt.Println(bucket.Name)
}

// 检查存储桶是否存在
exists, err := bucketSvc.Exists(ctx, "my-bucket")

// 获取存储桶位置
location, err := bucketSvc.GetLocation(ctx, "my-bucket")

// 删除存储桶
err = bucketSvc.Delete(ctx, "my-bucket")
// 或强制删除（RustFS 扩展，删除所有对象）
err = bucketSvc.Delete(ctx, "my-bucket", bucket.WithForceDelete(true))
```

### 📄 对象操作

```go
// 获取 Object 服务
objectSvc := client.Object()

// 从 reader 上传对象
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

// 下载对象
reader, objInfo, err := objectSvc.Get(ctx, "my-bucket", "my-object.txt")
defer reader.Close()

buf := make([]byte, 1024)
n, _ := reader.Read(buf)
fmt.Println(string(buf[:n]))

// 指定范围下载
reader, _, err := objectSvc.Get(ctx, "my-bucket", "my-object.txt",
    object.WithGetRange(0, 99), // 前 100 字节
)

// 获取对象信息
objInfo, err := objectSvc.Stat(ctx, "my-bucket", "my-object.txt")

// 列出对象
objectsCh := objectSvc.List(ctx, "my-bucket")
for obj := range objectsCh {
    if obj.Err != nil {
        log.Println(obj.Err)
        break
    }
    fmt.Println(obj.Key, obj.Size)
}

// 复制对象
copyInfo, err := objectSvc.Copy(ctx,
    "my-bucket", "copy.txt",     // 目标
    "my-bucket", "my-object.txt", // 来源
)

// 删除对象
err = objectSvc.Delete(ctx, "my-bucket", "my-object.txt")
```

### 🔄 多部分上传

```go
// 获取支持分片上传的 Object 服务
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

// 1. 初始化多部分上传
uploadID, err := multipartSvc.InitiateMultipartUpload(ctx, "my-bucket", "large-file.txt",
    object.WithContentType("text/plain"),
)

// 2. 上传分片
var parts []types.ObjectPart
part1, err := multipartSvc.UploadPart(ctx, "my-bucket", "large-file.txt",
    uploadID, 1, part1Data, partSize)
parts = append(parts, part1)

part2, err := multipartSvc.UploadPart(ctx, "my-bucket", "large-file.txt",
    uploadID, 2, part2Data, partSize)
parts = append(parts, part2)

// 3. 完成多部分上传
uploadInfo, err := multipartSvc.CompleteMultipartUpload(ctx, "my-bucket",
    "large-file.txt", uploadID, parts)

// 4. 需要时取消多部分上传
err = multipartSvc.AbortMultipartUpload(ctx, "my-bucket", "large-file.txt", uploadID)
```

> 📖 **完整示例**: 查看 [examples/rustfs/multipart.go](examples/rustfs/multipart.go)

### 🔐 预签名 URL

> **⏳ 待实现**: 预签名 URL 功能计划在后续版本提供。

### 🏷️ 对象标签

> **⏳ 待实现**: 对象标签功能计划在后续版本提供。

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

// 带重试的健康检查
result := client.HealthCheckWithRetry(opts, 3)
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

我们在两个目录中提供了全面的示例：

### 🔧 RustFS 示例 ([examples/rustfs](examples/rustfs/))

演示 RustFS 高级功能的示例：

- [存储桶操作示例](examples/rustfs/bucketops.go) - 创建、列出、删除存储桶
- [对象操作示例](examples/rustfs/objectops.go) - 上传、下载、复制对象
- [分片上传示例](examples/rustfs/multipart.go) - 大文件多部分上传
- [健康检查示例](examples/rustfs/health.go) - 服务健康监控
- [HTTP 追踪示例](examples/rustfs/trace.go) - 请求追踪和调试
- [对象标签示例](examples/rustfs/object_tagging.go) - 标签管理
- [存储桶策略与生命周期](examples/rustfs/bucket_policy_lifecycle.go) - 策略和生命周期配置
- [数据保护示例](examples/rustfs/data_protection.go) - 版本控制、复制、通知、日志

### 📦 S3 兼容示例 ([examples/s3](examples/s3/))

标准 S3 API 示例（35 个示例涵盖所有常用操作）：

- **存储桶操作** (5个): 创建、删除、列出、检查存在、获取位置
- **对象操作** (11个): 上传、下载、复制、删除、统计、列出、列出版本、文件上传/下载、批量删除、流式上传、进度显示
- **版本控制** (3个): 启用、暂停、状态
- **对象标签** (4个): 设置、获取、删除标签、上传带标签对象
- **存储桶策略** (3个): 设置、获取、删除策略
- **生命周期管理** (3个): 设置、获取、删除生命周期规则
- **预签名 URL** (3个): GET、PUT 和带响应头覆盖的 GET
- **健康检查** (1个): 服务健康监控

完整列表和使用说明请查看 [examples/s3/README.md](examples/s3/README.md)。

### 运行示例

```bash
# RustFS 示例
cd examples/rustfs
go run -tags example bucketops.go
go run -tags example objectops.go

# S3 示例
cd examples/s3
go run -tags example bucket-create.go
go run -tags example object-put.go
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

[⬆ 回到顶部](#-rustfs-go-sdk)

</div>
