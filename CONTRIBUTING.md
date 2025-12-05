# 贡献指南

感谢您对 RustFS Go SDK 项目的关注！我们欢迎所有形式的贡献。

## 如何贡献

### 报告问题

如果您发现了 bug 或有功能建议，请通过 GitHub Issues 提交。

### 提交代码

1. Fork 本仓库
2. 创建您的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交您的更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启一个 Pull Request

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 添加必要的注释和文档
- 为新功能添加单元测试

### 测试

在提交 PR 之前，请确保：

```bash
# 运行所有测试
go test ./...

# 检查代码格式
gofmt -s -w .

# 运行 linter
golangci-lint run
```

### 提交信息规范

提交信息应该清晰描述更改内容：

- `feat: 添加新功能`
- `fix: 修复 bug`
- `docs: 更新文档`
- `test: 添加测试`
- `refactor: 代码重构`

## 开发环境设置

1. 克隆仓库
```bash
git clone https://github.com/Scorpio69t/rustfs-go.git
cd rustfs-go
```

2. 安装依赖
```bash
go mod download
```

3. 运行测试
```bash
go test ./...
```

## 许可证

通过贡献代码，您同意您的贡献将在 Apache License 2.0 下授权。
