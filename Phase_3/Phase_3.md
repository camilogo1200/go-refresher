# ⚠️ Phase 3: Error Handling — The Go Way

[← Back to Main Roadmap](../README.md) | [← Previous: Phase 2](../Phase_2/Phase_2.md)

---

**Objective:** Master Go's explicit error handling philosophy as a core language feature, not an afterthought.

**Reference:** [Go Blog - Error Handling](https://go.dev/blog/error-handling-and-go)

**Prerequisites:** Phase 0-2

**Estimated Duration:** 1-2 weeks

---

## 📋 Table of Contents

1. [The `error` Interface](#31-the-error-interface)
2. [Creating Errors](#32-creating-errors)
3. [Error Wrapping](#33-error-wrapping-go-113)
4. [Error Handling Patterns](#34-error-handling-patterns)
5. [Panic and Recover](#35-panic-and-recover)
6. [Error Handling in Practice](#36-error-handling-in-practice)
7. [Interview Questions](#interview-questions)

---

## 3.1 The `error` Interface

### Definition

**Interview Question:** *"What is the `error` interface in Go? Why doesn't Go have exceptions?"*

```go
// The error interface (built-in)
type error interface {
    Error() string
}
```

**Key characteristics:**
- It's just an interface with one method
- Any type implementing `Error() string` is an error
- Errors are **values**, not exceptions

### Why Not Exceptions?

**Interview Question:** *"What are the advantages of Go's error handling over exceptions?"*

| Exceptions (Java/Python) | Errors as Values (Go) |
|--------------------------|----------------------|
| Hidden control flow | Explicit control flow |
| Can be ignored silently | Must be explicitly handled |
| Stack unwinding overhead | No special runtime cost |
| Try-catch boilerplate | if err != nil pattern |
| Unclear what can throw | Return type is clear |

```go
// Go's explicit approach
result, err := doSomething()
if err != nil {
    return fmt.Errorf("doing something: %w", err)
}
// Continue with result...

// vs. hidden exception flow:
// try { result = doSomething() } catch (Exception e) { ... }
```

### The `if err != nil` Pattern

**Interview Question:** *"Why is Go code full of `if err != nil`? Isn't this verbose?"*

```go
func processData(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()
    
    data, err := io.ReadAll(file)
    if err != nil {
        return err
    }
    
    result, err := parse(data)
    if err != nil {
        return err
    }
    
    return save(result)
}
```

**Philosophy:**
- Errors are expected, not exceptional
- Every error path is visible
- Forces developer to think about failure modes
- Self-documenting code

### Example: Basic Error Handling

```go
package main

import (
    "errors"
    "fmt"
    "strconv"
)

func parsePort(s string) (int, error) {
    port, err := strconv.Atoi(s)
    if err != nil {
        return 0, fmt.Errorf("invalid port %q: %w", s, err)
    }
    if port < 1 || port > 65535 {
        return 0, fmt.Errorf("port %d out of range [1-65535]", port)
    }
    return port, nil
}

func main() {
    port, err := parsePort("8080")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Printf("Port: %d\n", port)
}
```

---

## 3.2 Creating Errors

### `errors.New()` — Simple Static Errors

**Interview Question:** *"When would you use `errors.New()` vs. `fmt.Errorf()`?"*

```go
import "errors"

// Static error message
var ErrNotFound = errors.New("not found")
var ErrInvalidInput = errors.New("invalid input")

func findUser(id int) (*User, error) {
    // ...
    if notFound {
        return nil, ErrNotFound
    }
    return user, nil
}
```

**Use `errors.New()` for:**
- Sentinel errors (package-level)
- Simple, unchanging messages
- When error identity matters (comparison)

### `fmt.Errorf()` — Formatted Errors

```go
import "fmt"

func validateAge(age int) error {
    if age < 0 {
        return fmt.Errorf("age cannot be negative: %d", age)
    }
    if age > 150 {
        return fmt.Errorf("age %d exceeds maximum allowed", age)
    }
    return nil
}
```

**Use `fmt.Errorf()` for:**
- Dynamic error messages
- Including context values
- Error wrapping (with `%w`)

### Sentinel Errors

**Interview Question:** *"What are sentinel errors? Give examples from the standard library."*

Sentinel errors are package-level error values for comparison:

```go
// Standard library examples
io.EOF           // End of file/stream
sql.ErrNoRows    // No rows in result
context.Canceled // Context was canceled
context.DeadlineExceeded

// Define your own
var (
    ErrNotFound     = errors.New("resource not found")
    ErrUnauthorized = errors.New("unauthorized")
    ErrConflict     = errors.New("resource conflict")
)

// Usage
if err == sql.ErrNoRows {
    return nil, ErrNotFound
}
```

### Custom Error Types

**Interview Question:** *"When should you create a custom error type?"*

```go
// Custom error type with additional context
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}

// Custom error with HTTP status
type HTTPError struct {
    StatusCode int
    Message    string
}

func (e *HTTPError) Error() string {
    return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// Usage
func validateRequest(req *Request) error {
    if req.Email == "" {
        return &ValidationError{
            Field:   "email",
            Message: "required",
        }
    }
    return nil
}
```

**Use custom error types when:**
- Error needs structured data
- Caller needs to extract information
- Different error behaviors needed

### Example: Comprehensive Error Types

```go
package repository

import (
    "errors"
    "fmt"
)

// Sentinel errors
var (
    ErrNotFound     = errors.New("entity not found")
    ErrDuplicate    = errors.New("duplicate entity")
)

// Custom error type with context
type QueryError struct {
    Query   string
    Wrapped error
}

func (e *QueryError) Error() string {
    return fmt.Sprintf("query failed [%s]: %v", e.Query, e.Wrapped)
}

func (e *QueryError) Unwrap() error {
    return e.Wrapped
}

// Behavior interface for retryable errors
type RetryableError interface {
    error
    Retryable() bool
}

type TransientError struct {
    Message string
}

func (e *TransientError) Error() string {
    return e.Message
}

func (e *TransientError) Retryable() bool {
    return true
}
```

---

## 3.3 Error Wrapping (Go 1.13+)

### Wrapping with `%w`

**Interview Question:** *"How do you wrap errors in Go? What's the difference between `%w` and `%v`?"*

```go
// Wrap error with context
func readConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("reading config %s: %w", path, err)
    }
    
    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parsing config: %w", err)
    }
    
    return &cfg, nil
}
```

**`%w` vs `%v`:**

| `%w` | `%v` |
|------|------|
| Wraps error (preserves chain) | Converts to string |
| `errors.Is/As` work | Chain is broken |
| Use when caller needs original | Use when hiding implementation |

```go
// %w - preserves chain
err := fmt.Errorf("operation failed: %w", originalErr)
errors.Is(err, originalErr)  // true

// %v - breaks chain
err := fmt.Errorf("operation failed: %v", originalErr)
errors.Is(err, originalErr)  // false
```

### `errors.Unwrap()`

```go
wrapped := fmt.Errorf("outer: %w", 
    fmt.Errorf("middle: %w", 
        errors.New("inner")))

// Manual unwrapping
err := errors.Unwrap(wrapped)  // "middle: inner"
err = errors.Unwrap(err)       // "inner"
err = errors.Unwrap(err)       // nil
```

### `errors.Is()` — Checking Error Chain

**Interview Question:** *"When would you use `errors.Is()` vs `==` for error comparison?"*

```go
// errors.Is checks the entire chain
func handleError(err error) {
    if errors.Is(err, sql.ErrNoRows) {
        // Handle not found
    } else if errors.Is(err, context.DeadlineExceeded) {
        // Handle timeout
    }
}

// Custom Is method for specialized matching
type NotFoundError struct {
    Resource string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s not found", e.Resource)
}

func (e *NotFoundError) Is(target error) bool {
    // Match any NotFoundError regardless of Resource
    _, ok := target.(*NotFoundError)
    return ok
}
```

### `errors.As()` — Extracting Error Types

**Interview Question:** *"How do you extract a specific error type from a wrapped error chain?"*

```go
func handleHTTPError(err error) {
    var httpErr *HTTPError
    if errors.As(err, &httpErr) {
        fmt.Printf("HTTP %d: %s\n", httpErr.StatusCode, httpErr.Message)
        return
    }
    
    var validErr *ValidationError
    if errors.As(err, &validErr) {
        fmt.Printf("Invalid %s: %s\n", validErr.Field, validErr.Message)
        return
    }
    
    // Unknown error type
    fmt.Printf("Error: %v\n", err)
}
```

### Implementing Unwrap

```go
// Single wrapped error
type PathError struct {
    Op   string
    Path string
    Err  error
}

func (e *PathError) Error() string {
    return fmt.Sprintf("%s %s: %v", e.Op, e.Path, e.Err)
}

func (e *PathError) Unwrap() error {
    return e.Err
}

// Multiple wrapped errors (Go 1.20+)
type MultiError struct {
    Errors []error
}

func (e *MultiError) Error() string {
    return fmt.Sprintf("%d errors occurred", len(e.Errors))
}

func (e *MultiError) Unwrap() []error {
    return e.Errors
}
```

### Example: Error Wrapping in Practice

```go
package main

import (
    "errors"
    "fmt"
    "os"
)

var ErrPermissionDenied = errors.New("permission denied")

type FileError struct {
    Path string
    Err  error
}

func (e *FileError) Error() string {
    return fmt.Sprintf("file %s: %v", e.Path, e.Err)
}

func (e *FileError) Unwrap() error {
    return e.Err
}

func readSecureFile(path string) ([]byte, error) {
    if !hasPermission(path) {
        return nil, &FileError{
            Path: path,
            Err:  ErrPermissionDenied,
        }
    }
    
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, &FileError{
            Path: path,
            Err:  err,
        }
    }
    
    return data, nil
}

func main() {
    _, err := readSecureFile("/etc/shadow")
    
    // Check specific error
    if errors.Is(err, ErrPermissionDenied) {
        fmt.Println("Access denied!")
    }
    
    // Extract error details
    var fileErr *FileError
    if errors.As(err, &fileErr) {
        fmt.Printf("Failed on file: %s\n", fileErr.Path)
    }
}
```

---

## 3.4 Error Handling Patterns

### Early Return Pattern

**Interview Question:** *"What is the early return pattern? Why is it preferred?"*

```go
// BAD: Deep nesting
func processRequest(r *Request) error {
    if valid, err := validate(r); err == nil {
        if valid {
            if data, err := fetch(r); err == nil {
                if result, err := transform(data); err == nil {
                    return save(result)
                } else {
                    return err
                }
            } else {
                return err
            }
        }
        return errors.New("invalid request")
    } else {
        return err
    }
}

// GOOD: Early return
func processRequest(r *Request) error {
    valid, err := validate(r)
    if err != nil {
        return fmt.Errorf("validation: %w", err)
    }
    if !valid {
        return errors.New("invalid request")
    }
    
    data, err := fetch(r)
    if err != nil {
        return fmt.Errorf("fetching: %w", err)
    }
    
    result, err := transform(data)
    if err != nil {
        return fmt.Errorf("transforming: %w", err)
    }
    
    return save(result)
}
```

### Error Propagation with Context

**Interview Question:** *"How do you add context when propagating errors?"*

```go
func CreateUser(req *CreateUserRequest) (*User, error) {
    // Validate
    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }
    
    // Check uniqueness
    exists, err := userRepo.ExistsByEmail(req.Email)
    if err != nil {
        return nil, fmt.Errorf("checking email uniqueness: %w", err)
    }
    if exists {
        return nil, ErrEmailTaken
    }
    
    // Create user
    user, err := userRepo.Create(req.ToUser())
    if err != nil {
        return nil, fmt.Errorf("creating user: %w", err)
    }
    
    return user, nil
}
```

### Error Transformation

**Interview Question:** *"When would you transform an error instead of wrapping it?"*

```go
// Transform low-level errors to domain errors
func (r *UserRepository) FindByID(id string) (*User, error) {
    row := r.db.QueryRow("SELECT * FROM users WHERE id = $1", id)
    
    var user User
    err := row.Scan(&user.ID, &user.Name, &user.Email)
    
    switch {
    case errors.Is(err, sql.ErrNoRows):
        return nil, ErrUserNotFound  // Transform to domain error
    case err != nil:
        return nil, fmt.Errorf("querying user: %w", err)
    }
    
    return &user, nil
}
```

### Opaque Errors (Behavior-Based)

**Interview Question:** *"What are opaque errors? Why might you use them?"*

```go
// Define behavior, not type
type temporary interface {
    Temporary() bool
}

func isTemporary(err error) bool {
    var t temporary
    return errors.As(err, &t) && t.Temporary()
}

// Retry based on behavior
func fetchWithRetry(url string, maxRetries int) ([]byte, error) {
    var lastErr error
    for i := 0; i < maxRetries; i++ {
        data, err := fetch(url)
        if err == nil {
            return data, nil
        }
        lastErr = err
        
        if !isTemporary(err) {
            return nil, err  // Don't retry permanent errors
        }
        time.Sleep(time.Second * time.Duration(i+1))
    }
    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

### Example: Complete Error Handling Strategy

```go
package service

import (
    "context"
    "errors"
    "fmt"
)

// Domain errors
var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
    ErrConflict     = errors.New("conflict")
)

// Validation error
type ValidationError struct {
    Fields map[string]string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed: %v", e.Fields)
}

// Service layer
type UserService struct {
    repo UserRepository
}

func (s *UserService) UpdateUser(ctx context.Context, id string, update *UserUpdate) (*User, error) {
    // Authorization
    if !canEdit(ctx, id) {
        return nil, ErrUnauthorized
    }
    
    // Validation
    if errs := update.Validate(); len(errs) > 0 {
        return nil, &ValidationError{Fields: errs}
    }
    
    // Check existence
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("finding user %s: %w", id, err)
    }
    if user == nil {
        return nil, ErrNotFound
    }
    
    // Apply update
    user.Apply(update)
    
    // Save
    if err := s.repo.Save(ctx, user); err != nil {
        return nil, fmt.Errorf("saving user %s: %w", id, err)
    }
    
    return user, nil
}
```

---

## 3.5 Panic and Recover

### Understanding Panic

**Interview Question:** *"When should you use `panic()` in Go? When should you avoid it?"*

```go
// Panic stops normal execution
func mustParseURL(s string) *url.URL {
    u, err := url.Parse(s)
    if err != nil {
        panic(fmt.Sprintf("invalid URL %q: %v", s, err))
    }
    return u
}
```

**When to panic:**
- Truly unrecoverable errors
- Programming bugs (should never happen)
- Initialization failures
- `Must*` functions (convention)

**When NOT to panic:**
- Expected errors (network, user input)
- Recoverable situations
- Library code (return errors instead)

### Panic Mechanics

```go
func example() {
    defer fmt.Println("1: deferred")
    
    fmt.Println("2: before panic")
    panic("something went wrong")
    fmt.Println("3: after panic")  // Never executed
}

// Output:
// 2: before panic
// 1: deferred
// panic: something went wrong
```

**Key behavior:**
1. Normal execution stops
2. Deferred functions run (LIFO)
3. Panic propagates up call stack
4. Program crashes if not recovered

### Recover

**Interview Question:** *"How does `recover()` work? Where must it be called?"*

```go
func safeCall(f func()) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic recovered: %v", r)
        }
    }()
    
    f()
    return nil
}

// Usage
err := safeCall(func() {
    panic("oops!")
})
fmt.Println(err)  // "panic recovered: oops!"
```

**Rules:**
- `recover()` must be called directly in deferred function
- Returns `nil` if no panic
- Returns panic value if panicking
- Calling `recover()` stops the panic

### HTTP Handler Recovery

```go
func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                // Log the stack trace
                stack := debug.Stack()
                log.Printf("Panic: %v\n%s", err, stack)
                
                // Return 500 to client
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()
        
        next.ServeHTTP(w, r)
    })
}
```

### Never Panic Across API Boundaries

**Interview Question:** *"Why should library code never panic?"*

```go
// BAD: Library panics
func (c *Client) Get(url string) []byte {
    resp, err := http.Get(url)
    if err != nil {
        panic(err)  // Caller cannot handle!
    }
    // ...
}

