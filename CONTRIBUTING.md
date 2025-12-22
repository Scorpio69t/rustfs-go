# Contributing Guide

Thanks for your interest in the RustFS Go SDK! We welcome contributions of all kinds.

## How to Contribute

### Report Issues

If you find a bug or have a feature request, please file an issue on GitHub.

### Submit Code

1. Fork this repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m "feat: add amazing feature"`)
4. Push to your branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

### Code Style

- Follow the official Go coding guidelines
- Run `gofmt` to format code
- Add clear comments and GoDoc for public APIs
- Add unit tests for new functionality

### Testing

Before sending a PR, please run:

```bash
# Run all tests
go test ./...

# Check code format
gofmt -s -w .

# Run linter
golangci-lint run
```

### Commit Messages

Commit messages should clearly describe the change. Examples:

- `feat: add new feature`
- `fix: resolve panic when bucket name is empty`
- `docs: update README`
- `test: add health check tests`
- `refactor: simplify signer options`

## Development Environment

1. Clone the repository
```bash
git clone https://github.com/Scorpio69t/rustfs-go.git
cd rustfs-go
```

2. Install dependencies
```bash
go mod download
```

3. Run tests
```bash
go test ./...
```

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
