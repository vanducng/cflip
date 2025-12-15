# Golang Best Practices & Guidelines

This document outlines the best practices, coding standards, and guidelines for developing Go projects in our organization. These guidelines ensure consistency, maintainability, and high-quality code across all Go projects.

## Table of Contents
1. [Project Structure](#project-structure)
2. [Naming Conventions](#naming-conventions)
3. [Code Organization](#code-organization)
4. [Error Handling](#error-handling)
5. [Concurrency](#concurrency)
6. [Testing](#testing)
7. [Documentation](#documentation)
8. [Dependencies](#dependencies)
9. [Performance](#performance)
10. [Security](#security)
11. [Pre-commit & CI Guidelines](#pre-commit--ci-guidelines)

## Project Structure

### Standard Layout
```
project/
├── cmd/                    # Main applications
│   └── appname/
│       └── main.go
├── internal/               # Private application code
│   ├── auth/
│   ├── config/
│   └── service/
├── pkg/                    # Public library code
│   └── publicapi/
├── api/                    # API definitions
│   └── proto/
├── web/                    # Web assets
├── configs/                # Configuration files
├── init/                   # System init scripts
├── scripts/                # Build and utility scripts
├── build/                  # Build artifacts
├── deployments/            # Deployment configurations
├── test/                   # Test utilities
├── docs/                   # Documentation
├── tools/                  # Build tools
├── examples/               # Example applications
├── third_party/            # External dependencies
├── githooks/               # Git hooks
├── website/                # Website files
├── Makefile                # Build configuration
├── go.mod                  # Module definition
├── go.sum                  # Dependency checksums
└── README.md               # Project documentation
```

### Module Naming
- Use short, descriptive names: `github.com/company/project`
- Avoid `src` in the import path
- Keep the module name simple and memorable

## Naming Conventions

### Packages
- Use short, lowercase, single-word names when possible
- Don't use plural forms (use `user` not `users`)
- Avoid package name collisions with standard library
- Be consistent: `models` not `model`, `utils` not `util`

```go
// Good
package http
package user
package auth

// Bad
package httputil
package userService
package authentication
```

### Files
- Use lowercase, underscore-separated names
- Match file name to primary type/purpose
- Keep files focused (single responsibility)

```go
// Good
user_service.go
http_client.go
config_parser.go

// Bad
userService.go
HTTPClient.go
ConfigParser.go
```

### Variables & Constants
- Use camelCase for variables
- Use UPPER_SNAKE_CASE for exported constants
- Use camelCase for unexported constants
- Use meaningful names, avoid abbreviations

```go
// Good
const MaxRetries = 3
const defaultTimeout = 30 * time.Second

var httpClient HTTPClient
var userService UserService

// Bad
const MAX_RETRIES = 3
var defaultTimeout = 30 * time.Second
var hClient HTTPClient
```

### Functions & Methods
- Use camelCase
- Use verbs for functions that perform actions
- Use noun phrases for functions that return values

```go
// Good
func GetUser(id int) (*User, error)
func (u *User) Validate() error
func (u *User) String() string
func HashPassword(password string) string

// Bad
func user(id int) (*User, error)
func (u *User) validateUser() error
func (u *User) toString() string
```

### Interfaces
- Use `-er` suffix for interfaces with single methods
- Keep interfaces small and focused
- Name interfaces based on behavior, not implementation

```go
// Good
type Reader interface {
    Read(p []byte) (n int, err error)
}

type UserFinder interface {
    FindUser(id int) (*User, error)
}

// Bad
type UserInterface interface {
    GetUser(id int) (*User, error)
    UpdateUser(user *User) error
    DeleteUser(id int) error
}
```

## Code Organization

### Group Imports
Organize imports in three groups with blank lines:
1. Standard library
2. Third-party packages
3. Internal/local packages

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/gorilla/mux"
    "go.uber.org/zap"

    "github.com/company/project/internal/auth"
    "github.com/company/project/internal/config"
)
```

### Exported vs Unexported
- Export identifiers when they're part of the API
- Keep implementation details unexported
- Use getters/setters for exported fields when validation is needed

```go
type User struct {
    ID       int       // Exported
    Name     string    // Exported
    email    string    // Unexported
    password string    // Unexported
}

func (u *User) Email() string {
    return u.email
}

func (u *User) SetEmail(email string) error {
    if !isValidEmail(email) {
        return fmt.Errorf("invalid email format")
    }
    u.email = email
    return nil
}
```

### Constructor Pattern
Use `New` functions for constructors, return interfaces when possible:

```go
// Good
func NewUserService(repo UserRepository) UserService {
    return &userService{
        repo: repo,
        logger: zap.NewNop(),
    }
}

func NewFileCache(path string) (Cache, error) {
    if path == "" {
        return nil, fmt.Errorf("path cannot be empty")
    }
    return &fileCache{path: path}, nil
}

// Bad
func UserService(repo UserRepository) *userService {
    return &userService{repo: repo}
}
```

## Error Handling

### Always Handle Errors
Never ignore errors, even if they seem impossible:

```go
// Good
_, err := io.WriteString(w, "Hello")
if err != nil {
    log.Printf("Failed to write response: %v", err)
    return err
}

// Bad
_, _ = io.WriteString(w, "Hello")  // Never ignore errors!
```

### Error Wrapping
Use `fmt.Errorf` with `%w` for context preservation:

```go
// Good
func (s *UserService) GetUser(id int) (*User, error) {
    user, err := s.repo.FindUser(id)
    if err != nil {
        return nil, fmt.Errorf("failed to find user %d: %w", id, err)
    }
    return user, nil
}

// Bad
func (s *UserService) GetUser(id int) (*User, error) {
    user, err := s.repo.FindUser(id)
    if err != nil {
        return nil, err  // Context lost
    }
    return user, nil
}
```

### Custom Error Types
Create custom error types for better error handling:

```go
type ValidationError struct {
    Field string
    Value interface{}
    Rule  string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Rule)
}

func (e ValidationError) Is(target error) bool {
    _, ok := target.(ValidationError)
    return ok
}

// Usage
if errors.Is(err, ValidationError{}) {
    // Handle validation error
}
```

### Error Values
Define error variables for common errors:

```go
var (
    ErrUserNotFound = errors.New("user not found")
    ErrInvalidInput = errors.New("invalid input")
    ErrUnauthorized = errors.New("unauthorized access")
)
```

## Concurrency

### Goroutines
- Don't create goroutines without a clear lifecycle management plan
- Use structured concurrency patterns
- Always handle panic in goroutines

```go
// Good
func (s *Server) Start() error {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    errChan := make(chan error, 1)

    go func() {
        defer func() {
            if r := recover(); r != nil {
                errChan <- fmt.Errorf("panic in server: %v", r)
            }
        }()
        errChan <- s.httpServer.ListenAndServe()
    }()

    select {
    case err := <-errChan:
        return err
    case <-ctx.Done():
        return s.httpServer.Shutdown(ctx)
    }
}

// Bad
func (s *Server) Start() error {
    go s.httpServer.ListenAndServe()  // No error handling, no lifecycle management
    return nil
}
```

### Channels
- Prefer buffered channels for communication
- Use `select` for channel operations
- Close channels when done

```go
// Good
type Worker struct {
    jobs    <-chan Job
    results chan<- Result
    quit    <-chan struct{}
}

func (w *Worker) Start() {
    go func() {
        for {
            select {
            case job, ok := <-w.jobs:
                if !ok {
                    return
                }
                result := w.process(job)
                w.results <- result
            case <-w.quit:
                return
            }
        }
    }()
}

// Bad
func process(jobs <-chan Job, results chan<- Result) {
    for job := range jobs {  // Will deadlock if jobs channel is not closed
        results <- job.Process()
    }
}
```

### Sync Package
- Use `sync.Mutex` for simple mutual exclusion
- Use `sync.RWMutex` for read-heavy workloads
- Consider `sync/atomic` for simple operations

```go
// Good
type Counter struct {
    mu    sync.RWMutex
    count int
}

func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

func (c *Counter) Value() int {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.count
}

// Good for simple operations
type Flag struct {
    value int32
}

func (f *Flag) Set() {
    atomic.StoreInt32(&f.value, 1)
}

func (f *Flag) IsSet() bool {
    return atomic.LoadInt32(&f.value) == 1
}
```

## Testing

### Table-Driven Tests
Use table-driven tests for multiple scenarios:

```go
func TestUserService_ValidateEmail(t *testing.T) {
    tests := []struct {
        name     string
        email    string
        wantErr  bool
        errType  error
    }{
        {
            name:    "valid email",
            email:   "user@example.com",
            wantErr: false,
        },
        {
            name:    "invalid format",
            email:   "invalid-email",
            wantErr: true,
            errType: ValidationError,
        },
        {
            name:    "empty email",
            email:   "",
            wantErr: true,
            errType: ErrInvalidInput,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)

            if tt.wantErr {
                if err == nil {
                    t.Errorf("ValidateEmail() expected error, got nil")
                    return
                }
                if tt.errType != nil && !errors.Is(err, tt.errType) {
                    t.Errorf("ValidateEmail() expected error type %v, got %v", tt.errType, err)
                }
                return
            }

            if err != nil {
                t.Errorf("ValidateEmail() unexpected error: %v", err)
            }
        })
    }
}
```

### Mocks and Fakes
Use interfaces and test doubles:

```go
// Mock implementation
type MockUserRepository struct {
    users map[int]*User
    mu    sync.Mutex
}

func (m *MockUserRepository) FindUser(id int) (*User, error) {
    m.mu.Lock()
    defer m.mu.Unlock()

    user, ok := m.users[id]
    if !ok {
        return nil, ErrUserNotFound
    }
    return user, nil
}

// Test using mock
func TestUserService_GetUser(t *testing.T) {
    mockRepo := &MockUserRepository{
        users: map[int]*User{
            1: {ID: 1, Name: "John Doe"},
        },
    }

    service := NewUserService(mockRepo)

    user, err := service.GetUser(1)

    if err != nil {
        t.Fatalf("GetUser() unexpected error: %v", err)
    }

    if user.Name != "John Doe" {
        t.Errorf("GetUser() name = %v, want %v", user.Name, "John Doe")
    }
}
```

### Test Coverage
- Aim for 80%+ coverage
- Focus on testing business logic
- Test edge cases and error conditions
- Use `go test -race` for race condition detection

## Documentation

### Package Documentation
Every package should have a package comment:

```go
// Package auth provides authentication and authorization functionality.
// It supports JWT tokens, OAuth2, and basic authentication.
//
// Basic usage:
//
//   auth := auth.New(config.AuthConfig)
//   token, err := auth.Login(username, password)
//
package auth
```

### Function Documentation
Document exported functions with examples:

```go
// NewClient creates a new HTTP client with the given configuration.
// The client will use exponential backoff for retries and will
// timeout after the configured duration.
//
// Example:
//   client := NewClient(ClientConfig{
//       Timeout: 30 * time.Second,
//       Retries: 3,
//   })
//   resp, err := client.Get("https://api.example.com")
//
func NewClient(config ClientConfig) *Client {
    // implementation
}
```

### Code Comments
- Comment the why, not the what
- Exported functions and types must have comments
- Use `// TODO:` for temporary notes
- Use `// FIXME:` for known bugs

## Dependencies

### Version Management
- Use semantic versioning for your modules
- Pin dependency versions in go.mod
- Use `go mod tidy` regularly

```bash
# Add a dependency
go get github.com/pkg/errors@v1.0.0

# Update dependencies
go get -u ./...

# Tidy up dependencies
go mod tidy
```

### Dependency Selection
- Prefer well-maintained libraries
- Check for active development and community
- Avoid dependencies with many transitive dependencies
- Use standard library when possible

## Performance

### Profiling
Use the built-in profiler:

```go
import (
    _ "net/http/pprof"
    "net/http"
)

func main() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()

    // Application code
}
```

### Memory Management
- Reuse buffers and pools
- Avoid allocations in hot paths
- Use `strings.Builder` for string concatenation

```go
// Good - Use pool for frequently allocated objects
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 1024)
    },
}

func processData() {
    buf := bufferPool.Get().([]byte)
    defer bufferPool.Put(buf)

    // Use buffer
}

// Good - Use strings.Builder
func buildString(parts []string) string {
    var builder strings.Builder
    builder.Grow(len(parts) * 10) // Pre-allocate

    for _, part := range parts {
        builder.WriteString(part)
    }

    return builder.String()
}

// Bad
func buildString(parts []string) string {
    var result string
    for _, part := range parts {
        result += part  // Allocates new string each iteration
    }
    return result
}
```

### Concurrency Performance
- Use worker pools for CPU-bound tasks
- Use buffered channels to reduce blocking
- Consider `context` for cancellation and timeouts

## Security

### Input Validation
Always validate input:

```go
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
}

func (s *UserService) CreateUser(req CreateUserRequest) (*User, error) {
    if err := s.validator.Struct(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    // Check for existing user
    if exists, err := s.repo.EmailExists(req.Email); err != nil {
        return nil, err
    } else if exists {
        return nil, ErrUserAlreadyExists
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, fmt.Errorf("failed to hash password: %w", err)
    }

    // Create user
    user := &User{
        Email:    req.Email,
        Password: string(hashedPassword),
        Name:     req.Name,
    }

    return s.repo.Create(user)
}
```

### Secrets Management
- Never hardcode secrets
- Use environment variables or secret managers
- Log sensitive data carefully

```go
// Good
type Config struct {
    DatabaseURL string `env:"DATABASE_URL"`
    JWTSecret   string `env:"JWT_SECRET"`
    APIKey      string `env:"API_KEY"`
}

func LoadConfig() (*Config, error) {
    var cfg Config
    if err := env.Parse(&cfg); err != nil {
        return nil, fmt.Errorf("failed to parse environment: %w", err)
    }

    return &cfg, nil
}

// Bad
const (
    DatabaseURL = "postgres://user:pass@localhost/db"
    JWTSecret   = "super-secret-key"
    APIKey      = "api-key-123"
)
```

## Pre-commit & CI Guidelines

### Local Pre-commit Hooks
Our pre-commit configuration ensures code quality:

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v5.0.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files

  - repo: https://github.com/tekwizely/pre-commit-golang
    rev: v1.0.0-rc.4
    hooks:
      # Dependency management
      - id: go-mod-tidy

      # Code formatting
      - id: go-fmt
      - id: go-imports

      # Build verification
      - id: go-build-mod

      # Static analysis
      - id: go-vet-mod

      # Comprehensive linting
      - id: golangci-lint-mod
        args: [--timeout=5m]
```

### golangci-lint Configuration
Create `.golangci.yml` for consistent linting:

```yaml
# .golangci.yml
run:
  timeout: 5m
  modules-download-mode: readonly

linters:
  disable-all: true
  enable:
    # Default linters
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - typecheck
    - unused

    # Additional linters
    - gofmt
    - goimports
    - gosec
    - misspell
    - unconvert
    - unparam
    - nakedret
    - prealloc
    - scopelint
    - gocritic

linters-settings:
  gosec:
    excludes:
      - G204  # Subprocess launched with potential tainted input

  gocritic:
    enabled-tags:
      - performance
      - style
      - opinionated

    disabled-checks:
      - dupImport
      - ifElseChain
      - octalLiteral
      - whyNoLint
      - wrapperFunc

issues:
  exclude-rules:
    # Exclude some linters from running on tests files
    - path: _test\.go
      linters:
        - gocritic
        - errcheck
        - dupl
        - gosec
        - lll

  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0

output:
  format: colored-line-number
  print-issued-lines: true
  print-linter-name: true
```

### CI Pipeline Best Practices
1. **Run tests first** - Fail fast on test failures
2. **Parallelize linters** - Run linting in parallel with tests
3. **Skip caching** - Avoid caching issues; fresh builds ensure reliability
4. **Run on multiple OS** - Ensure cross-platform compatibility
5. **Generate coverage reports** - Track test coverage over time

### Checklist for Commits
Before committing, ensure:
- [ ] Code is formatted (`gofmt`, `goimports`)
- [ ] All tests pass (`go test ./...`)
- [ ] No linting errors (`golangci-lint run`)
- [ ] Dependencies are tidy (`go mod tidy`)
- [ ] Documentation is updated
- [ ] Error handling is comprehensive
- [ ] No TODOs or FIXMEs left in production code

## Additional Resources

- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Proverbs](https://go-proverbs.github.io/)
- [The Go Blog](https://blog.golang.org/)
- [golangci-lint Configuration](https://golangci-lint.run/usage/configuration/)

## Conclusion

Following these guidelines will help create maintainable, performant, and secure Go applications. Remember that guidelines are not strict rules - adapt them to your project's specific needs while maintaining consistency across the codebase.

Regular code reviews, refactoring, and staying up-to-date with Go best practices are essential for long-term project success.
