# examples Specification Deltas

## ADDED Requirements

### Requirement: S3 示例代码集
RustFS Go SDK SHALL 提供一套完整的 S3 API 使用示例，涵盖存储桶操作、对象操作、高级功能等，帮助用户学习和使用 SDK。

#### Scenario: 用户查找基础操作示例
- **WHEN** 用户需要了解如何创建存储桶
- **AND** 访问 `examples/s3/` 目录
- **THEN** 能找到 `bucket-create.go` 示例
- **AND** 示例包含清晰的中文注释和完整代码
- **AND** 示例可以独立编译和运行

#### Scenario: 用户查找对象上传示例
- **WHEN** 用户需要了解如何上传对象
- **AND** 访问 `examples/s3/` 目录
- **THEN** 能找到 `object-put.go` 和 `file-upload.go` 示例
- **AND** 示例演示了不同的上传方式（内存、文件）
- **AND** 示例展示了如何使用选项函数设置元数据、标签等

#### Scenario: 用户查找高级功能示例
- **WHEN** 用户需要了解如何使用版本控制
- **AND** 访问 `examples/s3/` 目录
- **THEN** 能找到 `versioning-enable.go`、`versioning-list.go` 等示例
- **AND** 示例演示了版本控制的完整流程

#### Scenario: 示例代码质量保证
- **WHEN** 用户运行任何示例
- **THEN** 示例代码能够成功编译
- **AND** 代码遵循 Go 代码规范（gofmt）
- **AND** 代码没有未使用的导入
- **AND** 错误处理清晰明确

### Requirement: 示例使用新 API
所有 S3 示例 SHALL 使用 RustFS SDK 的新模块化 API，包括服务化接口和选项模式。

#### Scenario: 使用 Bucket 服务
- **WHEN** 示例涉及存储桶操作
- **THEN** 使用 `client.Bucket()` 获取服务
- **AND** 使用 `bucket.WithXxx()` 选项函数
- **AND** 不使用旧的直接方法调用

#### Scenario: 使用 Object 服务
- **WHEN** 示例涉及对象操作
- **THEN** 使用 `client.Object()` 获取服务
- **AND** 使用 `object.WithXxx()` 选项函数
- **AND** 遵循新的参数顺序和命名约定

#### Scenario: 统一的客户端初始化
- **WHEN** 任何示例初始化客户端
- **THEN** 使用 `rustfs.New()` 而非 `minio.New()`
- **AND** 使用 `rustfs.Options` 结构体
- **AND** 导入路径为 `github.com/Scorpio69t/rustfs-go`

### Requirement: 示例代码独立性
每个示例文件 SHALL 能够独立编译和运行，不依赖其他示例文件。

#### Scenario: 单独编译示例
- **WHEN** 用户对单个示例文件执行 `go build`
- **THEN** 编译成功生成可执行文件
- **AND** 不报告缺少依赖或导入错误

#### Scenario: 独立运行示例
- **WHEN** 用户运行编译后的示例程序
- **AND** 提供了必要的配置（endpoint、credentials）
- **THEN** 程序能够成功执行
- **AND** 输出清晰的执行结果或错误信息

### Requirement: 示例分类和组织
示例 SHALL 按功能分类组织，使用清晰的命名规则，便于用户查找。

#### Scenario: 按功能查找示例
- **WHEN** 用户需要特定功能的示例
- **AND** 根据文件名搜索（如 "tagging"）
- **THEN** 能找到所有相关示例
- **AND** 文件名清晰表达功能（如 `tagging-object-set.go`）

#### Scenario: 示例索引文档
- **WHEN** 用户打开 `examples/s3/README.md`
- **THEN** 能看到所有示例的分类列表
- **AND** 每个示例有简短描述
- **AND** 包含使用说明和前置条件

### Requirement: 避免版权纠纷
示例代码 SHALL 为独立实现，不包含 MinIO 或其他第三方的版权声明。

#### Scenario: 代码独立性验证
- **WHEN** 检查任何示例文件
- **THEN** 不包含 MinIO 版权声明
- **AND** 不包含其他第三方版权声明
- **AND** 代码逻辑为独立重写，非直接复制

#### Scenario: 代码风格一致性
- **WHEN** 审查所有示例代码
- **THEN** 使用统一的代码模板和风格
- **AND** 注释、命名、结构与 RustFS 项目一致
- **AND** 代码质量符合 Go 最佳实践

### Requirement: 示例覆盖完整性
示例集 SHALL 覆盖 RustFS SDK 支持的所有主要 S3 API 功能。

#### Scenario: 核心功能覆盖
- **WHEN** 用户查看示例集
- **THEN** 包含所有核心存储桶操作示例
- **AND** 包含所有核心对象操作示例
- **AND** 包含文件上传下载示例

#### Scenario: 高级功能覆盖
- **WHEN** 用户需要高级功能示例
- **THEN** 包含版本控制、标签、加密示例
- **AND** 包含预签名 URL 示例
- **AND** 包含对象锁定、保留、生命周期等示例

#### Scenario: 功能映射完整性
- **WHEN** 对比 `old/examples/s3/` 中的 71 个示例
- **AND** 检查新示例集的覆盖范围
- **THEN** 所有旧示例的功能都有对应的新示例
- **AND** 新 SDK 不支持的功能有明确标注
