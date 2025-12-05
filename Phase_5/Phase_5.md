# ⚡ Phase 5: Concurrency & The Scheduler

[← Back to Main Roadmap](../README.md) | [← Previous: Phase 4](../Phase_4/Phase_4.md)

---

**Objective:** Master goroutines, channels, and synchronization as orchestration tools, not parallelism primitives.

**Reference:** [Go Memory Model](https://go.dev/ref/mem), [Effective Go - Concurrency](https://go.dev/doc/effective_go#concurrency)

**Prerequisites:** Phase 0-4

**Estimated Duration:** 3-4 weeks

---

## 📋 Table of Contents

1. [Concurrency vs. Parallelism](#51-concurrency-vs-parallelism)
2. [Goroutines](#52-goroutines)
3. [The GMP Scheduler Model](#53-the-gmp-scheduler-model)
4. [Work Stealing & Preemption](#54-work-stealing--preemption)
5. [Channels](#55-channels)
6. [The `select` Statement](#56-the-select-statement)
7. [Synchronization Primitives](#57-synchronization-primitives-sync-package)
8. [Atomic Operations](#58-atomic-operations-syncatomic)
9. [Context Package](#59-context-package)
10. [Concurrency Patterns](#510-concurrency-patterns)
11. [Error Handling & Race Detection](#511-error-handling--race-detection)
12. [Interview Questions](#interview-questions)

---

## 5.1 Concurrency vs. Parallelism

### Definitions

**Interview Question:** *"What is the difference between concurrency and parallelism?"*

```
Concurrency: Dealing with multiple things at once (STRUCTURE)
- Managing multiple tasks that can make progress
- About composition and design
- Can happen on single CPU

Parallelism: Doing multiple things at once (EXECUTION)
- Actually executing multiple tasks simultaneously
- About performance and speed
- Requires multiple CPUs
```

**Rob Pike's definition:** "Concurrency is about dealing with lots of things at once. Parallelism is about doing lots of things at once."

### Go's Model: CSP

**Interview Question:** *"What is CSP and how does Go implement it?"*

CSP = Communicating Sequential Processes (Tony Hoare, 1978)

```go
// Go's philosophy:
// "Don't communicate by sharing memory; share memory by communicating."

// BAD: Shared memory with locks
var counter int
var mu sync.Mutex

func increment() {
    mu.Lock()
    counter++
    mu.Unlock()
}

// GOOD: Communicate via channels
func counter(ch chan int) {
    count := 0
    for delta := range ch {
        count += delta
    }
}
```

### Example: Concurrency in Action

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    // Concurrent - structure
    go processA()
    go processB()
    go processC()
    
    // These MAY run in parallel if GOMAXPROCS > 1
    // But the code is about STRUCTURE, not parallelism
    
    time.Sleep(time.Second)
}

func processA() { fmt.Println("A") }
func processB() { fmt.Println("B") }
func processC() { fmt.Println("C") }
```

---

## 5.2 Goroutines

### Creating Goroutines

**Interview Question:** *"What is a goroutine? How does it differ from an OS thread?"*

```go
// Create goroutine with 'go' keyword
go myFunction()

// Anonymous goroutine
go func() {
    fmt.Println("Running in goroutine")
}()

// Goroutine with arguments
go func(msg string) {
    fmt.Println(msg)
}("hello")
```

### Goroutine vs. Thread Comparison

| Aspect | Goroutine | OS Thread |
|--------|-----------|-----------|
| Memory | ~2KB initial stack | ~1MB stack |
| Creation time | Microseconds | Milliseconds |
| Context switch | ~200ns (user space) | ~1μs (kernel) |
| Number practical | Millions | Thousands |
| Scheduling | Go runtime (M:N) | OS kernel (1:1) |

### Goroutine Lifecycle

```go
func main() {
    // Goroutine starts when 'go' is called
    go worker()
    
    // Goroutine ends when:
    // 1. Function returns
    // 2. Runtime terminates (main exits)
    // 3. panic (if not recovered)
    
    // WARNING: main doesn't wait for goroutines!
    time.Sleep(time.Second)  // Wait (bad practice)
}

func worker() {
    // Runs until return
    fmt.Println("Working...")
}
```

### Stack Growth

**Interview Question:** *"How does Go manage goroutine stack size?"*

```go
// Goroutine stack:
// - Starts at 2KB (configurable with GOSTACK)
// - Grows dynamically as needed
// - Maximum ~1GB
// - Uses contiguous stack (since Go 1.4)

// Stack growth mechanism:
// 1. Function prologue checks stack space
// 2. If insufficient, runtime allocates larger stack
// 3. Copies entire stack to new location
// 4. Updates all pointers
```

### Closure Capture Gotcha

**Interview Question:** *"What is the goroutine closure capture bug?"*

```go
// BUG (Pre-Go 1.22): All goroutines see same variable
for i := 0; i < 3; i++ {
    go func() {
        fmt.Println(i)  // Captures reference to i
    }()
}
// Output (likely): 3, 3, 3

// FIX 1: Pass as argument
for i := 0; i < 3; i++ {
    go func(n int) {
        fmt.Println(n)
    }(i)  // i copied to n
}

// FIX 2: Shadow variable
for i := 0; i < 3; i++ {
    i := i  // Shadow with new variable
    go func() {
        fmt.Println(i)
    }()
}

// Go 1.22+: Fixed! Loop variables are per-iteration
for i := 0; i < 3; i++ {
    go func() {
        fmt.Println(i)  // Works correctly
    }()
}
```

### Example: Goroutine Patterns

```go
package main

import (
    "fmt"
    "sync"
)

func main() {
    var wg sync.WaitGroup
    
    // Pattern: Launch workers
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            fmt.Printf("Worker %d starting\n", id)
            // Do work...
            fmt.Printf("Worker %d done\n", id)
        }(i)
    }
    
    // Wait for all goroutines to complete
    wg.Wait()
    fmt.Println("All workers completed")
}
```

---

## 5.3 The GMP Scheduler Model

### Components

**Interview Question:** *"Explain the GMP model in Go's scheduler."*

```
G - Goroutine
    - The unit of concurrent work
    - Contains stack, instruction pointer, state
    - Lightweight (~2KB)

M - Machine (OS Thread)
    - Executes goroutines
    - Managed by OS
    - Limited pool (GOMAXPROCS by default)

P - Processor (Logical Processor)
    - Context for executing Go code
    - Contains local run queue (LRQ)
    - Memory cache (mcache)
    - GOMAXPROCS determines count
```

```
Visual representation:

    +--------+     +--------+     +--------+
    |   G    |     |   G    |     |   G    |  (Goroutines)
    +--------+     +--------+     +--------+
         |              |              |
    +--------+     +--------+     +--------+
    |   P    |     |   P    |     |   P    |  (Processors)
    | [LRQ]  |     | [LRQ]  |     | [LRQ]  |
    +--------+     +--------+     +--------+
         |              |              |
    +--------+     +--------+     +--------+
    |   M    |     |   M    |     |   M    |  (OS Threads)
    +--------+     +--------+     +--------+
         |              |              |
    +------------------------------------------+
    |              Operating System             |
    +------------------------------------------+
```

### GOMAXPROCS

```go
import "runtime"

// Get current value
n := runtime.GOMAXPROCS(0)  // 0 = query without changing

// Set value
runtime.GOMAXPROCS(4)  // Limit to 4 P's

// Default: number of logical CPUs
runtime.GOMAXPROCS(runtime.NumCPU())
```

**Key insight:** GOMAXPROCS limits parallelism, not concurrency. You can have millions of goroutines with GOMAXPROCS=1.

### Scheduling Events

**Interview Question:** *"When does the Go scheduler switch between goroutines?"*

```go
// Goroutine yields to scheduler on:

// 1. Channel operations
ch <- value  // Send
<-ch         // Receive

// 2. Network I/O
conn.Read()
http.Get()

// 3. Blocking syscalls
file.Read()

// 4. time.Sleep
time.Sleep(time.Second)

// 5. Manual yield
runtime.Gosched()

// 6. Garbage collection

// 7. sync operations
mutex.Lock()
wg.Wait()

// 8. Function calls (preemption points)
```

### Example: Visualizing Scheduling

```go
package main

import (
    "fmt"
    "runtime"
)

func main() {
    runtime.GOMAXPROCS(1)  // Force single thread
    
    go func() {
        for i := 0; i < 5; i++ {
            fmt.Println("Goroutine A:", i)
            runtime.Gosched()  // Yield to scheduler
        }
    }()
    
    go func() {
        for i := 0; i < 5; i++ {
            fmt.Println("Goroutine B:", i)
            runtime.Gosched()
        }
    }()
    
    // Wait
    var input string
    fmt.Scanln(&input)
}
```

---

## 5.4 Work Stealing & Preemption

### Work Stealing

**Interview Question:** *"What is work stealing in Go's scheduler?"*

```
Work Stealing Algorithm:
1. P's local run queue (LRQ) is empty
2. P tries to steal from:
   a. Global run queue
   b. Network poller
   c. Other P's LRQ (steal half)

Benefits:
- Load balancing without centralization
- Good cache locality (prefer local queue)
- Scales well with CPU count
```

```go
// Visualization of stealing:
// P1 has goroutines: [G1, G2, G3, G4]
// P2 has goroutines: []  (empty)

// P2 steals half from P1:
// P1: [G1, G2]
// P2: [G3, G4]
```

### Preemption

**Interview Question:** *"How does Go preempt long-running goroutines?"*

```go
// Historical: Cooperative preemption (Go < 1.14)
// - Only at function calls
// - Tight loops could starve other goroutines

// Problem case:
func tight() {
    for {
        // No function calls = no preemption!
        // Other goroutines starve
    }
}

// Modern: Asynchronous preemption (Go 1.14+)
// - Signal-based (SIGURG)
// - Can interrupt at safe points
// - ~10ms quantum

// Now this doesn't starve others:
func tight() {
    for {
        // Runtime sends signal to interrupt
    }
}
```

### Example: Preemption Behavior

```go
package main

import (
    "fmt"
    "runtime"
    "time"
)

func main() {
    runtime.GOMAXPROCS(1)
    
    // Tight loop goroutine
    go func() {
        count := 0
        for {
            count++
            if count%1000000000 == 0 {
                fmt.Println("Still running...")
            }
        }
    }()
    
    // This will run thanks to async preemption
    go func() {
        for i := 0; i < 10; i++ {
            fmt.Println("Other goroutine:", i)
            time.Sleep(100 * time.Millisecond)
        }
    }()
    
    time.Sleep(2 * time.Second)
}
```

---

## 5.5 Channels

### Channel Fundamentals

**Interview Question:** *"What are the different types of channels and their behaviors?"*

```go
// Unbuffered channel (synchronous)
ch := make(chan int)

// Buffered channel (asynchronous up to capacity)
ch := make(chan int, 10)

// Directional channels (for function signatures)
func send(ch chan<- int) { ch <- 42 }      // Send-only
func recv(ch <-chan int) { v := <-ch }     // Receive-only
```

### Channel Operations

```go
ch := make(chan int)

// Send
ch <- 42

// Receive
v := <-ch

// Receive with ok (closed channel detection)
v, ok := <-ch
if !ok {
    fmt.Println("Channel closed")
}

// Close (only sender should close)
close(ch)

// Length and capacity
len(ch)  // Number of elements in buffer
cap(ch)  // Buffer capacity
```

### Channel Axioms (Critical for Interviews!)

**Interview Question:** *"What happens when you send/receive/close on a nil or closed channel?"*

| Operation | Nil Channel | Closed Channel |
|-----------|-------------|----------------|
| **Send** | Block forever | **PANIC** |
| **Receive** | Block forever | Returns zero value, ok=false |
| **Close** | **PANIC** | **PANIC** |

```go
// Demonstration
var nilCh chan int
// nilCh <- 1       // Blocks forever
// <-nilCh          // Blocks forever
// close(nilCh)     // PANIC!

closedCh := make(chan int)
close(closedCh)
// closedCh <- 1    // PANIC!
v, ok := <-closedCh // v=0, ok=false (safe)
// close(closedCh)  // PANIC!
```

### Unbuffered vs. Buffered

**Interview Question:** *"When should you use buffered vs. unbuffered channels?"*

```go
// Unbuffered: Synchronization point (rendezvous)
ch := make(chan int)
// Send blocks until receive (and vice versa)

// Use when:
// - Guaranteed handoff required
// - Synchronization is the goal
// - Backpressure needed

// Buffered: Decouples sender and receiver
ch := make(chan int, 100)
// Send blocks only when buffer is full

// Use when:
// - Known, bounded amount of work
// - Smoothing bursts
// - Implementing semaphore
```

### Range Over Channels

```go
ch := make(chan int)

go func() {
    for i := 0; i < 5; i++ {
        ch <- i
    }
    close(ch)  // Must close for range to exit
}()

for v := range ch {
    fmt.Println(v)
}
// Loop exits when channel is closed and empty
```

### Example: Channel Patterns

```go
package main

import (
    "fmt"
    "time"
)

// Pattern 1: Done channel for cancellation
func worker(done <-chan struct{}) {
    for {
        select {
        case <-done:
            fmt.Println("Worker stopped")
            return
        default:
            fmt.Println("Working...")
            time.Sleep(500 * time.Millisecond)
        }
    }
}

// Pattern 2: Result channel
func compute(n int) <-chan int {
    result := make(chan int)
    go func() {
        defer close(result)
        // Expensive computation
        time.Sleep(time.Second)
        result <- n * n
    }()
    return result
}

func main() {
    // Done channel
    done := make(chan struct{})
    go worker(done)
    time.Sleep(2 * time.Second)
    close(done)
    
    // Result channel
    resultCh := compute(5)
    result := <-resultCh
    fmt.Println("Result:", result)
}
```

---

## 5.6 The `select` Statement

### Basics

**Interview Question:** *"How does `select` work with multiple channels?"*

```go
select {
case v := <-ch1:
    fmt.Println("Received from ch1:", v)
case ch2 <- x:
    fmt.Println("Sent to ch2")
case <-time.After(time.Second):
    fmt.Println("Timeout")
default:
    fmt.Println("No communication ready")
}
```

**Behavior:**
- Evaluates all cases simultaneously
- If multiple ready: random selection (fairness)
- If none ready: blocks until one is ready
- `default`: makes select non-blocking

### Common Patterns

#### Timeout

```go
select {
case result := <-ch:
    process(result)
case <-time.After(5 * time.Second):
    return errors.New("timeout")
}
```

#### Non-blocking Send/Receive

```go
// Non-blocking receive
select {
case v := <-ch:
    fmt.Println("Received:", v)
default:
    fmt.Println("No value available")
}

// Non-blocking send
select {
case ch <- v:
    fmt.Println("Sent")
default:
    fmt.Println("Channel full, dropping")
}
```

#### Done Channel Pattern

```go
func worker(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            // Do work
        }
    }
}
```

### Example: Select Multiplexing

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    ch1 := make(chan string)
    ch2 := make(chan string)
    
    go func() {
        time.Sleep(1 * time.Second)
        ch1 <- "one"
    }()
    
    go func() {
        time.Sleep(2 * time.Second)
        ch2 <- "two"
    }()
    
    // Receive from whichever is ready first
    for i := 0; i < 2; i++ {
        select {
        case msg1 := <-ch1:
            fmt.Println("Received:", msg1)
        case msg2 := <-ch2:
            fmt.Println("Received:", msg2)
        }
    }
}
```

---

## 5.7 Synchronization Primitives (`sync` Package)

### Mutex

**Interview Question:** *"When should you use a mutex vs. a channel?"*

```go
import "sync"

type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

func (c *SafeCounter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.count
}
```

**Mutex vs. Channel:**
- Mutex: Protecting shared state, simple critical sections
- Channel: Communication, passing ownership, coordination

### RWMutex

```go
type Cache struct {
    mu   sync.RWMutex
    data map[string]string
}

func (c *Cache) Get(key string) (string, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    v, ok := c.data[key]
    return v, ok
}

func (c *Cache) Set(key, value string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = value
}
```

**Use when:** Read-heavy workloads (many readers, few writers)

### WaitGroup

**Interview Question:** *"How do you wait for multiple goroutines to complete?"*

```go
var wg sync.WaitGroup

for i := 0; i < 10; i++ {
    wg.Add(1)  // MUST be called before goroutine
    go func(id int) {
        defer wg.Done()
        process(id)
    }(i)
}

wg.Wait()  // Blocks until counter reaches 0
```

**Common bug:**
```go
// BUG: Add called inside goroutine (race condition)
for i := 0; i < 10; i++ {
    go func(id int) {
        wg.Add(1)  // WRONG! Race with Wait()
        defer wg.Done()
        process(id)
    }(i)
}
wg.Wait()  // May not wait for all!
```

### Once

```go
var once sync.Once
var instance *Database

func GetDatabase() *Database {
    once.Do(func() {
        instance = connectToDatabase()
    })
    return instance
}
```

**Guaranteed:** Function is executed exactly once, even with concurrent calls.

### Cond

```go
var mu sync.Mutex
var cond = sync.NewCond(&mu)
var ready bool

// Waiter
func wait() {
    mu.Lock()
    for !ready {
        cond.Wait()  // Releases lock, waits, reacquires
    }
    mu.Unlock()
}

// Signaler
func signal() {
    mu.Lock()
    ready = true
    cond.Broadcast()  // Wake all waiters
    mu.Unlock()
}
```

### Pool

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 4096)
    },
}

func process() {
    buf := bufferPool.Get().([]byte)
    defer bufferPool.Put(buf)
    
    // Use buffer...
}
```

### sync.Map

```go
var m sync.Map

// Store
m.Store("key", "value")

// Load
if v, ok := m.Load("key"); ok {
    fmt.Println(v)
}

// LoadOrStore
actual, loaded := m.LoadOrStore("key", "default")

// Delete
m.Delete("key")

// Range
m.Range(func(key, value interface{}) bool {
    fmt.Println(key, value)
    return true  // Continue iteration
})
```

**When to use:** Many goroutines, mostly disjoint keys, read-heavy. Otherwise use regular map + mutex.

### Example: Synchronization Patterns

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func main() {
    // Pattern: Worker pool with WaitGroup
    var wg sync.WaitGroup
    jobs := make(chan int, 100)
    
    // Start workers
    for w := 0; w < 3; w++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for job := range jobs {
                fmt.Printf("Worker %d processing job %d\n", id, job)
                time.Sleep(100 * time.Millisecond)
            }
        }(w)
    }
    
    // Send jobs
    for j := 0; j < 10; j++ {
        jobs <- j
    }
    close(jobs)
    
    // Wait for completion
    wg.Wait()
    fmt.Println("All jobs completed")
}
```

---

## 5.8 Atomic Operations (`sync/atomic`)

### Atomic Types (Go 1.19+)

**Interview Question:** *"When should you use atomic operations vs. mutexes?"*

```go
import "sync/atomic"

var counter atomic.Int64

func main() {
    counter.Store(0)
    
    var wg sync.WaitGroup
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counter.Add(1)
        }()
    }
    
    wg.Wait()
    fmt.Println(counter.Load())  // 1000
}
```

### Atomic Operations

```go
var x atomic.Int64

x.Store(42)          // Set value
v := x.Load()        // Get value
x.Add(10)            // Add and return new
old := x.Swap(100)   // Set and return old
x.CompareAndSwap(100, 200)  // CAS
```

### Compare-And-Swap (CAS)

**Interview Question:** *"Explain Compare-And-Swap. What are its use cases?"*

```go
// CAS: Only update if current value matches expected
var state atomic.Int32

func tryUpdate(expected, new int32) bool {
    return state.CompareAndSwap(expected, new)
}

// Use case: Lock-free counter increment
func increment() {
    for {
        old := state.Load()
        if state.CompareAndSwap(old, old+1) {
            return  // Success
        }
        // CAS failed, retry
    }
}
```

### Atomic Pointer

```go
type Config struct {
    Value string
}

var configPtr atomic.Pointer[Config]

func updateConfig(c *Config) {
    configPtr.Store(c)
}

func getConfig() *Config {
    return configPtr.Load()
}
```

### Atomic vs. Mutex

| Atomic | Mutex |
|--------|-------|
| Single value operations | Complex critical sections |
| Lock-free | Locks |
| Faster for simple ops | More flexible |
| Limited to specific ops | Any code |

---

## 5.9 Context Package

### Context Fundamentals

**Interview Question:** *"What is the context package used for? What are the rules for context usage?"*

```go
import "context"

// Context carries:
// 1. Cancellation signals
// 2. Deadlines/timeouts
// 3. Request-scoped values

// Root contexts
ctx := context.Background()  // Main programs
ctx := context.TODO()        // Placeholder
```

### Cancellation

```go
// Manual cancellation
ctx, cancel := context.WithCancel(context.Background())
defer cancel()  // Always defer cancel!

go func(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            fmt.Println("Cancelled:", ctx.Err())
            return
        default:
            // Work...
        }
    }
}(ctx)

time.Sleep(time.Second)
cancel()  // Trigger cancellation
```

### Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

select {
case <-time.After(10 * time.Second):
    fmt.Println("Operation completed")
case <-ctx.Done():
    fmt.Println("Timeout:", ctx.Err())
}
```

### Deadline

```go
deadline := time.Now().Add(30 * time.Second)
ctx, cancel := context.WithDeadline(context.Background(), deadline)
defer cancel()
```

### Values (Use Sparingly!)

```go
type contextKey string

const userIDKey contextKey = "userID"

func WithUserID(ctx context.Context, userID string) context.Context {
    return context.WithValue(ctx, userIDKey, userID)
}

func GetUserID(ctx context.Context) (string, bool) {
    userID, ok := ctx.Value(userIDKey).(string)
    return userID, ok
}
```

### Context Rules

**Interview Question:** *"What are the best practices for using context?"*

```go
// 1. Context is first parameter
func ProcessRequest(ctx context.Context, req *Request) error

// 2. Never store context in struct
type Handler struct {
    // ctx context.Context  // BAD!
}

// 3. Pass context down the call chain
func handler(ctx context.Context) {
    processA(ctx)
    processB(ctx)
}

// 4. Always defer cancel
ctx, cancel := context.WithTimeout(parent, timeout)
defer cancel()  // Releases resources

// 5. Use context-aware functions
data, err := db.QueryContext(ctx, query)
resp, err := client.Do(req.WithContext(ctx))
```

### Example: Complete Context Usage

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func main() {
    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    
    // Pass to worker
    result := make(chan string)
    go worker(ctx, result)
    
    select {
    case r := <-result:
        fmt.Println("Result:", r)
    case <-ctx.Done():
        fmt.Println("Timeout:", ctx.Err())
    }
}

func worker(ctx context.Context, result chan<- string) {
    // Simulate work that respects cancellation
    select {
    case <-time.After(2 * time.Second):
        result <- "completed"
    case <-ctx.Done():
        return
    }
}
```

---

## 5.10 Concurrency Patterns

### Worker Pool

**Interview Question:** *"Implement a worker pool pattern in Go."*

```go
func workerPool(numWorkers int, jobs <-chan Job) <-chan Result {
    results := make(chan Result)
    
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for job := range jobs {
                results <- process(job)
            }
        }(i)
    }
    
    // Close results when all workers done
    go func() {
        wg.Wait()
        close(results)
    }()
    
    return results
}
```

### Fan-Out / Fan-In

```go
// Fan-Out: One channel to many goroutines
func fanOut(input <-chan int, numWorkers int) []<-chan int {
    outputs := make([]<-chan int, numWorkers)
    for i := 0; i < numWorkers; i++ {
        outputs[i] = worker(input)
    }
    return outputs
}

// Fan-In: Many channels to one
func fanIn(inputs ...<-chan int) <-chan int {
    output := make(chan int)
    var wg sync.WaitGroup
    
    for _, in := range inputs {
        wg.Add(1)
        go func(ch <-chan int) {
            defer wg.Done()
            for v := range ch {
                output <- v
            }
        }(in)
    }
    
    go func() {
        wg.Wait()
        close(output)
    }()
    
    return output
}
```

### Pipeline

```go
func generator(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, n := range nums {
            out <- n
        }
    }()
    return out
}

func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            out <- n * n
        }
    }()
    return out
}

func main() {
    // Pipeline: generator -> square -> print
    for n := range square(generator(1, 2, 3, 4)) {
        fmt.Println(n)
    }
}
```

### Semaphore

```go
// Buffered channel as semaphore
sem := make(chan struct{}, maxConcurrency)

for _, item := range items {
    sem <- struct{}{}  // Acquire
    go func(item Item) {
        defer func() { <-sem }()  // Release
        process(item)
    }(item)
}
```

### Rate Limiting

```go
// Simple rate limiter with ticker
limiter := time.NewTicker(100 * time.Millisecond)
defer limiter.Stop()

for _, item := range items {
    <-limiter.C  // Wait for tick
    go process(item)
}

// Token bucket
type RateLimiter struct {
    tokens chan struct{}
}

func NewRateLimiter(rate int) *RateLimiter {
    rl := &RateLimiter{
        tokens: make(chan struct{}, rate),
    }
    go func() {
        ticker := time.NewTicker(time.Second / time.Duration(rate))
        for range ticker.C {
            select {
            case rl.tokens <- struct{}{}:
            default:
            }
        }
    }()
    return rl
}

func (rl *RateLimiter) Wait() {
    <-rl.tokens
}
```

### Graceful Shutdown

**Interview Question:** *"How do you implement graceful shutdown in Go?"*

```go
func main() {
    ctx, cancel := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer cancel()
    
    // Start server
    server := &http.Server{Addr: ":8080"}
    
    go func() {
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()
    
    // Wait for signal
    <-ctx.Done()
    
    // Graceful shutdown with timeout
    shutdownCtx, shutdownCancel := context.WithTimeout(
        context.Background(), 30*time.Second)
    defer shutdownCancel()
    
    if err := server.Shutdown(shutdownCtx); err != nil {
        log.Printf("Shutdown error: %v", err)
    }
}
```

---

## 5.11 Error Handling & Race Detection

### errgroup

**Interview Question:** *"How do you handle errors from multiple goroutines?"*

```go
import "golang.org/x/sync/errgroup"

func fetchAll(ctx context.Context, urls []string) error {
    g, ctx := errgroup.WithContext(ctx)
    
    for _, url := range urls {
        url := url  // Capture
        g.Go(func() error {
            return fetch(ctx, url)
        })
    }
    
    // Wait for all, return first error
    return g.Wait()
}
```

### Panic Recovery in Goroutines

```go
func safeGo(f func()) {
    go func() {
        defer func() {
            if r := recover(); r != nil {
                log.Printf("Goroutine panic: %v\n%s", r, debug.Stack())
            }
        }()
        f()
    }()
}
```

### Race Detector

**Interview Question:** *"How do you detect data races in Go?"*

```bash
# Build with race detector
go build -race myapp.go

# Test with race detector
go test -race ./...

# Run with race detector
go run -race main.go
```

```go
// Race example:
var counter int

func main() {
    for i := 0; i < 1000; i++ {
        go func() {
            counter++  // DATA RACE!
        }()
    }
}

// Race detector output:
// WARNING: DATA RACE
// Write at 0x... by goroutine X
// Previous write at 0x... by goroutine Y
```

---

## Interview Questions

### Beginner Level

1. **Q:** What keyword creates a goroutine?
   **A:** `go functionName()` or `go func() { }()`

2. **Q:** What happens if you close a channel twice?
   **A:** Panic! A closed channel cannot be closed again.

3. **Q:** What is the zero value of a channel?
   **A:** `nil`. Sending/receiving on nil blocks forever, closing panics.

### Intermediate Level

4. **Q:** Explain the channel axioms table (nil channel, closed channel behaviors).
   **A:** See table in section 5.5 - nil blocks forever (send/receive) or panics (close); closed panics (send/close) or returns zero (receive).

5. **Q:** When would you use `sync.RWMutex` vs `sync.Mutex`?
   **A:** RWMutex for read-heavy workloads where multiple readers can proceed concurrently. Mutex for write-heavy or simple cases.

6. **Q:** What is the GMP model?
   **A:** G=Goroutine, M=Machine(OS thread), P=Processor(logical CPU). P has local run queue, M executes G from P.

### Advanced Level

7. **Q:** How does Go's scheduler handle a goroutine making a blocking syscall?
   **A:** The M (OS thread) is blocked, so the P is detached and assigned to another M. When syscall completes, G is put back in a run queue.

8. **Q:** Implement a worker pool with proper shutdown.
   **A:** Use WaitGroup for workers, close jobs channel to signal done, wait for workers, then close results.

9. **Q:** What is false sharing and how do concurrent Go programs avoid it?
   **A:** When goroutines modify variables on same cache line causing invalidation. Pad structs to 64 bytes to separate cache lines.

---

## Summary

| Topic | Key Points |
|-------|------------|
| Concurrency | Structure (dealing with), not parallelism (doing) |
| Goroutines | 2KB stack, microsecond creation, M:N scheduling |
| GMP | G=goroutine, M=thread, P=processor context |
| Channels | Unbuffered=sync, Buffered=async, know the axioms! |
| Select | Multiplexing, random if multiple ready, default=non-blocking |
| Sync | Mutex, RWMutex, WaitGroup, Once, Pool, Map |
| Atomic | Lock-free single-value operations |
| Context | Cancellation, deadlines, values (sparingly) |
| Patterns | Worker pool, fan-out/fan-in, pipeline, semaphore |

**Next Phase:** [Phase 6 — Testing & Engineering Reliability](../Phase_6/Phase_6.md)

