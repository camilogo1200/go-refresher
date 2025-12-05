# 🧪 Phase 6: Testing & Engineering Reliability

[← Back to Main Roadmap](../README.md) | [← Previous: Phase 5](../Phase_5/Phase_5.md)

---

**Objective:** Master Go's testing ecosystem from unit tests to production verification.

**Reference:** [Go Testing Package](https://pkg.go.dev/testing)

**Prerequisites:** Phase 0-5

**Estimated Duration:** 2-3 weeks

---

## 📋 Table of Contents

1. [The `testing` Package](#61-the-testing-package)
2. [Running Tests](#62-running-tests)
3. [Table-Driven Tests](#63-table-driven-tests)
4. [Subtests and Sub-benchmarks](#64-subtests-and-sub-benchmarks)
5. [TestMain](#65-testmain)
6. [Benchmarking](#66-benchmarking)
7. [Fuzzing](#67-fuzzing-go-118)
8. [Code Coverage](#68-code-coverage)
9. [Mocking Strategies](#69-mocking-strategies)
10. [Integration Testing](#610-integration-testing)
11. [Example Functions](#611-example-functions)
12. [Interview Questions](#interview-questions)

---

## 6.1 The `testing` Package

### Test Function Signature

**Interview Question:** *"What are the naming conventions for test files and functions in Go?"*

```go
// File: math_test.go (must end with _test.go)
package math

import "testing"

// Function: must start with Test, followed by uppercase letter
func TestAdd(t *testing.T) {
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("Add(2, 3) = %d; want 5", result)
    }
}
```

**Naming Rules:**
- Test files: `*_test.go`
- Test functions: `TestXxx(t *testing.T)`
- Benchmark functions: `BenchmarkXxx(b *testing.B)`
- Fuzz functions: `FuzzXxx(f *testing.F)`
- Example functions: `ExampleXxx()`

### Testing Methods

```go
func TestExample(t *testing.T) {
    // Report failure, continue execution
    t.Error("failure message")
    t.Errorf("formatted %s message", "error")
    
    // Report failure, stop this test
    t.Fatal("fatal message")
    t.Fatalf("formatted %s message", "fatal")
    
    // Log (only shown on failure or -v)
    t.Log("info message")
    t.Logf("formatted %s message", "info")
    
    // Skip test
    t.Skip("skipping because...")
    t.Skipf("skipping: %s", reason)
    
    // Mark test for parallel execution
    t.Parallel()
    
    // Mark function as helper (better stack traces)
    t.Helper()
    
    // Cleanup function (runs after test)
    t.Cleanup(func() {
        // Cleanup resources
    })
}
```

### Test Helper Functions

**Interview Question:** *"What is `t.Helper()` and why should you use it?"*

```go
// Helper function - use t.Helper()
func assertEqual(t *testing.T, got, want int) {
    t.Helper()  // Marks this as helper - errors show caller's line
    if got != want {
        t.Errorf("got %d, want %d", got, want)
    }
}

func TestMath(t *testing.T) {
    result := Add(2, 3)
    assertEqual(t, result, 5)  // Error points HERE, not inside assertEqual
}
```

### Example: Complete Test File

```go
// calculator_test.go
package calculator

import (
    "testing"
)

func TestAdd(t *testing.T) {
    got := Add(2, 3)
    want := 5
    
    if got != want {
        t.Errorf("Add(2, 3) = %d; want %d", got, want)
    }
}

func TestDivide(t *testing.T) {
    // Test normal case
    got, err := Divide(10, 2)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if got != 5 {
        t.Errorf("Divide(10, 2) = %d; want 5", got)
    }
    
    // Test error case
    _, err = Divide(10, 0)
    if err == nil {
        t.Error("Divide(10, 0) should return error")
    }
}
```

---

## 6.2 Running Tests

### Basic Commands

```bash
# Run tests in current package
go test

# Run all tests in module
go test ./...

# Verbose output
go test -v ./...

# Run specific test
go test -run TestAdd ./...

# Run tests matching pattern
go test -run "TestUser.*" ./...

# Run specific subtest
go test -run "TestAdd/positive" ./...

# Multiple runs (catch flaky tests)
go test -count=10 ./...

# With race detector
go test -race ./...

# Set timeout
go test -timeout 30s ./...

# Short mode (skip long tests)
go test -short ./...
```

### Filtering Tests

```go
func TestLongRunning(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping long test in short mode")
    }
    // Long running test...
}
```

### Test Caching

```bash
# Go caches passing tests
# Force re-run with -count=1
go test -count=1 ./...

# Or clean cache
go clean -testcache
```

### Example: Test Execution

```bash
# Run with verbose and race detector
$ go test -v -race ./...

=== RUN   TestAdd
--- PASS: TestAdd (0.00s)
=== RUN   TestDivide
--- PASS: TestDivide (0.00s)
PASS
ok      myapp/calculator    0.005s
```

---

## 6.3 Table-Driven Tests

### The Idiomatic Pattern

**Interview Question:** *"What are table-driven tests and why are they idiomatic in Go?"*

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive numbers", 2, 3, 5},
        {"negative numbers", -2, -3, -5},
        {"mixed", -2, 3, 1},
        {"zeros", 0, 0, 0},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.expected {
                t.Errorf("Add(%d, %d) = %d; want %d",
                    tt.a, tt.b, got, tt.expected)
            }
        })
    }
}
```

**Benefits:**
- Easy to add new test cases
- Clear test case documentation
- Parallel execution per case
- Better failure messages

### Anonymous Struct Pattern

```go
func TestParse(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    int
        wantErr bool
    }{
        {
            name:  "valid number",
            input: "42",
            want:  42,
        },
        {
            name:    "invalid",
            input:   "abc",
            wantErr: true,
        },
        {
            name:  "negative",
            input: "-10",
            want:  -10,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Parse(tt.input)
            
            if tt.wantErr {
                if err == nil {
                    t.Error("expected error, got nil")
                }
                return
            }
            
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            
            if got != tt.want {
                t.Errorf("Parse(%q) = %d; want %d",
                    tt.input, got, tt.want)
            }
        })
    }
}
```

### Map-Based Tests

```go
func TestHTTPStatus(t *testing.T) {
    tests := map[string]struct {
        code int
        want string
    }{
        "OK":         {200, "OK"},
        "Not Found":  {404, "Not Found"},
        "Error":      {500, "Internal Server Error"},
    }
    
    for name, tt := range tests {
        t.Run(name, func(t *testing.T) {
            got := StatusText(tt.code)
            if got != tt.want {
                t.Errorf("StatusText(%d) = %q; want %q",
                    tt.code, got, tt.want)
            }
        })
    }
}
```

---

## 6.4 Subtests and Sub-benchmarks

### Subtests with t.Run

**Interview Question:** *"How do subtests help with test organization and parallel execution?"*

```go
func TestUser(t *testing.T) {
    // Setup shared by all subtests
    db := setupTestDB(t)
    
    t.Run("Create", func(t *testing.T) {
        t.Parallel()  // Can run in parallel with other subtests
        // Test creation...
    })
    
    t.Run("Update", func(t *testing.T) {
        t.Parallel()
        // Test update...
    })
    
    t.Run("Delete", func(t *testing.T) {
        t.Parallel()
        // Test deletion...
    })
}
```

### Filtering Subtests

```bash
# Run specific subtest
go test -run "TestUser/Create"

# Run multiple subtests with pattern
go test -run "TestUser/(Create|Update)"
```

### Parallel Subtests Pattern

```go
func TestParallel(t *testing.T) {
    tests := []struct {
        name  string
        input int
    }{
        {"case1", 1},
        {"case2", 2},
        {"case3", 3},
    }
    
    for _, tt := range tests {
        tt := tt  // Capture for parallel (Go < 1.22)
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()  // Run in parallel
            // Test using tt.input
        })
    }
}
```

### Setup and Teardown per Subtest

```go
func TestWithSetup(t *testing.T) {
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            resource := acquireResource()
            
            // Teardown (guaranteed to run)
            t.Cleanup(func() {
                resource.Close()
            })
            
            // Test
            result := useResource(resource)
            // assertions...
        })
    }
}
```

---

## 6.5 TestMain

### Package-Level Setup/Teardown

**Interview Question:** *"When would you use TestMain?"*

```go
func TestMain(m *testing.M) {
    // Setup before all tests
    db, err := setupDatabase()
    if err != nil {
        fmt.Println("Setup failed:", err)
        os.Exit(1)
    }
    
    // Run tests
    code := m.Run()
    
    // Teardown after all tests
    db.Close()
    cleanupTestData()
    
    // Exit with test result code
    os.Exit(code)
}
```

### Use Cases

```go
// 1. Database connection pool
var testDB *sql.DB

func TestMain(m *testing.M) {
    var err error
    testDB, err = sql.Open("postgres", os.Getenv("TEST_DB_URL"))
    if err != nil {
        log.Fatal(err)
    }
    defer testDB.Close()
    
    os.Exit(m.Run())
}

// 2. Docker container setup
func TestMain(m *testing.M) {
    ctx := context.Background()
    container, err := testcontainers.StartContainer(ctx, redisConfig)
    if err != nil {
        log.Fatal(err)
    }
    defer container.Terminate(ctx)
    
    os.Exit(m.Run())
}

// 3. Global configuration
func TestMain(m *testing.M) {
    // Set test environment
    os.Setenv("APP_ENV", "test")
    
    // Parse test flags
    flag.Parse()
    
    os.Exit(m.Run())
}
```

---

## 6.6 Benchmarking

### Benchmark Functions

**Interview Question:** *"How do you write and interpret benchmarks in Go?"*

```go
func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Add(2, 3)
    }
}
```

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkAdd ./...

# With memory allocation stats
go test -bench=. -benchmem ./...

# Custom duration
go test -bench=. -benchtime=5s ./...

# Multiple runs for statistics
go test -bench=. -count=5 ./...
```

