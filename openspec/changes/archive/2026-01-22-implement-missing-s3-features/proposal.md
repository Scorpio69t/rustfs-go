# Proposal: Complete Missing S3 API Features

## Metadata
- **Status**: Draft
- **Type**: Feature Enhancement
- **Priority**: Medium
- **Created**: 2026-01-21
- **Author**: AI Assistant

## Why

当前 RustFS Go SDK 已实现核心 S3 功能，但仍缺少关键企业级特性：

1. **安全与合规需求**：
   - 无法使用服务端加密（SSE-S3/SSE-C/SSE-KMS）保护数据
   - 缺少对象锁定和法律保留功能，无法满足合规要求（如 SEC 17a-4）
   - ACL 支持缺失，难以实现细粒度权限控制

2. **高可用性需求**：
   - 缺少跨区域复制功能，无法实现灾备
   - 缺少事件通知机制，无法构建事件驱动架构
   - CORS 配置缺失，限制了浏览器端直接访问

3. **功能完整性**：
   - 71 个 MinIO 示例中约 36 个因功能缺失无法迁移
   - 高级功能如对象组合、Select 查询、归档恢复未实现
   - Post Policy 缺失导致浏览器直传场景难以实现

4. **用户反馈**：
   - 企业用户要求完整的 S3 API 兼容性
   - 开发者需要完整的参考示例和最佳实践
   - 生产环境需要与 AWS S3 和 MinIO 完全兼容

**不解决的影响**：
- 无法满足企业级安全和合规要求
- 高可用架构无法实现
- SDK 采用率受限
- 与竞品（AWS SDK, MinIO SDK）相比功能不足

## Overview

迁移 MinIO SDK 中尚未实现的 S3 API 功能到新的 RustFS Go SDK 架构。当前 SDK 已实现核心功能（存储桶、对象、版本控制、标签、策略、生命周期），但仍有以下高级功能未迁移：

- CORS 配置
- 存储桶加密配置
- 对象加密（SSE-S3, SSE-C, SSE-KMS）
- 对象锁定和保留
- 存储桶复制
- 事件通知
- ACL 配置
- 对象追加
- 对象组合
- 对象恢复
- Select 对象查询
- Post Policy
- QoS 配置

## Problem Statement

### Current Limitations

1. **功能不完整**：许多 S3 兼容功能未实现，限制了 SDK 的使用场景
2. **示例无法迁移**：71 个旧示例中约 36 个因功能缺失无法迁移
3. **企业级特性缺失**：加密、合规性（对象锁定）、跨区复制等企业必需功能未实现
4. **API 不一致**：部分功能散落在 old 目录，未按新架构组织

### Why This Matters

- **用户需求**：企业用户需要完整的 S3 API 支持
- **兼容性**：与 AWS S3 和 MinIO 完全兼容
- **示例完整性**：提供完整的参考示例
- **生产就绪**：满足生产环境的安全和合规要求

## Proposed Solution

### High-Level Design

按照现有模块化架构（service-based + option functions）实现以下功能：

#### 1. Bucket Service Extensions (bucket/)

```go
// CORS configuration
SetCORS(ctx context.Context, bucketName string, config CORSConfig) error
GetCORS(ctx context.Context, bucketName string) (CORSConfig, error)
DeleteCORS(ctx context.Context, bucketName string) error

// Bucket encryption
SetEncryption(ctx context.Context, bucketName string, config EncryptionConfig) error
GetEncryption(ctx context.Context, bucketName string) (EncryptionConfig, error)
DeleteEncryption(ctx context.Context, bucketName string) error

// Bucket tagging (if not exists)
SetTagging(ctx context.Context, bucketName string, tags map[string]string) error
GetTagging(ctx context.Context, bucketName string) (map[string]string, error)
DeleteTagging(ctx context.Context, bucketName string) error

// Object locking
SetObjectLockConfig(ctx context.Context, bucketName string, config ObjectLockConfig) error
GetObjectLockConfig(ctx context.Context, bucketName string) (ObjectLockConfig, error)
```

#### 2. Object Service Extensions (object/)

```go
// Server-side encryption options
WithSSES3() PutOption
WithSSEKMS(keyID string, context map[string]string) PutOption
WithSSEC(key []byte) PutOption

// Object legal hold
SetLegalHold(ctx context.Context, bucketName, objectName string, hold LegalHoldStatus, opts ...Option) error
GetLegalHold(ctx context.Context, bucketName, objectName string, opts ...Option) (LegalHoldStatus, error)

// Object retention
SetRetention(ctx context.Context, bucketName, objectName string, retention RetentionMode, until time.Time, opts ...Option) error
GetRetention(ctx context.Context, bucketName, objectName string, opts ...Option) (RetentionMode, time.Time, error)

// Object ACL
GetACL(ctx context.Context, bucketName, objectName string, opts ...Option) (ACL, error)
SetACL(ctx context.Context, bucketName, objectName string, acl ACL, opts ...Option) error

// Compose objects
Compose(ctx context.Context, dst DestinationInfo, sources []SourceInfo, opts ...PutOption) (UploadInfo, error)

// Append object (RustFS extension)
Append(ctx context.Context, bucketName, objectName string, reader io.Reader, size int64, offset int64, opts ...PutOption) (UploadInfo, error)
```

