# Core API Specification

## ADDED Requirements

### Requirement: 签名器服务

RustFS SDK SHALL 提供完整的 AWS Signature V4 和 V2 签名实现，用于对所有 API 请求进行签名认证。

#### Scenario: V4 签名成功

- **WHEN** 客户端使用有效的访问密钥和密钥密钥调用签名器
- **AND** 提供完整的 HTTP 请求（方法、URL、头部、负载）
- **THEN** 签名器生成正确的 AWS Signature V4 Authorization 头
- **AND** 请求可以被 RustFS 服务器成功验证

#### Scenario: V4 预签名 URL 生成

- **WHEN** 客户端请求生成预签名 URL
- **AND** 指定过期时间（例如 1 小时）
- **THEN** 签名器生成包含签名参数的 URL
- **AND** URL 在过期时间内有效

#### Scenario: V2 签名兼容性

- **WHEN** 客户端配置使用 V2 签名
- **THEN** 签名器使用 AWS Signature V2 算法
- **AND** 与旧版本服务保持兼容

### Requirement: HTTP 传输层

RustFS SDK SHALL 提供可配置的 HTTP 传输层，支持自定义超时、代理、TLS 配置和连接池管理。

#### Scenario: 默认传输配置

- **WHEN** 客户端创建时未指定自定义传输
- **THEN** SDK 使用默认传输配置
- **AND** 连接超时为 30 秒
- **AND** 保持连接存活 30 秒
- **AND** 最大空闲连接为 100

#### Scenario: 自定义 TLS 配置

- **WHEN** 客户端配置启用 HTTPS
- **THEN** 传输层使用 TLS 1.2 或更高版本
- **AND** 支持自定义 TLS 配置（如证书验证）

#### Scenario: 代理支持

- **WHEN** 环境变量设置了 HTTP_PROXY 或 HTTPS_PROXY
- **THEN** 传输层自动使用代理服务器
- **AND** 支持自定义代理配置函数

### Requirement: 请求执行器

RustFS SDK SHALL 提供统一的请求执行器，负责构建请求、签名、执行、重试和错误处理。

#### Scenario: 请求执行成功

- **WHEN** 客户端发起 API 请求
- **THEN** 执行器构建完整的 HTTP 请求
- **AND** 调用签名器对请求签名
- **AND** 通过传输层发送请求
- **AND** 解析响应并返回结果

#### Scenario: 请求自动重试

- **WHEN** 请求遇到临时性错误（如 503 Service Unavailable）
- **THEN** 执行器使用指数退避算法自动重试
- **AND** 最多重试 MaxRetries 次（默认 10 次）
- **AND** 每次重试间隔递增（100ms, 200ms, 400ms, ...）

#### Scenario: URL 构建支持路径和虚拟主机风格

- **WHEN** 客户端配置 BucketLookup 类型
- **THEN** 执行器根据配置构建正确的 URL
- **AND** 路径风格: `https://endpoint/bucket/object`
- **AND** 虚拟主机风格: `https://bucket.endpoint/object`

### Requirement: Bucket 服务接口

RustFS SDK SHALL 提供清晰的 Bucket 服务接口，支持桶的创建、删除、列表和存在性检查。

#### Scenario: 创建桶

- **WHEN** 客户端调用 `client.Bucket().Create(ctx, "my-bucket")`
- **THEN** SDK 发送 PUT Bucket 请求到服务器
- **AND** 如果成功，返回 nil
- **AND** 如果桶已存在，返回 BucketAlreadyExists 错误

#### Scenario: 创建桶时指定选项

- **WHEN** 客户端调用 `client.Bucket().Create(ctx, "my-bucket", bucket.WithRegion("us-east-1"))`
- **THEN** SDK 在请求中包含区域配置
- **AND** 桶在指定区域创建

#### Scenario: 列出所有桶

- **WHEN** 客户端调用 `client.Bucket().List(ctx)`
- **THEN** SDK 返回所有可访问的桶列表
- **AND** 每个桶包含名称和创建时间信息

#### Scenario: 检查桶是否存在

- **WHEN** 客户端调用 `client.Bucket().Exists(ctx, "my-bucket")`
- **AND** 桶存在
- **THEN** 返回 true, nil
- **WHEN** 桶不存在
- **THEN** 返回 false, nil

### Requirement: Object 服务接口

