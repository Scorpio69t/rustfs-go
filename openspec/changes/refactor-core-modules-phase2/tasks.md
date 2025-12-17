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

- [ ] 3.1 完善 executor.go
  - [ ] 3.1.1 实现 makeTargetURL 方法（路径风格 vs 虚拟主机风格）
  - [ ] 3.1.2 实现 signRequest 方法（集成签名器）
  - [ ] 3.1.3 完善 Execute 方法的错误处理
  - [ ] 3.1.4 完善 shouldRetry 和 shouldRetryResponse 逻辑
  - [ ] 3.1.5 实现健康检查逻辑

- [ ] 3.2 完善 response.go
  - [ ] 3.2.1 完善 ParseObjectInfo 方法
  - [ ] 3.2.2 完善 ParseUploadInfo 方法
  - [ ] 3.2.3 添加更多响应解析辅助方法

- [ ] 3.3 核心执行器集成测试

## 4. 实现 Bucket 服务

- [ ] 4.1 创建服务接口 (bucket/service.go)
  - [ ] 4.1.1 定义 BucketService 接口
  - [ ] 4.1.2 定义 ConfigService 接口
  - [ ] 4.1.3 定义 PolicyService 接口
  - [ ] 4.1.4 定义选项函数类型

- [ ] 4.2 实现基础操作
  - [ ] 4.2.1 实现 Create (bucket/create.go)
  - [ ] 4.2.2 实现 Delete (bucket/delete.go)
  - [ ] 4.2.3 实现 Exists (bucket/exists.go)
  - [ ] 4.2.4 实现 List (bucket/list.go)

- [ ] 4.3 实现 bucket 服务入口 (bucket/bucket.go)
  - [ ] 4.3.1 实现 bucketService 结构体
  - [ ] 4.3.2 实现 NewBucketService 构造函数
  - [ ] 4.3.3 实现接口方法

- [ ] 4.4 Bucket 服务单元测试

## 5. 实现 Object 服务框架

- [ ] 5.1 创建服务接口 (object/service.go)
  - [ ] 5.1.1 定义 ObjectService 接口
  - [ ] 5.1.2 定义 UploadService 接口
  - [ ] 5.1.3 定义 DownloadService 接口
  - [ ] 5.1.4 定义 MultipartService 接口

- [ ] 5.2 实现选项函数 (object/options.go)
  - [ ] 5.2.1 实现 PutOption 函数集
  - [ ] 5.2.2 实现 GetOption 函数集
  - [ ] 5.2.3 实现 ListOption 函数集
  - [ ] 5.2.4 实现 DeleteOption、StatOption 等

- [ ] 5.3 创建 Object 服务入口 (object/object.go)
  - [ ] 5.3.1 实现 objectService 结构体
  - [ ] 5.3.2 实现 NewObjectService 构造函数
  - [ ] 5.3.3 实现 Upload() 方法返回 UploadService
  - [ ] 5.3.4 实现 Download() 方法返回 DownloadService

- [ ] 5.4 实现上传服务基础框架 (object/upload/)
  - [ ] 5.4.1 创建 upload.go 服务入口
  - [ ] 5.4.2 实现 uploadService 结构体
  - [ ] 5.4.3 实现 Put 方法的基础框架（调用 core.Execute）

## 6. 更新客户端入口

- [ ] 6.1 更新 client.go
  - [ ] 6.1.1 添加 bucketService 和 objectService 字段
  - [ ] 6.1.2 实现 Bucket() 方法
  - [ ] 6.1.3 实现 Object() 方法
  - [ ] 6.1.4 保留旧的快捷方法（标记为 Deprecated）

- [ ] 6.2 更新 options.go
  - [ ] 6.2.1 确保所有配置选项正确传递给子服务

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
