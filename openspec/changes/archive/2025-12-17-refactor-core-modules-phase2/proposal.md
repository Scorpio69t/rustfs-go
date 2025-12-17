# Change: 实施第二阶段核心模块重构

## Why

当前 RustFS Go SDK 已完成第一阶段基础架构搭建（types/, errors/, internal/core/, internal/signer/, internal/cache/），但核心功能模块仍在 `old/` 目录中。为了完成从 MinIO 风格到 RustFS 品牌的完全转变，需要实施第二阶段：核心模块实现和 Bucket/Object 服务接口定义。

## What Changes

- ✅ 完善 `internal/signer/` 包的 V4 和 V2 签名实现
- ✅ 实现 `internal/transport/` 包的 HTTP 传输层
- ✅ 完善 `internal/core/executor.go` 中的请求执行逻辑
- ✅ 创建 `bucket/` 服务接口和基础实现
- ✅ 创建 `object/` 服务接口和基础实现
- ✅ 实现 Bucket 基础操作（Create, Delete, List, Exists）
- ✅ 实现 Object 上传服务的函数选项模式
- ✅ 更新 Client 以支持新的服务接口

## Impact

- **影响的规范**: `core-api` (新增)
- **影响的代码**:
  - `internal/signer/v4.go` - V4 签名完整实现
  - `internal/signer/v2.go` - V2 签名完整实现
  - `internal/transport/transport.go` - HTTP 传输层
  - `internal/core/executor.go` - 请求执行和重试
  - `bucket/service.go` - Bucket 服务接口
  - `bucket/*.go` - Bucket 基础操作实现
  - `object/service.go` - Object 服务接口
  - `object/options.go` - Object 选项函数
  - `client.go` - 更新客户端入口
  - `examples/` - 更新示例代码

## Breaking Changes

无。此阶段是增量式重构，保持向后兼容。旧的 API 通过兼容层继续工作。

## Dependencies

- 依赖第一阶段完成的基础架构
- 参考 `old/` 目录中的现有实现
- 遵循 `REFACTORING_PLAN.md` 中的架构设计

## Timeline

预计完成时间：7-10 个工作日

- Day 1-2: 完善签名器和传输层
- Day 3-4: 完善核心执行器
- Day 5-6: 实现 Bucket 服务
- Day 7-8: 实现 Object 服务框架
- Day 9-10: 测试和示例更新