RustFS SDK SHALL 提供模块化的 Object 服务接口，将上传、下载、管理功能分离到独立的子服务中。

#### Scenario: 链式调用上传对象

- **WHEN** 客户端调用 `client.Object().Upload().Put(ctx, "bucket", "key", reader, size)`
- **THEN** SDK 通过上传服务执行上传操作
- **AND** 返回 UploadInfo 包含 ETag、大小等信息

#### Scenario: 使用函数选项配置上传

- **WHEN** 客户端调用 `client.Object().Upload().Put(ctx, "bucket", "key", reader, size, object.WithContentType("text/plain"), object.WithMetadata(meta))`
- **THEN** SDK 使用指定的 ContentType
- **AND** 附加用户元数据到对象

#### Scenario: 链式调用下载对象

- **WHEN** 客户端调用 `client.Object().Download().Get(ctx, "bucket", "key")`
- **THEN** SDK 返回 *Object（实现 io.ReadCloser）
- **AND** 客户端可以读取对象内容

### Requirement: 函数选项模式

RustFS SDK SHALL 使用函数选项模式（Functional Options Pattern）处理所有可选参数，提供清晰且可扩展的 API。

#### Scenario: 使用 WithContentType 选项

- **WHEN** 客户端传递 `object.WithContentType("application/json")`
- **THEN** 上传的对象 ContentType 被设置为 "application/json"

#### Scenario: 使用 WithMetadata 选项

- **WHEN** 客户端传递 `object.WithMetadata(map[string]string{"key": "value"})`
- **THEN** 用户元数据被附加到对象
- **AND** 服务器以 `x-amz-meta-key: value` 头存储

#### Scenario: 组合多个选项

- **WHEN** 客户端传递多个选项函数
- **THEN** 所有选项按顺序应用
- **AND** 后续选项可以覆盖前面的设置

### Requirement: 向后兼容性

RustFS SDK SHALL 保持与旧 API 的向后兼容性，提供兼容层或快捷方法。

#### Scenario: 旧的 PutObject 方法仍然工作

- **WHEN** 客户端调用旧的 `client.PutObject(ctx, "bucket", "key", reader, size, opts)`
- **THEN** SDK 将调用转发到新的 `client.Object().Upload().Put()`
- **AND** 方法标记为 Deprecated
- **AND** 功能完全正常工作

#### Scenario: 选项结构体到函数选项的转换

- **WHEN** 客户端使用旧的 `PutObjectOptions{}` 结构体
- **THEN** SDK 内部将其转换为函数选项
- **AND** 所有字段正确映射到新 API

### Requirement: 错误处理

RustFS SDK SHALL 提供一致且详细的错误处理，包括错误码、消息和请求 ID。

#### Scenario: API 错误包含完整信息

- **WHEN** 服务器返回错误响应
- **THEN** SDK 解析 XML 错误响应
- **AND** 返回包含 Code、Message、StatusCode、RequestID 的 Error
- **AND** 错误消息格式为 "Code: Message (RequestID: xxx)"

#### Scenario: 错误检查辅助函数

- **WHEN** 客户端需要检查特定错误类型
- **THEN** SDK 提供 `errors.IsNotFound(err)`, `errors.IsAccessDenied(err)` 等辅助函数
- **AND** 客户端无需手动检查错误码

## Design Notes

### 架构层次

```
Client (统一入口)
  ├── Bucket Service (桶操作)
  │     ├── Create, Delete, List, Exists
  │     ├── Config Service (配置子服务)
  │     └── Policy Service (策略子服务)
  ├── Object Service (对象操作)
  │     ├── Upload Service (上传子服务)
  │     ├── Download Service (下载子服务)
  │     └── Manage (管理方法)
  └── Internal Layers (内部实现)
        ├── Core Executor (请求执行)
        ├── Signer (签名)
        └── Transport (传输)
```

### 函数选项模式优势

1. **可扩展性**: 添加新选项无需修改函数签名
2. **可读性**: 选项名称自解释
3. **向后兼容**: 新选项不影响现有代码
4. **灵活性**: 客户端只需指定需要的选项

### 性能考虑

- 连接池复用，减少连接建立开销
- 支持并发上传（分片上传）
- 缓存桶位置信息，减少重定向
- 可配置的重试策略，平衡可靠性和延迟
