# 🚀 Advanced Go Roadmap: From Java Architect to Go Expert

**Target Audience:** Seasoned Software Architect (Java/OOP background).
**Goal:** Mastery of Go idioms, internals, memory management, and cloud-native patterns.
**Philosophy:** "Clear is better than clever." | "Share memory by communicating."
**Reference:** [The Go Programming Language Specification](https://go.dev/ref/spec)

---

## 📅 Phase 1: The Paradigm Shift & Core Syntax
**Goal:** Unlearn OOP "Best Practices" that are anti-patterns in Go.

### Topic 1.1: The Workspace & Philosophy
*   **Sub-topics:** `go.mod` (Modules), `gofmt`, Workspace semantics, Exported vs Unexported (Capitalization).
*   **Common Knowledge:** Go is opinionated. The compiler will fail on unused variables or imports.
*   **Interview Question:** "Why does Go forbid cyclic dependencies?"
    *   *Answer:* To ensure fast compilation times and clean architecture. Cycles imply tight coupling.
*   **Pitfall:** Trying to structure projects like Maven/Gradle (e.g., `src/main/java`).
*   **Real World:** Using `go work` (Go 1.18+) for local multi-module development in microservices repositories.

### Topic 1.2: Types, Structs & Composition (No Inheritance)
*   **Sub-topics:** `struct`, Type Embedding (Promoted fields), Pointer receivers vs. Value receivers.
*   **Deep Dive:** Go is not OOP. There is no `extends`. There is only composition.
*   **Interview Question:** "What is the difference between a Method defined on a Value receiver vs. a Pointer receiver?"
    *   *Answer:*
        1. **Performance:** Pointer avoids copying the struct.
        2. **Semantics:** Pointer allows mutating the struct.
        *   *Hint:* If you need to mutate, use Pointer. If the struct is large, use Pointer. Ideally, keep consistency.
*   **Pitfall:** Embedding a struct and assuming it works like polymorphism. Methods are promoted, but the "type" remains the wrapper.
*   **Insight:** Memory alignment in Structs matters. Ordering fields from largest (64-bit) to smallest (bool) can save RAM due to padding.

### Topic 1.3: Interfaces (Implicit Satisfaction)
*   **Sub-topics:** The `interface{}` (Any), Type Assertions, Type Switches, `nil` interface mechanics.
*   **Pro/Cons:**
    | Feature | Go | Java |
    | :--- | :--- | :--- |
    | **Implementation** | Implicit (Duck Typing) | Explicit (`implements`) |
    | **Definition** | Define where *used* (Consumer) | Define where *created* (Producer) |
*   **Interview Question:** "Explain the structure of an Interface value internally."
    *   *Details:* It is a 2-word header: `(Type Pointer, Value Pointer)`. An interface is only `nil` if BOTH are `nil`.
*   **Real World:** The "Accept Interfaces, Return Structs" pattern allows for easier mocking and testing without bloated abstract classes.

---

## 🧠 Phase 2: Memory Management & Data Structures
**Goal:** Understanding the Heap, Stack, and the Cost of Abstractions.

### Topic 2.1: Slices vs. Arrays
*   **Sub-topics:** Slice headers (`ptr`, `len`, `cap`), Slicing operations, `make` vs `new`.
*   **Deep Dive:** A Slice is a *descriptor* (lightweight view) of an underlying Array.
*   **Pitfall:** **The Re-Slicing Memory Leak.** Keeping a small slice of a massive array in memory prevents the garbage collector from reclaiming the massive array.
*   **Interview Question:** "How does `append` work internally?"
    *   *Answer:* Checks `cap`. If `len + 1 > cap`, it allocates a new array (usually 2x size), copies data, and returns the new slice header.
*   **Performance Tip:** Always pre-allocate slices if size is known: `make([]T, 0, 100)` to avoid resize allocations.

### Topic 2.2: Pointers, Stack & Heap (Escape Analysis)
*   **Sub-topics:** `&` and `*`, Escape Analysis, Inlining.
*   **Core Concept:** Go prefers Stack allocation (fast, self-cleaning). Variables move to Heap (GC pressure) only if they "escape" the function scope.
*   **Command:** `go build -gcflags="-m"` reveals escape analysis decisions.
*   **Interview Question:** "Is it always faster to pass a pointer?"
    *   *Answer:* **No.** Dereferencing pointers causes cache misses. Passing small structs by value is often faster because they stay on the stack and hit CPU cache lines.
*   **Real World:** High-frequency trading apps optimize struct sizes to fit in cache lines and strictly avoid pointer chasing in hot paths.

### Topic 2.3: Garbage Collection (GC)
*   **Sub-topics:** Tricolor Mark-and-Sweep, Write Barriers, `GOGC`, `GOMEMLIMIT`.
*   **Insight:** Go's GC is optimized for **low latency** (Stop-the-world < 500 microseconds), not throughput.
*   **Pitfall:** Generating excessive garbage (short-lived objects) forces the GC to run frequently, stealing CPU cycles from your app.

---

## ⚡ Phase 3: Concurrency & Asynchronous Programming
**Goal:** Mastering CSP (Communicating Sequential Processes) and the GMP Scheduler.

### Topic 3.1: Goroutines & GMP Architecture
*   **Sub-topics:** Kernel Threads vs. User Threads, M:N Scheduling, Context Switching costs.
*   **Architecture:**
    *   **G:** Goroutine (Stack starts at 2KB).
    *   **M:** Machine (OS Thread).
    *   **P:** Processor (Logical Context, usually = GOMAXPROCS).
*   **Interview Question:** "What happens if a Goroutine blocks on a System Call?"
    *   *Answer:* The P detaches from the M (which is blocked) and grabs a new M to keep executing other Gs.
*   **Pitfall:** **Goroutine Leaks.** Starting a goroutine that never exits (e.g., waiting on a nil channel). This is a permanent memory leak.

### Topic 3.2: Channels & Synchronization
*   **Sub-topics:** Buffered vs. Unbuffered, `select`, `close`, `sync.Mutex`, `sync.WaitGroup`, `sync.Map`.
*   **Best Practice:** Use Channels for data flow/ownership transfer. Use Mutex for state coherence.
*   **Interview Question:** "How do you implement a timeout using `select`?"
    ```go
    select {
    case res := <-ch:
        handle(res)
    case <-time.After(5 * time.Second):
        return errors.New("timeout")
    }
    ```
*   **Real World:** Worker Pools. Using a buffered channel as a semaphore to limit concurrent HTTP requests to an external API.

### Topic 3.3: The Context Package
*   **Sub-topics:** `context.Background`, `WithCancel`, `WithTimeout`, Value propagation.
*   **Golden Rule:** Context controls the lifecycle of a request.
*   **Pitfall:** Storing `context.Context` in a Struct.
    *   *Correct:* Pass `ctx` as the first argument to functions/methods.
    *   *Why?* Context is request-scoped, Structs are often instance-scoped.

---

## 🧪 Phase 4: Testing & Reliability (TDD/BDD)
**Goal:** Building bulletproof software.

### Topic 4.1: The Standard `testing` Library
*   **Sub-topics:** `TestMain`, Table-Driven Tests, Subtests (`t.Run`), Helpers (`t.Helper`).
*   **Standard:** **Table-Driven Tests** are the Go industry standard.
    ```go
    tests := []struct{ name, input, want }{...}
    for _, tt := range tests { t.Run(tt.name, ...) }
    ```
*   **Pitfall:** Using global variables in tests preventing parallel execution (`t.Parallel()`).

### Topic 4.2: Fuzzing & Benchmarking
*   **Sub-topics:** `func FuzzX(f *testing.F)`, `func BenchmarkX(b *testing.B)`, `benchstat`.
*   **Insight:** Benchmarks are useless without analyzing allocations (`b.ReportAllocs()`).
*   **Real World:** Use Fuzzing to find edge cases in parsers (JSON, YAML) or validation logic where random inputs cause Panics.

### Topic 4.3: Mocking vs. Stubs
*   **Libraries:** `testify/mock`, `gomock`.
*   **Philosophy:** Mock interfaces, not structs.
*   **Tip:** If an interface is too big to mock easily, the interface is arguably too big.

---

## 🏛️ Phase 5: Cloud Native Architecture & Patterns
**Goal:** System Design using Go.

### Topic 5.1: Project Layout & Clean Architecture
*   **Reference:** [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
*   **Structure:**
    *   `/cmd`: Entry points (main.go).
    *   `/internal`: Private application code (compiler enforced).
    *   `/pkg`: Library code ok to be imported by others.
*   **Pattern:** **Hexagonal Architecture (Ports & Adapters)**.
    *   *Core:* Domain Logic (Pure Go, no SQL/HTTP imports).
    *   *Adapters:* SQL Repositories, HTTP Handlers implement Core Interfaces.

### Topic 5.2: Dependency Injection
*   **Sub-topics:** Manual wiring (recommended), Google Wire (compile-time), Uber Dig (reflection-based).
*   **Interview Question:** "Why do Go engineers dislike Reflection-based DI frameworks (like Spring)?"
    *   *Answer:* Go values compile-time safety and readability ("Magic is bad"). Reflection is slow and hides dependency graphs.

### Topic 5.3: Event-Driven & Microservices
*   **Libraries:** Sarama (Kafka), NATS JetStream, Watermill (Generic Pub/Sub).
*   **Pattern: The Outbox Pattern.**
    *   *Problem:* How to save to DB and publish to Kafka atomically?
    *   *Solution:* Write entity AND event to DB in one transaction. Use a Go poller to read event table and publish.
*   **Pitfall:** Using the default `http.Client` without timeouts. It has NO timeout by default and will hang your production system indefinitely.

---

## 🎓 Recommended Learning Resources

1.  **Book:** "100 Go Mistakes and How to Avoid Them" by Teiva Harsanyi (Crucial for Java devs).
2.  **Book:** "Concurrency in Go" by Katherine Cox-Buday.
3.  **Site:** [Uber Go Style Guide](https://github.com/uber-go/guide) (The industry standard for code reviews).
4.  **Site:** [Go by Example](https://gobyexample.com) (Quick syntax ref).
5.  **Course:** "Ultimate Go" by Bill Kennedy (Ardan Labs) - The gold standard for internals.