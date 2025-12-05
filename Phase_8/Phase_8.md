# 📡 Phase 8: Network Programming & APIs

[← Back to Main Roadmap](../README.md) | [← Previous: Phase 7](../Phase_7/Phase_7.md)

---

**Objective:** Build robust networked services with Go's standard library and protocols.

**Reference:** [net/http Package](https://pkg.go.dev/net/http), [Go Blog - HTTP/2](https://go.dev/blog/h2push)

**Prerequisites:** Phase 0-7

**Estimated Duration:** 3-4 weeks

---

## 📋 Table of Contents

1. [The `net` Package](#81-the-net-package)
2. [HTTP Server](#82-http-server-nethttp)
3. [HTTP Routing (Go 1.22+)](#83-http-routing-go-122)
4. [Middleware Pattern](#84-middleware-pattern)
5. [HTTP Client](#85-http-client)
6. [JSON Handling](#86-json-handling)
7. [gRPC](#87-grpc)
8. [ConnectRPC](#88-connectrpc)
9. [API Design Best Practices](#89-api-design-best-practices)
10. [Resilience Patterns](#810-resilience-patterns)
11. [Interview Questions](#interview-questions)

---

## 8.1 The `net` Package

### TCP Server

**Interview Question:** *"How do you create a TCP server in Go?"*

```go
package main

import (
    "bufio"
    "fmt"
    "net"
)

func main() {
    // Listen on TCP port
    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        log.Fatal(err)
    }
    defer listener.Close()
    
    fmt.Println("Server listening on :8080")
    
    for {
        // Accept connection
        conn, err := listener.Accept()
        if err != nil {
            log.Println("Accept error:", err)
            continue
        }
        
        // Handle connection in goroutine
        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    defer conn.Close()
    
    reader := bufio.NewReader(conn)
    for {
        // Read until newline
        message, err := reader.ReadString('\n')
        if err != nil {
            return
        }
        
        // Echo back
        conn.Write([]byte("Echo: " + message))
    }
}
```

### TCP Client

```go
func main() {
    // Connect to server
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    // Send message
    conn.Write([]byte("Hello, server!\n"))
    
    // Read response
    response, err := bufio.NewReader(conn).ReadString('\n')
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Server response:", response)
}
```

### Timeouts

**Interview Question:** *"How do you handle timeouts in network operations?"*

```go
// Connection timeout
conn, err := net.DialTimeout("tcp", "localhost:8080", 5*time.Second)

// Read/Write deadlines
conn.SetDeadline(time.Now().Add(10 * time.Second))      // Both
conn.SetReadDeadline(time.Now().Add(10 * time.Second))  // Read only
conn.SetWriteDeadline(time.Now().Add(10 * time.Second)) // Write only

// Reset deadline (no timeout)
conn.SetDeadline(time.Time{})
```

### UDP Communication

```go
// UDP Server
func udpServer() {
    addr, _ := net.ResolveUDPAddr("udp", ":8080")
    conn, _ := net.ListenUDP("udp", addr)
    defer conn.Close()
    
    buffer := make([]byte, 1024)
    for {
        n, clientAddr, _ := conn.ReadFromUDP(buffer)
        fmt.Printf("Received from %v: %s\n", clientAddr, buffer[:n])
        conn.WriteToUDP([]byte("ACK"), clientAddr)
    }
}

// UDP Client
func udpClient() {
    conn, _ := net.Dial("udp", "localhost:8080")
    defer conn.Close()
    
    conn.Write([]byte("Hello UDP"))
    
    buffer := make([]byte, 1024)
    n, _ := conn.Read(buffer)
    fmt.Println("Response:", string(buffer[:n]))
}
```

---

## 8.2 HTTP Server (`net/http`)

### The Handler Interface

**Interview Question:** *"What is http.Handler and http.HandlerFunc?"*

```go
// The fundamental interface
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}

// HandlerFunc adapter - allows functions to be handlers
type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
    f(w, r)
}

// Usage
func hello(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello, World!"))
}

func main() {
    // As HandlerFunc
    http.Handle("/hello", http.HandlerFunc(hello))
    
    // Shortcut
    http.HandleFunc("/hello", hello)
    
    http.ListenAndServe(":8080", nil)
}
```

### Basic HTTP Server

```go
package main

import (
    "encoding/json"
    "log"
    "net/http"
)

type User struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

func main() {
    mux := http.NewServeMux()
    
    mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })
    
    mux.HandleFunc("GET /users/{id}", getUser)
    mux.HandleFunc("POST /users", createUser)
    
    server := &http.Server{
        Addr:         ":8080",
        Handler:      mux,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
    
    log.Println("Server starting on :8080")
    log.Fatal(server.ListenAndServe())
}

func getUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")  // Go 1.22+
    
    user := User{ID: id, Name: "Alice"}
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

func createUser(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    // Create user...
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}
```

### Server Configuration

**Interview Question:** *"What timeouts should you configure on an HTTP server?"*

```go
server := &http.Server{
    Addr:    ":8080",
    Handler: mux,
    
    // Time to read request headers
    ReadHeaderTimeout: 5 * time.Second,
    
    // Time to read entire request (headers + body)
    ReadTimeout: 15 * time.Second,
    
    // Time to write response
    WriteTimeout: 15 * time.Second,
    
    // Time between requests on keep-alive connections
    IdleTimeout: 60 * time.Second,
    
    // Maximum header size
    MaxHeaderBytes: 1 << 20, // 1MB
}
```

### HTTPS / TLS

```go
// Generate self-signed cert for development:
// go run $GOROOT/src/crypto/tls/generate_cert.go -host localhost

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", handler)
    
    // TLS configuration
    tlsConfig := &tls.Config{
        MinVersion: tls.VersionTLS12,
        CurvePreferences: []tls.CurveID{
            tls.CurveP256,
            tls.X25519,
        },
    }
    
    server := &http.Server{
        Addr:      ":443",
        Handler:   mux,
        TLSConfig: tlsConfig,
    }
    
    log.Fatal(server.ListenAndServeTLS("cert.pem", "key.pem"))
}
```

---

## 8.3 HTTP Routing (Go 1.22+)

### Method and Path Patterns

**Interview Question:** *"What routing improvements came in Go 1.22?"*

```go
mux := http.NewServeMux()

// Method matching
mux.HandleFunc("GET /users", listUsers)
mux.HandleFunc("POST /users", createUser)

// Path parameters
mux.HandleFunc("GET /users/{id}", getUser)
mux.HandleFunc("PUT /users/{id}", updateUser)
mux.HandleFunc("DELETE /users/{id}", deleteUser)

// Wildcards (catch remaining path)
mux.HandleFunc("GET /files/{path...}", serveFile)

// Exact match vs prefix
mux.HandleFunc("GET /api/", apiPrefix)  // Matches /api/*
mux.HandleFunc("GET /api", apiExact)    // Matches /api only
```

### Accessing Path Parameters

```go
func getUser(w http.ResponseWriter, r *http.Request) {
    // Go 1.22+ path values
    id := r.PathValue("id")
    
    if id == "" {
        http.Error(w, "ID required", http.StatusBadRequest)
        return
    }
    
    // Fetch user...
}

func serveFile(w http.ResponseWriter, r *http.Request) {
    // Wildcard captures remaining path
    path := r.PathValue("path")
    // path = "images/logo.png" for /files/images/logo.png
}
```

### Precedence Rules

```go
mux.HandleFunc("GET /users/{id}", getUser)      // Specific
mux.HandleFunc("GET /users/profile", getProfile) // More specific (wins)
mux.HandleFunc("GET /", catchAll)                // Least specific

// Precedence: More specific patterns take priority
// /users/profile -> getProfile (exact match)
// /users/123 -> getUser (parameter match)
// /anything -> catchAll (fallback)
```

### Host-Based Routing

```go
mux.HandleFunc("GET api.example.com/users", apiUsers)
mux.HandleFunc("GET www.example.com/users", webUsers)
```

---

## 8.4 Middleware Pattern

### Middleware Definition

**Interview Question:** *"What is HTTP middleware and how do you implement it in Go?"*

```go
// Middleware is a function that wraps a handler
type Middleware func(http.Handler) http.Handler

// Example: Logging middleware
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Call next handler
        next.ServeHTTP(w, r)
        
        // Log after
        log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}
```

### Common Middleware

```go
// Recovery middleware (catch panics)
func RecoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic: %v\n%s", err, debug.Stack())
                http.Error(w, "Internal Server Error", 500)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// CORS middleware
func CORSMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

// Authentication middleware
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        // Validate token...
        userID, err := validateToken(token)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        
        // Add to context
        ctx := context.WithValue(r.Context(), "userID", userID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Request ID middleware
func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }
        
        ctx := context.WithValue(r.Context(), "requestID", requestID)
        w.Header().Set("X-Request-ID", requestID)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Chaining Middleware

```go
// Chain applies middleware in order
func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        handler = middlewares[i](handler)
    }
    return handler
}

// Usage
func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("GET /users", getUsers)
    
    // Apply middleware chain
    handler := Chain(mux,
        RequestIDMiddleware,
        LoggingMiddleware,
        RecoveryMiddleware,
        CORSMiddleware,
    )
    
    http.ListenAndServe(":8080", handler)
}
```

### Response Writer Wrapper

```go
// Capture status code for logging
type responseWriter struct {
    http.ResponseWriter
    status int
    size   int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.status = code
    rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
    n, err := rw.ResponseWriter.Write(b)
    rw.size += n
    return n, err
}

func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Wrap response writer
        wrapped := &responseWriter{ResponseWriter: w, status: 200}
        
        next.ServeHTTP(wrapped, r)
        
        log.Printf("%s %s %d %d %v",
            r.Method, r.URL.Path, wrapped.status, wrapped.size, time.Since(start))
    })
}
```

---

## 8.5 HTTP Client

### Basic Client Usage

**Interview Question:** *"What are the best practices for HTTP client usage in Go?"*

```go
// DON'T use http.Get directly in production
resp, err := http.Get("https://api.example.com/data")  // No timeout!

// DO create a configured client
client := &http.Client{
    Timeout: 30 * time.Second,
}

resp, err := client.Get("https://api.example.com/data")
if err != nil {
    return err
}
defer resp.Body.Close()

body, err := io.ReadAll(resp.Body)
```

### Client Configuration

```go
// Production-ready client configuration
transport := &http.Transport{
    // Connection pool
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    MaxConnsPerHost:     100,
    IdleConnTimeout:     90 * time.Second,
    
    // TLS
    TLSClientConfig: &tls.Config{
        MinVersion: tls.VersionTLS12,
    },
    
    // Timeouts
    DialContext: (&net.Dialer{
        Timeout:   30 * time.Second,
        KeepAlive: 30 * time.Second,
    }).DialContext,
    TLSHandshakeTimeout:   10 * time.Second,
    ResponseHeaderTimeout: 10 * time.Second,
    ExpectContinueTimeout: 1 * time.Second,
}

client := &http.Client{
    Transport: transport,
    Timeout:   30 * time.Second,  // Total request timeout
}
```

### Making Requests

```go
// GET request
resp, err := client.Get(url)

// POST with JSON
data := map[string]string{"name": "Alice"}
jsonData, _ := json.Marshal(data)

resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))

// Custom request
req, err := http.NewRequestWithContext(ctx, "PUT", url, body)
if err != nil {
    return err
}

req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", "Bearer "+token)

resp, err := client.Do(req)
```

### Context Integration

```go
func fetchData(ctx context.Context, url string) ([]byte, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }
    
    return io.ReadAll(resp.Body)
}

// Usage with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

data, err := fetchData(ctx, "https://api.example.com/data")
```

### Retry Pattern

```go
func doWithRetry(req *http.Request, maxRetries int) (*http.Response, error) {
    var resp *http.Response
    var err error
    
    for i := 0; i < maxRetries; i++ {
        resp, err = client.Do(req)
        
        if err == nil && resp.StatusCode < 500 {
            return resp, nil
        }
        
        if resp != nil {
            resp.Body.Close()
        }
        
        // Exponential backoff with jitter
        backoff := time.Duration(math.Pow(2, float64(i))) * 100 * time.Millisecond
        jitter := time.Duration(rand.Intn(100)) * time.Millisecond
        time.Sleep(backoff + jitter)
    }
    
    return resp, err
}
```

---

## 8.6 JSON Handling

### Encoding and Decoding

**Interview Question:** *"What's the difference between json.Marshal and json.NewEncoder?"*

```go
// Marshal/Unmarshal - for byte slices
user := User{ID: "1", Name: "Alice"}
data, err := json.Marshal(user)  // Returns []byte

var decoded User
err = json.Unmarshal(data, &decoded)

// Encoder/Decoder - for streams (io.Reader/io.Writer)
// More efficient for HTTP (no intermediate []byte)
func handler(w http.ResponseWriter, r *http.Request) {
    // Decode from request body
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "Invalid JSON", 400)
        return
    }
    
    // Encode to response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}
```

### Struct Tags

```go
type User struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email,omitempty"`   // Omit if empty
    Password  string    `json:"-"`                  // Never include
    CreatedAt time.Time `json:"created_at"`
    Age       int       `json:"age,string"`         // Encode as string
}

// Output:
// {"id":"1","name":"Alice","created_at":"2024-01-01T00:00:00Z"}
```

### Custom Marshaling

```go
type Status int

const (
    StatusPending Status = iota
    StatusActive
    StatusInactive
)

func (s Status) MarshalJSON() ([]byte, error) {
    var str string
    switch s {
    case StatusPending:
        str = "pending"
    case StatusActive:
        str = "active"
    case StatusInactive:
        str = "inactive"
    default:
        return nil, fmt.Errorf("unknown status: %d", s)
    }
    return json.Marshal(str)
}

func (s *Status) UnmarshalJSON(data []byte) error {
    var str string
    if err := json.Unmarshal(data, &str); err != nil {
        return err
    }
    switch str {
    case "pending":
        *s = StatusPending
    case "active":
        *s = StatusActive
    case "inactive":
        *s = StatusInactive
    default:
        return fmt.Errorf("unknown status: %s", str)
    }
    return nil
}
```

### Working with Unknown JSON

```go
// map[string]interface{} for dynamic JSON
var data map[string]interface{}
json.Unmarshal(jsonBytes, &data)

// Access fields
name := data["name"].(string)
age := data["age"].(float64)  // Numbers are float64!

// Nested objects
address := data["address"].(map[string]interface{})
city := address["city"].(string)

// Arrays
items := data["items"].([]interface{})
```

### Performance: Alternative JSON Libraries

```go
// Standard library - safest, slowest
import "encoding/json"

// json-iterator - drop-in replacement, faster
import jsoniter "github.com/json-iterator/go"
var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Sonic - very fast (uses SIMD)
import "github.com/bytedance/sonic"
```

---

## 8.7 gRPC

### Protocol Buffers Definition

**Interview Question:** *"What are the benefits of gRPC over REST?"*

```protobuf
// user.proto
syntax = "proto3";

package user;

option go_package = "myapp/gen/user";

service UserService {
    rpc GetUser(GetUserRequest) returns (User);
    rpc ListUsers(ListUsersRequest) returns (stream User);  // Server streaming
    rpc CreateUsers(stream User) returns (CreateUsersResponse);  // Client streaming
    rpc Chat(stream Message) returns (stream Message);  // Bidirectional
}

message User {
    string id = 1;
    string name = 2;
    string email = 3;
}

message GetUserRequest {
    string id = 1;
}

message ListUsersRequest {
    int32 page_size = 1;
    string page_token = 2;
}

message CreateUsersResponse {
    int32 created_count = 1;
}

message Message {
    string content = 1;
}
```

### Code Generation

```bash
# Install tools
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate code
protoc --go_out=. --go-grpc_out=. user.proto
```

### gRPC Server

```go
package main

import (
    "context"
    "log"
    "net"
    
    pb "myapp/gen/user"
    "google.golang.org/grpc"
)

type userServer struct {
    pb.UnimplementedUserServiceServer
}

func (s *userServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    // Fetch user...
    return &pb.User{
        Id:    req.Id,
        Name:  "Alice",
        Email: "alice@example.com",
    }, nil
}

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatal(err)
    }
    
    server := grpc.NewServer()
    pb.RegisterUserServiceServer(server, &userServer{})
    
    log.Println("gRPC server listening on :50051")
    if err := server.Serve(lis); err != nil {
        log.Fatal(err)
    }
}
```

### gRPC Client

```go
func main() {
    conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    client := pb.NewUserServiceClient(conn)
    
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    
    user, err := client.GetUser(ctx, &pb.GetUserRequest{Id: "1"})
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("User: %v\n", user)
}
```

### Interceptors (Middleware)

```go
// Unary interceptor
func loggingInterceptor(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (interface{}, error) {
    start := time.Now()
    
    resp, err := handler(ctx, req)
    
    log.Printf("Method: %s, Duration: %v, Error: %v",
        info.FullMethod, time.Since(start), err)
    
    return resp, err
}

// Apply interceptor
server := grpc.NewServer(
    grpc.UnaryInterceptor(loggingInterceptor),
)
```

### gRPC vs REST

| Aspect | gRPC | REST |
|--------|------|------|
| Protocol | HTTP/2 | HTTP/1.1 or HTTP/2 |
| Format | Protocol Buffers (binary) | JSON (text) |
| Contract | Strict (.proto) | Flexible (OpenAPI) |
| Streaming | Built-in | Limited |
| Performance | Faster | Slower |
| Browser | Limited | Full support |
| Debugging | Harder | Easier (curl) |

---

## 8.8 ConnectRPC

### What is ConnectRPC?

**Interview Question:** *"What are the advantages of ConnectRPC over gRPC?"*

```go
// ConnectRPC: Modern RPC from Buf
// - HTTP/1.1 and HTTP/2 compatible
// - Works without gRPC-specific proxies
// - curl-friendly (supports JSON)
// - Same .proto definitions

// Benefits:
// - Simpler deployment (works with any HTTP proxy)
// - Better debugging (can use curl)
// - Same type safety as gRPC
```

### Connect Server

```go
package main

import (
    "context"
    "net/http"
    
    "connectrpc.com/connect"
    userv1 "myapp/gen/user/v1"
    "myapp/gen/user/v1/userv1connect"
)

type userServer struct{}

func (s *userServer) GetUser(
    ctx context.Context,
    req *connect.Request[userv1.GetUserRequest],
) (*connect.Response[userv1.User], error) {
    user := &userv1.User{
        Id:   req.Msg.Id,
        Name: "Alice",
    }
    return connect.NewResponse(user), nil
}

func main() {
    mux := http.NewServeMux()
    
    path, handler := userv1connect.NewUserServiceHandler(&userServer{})
    mux.Handle(path, handler)
    
    http.ListenAndServe(":8080", mux)
}
```

### Connect Client

```go
client := userv1connect.NewUserServiceClient(
    http.DefaultClient,
    "http://localhost:8080",
)

resp, err := client.GetUser(context.Background(),
    connect.NewRequest(&userv1.GetUserRequest{Id: "1"}),
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("User: %v\n", resp.Msg)
```

### Testing with curl

```bash
# JSON format (Connect)
curl -X POST http://localhost:8080/user.v1.UserService/GetUser \
  -H "Content-Type: application/json" \
  -d '{"id": "1"}'

# gRPC-Web format
curl -X POST http://localhost:8080/user.v1.UserService/GetUser \
  -H "Content-Type: application/grpc-web+json" \
  -d '{"id": "1"}'
```

---

## 8.9 API Design Best Practices

### Error Response Format

**Interview Question:** *"How should API errors be structured?"*

```go
// RFC 7807 Problem Details
type ProblemDetail struct {
    Type     string `json:"type"`               // URI identifying error type
    Title    string `json:"title"`              // Short description
    Status   int    `json:"status"`             // HTTP status code
    Detail   string `json:"detail,omitempty"`   // Human-readable explanation
    Instance string `json:"instance,omitempty"` // URI of specific occurrence
}

func writeError(w http.ResponseWriter, status int, err error) {
    problem := ProblemDetail{
        Type:   "https://example.com/errors/validation",
        Title:  "Validation Error",
        Status: status,
        Detail: err.Error(),
    }
    
    w.Header().Set("Content-Type", "application/problem+json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(problem)
}
```

### Input Validation

```go
type CreateUserRequest struct {
    Email    string `json:"email"`
    Name     string `json:"name"`
    Password string `json:"password"`
}

func (r *CreateUserRequest) Validate() error {
    if r.Email == "" {
        return errors.New("email is required")
    }
    if !strings.Contains(r.Email, "@") {
        return errors.New("invalid email format")
    }
    if len(r.Name) < 2 {
        return errors.New("name must be at least 2 characters")
    }
    if len(r.Password) < 8 {
        return errors.New("password must be at least 8 characters")
    }
    return nil
}

func createUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, err)
        return
    }
    
    if err := req.Validate(); err != nil {
        writeError(w, http.StatusBadRequest, err)
        return
    }
    
    // Create user...
}
```

### Pagination

```go
// Cursor-based (preferred for large datasets)
type ListUsersResponse struct {
    Users      []User  `json:"users"`
    NextCursor string  `json:"next_cursor,omitempty"`
    HasMore    bool    `json:"has_more"`
}

