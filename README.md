# RustFS Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/Scorpio69t/rustfs-go.svg)](https://pkg.go.dev/github.com/Scorpio69t/rustfs-go)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

RustFS Go SDK 是一个用于与 RustFS 对象存储系统交互的 Go 语言客户端库。它完全兼容 S3 API，提供了简洁易用的接口，支持所有标准的 S3 操作。

## 特性

- ✅ 完全兼容 S3 API
- ✅ 简洁直观的 API 设计
- ✅ 支持所有标准 S3 操作（存储桶、对象、多部分上传等）
- ✅ 支持预签名 URL
- ✅ 完整的错误处理和重试机制
- ✅ 支持流式上传/下载
- ✅ 完整的单元测试和示例代码

## 安装

```bash
go get github.com/Scorpio69t/rustfs-go
```

## 快速开始

### 初始化客户端

```go
package main

import (
    "context"
    "log"

    "github.com/Scorpio69t/rustfs-go/v1"
    "github.com/Scorpio69t/rustfs-go/v1/credentials"
)

func main() {
    // 初始化客户端
    client, err := rustfs.New("rustfs.example.com", &rustfs.Options{
        Creds:  credentials.NewStaticV4("your-access-key", "your-secret-key", ""),
        Secure: true,
        Region: "us-east-1",
    })
    if err != nil {
        log.Fatalln(err)
    }

    ctx := context.Background()
    // 使用客户端进行操作...
}
```

### 存储桶操作

```go
// 创建存储桶
err := client.MakeBucket(ctx, "my-bucket", rustfs.MakeBucketOptions{
    Region: "us-east-1",
})

// 列出所有存储桶
buckets, err := client.ListBuckets(ctx)

// 检查存储桶是否存在
exists, err := client.BucketExists(ctx, "my-bucket")

// 列出存储桶中的对象
objectsCh := client.ListObjects(ctx, "my-bucket", rustfs.ListObjectsOptions{
    Prefix:  "prefix/",
    MaxKeys: 100,
})
for obj := range objectsCh {
    fmt.Println(obj.Key)
}

// 删除存储桶
err = client.RemoveBucket(ctx, "my-bucket", rustfs.RemoveBucketOptions{})
```

### 对象操作

```go
// 上传对象
data := strings.NewReader("Hello, RustFS!")
uploadInfo, err := client.PutObject(ctx, "my-bucket", "my-object.txt", data, data.Size(), rustfs.PutObjectOptions{
    ContentType: "text/plain",
})

// 从文件上传
uploadInfo, err := client.FPutObject(ctx, "my-bucket", "file.txt", "/path/to/local/file.txt", rustfs.PutObjectOptions{
    ContentType: "text/plain",
})

// 下载对象
obj, err := client.GetObject(ctx, "my-bucket", "my-object.txt", rustfs.GetObjectOptions{})
defer obj.Reader.Close()
// 读取对象内容...

// 下载对象到文件
err = client.FGetObject(ctx, "my-bucket", "my-object.txt", "/path/to/local/file.txt", rustfs.GetObjectOptions{})

// 获取对象信息
objInfo, err := client.StatObject(ctx, "my-bucket", "my-object.txt", rustfs.StatObjectOptions{})

// 删除对象
err = client.RemoveObject(ctx, "my-bucket", "my-object.txt", rustfs.RemoveObjectOptions{})
```

### 多部分上传

