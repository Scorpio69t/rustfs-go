# pkg 包模块重构规范

## MODIFIED Requirements

### Requirement: pkg/signer 公共 API 包
系统 SHALL 提供 `pkg/signer` 包作为 AWS 签名功能的公共 API，供其他包使用。

#### Scenario: 使用 SignV4 进行标准签名
- **WHEN** 调用 `pkg/signer.SignV4` 函数并提供有效的请求参数
- **THEN** 返回正确格式的 AWS Signature V4 签名字符串
- **AND** 签名可用于 AWS S3 兼容的 API 请求

#### Scenario: 使用 SignV4STS 进行 STS 签名
- **WHEN** 调用 `pkg/signer.SignV4STS` 函数并提供 STS 凭证
- **THEN** 返回包含会话令牌的 AWS Signature V4 签名
- **AND** 签名可用于临时凭证的 API 请求

#### Scenario: pkg/signer 包不依赖其他 pkg 包
- **WHEN** 检查 `pkg/signer` 的依赖关系
- **THEN** 该包仅依赖标准库和内部包
- **AND** 不导入任何其他 `pkg/` 目录下的包

### Requirement: pkg/credentials 包版权声明
`pkg/credentials` 包的所有文件 SHALL 使用 RustFS 版权声明，移除 MinIO 版权信息。

#### Scenario: 版权声明更新
- **WHEN** 检查 `pkg/credentials` 包中的任意文件
- **THEN** 文件头部包含 RustFS Go SDK 版权声明
- **AND** 版权声明格式为 Apache License 2.0
- **AND** 包含对 MinIO 的致谢说明（如适用）

#### Scenario: 使用 pkg/signer 解决循环依赖
- **WHEN** `pkg/credentials/assume_role.go` 需要签名功能
- **THEN** 使用 `pkg/signer.SignV4STS` 而不是内部实现
- **AND** 不产生与 `internal/signer` 的循环依赖

#### Scenario: 公共 API 保持兼容
- **WHEN** 使用 `pkg/credentials` 包的公共 API
- **THEN** 所有公共接口和函数签名保持不变
- **AND** 现有代码无需修改即可继续使用
