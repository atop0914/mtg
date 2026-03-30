# Contributing to MTG

Thanks for your interest in contributing to MTG!

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git

### Getting Started

1. **Fork and clone the repository**

```bash
git clone https://github.com/YOUR_USERNAME/mtg.git
cdmtg
```

2. **Create a feature branch**

```bash
git checkout -b feature/your-feature-name
```

3. **Make your changes**

- Write clean, readable code
- Add tests for new functionality
- Update documentation as needed

4. **Run tests**

```bash
go test ./...
```

5. **Build and verify**

```bash
go build -o mtg ./cmd/mtg
```

6. **Commit and push**

```bash
git add .
git commit -m "feat: add your feature description"
git push origin feature/your-feature-name
```

7. **Open a Pull Request**

## Code Style

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Run `go fmt` before committing
- Keep functions small and focused
- Add comments for complex logic

## Reporting Issues

When reporting bugs, please include:

- Go version (`go version`)
- Operating system
- Configuration file (redact secrets)
- Error message or log output
- Steps to reproduce

## Feature Requests

Feel free to suggest new features by opening an issue with:

- Clear description of the feature
- Use case / motivation
- Potential implementation ideas (if any)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
