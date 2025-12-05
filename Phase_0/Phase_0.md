# 🧰 Phase 0: Environment, Toolchain & Mental Model

[← Back to Main Roadmap](../README.md)

---

**Objective:** Establish the development environment and internalize Go's philosophical foundations before writing code.

**Prerequisites:** Basic programming knowledge, familiarity with command-line interfaces.

**Estimated Duration:** 1-2 weeks

---

## 📋 Table of Contents

1. [Historical Context & Design Philosophy](#01-historical-context--design-philosophy)
2. [Compilation Model Deep Dive](#02-compilation-model-deep-dive)
3. [The Go Toolchain](#03-the-go-toolchain-the-go-command)
4. [Module System & Dependency Management](#04-module-system--dependency-management)
5. [Project Structure & Organization](#05-project-structure--organization)
6. [Modern Tooling](#06-modern-tooling-2024-2025-standards)
7. [Build Tags & Conditional Compilation](#07-build-tags--conditional-compilation)
8. [Interview Questions](#interview-questions)

---

## 0.1 Historical Context & Design Philosophy

### The Origin Story

**Interview Question:** *"Why was Go created? What problems does it solve?"*

Go was created at Google in 2007 by Robert Griesemer, Rob Pike, and Ken Thompson—legends who previously created Unix, Plan 9, UTF-8, and contributed to C. The language was publicly announced in 2009 and reached version 1.0 in 2012.

#### The Catalyst

The famous story goes that Go was conceived during a 45-minute C++ compilation. The creators were frustrated with:

1. **Slow compilation times** — Large C++ codebases took hours to build
2. **Complexity explosion** — C++11/14 added features that made the language harder to learn and maintain
3. **Dependency management chaos** — No standard way to manage third-party code
4. **Concurrency as an afterthought** — Threading models were bolted on, not designed in

#### The Creators' Backgrounds

| Creator | Background | Contribution to Go |
|---------|------------|-------------------|
| **Ken Thompson** | Created Unix, B language, Plan 9, UTF-8 | Systems programming DNA, simplicity focus |
| **Rob Pike** | Plan 9, Inferno, UTF-8, Limbo | Concurrency model (CSP), language design |
| **Robert Griesemer** | V8 JavaScript engine, Java HotSpot | Compiler architecture, type system |

Their combined experience of 100+ years in systems programming shaped Go's "less is more" philosophy.

### The Design Philosophy

**Interview Question:** *"What is idiomatic Go? How is Go's philosophy different from other languages?"*

#### 1. Simplicity Over Features

Go intentionally omits features that other languages consider essential:

| Omitted Feature | Why Omitted | Go Alternative |
|-----------------|-------------|----------------|
| Inheritance | Creates brittle hierarchies | Composition via embedding |
| Exceptions | Hidden control flow | Explicit error returns |
| Generics (until 1.18) | Complexity vs. benefit | Interfaces, code generation |
| Method overloading | Ambiguity in dispatch | Unique method names |
| Default parameters | Obscures function calls | Functional options pattern |
| Operator overloading | Surprises in arithmetic | Explicit method calls |

**Key Insight:** Go shifts complexity from the language to the developer's judgment. You can't hide complexity in clever abstractions—you must face it explicitly.

#### 2. Orthogonality Principle

Features in Go are designed to compose without interference:

- Goroutines work with any function
- Interfaces work with any type
- Channels work with any data type
- `defer` works in any function

No special cases, no exceptions to rules.

#### 3. The "One Way" Culture

Go enforces a single canonical style:

```go
// WRONG - Go will not compile this
if condition
{
    // ...
}

// RIGHT - Opening brace on same line (enforced by specification)
if condition {
    // ...
}
```

The `go fmt` tool removes all style debates. Every Go codebase looks the same.

**Why This Matters:**
- New team members are productive immediately
- Code reviews focus on logic, not style
- Automated refactoring is reliable

### Example: Philosophy in Action

```go
// Other languages might have:
// try { file.Read() } catch (IOException e) { ... }

// Go makes error handling explicit:
data, err := file.Read()
if err != nil {
    return fmt.Errorf("reading file: %w", err)
}
// Continue with data...
```

This forces you to think about every error condition—no hidden paths.

---

## 0.2 Compilation Model Deep Dive

### AOT (Ahead-of-Time) Compilation

**Interview Question:** *"How does Go's compilation differ from Java or Python? What are the trade-offs?"*

| Aspect | Go (AOT) | Java (JIT) | Python (Interpreted) |
|--------|----------|------------|---------------------|
| **Compilation** | Source → Binary | Source → Bytecode → JIT | Source → Bytecode (at runtime) |
| **Startup Time** | Instant | Slow (JVM warmup) | Medium |
| **Peak Performance** | High (optimized ahead) | Very High (after warmup) | Low |
| **Binary Size** | Large (self-contained) | Small (needs JVM) | Tiny (needs interpreter) |
| **Deployment** | Single file | JAR + JVM | Source + interpreter |
| **Cross-compilation** | Built-in | N/A (JVM portable) | N/A |

#### How Go Compilation Works

```
Source Code (.go files)
        ↓
   Lexer/Parser
        ↓
 Abstract Syntax Tree (AST)
        ↓
   Type Checking
        ↓
  SSA (Single Static Assignment) IR
        ↓
   Optimization Passes
        ↓
  Machine Code Generation
        ↓
     Linking
        ↓
  Executable Binary
```

**Key Point:** The entire standard library and runtime are compiled and linked into your binary. A "Hello World" in Go is ~2MB because it includes:
- The Go runtime (scheduler)
- The garbage collector
- The memory allocator
- All imported packages

### Static Linking

**Interview Question:** *"Why are Go binaries self-contained? What are the benefits and drawbacks?"*

Go uses static linking by default:

```bash
# Build and check dependencies
$ go build -o myapp main.go
$ ldd myapp
    not a dynamic executable  # No shared library dependencies!
```

**Benefits:**
1. **Deployment simplicity** — Copy one file, done
2. **No DLL hell** — No version conflicts with system libraries
3. **Container-friendly** — Can use `scratch` (empty) Docker images
4. **Security** — No runtime library injection attacks

**Drawbacks:**
1. **Binary size** — Each binary is 5-20MB minimum
2. **No shared memory** — 10 Go services = 10 copies of runtime in RAM
3. **Security patches** — Must recompile to get stdlib fixes

### Cross-Compilation

**Interview Question:** *"How do you build a Go binary for Linux when developing on macOS?"*

Go's cross-compilation is trivial—no additional tools needed:

```bash
# Build for Linux AMD64 on any platform
GOOS=linux GOARCH=amd64 go build -o myapp-linux main.go

# Build for Windows ARM64
GOOS=windows GOARCH=arm64 go build -o myapp.exe main.go

# Build for macOS M1
GOOS=darwin GOARCH=arm64 go build -o myapp-mac main.go
```

#### Supported Platforms

| GOOS | GOARCH Options |
|------|----------------|
| `linux` | `amd64`, `arm64`, `386`, `arm`, `mips`, `ppc64`, `riscv64`, `s390x` |
| `darwin` | `amd64`, `arm64` |
| `windows` | `amd64`, `arm64`, `386` |
| `freebsd` | `amd64`, `arm64`, `386`, `arm` |
| `js` | `wasm` |

### CGO Trade-offs

**Interview Question:** *"What is CGO? When would you use it, and what are the risks?"*

CGO allows Go to call C code:

```go
// #include <stdlib.h>
// #include <openssl/sha.h>
import "C"

func main() {
    cs := C.CString("hello")
    defer C.free(unsafe.Pointer(cs))
    // ... use C libraries
}
```

**Trade-offs:**

| CGO Enabled | CGO Disabled (`CGO_ENABLED=0`) |
|-------------|-------------------------------|
| Can use C libraries (OpenSSL, SQLite) | Pure Go only |
| Dynamic linking (needs libc) | Static binary |
| Cross-compilation requires C toolchain | Trivial cross-compilation |
| Potential memory safety issues | Full memory safety |
| Slower function calls (FFI overhead) | No FFI overhead |

**Best Practice:** Avoid CGO when possible. Use pure Go alternatives:
- `modernc.org/sqlite` instead of CGO SQLite
- `crypto/tls` instead of OpenSSL
- Pure Go image libraries

### Example: Complete Build Workflow

```bash
# Development build (fast, includes debug info)
go build -o myapp ./cmd/api

# Production build (optimized, stripped, static)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w" -o myapp ./cmd/api

# Check binary size
ls -lh myapp
# -rwxr-xr-x 1 user user 6.2M Dec  5 10:00 myapp

# Verify static linking
file myapp
# myapp: ELF 64-bit LSB executable, x86-64, statically linked
```

---

## 0.3 The Go Toolchain (The `go` Command)

### Overview

**Interview Question:** *"What tools does Go provide out of the box? How does this compare to other ecosystems?"*

Go provides a unified CLI that replaces:
- Build tools (Make, Maven, Gradle)
- Package managers (npm, pip, Maven)
- Linters (ESLint, Pylint)
- Formatters (Prettier, Black)
- Test runners (Jest, pytest)
- Documentation generators

All with a single command: `go`

### Essential Commands Reference

#### `go build` — Compilation

```bash
# Build current package
go build

# Build specific package
go build ./cmd/api

# Build with output name
go build -o myapp ./cmd/api

# Build all packages (check compilation)
go build ./...

# Build with race detector
go build -race ./cmd/api

# Build with optimizations and stripping
go build -ldflags="-s -w" -o myapp ./cmd/api
```

**Flags Deep Dive:**

| Flag | Purpose | Example |
|------|---------|---------|
| `-o` | Output filename | `-o myapp` |
| `-v` | Verbose (show packages) | `-v` |
| `-race` | Enable race detector | `-race` |
| `-ldflags` | Linker flags | `-ldflags="-s -w"` |
| `-gcflags` | Compiler flags | `-gcflags="-m"` (escape analysis) |
| `-trimpath` | Remove file paths from binary | `-trimpath` |

#### `go run` — Compile and Execute

```bash
# Run single file
go run main.go

# Run package
go run ./cmd/api

# Run with arguments
go run main.go -port 8080

# Run with race detector
go run -race main.go
```

**Interview Question:** *"What's the difference between `go run` and `go build`?"*

`go run`:
- Compiles to a temporary directory
- Executes immediately
- Cleans up after exit
- Good for development iteration

`go build`:
- Compiles to current directory (or specified output)
- Creates persistent executable
- Used for deployment

#### `go test` — Testing Framework

```bash
# Run tests in current package
go test

# Run all tests in module
go test ./...

# Verbose output
go test -v ./...

# Run specific test
go test -run TestMyFunction ./...

# Run with coverage
go test -cover ./...

# Generate coverage profile
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. ./...

# Run with race detector
go test -race ./...

# Set timeout
go test -timeout 30s ./...
```

#### `go fmt` — Code Formatting

```bash
# Format current package
go fmt

# Format all packages
go fmt ./...

# Check if formatting is needed (CI)
gofmt -l .  # Lists files that need formatting
```

**Key Point:** `go fmt` is non-negotiable. Run it before every commit.

#### `go vet` — Static Analysis

```bash
# Vet current package
go vet

# Vet all packages
go vet ./...
```

**What `go vet` catches:**
- Printf format string mismatches
- Unreachable code
- Suspicious constructs
- Lock copying
- Nil pointer dereferences (some cases)

#### `go mod` — Module Management

```bash
# Initialize new module
go mod init github.com/user/project

# Add missing dependencies, remove unused
go mod tidy

# Download dependencies to local cache
go mod download

# Vendor dependencies into ./vendor
go mod vendor

# Show dependency graph
go mod graph

# Explain why a dependency exists
go mod why github.com/some/package

# Edit go.mod programmatically
go mod edit -require github.com/gin-gonic/gin@v1.9.0
```

#### `go get` — Dependency Management

```bash
# Add dependency
go get github.com/gin-gonic/gin

# Add specific version
go get github.com/gin-gonic/gin@v1.9.0

# Update to latest
go get -u github.com/gin-gonic/gin

# Update all dependencies
go get -u ./...

# Install tool binary
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

**Note:** Since Go 1.17, `go get` only modifies `go.mod`. Use `go install` for installing binaries.

#### `go doc` — Documentation

```bash
# View package documentation
go doc fmt

# View specific function
go doc fmt.Printf

# View all (including unexported)
go doc -all fmt

# Start documentation server
godoc -http=:6060
# Then visit http://localhost:6060
```

### Example: Complete Development Workflow

```bash
# 1. Create new project
mkdir myproject && cd myproject
go mod init github.com/user/myproject

# 2. Write code
cat > main.go << 'EOF'
package main

import "fmt"

func main() {
    fmt.Println("Hello, Go!")
}
EOF

# 3. Format and vet
go fmt ./...
go vet ./...

# 4. Run tests (create test file first)
go test -v ./...

# 5. Run during development
go run main.go

# 6. Build for production
CGO_ENABLED=0 go build -ldflags="-s -w" -o myproject

# 7. Cross-compile for deployment
GOOS=linux GOARCH=amd64 go build -o myproject-linux
```

---

## 0.4 Module System & Dependency Management

### The `go.mod` File

**Interview Question:** *"Explain Go modules. How does dependency resolution work?"*

The `go.mod` file defines your module:

```go
module github.com/user/myproject

go 1.22

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/jackc/pgx/v5 v5.5.0
)

require (
    // Indirect dependencies (transitive)
    github.com/bytedance/sonic v1.10.2 // indirect
    golang.org/x/crypto v0.16.0 // indirect
)

replace github.com/broken/lib => github.com/fixed/lib v1.0.0

exclude github.com/vulnerable/lib v0.9.0
```

**Directives:**

| Directive | Purpose |
|-----------|---------|
| `module` | Declares module path (import path) |
| `go` | Minimum Go version required |
| `require` | Direct dependencies with versions |
| `require // indirect` | Transitive dependencies |
| `replace` | Override dependency source |
| `exclude` | Prevent specific version from being used |

### The `go.sum` File

**Interview Question:** *"What is `go.sum` and why is it important for security?"*

The `go.sum` file contains cryptographic checksums:

```
github.com/gin-gonic/gin v1.9.1 h1:4idEAncQnU5cB7BeOkPtxjfCSye0AAm1R0RVIqJ+Jmg=
github.com/gin-gonic/gin v1.9.1/go.mod h1:hPrL+YrJ2F/XfmRxyzw==
```

**Security Model:**
1. First download records hash in `go.sum`
2. Future downloads verify against recorded hash
3. Tampering detection via checksum mismatch
4. Public checksum database at `sum.golang.org`

**Best Practice:** Always commit `go.sum` to version control.

### Minimal Version Selection (MVS)

**Interview Question:** *"How does Go's dependency resolution differ from npm or Maven?"*

Go uses Minimal Version Selection:

```
Your project requires:     A v1.0.0
A v1.0.0 requires:        B v1.2.0
Another dep requires:     B v1.5.0

Result: B v1.5.0 is selected (minimum version that satisfies all)
```

**Contrast with other systems:**

| System | Strategy | Result |
|--------|----------|--------|
| Go (MVS) | Minimum satisfying version | Deterministic, reproducible |
| npm | Latest compatible version | May break unexpectedly |
| Maven | Nearest wins | Can pick older version |

**Benefits of MVS:**
- Reproducible builds without lockfile
- No "it works on my machine" issues
- Predictable upgrade paths

### Private Modules

**Interview Question:** *"How do you use private Git repositories as Go dependencies in an enterprise environment?"*

```bash
# Tell Go these paths are private (don't use public proxy)
export GOPRIVATE=github.com/mycompany/*,gitlab.internal.com/*

# Disable checksum database for private modules
export GONOSUMDB=github.com/mycompany/*

# Optionally disable proxy entirely for private
export GONOPROXY=github.com/mycompany/*
```

**Authentication:**

```bash
# Git credential helper for HTTPS
git config --global credential.helper store

# Or use SSH
git config --global url."git@github.com:".insteadOf "https://github.com/"

# Or use .netrc
echo "machine github.com login user password token" >> ~/.netrc
```

### Workspace Mode (`go.work`)

**Interview Question:** *"How do you develop multiple related modules locally?"*

Go 1.18+ supports workspaces:

```bash
# Project structure
mywork/
├── go.work        # Workspace file
├── api-service/
│   ├── go.mod
│   └── main.go
└── shared-lib/
    ├── go.mod
    └── utils.go
```

```go
// go.work
go 1.22

use (
    ./api-service
    ./shared-lib
)
```

**Benefits:**
- Edit `shared-lib` and `api-service` sees changes immediately
- No need for `replace` directives
- Not committed to version control (developer-local)

### Example: Managing Dependencies

```bash
# Add a new dependency
go get github.com/redis/go-redis/v9

# Update a specific dependency
go get -u github.com/gin-gonic/gin

# Update all dependencies
go get -u ./...

# Clean up unused dependencies
go mod tidy

# Check for available updates
go list -m -u all

# Downgrade a dependency
go get github.com/gin-gonic/gin@v1.8.0

# Use replace for local development
go mod edit -replace github.com/user/lib=../local-lib
```

---

## 0.5 Project Structure & Organization

### Executable vs. Library

**Interview Question:** *"What distinguishes a Go executable from a library?"*

**Executable (Application):**
```go
// main.go
package main  // Must be "main"

func main() { // Must have this function
    // Entry point
}
```

**Library (Dependency):**
```go
// calculator/math.go
package calculator  // Any name except "main"

func Add(a, b int) int {
    return a + b
}
// No main() function
```

### Standard Project Layout

**Interview Question:** *"Describe a typical Go project structure. What are the conventions?"*

```
myproject/
├── cmd/                    # Entry points (executables)
│   ├── api/
│   │   └── main.go        # go build ./cmd/api
│   └── worker/
│       └── main.go        # go build ./cmd/worker
│
├── internal/              # Private packages (compiler-enforced)
│   ├── handler/
│   │   └── user.go
│   ├── service/
│   │   └── user.go
│   └── repository/
│       └── user.go
│
├── pkg/                   # Public packages (optional, debated)
│   └── httpclient/
│       └── client.go
│
├── api/                   # API definitions (OpenAPI, Proto)
│   └── openapi.yaml
│
├── web/                   # Web assets (if applicable)
│   └── static/
│
├── configs/               # Configuration files
│   └── config.yaml
│
├── scripts/               # Build/deployment scripts
│   └── build.sh
│
├── test/                  # Integration tests
│   └── integration_test.go
│
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### The `cmd/` Convention

For projects with multiple executables:

```go
// cmd/api/main.go
package main

import "github.com/user/myproject/internal/handler"

func main() {
    h := handler.New()
    http.ListenAndServe(":8080", h)
}

// cmd/worker/main.go
package main

import "github.com/user/myproject/internal/worker"

func main() {
    w := worker.New()
    w.Run()
}
```

Build each separately:
```bash
go build -o api-server ./cmd/api
go build -o background-worker ./cmd/worker
```

### The `internal/` Enforced Privacy

**Interview Question:** *"What is special about the `internal/` directory in Go?"*

The Go compiler **enforces** that `internal/` packages can only be imported by code in the same module subtree:

```
myproject/
├── cmd/
│   └── api/
│       └── main.go      # CAN import internal/handler
├── internal/
│   └── handler/
│       └── user.go
└── go.mod
```

```go
// ❌ Another project trying to import:
import "github.com/user/myproject/internal/handler"
// Error: use of internal package not allowed
```

**Use Case:** Hide implementation details while exposing a public API.

### The `pkg/` Debate

**Interview Question:** *"Should you use a `pkg/` directory? What's the controversy?"*

**Historical Context:**
- Kubernetes, Docker, and early Go projects used `pkg/`
- Meant to signal "these packages are safe to import"

**Modern View:**
- Adds unnecessary nesting
- Everything not in `internal/` is already public
- Many projects skip it entirely

**Recommendation:** Use `pkg/` only if you have a clear public API that benefits from separation. Otherwise, put packages at the root.

### Example: Minimal vs. Full Structure

**Minimal (small project):**
```
myapp/
├── main.go
├── handler.go
├── handler_test.go
├── go.mod
└── go.sum
```

**Full (large project):**
```
myplatform/
├── cmd/
│   ├── api/
│   ├── worker/
│   └── cli/
├── internal/
│   ├── domain/
│   ├── handler/
│   ├── service/
│   └── repository/
├── pkg/
│   └── sdk/
├── test/
├── configs/
├── scripts/
├── go.mod
└── go.sum
```

---

## 0.6 Modern Tooling (2024-2025 Standards)

### gopls (Language Server)

**Interview Question:** *"What is gopls and why is it important?"*

`gopls` is the official Go language server providing:
- Autocomplete
- Go to definition
- Find references
- Rename refactoring
- Code actions

**Installation:**
```bash
go install golang.org/x/tools/gopls@latest
```

**VS Code Integration:**
Install "Go" extension → automatically uses `gopls`

### govulncheck (Security Scanning)

**Interview Question:** *"How do you check for vulnerabilities in Go dependencies?"*

```bash
# Install
go install golang.org/x/vuln/cmd/govulncheck@latest

# Scan current module
govulncheck ./...

# Output example:
# Vulnerability #1: GO-2023-2186
#   A malformed DNS message can cause a panic in net.
#   ...
```

**Best Practice:** Run in CI pipeline before deployment.

### staticcheck (Extended Analysis)

Beyond `go vet`, `staticcheck` finds:
- Unused code
- Deprecated function usage
- Performance issues
- Stylistic problems

```bash
# Install
go install honnef.co/go/tools/cmd/staticcheck@latest

# Run
staticcheck ./...
```

### golangci-lint (Aggregated Linting)

**Interview Question:** *"What linting tools do you use for Go projects?"*

`golangci-lint` runs 50+ linters in parallel:

```bash
# Install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run with default linters
golangci-lint run

# Run all linters
golangci-lint run --enable-all
```

**Configuration (`.golangci.yml`):**
```yaml
linters:
  enable:
    - gofmt
    - govet
    - errcheck
    - staticcheck
    - gosec
    - gocritic
    - prealloc
  
linters-settings:
  govet:
    check-shadowing: true
  
issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - errcheck
```

### Delve Debugger

**Interview Question:** *"How do you debug Go applications? Can you inspect goroutines?"*

```bash
# Install
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug a program
dlv debug ./cmd/api

# Debug existing binary
dlv exec ./myapp

# Attach to running process
dlv attach <pid>

# Debug tests
dlv test ./...
```

**Delve Commands:**
```
(dlv) break main.go:25    # Set breakpoint
(dlv) continue            # Run until breakpoint
(dlv) next                # Step over
(dlv) step                # Step into
(dlv) print myVar         # Print variable
(dlv) goroutines          # List all goroutines
(dlv) goroutine 5         # Switch to goroutine 5
(dlv) stack               # Show stack trace
```

### Example: Complete Linting Setup

```bash
# Create configuration
cat > .golangci.yml << 'EOF'
run:
  timeout: 5m

linters:
  enable:
    - gofmt
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - ineffassign
    - typecheck

issues:
  max-issues-per-linter: 0
  max-same-issues: 0
EOF

# Run linting
golangci-lint run ./...

# Fix auto-fixable issues
golangci-lint run --fix ./...
```

---

## 0.7 Build Tags & Conditional Compilation

### Build Tag Syntax

**Interview Question:** *"How do you include different code for different platforms or build conditions?"*

```go
//go:build linux
// +build linux (old syntax, still works)

package main

// This file only compiles on Linux
```

**Boolean Expressions:**
```go
//go:build linux && amd64
//go:build linux || darwin
//go:build !windows
//go:build (linux || darwin) && !cgo
```

### Platform-Specific Code

**File naming convention:**
```
mypackage/
├── file.go           # All platforms
├── file_linux.go     # Linux only
├── file_darwin.go    # macOS only
├── file_windows.go   # Windows only
├── file_amd64.go     # AMD64 only
└── file_linux_amd64.go  # Linux + AMD64
```

**Example: Platform-specific implementation:**

```go
// signal_unix.go
//go:build unix

package main

import "syscall"

func setupSignals() {
    // Unix-specific signal handling
    signal.Notify(c, syscall.SIGTERM)
}

// signal_windows.go
//go:build windows

package main

func setupSignals() {
    // Windows-specific signal handling
}
```

### Integration Test Isolation

```go
//go:build integration

package mypackage

import "testing"

func TestDatabaseIntegration(t *testing.T) {
    // This test only runs with: go test -tags=integration
}
```

```bash
# Run only unit tests
go test ./...

# Run with integration tests
go test -tags=integration ./...
```

### Example: Feature Flags

```go
//go:build feature_newui

package main

func init() {
    enableNewUI = true
}
```

```bash
# Build with feature enabled
go build -tags=feature_newui -o myapp ./cmd/api
```

---

## Interview Questions

### Beginner Level

1. **Q:** What command creates a new Go module?
   **A:** `go mod init <module-path>`

2. **Q:** How do you format Go code?
   **A:** `go fmt ./...` — it's non-negotiable and enforced.

3. **Q:** What is the entry point of a Go program?
   **A:** The `main` function in `package main`.

### Intermediate Level

4. **Q:** How would you cross-compile a Go binary for Linux ARM64 on macOS?
   **A:** `GOOS=linux GOARCH=arm64 go build -o myapp ./cmd/api`

5. **Q:** What's the difference between `internal/` and `pkg/` directories?
   **A:** `internal/` is compiler-enforced private (can't be imported outside module), `pkg/` is a convention for public packages.

6. **Q:** How does `go.sum` protect against supply chain attacks?
   **A:** It stores cryptographic hashes of dependencies, verified against a global checksum database.

### Advanced Level

7. **Q:** Explain Minimal Version Selection. Why did Go choose this over other approaches?
   **A:** MVS selects the minimum version satisfying all constraints. It's deterministic without lockfiles, produces reproducible builds, and avoids "version creep" issues.

8. **Q:** When would you use CGO, and what are the trade-offs?
   **A:** CGO is needed for C libraries (e.g., hardware drivers, some crypto). Trade-offs: breaks cross-compilation, adds dynamic linking, potential memory safety issues, slower FFI calls.

9. **Q:** How do you debug goroutine-related issues?
   **A:** Use `dlv` (Delve) with `goroutines` command, or add `-race` flag to detect data races, or use `runtime.Stack()` for programmatic inspection.

---

## Summary

Phase 0 establishes the foundation for Go development:

| Topic | Key Takeaway |
|-------|--------------|
| Philosophy | Simplicity is intentional; explicit over implicit |
| Compilation | AOT, static linking, easy cross-compilation |
| Toolchain | One `go` command replaces entire ecosystem |
| Modules | `go.mod` + `go.sum` + MVS = reproducible builds |
| Structure | `cmd/`, `internal/`, flat packages |
| Tooling | `gopls`, `golangci-lint`, `dlv` are essential |
| Build Tags | `//go:build` for conditional compilation |

**Next Phase:** [Phase 1 — Lexical Elements & Language Fundamentals](../Phase_1/Phase_1.md)
