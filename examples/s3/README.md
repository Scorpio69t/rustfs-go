# S3 API 示例集

本目录包含 RustFS Go SDK 的完整 S3 API 使用示例，帮助用户学习如何使用 SDK 的各种功能。

## 📋 前置条件

运行这些示例前，请确保：

1. **RustFS 服务运行中**
   ```bash
   # 使用 Docker 启动本地 MinIO 服务器（用于测试）
   docker run -p 9000:9000 -p 9001:9001 \
     -e "MINIO_ROOT_USER=minioadmin" \
     -e "MINIO_ROOT_PASSWORD=minioadmin" \
     minio/minio server /data --console-address ":9001"
   ```

2. **配置访问凭证**
   - 修改示例中的 `accessKey` 和 `secretKey`
   - 或设置环境变量 `ACCESS_KEY` 和 `SECRET_KEY`

3. **安装依赖**
   ```bash
   cd examples/s3
   go mod download
   ```

## 🚀 运行示例

### 编译并运行单个示例

```bash
# 编译
go build -tags example bucket-create.go

# 运行
./bucket-create
```

### 直接运行（不编译）

```bash
go run -tags example bucket-create.go
```

## 📚 示例分类

### 🗂️ 存储桶操作

| 示例文件 | 功能描述 |
|---------|---------|
| `bucket-create.go` | 创建存储桶 |
| `bucket-delete.go` | 删除存储桶 |
| `bucket-list.go` | 列出所有存储桶 |
| `bucket-exists.go` | 检查存储桶是否存在 |
| `bucket-location.go` | 获取存储桶位置 |

### 🏷️ Bucket Tagging

| 示例文件 | 功能描述 |
|---------|---------|
| `bucket-tagging-set.go` | Set bucket tags |
| `bucket-tagging-get.go` | Get bucket tags |
| `bucket-tagging-delete.go` | Delete bucket tags |

### 📦 对象基础操作

| 示例文件 | 功能描述 |
|---------|---------|
| `object-put.go` | 上传对象（从内存） |
| `object-get.go` | 下载对象 |
| `object-get-override-headers.go` | Get object with response header overrides |
| `object-stat.go` | 获取对象信息 |
| `object-copy.go` | 复制对象 |
| `object-delete.go` | 删除单个对象 |
| `object-delete-multiple.go` | 批量删除对象 |
| `object-list.go` | 列出对象 |
| `object-list-versions.go` | 列出对象版本 |
| `object-list-max-keys.go` | List objects with max-keys and start-after |
| `object-list-multipart-uploads.go` | List multipart uploads |
| `object-list-parts.go` | List multipart upload parts |
| `object-put-streaming.go` | 流式上传对象 |
| `object-put-progress.go` | 带进度显示的上传 |
| `object-put-checksum.go` | Upload with checksum mode |
| `object-put-s3-accelerate.go` | Upload with S3 Accelerate |

### 📁 文件操作

| 示例文件 | 功能描述 |
|---------|---------|
| `file-upload.go` | 从文件上传对象 |
| `file-download.go` | 下载对象到文件 |

### 🔄 版本控制

| 示例文件 | 功能描述 |
|---------|---------|
| `versioning-enable.go` | 启用版本控制 |
| `versioning-suspend.go` | 暂停版本控制 |
| `versioning-status.go` | 获取版本控制状态 |

### 🏷️ 对象标签

| 示例文件 | 功能描述 |
|---------|---------|
| `tagging-object-set.go` | 设置对象标签 |
| `tagging-object-get.go` | 获取对象标签 |
| `tagging-object-delete.go` | 删除对象标签 |
| `tagging-object-put-with-tags.go` | 上传带标签的对象 |

### 🔗 预签名 URL

| 示例文件 | 功能描述 |
|---------|---------|
| `presigned-get.go` | 生成预签名 GET URL |
| `presigned-head.go` | 生成预签名 HEAD URL |
| `presigned-put.go` | 生成预签名 PUT URL |
| `presigned-get-override-headers.go` | 预签名 GET 并覆盖响应头 |
| `presigned-post-policy.go` | Generate presigned POST policy |

### 🔐 Encryption

