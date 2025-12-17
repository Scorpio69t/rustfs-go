# 第二阶段核心模块重构任务清单

## 1. 完善签名器模块 (internal/signer/)

- [x] 1.1 完善 V4 签名器实现
  - [x] 1.1.1 实现完整的规范请求生成 (createCanonicalRequest)
  - [x] 1.1.2 实现完整的待签名字符串生成 (createStringToSign)
  - [x] 1.1.3 实现签名密钥派生 (deriveSigningKey)
  - [x] 1.1.4 实现 Authorization 头构建 (buildAuthorizationHeader)
  - [x] 1.1.5 实现预签名 URL 生成 (Presign)
  - [ ] 1.1.6 支持流式签名 (Chunk Signature) - 后续实现

- [x] 1.2 完善 V2 签名器实现
  - [x] 1.2.1 实现 V2 签名算法 (Sign)
  - [x] 1.2.2 实现 V2 预签名 URL (Presign)
  - [x] 1.2.3 实现规范化头部和资源处理

- [x] 1.3 签名器单元测试
  - [x] 1.3.1 V4 签名测试用例 (16 个测试，全部通过)
  - [x] 1.3.2 V2 签名测试用例 (11 个测试，全部通过)
  - [x] 1.3.3 预签名 URL 测试用例 (V4/V2 各3个，全部通过)

## 2. 实现传输层 (internal/transport/)

- [x] 2.1 创建 transport.go
  - [x] 2.1.1 实现 DefaultTransport 函数
  - [x] 2.1.2 实现 TransportOptions 结构体
  - [x] 2.1.3 实现 NewTransport 函数
  - [x] 2.1.4 配置 TLS 和连接池参数
  - [x] 2.1.5 实现 NewHTTPClient 辅助函数
  - [x] 2.1.6 支持自定义 CA 证书（SSL_CERT_FILE）

- [ ] 2.2 创建 trace.go (可选 - 后续实现)
  - [ ] 2.2.1 实现 HTTP 请求追踪
  - [ ] 2.2.2 添加调试日志支持

- [x] 2.3 传输层单元测试
  - [x] 2.3.1 DefaultTransport 测试（2 个测试用例）
  - [x] 2.3.2 NewTransport 测试（3 个测试用例）
  - [x] 2.3.3 TLS 配置测试
  - [x] 2.3.4 代理配置测试
  - [x] 2.3.5 连接池配置测试
  - [x] 2.3.6 NewHTTPClient 测试（3 个测试用例）
  - [x] 2.3.7 性能基准测试（2 个 benchmark）

## 3. 完善核心执行器 (internal/core/)

- [x] 3.1 完善 executor.go
  - [x] 3.1.1 实现 makeTargetURL 方法（路径风格 vs 虚拟主机风格）
  - [x] 3.1.2 实现 signRequest 方法（集成签名器）
  - [x] 3.1.3 完善 Execute 方法的错误处理
  - [x] 3.1.4 完善 shouldRetry 和 shouldRetryResponse 逻辑
  - [x] 3.1.5 实现辅助方法（isVirtualHostStyleRequest, encodePath 等）
  - [ ] 3.1.6 实现健康检查逻辑（可选，后续实现）

- [x] 3.2 完善 response.go（已有基础实现）
  - [x] 3.2.1 完善 ParseObjectInfo 方法
  - [x] 3.2.2 完善 ParseUploadInfo 方法
  - [x] 3.2.3 添加更多响应解析辅助方法

- [x] 3.3 核心执行器集成测试
  - [x] 46 个测试用例全部通过
  - [x] 包含 URL 构建、重试逻辑、执行流程等完整测试

## 4. 实现 Bucket 服务

- [x] 4.1 创建服务接口 (bucket/service.go)
  - [x] 4.1.1 定义 Service 接口（Create, Delete, Exists, List, GetLocation）
  - [x] 4.1.2 定义 CreateOption 和 DeleteOption 函数类型
  - [x] 4.1.3 实现选项函数（WithRegion, WithObjectLocking, WithForceCreate, WithForceDelete）

- [x] 4.2 实现基础操作
  - [x] 4.2.1 实现 Create (bucket/create.go)
    - [x] 桶名验证
    - [x] 区域配置支持
    - [x] 对象锁定支持
    - [x] CreateBucketConfiguration XML 生成
    - [x] 位置缓存更新
  - [x] 4.2.2 实现 Delete (bucket/delete.go)
    - [x] 桶名验证
    - [x] 强制删除选项
    - [x] 位置缓存清理
  - [x] 4.2.3 实现 Exists (bucket/exists.go)
    - [x] HEAD 请求检查
    - [x] NoSuchBucket 错误处理
    - [x] 404 状态码处理
  - [x] 4.2.4 实现 List (bucket/list.go)
    - [x] ListAllMyBucketsResult XML 解析
    - [x] 返回桶信息列表
  - [x] 4.2.5 实现 GetLocation (bucket/list.go)
    - [x] 位置缓存查询
    - [x] LocationConstraint XML 解析
    - [x] 默认 us-east-1 处理

- [x] 4.3 实现 bucket 服务入口 (bucket/bucket.go)
  - [x] 4.3.1 实现 bucketService 结构体
  - [x] 4.3.2 实现 NewService 构造函数
  - [x] 4.3.3 实现选项应用函数
  - [x] 4.3.4 实现桶名验证函数

- [x] 4.4 辅助功能 (bucket/utils.go, bucket/errors.go)
  - [x] SHA256 哈希计算
  - [x] 响应关闭工具
  - [x] 错误解析工具
  - [x] 自定义错误类型

