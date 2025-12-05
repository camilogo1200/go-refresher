# ☁️ Phase 10: Cloud Native & Production

[← Back to Main Roadmap](../README.md) | [← Previous: Phase 9](../Phase_9/Phase_9.md)

---

**Objective:** Deploy, observe, and operate Go services in production environments.

**Reference:** [12-Factor App](https://12factor.net/), [Prometheus Go Client](https://github.com/prometheus/client_golang)

**Prerequisites:** Phase 0-9

**Estimated Duration:** 2-3 weeks

---

## 📋 Table of Contents

1. [Containerization](#101-containerization)
2. [Configuration Management](#102-configuration-management)
3. [Structured Logging](#103-structured-logging-logslog)
4. [Distributed Tracing](#104-distributed-tracing-opentelemetry)
5. [Metrics](#105-metrics-prometheus)
6. [Health Checks](#106-health-checks)
7. [Graceful Shutdown](#107-graceful-shutdown)
8. [Profile-Guided Optimization](#108-profile-guided-optimization-pgo)
9. [Profiling in Production](#109-profiling-in-production)
10. [Interview Questions](#interview-questions)

---

## 10.1 Containerization

### Multi-Stage Docker Build

**Interview Question:** *"How do you build optimized Docker images for Go applications?"*

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=${VERSION}" \
    -o /app/server ./cmd/server

# Runtime stage
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /app/server /server

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/server"]
```

### Base Image Options

| Image | Size | Shell | Use Case |
|-------|------|-------|----------|
| `golang:1.22` | ~800MB | Yes | Building only |
| `alpine:3.19` | ~7MB | Yes | Need shell/tools |
| `gcr.io/distroless/static` | ~2MB | No | Production |
| `scratch` | 0MB | No | Minimal, static binaries |

### Scratch Image

```dockerfile
FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM scratch
COPY --from=builder /server /server
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/server"]
```

### Build Flags Explained

```bash
# CGO_ENABLED=0
# - Disable cgo (C interop)
# - Produces static binary
# - Required for scratch/distroless

# -ldflags="-w -s"
# - -w: Omit DWARF symbol table
# - -s: Omit symbol table
# - Reduces binary size ~30%

# Version injection
go build -ldflags="-X main.version=1.0.0 -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

```go
// Access injected variables
var (
    version   = "dev"
    buildTime = "unknown"
)

func main() {
    fmt.Printf("Version: %s, Built: %s\n", version, buildTime)
}
```

### Docker Compose for Development

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://user:pass@db:5432/mydb?sslmode=disable
      - REDIS_URL=redis://cache:6379
    depends_on:
      - db
      - cache

  db:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
      POSTGRES_DB: mydb
    volumes:
      - postgres_data:/var/lib/postgresql/data

  cache:
    image: redis:7-alpine

volumes:
  postgres_data:
```

---

## 10.2 Configuration Management

### 12-Factor App Configuration

**Interview Question:** *"How should production Go applications handle configuration?"*

```go
// Configuration from environment variables (12-Factor App)
type Config struct {
    Port        int           `env:"PORT" envDefault:"8080"`
    DatabaseURL string        `env:"DATABASE_URL,required"`
    RedisURL    string        `env:"REDIS_URL" envDefault:"localhost:6379"`
    LogLevel    string        `env:"LOG_LEVEL" envDefault:"info"`
    Timeout     time.Duration `env:"TIMEOUT" envDefault:"30s"`
}

// Manual parsing
func LoadConfig() (*Config, error) {
    cfg := &Config{}
    
    // Required
    cfg.DatabaseURL = os.Getenv("DATABASE_URL")
    if cfg.DatabaseURL == "" {
        return nil, errors.New("DATABASE_URL is required")
    }
    
    // With default
    if port := os.Getenv("PORT"); port != "" {
        p, err := strconv.Atoi(port)
        if err != nil {
            return nil, fmt.Errorf("invalid PORT: %w", err)
        }
        cfg.Port = p
    } else {
        cfg.Port = 8080
    }
    
    return cfg, nil
}
```

### Using envconfig Library

```go
import "github.com/kelseyhightower/envconfig"

type Config struct {
    Port        int           `envconfig:"PORT" default:"8080"`
    DatabaseURL string        `envconfig:"DATABASE_URL" required:"true"`
    RedisURL    string        `envconfig:"REDIS_URL" default:"localhost:6379"`
    Debug       bool          `envconfig:"DEBUG" default:"false"`
    Timeout     time.Duration `envconfig:"TIMEOUT" default:"30s"`
}

func LoadConfig() (*Config, error) {
    var cfg Config
    if err := envconfig.Process("", &cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}
```

### Configuration Validation

```go
func (c *Config) Validate() error {
    if c.Port < 1 || c.Port > 65535 {
        return fmt.Errorf("invalid port: %d", c.Port)
    }
    
    if _, err := url.Parse(c.DatabaseURL); err != nil {
        return fmt.Errorf("invalid database URL: %w", err)
    }
    
    if c.Timeout < time.Second {
        return fmt.Errorf("timeout too short: %v", c.Timeout)
    }
    
    return nil
}

func main() {
    cfg, err := LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    if err := cfg.Validate(); err != nil {
        log.Fatalf("Invalid config: %v", err)
    }
    
    // Continue startup...
}
```

### Secret Management

```go
// DON'T: Secrets in code or config files
const apiKey = "secret123"  // BAD!

// DO: Secrets from environment/vault
apiKey := os.Getenv("API_KEY")

// DO: Use secret management services
// - AWS Secrets Manager
// - HashiCorp Vault
// - Kubernetes Secrets
// - Google Secret Manager
```

---

## 10.3 Structured Logging (`log/slog`)

### Basic slog Usage (Go 1.21+)

**Interview Question:** *"What are the benefits of structured logging?"*

```go
import "log/slog"

func main() {
    // Default text handler
    slog.Info("Starting server", "port", 8080)
    // Output: 2024/01/15 10:30:00 INFO Starting server port=8080
    
    // JSON handler for production
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    logger.Info("Starting server", "port", 8080)
    // Output: {"time":"2024-01-15T10:30:00Z","level":"INFO","msg":"Starting server","port":8080}
}
```

### Configuring Logger

```go
func NewLogger(level string, format string) *slog.Logger {
    var lvl slog.Level
    switch strings.ToLower(level) {
    case "debug":
        lvl = slog.LevelDebug
    case "info":
        lvl = slog.LevelInfo
    case "warn":
        lvl = slog.LevelWarn
    case "error":
        lvl = slog.LevelError
    default:
        lvl = slog.LevelInfo
    }
    
    opts := &slog.HandlerOptions{
        Level: lvl,
        AddSource: true,  // Add file:line
    }
    
    var handler slog.Handler
    switch format {
    case "json":
        handler = slog.NewJSONHandler(os.Stdout, opts)
    default:
        handler = slog.NewTextHandler(os.Stdout, opts)
    }
    
    return slog.New(handler)
}
```

### Logging with Context

```go
// Add attributes to all logs from this logger
logger := slog.Default().With(
    "service", "user-api",
    "version", "1.0.0",
)

// Context-aware logging
func handleRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) {
    requestID := ctx.Value("requestID").(string)
    
    logger := slog.Default().With(
        "request_id", requestID,
        "method", r.Method,
        "path", r.URL.Path,
    )
    
    logger.InfoContext(ctx, "Handling request")
    
    // ... process request
    
    logger.InfoContext(ctx, "Request completed", "status", 200, "duration_ms", 42)
}
```

### Log Groups

```go
slog.Info("User action",
    slog.Group("user",
        slog.String("id", "123"),
        slog.String("email", "user@example.com"),
    ),
    slog.Group("request",
        slog.String("method", "POST"),
        slog.String("path", "/api/users"),
    ),
)
// JSON: {"user":{"id":"123","email":"user@example.com"},"request":{"method":"POST","path":"/api/users"}}
```

### Custom Handler

```go
// Handler that filters sensitive data
type SanitizingHandler struct {
    handler slog.Handler
}

func (h *SanitizingHandler) Handle(ctx context.Context, r slog.Record) error {
    // Redact sensitive fields
    r.Attrs(func(a slog.Attr) bool {
        if a.Key == "password" || a.Key == "token" {
            a.Value = slog.StringValue("[REDACTED]")
        }
        return true
    })
    return h.handler.Handle(ctx, r)
}
```

---

## 10.4 Distributed Tracing (OpenTelemetry)

### Setting Up OpenTelemetry

**Interview Question:** *"What is distributed tracing and why is it important?"*

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func initTracer(ctx context.Context) (*trace.TracerProvider, error) {
    // Create OTLP exporter
    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint("otel-collector:4317"),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return nil, err
    }
    
    // Create resource with service info
    res, err := resource.Merge(
        resource.Default(),
        resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceName("my-service"),
            semconv.ServiceVersion("1.0.0"),
        ),
    )
    if err != nil {
        return nil, err
    }
    
    // Create TracerProvider
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(res),
        trace.WithSampler(trace.TraceIDRatioBased(0.1)), // Sample 10%
    )
    
    otel.SetTracerProvider(tp)
    
    return tp, nil
}
```

### Creating Spans

```go
import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("my-service")

func processOrder(ctx context.Context, orderID string) error {
    // Create span
    ctx, span := tracer.Start(ctx, "processOrder")
    defer span.End()
    
    // Add attributes
    span.SetAttributes(
        attribute.String("order.id", orderID),
        attribute.Int("order.items", 5),
    )
    
    // Nested span
    if err := validateOrder(ctx, orderID); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }
    
    return nil
}

func validateOrder(ctx context.Context, orderID string) error {
    ctx, span := tracer.Start(ctx, "validateOrder")
    defer span.End()
    
    // Validation logic...
    return nil
}
```

### HTTP Instrumentation

```go
import (
    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
    // Wrap handler for automatic tracing
    handler := http.HandlerFunc(myHandler)
    wrappedHandler := otelhttp.NewHandler(handler, "my-handler")
    
    http.Handle("/", wrappedHandler)
    http.ListenAndServe(":8080", nil)
}

// HTTP client with tracing
client := &http.Client{
    Transport: otelhttp.NewTransport(http.DefaultTransport),
}
```

### Trace Propagation

```go
// Propagation happens automatically via context
// When making outbound HTTP calls, trace context is added to headers

func callExternalService(ctx context.Context) error {
    req, _ := http.NewRequestWithContext(ctx, "GET", "http://other-service/api", nil)
    
    // Inject trace context into headers
    otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
    
    resp, err := client.Do(req)
    // ...
}

// Extract trace context from incoming request
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := otel.GetTextMapPropagator().Extract(r.Context(), 
        propagation.HeaderCarrier(r.Header))
    
    // Use ctx for creating spans - they'll be linked to the trace
}
```

---

## 10.5 Metrics (Prometheus)

### Setting Up Prometheus Metrics

**Interview Question:** *"What are the different types of Prometheus metrics?"*

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    // Counter - only increases
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    // Gauge - can increase or decrease
    activeConnections = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "active_connections",
            Help: "Number of active connections",
        },
    )
    
    // Histogram - distribution of values
    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets, // 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
        },
        []string{"method", "path"},
    )
    
    // Summary - similar to histogram but calculates quantiles
    responseSize = promauto.NewSummaryVec(
        prometheus.SummaryOpts{
            Name:       "http_response_size_bytes",
            Help:       "HTTP response size in bytes",
            Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
        },
        []string{"method", "path"},
    )
)
```

### Using Metrics

```go
func metricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Wrap response writer to capture status
        wrapped := &responseWriter{ResponseWriter: w, status: 200}
        
        next.ServeHTTP(wrapped, r)
        
        duration := time.Since(start).Seconds()
        
        // Record metrics
        httpRequestsTotal.WithLabelValues(
            r.Method, r.URL.Path, strconv.Itoa(wrapped.status),
        ).Inc()
        
        httpRequestDuration.WithLabelValues(
            r.Method, r.URL.Path,
        ).Observe(duration)
    })
}

func main() {
    mux := http.NewServeMux()
    
    // Application routes
    mux.HandleFunc("/api/users", handleUsers)
    
    // Metrics endpoint
    mux.Handle("/metrics", promhttp.Handler())
    
    // Apply middleware
    handler := metricsMiddleware(mux)
    
    http.ListenAndServe(":8080", handler)
}
```

### Metric Naming Conventions

```
# Format: namespace_subsystem_name_unit

# Good examples:
http_requests_total
http_request_duration_seconds
process_cpu_seconds_total
db_connections_open

# Bad examples:
requests          # Too vague
httpRequestCount  # Wrong case
request_time      # Missing unit
```

### Custom Collectors

```go
// For metrics that need to be fetched on-demand
type dbStatsCollector struct {
    db *sql.DB
    
    openDesc *prometheus.Desc
    idleDesc *prometheus.Desc
}

func NewDBStatsCollector(db *sql.DB) *dbStatsCollector {
    return &dbStatsCollector{
        db: db,
        openDesc: prometheus.NewDesc(
            "db_connections_open",
            "Number of open database connections",
            nil, nil,
        ),
        idleDesc: prometheus.NewDesc(
            "db_connections_idle",
            "Number of idle database connections",
            nil, nil,
        ),
    }
}

func (c *dbStatsCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- c.openDesc
    ch <- c.idleDesc
}

func (c *dbStatsCollector) Collect(ch chan<- prometheus.Metric) {
    stats := c.db.Stats()
    ch <- prometheus.MustNewConstMetric(c.openDesc, prometheus.GaugeValue, float64(stats.OpenConnections))
    ch <- prometheus.MustNewConstMetric(c.idleDesc, prometheus.GaugeValue, float64(stats.Idle))
}
```

---

## 10.6 Health Checks

### Kubernetes Probes

**Interview Question:** *"What are the different types of Kubernetes health probes?"*

```go
// Liveness: Is the process alive?
// Readiness: Is the service ready for traffic?
// Startup: Has the service finished starting?

type HealthChecker struct {
    db    *sql.DB
    redis *redis.Client
}

func (h *HealthChecker) LivenessHandler(w http.ResponseWriter, r *http.Request) {
    // Simple check - process is running
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func (h *HealthChecker) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    
    // Check dependencies
    if err := h.db.PingContext(ctx); err != nil {
        http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
        return
    }
    
    if err := h.redis.Ping(ctx).Err(); err != nil {
        http.Error(w, "Redis unavailable", http.StatusServiceUnavailable)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Ready"))
}

func (h *HealthChecker) StartupHandler(w http.ResponseWriter, r *http.Request) {
    // Check if initialization is complete
    if !isInitialized {
        http.Error(w, "Still starting", http.StatusServiceUnavailable)
        return
    }
    w.WriteHeader(http.StatusOK)
}
```

### Kubernetes Configuration

```yaml
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: app
    image: myapp:latest
    ports:
    - containerPort: 8080
    livenessProbe:
      httpGet:
        path: /healthz
        port: 8080
      initialDelaySeconds: 5
      periodSeconds: 10
      failureThreshold: 3
    readinessProbe:
      httpGet:
        path: /ready
        port: 8080
      initialDelaySeconds: 5
      periodSeconds: 5
      failureThreshold: 3
    startupProbe:
      httpGet:
        path: /startup
        port: 8080
      initialDelaySeconds: 0
      periodSeconds: 5
      failureThreshold: 30  # 30 * 5s = 150s max startup time
```

---

## 10.7 Graceful Shutdown

### Signal Handling

**Interview Question:** *"How do you implement graceful shutdown in Go?"*

```go
func main() {
    // Create context that cancels on SIGINT/SIGTERM
    ctx, cancel := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer cancel()
    
    // Initialize services
    db := initDB()
    server := initServer()
    
    // Start server in goroutine
    go func() {
        log.Println("Server starting on :8080")
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()
    
    // Wait for shutdown signal
    <-ctx.Done()
    log.Println("Shutdown signal received")
    
    // Create shutdown context with timeout
    shutdownCtx, shutdownCancel := context.WithTimeout(
        context.Background(), 30*time.Second)
    defer shutdownCancel()
    
    // Graceful shutdown sequence
    log.Println("Shutting down server...")
    if err := server.Shutdown(shutdownCtx); err != nil {
        log.Printf("Server shutdown error: %v", err)
    }
    
    log.Println("Closing database...")
    if err := db.Close(); err != nil {
        log.Printf("Database close error: %v", err)
    }
    
    log.Println("Shutdown complete")
}
```

### Complete Shutdown Pattern

```go
type Application struct {
    server *http.Server
    db     *sql.DB
    redis  *redis.Client
    wg     sync.WaitGroup
}

func (a *Application) Shutdown(ctx context.Context) error {
    var errs []error
    
    // 1. Stop accepting new requests
    if err := a.server.Shutdown(ctx); err != nil {
        errs = append(errs, fmt.Errorf("server shutdown: %w", err))
    }
    
    // 2. Wait for in-flight work
    done := make(chan struct{})
    go func() {
        a.wg.Wait()
        close(done)
    }()
    
    select {
    case <-done:
        // All work completed
    case <-ctx.Done():
        errs = append(errs, fmt.Errorf("timeout waiting for work"))
    }
    
    // 3. Close connections
    if err := a.db.Close(); err != nil {
        errs = append(errs, fmt.Errorf("db close: %w", err))
    }
    
    if err := a.redis.Close(); err != nil {
        errs = append(errs, fmt.Errorf("redis close: %w", err))
    }
    
    if len(errs) > 0 {
        return errors.Join(errs...)
    }
    return nil
}
```

---

## 10.8 Profile-Guided Optimization (PGO)

### What is PGO?

**Interview Question:** *"What is Profile-Guided Optimization and how do you use it?"*

```bash
# PGO uses runtime profiles to optimize builds
# Typical improvement: 2-7%

# 1. Build without PGO and deploy
go build -o myapp ./cmd/server

# 2. Collect CPU profile in production
# (run for representative workload period)
curl http://localhost:8080/debug/pprof/profile?seconds=30 > default.pgo

# 3. Rebuild with profile
go build -pgo=default.pgo -o myapp ./cmd/server

# 4. Or auto-detect (place default.pgo in main package)
go build -pgo=auto -o myapp ./cmd/server
```

### Collecting Profiles

```go
import _ "net/http/pprof"

func main() {
    // pprof endpoints automatically registered
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // Application server
    http.ListenAndServe(":8080", handler)
}
```

```bash
# Collect 30-second CPU profile
curl -o cpu.pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Rename for PGO
mv cpu.pprof default.pgo
```

---

## 10.9 Profiling in Production

### pprof Endpoints

```go
import (
    "net/http"
    _ "net/http/pprof"  // Registers handlers
)

func main() {
    // Separate port for profiling (security!)
    go func() {
        http.ListenAndServe("localhost:6060", nil)
    }()
}

// Available endpoints:
// /debug/pprof/           - Index page
// /debug/pprof/profile    - CPU profile (30s default)
// /debug/pprof/heap       - Heap memory profile
// /debug/pprof/goroutine  - Goroutine stacks
// /debug/pprof/block      - Block (contention) profile
// /debug/pprof/mutex      - Mutex contention profile
// /debug/pprof/trace      - Execution trace
```

### Using go tool pprof

```bash
# Interactive CPU profile analysis
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Heap analysis
go tool pprof http://localhost:6060/debug/pprof/heap

# Common commands in pprof:
# top      - Show top functions
# top10    - Top 10 functions
# list fn  - Show source for function
# web      - Generate SVG visualization
# pdf      - Generate PDF

# Save profile to file
curl -o cpu.pprof http://localhost:6060/debug/pprof/profile?seconds=30
go tool pprof cpu.pprof

# Compare profiles
go tool pprof -base old.pprof new.pprof
```

### Memory Profiling

```bash
# Current heap
go tool pprof http://localhost:6060/debug/pprof/heap

# Allocation since start
go tool pprof http://localhost:6060/debug/pprof/allocs

# In pprof:
(pprof) top        # Top memory consumers
(pprof) list main  # Source annotation
```

### Goroutine Analysis

```bash
# View goroutine stacks
curl http://localhost:6060/debug/pprof/goroutine?debug=2

# In pprof
go tool pprof http://localhost:6060/debug/pprof/goroutine
(pprof) top        # Functions with most goroutines
(pprof) traces     # Full stack traces
```

### Execution Tracing

```bash
# Collect 5-second trace
curl -o trace.out http://localhost:6060/debug/pprof/trace?seconds=5

# Analyze in browser
go tool trace trace.out
```

### Continuous Profiling

```go
// For production, use continuous profiling services:
// - Parca
// - Pyroscope
// - Datadog Continuous Profiler
// - Google Cloud Profiler

import "cloud.google.com/go/profiler"

func main() {
    if err := profiler.Start(profiler.Config{
        Service:        "my-service",
        ServiceVersion: "1.0.0",
        ProjectID:      "my-project",
    }); err != nil {
        log.Fatalf("Failed to start profiler: %v", err)
    }
}
```

---

## Interview Questions

### Beginner Level

1. **Q:** Why use multi-stage Docker builds for Go?
   **A:** Separate build (large Go toolchain) from runtime (small binary). Final image is minimal.

2. **Q:** What's the difference between liveness and readiness probes?
   **A:** Liveness: Is process alive? (restart if fails). Readiness: Can it accept traffic? (remove from load balancer if fails).

3. **Q:** How do you expose Prometheus metrics in Go?
   **A:** Use prometheus/client_golang, register metrics, serve `/metrics` endpoint with `promhttp.Handler()`.

### Intermediate Level

4. **Q:** How do you handle graceful shutdown?
   **A:** Catch SIGTERM/SIGINT with `signal.NotifyContext`, call `server.Shutdown()`, wait for in-flight requests, close connections.

5. **Q:** What are the Prometheus metric types?
   **A:** Counter (only increases), Gauge (up/down), Histogram (distribution with buckets), Summary (distribution with quantiles).

6. **Q:** What is distributed tracing?
   **A:** Tracking requests across services using trace IDs and spans. OpenTelemetry is the standard. Propagate context via headers.

### Advanced Level

7. **Q:** What is PGO and how does it improve performance?
   **A:** Profile-Guided Optimization uses runtime profiles to optimize builds. Compiler makes better inlining/branch decisions. 2-7% improvement typical.

8. **Q:** How would you debug a goroutine leak in production?
   **A:** Use `/debug/pprof/goroutine` endpoint, analyze with `go tool pprof`, look for growing goroutine count, identify blocking operations.

9. **Q:** Design a health check system for a microservice.
   **A:** Separate liveness (process ok) and readiness (dependencies ok) endpoints, check DB/cache connectivity with timeouts, don't fail readiness during startup.

---

## Summary

| Topic | Key Points |
|-------|------------|
| Docker | Multi-stage builds, distroless/scratch base, CGO_ENABLED=0, ldflags |
| Config | Environment variables (12-factor), validation at startup, secrets from vault |
| Logging | log/slog, structured JSON, context propagation, log levels |
| Tracing | OpenTelemetry, spans, context propagation, sampling |
| Metrics | Prometheus, counter/gauge/histogram, naming conventions, /metrics endpoint |
| Health | Liveness (alive?), Readiness (ready?), Startup (initialized?) |
| Shutdown | signal.NotifyContext, server.Shutdown(), wait for in-flight, close connections |
| PGO | Collect CPU profile, rebuild with -pgo flag, 2-7% improvement |
| Profiling | pprof endpoints, CPU/heap/goroutine/trace, continuous profiling |

**Next Phase:** [Phase 11 — Go Runtime Internals](../Phase_11/Phase_11.md)