// Offset-based (simpler, but has issues with mutations)
type ListUsersResponse struct {
    Users      []User `json:"users"`
    Page       int    `json:"page"`
    PageSize   int    `json:"page_size"`
    TotalCount int    `json:"total_count"`
}
```

### Versioning

```go
// URL path versioning (most common)
mux.HandleFunc("GET /v1/users", handleUsersV1)
mux.HandleFunc("GET /v2/users", handleUsersV2)

// Header versioning
func handler(w http.ResponseWriter, r *http.Request) {
    version := r.Header.Get("API-Version")
    switch version {
    case "2":
        handleV2(w, r)
    default:
        handleV1(w, r)
    }
}
```

### Idempotency

```go
// Idempotency key for safe retries
func createPayment(w http.ResponseWriter, r *http.Request) {
    idempotencyKey := r.Header.Get("Idempotency-Key")
    if idempotencyKey == "" {
        writeError(w, http.StatusBadRequest, errors.New("Idempotency-Key required"))
        return
    }
    
    // Check if already processed
    if result, found := cache.Get(idempotencyKey); found {
        json.NewEncoder(w).Encode(result)
        return
    }
    
    // Process payment...
    result := processPayment(...)
    
    // Store result for idempotency
    cache.Set(idempotencyKey, result, 24*time.Hour)
    
    json.NewEncoder(w).Encode(result)
}
```

---

## 8.10 Resilience Patterns

### Circuit Breaker

**Interview Question:** *"Explain the circuit breaker pattern."*

```go
type CircuitBreaker struct {
    failures    int
    successes   int
    state       string // "closed", "open", "half-open"
    threshold   int
    resetAfter  time.Duration
    lastFailure time.Time
    mu          sync.Mutex
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    // Check state
    switch cb.state {
    case "open":
        if time.Since(cb.lastFailure) > cb.resetAfter {
            cb.state = "half-open"
        } else {
            return errors.New("circuit breaker is open")
        }
    }
    
    // Execute function
    err := fn()
    
    if err != nil {
        cb.failures++
        cb.lastFailure = time.Now()
        
        if cb.failures >= cb.threshold {
            cb.state = "open"
        }
        return err
    }
    
    // Success
    if cb.state == "half-open" {
        cb.successes++
        if cb.successes >= 3 {
            cb.state = "closed"
            cb.failures = 0
            cb.successes = 0
        }
    }
    
    return nil
}
```

### Rate Limiting

```go
// Token bucket rate limiter
type RateLimiter struct {
    rate     float64
    capacity float64
    tokens   float64
    lastTime time.Time
    mu       sync.Mutex
}