### Benchmark Output

```
BenchmarkAdd-8      1000000000        0.5 ns/op        0 B/op        0 allocs/op
│             │              │             │             │
│             │              │             │             └─ allocations per op
│             │              │             └─ bytes allocated per op
│             │              └─ time per operation
│             └─ number of iterations
└─ number of CPUs
```

### Benchmark Best Practices

```go
func BenchmarkStringConcat(b *testing.B) {
    // Setup (not timed)
    data := generateTestData()
    
    b.ResetTimer()  // Start timing from here
    
    for i := 0; i < b.N; i++ {
        _ = processData(data)
    }
}

func BenchmarkWithAllocs(b *testing.B) {
    b.ReportAllocs()  // Report allocations
    
    for i := 0; i < b.N; i++ {
        _ = make([]byte, 1024)
    }
}

func BenchmarkSetBytes(b *testing.B) {
    data := make([]byte, 1024)
    b.SetBytes(int64(len(data)))  // Report throughput
    
    for i := 0; i < b.N; i++ {
        processBytes(data)
    }
}
```

### Parallel Benchmarks

```go
func BenchmarkParallel(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        // Each goroutine runs this
        for pb.Next() {
            doWork()
        }
    })
}
```

### Comparing Benchmarks

```bash
# Save baseline
go test -bench=. -count=10 > old.txt

# Make changes, save new results
go test -bench=. -count=10 > new.txt

# Compare with benchstat
go install golang.org/x/perf/cmd/benchstat@latest
benchstat old.txt new.txt
```