// GOOD: Library returns error
func (c *Client) Get(url string) ([]byte, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("http get: %w", err)
    }
    // ...
}
```

### Example: Panic/Recover Patterns

```go
package main

import (
    "fmt"
    "runtime/debug"
)

// Must pattern - panic on error
func MustCompileRegex(pattern string) *regexp.Regexp {
    re, err := regexp.Compile(pattern)
    if err != nil {
        panic(fmt.Sprintf("invalid regex %q: %v", pattern, err))
    }
    return re
}

// Safe goroutine wrapper
func safeGo(f func()) {
    go func() {
        defer func() {
            if r := recover(); r != nil {
                fmt.Printf("Goroutine panic: %v\n%s", r, debug.Stack())
            }
        }()
        f()
    }()
}

// Convert panic to error
func tryParse(s string) (result int, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("parse failed: %v", r)
        }
    }()
    
    // Some parsing that might panic
    result = riskyParse(s)
    return result, nil
}

func main() {
    // Must pattern - use for constants/init
    var emailRegex = MustCompileRegex(`^[a-z]+@[a-z]+\.[a-z]+$`)
    
    // Safe goroutine
    safeGo(func() {
        panic("goroutine panic!")
    })
    
    // Continue running...
    fmt.Println("Main continues")
}
```

---

## 3.6 Error Handling in Practice

### Don't Ignore Errors

**Interview Question:** *"Is it ever okay to ignore errors in Go?"*

```go
// BAD: Ignoring error
data, _ := json.Marshal(obj)  // Could fail!