| 示例文件 | 功能描述 |
|---------|---------|
| `encryption-sse-s3-put.go` | Upload with SSE-S3 |
| `encryption-sse-s3-get.go` | Download with SSE-S3 |
| `encryption-sse-c-put.go` | Upload with SSE-C |
| `encryption-sse-c-get.go` | Download with SSE-C |
| `encryption-bucket-set.go` | Set bucket encryption |
| `encryption-bucket-get.go` | Get bucket encryption |
| `encryption-bucket-delete.go` | Delete bucket encryption |
| `encryption-bucket-config.go` | Bucket encryption configuration (legacy) |
| `debug-sse-c.go` | Debug SSE-C headers |
| `object-put-client-encryption.go` | Upload with client-side encryption (CSE) |
| `object-get-client-encryption.go` | Download with client-side encryption (CSE) |

### 📋 存储桶策略和生命周期

| 示例文件 | 功能描述 |
|---------|---------|
| `policy-set.go` | 设置存储桶策略 |
| `policy-get.go` | 获取存储桶策略 |
| `policy-delete.go` | 删除存储桶策略 |
| `lifecycle-set.go` | 设置生命周期规则 |
| `lifecycle-get.go` | 获取生命周期规则 |
| `lifecycle-delete.go` | 删除生命周期规则 |

### 🏥 健康检查

| 示例文件 | 功能描述 |
|---------|---------|
| `health-check.go` | 服务健康检查和监控 |

### 🔄 跨区复制

| 示例文件 | 功能描述 |
|---------|---------|
| `replication-set.go` | Set replication configuration |
| `replication-get.go` | Get replication configuration |
| `replication-metrics.go` | Get replication metrics |
| `replication-delete.go` | Delete replication configuration |

### 🔔 事件通知

| 示例文件 | 功能描述 |
|---------|---------|
| `notification-set.go` | Set notification configuration |
| `notification-get.go` | Get notification configuration |
| `notification-listen.go` | Listen for notifications |
| `notification-delete.go` | Delete notification configuration |

### 🌐 CORS 配置

| 示例文件 | 功能描述 |
|---------|---------|
| `cors-set.go` | Set CORS configuration |
| `cors-get.go` | Get CORS configuration |
| `cors-delete.go` | Delete CORS configuration |

### 🔑 访问控制

| 示例文件 | 功能描述 |
|---------|---------|
| `acl-object-get.go` | Get object ACL |
| `acl-object-set.go` | Set object ACL |

### 🔒 Object Lock

| 示例文件 | 功能描述 |
|---------|---------|
| `object-lock-config-set.go` | Set object lock configuration |
| `object-lock-config-get.go` | Get object lock configuration |
| `object-legal-hold-set.go` | Set legal hold |
| `object-legal-hold-get.go` | Get legal hold |
| `object-retention-set.go` | Set retention |
| `object-retention-get.go` | Get retention |

### 🧩 Advanced Object Operations

| 示例文件 | 功能描述 |
|---------|---------|
| `object-compose.go` | Compose objects |
| `object-append.go` | Append to object |
| `object-copy-with-new-tags.go` | Copy object and replace tags |

### 🔍 Object Select

| 示例文件 | 功能描述 |
|---------|---------|
| `object-select-csv.go` | Select CSV objects |
| `object-select-json.go` | Select JSON objects |

### 🧊 Object Restore

| 示例文件 | 功能描述 |
|---------|---------|
| `object-restore.go` | Restore an archived object |

### 🧪 End-to-End & Performance

| 示例文件 | 功能描述 |
|---------|---------|
| `e2e-scenario.go` | End-to-end scenario test |
| `perf-sse.go` | SSE-S3 upload performance comparison |
| `perf-large-object.go` | Large object upload/download performance |
| `perf-concurrent.go` | Concurrent upload performance |

## 💡 使用提示

### 配置管理

建议使用环境变量管理凭证：

```go
import "os"

endpoint := os.Getenv("RUSTFS_ENDPOINT")
if endpoint == "" {
    endpoint = "127.0.0.1:9000"
}

accessKey := os.Getenv("ACCESS_KEY")
if accessKey == "" {
    accessKey = "minioadmin"
}

secretKey := os.Getenv("SECRET_KEY")
if secretKey == "" {
    secretKey = "minioadmin"
}
```

### 错误处理

所有示例都包含完整的错误处理：

```go
if err != nil {
    log.Fatalf("操作失败: %v", err)
}
```

### 上下文管理

示例使用 `context.Background()`，生产环境建议使用带超时的上下文：

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

## 🔗 相关资源

- [RustFS Go SDK 文档](../../README.zh.md)
- [API 参考](https://pkg.go.dev/github.com/Scorpio69t/rustfs-go)
- [问题反馈](https://github.com/Scorpio69t/rustfs-go/issues)

## 📝 贡献

欢迎提交新的示例或改进现有示例！请参考 [CONTRIBUTING.md](../../CONTRIBUTING.md)。
