# 设计文档：S3 示例迁移

## Context
从 MinIO SDK 的旧 API 迁移到 RustFS SDK 的新模块化 API，需要重构 71 个示例文件。这些示例展示了 S3 兼容存储的各种功能，从基础的存储桶和对象操作到高级的加密、版本控制和对象锁定等特性。

**约束：**
- 必须避免版权纠纷：重写代码逻辑，不直接复制
- 必须适配新 API：使用服务化接口和选项模式
- 必须保持功能完整性：所有原有功能都要支持
- 必须易于理解：清晰的注释和示例结构

**利益相关者：**
- SDK 用户：需要清晰的示例来学习如何使用
- 迁移用户：从 MinIO SDK 迁移的用户需要参考
- 开发团队：需要维护和扩展示例

## Goals / Non-Goals

### Goals
- ✅ 将所有 71 个 S3 示例迁移到新 API
- ✅ 重构代码以避免版权问题（独立实现）
- ✅ 采用统一的代码风格和模板
- ✅ 提供清晰的中文注释
- ✅ 确保示例可独立运行
- ✅ 按功能分类组织示例

### Non-Goals
- ❌ 不创建新的功能（仅迁移现有示例）
- ❌ 不修改核心 SDK 代码
- ❌ 不添加单元测试（示例代码本身就是测试）
- ❌ 不支持已废弃的 API（如果新 SDK 不支持）

## Decisions

### Decision 1: API 迁移模式

**旧 API 风格（MinIO SDK）：**
```go
s3Client, _ := minio.New("endpoint", &minio.Options{
    Creds: credentials.NewStaticV4("key", "secret", ""),
    Secure: true,
})

info, err := s3Client.PutObject(ctx, "bucket", "object",
    reader, size, minio.PutObjectOptions{
        ContentType: "text/plain",
    })
```

**新 API 风格（RustFS SDK）：**
```go
client, _ := rustfs.New("endpoint", &rustfs.Options{
    Credentials: credentials.NewStaticV4("key", "secret", ""),
    Secure: true,
})

objectSvc := client.Object()
info, err := objectSvc.Put(ctx, "bucket", "object",
    reader, size,
    object.WithContentType("text/plain"),
)
```

**关键变化：**
1. 服务分离：`client.Bucket()` 和 `client.Object()`
2. 选项函数：`object.WithXxx()` 代替结构体选项
3. 包名变更：`minio` → `rustfs`
4. 导入路径：`github.com/minio/minio-go/v7` → `github.com/Scorpio69t/rustfs-go`

### Decision 2: 代码重构策略

为避免版权问题，采用以下策略：
1. **重新实现**：理解功能后独立编写，不直接复制
2. **改进命名**：使用更描述性的变量和函数名
3. **添加注释**：详细的中文注释解释每一步
4. **重组结构**：优化代码布局和错误处理
5. **移除版权**：删除所有 MinIO 版权声明

**示例对比：**

旧代码（MinIO）：
```go
// MinIO 版权声明...
func main() {
    s3Client, err := minio.New("s3.amazonaws.com", &minio.Options{
        Creds:  credentials.NewStaticV4("YOUR-ACCESSKEYID", "YOUR-SECRETACCESSKEY", ""),
        Secure: true,
    })
    if err != nil {
        log.Fatalln(err)
    }
    // 简单的操作
}
```

新代码（RustFS）：
```go
//go:build example
// +build example

// 示例：创建存储桶
// 演示如何使用 RustFS Go SDK 创建一个新的存储桶
package main

import (
    "context"
    "log"

    "github.com/Scorpio69t/rustfs-go"
    "github.com/Scorpio69t/rustfs-go/bucket"
    "github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
    // 配置连接参数
    endpoint := "127.0.0.1:9000"
    accessKey := "YOUR-ACCESS-KEY"
    secretKey := "YOUR-SECRET-KEY"

    // 初始化客户端
    client, err := rustfs.New(endpoint, &rustfs.Options{
        Credentials: credentials.NewStaticV4(accessKey, secretKey, ""),
        Secure:      false, // 本地测试使用 HTTP
    })
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    ctx := context.Background()
    bucketSvc := client.Bucket()

    // 创建存储桶
    bucketName := "my-bucket"
    err = bucketSvc.Create(ctx, bucketName,
        bucket.WithRegion("us-east-1"),
    )
    if err != nil {
        log.Fatalf("Failed to create bucket: %v", err)
    }

    log.Printf("Successfully created bucket: %s", bucketName)
}
```

### Decision 3: 目录组织

采用扁平化结构，按功能命名：

