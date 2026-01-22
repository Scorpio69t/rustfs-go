# core-api Specification

## Purpose
TBD - created by archiving change data-protection-and-events. Update Purpose after archive.
## Requirements
### Requirement: 桶版本控制管理
RustFS Go SDK SHALL 提供桶版本控制的开启/暂停配置接口，并能读取当前状态。

#### Scenario: 启用桶版本控制
- **WHEN** 客户端调用开启版本控制接口（如设置为 Enabled）
- **AND** 随后读取版本控制状态
- **THEN** 返回状态为 Enabled
- **AND** 对象新写入将生成唯一 versionId

#### Scenario: 暂停桶版本控制
- **WHEN** 客户端将版本控制状态设置为 Suspended
- **AND** 再次读取状态
- **THEN** 返回状态为 Suspended
- **AND** 新上传对象不再生成新的 versionId

### Requirement: 版本化对象操作与枚举
RustFS Go SDK SHALL 支持在 Get/Head/Delete/Copy 中传入 `versionId`，并提供对象版本与删除标记的列举接口。

#### Scenario: 读取指定 versionId 的对象
- **WHEN** 客户端在 Get 或 Head 请求中指定有效 versionId
- **THEN** 返回的对象内容与元数据对应该版本
- **AND** 返回头部包含匹配的 versionId

#### Scenario: 列举对象版本与删除标记
- **WHEN** 客户端调用 ListObjectVersions
- **THEN** 返回结果包含版本条目与 DeleteMarker 条目
- **AND** 支持前缀、分页等过滤参数

#### Scenario: 删除指定版本的对象
- **WHEN** 客户端在 Delete 请求中指定 versionId
- **THEN** 成功删除对应版本或创建删除标记
- **AND** API 返回 versionId 或删除标记相关信息

### Requirement: 跨区复制配置
RustFS Go SDK SHALL 提供桶跨区复制配置的设置、获取与删除接口，支持多规则、过滤器和目标定义，并进行必要校验。

#### Scenario: 设置并获取复制配置
- **WHEN** 客户端提交复制配置，包含至少一条规则（ID、状态、前缀/标签过滤器、目标桶/区域、存储类型等）
- **AND** 随后获取复制配置
- **THEN** 返回的配置与提交内容一致，包含所有规则与目标

#### Scenario: 校验复制配置约束
- **WHEN** 提交的复制配置缺少目标桶或规则状态无效
- **THEN** SDK 返回校验错误，阻止无效配置下发

### Requirement: 桶事件通知配置
RustFS Go SDK SHALL 提供桶事件通知配置的设置、获取与删除接口，支持常见对象事件类型与多种目标（队列/主题/回调）。

#### Scenario: 设置并获取事件通知
- **WHEN** 客户端为桶提交事件通知配置，包含事件类型（如 ObjectCreated、ObjectRemoved 等）及目标（Queue/Topic/Callback）
- **AND** 随后获取通知配置
- **THEN** 返回的配置与提交内容一致，包含目标与过滤条件（前缀/标签）

#### Scenario: 通知配置校验
- **WHEN** 配置包含不支持的事件类型或缺少目标标识
- **THEN** SDK 返回校验错误并拒绝提交

### Requirement: 桶访问日志配置
RustFS Go SDK SHALL 提供桶访问日志配置的设置、获取与删除接口，支持目标桶、前缀及权限校验。

#### Scenario: 设置并获取访问日志配置
- **WHEN** 客户端为桶设置访问日志，指定目标桶与前缀
- **AND** 随后读取访问日志配置
- **THEN** 返回的配置与提交值一致

#### Scenario: 访问日志配置校验
- **WHEN** 配置缺少目标桶或目标桶权限不足
- **THEN** SDK 返回校验错误或透出服务端的权限错误信息

