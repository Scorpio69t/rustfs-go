# 项目上下文

## Purpose

RustFS Go SDK 是一个专为 RustFS 对象存储系统设计的 Go 语言客户端库，完全兼容 S3 协议。

**项目主要目标：**
- 🎯 建立独立的品牌身份，摆脱 MinIO 依赖痕迹
- 🏗️ 采用现代化的 Go 项目结构，提高代码可维护性
- 🔧 提供更清晰、更易用的 API 设计
- 📦 优化包组织，实现职责分离和接口抽象
- ✅ 保持与 S3 API 的完全兼容性

**当前阶段：**
项目正在进行大规模重构，从旧的扁平化结构（30+ 个 api-*.go 文件）迁移到模块化的新架构。重构参考基准代码位于 `old/` 目录，已完成第一阶段基础架构搭建，正在进行核心模块实现。

## 技术栈

- **语言**: Go 1.25+
- **HTTP 客户端**: 标准库 `net/http`
- **加密签名**: 标准库 `crypto/hmac`, `crypto/sha256`
- **XML 处理**: 标准库 `encoding/xml`
- **JSON 处理**: 标准库 `encoding/json`
- **测试框架**: 标准库 `testing`
- **依赖管理**: Go Modules

**主要依赖：**
- `github.com/minio/md5-simd` - 高性能 MD5 计算
- `github.com/minio/crc64nvme` - CRC64 校验
- `github.com/dustin/go-humanize` - 人类可读格式
- `golang.org/x/net` - 网络工具
- `golang.org/x/crypto` - 加密工具

## 项目约定

### 代码风格

**命名规范：**
- 包名：小写单词，如 `bucket`, `object`, `presign`
- 接口名：动词+Service，如 `BucketService`, `UploadService`
- 结构体：名词/形容词+名词，如 `BucketInfo`, `UploadOptions`
- 方法：动词开头，如 `Create`, `Delete`, `List`
- 选项函数：With+属性，如 `WithRegion`, `WithMetadata`
- 错误码：ErrCode+描述，如 `ErrCodeNoSuchBucket`

**代码组织：**
- 所有公共 API 必须有 GoDoc 注释
- 使用 `context.Context` 作为第一个参数
- 使用函数选项模式处理可选参数
- 错误信息要清晰且可操作
- 避免导出内部实现细节（使用 `internal/` 包）

**格式化：**
- 使用 `gofmt` 格式化代码
- 使用 `golangci-lint` 进行代码检查
- 遵循 Go 官方代码规范

### 架构模式

**分层架构：**
```
用户层 (Client)
  ↓
功能模块层 (bucket/, object/)
  ↓
核心层 (internal/core/, internal/signer/, internal/transport/)
  ↓
基础设施层 (pkg/credentials/, errors/, types/)
```

**设计原则：**
1. **接口抽象** - 核心功能基于接口，便于测试和扩展
2. **职责分离** - 每个包只负责单一功能
3. **依赖注入** - 通过 Options 模式注入依赖
4. **错误处理** - 统一的错误类型和错误码
5. **向后兼容** - 提供兼容层支持旧 API

**模块组织：**
- `bucket/` - 存储桶操作模块
  - `bucket.go` - 桶服务入口
  - `config/` - 桶配置子模块（CORS、加密、生命周期等）
  - `policy/` - 桶策略子模块
- `object/` - 对象操作模块
  - `upload/` - 上传子模块（简单、分片、流式）
  - `download/` - 下载子模块
  - `manage/` - 管理子模块（复制、删除、标签等）
  - `presign/` - 预签名子模块
- `internal/` - 内部实现（不导出）
  - `core/` - 核心请求处理
  - `signer/` - 签名处理
  - `transport/` - 传输层
  - `cache/` - 缓存
- `pkg/` - 公共工具包（可独立使用）
  - `credentials/` - 凭证管理
  - `encrypt/` - 加密工具
  - `lifecycle/` - 生命周期配置
  - `policy/` - 策略定义
- `types/` - 公共类型定义
- `errors/` - 错误定义

### 测试策略