### Example: Complete Benchmark

```go
func BenchmarkStringBuilder(b *testing.B) {
    for i := 0; i < b.N; i++ {
        var builder strings.Builder
        for j := 0; j < 100; j++ {
            builder.WriteString("hello")
        }
        _ = builder.String()
    }
}

func BenchmarkStringConcat(b *testing.B) {
    for i := 0; i < b.N; i++ {
        var s string
        for j := 0; j < 100; j++ {
            s += "hello"
        }
        _ = s
    }
}

// Results:
// BenchmarkStringBuilder-8   500000   2400 ns/op    1024 B/op   4 allocs/op
// BenchmarkStringConcat-8     10000  140000 ns/op  53000 B/op  99 allocs/op
```

---

## 6.7 Fuzzing (Go 1.18+)

### Fuzz Function Signature

**Interview Question:** *"What is fuzzing and when should you use it?"*

```go
func FuzzParse(f *testing.F) {
    // Add seed corpus
    f.Add("123")
    f.Add("-456")
    f.Add("0")
    
    // Fuzz target
    f.Fuzz(func(t *testing.T, input string) {
        result, err := Parse(input)
        if err != nil {
            return  // Invalid input is okay
        }
        
        // Check invariants
        if result < 0 && !strings.HasPrefix(input, "-") {
            t.Errorf("negative result from non-negative input: %s -> %d",
                input, result)
        }
    })
}
```

