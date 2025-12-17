# pkg 包重构技术规范

## 概述

本规范定义了 `pkg/` 目录下各个包的重构标准和实现细节。

## 1. 版权声明标准

### 1.1 新版权声明模板

所有重构后的文件应使用以下版权声明：

```go
/*
 * RustFS Go SDK
 * Copyright 2025 RustFS Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * This code is derived from MinIO Go SDK (https://github.com/minio/minio-go)
 * and has been substantially modified for RustFS compatibility.
 */
```

### 1.2 致谢说明

- 保留对 MinIO 的致谢（最后一行）
- 明确说明代码已被"实质性修改"
- 强调 RustFS 兼容性

## 2. pkg/credentials 包规范

### 2.1 包结构

```
pkg/credentials/
├── credentials.go       # 核心接口和类型
├── static.go           # 静态凭证提供者
├── env.go              # 环境变量凭证提供者
├── file.go             # 文件凭证提供者
├── chain.go            # 凭证链
├── iam.go              # IAM 角色凭证
├── sts_assume_role.go  # STS AssumeRole（可选）
├── sts_web_identity.go # STS Web Identity（可选）
├── error.go            # 错误类型
├── doc.go              # 包文档
└── *_test.go           # 测试文件
```

### 2.2 核心接口

```go
// Provider 凭证提供者接口
type Provider interface {
    // Retrieve 获取凭证
    Retrieve() (Value, error)

    // IsExpired 检查凭证是否过期
    IsExpired() bool
}

// Value 凭证值
type Value struct {
    AccessKeyID     string
    SecretAccessKey string
    SessionToken    string
    SignerType      SignatureType
}

// Credentials 凭证管理器
type Credentials struct {
    provider Provider
    // ... 其他字段
}
```

### 2.3 环境变量支持

#### 2.3.1 AWS 环境变量
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `AWS_SESSION_TOKEN`

#### 2.3.2 RustFS 环境变量（向后兼容 MinIO）
- `RUSTFS_ACCESS_KEY` / `MINIO_ACCESS_KEY`
- `RUSTFS_SECRET_KEY` / `MINIO_SECRET_KEY`
- `RUSTFS_ROOT_USER` / `MINIO_ROOT_USER`
- `RUSTFS_ROOT_PASSWORD` / `MINIO_ROOT_PASSWORD`

### 2.4 命名规范

#### 2.4.1 类型命名
- ❌ `EnvMinio` -> ✅ `EnvRustFS` 或 `EnvProvider`
- ❌ `FileMinioClient` -> ✅ `FileRustFS` 或 `FileProvider`
- ✅ `EnvAWS` -> 保持不变（AWS 标准）

#### 2.4.2 函数命名
- ❌ `NewEnvMinio()` -> ✅ `NewEnvRustFS()` 或 `NewEnv()`
- ❌ `NewFileMinioClient()` -> ✅ `NewFileRustFS()` 或 `NewFile()`

### 2.5 STS 功能决策

#### 2.5.1 保留的 STS 功能
- ✅ `AssumeRole` - AWS 标准功能
- ✅ `WebIdentity` - AWS 标准功能

#### 2.5.2 可选的 STS 功能
- ❓ `ClientGrants` - MinIO 扩展，评估 RustFS 是否支持
- ❓ `CustomIdentity` - MinIO 扩展，评估 RustFS 是否支持
- ❓ `LDAPIdentity` - MinIO 扩展，评估 RustFS 是否支持
- ❓ `TLSIdentity` - MinIO 扩展，评估 RustFS 是否支持

## 3. pkg/signer 包规范

### 3.1 处理方案

**推荐方案**: 删除 `pkg/signer`

**理由**:
1. 已有完整的 `internal/signer` 实现
2. 签名逻辑应该是内部实现细节
3. 避免代码重复

**如果需要公共 API**:
- 在 `pkg/credentials` 中提供签名类型定义
- 不暴露签名实现细节

### 3.2 迁移计划

如果有外部代码依赖 `pkg/signer`:
1. 提供迁移指南
2. 标记为 Deprecated
3. 在下一个主版本中移除

## 4. pkg/s3utils 包规范

### 4.1 保留的功能

```go
// EncodePath URL 路径编码
func EncodePath(path string) string

// QueryEncode URL 查询参数编码
func QueryEncode(v url.Values) string

// IsValidBucketName 验证桶名
func IsValidBucketName(bucketName string) error

// IsValidObjectName 验证对象名
func IsValidObjectName(objectName string) error
```

### 4.2 重构要点