**测试要求：**
- 单元测试覆盖率 > 70%
- 关键路径必须有集成测试
- 使用 table-driven 测试风格
- Mock 外部依赖（HTTP 客户端、凭证服务等）

**测试组织：**
- 每个包都有对应的 `*_test.go` 文件
- 测试文件与源文件在同一目录
- 集成测试放在 `test/` 目录（如需要）

**测试命名：**
- 测试函数：`TestFunctionName`
- 基准测试：`BenchmarkFunctionName`
- 示例函数：`ExampleFunctionName`

### Git 工作流

**分支策略：**
- `main` - 主分支，稳定代码
- `develop` - 开发分支
- `feature/*` - 功能分支
- `refactor/*` - 重构分支

**提交规范：**
使用约定式提交格式：
```
<type>(<scope>): <subject>

<body>

<footer>
```

**提交类型：**
- `feat(module):` - 添加新功能
- `fix(module):` - 修复 bug
- `refactor(module):` - 重构代码
- `docs(module):` - 更新文档
- `test(module):` - 添加测试
- `chore:` - 其他变更（构建、工具等）

**示例：**
```
feat(bucket): 添加桶生命周期配置功能

实现 SetLifecycle、GetLifecycle、DeleteLifecycle 方法
支持完整的生命周期规则配置

Closes #123
```

## 领域上下文

**RustFS 对象存储：**
- RustFS 是一个高性能的分布式对象存储系统
- 完全兼容 Amazon S3 API
- 支持标准 S3 操作：存储桶管理、对象操作、多部分上传等
- 支持预签名 URL、对象标签、生命周期管理等高级功能

**S3 API 兼容性：**
- 所有 API 调用必须符合 S3 协议规范
- 使用 AWS Signature Version 4 进行请求签名
- 支持虚拟主机风格和路径风格的 URL
- 错误响应遵循 S3 错误格式

**重构上下文：**
- **基准代码**: `old/` 目录包含重构前的完整代码，作为功能参考
- **重构目标**: 从扁平化结构（30+ api-*.go 文件）迁移到模块化架构
- **当前进度**:
  - ✅ 第一阶段：基础架构搭建（types/, errors/, internal/core/）
  - 🔄 第二阶段：核心模块实现（进行中）
  - ⏳ 第三阶段：Bucket 模块实现
  - ⏳ 第四阶段：Object 模块实现
  - ⏳ 第五阶段：兼容层和测试

**关键文档：**
- `REFACTORING_PLAN.md` - 详细的重构方案和架构设计
- `IMPLEMENTATION_TODO.md` - 分阶段实施计划和任务清单
- `old/` - 重构前的基准代码，包含所有功能实现

## 重要约束

**兼容性约束：**
- 必须保持与 S3 API 的完全兼容
- 必须支持所有标准 S3 操作
- 错误响应格式必须符合 S3 规范

**性能约束：**
- 支持大文件流式上传/下载
- 支持并发操作
- 最小化内存占用

**安全约束：**
- 凭证信息不得记录在日志中
- 支持 HTTPS 传输
- 实现正确的请求签名验证

**重构约束：**
- 重构过程中保持功能完整性
- 提供向后兼容的 API（标记为 deprecated）
- 所有重构变更必须通过测试验证

## 外部依赖

**核心依赖：**
- `github.com/minio/md5-simd` - 高性能 MD5 哈希计算
- `github.com/minio/crc64nvme` - CRC64 校验和计算
- `golang.org/x/net` - 网络工具和 HTTP/2 支持
- `golang.org/x/crypto` - 加密算法支持

**工具依赖：**
- `github.com/dustin/go-humanize` - 人类可读格式转换
- `github.com/google/uuid` - UUID 生成
- `github.com/rs/xid` - 分布式 ID 生成

**参考实现：**
- MinIO Go SDK (`github.com/minio/minio-go`) - 主要参考实现，但需要去除所有 MinIO 品牌痕迹
- AWS S3 API 文档 - API 规范参考

**服务依赖：**
- RustFS 对象存储服务器 - 目标服务
- 支持 S3 兼容的对象存储服务（用于测试）
