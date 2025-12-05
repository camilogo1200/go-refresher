# 🎓 Phase 12: Modern Go Features (1.22-1.24+)

[← Back to Main Roadmap](../README.md) | [← Previous: Phase 11](../Phase_11/Phase_11.md)

---

**Objective:** Stay current with the latest language evolution and upcoming features.

**Reference:** [Go Release Notes](https://go.dev/doc/devel/release), [Go Proposals](https://github.com/golang/go/issues?q=is%3Aissue+label%3AProposal)

**Prerequisites:** Phase 0-11

**Estimated Duration:** Ongoing (Updates with each Go release)

---

## 📋 Table of Contents

1. [Iterators (Go 1.23+)](#121-iterators-go-123)
2. [Enhanced HTTP Routing (Go 1.22)](#122-enhanced-http-routing-go-122)
3. [Loop Variable Fix (Go 1.22)](#123-loop-variable-fix-go-122)
4. [Range Over Integers (Go 1.22)](#124-range-over-integers-go-122)
5. [GOMEMLIMIT and GC Improvements](#125-gomemlimit-and-gc-improvements)
6. [Structured Logging (Go 1.21)](#126-structured-logging-go-121)
7. [Recent Additions Summary](#127-recent-additions-summary)
8. [Upcoming Features](#128-upcoming-features)
9. [Migration Strategies](#129-migration-strategies)
10. [Interview Questions](#interview-questions)

---

## 12.1 Iterators (Go 1.23+)

### Iterator Functions

**Interview Question:** *"What are iterators in Go 1.23+ and how do they work?"*

```go
// Iterator types defined in iter package
package iter

// Single value iterator
type Seq[V any] func(yield func(V) bool)

// Key-value iterator
type Seq2[K, V any] func(yield func(K, V) bool)
```

### Creating Iterators

```go
// Basic iterator - yields values until stopped
func countTo(n int) iter.Seq[int] {
    return func(yield func(int) bool) {
        for i := 1; i <= n; i++ {
            if !yield(i) {
                return  // Consumer stopped
            }
        }
    }
}

// Usage with range
for v := range countTo(5) {
    fmt.Println(v)  // 1, 2, 3, 4, 5
}

// Early termination
for v := range countTo(100) {
    if v > 3 {
        break  // yield returns false, iterator stops
    }
    fmt.Println(v)  // 1, 2, 3
}
```

### Two-Value Iterators

```go
// Iterator yielding key-value pairs
func enumerate[T any](s []T) iter.Seq2[int, T] {
    return func(yield func(int, T) bool) {
        for i, v := range s {
            if !yield(i, v) {
                return
            }
        }
    }
}

// Usage
for i, v := range enumerate([]string{"a", "b", "c"}) {
    fmt.Printf("%d: %s\n", i, v)
}
```

### Standard Library Iterators

```go
import (
    "maps"
    "slices"
)

// Slice iteration
s := []int{1, 2, 3, 4, 5}

for i, v := range slices.All(s) {
    fmt.Println(i, v)
}

for v := range slices.Values(s) {
    fmt.Println(v)
}

for v := range slices.Backward(s) {
    fmt.Println(v)  // 5, 4, 3, 2, 1
}

// Map iteration
m := map[string]int{"a": 1, "b": 2}

for k := range maps.Keys(m) {
    fmt.Println(k)
}

for v := range maps.Values(m) {
    fmt.Println(v)
}

for k, v := range maps.All(m) {
    fmt.Println(k, v)
}
```

### Pull Iterators

```go
// Convert push iterator to pull (for imperative consumption)
import "iter"

func main() {
    seq := countTo(5)
    
    // Pull iterator - manual control
    next, stop := iter.Pull(seq)
    defer stop()  // Important: release resources
    
    for {
        v, ok := next()
        if !ok {
            break
        }
        fmt.Println(v)
    }
}
```

### Custom Collection Iterators

```go
type Set[T comparable] struct {
    items map[T]struct{}
}

func (s *Set[T]) Add(item T) {
    s.items[item] = struct{}{}
}

// Implement iterator
func (s *Set[T]) All() iter.Seq[T] {
    return func(yield func(T) bool) {
        for item := range s.items {
            if !yield(item) {
                return
            }
        }
    }
}

// Usage
set := &Set[int]{items: make(map[int]struct{})}
set.Add(1)
set.Add(2)
set.Add(3)

for v := range set.All() {
    fmt.Println(v)
}
```

### Iterator Composition

```go
// Filter iterator
func filter[T any](seq iter.Seq[T], pred func(T) bool) iter.Seq[T] {
    return func(yield func(T) bool) {
        for v := range seq {
            if pred(v) {
                if !yield(v) {
                    return
                }
            }
        }
    }
}

// Map iterator
func mapIter[T, U any](seq iter.Seq[T], fn func(T) U) iter.Seq[U] {
    return func(yield func(U) bool) {
        for v := range seq {
            if !yield(fn(v)) {
                return
            }
        }
    }
}

// Chain iterators
nums := countTo(10)
evens := filter(nums, func(n int) bool { return n%2 == 0 })
doubled := mapIter(evens, func(n int) int { return n * 2 })

for v := range doubled {
    fmt.Println(v)  // 4, 8, 12, 16, 20
}
```

---

## 12.2 Enhanced HTTP Routing (Go 1.22)

### Method Patterns

**Interview Question:** *"What routing improvements came in Go 1.22?"*

```go
mux := http.NewServeMux()

// Before Go 1.22: Manual method checking
mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        // handle GET
    case "POST":
        // handle POST
    default:
        http.Error(w, "Method not allowed", 405)
    }
})

// Go 1.22+: Method in pattern
mux.HandleFunc("GET /users", listUsers)
mux.HandleFunc("POST /users", createUser)
mux.HandleFunc("GET /users/{id}", getUser)
mux.HandleFunc("PUT /users/{id}", updateUser)
mux.HandleFunc("DELETE /users/{id}", deleteUser)
```

### Path Parameters

```go
mux := http.NewServeMux()

// Path parameters with {name}
mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")  // New in 1.22!
    fmt.Fprintf(w, "User ID: %s", id)
})

// Multiple parameters
mux.HandleFunc("GET /users/{userID}/posts/{postID}", func(w http.ResponseWriter, r *http.Request) {
    userID := r.PathValue("userID")
    postID := r.PathValue("postID")
    fmt.Fprintf(w, "User: %s, Post: %s", userID, postID)
})
```

### Wildcard Matching

```go
// {name...} matches remaining path segments
mux.HandleFunc("GET /files/{path...}", func(w http.ResponseWriter, r *http.Request) {
    path := r.PathValue("path")
    // /files/images/logo.png -> path = "images/logo.png"
    fmt.Fprintf(w, "File path: %s", path)
})
```

### Precedence Rules

```go
mux := http.NewServeMux()

// More specific patterns take precedence
mux.HandleFunc("GET /users/{id}", getUser)       // Less specific
mux.HandleFunc("GET /users/me", getCurrentUser)  // More specific - wins!

// Pattern specificity:
// 1. Longer patterns
// 2. More literal segments
// 3. Method-specific over any-method

// These are ordered by precedence (highest first):
// GET /users/me
// GET /users/{id}
// /users/{id}
// GET /users/
// /users/
```

### Host-Based Routing

```go
mux := http.NewServeMux()

// Route by host
mux.HandleFunc("GET api.example.com/users", apiUsers)
mux.HandleFunc("GET www.example.com/users", webUsers)
```

### Eliminating Third-Party Routers

```go
// Before Go 1.22: Often needed chi, gorilla/mux, etc.
// Now: ServeMux covers most use cases!

// Still consider third-party for:
// - Middleware ecosystem
// - Complex matching (regex)
// - OpenAPI integration
// - Already using (don't migrate for its own sake)
```

---

## 12.3 Loop Variable Fix (Go 1.22)

### The Historical Problem

**Interview Question:** *"What was the loop variable capture bug and how was it fixed?"*

```go
// Before Go 1.22: DANGEROUS!
func main() {
    funcs := make([]func(), 3)
    
    for i := 0; i < 3; i++ {
        funcs[i] = func() {
            fmt.Println(i)  // Captures reference to i
        }
    }
    
    for _, f := range funcs {
        f()  // Prints: 3, 3, 3 (not 0, 1, 2!)
    }
}

// The problem: All closures share same variable
// By the time they execute, i has final value (3)
```

### Old Workarounds

```go
// Workaround 1: Shadow variable
for i := 0; i < 3; i++ {
    i := i  // Shadow with new variable
    funcs[i] = func() {
        fmt.Println(i)
    }
}

// Workaround 2: Pass as argument
for i := 0; i < 3; i++ {
    funcs[i] = func(n int) func() {
        return func() {
            fmt.Println(n)
        }
    }(i)
}
```

### Go 1.22 Fix

```go
// Go 1.22+: Loop variables are per-iteration
func main() {
    funcs := make([]func(), 3)
    
    for i := 0; i < 3; i++ {
        funcs[i] = func() {
            fmt.Println(i)  // Now captures copy!
        }
    }
    
    for _, f := range funcs {
        f()  // Prints: 0, 1, 2 (correct!)
    }
}
```

### Range Loops Too

```go
// Also fixed for range loops
items := []string{"a", "b", "c"}

for _, item := range items {
    go func() {
        fmt.Println(item)  // Go 1.22+: Prints each item correctly
    }()
}
```

### Compatibility

```go
// Go 1.22+ applies fix based on go.mod version

// go.mod with go 1.22+: New behavior (per-iteration)
// go.mod with go 1.21 or lower: Old behavior (shared)

// Explicit control (testing):
// GOEXPERIMENT=loopvar go run main.go
```

---

## 12.4 Range Over Integers (Go 1.22)

### Integer Range

**Interview Question:** *"How does range over integers work in Go 1.22+?"*

```go
// Go 1.22+: Range over integers!
for i := range 5 {
    fmt.Println(i)  // 0, 1, 2, 3, 4
}

// Equivalent to:
for i := 0; i < 5; i++ {
    fmt.Println(i)
}

// Use cases:
// Simpler iteration when you don't need fine control
// Cleaner than range over dummy slice
```

### Comparison with Traditional For

```go
// Traditional - more control
for i := 0; i < n; i++ {
    // Full control: start, condition, increment
}

// Range over int - simpler
for i := range n {
    // Always starts at 0, increments by 1
}

// Can't do with range:
for i := 10; i >= 0; i-- { }  // Countdown
for i := 0; i < 100; i += 10 { }  // Step by 10
```

---

## 12.5 GOMEMLIMIT and GC Improvements

### GOMEMLIMIT (Go 1.19+)

**Interview Question:** *"What is GOMEMLIMIT and when should you use it?"*

```bash
# Set soft memory limit
GOMEMLIMIT=1GiB ./myapp

# Supported units: B, KiB, MiB, GiB, TiB
# Also: KB, MB, GB, TB (powers of 1000)
```

```go
// Programmatic setting
import "runtime/debug"

debug.SetMemoryLimit(1 << 30)  // 1 GiB
```

### How GOMEMLIMIT Works

```
Without GOMEMLIMIT:
- GC based only on GOGC
- Can OOM in memory-constrained environments

With GOMEMLIMIT:
- Soft limit on total Go memory
- GC runs more aggressively near limit
- Helps prevent OOM
- Not a hard limit (can exceed temporarily)
```

### GOGC + GOMEMLIMIT Together

```go
// GOGC controls GC frequency (% growth trigger)
// GOMEMLIMIT provides safety net

// Example for container with 1GB limit:
// GOGC=100        - Normal GC behavior
// GOMEMLIMIT=800MiB - Leave 200MB for non-Go memory

// Near limit: GC overrides GOGC to avoid OOM
```

### GC Improvements Timeline

```
Go 1.19: GOMEMLIMIT introduced
Go 1.20: Improved GC pacing
Go 1.21: Better memory limit behavior
Go 1.22: Reduced tail latency
Go 1.23: Improved arena support (experimental)
```

---

## 12.6 Structured Logging (Go 1.21)

### log/slog Package

**Interview Question:** *"What are the benefits of log/slog over log package?"*

```go
import "log/slog"

// Structured logging with typed attributes
slog.Info("User logged in",
    slog.String("user_id", "123"),
    slog.Int("attempt", 1),
    slog.Duration("latency", 42*time.Millisecond),
)

// Output (JSON handler):
// {"time":"2024-01-15T10:30:00Z","level":"INFO","msg":"User logged in","user_id":"123","attempt":1,"latency":"42ms"}
```

### Handlers

```go
// Text handler (development)
logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

// JSON handler (production)
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
    AddSource: true,
}))

// Set as default
slog.SetDefault(logger)
```

### Log Levels

```go
slog.Debug("Debug message")
slog.Info("Info message")
slog.Warn("Warning message")
slog.Error("Error message")

// With context
slog.InfoContext(ctx, "Request processed",
    slog.String("request_id", requestID),
)
```

### Logger Groups and Attributes

```go
// Add common attributes to logger
logger := slog.Default().With(
    slog.String("service", "user-api"),
    slog.String("version", "1.0.0"),
)

// Groups for nested structure
slog.Info("Request",
    slog.Group("user",
        slog.String("id", "123"),
        slog.String("role", "admin"),
    ),
)
// {"msg":"Request","user":{"id":"123","role":"admin"}}
```

---

## 12.7 Recent Additions Summary

### Go 1.21 (August 2023)

```go
// 1. min/max builtins
minimum := min(1, 2, 3)  // 1
maximum := max(1, 2, 3)  // 3

// 2. clear builtin
m := map[string]int{"a": 1, "b": 2}
clear(m)  // Empty map, keeps capacity

s := []int{1, 2, 3}
clear(s)  // Zero all elements

// 3. log/slog package (see above)

// 4. slices/maps packages (standard library)
import "slices"
import "maps"

slices.Sort(s)
slices.Contains(s, 2)
slices.Clone(s)

maps.Clone(m)
maps.Equal(m1, m2)
```

### Go 1.22 (February 2024)

```go
// 1. Enhanced routing (see above)
mux.HandleFunc("GET /users/{id}", handler)

// 2. Loop variable fix (see above)

// 3. Range over integers
for i := range 10 { }

// 4. math/rand/v2 package
import "math/rand/v2"
n := rand.IntN(100)  // [0, 100)
```

### Go 1.23 (August 2024)

```go
// 1. Iterators (see above)
for v := range myIterator { }

// 2. Timer/Ticker changes
timer := time.NewTimer(time.Second)
// Now uses monotonic clock
// Stop returns whether timer was active

// 3. unique package
import "unique"
handle := unique.Make("string")
// Interning for memory efficiency
```

### Go 1.24 (Expected February 2025)

```go
// Expected features (subject to change):
// - Generic type aliases
// - Improved error handling proposals
// - Performance improvements
// - More iterator support
```

---

## 12.8 Upcoming Features

### Generic Methods (Proposal)

**Interview Question:** *"Why can't Go methods have type parameters?"*

```go
// Currently NOT allowed:
type Container struct {
    items []any
}

func (c *Container) Get[T any](index int) T {  // COMPILE ERROR
    return c.items[index].(T)
}

// Workaround: Top-level functions
func Get[T any](c *Container, index int) T {
    return c.items[index].(T)
}

// Proposal status: Under discussion
// Challenges: Method set determination, interface satisfaction
```

### Sum Types (Proposal)

```go
// Currently: Use interfaces or code generation
type Result interface {
    isResult()
}

type Ok struct { Value int }
type Err struct { Error error }

func (Ok) isResult() {}
func (Err) isResult() {}

// Proposed (if accepted):
type Result = Ok | Err

// Benefits:
// - Exhaustive switch checking
// - Memory efficiency
// - Clearer intent
```

### Error Handling (`?` Operator Proposals)

```go
// Current
func process() error {
    data, err := fetchData()
    if err != nil {
        return err
    }
    
    result, err := transform(data)
    if err != nil {
        return err
    }
    
    return save(result)
}

// Various proposals (not accepted):
func process() error {
    data := fetchData()?  // Return error if non-nil
    result := transform(data)?
    return save(result)
}

// Status: Multiple proposals rejected
// Go team prefers explicit error handling
```

### Weak Pointers (Go 1.24?)

```go
// Proposal: Pointers that don't prevent GC
import "weak"

type Cache struct {
    items map[string]weak.Pointer[Value]
}

func (c *Cache) Get(key string) (*Value, bool) {
    if wp, ok := c.items[key]; ok {
        if v := wp.Value(); v != nil {
            return v, true
        }
        // Value was collected
        delete(c.items, key)
    }
    return nil, false
}
```

---

## 12.9 Migration Strategies

### Updating go.mod

```bash
# Update Go version
go mod edit -go=1.22

# Verify compatibility
go vet ./...
go test ./...
```

### Gradual Adoption

```go
// Use build tags for version-specific code
//go:build go1.22

package main

// Go 1.22+ specific code
func handler(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")  // Only in 1.22+
}
```

### Testing Across Versions

```yaml
# GitHub Actions
jobs:
  test:
    strategy:
      matrix:
        go: ['1.21', '1.22', '1.23']
    steps:
      - uses: actions/setup-go@v4
        with:
          go-version: ${{ matrix.go }}
      - run: go test ./...
```

### Breaking Changes Checklist

```markdown
Before upgrading Go version:
1. [ ] Read release notes
2. [ ] Check for breaking changes
3. [ ] Update dependencies
4. [ ] Run full test suite
5. [ ] Test in staging environment
6. [ ] Monitor after production deploy

Common issues:
- [ ] Loop variable behavior change (1.22)
- [ ] Routing precedence changes (1.22)
- [ ] Timer/Ticker behavior (1.23)
```

---

## Interview Questions

### Beginner Level

1. **Q:** What did Go 1.22 change about loop variables?
   **A:** Loop variables are now per-iteration, not shared. Closures capture a copy, fixing the common goroutine bug.

2. **Q:** How do you add path parameters in Go 1.22+ routing?
   **A:** Use `{name}` in pattern: `"GET /users/{id}"`. Access with `r.PathValue("id")`.

3. **Q:** What is `range 10` in Go 1.22+?
   **A:** Range over integer - iterates 0 to 9. Equivalent to `for i := 0; i < 10; i++`.

### Intermediate Level

4. **Q:** Explain GOMEMLIMIT and when to use it.
   **A:** Soft memory limit for GC. Use in containers to prevent OOM. GC runs more aggressively near limit.

5. **Q:** What are the iterator types in Go 1.23+?
   **A:** `iter.Seq[V]` for single values, `iter.Seq2[K,V]` for key-value pairs. Both use yield function pattern.

6. **Q:** Why use log/slog over the log package?
   **A:** Structured logging (key-value), multiple handlers (JSON/text), log levels, context support, better performance.

### Advanced Level

7. **Q:** How do iterators interact with early termination?
   **A:** When `yield` returns false, iterator should stop. Pull iterators need explicit `stop()` call.

8. **Q:** Why doesn't Go have generic methods?
   **A:** Method sets must be determinable at compile time for interface satisfaction. Generic methods make this undecidable.

9. **Q:** How should you handle migration to new loop variable semantics?
   **A:** Go version in go.mod determines behavior. Test thoroughly, especially closure-heavy code. Old workarounds still work.

---

## Summary

| Feature | Version | Key Points |
|---------|---------|------------|
| Iterators | 1.23 | `iter.Seq`, `iter.Seq2`, yield pattern, range compatible |
| HTTP Routing | 1.22 | Method patterns, path parameters, wildcards |
| Loop Variables | 1.22 | Per-iteration capture, fixes closure bug |
| Range Int | 1.22 | `range n` iterates 0 to n-1 |
| GOMEMLIMIT | 1.19 | Soft memory limit for GC |
| log/slog | 1.21 | Structured logging, JSON/text handlers |
| min/max/clear | 1.21 | Built-in functions |

### Staying Current

```bash
# Follow Go development
https://go.dev/blog/
https://github.com/golang/go/issues
https://go.dev/doc/devel/release

# Track proposals
https://github.com/golang/go/labels/Proposal

# Release schedule
# February: Major release (1.N)
# August: Major release (1.N+1)
# Monthly: Patch releases
```

---

## 🎉 Congratulations!

You've completed the Go Mastery Roadmap. You now have comprehensive knowledge from fundamentals to runtime internals, positioning you for:

- **Technical Interviews:** Deep understanding of Go's design and implementation
- **System Design:** Cloud-native architecture with Go
- **Performance Optimization:** Memory, GC, and scheduler tuning
- **Production Systems:** Observability, resilience, and operations

**Continue Learning:**
- Contribute to open source Go projects
- Read the Go standard library source code
- Follow Go proposals and participate in discussions
- Build production systems and learn from experience

*"Simplicity is the ultimate sophistication." — Leonardo da Vinci*

*"Clear is better than clever." — The Go Proverbs*