1. **更新版权声明**
2. **简化实现**: 移除不必要的复杂逻辑
3. **S3 兼容性**: 确保符合 S3 规范
4. **RustFS 特性**: 添加 RustFS 特定验证（如有）

### 4.3 验证规则

#### 4.3.1 桶名规则（S3 标准）
- 长度: 3-63 字符
- 字符: 小写字母、数字、连字符
- 不能以连字符开头或结尾
- 不能包含连续的连字符
- 不能是 IP 地址格式

#### 4.3.2 对象名规则（S3 标准）
- 长度: 1-1024 字符
- 支持 UTF-8 字符
- 某些字符需要 URL 编码

## 5. pkg/set 包规范

### 5.1 评估标准

检查以下使用场景：
1. 是否在核心代码中使用？
2. 是否在公共 API 中暴露？
3. 是否有替代方案（如 `map[string]struct{}`）？

### 5.2 处理方案

**方案 A**: 删除（如果未使用或有简单替代）
**方案 B**: 移至 `internal/set`（如果仅内部使用）
**方案 C**: 重构并保留（如果是公共 API 的一部分）

## 6. 代码风格规范

### 6.1 注释规范

```go
// Good: 清晰描述功能
// NewStaticV4 creates a new static credentials provider with V4 signature.

// Bad: 引用 MinIO
// NewStaticV4 creates MinIO credentials with V4 signature.
```

### 6.2 错误消息规范

```go
// Good: 通用描述
return fmt.Errorf("failed to retrieve credentials: %w", err)

// Bad: MinIO 特定
return fmt.Errorf("MinIO STS request failed: %w", err)
```

### 6.3 示例代码规范

```go
// Good: RustFS 示例
// Example:
//   creds := credentials.NewStaticV4("access-key", "secret-key", "")
//   client, err := rustfs.New("rustfs.example.com", &rustfs.Options{
//       Credentials: creds,
//   })

// Bad: MinIO 示例
// Example:
//   client, err := minio.New("play.min.io", &minio.Options{...})
```

## 7. 测试规范

### 7.1 测试文件命名

- 单元测试: `*_test.go`
- 基准测试: 包含在单元测试文件中
- 示例测试: `Example*` 函数

### 7.2 测试覆盖率要求

- 核心功能: > 80%
- 辅助功能: > 60%
- 总体: > 60%

### 7.3 测试用例要求

每个公共函数至少包含：
1. 正常情况测试
2. 错误情况测试
3. 边界条件测试

## 8. 迁移兼容性

### 8.1 向后兼容性

**必须保持兼容**:
- 公共 API 接口
- 环境变量名（支持 MINIO_* 作为别名）
- 配置文件格式

**可以变更**:
- 内部实现
- 私有函数
- 错误消息格式

### 8.2 弃用策略

如果需要移除某些功能：
1. 标记为 `Deprecated`
2. 在文档中说明替代方案
3. 在下一个主版本中移除

## 9. 性能要求

### 9.1 基准测试

每个核心功能应提供基准测试：

```go
func BenchmarkNewStaticV4(b *testing.B) {
    for i := 0; i < b.N; i++ {
        NewStaticV4("access", "secret", "")
    }
}
```

### 9.2 性能目标

- 凭证获取: < 1μs
- 凭证验证: < 100ns
- URL 编码: < 1μs

## 10. 安全要求

### 10.1 敏感信息处理

- ❌ 不在日志中输出密钥
- ❌ 不在错误消息中包含密钥
- ✅ 使用安全的内存清理（如需要）

### 10.2 凭证存储

- ✅ 支持环境变量
- ✅ 支持配置文件（权限检查）
- ✅ 支持 IAM 角色
- ❌ 不在代码中硬编码凭证

## 附录

### A. 包依赖关系

```
pkg/credentials
├── 依赖: 无外部依赖（标准库）
└── 被依赖: client.go, internal/core/executor.go

pkg/s3utils
├── 依赖: 无外部依赖（标准库）
└── 被依赖: bucket/, object/, internal/core/

pkg/signer (待删除)
├── 依赖: 无
└── 被依赖: 无（已被 internal/signer 替代）

pkg/set (待评估)
├── 依赖: github.com/tinylib/msgp
└── 被依赖: 待确认
```

### B. 重构检查清单

每个文件重构后检查：
- [ ] 版权声明已更新
- [ ] MinIO 引用已移除
- [ ] 文档注释已更新
- [ ] 测试用例已更新
- [ ] 测试全部通过
- [ ] 代码格式正确
- [ ] 无 linter 警告