### Running Fuzzing

```bash
# Run fuzz test (indefinitely until stopped or failure)
go test -fuzz=FuzzParse

# Run for specific duration
go test -fuzz=FuzzParse -fuzztime=30s

# Run with specific worker count
go test -fuzz=FuzzParse -parallel=4
```

### Corpus Management

```
testdata/
└── fuzz/
    └── FuzzParse/
        ├── seed1    # Seed corpus
        └── crash-xxx # Crash-inducing inputs (auto-saved)
```

### Example: Fuzzing a Parser

```go
func FuzzJSON(f *testing.F) {
    // Seed with valid inputs
    f.Add(`{"name": "test"}`)
    f.Add(`{"count": 42}`)
    f.Add(`[]`)
    f.Add(`null`)
    
    f.Fuzz(func(t *testing.T, data string) {
        var v interface{}
        
        // Should not panic
        err := json.Unmarshal([]byte(data), &v)
        
        if err != nil {
            return  // Invalid JSON is expected
        }
        
        // Round-trip should work
        encoded, err := json.Marshal(v)
        if err != nil {
            t.Fatalf("Marshal failed: %v", err)
        }
        
        var v2 interface{}
        if err := json.Unmarshal(encoded, &v2); err != nil {
            t.Fatalf("Round-trip failed: %v", err)
        }
    })
}
```

---

## 6.8 Code Coverage

### Generating Coverage

```bash
# Show coverage percentage
go test -cover ./...

# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

### Coverage Modes

```bash
# set: did this statement run? (default)
go test -covermode=set -coverprofile=coverage.out

# count: how many times did it run?
go test -covermode=count -coverprofile=coverage.out

# atomic: like count but thread-safe (for concurrent tests)
go test -covermode=atomic -coverprofile=coverage.out
```

### Coverage Output

```
$ go tool cover -func=coverage.out

myapp/calculator.go:10:    Add             100.0%
myapp/calculator.go:14:    Divide          80.0%
myapp/calculator.go:24:    Multiply        100.0%
total:                     (statements)    93.3%
```

### Coverage in CI

```yaml
# GitHub Actions example
- name: Test with coverage
  run: go test -coverprofile=coverage.out ./...

- name: Upload coverage
  uses: codecov/codecov-action@v3
  with:
    file: ./coverage.out
```

---

## 6.9 Mocking Strategies

### Interface-Based Mocking

**Interview Question:** *"What is the preferred approach to mocking in Go?"*

```go
// Define interface at point of use (consumer)
type UserStore interface {
    GetUser(id string) (*User, error)
    SaveUser(user *User) error
}

// Production implementation
type PostgresUserStore struct {
    db *sql.DB
}

func (s *PostgresUserStore) GetUser(id string) (*User, error) {
    // Real database query
}