// Acceptable only when:
// 1. You're certain it can't fail
// 2. Failure doesn't matter
// 3. You've documented why

// Example of acceptable ignore:
fmt.Fprintf(w, "hello")  // Writing to stdout rarely fails
_ = conn.Close()          // Best-effort cleanup
```

### Don't Over-Wrap

**Interview Question:** *"What's wrong with wrapping errors at every level?"*

```go
// BAD: Every level adds redundant context
// Error: "handler: service: repository: database: connection refused"

// GOOD: Add context only when it helps
// Error: "creating user: checking uniqueness: connection refused"

// Rule: Add context at boundaries or when it clarifies
```

### Error Logging Strategy

**Interview Question:** *"Where should you log errors?"*

```go
// BAD: Log at every level
func (r *Repo) FindUser(id string) (*User, error) {
    user, err := r.db.Find(id)
    if err != nil {
        log.Printf("FindUser failed: %v", err)  // Logged here
        return nil, err
    }
    return user, nil
}

func (s *Service) GetUser(id string) (*User, error) {
    user, err := s.repo.FindUser(id)
    if err != nil {
        log.Printf("GetUser failed: %v", err)  // And here
        return nil, err
    }
    return user, nil
}

// GOOD: Log once at the top (handler)
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    user, err := h.service.GetUser(id)
    if err != nil {
        log.Printf("GetUser error: %v", err)  // Log once
        http.Error(w, "Internal Error", 500)
        return
    }
    // ...
}
```

### Using `errgroup` for Concurrent Errors

**Interview Question:** *"How do you handle errors from multiple goroutines?"*

```go
import "golang.org/x/sync/errgroup"

