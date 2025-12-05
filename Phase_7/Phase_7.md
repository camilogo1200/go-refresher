# 📐 Phase 7: Idiomatic Go Design & Architecture

[← Back to Main Roadmap](../README.md) | [← Previous: Phase 6](../Phase_6/Phase_6.md)

---

**Objective:** Apply Go-specific patterns that embrace simplicity and composition.

**Reference:** [Effective Go](https://go.dev/doc/effective_go), [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

**Prerequisites:** Phase 0-6

**Estimated Duration:** 2-3 weeks

---

## 📋 Table of Contents

1. [Package Design Principles](#71-package-design-principles)
2. [SOLID Principles in Go](#72-solid-principles-in-go)
3. [Composition Over Inheritance](#73-composition-over-inheritance)
4. [Dependency Injection](#74-dependency-injection)
5. [Functional Options Pattern](#75-functional-options-pattern)
6. [Error Handling Architecture](#76-error-handling-architecture)
7. [Clean Architecture vs. Go Pragmatism](#77-clean-architecture-vs-go-pragmatism)
8. [Common Patterns](#78-common-patterns)
9. [Anti-Patterns](#79-anti-patterns)
10. [Interview Questions](#interview-questions)

---

## 7.1 Package Design Principles

### Package Naming

**Interview Question:** *"What are the rules and conventions for Go package names?"*

```go
// GOOD: Short, lowercase, no underscores or mixedCaps
package http
package json
package user
package auth

// BAD: Verbose, utility names, stutter
package httputils        // Too generic
package user_service     // No underscores
package UserRepository   // No mixed caps
```

**Naming Rules:**
- Short, concise, lowercase
- Singular (not plural)
- No underscores or mixedCaps
- Descriptive of contents
- Avoid: `util`, `common`, `base`, `helpers`

### Package Purpose

**Interview Question:** *"How should you organize packages in a Go project?"*

```
// Each package should have ONE clear purpose

// GOOD: Clear, focused packages
myapp/
├── user/           # User domain logic
├── auth/           # Authentication
├── storage/        # Data persistence
└── api/            # HTTP handlers

// BAD: Kitchen sink packages
myapp/
├── utils/          # What's in here?
├── helpers/        # Vague
├── common/         # Too broad
└── models/         # Mixed concerns
```

### Naming by Purpose, Not Contents

```go
// BAD: Named after what it contains
package models
type User struct{}

// GOOD: Named after what it provides
package user
type User struct{}
```

### Export Rules

**Interview Question:** *"Explain Go's visibility rules."*

```go
// Uppercase = exported (public)
func PublicFunction() {}
type PublicType struct {
    PublicField string
}

// Lowercase = unexported (package-private)
func privateFunction() {}
type privateType struct {
    privateField string
}

// Exported type with unexported fields
type User struct {
    ID       string    // Exported
    password string    // Unexported - hidden from other packages
}
```

### The `internal/` Package

```
myapp/
├── cmd/
│   └── server/
│       └── main.go
├── internal/           # Can only be imported by myapp and children
│   ├── cache/          # Internal caching
│   └── metrics/        # Internal metrics
└── pkg/                # Can be imported by anyone (optional)
    └── client/
```

```go
// From cmd/server/main.go:
import "myapp/internal/cache"  // OK

// From another module:
import "myapp/internal/cache"  // COMPILE ERROR
```

### Cyclic Import Prevention

```go
// PROBLEM: A imports B, B imports A
// package a
import "myapp/b"

// package b
import "myapp/a"  // Compilation error!

// SOLUTIONS:

// 1. Move shared types to third package
// package types
type User struct{}

// package a
import "myapp/types"

// package b
import "myapp/types"

// 2. Use interfaces at boundaries
// package a
type UserStore interface {
    GetUser(id string) (*User, error)
}

// package b implements UserStore without importing a
```

### Example: Package Structure

```
myapp/
├── cmd/
│   └── api/
│       └── main.go         # Composition root
├── internal/
│   ├── user/
│   │   ├── user.go         # User domain type
│   │   ├── service.go      # User business logic
│   │   └── repository.go   # User data access
│   ├── auth/
│   │   ├── auth.go
│   │   └── jwt.go
│   └── platform/
│       ├── postgres/
│       └── redis/
├── api/
│   └── http/
│       └── handler.go      # HTTP handlers
├── go.mod
└── go.sum
```

---

## 7.2 SOLID Principles in Go

### Single Responsibility Principle

**Interview Question:** *"How does Go approach Single Responsibility?"*

```go
// BAD: Multiple responsibilities
type UserService struct {
    db     *sql.DB
    cache  *redis.Client
    mailer *smtp.Client
}

func (s *UserService) CreateUser(u *User) error { ... }
func (s *UserService) SendEmail(to, body string) error { ... }  // Wrong place!
func (s *UserService) CacheUser(u *User) error { ... }          // Wrong place!

// GOOD: Focused responsibilities
type UserService struct {
    repo   UserRepository
    events EventPublisher
}

func (s *UserService) CreateUser(u *User) error {
    if err := s.repo.Save(u); err != nil {
        return err
    }
    return s.events.Publish("user.created", u)
}
```

### Open/Closed Principle

```go
// Extend behavior without modifying existing code

// BAD: Must modify function for new shapes
func Area(shape string, dimensions ...float64) float64 {
    switch shape {
    case "circle":
        return math.Pi * dimensions[0] * dimensions[0]
    case "rectangle":
        return dimensions[0] * dimensions[1]
    // Must add more cases here...
    }
    return 0
}

// GOOD: Open for extension, closed for modification
type Shape interface {
    Area() float64
}

type Circle struct{ Radius float64 }
func (c Circle) Area() float64 { return math.Pi * c.Radius * c.Radius }

type Rectangle struct{ Width, Height float64 }
func (r Rectangle) Area() float64 { return r.Width * r.Height }

// New shapes don't require modifying existing code
type Triangle struct{ Base, Height float64 }
func (t Triangle) Area() float64 { return 0.5 * t.Base * t.Height }

func TotalArea(shapes []Shape) float64 {
    total := 0.0
    for _, s := range shapes {
        total += s.Area()
    }
    return total
}
```

### Liskov Substitution Principle

```go
// Interface implementations must be interchangeable

type Reader interface {
    Read(p []byte) (n int, err error)
}

// All these satisfy Reader and can be used interchangeably
var r Reader
r = os.Stdin           // File
r = strings.NewReader("hello")  // String
r = bytes.NewBuffer(data)       // Buffer

// Function works with ANY Reader
func Process(r Reader) error {
    // Implementation doesn't care about concrete type
    buf := make([]byte, 1024)
    n, err := r.Read(buf)
    // ...
}
```

### Interface Segregation Principle

**Interview Question:** *"What does 'accept interfaces, return structs' mean?"*

```go
// BAD: Large interface
type UserManager interface {
    CreateUser(u *User) error
    UpdateUser(u *User) error
    DeleteUser(id string) error
    GetUser(id string) (*User, error)
    ListUsers() ([]*User, error)
    SearchUsers(query string) ([]*User, error)
    ExportUsers() ([]byte, error)
    ImportUsers(data []byte) error
}

// GOOD: Small, focused interfaces
type UserCreator interface {
    CreateUser(u *User) error
}

type UserGetter interface {
    GetUser(id string) (*User, error)
}

type UserLister interface {
    ListUsers() ([]*User, error)
}

// Compose interfaces when needed
type UserReadWriter interface {
    UserGetter
    UserCreator
}

// Functions accept minimal interface needed
func GetUserHandler(getter UserGetter) http.Handler {
    // Only needs GetUser
}
```

### Dependency Inversion Principle

```go
// BAD: Depend on concrete types
type UserService struct {
    db *PostgresDB  // Concrete type
}

// GOOD: Depend on interfaces
type UserRepository interface {
    Save(u *User) error
    Find(id string) (*User, error)
}

type UserService struct {
    repo UserRepository  // Interface
}

// Now can use any implementation
func main() {
    // Production
    service := NewUserService(&PostgresRepository{})
    
    // Testing
    service := NewUserService(&FakeRepository{})
}
```

---

## 7.3 Composition Over Inheritance

### Struct Embedding

**Interview Question:** *"Explain embedding in Go. Is it inheritance?"*

```go
// Embedding promotes methods and fields
type Logger struct {
    prefix string
}

func (l *Logger) Log(msg string) {
    fmt.Printf("[%s] %s\n", l.prefix, msg)
}

type Server struct {
    *Logger  // Embedded - Logger's methods are promoted
    addr string
}

func main() {
    s := &Server{
        Logger: &Logger{prefix: "SERVER"},
        addr:   ":8080",
    }
    
    s.Log("Starting...")  // Calls Logger.Log directly
}
```

### Embedding Is NOT Inheritance

```go
type Animal struct {
    Name string
}

func (a *Animal) Speak() {
    fmt.Println("...")
}

type Dog struct {
    *Animal
}

func (d *Dog) Speak() {
    fmt.Println("Woof!")
}

func main() {
    d := &Dog{Animal: &Animal{Name: "Buddy"}}
    d.Speak()         // "Woof!" - Dog's method
    d.Animal.Speak()  // "..." - Animal's method
    
    // But: Dog is NOT an Animal (no polymorphism)
    var a *Animal = d  // COMPILE ERROR!
}
```

### Interface Embedding

```go
// Compose interfaces
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

type ReadWriter interface {
    Reader
    Writer
}

// Any type implementing ReadWriter also implements Reader and Writer
type Buffer struct{}

func (b *Buffer) Read(p []byte) (int, error)  { return 0, nil }
func (b *Buffer) Write(p []byte) (int, error) { return 0, nil }

var rw ReadWriter = &Buffer{}  // OK
var r Reader = rw              // OK
var w Writer = rw              // OK
```

### When to Use Embedding

```go
// GOOD: Adding behavior without "is-a" relationship
type Metrics struct {
    requestCount int64
    mu           sync.Mutex
}

type Server struct {
    *Metrics  // Server has metrics capability
    // ...
}

// AVOID: Embedding for code reuse when not logically appropriate
type User struct {
    *Logger  // User "is-a" logger? No! Inject instead.
}
```

### Example: Composition Pattern

```go
// Build complex types from simple ones

type Timestamps struct {
    CreatedAt time.Time
    UpdatedAt time.Time
}

type SoftDelete struct {
    DeletedAt *time.Time
}

type User struct {
    ID   string
    Name string
    Timestamps   // Embedded
    SoftDelete   // Embedded
}

func main() {
    u := User{
        ID:   "1",
        Name: "Alice",
        Timestamps: Timestamps{
            CreatedAt: time.Now(),
        },
    }
    
    fmt.Println(u.CreatedAt)  // Accessed directly
}
```

---

## 7.4 Dependency Injection

### Constructor Injection

**Interview Question:** *"How do you implement dependency injection in Go?"*

```go
// Dependencies passed through constructor
type UserService struct {
    repo   UserRepository
    logger Logger
    cache  Cache
}

func NewUserService(repo UserRepository, logger Logger, cache Cache) *UserService {
    return &UserService{
        repo:   repo,
        logger: logger,
        cache:  cache,
    }
}

// Usage (composition root - usually main)
func main() {
    db := postgres.Connect()
    repo := postgres.NewUserRepository(db)
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    cache := redis.NewCache()
    
    userService := NewUserService(repo, logger, cache)
    
    // Start application...
}
```

### Method Injection

```go
// When dependency varies per call
type ReportService struct{}

func (s *ReportService) Generate(ctx context.Context, w io.Writer) error {
    // w is injected per call
    return nil
}

// Usage
func main() {
    svc := &ReportService{}
    
    // Different writers per call
    svc.Generate(ctx, os.Stdout)
    svc.Generate(ctx, &bytes.Buffer{})
    svc.Generate(ctx, file)
}
```

### Wire (Compile-Time DI)

```go
// wire.go
//go:build wireinject

package main

import "github.com/google/wire"

func InitializeApp() (*Application, error) {
    wire.Build(
        NewDatabase,
        NewUserRepository,
        NewUserService,
        NewHTTPServer,
        NewApplication,
    )
    return nil, nil
}
```

### Avoid Magic DI Frameworks

**Interview Question:** *"Why does Go discourage reflection-based DI frameworks?"*

```go
// BAD: Magic DI with reflection
type Container struct {
    // Uses reflection to resolve dependencies
}

func (c *Container) Resolve(target interface{}) error {
    // Magic happens here...
}

// Problems:
// - Errors at runtime, not compile time
// - Hard to understand flow
// - Hidden dependencies
// - Against Go philosophy

// GOOD: Explicit wiring
func main() {
    // Clear, explicit, compile-time checked
    db := NewDB()
    repo := NewRepo(db)
    service := NewService(repo)
    handler := NewHandler(service)
    
    http.ListenAndServe(":8080", handler)
}
```

---

## 7.5 Functional Options Pattern

### The Problem

**Interview Question:** *"Explain the functional options pattern and when to use it."*

```go
// BAD: Constructor with many parameters
func NewServer(addr string, port int, timeout time.Duration, 
    maxConns int, tls *tls.Config, logger *log.Logger) *Server {
    // ...
}

// Problems:
// - Hard to read
// - Must provide all values
// - Breaking change to add new option
// - Unclear what's required vs optional

// BAD: Config struct (slightly better)
type ServerConfig struct {
    Addr     string
    Port     int
    Timeout  time.Duration
    MaxConns int
    TLS      *tls.Config
    Logger   *log.Logger
}

func NewServer(cfg ServerConfig) *Server {
    // Still: what's the default? What's required?
}
```

### The Solution

```go
type Server struct {
    addr     string
    port     int
    timeout  time.Duration
    maxConns int
    tls      *tls.Config
    logger   *log.Logger
}

// Option is a function that configures Server
type Option func(*Server)

// Option constructors
func WithPort(port int) Option {
    return func(s *Server) {
        s.port = port
    }
}

func WithTimeout(d time.Duration) Option {
    return func(s *Server) {
        s.timeout = d
    }
}

func WithMaxConns(n int) Option {
    return func(s *Server) {
        s.maxConns = n
    }
}

func WithTLS(cfg *tls.Config) Option {
    return func(s *Server) {
        s.tls = cfg
    }
}

func WithLogger(l *log.Logger) Option {
    return func(s *Server) {
        s.logger = l
    }
}

// Constructor with functional options
func NewServer(addr string, opts ...Option) *Server {
    // Default values
    s := &Server{
        addr:     addr,
        port:     8080,
        timeout:  30 * time.Second,
        maxConns: 100,
        logger:   log.Default(),
    }
    
    // Apply options
    for _, opt := range opts {
        opt(s)
    }
    
    return s
}
```

### Usage

```go
func main() {
    // Basic usage with defaults
    s1 := NewServer("localhost")
    
    // With some options
    s2 := NewServer("localhost",
        WithPort(9090),
        WithTimeout(60*time.Second),
    )
    
    // With all options
    s3 := NewServer("localhost",
        WithPort(443),
        WithTimeout(120*time.Second),
        WithMaxConns(1000),
        WithTLS(tlsConfig),
        WithLogger(customLogger),
    )
}
```

### Benefits

| Benefit | Explanation |
|---------|-------------|
| Backward compatible | Add new options without breaking existing code |
| Self-documenting | `WithTimeout(5*time.Second)` is clear |
| Defaults built-in | Required params explicit, optional have defaults |
| Flexible | Any combination of options |
| Type-safe | Compiler checks option functions |

---

## 7.6 Error Handling Architecture

### Error Types per Domain

```go
// domain/user/errors.go
package user

import "errors"

// Sentinel errors
var (
    ErrNotFound     = errors.New("user not found")
    ErrDuplicate    = errors.New("user already exists")
    ErrInvalidEmail = errors.New("invalid email format")
)

// Structured error
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}
```

### Error Wrapping at Boundaries

```go
// Layer 1: Repository
func (r *Repository) Find(id string) (*User, error) {
    row := r.db.QueryRow(query, id)
    if err := row.Scan(...); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrNotFound  // Domain error
        }
        return nil, fmt.Errorf("query user: %w", err)  // Wrap
    }
    return user, nil
}

// Layer 2: Service (adds context)
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    user, err := s.repo.Find(id)
    if err != nil {
        return nil, fmt.Errorf("get user %s: %w", id, err)  // Add context
    }
    return user, nil
}

// Layer 3: Handler (decides response)
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    user, err := h.service.GetUser(r.Context(), id)
    if err != nil {
        if errors.Is(err, user.ErrNotFound) {
            http.Error(w, "User not found", http.StatusNotFound)
            return
        }
        h.logger.Error("get user", "error", err)  // Log once here
        http.Error(w, "Internal error", http.StatusInternalServerError)
        return
    }
    // ...
}
```

### Error Behavior Interfaces

```go
// Define error behaviors
type temporary interface {
    Temporary() bool
}

type timeout interface {
    Timeout() bool
}

// Implement on error types
type NetworkError struct {
    Op  string
    Err error
}

func (e *NetworkError) Error() string {
    return fmt.Sprintf("network error during %s: %v", e.Op, e.Err)
}

func (e *NetworkError) Temporary() bool {
    return true  // Network errors are typically temporary
}

func (e *NetworkError) Timeout() bool {
    return errors.Is(e.Err, context.DeadlineExceeded)
}

// Check behavior, not type
func shouldRetry(err error) bool {
    var t temporary
    if errors.As(err, &t) {
        return t.Temporary()
    }
    return false
}
```

---

## 7.7 Clean Architecture vs. Go Pragmatism

### The Problem with Clean Architecture

**Interview Question:** *"Is Clean Architecture appropriate for Go projects?"*

```
// Traditional Clean Architecture layers:
Entities → Use Cases → Controllers → Presenters → DB/UI

// Problems in Go:
// - Too many abstractions for simple cases
// - Interfaces defined at wrong level
// - Over-engineering for small services
// - Against Go's "accept interfaces, return structs"
```

### Go's Pragmatic Approach

```
// Typical Go service layers:
Handler → Service → Repository

// That's it! Only add layers when needed.
```

### Vertical Slices

```
// Instead of layers by type:
handlers/
    user_handler.go
    product_handler.go
services/
    user_service.go
    product_service.go
repositories/
    user_repo.go
    product_repo.go

// Group by feature (vertical slice):
user/
    handler.go
    service.go
    repository.go
    user.go
product/
    handler.go
    service.go
    repository.go
    product.go
```

### When to Add Abstraction

```go
// Start simple
type UserService struct {
    db *sql.DB  // Direct dependency (fine for small services)
}

// Add interface when needed (second implementation, testing)
type UserRepository interface {
    Find(id string) (*User, error)
}

type UserService struct {
    repo UserRepository  // Now abstracted
}
```

### Example: Pragmatic Structure

```go
// Simple CRUD service - no need for complex architecture

// user/user.go - domain type
type User struct {
    ID    string
    Email string
    Name  string
}

// user/service.go - business logic
type Service struct {
    db *sql.DB
}

func (s *Service) Create(ctx context.Context, u *User) error {
    _, err := s.db.ExecContext(ctx, 
        "INSERT INTO users (id, email, name) VALUES ($1, $2, $3)",
        u.ID, u.Email, u.Name)
    return err
}

func (s *Service) Find(ctx context.Context, id string) (*User, error) {
    var u User
    err := s.db.QueryRowContext(ctx,
        "SELECT id, email, name FROM users WHERE id = $1", id).
        Scan(&u.ID, &u.Email, &u.Name)
    if err == sql.ErrNoRows {
        return nil, ErrNotFound
    }
    return &u, err
}

// user/handler.go - HTTP handlers
type Handler struct {
    svc *Service
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    user, err := h.svc.Find(r.Context(), id)
    if errors.Is(err, ErrNotFound) {
        http.Error(w, "Not found", 404)
        return
    }
    if err != nil {
        http.Error(w, "Internal error", 500)
        return
    }
    json.NewEncoder(w).Encode(user)
}
```

---

## 7.8 Common Patterns

### Constructor Pattern

```go
type Config struct {
    Timeout time.Duration
}

type Service struct {
    config Config
    db     *sql.DB
}

// NewService validates and constructs
func NewService(db *sql.DB, cfg Config) (*Service, error) {
    if db == nil {
        return nil, errors.New("db is required")
    }
    if cfg.Timeout <= 0 {
        cfg.Timeout = 30 * time.Second  // Default
    }
    return &Service{
        db:     db,
        config: cfg,
    }, nil
}
```

### Builder Pattern

```go
type QueryBuilder struct {
    table   string
    columns []string
    where   []string
    args    []interface{}
    limit   int
}

func NewQueryBuilder(table string) *QueryBuilder {
    return &QueryBuilder{table: table}
}

func (b *QueryBuilder) Select(cols ...string) *QueryBuilder {
    b.columns = cols
    return b
}

func (b *QueryBuilder) Where(condition string, arg interface{}) *QueryBuilder {
    b.where = append(b.where, condition)
    b.args = append(b.args, arg)
    return b
}

func (b *QueryBuilder) Limit(n int) *QueryBuilder {
    b.limit = n
    return b
}

func (b *QueryBuilder) Build() (string, []interface{}) {
    // Build SQL string
    return sql, b.args
}

// Usage
sql, args := NewQueryBuilder("users").
    Select("id", "name").
    Where("status = ?", "active").
    Limit(10).
    Build()
```

### Result Type Pattern

```go
// For operations that can fail with data
type Result[T any] struct {
    Data  T
    Error error
}

func FetchUser(id string) Result[*User] {
    user, err := db.Find(id)
    return Result[*User]{Data: user, Error: err}
}

// Or with channels
func FetchAsync(id string) <-chan Result[*User] {
    ch := make(chan Result[*User], 1)
    go func() {
        user, err := db.Find(id)
        ch <- Result[*User]{Data: user, Error: err}
    }()
    return ch
}
```

---

## 7.9 Anti-Patterns

### God Packages

```go
// BAD: Everything in one package
package app

type User struct{}
type Order struct{}
type Product struct{}
type UserService struct{}
type OrderService struct{}
type AuthService struct{}
// ... 50 more types

// GOOD: Split by domain
package user
package order
package product
package auth
```

### Stuttering

```go
// BAD: Package name repeated in type
package user

type UserService struct{}  // user.UserService
type UserRepository struct{}

// GOOD: No stutter
package user

type Service struct{}     // user.Service
type Repository struct{}  // user.Repository
```

### Interface Pollution

```go
// BAD: Interface defined before needed
type UserRepository interface {
    Find(id string) (*User, error)
    Create(u *User) error
    Update(u *User) error
    Delete(id string) error
    List() ([]*User, error)
    Search(q string) ([]*User, error)
}

// Only one implementation exists!
type PostgresUserRepo struct{}

// GOOD: Start with struct, add interface when needed
type UserRepo struct{ db *sql.DB }

// Add interface later when second implementation needed
```

### Premature Abstraction

```go
// BAD: Over-engineered for "flexibility"
type DataProcessor interface {
    Process(data interface{}) (interface{}, error)
}

type UserDataProcessor struct{}
func (p *UserDataProcessor) Process(data interface{}) (interface{}, error) {
    u := data.(*User)
    // ...
}

// GOOD: Simple and direct
func ProcessUser(u *User) (*ProcessedUser, error) {
    // ...
}
```

### Empty Interface Abuse

```go
// BAD: Using interface{} when specific type works
func ProcessData(data interface{}) {
    // Type assertions everywhere
    switch v := data.(type) {
    case string:
    case int:
    case *User:
    }
}

// GOOD: Use generics or specific types
func ProcessString(s string) {}
func ProcessInt(n int) {}
func ProcessUser(u *User) {}

// Or with generics
func Process[T Processable](data T) {}
```

---

## Interview Questions

### Beginner Level

1. **Q:** What should package names look like?
   **A:** Short, lowercase, singular nouns. No underscores, mixedCaps, or generic names like "utils".

2. **Q:** What does uppercase vs lowercase mean for Go identifiers?
   **A:** Uppercase = exported (public), lowercase = unexported (package-private).

3. **Q:** How do you prevent cyclic imports?
   **A:** Use interfaces at boundaries, move shared types to third package, or restructure packages.

### Intermediate Level

4. **Q:** Explain "accept interfaces, return structs".
   **A:** Functions should accept interfaces (flexibility for callers) but return concrete types (clarity for users, avoids nil interface issues).

5. **Q:** What is the functional options pattern?
   **A:** Using `func(*Type)` functions to configure struct. Benefits: backward compatible, self-documenting, clear defaults.

6. **Q:** When should you define an interface in Go?
   **A:** When you have (or anticipate) multiple implementations, for testing boundaries, or at consumption point. Don't preemptively create interfaces.

### Advanced Level

7. **Q:** Is embedding inheritance?
   **A:** No. Embedding promotes methods/fields but doesn't create "is-a" relationship. No polymorphism - embedded type is not a subtype.

8. **Q:** How would you structure a Go microservice?
   **A:** Pragmatically: Handler → Service → Repository layers. Use vertical slices (group by feature). Only abstract when needed.

9. **Q:** Why avoid DI containers in Go?
   **A:** Runtime errors vs compile-time, hidden dependencies, against Go's explicit philosophy. Prefer explicit wiring in main.

---

## Summary

| Topic | Key Points |
|-------|------------|
| Packages | Short, lowercase, single purpose, no utils |
| SOLID | Interfaces for abstraction, composition over inheritance |
| Composition | Embedding promotes but isn't inheritance |
| DI | Constructor injection, explicit wiring, no magic |
| Options | Functional options for configurable constructors |
| Errors | Domain types, wrap at boundaries, behavior interfaces |
| Architecture | Pragmatic layers, vertical slices, add abstraction when needed |
| Anti-patterns | God packages, stuttering, interface pollution, premature abstraction |

**Next Phase:** [Phase 8 — Network Programming & APIs](../Phase_8/Phase_8.md)