- [x] 4.5 Bucket 服务单元测试 (bucket/bucket_test.go)
  - [x] TestCreate（4 个测试用例）
  - [x] TestDelete（3 个测试用例）
  - [x] TestExists（3 个测试用例）
  - [x] TestList（2 个测试用例）
  - [x] TestGetLocation（3 个测试用例）
  - [x] TestValidateBucketName（6 个测试用例）
  - [x] TestApplyCreateOptions（4 个测试用例）
  - [x] TestApplyDeleteOptions（2 个测试用例）
  - [x] 性能基准测试（2 个 benchmark）

## 5. 实现 Object 服务框架

- [x] 5.1 创建服务接口 (object/service.go)
  - [x] 5.1.1 定义 Service 接口（Put, Get, Stat, Delete, List, Copy）
  - [x] 5.1.2 定义选项函数类型（PutOption, GetOption, StatOption, DeleteOption, ListOption, CopyOption）

- [x] 5.2 实现选项函数 (object/options.go)
  - [x] 5.2.1 实现 PutOptions 结构体和选项函数
    - [x] WithContentType, WithContentEncoding, WithContentDisposition
    - [x] WithUserMetadata, WithUserTags, WithStorageClass
    - [x] WithPartSize
  - [x] 5.2.2 实现 GetOptions 结构体和选项函数
    - [x] WithGetRange, WithVersionID
  - [x] 5.2.3 实现 ListOptions 结构体和选项函数
    - [x] WithListPrefix, WithListRecursive, WithListMaxKeys
  - [x] 5.2.4 实现 DeleteOptions, StatOptions, CopyOptions 结构体
    - [x] WithVersionID（通用）, WithCopySourceVersionID, WithCopyMetadata

- [x] 5.3 创建 Object 服务入口 (object/object.go)
  - [x] 5.3.1 实现 objectService 结构体
  - [x] 5.3.2 实现 NewService 构造函数
  - [x] 5.3.3 实现所有接口方法的框架（标记为 TODO）
    - [x] Put, Get, Stat, Delete, List, Copy
  - [x] 5.3.4 实现选项应用函数
    - [x] applyPutOptions, applyGetOptions, applyStatOptions
    - [x] applyDeleteOptions, applyListOptions, applyCopyOptions
  - [x] 5.3.5 实现验证函数
    - [x] validateBucketName, validateObjectName

- [x] 5.4 辅助功能 (object/errors.go, types/object.go)
  - [x] 错误定义（ErrInvalidBucketName, ErrInvalidObjectName, ErrNotImplemented）
  - [x] 添加 CopyInfo 类型到 types/object.go

- [x] 5.5 Object 服务单元测试 (object/object_test.go)
  - [x] TestValidateBucketName（4 个测试用例）
  - [x] TestValidateObjectName（3 个测试用例）
  - [x] TestApplyPutOptions（3 个测试用例）
  - [x] TestApplyGetOptions（2 个测试用例）
  - [x] TestApplyListOptions（3 个测试用例）
  - [x] TestNewService（1 个测试用例）
  - [x] 性能基准测试（2 个 benchmark）

## 6. 更新客户端入口

- [x] 6.1 更新 client.go
  - [x] 6.1.1 添加 bucketService 和 objectService 字段
  - [x] 6.1.2 实现 Bucket() 方法
  - [x] 6.1.3 实现 Object() 方法
  - [x] 6.1.4 保留旧的快捷方法（标记为 Deprecated）
  - [x] 6.1.5 实现 IP 地址自动使用路径风格桶查找
  - [x] 6.1.6 创建向后兼容层（compat.go）

- [x] 6.2 更新 options.go
  - [x] 6.2.1 确保所有配置选项正确传递给子服务
  - [x] 6.2.2 添加选项验证和默认值设置

- [x] 6.3 客户端单元测试（client_test.go）
  - [x] TestNew（4 个测试用例，全部通过）
  - [x] TestClientMethods（6 个测试用例，全部通过）
  - [x] TestBackwardCompatibility（4 个测试用例，全部通过）
  - [x] TestParseEndpointURL（4 个测试用例，全部通过）
  - [x] 性能基准测试（3 个 benchmark）

## 7. 文档和示例

- [ ] 7.1 更新 examples/
  - [ ] 7.1.1 更新 examples/rustfs/bucketops.go 使用新 API
  - [ ] 7.1.2 更新 examples/rustfs/objectops.go 使用新 API
  - [ ] 7.1.3 确保示例可以正常运行

- [ ] 7.2 更新 README.md
  - [ ] 7.2.1 添加新 API 使用示例
  - [ ] 7.2.2 更新快速开始部分

## 8. 测试验证

- [ ] 8.1 单元测试
  - [ ] 8.1.1 运行所有单元测试确保通过
  - [ ] 8.1.2 确保测试覆盖率 > 60%

- [ ] 8.2 集成测试
  - [ ] 8.2.1 测试与真实 RustFS 服务器的交互
  - [ ] 8.2.2 测试向后兼容性

- [ ] 8.3 构建验证
  - [ ] 8.3.1 运行 `go build ./...` 确保编译通过
  - [ ] 8.3.2 运行 `go mod tidy` 确保依赖正确

## 验收标准

- ✅ 所有代码编译通过，无 lint 错误
- ✅ 核心功能单元测试通过
- ✅ 示例代码可以成功运行
- ✅ 新 API 与旧 API 共存，保持向后兼容
- ✅ 文档更新完整，清晰说明新旧 API 的使用方式