```
examples/s3/
├── README.md                          # 示例索引和使用说明
├── go.mod                             # 依赖管理
├── go.sum
│
# 基础操作
├── bucket-create.go                   # 创建存储桶
├── bucket-delete.go                   # 删除存储桶
├── bucket-list.go                     # 列出存储桶
├── bucket-exists.go                   # 检查存储桶
│
# 对象操作
├── object-put.go                      # 上传对象
├── object-get.go                      # 下载对象
├── object-copy.go                     # 复制对象
├── object-delete.go                   # 删除对象
├── object-stat.go                     # 对象信息
├── object-list.go                     # 列出对象
│
# 文件操作
├── file-upload.go                     # 从文件上传
├── file-download.go                   # 下载到文件
│
# 高级功能
├── versioning-enable.go               # 启用版本控制
├── versioning-suspend.go              # 暂停版本控制
├── versioning-list.go                 # 列出版本
├── tagging-object-set.go              # 设置对象标签
├── tagging-object-get.go              # 获取对象标签
├── encryption-sse-put.go              # SSE 上传
├── encryption-sse-get.go              # SSE 下载
├── presigned-get.go                   # 预签名 GET
├── presigned-put.go                   # 预签名 PUT
│
# ... 更多示例
```

**命名规则：**
- `<resource>-<action>.go`：如 `bucket-create.go`
- `<feature>-<resource>-<action>.go`：如 `versioning-bucket-enable.go`
- 使用小写加连字符（kebab-case）

### Decision 4: 示例代码模板

每个示例遵循统一模板：

```go
//go:build example
// +build example

// 示例：<功能描述>
// 演示如何使用 RustFS Go SDK <具体操作>
package main

import (
    "context"
    "log"

    "github.com/Scorpio69t/rustfs-go"
    "github.com/Scorpio69t/rustfs-go/<package>"
    "github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
    // 1. 配置参数
    const (
        endpoint  = "127.0.0.1:9000"
        accessKey = "YOUR-ACCESS-KEY"
        secretKey = "YOUR-SECRET-KEY"
    )

    // 2. 初始化客户端
    client, err := rustfs.New(endpoint, &rustfs.Options{
        Credentials: credentials.NewStaticV4(accessKey, secretKey, ""),
        Secure:      false,
    })
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    // 3. 执行操作
    ctx := context.Background()
    // ... 具体操作代码

    // 4. 输出结果
    log.Printf("Operation completed successfully")
}
```

## Alternatives Considered

### Alternative 1: 直接复制并修改导入
**拒绝原因：** 可能引起版权纠纷，代码风格不统一

### Alternative 2: 只迁移核心示例
**拒绝原因：** 功能不完整，用户缺少参考

### Alternative 3: 创建完全不同的示例
**拒绝原因：** 不利于迁移用户对照学习

## Risks / Trade-offs

### Risk 1: 代码重构可能引入错误
**缓解措施：**
- 每个示例独立编译测试
- 核心示例进行功能验证
- 参考已有的 rustfs 示例确保正确性

### Risk 2: 新 API 可能不支持某些功能
**缓解措施：**
- 提前识别不支持的功能
- 在文档中标注差异
- 为缺失功能创建 issue 跟踪

### Risk 3: 迁移工作量大（71个文件）
**缓解措施：**
- 分批迁移（6批次）
- 使用模板加速开发
- 优先迁移高频使用的核心功能

## Migration Plan

### 分批迁移策略

**第一批（基础操作，~30 个文件）：**
- 存储桶：创建、删除、列表、存在性
- 对象：上传、下载、复制、删除、信息
- 优先级：高

**第二批（版本和标签，~12 个文件）：**
- 版本控制相关
- 对象和存储桶标签
- 优先级：中

**第三批（加密和安全，~13 个文件）：**
- 服务端加密
- 客户端加密
- 对象锁定和保留
- 优先级：中

**第四批（策略和配置，~10 个文件）：**
- 存储桶策略
- 生命周期
- 复制
- 通知
- 优先级：中

**第五批（预签名 URL，~5 个文件）：**
- 预签名操作
- 优先级：高

**第六批（高级功能，~8 个文件）：**
- 流式上传
- 对象查询
- 健康检查
- 优先级：低

### 回滚计划
如果迁移过程中发现重大问题：
1. 停止当前批次的迁移
2. 回滚已提交的代码
3. 重新评估策略
4. 修复问题后继续

## Open Questions
- [ ] 是否需要为每个示例创建对应的集成测试？
- [ ] 是否需要添加 Makefile 来批量运行示例？
- [ ] 是否需要创建示例运行的 Docker 环境？
- [ ] 某些高级功能（如 S3 Select）新 SDK 是否支持？
