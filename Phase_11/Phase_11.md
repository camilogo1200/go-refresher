# 🔬 Phase 11: Go Runtime Internals

[← Back to Main Roadmap](../README.md) | [← Previous: Phase 10](../Phase_10/Phase_10.md)

---

**Objective:** Deep understanding of how Go executes code for expert-level debugging and optimization.

**Reference:** [Go Runtime Source](https://github.com/golang/go/tree/master/src/runtime), [Go Memory Model](https://go.dev/ref/mem)

**Prerequisites:** Phase 0-10

**Estimated Duration:** 3-4 weeks (Advanced)

---

## 📋 Table of Contents

1. [The Go Runtime Overview](#111-the-go-runtime-overview)
2. [Memory Allocator](#112-memory-allocator)
3. [Stack Management](#113-stack-management)
4. [Garbage Collector Internals](#114-garbage-collector-internals)
5. [Scheduler Internals](#115-scheduler-internals)
6. [Channel Internals](#116-channel-internals)
7. [Interface Internals](#117-interface-internals)
8. [Reflection](#118-reflection)
9. [Runtime Debugging](#119-runtime-debugging)
10. [Interview Questions](#interview-questions)

---

## 11.1 The Go Runtime Overview

### What's in the Runtime?

**Interview Question:** *"What components are included in every Go binary?"*

```
Every Go binary includes:
┌─────────────────────────────────────┐
│           Go Binary                  │
├─────────────────────────────────────┤
│  Your Application Code              │
├─────────────────────────────────────┤
│  Standard Library (used portions)   │
├─────────────────────────────────────┤
│  Runtime:                           │
│  ├── Scheduler (GMP)                │
│  ├── Memory Allocator               │
│  ├── Garbage Collector              │
│  ├── Stack Manager                  │
│  ├── Channel Implementation         │
│  ├── defer/panic/recover            │
│  ├── Reflection Support             │
│  └── OS Interface                   │
└─────────────────────────────────────┘
```

### The `runtime` Package

```go
import "runtime"

// Scheduler control
runtime.GOMAXPROCS(4)        // Set/get P count
runtime.NumGoroutine()       // Current goroutine count
runtime.Gosched()            // Yield to scheduler
runtime.Goexit()             // Terminate current goroutine

// Memory information
var m runtime.MemStats
runtime.ReadMemStats(&m)
fmt.Printf("Alloc: %d MB\n", m.Alloc/1024/1024)
fmt.Printf("HeapObjects: %d\n", m.HeapObjects)
fmt.Printf("GC Cycles: %d\n", m.NumGC)

// GC control
runtime.GC()                 // Trigger GC
debug.SetGCPercent(200)      // Set GOGC

// Stack trace
buf := make([]byte, 4096)
n := runtime.Stack(buf, false)  // Current goroutine
n := runtime.Stack(buf, true)   // All goroutines

// Caller information
pc, file, line, ok := runtime.Caller(0)
```

### Environment Variables

```bash
# Scheduler
GOMAXPROCS=4           # Number of OS threads for goroutines

# GC tuning
GOGC=100               # GC target percentage (default 100)
GOMEMLIMIT=1GiB        # Soft memory limit (Go 1.19+)

# Debugging
GODEBUG=gctrace=1                    # GC trace
GODEBUG=schedtrace=1000              # Scheduler trace every 1000ms
GODEBUG=scheddetail=1                # Detailed scheduler trace
GODEBUG=allocfreetrace=1             # Allocation trace
GODEBUG=madvdontneed=1               # Memory advice (Linux)

# Panic behavior
GOTRACEBACK=all        # Print all goroutine stacks on panic
GOTRACEBACK=crash      # Trigger core dump on panic
```

---

## 11.2 Memory Allocator

### Memory Hierarchy

**Interview Question:** *"Explain Go's memory allocation architecture."*

```
┌────────────────────────────────────────────────────────┐
│                        mheap                            │
│        (Global heap manager, single instance)           │
├────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────┐  │
│  │              mcentral (per size class)            │  │
│  │      (Shared cache for each size class)           │  │
│  └──────────────────────────────────────────────────┘  │
├────────────────────────────────────────────────────────┤
│  ┌─────────┐  ┌─────────┐  ┌─────────┐               │
│  │ mcache  │  │ mcache  │  │ mcache  │  (Per-P)      │
│  │  (P0)   │  │  (P1)   │  │  (P2)   │               │
│  └─────────┘  └─────────┘  └─────────┘               │
└────────────────────────────────────────────────────────┘

Allocation path (small objects):
1. mcache (no lock) → fastest
2. mcentral (lock per size class) → fast
3. mheap (global lock) → slow
```

### Size Classes

```go
// Go uses ~70 size classes from 8 bytes to 32KB
// Objects > 32KB are allocated directly from heap

// Size class examples:
// Class  Size   Objects per span
//   1      8      1024
//   2     16       512
//   3     24       341
//   ...
//  67    32768        1

// Query size class for an object
import "runtime"

func printSizeClass(size uintptr) {
    // Internal: runtime maps sizes to classes
    // 8, 16, 24, 32, 48, 64, 80, 96, 112, 128...
}
```

### Span Management

```go
// mspan: A contiguous run of pages containing objects of one size class

type mspan struct {
    next       *mspan      // Next span in list
    prev       *mspan      // Previous span
    startAddr  uintptr     // Start address
    npages     uintptr     // Number of pages
    freeindex  uintptr     // Next free object
    nelems     uintptr     // Number of objects
    allocCount uint16      // Allocated objects
    spanclass  spanClass   // Size class and noscan
    // ...
}

// Span states:
// mSpanDead    - Released to OS
// mSpanInUse   - Allocated for objects
// mSpanManual  - Manual management (stacks)
// mSpanFree    - Free for allocation
```

### Tiny Allocator

**Interview Question:** *"What is Go's tiny allocator?"*

```go
// Objects < 16 bytes without pointers use "tiny allocator"
// Multiple tiny objects packed into single 16-byte block

// Example: Multiple small allocations
var a int8   // 1 byte
var b int8   // 1 byte
var c int8   // 1 byte
// These may share a single 16-byte block!

// Benefits:
// - Reduces memory fragmentation
// - Fewer allocations tracked by GC
// - Faster allocation

// Restrictions:
// - Only for objects without pointers (noscan)
// - Size < 16 bytes
```

### Large Object Allocation

```go
// Objects > 32KB allocated directly from heap
// Uses mheap.allocLarge()

// These bypass size class system
bigSlice := make([]byte, 64*1024)  // 64KB - large allocation

// Large objects:
// - Rounded to page size (8KB on most systems)
// - Allocated in contiguous pages
// - Higher fragmentation risk
```

---

## 11.3 Stack Management

### Goroutine Stack

**Interview Question:** *"How does Go manage goroutine stacks?"*

```go
// Stack characteristics:
// - Initial size: 2KB (configurable)
// - Maximum size: 1GB (64-bit), 250MB (32-bit)
// - Growth: 2x when needed
// - Contiguous (since Go 1.4)

// Stack growth process:
// 1. Function prologue checks stack space
// 2. If insufficient, runtime.morestack called
// 3. New larger stack allocated
// 4. Old stack copied to new location
// 5. Pointers updated
// 6. Old stack freed
```

### Stack Layout

```
High Address
┌───────────────────┐
│    Arguments      │  ← Passed to function
├───────────────────┤
│  Return address   │
├───────────────────┤
│   Frame pointer   │  ← Optional (for profiling)
├───────────────────┤
│  Local variables  │
├───────────────────┤
│   Saved registers │
├───────────────────┤
│  Space for calls  │  ← For calling other functions
└───────────────────┘
Low Address (SP points here)
```

### Stack vs Heap Decision

```go
// Compiler decides via escape analysis
// go build -gcflags="-m" shows decisions

func stackAlloc() int {
    x := 42           // Stack - doesn't escape
    return x
}

func heapAlloc() *int {
    x := 42           // Heap - escapes via return
    return &x
}

func sliceStack() {
    s := make([]int, 100)    // Stack - size known, doesn't escape
    _ = s
}

func sliceHeap(n int) []int {
    s := make([]int, n)      // Heap - size unknown at compile time
    return s
}
```

---

## 11.4 Garbage Collector Internals

### Tricolor Mark-and-Sweep

**Interview Question:** *"Explain Go's tricolor garbage collection algorithm."*

```
Three colors represent object states:

WHITE: Potentially garbage (not yet visited)
GRAY:  Reachable, but children not scanned
BLACK: Reachable, all children scanned

Algorithm:
1. Initially all objects are white
2. Root objects marked gray
3. Process gray objects:
   - Mark object black
   - Mark children gray
4. Repeat until no gray objects
5. White objects are garbage

Invariant: No black object points to white object
(Maintained by write barrier)
```

### GC Phases

```
┌─────────────────────────────────────────────────────┐
│ Phase 1: Mark Setup (STW)                           │
│ - Enable write barrier                              │
│ - Prepare for marking                               │
│ Duration: ~10-30 microseconds                       │
├─────────────────────────────────────────────────────┤
│ Phase 2: Marking (Concurrent)                       │
│ - Scan stacks, globals, heap                        │
│ - Application runs simultaneously                   │
│ - Write barrier tracks mutations                    │
├─────────────────────────────────────────────────────┤
│ Phase 3: Mark Termination (STW)                     │
│ - Drain mark work                                   │
│ - Disable write barrier                             │
│ Duration: ~10-30 microseconds                       │
├─────────────────────────────────────────────────────┤
│ Phase 4: Sweeping (Concurrent)                      │
│ - Reclaim unmarked spans                            │
│ - Done lazily during allocation                     │
└─────────────────────────────────────────────────────┘
```

### Write Barrier

```go
// Write barrier ensures tricolor invariant during concurrent marking

// Conceptually (actual implementation is more complex):
func writeBarrier(slot *unsafe.Pointer, ptr unsafe.Pointer) {
    // Shade the old pointer (if any)
    shade(*slot)
    // Shade the new pointer
    shade(ptr)
    // Perform the actual write
    *slot = ptr
}

// This prevents black->white pointers
// If black object gets new white child, child becomes gray
```

### GC Pacing

```go
// GC tries to complete before heap doubles
// Controlled by GOGC and GOMEMLIMIT

// GOGC=100 (default): GC when heap is 100% larger than live heap
// Live heap: 100MB
// Next GC: 200MB

// GOMEMLIMIT helps prevent OOM:
// - Triggers more aggressive GC near limit
// - Useful for containers with memory limits

// Pacing formula (simplified):
// GC_trigger = live_heap * (1 + GOGC/100)
```

### GC Trace

```bash
GODEBUG=gctrace=1 ./myapp

# Output format:
# gc 1 @0.014s 2%: 0.52+1.0+0.013 ms clock, 4.2+0/2.0/1.5+0.10 ms cpu, 4->4->1 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 8 P

# gc 1         - GC number
# @0.014s      - Time since start
# 2%           - % time in GC
# 0.52+1.0+0.013 ms clock:
#   - STW mark setup
#   - Concurrent mark
#   - STW mark termination
# 4->4->1 MB:
#   - Heap before
#   - Heap after
#   - Live heap
# 4 MB goal    - Target heap size
# 8 P          - GOMAXPROCS
```

---

## 11.5 Scheduler Internals

### GMP Model Deep Dive

**Interview Question:** *"Explain the GMP scheduler in detail."*

```go
// G - Goroutine
type g struct {
    stack       stack     // Stack bounds
    stackguard0 uintptr   // Stack growth check
    m           *m        // Current M (nil if not running)
    sched       gobuf     // Saved context
    atomicstatus atomic.Uint32  // Status
    goid         uint64   // Goroutine ID
    // ...
}

// M - Machine (OS Thread)
type m struct {
    g0          *g        // Goroutine for scheduler
    curg        *g        // Current running goroutine
    p           puintptr  // Associated P (nil if not running)
    nextp       puintptr  // Next P to acquire
    spinning    bool      // Looking for work
    // ...
}

// P - Processor
type p struct {
    status      uint32        // Status
    link        puintptr      // Next P in list
    m           muintptr      // Associated M
    runq        [256]guintptr // Local run queue
    runqhead    uint32        // Head of run queue
    runqtail    uint32        // Tail of run queue
    runnext     guintptr      // Next G to run
    mcache      *mcache       // Memory cache
    // ...
}
```

### Goroutine States

```
            ┌──────────────┐
            │  _Gidle      │  (New, not yet started)
            └──────┬───────┘
                   │ go func()
                   ▼
            ┌──────────────┐
            │  _Grunnable  │◄──────────────────────┐
            └──────┬───────┘                       │
                   │ scheduled                     │
                   ▼                               │
            ┌──────────────┐                       │
    ┌──────►│  _Grunning   │───────────────────────┤
    │       └──────┬───────┘                       │
    │              │                               │
    │   ┌──────────┼────────────┐                  │
    │   │          │            │                  │
    │   │ syscall  │ chan/lock  │ preempt          │
    │   ▼          ▼            ▼                  │
    │ ┌────────┐ ┌──────────┐ ┌──────────┐        │
    │ │_Gsyscall│ │_Gwaiting │ │_Gpreempted│       │
    │ └────────┘ └──────────┘ └──────────┘        │
    │   │          │            │                  │
    │   └──────────┴────────────┴──────────────────┘
    │              │
    │              │ return
    │              ▼
    │       ┌──────────────┐
    └───────┤  _Gdead      │  (Finished)
            └──────────────┘
```

### Scheduling Algorithm

```go
// Simplified scheduling loop (runtime.schedule)
func schedule() {
    // 1. Get current P
    pp := getg().m.p.ptr()
    
    // 2. Try local run queue (no lock!)
    if gp := pp.runqget(); gp != nil {
        execute(gp)
        return
    }
    
    // 3. Try global run queue (with lock)
    if gp := globrunqget(pp); gp != nil {
        execute(gp)
        return
    }
    
    // 4. Poll network
    if gp := netpoll(); gp != nil {
        execute(gp)
        return
    }
    
    // 5. Steal from other P's
    for i := 0; i < 4; i++ {
        for _, p2 := range allp {
            if gp := runqsteal(pp, p2); gp != nil {
                execute(gp)
                return
            }
        }
    }
    
    // 6. Nothing to do - stop this M
    stopm()
}
```

### Scheduler Trace

```bash
GODEBUG=schedtrace=1000,scheddetail=1 ./myapp

# Output:
# SCHED 0ms: gomaxprocs=8 idleprocs=7 threads=2 spinningthreads=0 idlethreads=0 runqueue=0 [0 0 0 0 0 0 0 0]
#
# gomaxprocs    - GOMAXPROCS value
# idleprocs     - Idle P count
# threads       - Total M count
# spinningthreads - M's looking for work
# runqueue      - Global run queue length
# [0 0 0 0...]  - Per-P local run queue lengths
```

---

## 11.6 Channel Internals

### hchan Structure

**Interview Question:** *"How are channels implemented internally?"*

```go
// Channel internal structure (simplified)
type hchan struct {
    qcount   uint           // Elements in queue
    dataqsiz uint           // Size of circular queue
    buf      unsafe.Pointer // Pointer to buffer
    elemsize uint16         // Element size
    closed   uint32         // Closed flag
    elemtype *_type         // Element type
    sendx    uint           // Send index
    recvx    uint           // Receive index
    recvq    waitq          // Receivers waiting
    sendq    waitq          // Senders waiting
    lock     mutex          // Protects all fields
}

// Wait queue entry
type sudog struct {
    g     *g              // Waiting goroutine
    elem  unsafe.Pointer  // Data element
    next  *sudog          // Next waiter
    prev  *sudog          // Previous waiter
    // ...
}
```

### Unbuffered Channel Send

```
Sender:                          Receiver:
┌─────────────────┐              ┌─────────────────┐
│ ch <- value     │              │ v := <-ch       │
└────────┬────────┘              └────────┬────────┘
         │                                │
         ▼                                ▼
┌─────────────────────────────────────────────────┐
│              Check for receiver                  │
├─────────────────────────────────────────────────┤
│ If receiver waiting:                             │
│   1. Copy value directly to receiver's stack    │
│   2. Wake receiver                              │
│   3. Continue                                   │
├─────────────────────────────────────────────────┤
│ If no receiver:                                  │
│   1. Add sender to sendq                        │
│   2. Park goroutine                             │
│   3. Wait to be woken                           │
└─────────────────────────────────────────────────┘
```

### Buffered Channel

```
Buffer (circular queue):
┌───┬───┬───┬───┬───┐
│ 0 │ 1 │ 2 │ 3 │ 4 │  dataqsiz=5
└───┴───┴───┴───┴───┘
      ▲           ▲
    recvx       sendx

Send:
- If buf not full: write to buf[sendx], sendx++
- If buf full: block, add to sendq

Receive:
- If buf not empty: read from buf[recvx], recvx++
- If buf empty: block, add to recvq
```

### Select Implementation

```go
// Select uses a special algorithm to prevent deadlocks

// 1. Lock all channels in address order (prevents deadlock)
// 2. Check each case for readiness
// 3. If none ready and no default: park on all channels
// 4. When woken: determine which case fired
// 5. Unlock all channels

// Random selection among ready cases provides fairness
```

---

## 11.7 Interface Internals

### Interface Representation

**Interview Question:** *"How are interfaces represented in memory?"*

```go
// Non-empty interface (has methods)
type iface struct {
    tab  *itab          // Type information + method table
    data unsafe.Pointer // Pointer to actual data
}

// Empty interface (interface{} / any)
type eface struct {
    _type *_type        // Type information only
    data  unsafe.Pointer // Pointer to actual data
}

// itab - Interface table
type itab struct {
    inter *interfacetype  // Interface type
    _type *_type          // Concrete type
    hash  uint32          // Copy of _type.hash (for switch)
    _     [4]byte         // Padding
    fun   [1]uintptr      // Method table (variable size)
}
```

### Interface Assignment

```go
type Writer interface {
    Write([]byte) (int, error)
}

type MyWriter struct{}
func (m *MyWriter) Write(b []byte) (int, error) { return len(b), nil }

var w Writer = &MyWriter{}

// In memory:
// w.tab -> itab {
//     inter: *Writer interface type
//     _type: *MyWriter type
//     fun:   [&MyWriter.Write]
// }
// w.data -> &MyWriter{}
```

### Type Assertion Cost

```go
// Type assertion: v, ok := i.(T)

// Implementation:
// 1. Check if itab._type matches T
// 2. If match: return data, true
// 3. If no match: return zero, false

// For concrete types: O(1) - direct comparison
// For interfaces: O(n) - check method set

// itab caching: Go caches itab lookups
// First assertion: Hash table lookup
// Subsequent: Cached
```

### Interface Nil Gotcha

```go
type MyError struct{}
func (e *MyError) Error() string { return "error" }

func returnsError() error {
    var err *MyError = nil
    return err  // Returns non-nil interface!
}

func main() {
    err := returnsError()
    fmt.Println(err == nil)  // false!
    
    // err = iface{tab: *itab for MyError, data: nil}
    // Interface is not nil because tab is set!
}

// Correct approach:
func returnsError() error {
    var err *MyError = nil
    if err == nil {
        return nil  // Return nil interface
    }
    return err
}
```

---

## 11.8 Reflection

### reflect Package Basics

**Interview Question:** *"How does reflection work in Go?"*

```go
import "reflect"

func inspect(x interface{}) {
    t := reflect.TypeOf(x)   // Type information
    v := reflect.ValueOf(x)  // Value information
    
    fmt.Println("Type:", t)
    fmt.Println("Kind:", t.Kind())
    fmt.Println("Value:", v)
    
    // For structs
    if t.Kind() == reflect.Struct {
        for i := 0; i < t.NumField(); i++ {
            field := t.Field(i)
            value := v.Field(i)
            fmt.Printf("%s: %v\n", field.Name, value)
        }
    }
}

func main() {
    type User struct {
        Name string
        Age  int
    }
    inspect(User{Name: "Alice", Age: 30})
}
```

### Modifying Values

```go
func modify(x interface{}) {
    v := reflect.ValueOf(x)
    
    // Must pass pointer to modify
    if v.Kind() != reflect.Ptr {
        panic("must pass pointer")
    }
    
    // Get element (dereference)
    v = v.Elem()
    
    if v.Kind() == reflect.Struct {
        field := v.FieldByName("Name")
        if field.CanSet() {
            field.SetString("Bob")
        }
    }
}

func main() {
    type User struct {
        Name string
    }
    u := User{Name: "Alice"}
    modify(&u)
    fmt.Println(u.Name)  // "Bob"
}
```

### Reflection Performance

```go
// Reflection is slow! Avoid in hot paths.

// Direct call: ~1ns
func direct(u User) string { return u.Name }

// Reflection: ~100ns (100x slower!)
func reflected(x interface{}) string {
    return reflect.ValueOf(x).FieldByName("Name").String()
}

// Why slow?
// - Type checks at runtime
// - Memory allocations
// - Cannot be inlined
// - No compile-time optimization
```

### Common Use Cases

```go
// 1. Serialization (encoding/json)
json.Marshal(v)  // Uses reflection internally

// 2. ORM field mapping
db.Find(&users)  // Maps columns to struct fields

// 3. Dependency injection
container.Resolve(&service)

// 4. Testing/mocking
mock.On("Method").Return(value)

// 5. Generic utilities (pre-generics)
func contains(slice interface{}, item interface{}) bool
```

---

## 11.9 Runtime Debugging

### GODEBUG Options

```bash
# GC tracing
GODEBUG=gctrace=1 ./app
# gc 1 @0.009s 2%: 0.12+0.34+0.01 ms clock...

# Scheduler tracing
GODEBUG=schedtrace=1000 ./app
# SCHED 1000ms: gomaxprocs=8 idleprocs=7...

# Memory allocation tracing
GODEBUG=allocfreetrace=1 ./app

# Invalidate cached pointers (debugging)
GODEBUG=invalidptr=1 ./app

# Async preemption (disable for debugging)
GODEBUG=asyncpreemptoff=1 ./app
```

### Runtime Functions for Debugging

```go
import "runtime"

// Get goroutine ID (unofficial, don't rely on it!)
func getGoroutineID() uint64 {
    b := make([]byte, 64)
    b = b[:runtime.Stack(b, false)]
    // Parse "goroutine 123 [...]"
    // ...
}

// Count goroutines
fmt.Println(runtime.NumGoroutine())

// Force GC
runtime.GC()

// Memory stats
var m runtime.MemStats
runtime.ReadMemStats(&m)

// Set finalizer (debugging only)
runtime.SetFinalizer(obj, func(o *MyType) {
    fmt.Println("Object finalized")
})
```

### Stack Traces

```go
import "runtime/debug"

// Print current stack
debug.PrintStack()

// Get stack as string
stack := debug.Stack()

// All goroutines
buf := make([]byte, 1024*1024)
n := runtime.Stack(buf, true)  // true = all goroutines
fmt.Println(string(buf[:n]))
```

---

## Interview Questions

### Beginner Level

1. **Q:** What is included in Go's runtime?
   **A:** Scheduler, memory allocator, garbage collector, stack manager, channel implementation, defer/panic/recover.

2. **Q:** How big is the initial goroutine stack?
   **A:** 2KB. It grows dynamically up to 1GB.

3. **Q:** What does GOGC=100 mean?
   **A:** GC runs when heap reaches 100% growth over live heap. If live heap is 100MB, GC triggers at 200MB.

### Intermediate Level

4. **Q:** Explain tricolor garbage collection.
   **A:** Objects are white (potential garbage), gray (reachable, unscanned children), or black (reachable, scanned). Process gray until none remain. White objects are garbage.

5. **Q:** What is the write barrier?
   **A:** Code that runs on pointer writes during GC to maintain the tricolor invariant (no black->white pointers).

6. **Q:** How do channels block goroutines?
   **A:** Blocked goroutines added to sendq/recvq wait queues. When partner arrives, goroutine is woken directly.

### Advanced Level

7. **Q:** Explain Go's memory allocator hierarchy.
   **A:** mcache (per-P, no lock) → mcentral (per-size-class, locked) → mheap (global, locked). Objects grouped by size class.

8. **Q:** Why can an interface holding nil not equal nil?
   **A:** Interface is `(type, value)` pair. When type is set but value is nil, interface isn't nil because type metadata exists.

9. **Q:** How does the scheduler implement work stealing?
   **A:** Idle P steals half of runqueue from random busy P. Global queue checked periodically. Prevents load imbalance.

---

## Summary

| Topic | Key Points |
|-------|------------|
| Runtime | Scheduler + Allocator + GC + Stacks in every binary |
| Allocator | mcache → mcentral → mheap, size classes, tiny allocator |
| Stacks | 2KB initial, grows 2x, contiguous, copied on growth |
| GC | Tricolor mark-sweep, concurrent, write barrier, STW phases minimal |
| Scheduler | GMP model, work stealing, local/global queues, preemption |
| Channels | hchan struct, circular buffer, wait queues, lock per channel |
| Interfaces | iface (methods) vs eface (empty), itab caching, nil gotcha |
| Reflection | TypeOf/ValueOf, slow (~100x), avoid in hot paths |
| Debugging | GODEBUG, gctrace, schedtrace, runtime/debug |

**Next Phase:** [Phase 12 — Modern Go Features](../Phase_12/Phase_12.md)