func fetchAll(urls []string) ([][]byte, error) {
    g, ctx := errgroup.WithContext(context.Background())
    results := make([][]byte, len(urls))
    
    for i, url := range urls {
        i, url := i, url  // Capture for goroutine
        g.Go(func() error {
            data, err := fetchURL(ctx, url)
            if err != nil {
                return fmt.Errorf("fetching %s: %w", url, err)
            }
            results[i] = data
            return nil
        })
    }
    
    if err := g.Wait(); err != nil {
        return nil, err  // First error
    }
    
    return results, nil
}
```

### Error Message Conventions

**Interview Question:** *"What are the conventions for error messages in Go?"*

```go
// Convention: lowercase, no punctuation, no "error:" prefix

// GOOD
errors.New("connection refused")
fmt.Errorf("invalid port: %d", port)

// BAD
errors.New("Connection Refused.")
errors.New("Error: connection refused")
fmt.Errorf("Invalid port: %d.", port)

// Wrapped errors read naturally
// "opening config: reading file: permission denied"
```

---

## Interview Questions

### Beginner Level

1. **Q:** What does the `error` interface require?
   **A:** A single method: `Error() string`

2. **Q:** What's the difference between `errors.New()` and `fmt.Errorf()`?
   **A:** `errors.New()` creates simple static errors. `fmt.Errorf()` allows formatted strings and error wrapping with `%w`.

3. **Q:** How do you check if an error equals a specific value?
   **A:** Use `errors.Is(err, target)` to check the entire error chain.

### Intermediate Level

4. **Q:** What's the difference between `%w` and `%v` in `fmt.Errorf()`?
   **A:** `%w` wraps the error preserving the chain for `errors.Is/As`. `%v` converts to string, breaking the chain.

5. **Q:** When should you use `panic()` vs. returning an error?
   **A:** Panic for unrecoverable programming bugs. Return errors for expected failures (network, user input, etc.).

6. **Q:** How do you extract a specific error type from a wrapped chain?
   **A:** `errors.As(err, &target)` - it searches the chain and populates target if found.

### Advanced Level

7. **Q:** How would you implement a custom error type that supports `errors.Is()` with special matching logic?
   **A:** Implement `Is(target error) bool` method on your error type.

8. **Q:** What happens if `recover()` is called outside a deferred function?
   **A:** It returns `nil` and has no effect. `recover()` only works directly inside a deferred function.

9. **Q:** Design an error handling strategy for a REST API.
   **A:** 
   - Define domain errors (sentinel and custom types)
   - Transform DB/external errors to domain errors at repository layer
   - Wrap with context at service layer
   - Map to HTTP status codes at handler layer
   - Log once at handler with full error chain
   - Return structured error response to client

---

## Summary

| Topic | Key Points |
|-------|------------|
| Error Interface | `type error interface { Error() string }` |
| Creation | `errors.New()` for static, `fmt.Errorf()` for formatted |
| Wrapping | `%w` preserves chain, `%v` breaks it |
| Inspection | `errors.Is()` for value, `errors.As()` for type |
| Panic | Only for unrecoverable bugs, never in libraries |
| Patterns | Early return, wrap at boundaries, log once |

**Next Phase:** [Phase 4 — Memory Management & Performance](../Phase_4/Phase_4.md)

