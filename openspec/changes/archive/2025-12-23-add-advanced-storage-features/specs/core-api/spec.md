## ADDED Requirements
### Requirement: 对象预签名 URL 支持
RustFS Go SDK SHALL 提供对象 GET/PUT 的预签名 URL 生成能力，允许配置过期时间及可选响应/请求头。

#### Scenario: 生成 GET 预签名 URL
- **WHEN** 客户端调用预签名接口为对象生成 GET URL，并指定过期时间（例如 15 分钟）与可选响应头（`response-content-type` 等）
- **THEN** 返回的 URL 包含签名查询参数和过期时间
- **AND** 在过期前使用该 URL 发起 GET 请求能够成功下载对象
- **AND** 服务器响应中应用了指定的响应头

#### Scenario: 生成 PUT 预签名 URL 并上传成功
- **WHEN** 客户端为对象生成 PUT 预签名 URL，并在签名时指定 Content-Type 或自定义头约束
- **AND** 客户端使用匹配的头通过该 URL 执行 HTTP PUT 上传
- **THEN** 上传请求在过期前成功，服务器接受对象并返回 2xx 状态
- **AND** 重新获取对象时其元数据（如 Content-Type）与签名约束一致

### Requirement: 对象标签管理
RustFS Go SDK SHALL 支持对象标签的设置、获取与删除操作，映射到 S3 兼容的标签 API。

#### Scenario: 设置并读取对象标签
- **WHEN** 客户端调用标签设置接口为对象写入标签键值对
- **AND** 随后调用标签获取接口
- **THEN** 返回的标签列表与设置时一致，包含所有键值对

#### Scenario: 删除对象标签
- **WHEN** 客户端调用删除标签接口
- **THEN** 服务器删除对象标签
- **AND** 再次获取标签返回空结果或 NotFound 错误

### Requirement: 桶策略管理
RustFS Go SDK SHALL 提供桶策略的设置、获取与删除接口，接受和返回标准 S3 Policy JSON。

#### Scenario: 设置并获取桶策略
- **WHEN** 客户端为指定桶设置策略 JSON
- **AND** 随后读取桶策略
- **THEN** 返回的策略内容与设置值一致
- **AND** 删除桶策略后再次读取返回空策略或 NotFound 错误

### Requirement: 桶生命周期管理
RustFS Go SDK SHALL 支持桶生命周期配置的设置、获取与删除，遵循 S3 生命周期 XML 结构。

#### Scenario: 设置并获取桶生命周期配置
- **WHEN** 客户端提交包含规则 ID、前缀/过滤器、过期或转换动作的生命周期配置
- **AND** 随后获取生命周期配置
- **THEN** 返回的配置与提交内容一致，包含所有规则
- **AND** 删除配置后再次获取返回空配置或 NotFound 错误

### Requirement: 服务端加密选项支持
RustFS Go SDK SHALL 在对象上传/下载/预签名相关接口支持 SSE-S3 与 SSE-C 选项，并正确传递所需头信息。

#### Scenario: 使用 SSE-S3 上传对象
- **WHEN** 客户端在上传或 fput 调用中启用 SSE-S3 选项
- **THEN** SDK 在请求中包含服务端加密头
- **AND** 对象被以 SSE-S3 模式存储，后续 GET 请求返回的头部表明已加密

#### Scenario: 使用 SSE-C 上传并下载对象
- **WHEN** 客户端在上传时提供 SSE-C 密钥、算法与 MD5
- **THEN** 请求包含匹配的 SSE-C 头，上传成功
- **AND** 客户端使用相同密钥下载对象成功，使用错误密钥会返回解密相关错误

### Requirement: 基于文件的 fput/fget 便捷操作
RustFS Go SDK SHALL 提供基于文件路径的 fput/fget 便捷方法，自动处理文件打开/关闭、内容类型推断和校验。

#### Scenario: 使用 fput 直接上传文件
- **WHEN** 客户端调用 fput 传入文件路径、目标桶名与对象键
- **THEN** SDK 打开文件并流式上传
- **AND** 自动推断常见文件类型作为 Content-Type，可通过选项覆盖
- **AND** 上传完成后返回对象信息（ETag、大小），并在出错时关闭文件句柄

#### Scenario: 使用 fget 直接下载到文件
- **WHEN** 客户端调用 fget 下载对象到指定文件路径
- **THEN** SDK 将对象内容写入文件（覆盖或使用安全写入策略）
- **AND** 下载完成后文件大小与返回的 Content-Length 一致
- **AND** 支持可选进度/校验或加密选项与其他下载接口保持一致