```go
// 初始化多部分上传
uploadID, err := client.InitiateMultipartUpload(ctx, "my-bucket", "large-file.txt", rustfs.PutObjectOptions{
    ContentType: "text/plain",
})

// 上传分片
part1, err := client.UploadPart(ctx, "my-bucket", "large-file.txt", uploadID, 1, part1Data, partSize, rustfs.PutObjectPartOptions{})
part2, err := client.UploadPart(ctx, "my-bucket", "large-file.txt", uploadID, 2, part2Data, partSize, rustfs.PutObjectPartOptions{})

// 完成多部分上传
parts := []rustfs.CompletePart{
    {PartNumber: part1.PartNumber, ETag: part1.ETag},
    {PartNumber: part2.PartNumber, ETag: part2.ETag},
}
uploadInfo, err := client.CompleteMultipartUpload(ctx, "my-bucket", "large-file.txt", uploadID, parts, rustfs.PutObjectOptions{})

// 取消多部分上传
err = client.AbortMultipartUpload(ctx, "my-bucket", "large-file.txt", uploadID, rustfs.AbortMultipartUploadOptions{})
```

### 预签名 URL

```go
// 生成预签名 GET URL（1小时有效）
presignedURL, err := client.PresignedGetObject(ctx, "my-bucket", "my-object.txt", time.Hour, url.Values{})

// 生成预签名 PUT URL
presignedPutURL, err := client.PresignedPutObject(ctx, "my-bucket", "upload.txt", time.Hour)

// 生成预签名 POST URL
policy := &rustfs.PostPolicy{
    Expiration: time.Now().Add(time.Hour),
    Conditions: []map[string]interface{}{
        {"bucket": "my-bucket"},
        {"key": "post-object.txt"},
    },
}
postURL, formData, err := client.PresignedPostPolicy(ctx, policy)
```

### 对象复制

```go
// 复制对象
copyInfo, err := client.CopyObject(ctx, "source-bucket", "source-object.txt",
    "dest-bucket", "dest-object.txt", rustfs.CopyObjectOptions{
        ContentType: "text/plain",
    })
```

### 对象标签

```go
// 设置对象标签
err := client.SetObjectTagging(ctx, "my-bucket", "my-object.txt", map[string]string{
    "environment": "production",
    "project":     "rustfs-go",
})

// 获取对象标签
tags, err := client.GetObjectTagging(ctx, "my-bucket", "my-object.txt")

// 删除对象标签
err = client.RemoveObjectTagging(ctx, "my-bucket", "my-object.txt")
```

## 凭证管理

### 静态凭证

```go
creds := credentials.NewStaticV4("access-key", "secret-key", "")
```

### 环境变量凭证

```go
creds := credentials.NewEnvAWS()
// 从环境变量读取:
// AWS_ACCESS_KEY_ID
// AWS_SECRET_ACCESS_KEY
// AWS_SESSION_TOKEN
```

## 配置选项

```go
client, err := rustfs.New("rustfs.example.com", &rustfs.Options{
    Creds:        credentials.NewStaticV4("access-key", "secret-key", ""),
    Secure:       true,              // 使用 HTTPS
    Region:       "us-east-1",       // 区域
    BucketLookup: rustfs.BucketLookupDNS, // 存储桶查找方式
    Transport:    nil,               // 自定义 HTTP Transport
})
```

## 示例代码

更多示例代码请查看 [examples](examples/) 目录：

- [存储桶操作示例](examples/bucketops.go)
- [对象操作示例](examples/objectops.go)
- [多部分上传示例](examples/multipart.go)
- [预签名 URL 示例](examples/presigned.go)

## API 文档

完整的 API 文档请访问: https://pkg.go.dev/github.com/Scorpio69t/rustfs-go

## 许可证

本项目采用 Apache License 2.0 许可证。详情请查看 [LICENSE](LICENSE) 文件。

## 贡献

欢迎贡献代码！请查看 [CONTRIBUTING.md](CONTRIBUTING.md) 了解贡献指南。

## 参考资源

- [MinIO Go SDK](https://github.com/minio/minio-go) - 主要参考实现
- [AWS S3 API 文档](https://docs.aws.amazon.com/AmazonS3/latest/API/Welcome.html) - API 规范
- [AWS Signature Version 4](https://docs.aws.amazon.com/general/latest/gr/signature-version-4.html) - 签名算法

## 支持

如有问题或建议，请提交 [Issue](https://github.com/Scorpio69t/rustfs-go/v1/issues)。