// Test implementation (fake)
type FakeUserStore struct {
    Users map[string]*User
    Err   error
}

func (s *FakeUserStore) GetUser(id string) (*User, error) {
    if s.Err != nil {
        return nil, s.Err
    }
    return s.Users[id], nil
}

// Test
func TestUserService(t *testing.T) {
    store := &FakeUserStore{
        Users: map[string]*User{
            "1": {ID: "1", Name: "Alice"},
        },
    }
    
    service := NewUserService(store)
    user, err := service.GetUser("1")
    
    // assertions...
}
```

### Manual Fakes vs Generated Mocks

**Interview Question:** *"Should you use mock generation libraries in Go?"*

| Manual Fakes | Generated Mocks (gomock, mockery) |
|--------------|-----------------------------------|
| Simple, readable | Complex, generated code |
| Full control | Automatic expectation verification |
| No dependencies | Requires tool dependency |
| Preferred in Go | Use when expectations matter |

### When to Mock

```go
// MOCK these (external dependencies):
// - Databases
// - External APIs
// - File system
// - Time
// - Random numbers

// DON'T mock these:
// - Internal packages
// - Pure functions
// - Standard library
```

### Testing with Time

```go
// Define interface for time
type Clock interface {
    Now() time.Time
}

// Real implementation
type RealClock struct{}

func (c RealClock) Now() time.Time {
    return time.Now()
}

// Test implementation
type FakeClock struct {
    time time.Time
}

func (c *FakeClock) Now() time.Time {
    return c.time
}

func (c *FakeClock) Advance(d time.Duration) {
    c.time = c.time.Add(d)
}

// Usage
type TokenService struct {
    clock Clock
}

func (s *TokenService) IsExpired(token Token) bool {
    return token.ExpiresAt.Before(s.clock.Now())
}
```

### Example: HTTP Client Mocking

```go
// Interface for HTTP client
type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}

// Service using the client
type APIClient struct {
    client HTTPClient
    baseURL string
}

// Test with fake response
func TestAPIClient(t *testing.T) {
    // Create test server
    server := httptest.NewServer(http.HandlerFunc(
        func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(http.StatusOK)
            w.Write([]byte(`{"id": 1, "name": "test"}`))
        },
    ))
    defer server.Close()
    
    client := &APIClient{
        client:  server.Client(),
        baseURL: server.URL,
    }
    
    result, err := client.GetResource("1")
    // assertions...
}
```

---

## 6.10 Integration Testing

### Build Tags for Separation

```go
//go:build integration

package myapp

import "testing"

func TestDatabaseIntegration(t *testing.T) {
    // Requires real database
    db := connectToTestDB()
    defer db.Close()
    
    // Test...
}
```

```bash
# Run only unit tests (default)
go test ./...

# Run integration tests
go test -tags=integration ./...

# Run all tests
go test -tags=integration ./...
```

### Using testcontainers-go

```go
//go:build integration

package myapp

import (
    "context"
    "testing"
    
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

func TestWithPostgres(t *testing.T) {
    ctx := context.Background()
    
    // Start container
    container, err := testcontainers.GenericContainer(ctx,
        testcontainers.GenericContainerRequest{
            ContainerRequest: testcontainers.ContainerRequest{
                Image:        "postgres:15",
                ExposedPorts: []string{"5432/tcp"},
                Env: map[string]string{
                    "POSTGRES_PASSWORD": "test",
                    "POSTGRES_DB":       "testdb",
                },
                WaitingFor: wait.ForLog("database system is ready"),
            },
            Started: true,
        })
    if err != nil {
        t.Fatal(err)
    }
    defer container.Terminate(ctx)
    
    // Get connection string
    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "5432")
    
    // Connect and test
    dsn := fmt.Sprintf("postgres://postgres:test@%s:%s/testdb",
        host, port.Port())
    
    // Run tests...
}
```

### Cleanup Functions

```go
func TestWithCleanup(t *testing.T) {
    // Create resource
    tempDir, err := os.MkdirTemp("", "test-*")
    if err != nil {
        t.Fatal(err)
    }
    
    // Register cleanup (runs after test, even on failure)
    t.Cleanup(func() {
        os.RemoveAll(tempDir)
    })
    
    // Test using tempDir...
}
```

---

## 6.11 Example Functions

### Executable Documentation

**Interview Question:** *"What are example functions and how are they verified?"*

```go
func ExampleAdd() {
    result := Add(2, 3)
    fmt.Println(result)
    // Output: 5
}