#### 3. New Modules

```go
// pkg/cors - CORS configuration types
package cors

type Config struct {
    CORSRules []CORSRule
}

type CORSRule struct {
    AllowedOrigins []string
    AllowedMethods []string
    AllowedHeaders []string
    ExposeHeaders  []string
    MaxAgeSeconds  int
}

// pkg/sse - Server-side encryption types
package sse

type Configuration struct {
    Rules []Rule
}

type Rule struct {
    ApplySSEByDefault DefaultEncryption
}

type DefaultEncryption struct {
    SSEAlgorithm   string
    KMSMasterKeyID string
}

// pkg/objectlock - Object lock types
package objectlock

type Config struct {
    ObjectLockEnabled string
    Rule              *Rule
}

type Rule struct {
    DefaultRetention DefaultRetention
}

type DefaultRetention struct {
    Mode  RetentionMode
    Days  int
    Years int
}

// pkg/notification - Event notification (if not exists)
package notification

type Configuration struct {
    QueueConfigurations   []QueueConfig
    TopicConfigurations   []TopicConfig
    LambdaConfigurations  []LambdaConfig
}
```

### API Migration Plan

| Old API | New API | Package | Priority |
|---------|---------|---------|----------|
| SetBucketCors | bucket.SetCORS | bucket | High |
| SetBucketEncryption | bucket.SetEncryption | bucket | High |
| SetBucketReplication | bucket.SetReplication | bucket | Medium |
| SetBucketNotification | bucket.SetNotification | bucket | Medium |
| GetObjectAcl | object.GetACL | object | Low |
| PutObjectLegalHold | object.SetLegalHold | object | Medium |
| PutObjectRetention | object.SetRetention | object | Medium |
| ComposeObject | object.Compose | object | Medium |
| AppendObject | object.Append | object | Low |
| SelectObjectContent | object.Select | object | Low |
| RestoreObject | object.Restore | object | Low |

### Implementation Phases

**Phase 1: Security & Compliance (2 weeks)**
- Object encryption (SSE-S3, SSE-C, SSE-KMS)
- Bucket encryption configuration
- Object locking, legal hold, retention

**Phase 2: Configuration & Management (1 week)**
- CORS configuration
- Bucket tagging
- ACL configuration

**Phase 3: Advanced Operations (2 weeks)**
- Bucket replication
- Event notification
- Object composition
- Object append

**Phase 4: Query & Restore (1 week)**
- Select object content
- Restore archived objects
- Post Policy

## Alternatives Considered

### Alternative 1: Keep Using Old SDK
- ❌ Maintains technical debt
- ❌ Inconsistent API patterns
- ❌ Hard to maintain

### Alternative 2: Implement Only Core Features
- ❌ Incomplete S3 compatibility
- ❌ Limits use cases
- ✅ Faster delivery

### Alternative 3: Gradual Migration (Chosen)
- ✅ Incremental value delivery
- ✅ Manageable scope
- ✅ Can prioritize by demand

## Impact Assessment

### Benefits
1. **Complete S3 Compatibility**: 100% feature parity with MinIO SDK
2. **Enterprise Ready**: Support for encryption, compliance, replication
3. **Better Architecture**: Consistent with new modular design
4. **More Examples**: Can migrate all 71 examples from old SDK
5. **Future Proof**: Easy to extend with new features

### Risks
1. **Development Time**: ~6 weeks total for full implementation
2. **Testing Complexity**: Need infrastructure for encryption, replication testing
3. **Breaking Changes**: Minimal, mostly additive

### Migration Path
- All new APIs are additive
- No breaking changes to existing code
- Old SDK examples can be migrated incrementally

## Success Criteria

- [ ] All priority 1 (High) features implemented
- [ ] Unit tests with >80% coverage
- [ ] Integration tests with real RustFS server
- [ ] API documentation complete
- [ ] At least 20 new examples added
- [ ] Old SDK examples migrated (target: 50+ examples)

## Open Questions

1. Do we need to support SSE-KMS? (AWS-specific feature)
2. Should we implement Post Policy for browser uploads?
3. Priority order for Phase 3 features?
4. Testing infrastructure requirements?

## References

- [AWS S3 API Documentation](https://docs.aws.amazon.com/AmazonS3/latest/API/)
- [MinIO SDK Go Documentation](https://min.io/docs/minio/linux/developers/go/API.html)
- Current implementation: `bucket/`, `object/` packages
- Old implementation: `old/api-*.go` files
