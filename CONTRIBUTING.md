# Contributing to bc-lib

First off, thank you for considering contributing to bc-lib! 🎉

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues. When you create a bug report, include:

- A clear and descriptive title
- Steps to reproduce the issue
- Expected vs. actual behavior
- Code samples if applicable
- Your Go version and OS

### Suggesting Enhancements

Enhancement suggestions are welcome! Please provide:

- A clear description of the enhancement
- Use cases and examples
- Why it would be useful to most users

### Pull Requests

1. Fork the repo and create your branch from `main`
2. If you've added code, add tests
3. Ensure the test suite passes (`go test ./...`)
4. Make sure your code follows Go conventions (`gofmt`, `golint`)
5. Write clear commit messages
6. Open a pull request!

## Development Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/bc-lib
cd bc-lib

# Install dependencies
go mod download

# Run tests
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Format code
gofmt -w .
```

## Coding Guidelines

### Go Style

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- Run `gofmt` before committing
- Keep functions focused and small
- Write descriptive variable names

### Testing

- Write tests for new features
- Aim for >80% code coverage
- Use table-driven tests where appropriate
- Mock external dependencies

Example test:

```go
func TestNewAddress(t *testing.T) {
    tests := []struct {
        name    string
        input   []byte
        wantErr bool
    }{
        {
            name:    "valid 20 bytes",
            input:   make([]byte, 20),
            wantErr: false,
        },
        {
            name:    "invalid length",
            input:   make([]byte, 10),
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := NewAddress(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("NewAddress() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Documentation

- Add GoDoc comments for all exported functions, types, and constants
- Include examples in documentation where helpful
- Update README.md if you change user-facing features

Example GoDoc:

```go
// AddressFromHex creates an Address from a hex string.
// Accepts both "0x" prefixed and non-prefixed strings.
//
// Example:
//   addr, err := AddressFromHex("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
func AddressFromHex(s string) (Address, error) {
    // implementation
}
```

## Commit Messages

Use clear, descriptive commit messages:

```
feat: add Solana chain support
fix: handle nil pointer in GetBalance
docs: update getting started guide
test: add tests for transaction builder
refactor: simplify RPC provider logic
```

Prefixes:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `test`: Adding or updating tests
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `perf`: Performance improvement
- `chore`: Maintenance tasks

## Adding a New Chain

To add support for a new blockchain:

1. Implement the `Chain` interface in `pkg/chains/`
2. Add chain ID constants to `pkg/chains/interface.go`
3. Update `client.createChain()` to handle the new chain
4. Add tests
5. Update documentation
6. Add example usage

## Review Process

1. Maintainers will review your PR
2. Address any feedback or requested changes
3. Once approved, your PR will be merged
4. Your contribution will be acknowledged in the changelog!

## Questions?

Feel free to open an issue with the `question` label, or reach out to the maintainers.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

Thank you for contributing to bc-lib! 🚀