func ExampleDivide() {
    result, _ := Divide(10, 2)
    fmt.Println(result)
    // Output: 5
}

func ExampleUser_Name() {
    u := User{Name: "Alice"}
    fmt.Println(u.Name)
    // Output: Alice
}
```

### Output Verification

```go
// Exact match
func ExampleSort() {
    s := []int{3, 1, 2}
    sort.Ints(s)
    fmt.Println(s)
    // Output: [1 2 3]
}

// Unordered output (for maps, goroutines)
func ExamplePrint() {
    m := map[string]int{"a": 1, "b": 2}
    for k, v := range m {
        fmt.Printf("%s=%d\n", k, v)
    }
    // Unordered output:
    // a=1
    // b=2
}
```

### Example Naming Conventions

```go
// Package example
func Example() {}

// Function example
func ExampleFunction() {}

// Type example
func ExampleType() {}

// Method example
func ExampleType_Method() {}

// Multiple examples for same target
func ExampleFunction_basic() {}
func ExampleFunction_advanced() {}
```

---

## Interview Questions

### Beginner Level

1. **Q:** What suffix must test files have?
   **A:** `_test.go`

2. **Q:** What's the difference between `t.Error()` and `t.Fatal()`?
   **A:** `Error` reports failure and continues, `Fatal` reports and stops the test.

3. **Q:** How do you run tests in verbose mode?
   **A:** `go test -v ./...`

### Intermediate Level

4. **Q:** Explain table-driven tests and their benefits.
   **A:** Test cases in slice/map, iterate with t.Run. Benefits: easy to add cases, clear documentation, parallel execution.

5. **Q:** When would you use `TestMain`?
   **A:** Package-level setup/teardown: database connections, Docker containers, global configuration.

6. **Q:** How does the race detector work?
   **A:** Instruments code to track memory access, detects when same location accessed by different goroutines without synchronization. Use with `-race` flag.

### Advanced Level

7. **Q:** How do you mock dependencies in Go?
   **A:** Interface-based: define interface at consumer, create fake implementation for tests. Prefer manual fakes over generated mocks.

8. **Q:** Explain fuzzing and when to use it.
   **A:** Automated random input generation to find edge cases. Use for parsers, deserializers, input validation. `func FuzzXxx(f *testing.F)` with `f.Fuzz()`.

9. **Q:** How would you set up integration tests with real databases?
   **A:** Use build tags (`//go:build integration`), testcontainers-go for Docker, TestMain for setup/teardown, `t.Cleanup()` for per-test cleanup.

---

## Summary

| Topic | Key Points |
|-------|------------|
| Test Functions | `TestXxx(t *testing.T)` in `*_test.go` |
| Running | `go test ./...`, `-v`, `-run`, `-race` |
| Table-Driven | Slice of cases + `t.Run()` for each |
| Subtests | `t.Run()` for organization, `t.Parallel()` |
| TestMain | `func TestMain(m *testing.M)` for setup/teardown |
| Benchmarks | `BenchmarkXxx(b *testing.B)`, `-bench=.`, `-benchmem` |
| Fuzzing | `FuzzXxx(f *testing.F)`, `-fuzz=FuzzXxx` |
| Coverage | `-cover`, `-coverprofile`, `go tool cover` |
| Mocking | Interface-based, manual fakes preferred |

**Next Phase:** [Phase 7 — Idiomatic Go Design & Architecture](../Phase_7/Phase_7.md)

