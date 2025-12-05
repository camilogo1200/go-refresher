# 🔠 Phase 2: The Type System Deep Dive

[← Back to Main Roadmap](../README.md) | [← Previous: Phase 1](../Phase_1/Phase_1.md)

---

**Objective:** Master Go's static, strongly-typed system and understand the memory implications of every type choice.

**Reference:** [Go Language Specification - Types](https://go.dev/ref/spec#Types)

**Prerequisites:** Phase 0-1

**Estimated Duration:** 3-4 weeks

---

## 📋 Table of Contents

1. [Type System Fundamentals](#21-type-system-fundamentals)
2. [Basic Types (Primitives)](#22-basic-types-primitives)
3. [Composite Types](#23-composite-types)
4. [Pointer Types](#24-pointer-types)
5. [Function Types](#25-function-types)
6. [Methods & Receivers](#26-methods--receivers)
7. [Interface Types](#27-interface-types-critical-section)
8. [Generics](#28-generics-type-parameters--go-118)
9. [Interview Questions](#interview-questions)

---

## 2.1 Type System Fundamentals

### Static vs. Dynamic Typing

**Interview Question:** *"Is Go statically or dynamically typed? What are the implications?"*

Go is **statically typed** — types are checked at compile time:

```go
var x int = 42
x = "hello"  // Compile error: cannot use "hello" (type string) as type int
```

**Benefits:**
- Errors caught at compile time
- Better IDE support (autocomplete, refactoring)
- Performance (no runtime type checks)
- Self-documenting code

### Strong vs. Weak Typing

**Interview Question:** *"Why can't I add an int and a float64 in Go without conversion?"*

Go is **strongly typed** — no implicit conversions:

```go
var i int = 42
var f float64 = 3.14

// sum := i + f  // ERROR: mismatched types
sum := float64(i) + f  // OK: explicit conversion
```

**Even between related types:**
```go
type UserID int
type ProductID int

var uid UserID = 1
var pid ProductID = 2

// uid = pid  // ERROR: cannot use ProductID as UserID
uid = UserID(pid)  // OK: explicit conversion
```

### Structural Typing (for Interfaces)

**Interview Question:** *"Does Go use nominal or structural typing?"*

Go uses **structural typing** for interfaces — no `implements` keyword:

```go
type Writer interface {
    Write([]byte) (int, error)
}

// MyBuffer implements Writer implicitly by having the Write method
type MyBuffer struct {
    data []byte
}

func (b *MyBuffer) Write(p []byte) (int, error) {
    b.data = append(b.data, p...)
    return len(p), nil
}

// No declaration needed - MyBuffer satisfies Writer automatically
var w Writer = &MyBuffer{}
```

### Named vs. Unnamed Types

**Interview Question:** *"What is an underlying type? Why does it matter?"*

```go
// Named types
type MyInt int           // MyInt is a named type
type MySlice []string    // MySlice is a named type

// Unnamed types (type literals)
var x []int              // []int is unnamed
var y map[string]bool    // map[string]bool is unnamed

// Underlying type determination
type MyInt int           // Underlying: int
type YourInt MyInt       // Underlying: int (recursive!)
type StringSlice []string // Underlying: []string
```

**Type Identity Rules:**
```go
type A int
type B int

var a A = 1
var b B = 2

// a = b  // ERROR: different named types
a = A(b)  // OK: explicit conversion (same underlying type)
```

### Type Definitions vs. Type Aliases

**Interview Question:** *"What's the difference between `type A B` and `type A = B`?"*

```go
// Type definition - creates NEW type
type UserID int

// Type alias - creates another name for SAME type  
type ID = int

func main() {
    var uid UserID = 1
    var id ID = 2
    var i int = 3
    
    // uid = i    // ERROR: different types
    // uid = id   // ERROR: UserID vs int (ID is just int)
    
    id = i       // OK: ID is exactly int
    i = id       // OK: same type
}
```

**Use cases for aliases:**
- Gradual refactoring (`type OldName = NewName`)
- Shortening long type names
- Cross-package type exposure

### Example: Type System in Action

```go
package main

import "fmt"

// Type definition with methods
type Celsius float64
type Fahrenheit float64

func (c Celsius) ToFahrenheit() Fahrenheit {
    return Fahrenheit(c*9/5 + 32)
}

func (f Fahrenheit) ToCelsius() Celsius {
    return Celsius((f - 32) * 5 / 9)
}

func main() {
    boiling := Celsius(100)
    fmt.Printf("%v°C = %v°F\n", boiling, boiling.ToFahrenheit())
    
    // Type safety prevents accidental mixing
    // boiling = Fahrenheit(212)  // ERROR!
    boiling = Fahrenheit(212).ToCelsius()  // OK
}
```

---

## 2.2 Basic Types (Primitives)

### Boolean Type

**Interview Question:** *"Can you convert between bool and int in Go?"*

```go
var b bool = true
// var i int = int(b)  // ERROR: cannot convert bool to int

// Must use explicit conversion
func boolToInt(b bool) int {
    if b {
        return 1
    }
    return 0
}
```

**Zero value:** `false`

### Numeric Types

#### Integer Types

| Type | Size | Range |
|------|------|-------|
| `int8` | 8 bits | -128 to 127 |
| `int16` | 16 bits | -32,768 to 32,767 |
| `int32` | 32 bits | -2.1B to 2.1B |
| `int64` | 64 bits | ±9.2 quintillion |
| `uint8` (byte) | 8 bits | 0 to 255 |
| `uint16` | 16 bits | 0 to 65,535 |
| `uint32` | 32 bits | 0 to 4.3B |
| `uint64` | 64 bits | 0 to 18.4 quintillion |
| `int` | 32 or 64 | Platform-dependent |
| `uint` | 32 or 64 | Platform-dependent |
| `uintptr` | 32 or 64 | For pointer arithmetic |

**Interview Question:** *"When should you use `int` vs. `int64`?"*

```go
// Use int for:
// - Loop counters
// - Array/slice indices
// - General-purpose integers

// Use int64/int32 for:
// - API boundaries (serialization)
// - Database columns
// - When size matters for storage

// CAUTION: int size varies by platform!
var x int = 1 << 40  // OK on 64-bit, overflow on 32-bit
```

#### Overflow Behavior

**Interview Question:** *"What happens when an integer overflows in Go?"*

```go
var u uint8 = 255
u++
fmt.Println(u)  // 0 (wraps around, no error!)

var i int8 = 127
i++
fmt.Println(i)  // -128 (wraps around)

// Detection requires explicit checks
func addSafe(a, b int) (int, error) {
    if a > 0 && b > math.MaxInt-a {
        return 0, errors.New("overflow")
    }
    return a + b, nil
}
```

### String Type (Deep Dive)

**Interview Question:** *"What is a string's internal representation in Go?"*

A string in Go is a **read-only slice of bytes** with a 2-word header:

```go
// Internal structure (conceptual)
type stringHeader struct {
    Data uintptr  // Pointer to byte array
    Len  int      // Length in bytes
}
```

#### String Internals

```go
s := "Hello, 世界"

// len() returns BYTES, not characters!
fmt.Println(len(s))  // 13 (5 + 2 + 6 bytes for UTF-8)

// Indexing returns BYTES
fmt.Println(s[0])  // 72 ('H')
fmt.Println(s[7])  // 228 (first byte of '世')

// Rune count
fmt.Println(utf8.RuneCountInString(s))  // 9 characters

// Correct iteration for Unicode
for i, r := range s {
    fmt.Printf("byte %d: %c (U+%04X)\n", i, r, r)
}
```

#### String Immutability

**Interview Question:** *"Are strings mutable in Go? How do you modify a string?"*

```go
s := "hello"
// s[0] = 'H'  // ERROR: cannot assign to s[0]

// Must create new string
s = "H" + s[1:]  // OK but inefficient

// Or convert to []byte
b := []byte(s)
b[0] = 'H'
s = string(b)  // Creates new string
```

#### String Performance

**Interview Question:** *"Why is string concatenation in a loop inefficient?"*

```go
// SLOW: O(n²) - creates new string each iteration
func slowConcat(items []string) string {
    result := ""
    for _, item := range items {
        result += item  // Allocates new string!
    }
    return result
}

// FAST: O(n) - uses buffer
func fastConcat(items []string) string {
    var builder strings.Builder
    for _, item := range items {
        builder.WriteString(item)  // Appends to buffer
    }
    return builder.String()
}
```

### Rune Type & Unicode

**Interview Question:** *"What is a rune? How is it different from a byte?"*

```go
// rune is alias for int32, represents Unicode code point
var r rune = '日'  // U+65E5

// byte is alias for uint8, represents single byte
var b byte = 'A'  // 65

// String iteration
s := "日本語"

// Byte iteration (WRONG for Unicode)
for i := 0; i < len(s); i++ {
    fmt.Printf("%d: %x\n", i, s[i])  // Individual bytes
}

// Rune iteration (CORRECT)
for i, r := range s {
    fmt.Printf("%d: %c\n", i, r)  // Complete characters
}
// Output:
// 0: 日
// 3: 本
// 6: 語
```

### The `strings` Package

**Interview Question:** *"Name the most commonly used `strings` package functions."*

| Category | Functions |
|----------|-----------|
| Search | `Contains`, `Index`, `LastIndex`, `Count` |
| Prefix/Suffix | `HasPrefix`, `HasSuffix` |
| Case | `ToUpper`, `ToLower`, `EqualFold` |
| Trim | `TrimSpace`, `Trim`, `TrimPrefix`, `TrimSuffix` |
| Split/Join | `Split`, `Fields`, `Join` |
| Replace | `Replace`, `ReplaceAll` |
| Build | `Builder.WriteString`, `Builder.String` |

```go
// Common patterns
s := "  Hello, World!  "

strings.TrimSpace(s)           // "Hello, World!"
strings.ToLower(s)             // "  hello, world!  "
strings.Contains(s, "World")   // true
strings.Split("a,b,c", ",")    // []string{"a", "b", "c"}
strings.Join([]string{"a","b"}, "-")  // "a-b"
strings.ReplaceAll(s, "l", "L") // "  HeLLo, WorLd!  "

// Case-insensitive comparison
strings.EqualFold("GO", "go")  // true
```

### The `strconv` Package

**Interview Question:** *"How do you convert between strings and numbers in Go?"*

```go
// String to number
i, err := strconv.Atoi("42")           // int
i64, err := strconv.ParseInt("42", 10, 64)  // int64
f, err := strconv.ParseFloat("3.14", 64)    // float64
b, err := strconv.ParseBool("true")         // bool

// Number to string
s := strconv.Itoa(42)                    // "42"
s = strconv.FormatInt(42, 16)            // "2a" (hex)
s = strconv.FormatFloat(3.14, 'f', 2, 64) // "3.14"
s = strconv.FormatBool(true)             // "true"

// Error handling is critical!
if i, err := strconv.Atoi("not a number"); err != nil {
    fmt.Println("Parse error:", err)
}
```

---

## 2.3 Composite Types

### Arrays

**Interview Question:** *"What's unique about arrays in Go compared to most languages?"*

Arrays in Go:
1. Have **fixed size** (part of the type)
2. Are **value types** (assignment copies)

```go
// Size is part of the type
var a [3]int
var b [4]int
// a = b  // ERROR: [3]int and [4]int are different types!

// Value semantics - copies entire array
a := [3]int{1, 2, 3}
b := a        // Copy!
b[0] = 100
fmt.Println(a[0])  // 1 (unchanged)

// Array literal with inferred size
c := [...]int{1, 2, 3, 4, 5}  // [5]int

// Arrays are comparable
x := [3]int{1, 2, 3}
y := [3]int{1, 2, 3}
fmt.Println(x == y)  // true
```

**When to use arrays:**
- Fixed-size data (RGB pixels, coordinates)
- Hash digests (`[32]byte`)
- When you need value semantics

### Slices (The Workhorse)

**Interview Question:** *"Explain the internal structure of a slice. What are length and capacity?"*

A slice is a **view into an array** with a 3-word header:

```go
// Internal structure (conceptual)
type sliceHeader struct {
    Data uintptr  // Pointer to backing array element
    Len  int      // Number of elements
    Cap  int      // Capacity (elements until end of backing array)
}
```

#### Slice Creation

```go
// Literal
s1 := []int{1, 2, 3}

// make with length
s2 := make([]int, 5)  // len=5, cap=5, zeroed

// make with length and capacity
s3 := make([]int, 5, 10)  // len=5, cap=10

// Slicing an array or slice
arr := [5]int{1, 2, 3, 4, 5}
s4 := arr[1:4]  // []int{2, 3, 4}, len=3, cap=4

// Full slice expression (limits capacity)
s5 := arr[1:4:4]  // len=3, cap=3 (can't grow beyond)
```

#### Slice Gotchas

**Interview Question:** *"What are the common pitfalls when working with slices?"*

**1. Sharing backing arrays:**
```go
original := []int{1, 2, 3, 4, 5}
slice := original[1:3]  // {2, 3}

slice[0] = 100
fmt.Println(original)  // [1 100 3 4 5] - Modified!
```

**2. Append may or may not create new array:**
```go
s := make([]int, 3, 5)  // len=3, cap=5
s[0], s[1], s[2] = 1, 2, 3

s2 := append(s, 4)  // Fits in capacity
fmt.Println(&s[0] == &s2[0])  // true (same backing array!)

s3 := append(s2, 5, 6)  // Exceeds capacity
fmt.Println(&s2[0] == &s3[0])  // false (new array allocated)
```

**3. Memory leak from large backing array:**
```go
func getFirst(data []byte) []byte {
    return data[:1]  // Still references entire backing array!
}

// Fix: copy the data
func getFirstSafe(data []byte) []byte {
    result := make([]byte, 1)
    copy(result, data[:1])
    return result
}
```

**4. Nil slice vs. empty slice:**
```go
var nilSlice []int      // nil, len=0, cap=0
emptySlice := []int{}   // not nil, len=0, cap=0
makeSlice := make([]int, 0)  // not nil, len=0, cap=0

// Both work the same for most operations
fmt.Println(len(nilSlice) == len(emptySlice))  // true
nilSlice = append(nilSlice, 1)  // Works!

// But different in JSON
json.Marshal(nilSlice)   // "null"
json.Marshal(emptySlice) // "[]"
```

#### Append Mechanics

**Interview Question:** *"How does slice growth work? What's the growth factor?"*

```go
s := []int{}
for i := 0; i < 10; i++ {
    s = append(s, i)
    fmt.Printf("len=%d cap=%d\n", len(s), cap(s))
}
// len=1 cap=1
// len=2 cap=2
// len=3 cap=4   (doubled)
// len=4 cap=4
// len=5 cap=8   (doubled)
// ...
```

**Growth strategy (Go 1.18+):**
- Small slices: double capacity
- Large slices (>256): grow by ~25%

### Maps

**Interview Question:** *"What types can be used as map keys? What happens if you read from a nil map?"*

#### Map Fundamentals

```go
// Creation
m1 := make(map[string]int)
m2 := map[string]int{"a": 1, "b": 2}

// Read (returns zero value if missing)
v := m["key"]  // 0 if not present

// Read with existence check
v, ok := m["key"]
if !ok {
    fmt.Println("key not found")
}

// Write
m["key"] = 42

// Delete (safe on nil and missing)
delete(m, "key")

// Iteration (random order!)
for k, v := range m {
    fmt.Println(k, v)
}
```

#### Map Key Constraints

**Interview Question:** *"Can you use a slice as a map key?"*

Map keys must be **comparable** (support `==`):

| Valid Keys | Invalid Keys |
|------------|--------------|
| `int`, `string`, `float64` | `[]int` (slices) |
| `*MyType` (pointers) | `map[K]V` (maps) |
| `[3]int` (arrays) | `func()` (functions) |
| `struct` (if all fields comparable) | `struct` with slice field |

```go
// Array as key (works!)
type Point [2]int
positions := map[Point]string{
    {0, 0}: "origin",
    {1, 1}: "diagonal",
}

// Struct as key (if comparable)
type Person struct {
    Name string
    Age  int
}
people := map[Person]bool{
    {"Alice", 30}: true,
}
```

#### Nil Map Behavior

```go
var m map[string]int  // nil

// Read is safe (returns zero value)
v := m["key"]  // v = 0

// Write PANICS!
// m["key"] = 1  // panic: assignment to entry in nil map

// Delete is safe
delete(m, "key")  // No-op

// Iteration is safe
for k, v := range m {  // No iterations
    fmt.Println(k, v)
}
```

#### Map Internals (Interview Focus)

**Interview Question:** *"How is a Go map implemented internally?"*

```
Map structure:
- Header with bucket count
- Array of buckets
- Each bucket holds 8 key/value pairs
- Overflow buckets for collision handling

Bucket structure:
+------------------+
| tophash[8]       |  <- High byte of hash for quick comparison
| keys[8]          |
| values[8]        |
| overflow pointer |
+------------------+
```

**Key properties:**
- Load factor triggers growth (~6.5 elements/bucket)
- Iteration order randomized intentionally
- Not safe for concurrent access (use `sync.Map` or mutex)

### Structs

**Interview Question:** *"How does struct embedding differ from inheritance?"*

#### Struct Basics

```go
type Person struct {
    Name string
    Age  int
}

// Creating structs
p1 := Person{Name: "Alice", Age: 30}
p2 := Person{"Bob", 25}  // Positional (fragile, avoid)
p3 := Person{}  // Zero value: {"", 0}

// Accessing fields
fmt.Println(p1.Name)

// Pointer to struct
pp := &Person{Name: "Charlie", Age: 35}
fmt.Println(pp.Name)  // Automatic dereference
```

#### Memory Layout & Padding

**Interview Question:** *"How can you optimize struct memory layout?"*

```go
// Inefficient layout (24 bytes with padding)
type Inefficient struct {
    a bool    // 1 byte + 7 padding
    b int64   // 8 bytes
    c bool    // 1 byte + 7 padding
}

// Efficient layout (16 bytes)
type Efficient struct {
    b int64   // 8 bytes
    a bool    // 1 byte
    c bool    // 1 byte + 6 padding
}

fmt.Println(unsafe.Sizeof(Inefficient{}))  // 24
fmt.Println(unsafe.Sizeof(Efficient{}))    // 16
```

**Rule:** Order fields from largest to smallest to minimize padding.

#### Struct Embedding (Composition)

```go
type Address struct {
    Street string
    City   string
}

type Employee struct {
    Person          // Embedded (anonymous field)
    Address         // Embedded
    EmployeeID int
}

func main() {
    e := Employee{
        Person:     Person{Name: "Alice", Age: 30},
        Address:    Address{Street: "123 Main", City: "NYC"},
        EmployeeID: 12345,
    }
    
    // Promoted fields
    fmt.Println(e.Name)    // From Person
    fmt.Println(e.Street)  // From Address
    
    // Full path also works
    fmt.Println(e.Person.Name)
}
```

**Key difference from inheritance:**
- No polymorphism (Employee is NOT a Person)
- Field promotion is syntactic sugar
- Embedded type's methods are promoted

#### Struct Tags

**Interview Question:** *"What are struct tags and how are they used?"*

```go
type User struct {
    ID        int    `json:"id" db:"user_id"`
    Email     string `json:"email" validate:"required,email"`
    Password  string `json:"-"`  // Omit from JSON
    CreatedAt time.Time `json:"created_at,omitempty"`
}

// Tags are accessed via reflection
t := reflect.TypeOf(User{})
field, _ := t.FieldByName("Email")
fmt.Println(field.Tag.Get("json"))      // "email"
fmt.Println(field.Tag.Get("validate"))  // "required,email"
```

#### Empty Struct

**Interview Question:** *"What is the size of an empty struct? When would you use it?"*

```go
type empty struct{}

fmt.Println(unsafe.Sizeof(empty{}))  // 0 bytes!

// Use cases:

// 1. Set implementation
set := make(map[string]struct{})
set["item"] = struct{}{}
if _, exists := set["item"]; exists {
    fmt.Println("Found!")
}

// 2. Signal channel (no data needed)
done := make(chan struct{})
go func() {
    // work...
    close(done)  // Signal completion
}()
<-done
```

---

## 2.4 Pointer Types

**Interview Question:** *"When should you use pointers vs. values in Go?"*

### Pointer Basics

```go
x := 42
p := &x      // p is *int, points to x
fmt.Println(*p)  // 42 (dereference)

*p = 100
fmt.Println(x)   // 100

// Nil pointer
var nilPtr *int
// fmt.Println(*nilPtr)  // PANIC: nil pointer dereference

// Check before use
if nilPtr != nil {
    fmt.Println(*nilPtr)
}
```

### No Pointer Arithmetic

```go
arr := [3]int{1, 2, 3}
p := &arr[0]

// In C: p++
// In Go: NOT ALLOWED (without unsafe)

// Use slice or index instead
for i := range arr {
    fmt.Println(arr[i])
}
```

### When to Use Pointers

| Use Pointer | Use Value |
|-------------|-----------|
| Need to modify the value | Small, immutable data |
| Large struct (>64 bytes) | Small struct |
| Implementing interfaces with mutation | Thread-safe by copy |
| Optional value (nil = absent) | Always present |
| Shared mutable state | Independent copies |

### Example: Pointer Patterns

```go
// Pattern 1: Optional return
func findUser(id int) *User {
    // ... search
    if notFound {
        return nil
    }
    return &user
}

// Pattern 2: Constructor returning pointer
func NewUser(name string) *User {
    return &User{
        Name:      name,
        CreatedAt: time.Now(),
    }
}

// Pattern 3: Modify receiver
func (u *User) SetName(name string) {
    u.Name = name  // Modifies original
}

// Pattern 4: Avoid copying large struct
func processLargeData(data *LargeStruct) {
    // Work with data without copying
}
```

---

## 2.5 Function Types

**Interview Question:** *"Are functions first-class citizens in Go? What does that mean?"*

### Functions as Values

```go
// Function type
type Operation func(int, int) int

// Function as variable
var add Operation = func(a, b int) int {
    return a + b
}

// Function as parameter
func apply(op Operation, a, b int) int {
    return op(a, b)
}

result := apply(add, 3, 4)  // 7
```

### Anonymous Functions

```go
// Inline definition
result := func(x int) int {
    return x * x
}(5)  // Immediately invoked

// Stored for later
square := func(x int) int {
    return x * x
}
fmt.Println(square(5))  // 25
```

### Closures

**Interview Question:** *"What is a closure? How does variable capture work?"*

```go
func counter() func() int {
    count := 0  // Captured by closure
    return func() int {
        count++  // Modifies captured variable
        return count
    }
}

c := counter()
fmt.Println(c())  // 1
fmt.Println(c())  // 2
fmt.Println(c())  // 3

// Each call to counter() creates new closure
c2 := counter()
fmt.Println(c2())  // 1 (independent)
```

**Capture is by reference:**
```go
funcs := make([]func(), 3)
for i := 0; i < 3; i++ {
    funcs[i] = func() {
        fmt.Println(i)  // Captures reference to i
    }
}
// In Go < 1.22: prints 3, 3, 3
// In Go >= 1.22: prints 0, 1, 2 (fixed!)
```

### Higher-Order Functions

```go
// Map function (generic style)
func Map[T, U any](slice []T, f func(T) U) []U {
    result := make([]U, len(slice))
    for i, v := range slice {
        result[i] = f(v)
    }
    return result
}

// Filter function
func Filter[T any](slice []T, predicate func(T) bool) []T {
    var result []T
    for _, v := range slice {
        if predicate(v) {
            result = append(result, v)
        }
    }
    return result
}

// Usage
numbers := []int{1, 2, 3, 4, 5}
doubled := Map(numbers, func(n int) int { return n * 2 })
evens := Filter(numbers, func(n int) bool { return n%2 == 0 })
```

---

## 2.6 Methods & Receivers

**Interview Question:** *"What's the difference between a value receiver and a pointer receiver?"*

### Method Declaration

```go
type Rectangle struct {
    Width, Height float64
}

// Value receiver - operates on copy
func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

// Pointer receiver - operates on original
func (r *Rectangle) Scale(factor float64) {
    r.Width *= factor
    r.Height *= factor
}
```

### Value vs. Pointer Receivers

| Aspect | Value Receiver | Pointer Receiver |
|--------|----------------|------------------|
| Modification | Cannot modify original | Can modify original |
| Copy | Copies entire value | Copies pointer (8 bytes) |
| Nil safety | Cannot be nil | Can handle nil |
| Interface | Type and *Type satisfy | Only *Type satisfies |

```go
r := Rectangle{10, 5}

// Value receiver - r unchanged
area := r.Area()

// Pointer receiver - r modified
r.Scale(2)  // Go automatically takes address: (&r).Scale(2)

// Explicit pointer
pr := &Rectangle{10, 5}
pr.Scale(2)  // Direct pointer call
area = pr.Area()  // Go automatically dereferences: (*pr).Area()
```

### Method Sets (Critical for Interfaces)

**Interview Question:** *"Explain method sets. Why can't a value satisfy an interface requiring pointer receiver methods?"*

```go
type Resizer interface {
    Resize(factor float64)
}

type Shape struct {
    Size float64
}

func (s *Shape) Resize(factor float64) {
    s.Size *= factor
}

// Method sets:
// - Type Shape: no methods (Resize has pointer receiver)
// - Type *Shape: Resize method

var r Resizer

// r = Shape{}   // ERROR: Shape doesn't implement Resizer
r = &Shape{}     // OK: *Shape implements Resizer
```

### Nil Receivers

**Interview Question:** *"Can you call a method on a nil pointer in Go?"*

```go
type List struct {
    Value int
    Next  *List
}

func (l *List) Length() int {
    if l == nil {
        return 0  // Handle nil gracefully
    }
    return 1 + l.Next.Length()
}

var list *List = nil
fmt.Println(list.Length())  // 0 (works!)
```

### Example: Complete Method Pattern

```go
package main

import "fmt"

type Account struct {
    balance int
}

// Constructor
func NewAccount(initial int) *Account {
    return &Account{balance: initial}
}

// Value receiver - read-only
func (a Account) Balance() int {
    return a.balance
}

// Pointer receiver - mutation
func (a *Account) Deposit(amount int) {
    if amount > 0 {
        a.balance += amount
    }
}

func (a *Account) Withdraw(amount int) error {
    if amount > a.balance {
        return fmt.Errorf("insufficient funds")
    }
    a.balance -= amount
    return nil
}

func main() {
    acc := NewAccount(100)
    acc.Deposit(50)
    acc.Withdraw(30)
    fmt.Println(acc.Balance())  // 120
}
```

---

## 2.7 Interface Types (Critical Section)

**Interview Question:** *"How do interfaces work in Go? What's the nil interface gotcha?"*

### Interface Fundamentals

```go
// Interface definition
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// Interface composition
type ReadWriter interface {
    Reader
    Writer
}
```

### Implicit Implementation

```go
// MyBuffer implements Reader and Writer without declaring it
type MyBuffer struct {
    data []byte
}

func (b *MyBuffer) Read(p []byte) (n int, err error) {
    n = copy(p, b.data)
    return n, io.EOF
}

func (b *MyBuffer) Write(p []byte) (n int, err error) {
    b.data = append(b.data, p...)
    return len(p), nil
}

// Compile-time interface check
var _ Reader = (*MyBuffer)(nil)
var _ Writer = (*MyBuffer)(nil)
```

### The Nil Interface Gotcha

**Interview Question:** *"What is the nil interface gotcha? How can you avoid it?"*

```go
type error interface {
    Error() string
}

type MyError struct {
    Message string
}

func (e *MyError) Error() string {
    return e.Message
}

func getError(fail bool) error {
    var err *MyError = nil
    if fail {
        err = &MyError{"failed"}
    }
    return err  // Returns non-nil interface with nil value!
}

func main() {
    err := getError(false)
    
    // GOTCHA!
    if err != nil {
        fmt.Println("error:", err)  // This RUNS!
    }
    
    // The interface value is:
    // {type: *MyError, value: nil}
    // This is NOT equal to nil interface: {type: nil, value: nil}
}
```

**Fix:**
```go
func getError(fail bool) error {
    if fail {
        return &MyError{"failed"}
    }
    return nil  // Return nil interface explicitly
}
```

### Interface Internals

**Interview Question:** *"How are interfaces represented at runtime?"*

```go
// Interface with methods (iface)
type iface struct {
    tab  *itab   // Type info + method table
    data unsafe.Pointer  // Pointer to actual value
}

// Empty interface (eface)
type eface struct {
    _type *_type  // Type info
    data  unsafe.Pointer
}
```

```
Interface variable holding *MyBuffer:

+-------------------+
| itab pointer -----|---> +------------------+
+-------------------+     | inter: Reader    |
| data pointer -----|     | type: *MyBuffer  |
+-------------------+     | methods[0]: Read |
         |                +------------------+
         v
    +----------+
    | MyBuffer |
    | data     |
    +----------+
```

### Interface Design Principles

**Interview Question:** *"What does 'accept interfaces, return structs' mean?"*

```go
// GOOD: Accept interface
func ProcessData(r io.Reader) error {
    // Can accept *os.File, *bytes.Buffer, *http.Response.Body, etc.
    data, err := io.ReadAll(r)
    // ...
}

// GOOD: Return concrete type
func NewBuffer() *bytes.Buffer {
    return &bytes.Buffer{}
}

// BAD: Return interface (loses type information)
func NewReader() io.Reader {
    return &bytes.Buffer{}
}
```

**Interface guidelines:**
1. **Small interfaces:** One method is ideal (`io.Reader`, `fmt.Stringer`)
2. **Consumer-defined:** Define interface where it's used, not implemented
3. **Don't export interfaces prematurely:** Start with concrete types

### Type Assertions & Type Switches

```go
var i interface{} = "hello"

// Type assertion (panics if wrong)
s := i.(string)

// Type assertion with ok (safe)
s, ok := i.(string)
if ok {
    fmt.Println("string:", s)
}

// Type switch
switch v := i.(type) {
case int:
    fmt.Println("int:", v)
case string:
    fmt.Println("string:", v)
case nil:
    fmt.Println("nil")
default:
    fmt.Println("unknown type:", reflect.TypeOf(v))
}
```

---

## 2.8 Generics (Type Parameters) — Go 1.18+

**Interview Question:** *"How do generics work in Go? What are constraints?"*

### Type Parameter Basics

```go
// Generic function
func Min[T constraints.Ordered](a, b T) T {
    if a < b {
        return a
    }
    return b
}

// Usage
fmt.Println(Min(3, 5))       // Type inferred: int
fmt.Println(Min("a", "b"))   // Type inferred: string
fmt.Println(Min[float64](3.14, 2.71))  // Explicit type
```

### Constraints

```go
// any - no restrictions
func Print[T any](v T) {
    fmt.Println(v)
}

// comparable - supports == and !=
func Contains[T comparable](slice []T, target T) bool {
    for _, v := range slice {
        if v == target {
            return true
        }
    }
    return false
}

// Custom constraint
type Number interface {
    int | int64 | float64
}

func Sum[T Number](values []T) T {
    var sum T
    for _, v := range values {
        sum += v
    }
    return sum
}

// Approximation constraint
type Integer interface {
    ~int | ~int64  // Types with underlying int/int64
}

type MyInt int  // Underlying type is int
// MyInt satisfies Integer constraint
```

### Generic Types

```go
// Generic struct
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    item := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return item, true
}

// Usage
intStack := &Stack[int]{}
intStack.Push(1)
intStack.Push(2)
```

### Generic Implementation Details

**Interview Question:** *"How does Go implement generics? What is GC shape stenciling?"*

Go uses **GC shape stenciling**:
- Groups types by "GC shape" (pointer vs. non-pointer)
- Generates specialized code per shape, not per type
- Dictionary passed for type-specific operations

```
Unlike C++ full monomorphization:
- C++: func<int>, func<float>, func<MyType> = 3 copies
- Go: func[pointer-shape], func[non-pointer-shape] = 2 shapes + dictionary
```

**Trade-offs:**
- Less code bloat than C++ templates
- Slight runtime overhead for dictionary access
- Near-equivalent performance in most cases

### Current Limitations

```go
// NO: Method type parameters
type Container[T any] struct{}
// func (c Container[T]) Map[U any](f func(T) U) Container[U] // ERROR!

// NO: Type parameter in method receiver position only
// Type parameters must be on the type, not just the method

// WORKAROUND: Top-level generic function
func Map[T, U any](c Container[T], f func(T) U) Container[U] {
    // ...
}
```

---

## Interview Questions

### Beginner Level

1. **Q:** What is the zero value of a map?
   **A:** `nil`. Reading returns zero value, writing panics.

2. **Q:** Can you use a slice as a map key?
   **A:** No, slices are not comparable. Use arrays or convert to string.

3. **Q:** What does `type A = B` do?
   **A:** Creates a type alias. A and B are the same type.

### Intermediate Level

4. **Q:** Why is `strings.Builder` more efficient than `+` concatenation?
   **A:** `+` creates a new string each time (O(n²) for n concatenations). `Builder` uses an internal buffer and only creates the final string once.

5. **Q:** What is the nil interface gotcha?
   **A:** An interface holding a nil pointer is not equal to nil. `var err error = (*MyError)(nil); err == nil` is `false`.

6. **Q:** When would you use a value receiver vs. pointer receiver?
   **A:** Pointer: need to modify, large struct, consistency. Value: small immutable data, need value semantics, thread-safety through copying.

### Advanced Level

7. **Q:** Explain the slice `append` capacity growth strategy.
   **A:** Small slices double. Large slices (>256 elements) grow by ~25%. Returns new slice that may or may not share backing array.

8. **Q:** How can struct field ordering affect memory usage?
   **A:** Padding for alignment. Order fields from largest to smallest to minimize padding. Use `unsafe.Sizeof` to measure.

9. **Q:** Why can't a value type satisfy an interface that requires pointer receiver methods?
   **A:** The method set of type `T` includes only value receiver methods. The method set of `*T` includes both. If interface requires pointer receiver method, only `*T` satisfies it.

---

## Summary

| Topic | Key Points |
|-------|------------|
| Type System | Static, strong, structural (interfaces) |
| Strings | Immutable, UTF-8, use `strings.Builder` |
| Slices | 3-word header, shared backing array gotchas |
| Maps | Must initialize, keys must be comparable |
| Structs | Value semantics, embedding for composition |
| Interfaces | Implicit, nil gotcha, iface/eface internals |
| Generics | Type parameters, constraints, GC shape stenciling |

**Next Phase:** [Phase 3 — Error Handling](../Phase_3/Phase_3.md)

