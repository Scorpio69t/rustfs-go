# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.0.0] - 2025-01-XX

### Added

#### 核心功能
- ✅ **完整的 S3 API 兼容性** - 支持所有标准 S3 操作
- ✅ **模块化设计** - 分离 Bucket 和 Object 服务，提供更清晰的 API
- ✅ **AWS Signature V4/V2** - 完整的签名支持，包括流式签名
- ✅ **健康检查** - 内置服务健康检查，支持重试机制
- ✅ **HTTP 追踪** - 请求性能追踪和调试支持
- ✅ **传输层优化** - 可配置的 HTTP 传输，支持连接池、超时、TLS 等

#### Bucket 操作
- 创建/删除存储桶（支持区域、对象锁定、强制删除等选项）
- 列出所有存储桶
- 检查存储桶是否存在
- 获取存储桶位置

#### Object 操作
- 上传/下载对象（支持元数据、标签、存储类等）
- 获取对象信息和元数据
- 删除对象
- 列出对象（支持前缀、递归、最大数量等过滤）
- 复制对象（支持元数据替换、条件复制等）

#### 分片上传
- 初始化分片上传
- 上传分片
- 完成分片上传
- 取消分片上传

#### 高级功能
- 流式签名支持 (AWS Signature V4 Chunked Upload)
- 位置缓存优化（减少重复的 GetBucketLocation 请求）
- 智能重试机制（支持网络错误和特定 HTTP 状态码重试）
- 自动路径风格选择（IP 地址自动使用 Path-style）

### Technical Details

#### 新增模块
- `internal/signer/` - AWS 签名实现（V4/V2/流式）
- `internal/transport/` - HTTP 传输层和追踪
- `internal/core/` - 核心执行器和健康检查
- `bucket/` - 存储桶服务
- `object/` - 对象服务

#### 测试覆盖
- 单元测试覆盖率 > 60%
- 总计 150+ 个测试用例
- 所有核心功能经过完整测试

#### 示例代码
- Bucket 操作示例 (`examples/rustfs/bucketops.go`)
- Object 操作示例 (`examples/rustfs/objectops.go`)
- 分片上传示例 (`examples/rustfs/multipart.go`)
- 健康检查示例 (`examples/rustfs/health.go`)
- HTTP 追踪示例 (`examples/rustfs/trace.go`)

### Changed
- 采用新的模块化 API 设计
- 使用选项函数模式提供更灵活的配置
- 改进错误处理和类型定义

### Dependencies
- Go 1.25+
- github.com/google/uuid v1.6.0
- golang.org/x/net v0.25.0

### Documentation
- 完整的 README（中英文）
- API 文档注释
- 详细的使用示例
- OpenSpec 规范文档

---

## 未来计划

### v1.1.0 (计划中)
- [ ] 预签名 URL 支持
- [ ] 对象标签管理
- [ ] 存储桶策略管理
- [ ] 生命周期管理
- [ ] 服务端加密支持

### v1.2.0 (计划中)
- [ ] 对象版本控制
- [ ] 跨区域复制
- [ ] 事件通知
- [ ] 访问日志记录

---

## 贡献

欢迎提交 Issue 和 Pull Request！

请查看 [CONTRIBUTING.md](CONTRIBUTING.md) 了解详细的贡献指南。

## 许可证

Apache License 2.0 - 详见 [LICENSE](LICENSE) 文件