func (rl *RateLimiter) Allow() bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    now := time.Now()
    elapsed := now.Sub(rl.lastTime).Seconds()
    
    // Add tokens based on elapsed time
    rl.tokens = math.Min(rl.capacity, rl.tokens+elapsed*rl.rate)
    rl.lastTime = now
    
    if rl.tokens >= 1 {
        rl.tokens--
        return true
    }
    return false
}

// Middleware
func RateLimitMiddleware(limiter *RateLimiter) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

### Bulkhead Pattern

```go
// Limit concurrent requests to a service
type Bulkhead struct {
    sem chan struct{}
}

func NewBulkhead(maxConcurrent int) *Bulkhead {
    return &Bulkhead{
        sem: make(chan struct{}, maxConcurrent),
    }
}

func (b *Bulkhead) Execute(ctx context.Context, fn func() error) error {
    select {
    case b.sem <- struct{}{}:
        defer func() { <-b.sem }()
        return fn()
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

---

## Interview Questions

### Beginner Level

1. **Q:** What is `http.Handler`?
   **A:** An interface with `ServeHTTP(ResponseWriter, *Request)`. Any type implementing this can handle HTTP requests.

2. **Q:** How do you get path parameters in Go 1.22+?
   **A:** `r.PathValue("paramName")` for routes defined with `{paramName}`.

3. **Q:** What's the difference between `json.Marshal` and `json.NewEncoder`?
   **A:** Marshal returns `[]byte`, Encoder writes directly to `io.Writer` (more efficient for streams like HTTP).

### Intermediate Level

4. **Q:** How should HTTP client timeouts be configured?
   **A:** Set `Client.Timeout` for total request time, configure Transport for granular control (dial, TLS handshake, response headers).

5. **Q:** Explain middleware chaining in Go.
   **A:** Middleware are `func(http.Handler) http.Handler`. Chain by wrapping: `recovery(logging(auth(handler)))`.

6. **Q:** What's the difference between gRPC and ConnectRPC?
   **A:** ConnectRPC works with HTTP/1.1, curl-friendly, doesn't require gRPC-specific proxies, same .proto definitions.

### Advanced Level

7. **Q:** Implement a basic circuit breaker.
   **A:** Track failures, open circuit after threshold, allow retry after timeout (half-open state).

8. **Q:** How would you implement graceful shutdown for an HTTP server?
   **A:** `signal.NotifyContext()` for signals, `server.Shutdown(ctx)` with timeout context.

9. **Q:** Design an API that handles idempotent operations.
   **A:** Use Idempotency-Key header, cache results by key, return cached result on retry.

---

## Summary

| Topic | Key Points |
|-------|------------|
| TCP/UDP | `net.Listen`, `net.Dial`, timeouts with `SetDeadline` |
| HTTP Server | `http.Handler` interface, `ServeMux`, configure timeouts |
| Routing | Go 1.22+ patterns: `"GET /users/{id}"`, `r.PathValue()` |
| Middleware | `func(http.Handler) http.Handler`, chain for logging/auth/recovery |
| HTTP Client | Configure timeouts, reuse client, context for cancellation |
| JSON | Tags for field mapping, Encoder/Decoder for streams |
| gRPC | Protocol Buffers, code generation, interceptors |
| ConnectRPC | HTTP/1.1 compatible, curl-friendly, same type safety |
| Resilience | Circuit breaker, rate limiting, bulkhead, retries with backoff |

**Next Phase:** [Phase 9 — Data Persistence](../Phase_9/Phase_9.md)

