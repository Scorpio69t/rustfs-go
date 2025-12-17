# pkg 包使用情况分析

## 分析日期
2025-12-17

## 包使用情况统计

### 1. pkg/credentials ✅ **被使用**
**使用位置**:
- `client.go` - 客户端初始化
- `options.go` - 配置选项
- `internal/core/executor.go` - 执行器
- `internal/signer/signer.go` - 签名器
- `examples/rustfs/*.go` - 所有示例

**状态**: 🔴 **必须保留并重构**
**优先级**: P0 - 最高
**重构工作量**: 大（20+ 文件）

### 2. pkg/signer ⚠️ **部分使用**
**使用位置**:
- `pkg/credentials/assume_role.go` - 仅导入声明

**状态**: 🟡 **可以删除**
**理由**:
- 已有 `internal/signer` 完整实现
- 仅在 `pkg/credentials` 中有导入引用
- 实际签名逻辑使用 `internal/signer`

**推荐**: 删除 `pkg/signer`，清理 `pkg/credentials` 中的引用

### 3. pkg/s3utils ⚠️ **部分使用**
**使用位置**:
- `pkg/signer/request-signature-v2.go`
- `pkg/signer/request-signature-v4.go`

**状态**: 🟡 **评估后决定**
**理由**:
- 仅被 `pkg/signer` 使用
- 如果删除 `pkg/signer`，则 `pkg/s3utils` 也可删除
- 相关功能已在 `internal/signer/utils.go` 实现

**推荐**: 删除 `pkg/s3utils`

### 4. pkg/set ❌ **未使用**
**使用位置**: 无

**状态**: 🔴 **可以删除**
**理由**: 项目中没有任何地方使用此包

**推荐**: 删除 `pkg/set`

### 5. pkg/lifecycle ❌ **未使用**
**使用位置**: 无

**状态**: 🔴 **可以删除**
**推荐**: 删除

### 6. pkg/notification ❌ **未使用**
**使用位置**: 无

**状态**: 🔴 **可以删除**
**推荐**: 删除

### 7. pkg/policy ❌ **未使用**
**使用位置**: 无

**状态**: 🔴 **可以删除**
**推荐**: 删除

### 8. pkg/replication ❌ **未使用**
**使用位置**: 无

**状态**: 🔴 **可以删除**
**推荐**: 删除

### 9. pkg/encrypt ❌ **未使用**
**使用位置**: 无

**状态**: 🔴 **可以删除**
**推荐**: 删除

### 10. pkg/kvcache ❌ **未使用**
**使用位置**: 无

**状态**: 🔴 **可以删除**
**理由**: 已有 `internal/cache` 实现

**推荐**: 删除

### 11. pkg/singleflight ❌ **未使用**
**使用位置**: 无

**状态**: 🔴 **可以删除**
**推荐**: 删除

### 12. pkg/sse ❌ **未使用**
**使用位置**: 无

**状态**: 🔴 **可以删除**
**推荐**: 删除

### 13. pkg/tags ❌ **未使用**
**使用位置**: 无

**状态**: 🔴 **可以删除**
**推荐**: 删除

## 总结

### 保留的包（需要重构）
1. ✅ **pkg/credentials** - 核心凭证管理（必须）

### 删除的包
1. ❌ **pkg/signer** - 已被 `internal/signer` 替代
2. ❌ **pkg/s3utils** - 已被 `internal/signer/utils.go` 替代
3. ❌ **pkg/set** - 未使用
4. ❌ **pkg/lifecycle** - 未使用
5. ❌ **pkg/notification** - 未使用
6. ❌ **pkg/policy** - 未使用
7. ❌ **pkg/replication** - 未使用
8. ❌ **pkg/encrypt** - 未使用
9. ❌ **pkg/kvcache** - 未使用（已有 `internal/cache`）
10. ❌ **pkg/singleflight** - 未使用
11. ❌ **pkg/sse** - 未使用
12. ❌ **pkg/tags** - 未使用

## 重构策略

### 第一步：清理未使用的包（快速）
删除以下未使用的包：
```bash
rm -rf pkg/signer
rm -rf pkg/s3utils
rm -rf pkg/set
rm -rf pkg/lifecycle
rm -rf pkg/notification
rm -rf pkg/policy
rm -rf pkg/replication
rm -rf pkg/encrypt
rm -rf pkg/kvcache
rm -rf pkg/singleflight
rm -rf pkg/sse
rm -rf pkg/tags
```

### 第二步：重构 pkg/credentials（重点）
1. 更新所有文件的版权声明
2. 重命名 MinIO 特定的类型和函数
3. 合并重复的功能（env、file）
4. 简化 STS 功能
5. 更新所有测试
6. 更新文档

### 第三步：验证和测试
1. 运行所有测试
2. 运行所有示例
3. 更新文档

## 工作量估算

| 任务 | 工作量 | 优先级 |
|------|--------|--------|
| 删除未使用的包 | 0.5 天 | P0 |
| 重构 credentials 核心文件 | 1 天 | P0 |
| 重构 credentials 环境变量 | 0.5 天 | P0 |
| 重构 credentials 文件凭证 | 0.5 天 | P0 |
| 重构 credentials STS 功能 | 1 天 | P1 |
| 更新测试用例 | 1 天 | P0 |
| 更新文档 | 0.5 天 | P1 |
| **总计** | **5-6 天** | - |

## 风险评估

### 高风险
- ❌ 无

### 中风险
- ⚠️ STS 功能重构可能影响某些用户
  - **缓解**: 保留核心 STS 功能，只移除 MinIO 特定扩展

### 低风险
- ✅ 删除未使用的包 - 无影响
- ✅ 更新版权声明 - 无功能影响
- ✅ 重命名内部类型 - 不影响公共 API

## 建议

1. **快速清理**: 先删除所有未使用的包（立即见效）
2. **增量重构**: 按文件逐个重构 `pkg/credentials`
3. **测试驱动**: 每个重构步骤都先运行测试
4. **文档同步**: 代码和文档同步更新

## 下一步行动

1. ✅ 创建 OpenSpec 提案（本文档）
2. ⏭️ 执行阶段 1: 删除未使用的包
3. ⏭️ 执行阶段 2: 重构 pkg/credentials
4. ⏭️ 执行阶段 3: 测试和验证
