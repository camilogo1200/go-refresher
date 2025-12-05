# 💾 Phase 4: Memory Management & Performance

[← Back to Main Roadmap](../README.md) | [← Previous: Phase 3](../Phase_3/Phase_3.md)

---

**Objective:** Understand where bytes live and how to minimize runtime overhead.

**Reference:** [Go Memory Model](https://go.dev/ref/mem)

**Prerequisites:** Phase 0-3

**Estimated Duration:** 2-3 weeks

---

## 📋 Table of Contents

1. [Stack vs. Heap Allocation](#41-stack-vs-heap-allocation)
2. [Escape Analysis](#42-escape-analysis)
3. [Value Semantics vs. Pointer Semantics](#43-value-semantics-vs-pointer-semantics)
4. [Memory Layout & Alignment](#44-memory-layout--alignment)
5. [The Garbage Collector](#45-the-garbage-collector)
6. [GC Tuning](#46-gc-tuning)
7. [Reducing GC Pressure](#47-reducing-gc-pressure)
8. [The `unsafe` Package](#48-the-unsafe-package)
9. [Interview Questions](#interview-questions)

---

## 4.1 Stack vs. Heap Allocation

### Stack Memory

**Interview Question:** *"How does stack allocation work in Go? What are the benefits?"*

```
Stack characteristics:
- Per-goroutine (starts at 2KB)
- Grows contiguously (since Go 1.4)
- Allocation is cheap (pointer bump)
- Automatic cleanup (no GC)
- LIFO ordering
```

```go
func stackAllocation() {
    x := 42           // Stack allocated
    arr := [100]int{} // Stack allocated (fixed size)
    // When function returns, stack frame is popped
    // No GC involvement
}
```

**Stack growth mechanism:**
```
Initial: 2KB
Growth:  Copy entire stack to new, larger allocation
Shrink:  During GC if stack is underutilized
Maximum: ~1GB (configurable)
```

### Heap Memory

**Interview Question:** *"When does Go allocate on the heap vs. the stack?"*

```
Heap characteristics:
- Shared across goroutines
- Requires garbage collection
- Allocation involves runtime
- Fragmentation concerns
- Non-deterministic cleanup
```

```go
func heapAllocation() *int {
    x := 42
    return &x  // Must escape to heap (pointer survives function)
}
```

### Why It Matters

**Interview Question:** *"Why is stack allocation preferred over heap allocation?"*

| Stack | Heap |
|-------|------|
| O(1) allocation (pointer bump) | Runtime allocation |
| Automatic cleanup | GC overhead |
| Cache-friendly (locality) | Scattered in memory |
| No fragmentation | Fragmentation possible |
| No GC scanning needed | GC must track |

### Example: Allocation Visualization

```go
package main

func main() {
    // Stack allocated
    x := 42
    y := [10]int{1, 2, 3}
    
    // Heap allocated (escapes)
    z := make([]int, 100)  // Size determined at runtime
    p := new(int)          // Explicit heap allocation
    q := &x                // If passed to escaping function
    
    _ = z
    _ = p
    _ = q
}
```

---

## 4.2 Escape Analysis

### What is Escape Analysis?

**Interview Question:** *"What is escape analysis? How do you view escape analysis output?"*

Escape analysis is a compile-time determination of whether a value can remain on the stack or must be allocated on the heap.

```bash
# View escape analysis decisions
go build -gcflags="-m" main.go

# More verbose output
go build -gcflags="-m -m" main.go
```

### Common Escape Causes

**Interview Question:** *"What causes a variable to escape to the heap?"*

#### 1. Returning pointer to local variable

```go
func escapes() *int {
    x := 42
    return &x  // x escapes to heap
}
```

#### 2. Storing pointer in interface

```go
func escapes2() {
    x := 42
    var i interface{} = &x  // x escapes (interface boxing)
}
```

#### 3. Closure capturing pointer

```go
func escapes3() func() *int {
    x := 42
    return func() *int {
        return &x  // x escapes (captured by closure)
    }
}
```

#### 4. Unknown slice capacity

```go
func escapes4(n int) {
    s := make([]int, n)  // n not known at compile time - escapes
    _ = s
}

func noEscape() {
    s := make([]int, 100)  // Known size - may stay on stack
    _ = s
}
```

#### 5. Sending pointer to channel

```go
func escapes5(ch chan *int) {
    x := 42
    ch <- &x  // x escapes (leaves goroutine)
}
```

#### 6. Values too large for stack

```go
func escapes6() {
    arr := [1000000]int{}  // Too large, escapes
    _ = arr
}
```

### Preventing Escapes

**Interview Question:** *"How can you write code that avoids unnecessary heap allocations?"*

```go
// BAD: Returns pointer, forces heap allocation
func NewConfig() *Config {
    return &Config{
        Port: 8080,
    }
}

// BETTER: Return value if struct is small
func NewConfig() Config {
    return Config{
        Port: 8080,
    }
}

// BAD: Unknown capacity forces escape
func process(items []string) []string {
    result := make([]string, 0)  // May escape
    // ...
}

// BETTER: Pre-allocate with known capacity
func process(items []string) []string {
    result := make([]string, 0, len(items))  // Hint helps
    // ...
}

// BAD: Interface boxing
func log(v interface{}) {
    fmt.Println(v)  // v boxed
}

// BETTER: Type-specific or generics
func log[T any](v T) {
    fmt.Println(v)
}
```

### Example: Analyzing Escape Behavior

```go
// Save as escape.go
package main

func noEscape() int {
    x := 42
    return x  // Value returned, no escape
}

func escapePointer() *int {
    x := 42    // moved to heap: x escapes
    return &x
}

func escapeSlice(n int) []int {
    // make([]int, n) escapes to heap (unknown size)
    return make([]int, n)
}

func noEscapeSlice() []int {
    // May stay on stack (known size, small)
    s := make([]int, 10)
    return s  // But returning it causes escape!
}

func main() {
    _ = noEscape()
    _ = escapePointer()
    _ = escapeSlice(10)
    _ = noEscapeSlice()
}
```

```bash
$ go build -gcflags="-m" escape.go
# Output shows which variables escape
./escape.go:8:2: moved to heap: x
./escape.go:13:13: make([]int, n) escapes to heap
./escape.go:18:11: make([]int, 10) escapes to heap
```

---

## 4.3 Value Semantics vs. Pointer Semantics

### Value Semantics

**Interview Question:** *"What are the benefits of value semantics in Go?"*

```go
type Point struct {
    X, Y float64
}

func (p Point) Distance(q Point) float64 {
    dx := p.X - q.X
    dy := p.Y - q.Y
    return math.Sqrt(dx*dx + dy*dy)
}

func main() {
    p1 := Point{0, 0}
    p2 := Point{3, 4}
    
    // p1 and p2 are copied - no aliasing
    dist := p1.Distance(p2)
}
```

**Benefits:**
- No aliasing (mutation is explicit)
- Better cache locality
- No nil pointer risks
- GC doesn't trace (if no pointers in type)
- Thread-safe by default

### Pointer Semantics

```go
type LargeStruct struct {
    Data [1000]byte
}

func (l *LargeStruct) Process() {
    // Operates on original, no copy
}

func main() {
    ls := &LargeStruct{}
    ls.Process()  // No 1000-byte copy
}
```

**Benefits:**
- Efficient for large structures
- Enables mutation
- Required for some interfaces
- Single source of truth

### The 64-Byte Rule

**Interview Question:** *"When should you use value vs. pointer semantics?"*

```go
// Small struct - prefer value semantics
type Point struct {
    X, Y float64  // 16 bytes
}

// Large struct - prefer pointer semantics
type Image struct {
    Pixels [1024 * 1024]byte  // 1MB
}

// Rule of thumb: ~64 bytes is the threshold
// Below 64 bytes: value is often faster (no indirection)
// Above 64 bytes: pointer avoids copy overhead
```

### Consistency Rule

**Interview Question:** *"Why is consistency important when choosing value vs. pointer semantics?"*

```go
// GOOD: Consistent pointer semantics
type User struct {
    Name string
    Age  int
}

func (u *User) SetName(name string) { u.Name = name }
func (u *User) GetName() string { return u.Name }
func (u *User) SetAge(age int) { u.Age = age }
func (u *User) GetAge() int { return u.Age }

// BAD: Mixed semantics (confusing)
func (u *User) SetName(name string) { u.Name = name }
func (u User) GetName() string { return u.Name }  // Why value here?
```

**Rule:** Pick one semantic per type and stick to it.

### Example: Choosing Semantics

```go
package main

import "fmt"

// Small, immutable - value semantics
type Color struct {
    R, G, B uint8
}

func (c Color) Brightness() float64 {
    return float64(c.R+c.G+c.B) / 3.0
}

// Mutable state - pointer semantics
type Counter struct {
    count int
}

func NewCounter() *Counter {
    return &Counter{}
}

func (c *Counter) Increment() {
    c.count++
}

func (c *Counter) Value() int {
    return c.count
}

// Large data - pointer semantics
type Document struct {
    Content []byte
    Meta    map[string]string
}

func NewDocument(content []byte) *Document {
    return &Document{
        Content: content,
        Meta:    make(map[string]string),
    }
}

func main() {
    // Value type - copied
    red := Color{255, 0, 0}
    redCopy := red
    fmt.Println(red.Brightness())
    
    // Pointer type - shared
    counter := NewCounter()
    counter.Increment()
    fmt.Println(counter.Value())
}
```

---

## 4.4 Memory Layout & Alignment

### Alignment Requirements

**Interview Question:** *"What is memory alignment and why does it matter?"*

```go
// Types must be aligned to their size
// int64: 8-byte aligned
// int32: 4-byte aligned
// int16: 2-byte aligned
// int8:  1-byte aligned
```

```go
import "unsafe"

type Example struct {
    a bool   // 1 byte
    b int64  // 8 bytes, needs 8-byte alignment
    c bool   // 1 byte
}

// Memory layout with padding:
// [a:1][pad:7][b:8][c:1][pad:7] = 24 bytes!

fmt.Println(unsafe.Sizeof(Example{}))  // 24
```

### Struct Field Ordering

**Interview Question:** *"How can you optimize struct memory layout?"*

```go
// Inefficient layout (24 bytes)
type Inefficient struct {
    a bool    // 1 + 7 padding
    b int64   // 8
    c bool    // 1 + 7 padding
}

// Efficient layout (16 bytes)
type Efficient struct {
    b int64   // 8
    a bool    // 1
    c bool    // 1 + 6 padding
}

fmt.Println(unsafe.Sizeof(Inefficient{}))  // 24
fmt.Println(unsafe.Sizeof(Efficient{}))    // 16
```

**Rule:** Order fields from largest to smallest.

### Inspection Tools

```go
import "unsafe"

type MyStruct struct {
    A int32
    B int64
    C int16
}

func main() {
    fmt.Println("Size:", unsafe.Sizeof(MyStruct{}))
    fmt.Println("Align:", unsafe.Alignof(MyStruct{}))
    
    var s MyStruct
    fmt.Println("Offset A:", unsafe.Offsetof(s.A))
    fmt.Println("Offset B:", unsafe.Offsetof(s.B))
    fmt.Println("Offset C:", unsafe.Offsetof(s.C))
}
```

### Cache Line Awareness

**Interview Question:** *"What is false sharing and how do you prevent it?"*

```go
// Modern CPUs have 64-byte cache lines
// False sharing: two goroutines modify variables on same cache line

// BAD: Counters share cache line
type Counters struct {
    A int64  // Same cache line
    B int64  // Same cache line - contention!
}

// GOOD: Pad to separate cache lines
type Counters struct {
    A   int64
    _   [56]byte  // Padding to 64 bytes
    B   int64
    _   [56]byte
}
```

### Example: Memory Layout Analysis

```go
package main

import (
    "fmt"
    "unsafe"
)

// Analyze different struct layouts
type BadLayout struct {
    flag1 bool
    count int64
    flag2 bool
    value int32
    flag3 bool
}

type GoodLayout struct {
    count int64
    value int32
    flag1 bool
    flag2 bool
    flag3 bool
    // 1 byte padding
}

func main() {
    fmt.Printf("BadLayout:  %d bytes\n", unsafe.Sizeof(BadLayout{}))
    fmt.Printf("GoodLayout: %d bytes\n", unsafe.Sizeof(GoodLayout{}))
    
    // Output:
    // BadLayout:  40 bytes
    // GoodLayout: 16 bytes
}
```

---

## 4.5 The Garbage Collector

### Tricolor Mark-and-Sweep

**Interview Question:** *"Explain how Go's garbage collector works."*

```
Tricolor algorithm:
1. WHITE: Potentially garbage (not yet scanned)
2. GRAY:  Reachable, but children not scanned
3. BLACK: Reachable, all children scanned

Process:
1. Start: All objects WHITE
2. Mark roots (stacks, globals) as GRAY
3. For each GRAY object:
   - Scan its pointers
   - Mark referenced objects GRAY
   - Mark current object BLACK
4. Repeat until no GRAY objects
5. Sweep: Reclaim WHITE objects
```

### Concurrent Collection

**Interview Question:** *"How does Go's GC minimize pause times?"*

```
GC Phases:
1. Mark Setup (STW ~10-30μs)
   - Enable write barrier
   - Prepare for marking
   
2. Marking (Concurrent)
   - Most work happens here
   - Application runs alongside
   
3. Mark Termination (STW ~10-30μs)
   - Finish marking
   - Disable write barrier
   
4. Sweeping (Concurrent)
   - Reclaim memory
   - Application runs
```

**Write Barrier:** Ensures objects aren't incorrectly collected during concurrent marking.

### GC Triggers

```go
// Automatic triggers:
// 1. Heap doubles (controlled by GOGC)
// 2. 2 minutes without GC
// 3. OS memory pressure

// Manual trigger (rarely needed):
runtime.GC()

// View GC statistics:
var stats runtime.MemStats
runtime.ReadMemStats(&stats)
fmt.Printf("Alloc: %d MB\n", stats.Alloc/1024/1024)
fmt.Printf("TotalAlloc: %d MB\n", stats.TotalAlloc/1024/1024)
fmt.Printf("NumGC: %d\n", stats.NumGC)
```

### Example: Monitoring GC

```go
package main

import (
    "fmt"
    "runtime"
    "time"
)

func main() {
    // Enable GC trace
    // Run with: GODEBUG=gctrace=1 go run main.go
    
    var stats runtime.MemStats
    
    // Allocate memory
    var data [][]byte
    for i := 0; i < 100; i++ {
        data = append(data, make([]byte, 1024*1024))  // 1MB each
        
        runtime.ReadMemStats(&stats)
        fmt.Printf("Iteration %d: Alloc=%dMB, NumGC=%d\n",
            i, stats.Alloc/1024/1024, stats.NumGC)
        
        time.Sleep(100 * time.Millisecond)
    }
}
```

---

## 4.6 GC Tuning

### GOGC Environment Variable

**Interview Question:** *"How do you tune the Go garbage collector?"*

```bash
# Default: GC triggers at 100% heap growth
GOGC=100 ./myapp

# Less frequent GC (more memory, less CPU)
GOGC=200 ./myapp

# More frequent GC (less memory, more CPU)
GOGC=50 ./myapp

# Disable GC (careful!)
GOGC=off ./myapp
```

```go
// Runtime adjustment
import "runtime/debug"

debug.SetGCPercent(50)   // More aggressive
debug.SetGCPercent(200)  // Less aggressive
debug.SetGCPercent(-1)   // Disable GC
```

### GOMEMLIMIT (Go 1.19+)

**Interview Question:** *"What is GOMEMLIMIT and when would you use it?"*

```bash
# Set soft memory limit
GOMEMLIMIT=1GiB ./myapp

# Common in containerized environments
GOMEMLIMIT=512MiB ./myapp
```

```go
// Runtime adjustment
import "runtime/debug"

debug.SetMemoryLimit(1024 * 1024 * 1024)  // 1 GiB
```

**Benefits:**
- Prevents OOM in containers
- Better memory utilization
- Replaces "memory ballast" hack

### When to Tune

**Interview Question:** *"When should you tune the GC?"*

```go
// Profile first, tune second!
// Tuning scenarios:

// 1. High-throughput, memory-rich
GOGC=200  // Less frequent GC
GOMEMLIMIT=4GiB  // Cap total memory

// 2. Latency-sensitive
GOGC=50  // More frequent, smaller pauses

// 3. Memory-constrained container
GOMEMLIMIT=256MiB  // Hard limit
GOGC=100  // Default behavior within limit

// 4. Batch processing
GOGC=off  // Disable during batch
runtime.GC()  // Explicit GC at end
```

### Example: GC Tuning in Practice

```go
package main

import (
    "fmt"
    "runtime"
    "runtime/debug"
    "time"
)

func main() {
    // Set memory limit for container
    debug.SetMemoryLimit(512 * 1024 * 1024)  // 512MB
    
    // Monitor GC behavior
    go func() {
        for {
            var stats runtime.MemStats
            runtime.ReadMemStats(&stats)
            
            fmt.Printf("HeapAlloc: %dMB, NumGC: %d, PauseTotal: %dms\n",
                stats.HeapAlloc/1024/1024,
                stats.NumGC,
                stats.PauseTotalNs/1000000)
            
            time.Sleep(time.Second)
        }
    }()
    
    // Application work...
    select {}
}
```

---

## 4.7 Reducing GC Pressure

### Object Pooling with `sync.Pool`

**Interview Question:** *"How does `sync.Pool` work and when should you use it?"*

```go
import "sync"

var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 4096)
    },
}

func processRequest() {
    // Get buffer from pool (or create new)
    buf := bufferPool.Get().([]byte)
    
    // Reset buffer before use
    buf = buf[:0]
    
    // Use buffer...
    
    // Return to pool
    bufferPool.Put(buf)
}
```

**Characteristics:**
- Objects may be garbage collected
- Not guaranteed to retain objects between GCs
- Best for temporary, frequently allocated objects
- Reduces allocation pressure in hot paths

### Pre-allocation

**Interview Question:** *"How does pre-allocation improve performance?"*

```go
// BAD: Multiple allocations during append
func slowBuild(n int) []int {
    var result []int
    for i := 0; i < n; i++ {
        result = append(result, i)  // May reallocate
    }
    return result
}

// GOOD: Single allocation
func fastBuild(n int) []int {
    result := make([]int, 0, n)  // Pre-allocate capacity
    for i := 0; i < n; i++ {
        result = append(result, i)  // No reallocation
    }
    return result
}

// For maps too
m := make(map[string]int, expectedSize)
```

### String Building

**Interview Question:** *"What's the most efficient way to build strings in Go?"*

```go
// BAD: O(n²) - creates new string each iteration
func slowConcat(items []string) string {
    result := ""
    for _, item := range items {
        result += item
    }
    return result
}

// GOOD: O(n) - uses buffer
func fastConcat(items []string) string {
    var builder strings.Builder
    
    // Pre-calculate total size
    size := 0
    for _, item := range items {
        size += len(item)
    }
    builder.Grow(size)
    
    for _, item := range items {
        builder.WriteString(item)
    }
    return builder.String()
}
```

### Avoiding Interface Boxing

**Interview Question:** *"What is interface boxing and how does it affect allocation?"*

```go
// Interface boxing allocates
func logAny(v interface{}) {
    fmt.Println(v)  // v is boxed
}

func main() {
    x := 42
    logAny(x)  // Boxing: allocates iface{type: int, data: &x}
}

// Avoid with generics (Go 1.18+)
func logTyped[T any](v T) {
    fmt.Println(v)  // No boxing for concrete types
}

// Or type-specific functions
func logInt(v int) {
    fmt.Println(v)
}
```

### Example: Complete Optimization

```go
package main

import (
    "bytes"
    "sync"
)

// Object pool for buffers
var bufPool = sync.Pool{
    New: func() interface{} {
        return bytes.NewBuffer(make([]byte, 0, 4096))
    },
}

// ProcessData demonstrates allocation-conscious code
func ProcessData(items []string) string {
    // Get buffer from pool
    buf := bufPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufPool.Put(buf)
    }()
    
    // Pre-calculate size for efficiency
    totalLen := 0
    for _, item := range items {
        totalLen += len(item) + 1  // +1 for separator
    }
    buf.Grow(totalLen)
    
    // Build string efficiently
    for i, item := range items {
        if i > 0 {
            buf.WriteByte(',')
        }
        buf.WriteString(item)
    }
    
    return buf.String()
}

func main() {
    items := []string{"apple", "banana", "cherry"}
    result := ProcessData(items)
    println(result)
}
```

---

## 4.8 The `unsafe` Package

### When to Use `unsafe`

**Interview Question:** *"When is it appropriate to use the `unsafe` package?"*

```go
// Legitimate use cases:
// 1. Performance-critical code (after profiling!)
// 2. C interop (cgo)
// 3. Low-level system programming
// 4. Serialization/deserialization
// 5. Memory-mapped I/O

// AVOID for:
// - Regular application code
// - When safe alternatives exist
// - When you don't understand implications
```

### `unsafe.Pointer`

```go
import "unsafe"

func main() {
    x := 42
    
    // Regular pointer
    p := &x
    
    // unsafe.Pointer - generic pointer
    up := unsafe.Pointer(p)
    
    // Convert to different pointer type
    fp := (*float64)(up)  // DANGER: Type punning
    
    // Convert to uintptr for arithmetic
    addr := uintptr(up)
    addr += 8
    newUp := unsafe.Pointer(addr)
}
```

### Zero-Copy String Conversion

**Interview Question:** *"How can you convert between string and []byte without copying?"*

```go
import "unsafe"

// Zero-copy string to []byte (DANGER: mutation will corrupt string!)
func stringToBytes(s string) []byte {
    return unsafe.Slice(unsafe.StringData(s), len(s))
}

// Zero-copy []byte to string
func bytesToString(b []byte) string {
    return unsafe.String(unsafe.SliceData(b), len(b))
}

// Go 1.20+ provides these safely
```

### Struct Field Access

```go
import "unsafe"

type Secret struct {
    public  int
    private int  // unexported
}

func hackPrivate(s *Secret) int {
    // Get offset of private field
    offset := unsafe.Offsetof(s.private)
    
    // Get pointer to private field
    privatePtr := (*int)(unsafe.Pointer(
        uintptr(unsafe.Pointer(s)) + offset,
    ))
    
    return *privatePtr
}
```

### Risks of `unsafe`

```go
// 1. Memory corruption
x := 42
p := unsafe.Pointer(&x)
*(*[1000]byte)(p) = [1000]byte{}  // Overwrites adjacent memory!

// 2. Undefined behavior
uptr := uintptr(unsafe.Pointer(&x))
// GC may move x here!
p2 := unsafe.Pointer(uptr)  // May point to wrong location

// 3. Race conditions
// unsafe operations are not atomic

// 4. Breaking portability
// Size/alignment may differ between platforms
```

### Example: Safe Usage Patterns

```go
package main

import (
    "fmt"
    "unsafe"
)

// Pattern 1: Type size inspection
func inspectSize[T any]() {
    var zero T
    fmt.Printf("%T: size=%d, align=%d\n",
        zero, unsafe.Sizeof(zero), unsafe.Alignof(zero))
}

// Pattern 2: Atomic pointer operations (use sync/atomic instead!)
// This is just for illustration

// Pattern 3: Zero-copy conversion (Go 1.20+)
func zeroCopyBytesToString(b []byte) string {
    if len(b) == 0 {
        return ""
    }
    return unsafe.String(&b[0], len(b))
}

func main() {
    inspectSize[int64]()
    inspectSize[string]()
    inspectSize[[]int]()
    
    b := []byte("hello")
    s := zeroCopyBytesToString(b)
    fmt.Println(s)
    
    // WARNING: b and s share memory!
    // Modifying b will corrupt s
}
```

---

## Interview Questions

### Beginner Level

1. **Q:** What are the two memory regions for variable allocation?
   **A:** Stack (per-goroutine, automatic cleanup) and Heap (shared, garbage collected).

2. **Q:** How can you see escape analysis output?
   **A:** `go build -gcflags="-m"` shows which variables escape to heap.

3. **Q:** What is the zero value for a pointer?
   **A:** `nil`

### Intermediate Level

4. **Q:** Why might returning a pointer from a function cause heap allocation?
   **A:** The pointed-to value must outlive the function, so it escapes to heap.

5. **Q:** What is the purpose of `GOMEMLIMIT`?
   **A:** Sets a soft memory limit for the Go runtime, preventing OOM in containers and improving GC behavior.

6. **Q:** How does `sync.Pool` help reduce GC pressure?
   **A:** It recycles objects between uses, reducing allocation frequency. Objects may still be GC'd.

### Advanced Level

7. **Q:** Explain the tricolor mark-and-sweep algorithm.
   **A:** Objects are WHITE (unknown), GRAY (reachable, unscanned), or BLACK (reachable, scanned). GC marks roots gray, then iteratively scans gray objects until none remain. White objects are garbage.

8. **Q:** What is false sharing and how do you prevent it?
   **A:** When goroutines modify variables on the same cache line, causing cache invalidation. Prevent by padding structs to 64 bytes.

9. **Q:** What are the risks of using `unsafe.Pointer`?
   **A:** Memory corruption, undefined behavior (especially with uintptr), race conditions, platform-specific behavior, GC interference.

---

## Summary

| Topic | Key Points |
|-------|------------|
| Stack vs. Heap | Stack is fast, per-goroutine; Heap is shared, GC'd |
| Escape Analysis | `go build -gcflags="-m"` to view decisions |
| Semantics | Value for small/immutable, Pointer for large/mutable |
| Alignment | Order fields largest-to-smallest to minimize padding |
| GC | Concurrent tricolor mark-and-sweep |
| Tuning | `GOGC` for frequency, `GOMEMLIMIT` for cap |
| Optimization | sync.Pool, pre-allocation, strings.Builder |
| unsafe | Use sparingly, understand implications |

**Next Phase:** [Phase 5 — Concurrency & The Scheduler](../Phase_5/Phase_5.md)

