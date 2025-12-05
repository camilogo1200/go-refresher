# 📘 Phase 1: Lexical Elements & Language Fundamentals

[← Back to Main Roadmap](../README.md) | [← Previous: Phase 0](../Phase_0/Phase_0.md)

---

**Objective:** Master the atomic building blocks of Go source code as defined in the language specification.

**Reference:** [Go Language Specification - Lexical Elements](https://go.dev/ref/spec#Lexical_elements)

**Prerequisites:** Phase 0 (Environment & Toolchain)

**Estimated Duration:** 2-3 weeks

---

## 📋 Table of Contents

1. [Source Code Representation](#11-source-code-representation)
2. [Identifiers & Naming](#12-identifiers--naming)
3. [Keywords](#13-keywords-the-reserved-25)
4. [Operators & Delimiters](#14-operators--delimiters)
5. [Literals & Constants](#15-literals--constants)
6. [Variables & Zero Values](#16-variables--zero-values)
7. [Control Flow Structures](#17-control-flow-structures)
8. [The `defer` Statement](#18-the-defer-statement)
9. [`goto` and Labels](#19-goto-and-labels)
10. [Environment Variables & OS Interaction](#110-environment-variables--os-interaction)
11. [Interview Questions](#interview-questions)

---

## 1.1 Source Code Representation

### UTF-8 Encoding Requirement

**Interview Question:** *"What character encoding does Go source code use? Can I use non-ASCII identifiers?"*

Go source code is **always** UTF-8 encoded. This is not optional.

```go
// Valid Go - Unicode identifiers are allowed!
var 日本語 = "Japanese"
var Ñoño = "Spanish"
var α, β = 1, 2

func 計算(x int) int {
    return x * 2
}
```

**Key Points:**
- No BOM (Byte Order Mark) allowed
- All string literals are UTF-8 by default
- Unicode letters can be identifiers (but ASCII is conventional)

### Semicolon Insertion Rules

**Interview Question:** *"Does Go use semicolons? Why don't we see them in code?"*

Go uses semicolons, but they're **automatically inserted** by the lexer:

**Rule:** If the last token before a newline is one of:
- An identifier (`x`, `myFunc`)
- A literal (`42`, `"hello"`)
- One of: `break`, `continue`, `fallthrough`, `return`
- One of: `++`, `--`, `)`, `]`, `}`

Then a semicolon is inserted after that token.

```go
// What you write:
func main() {
    x := 5
    fmt.Println(x)
}

// What the lexer sees:
func main() {
    x := 5;
    fmt.Println(x);
};
```

**This is why braces must be on the same line:**

```go
// WRONG - semicolon inserted after if condition!
if condition
{  // This becomes: if condition;
    // ...
}

// RIGHT
if condition {
    // ...
}
```

### Comments

**Interview Question:** *"How do documentation comments work in Go?"*

```go
// Line comment - for implementation notes

/*
Block comment - less common,
spans multiple lines
*/

// Package mypackage provides utilities for...
// 
// This is a documentation comment (godoc).
// It must immediately precede the declaration.
package mypackage

// Add returns the sum of two integers.
// It handles overflow by wrapping.
func Add(a, b int) int {
    return a + b
}
```

**Documentation conventions:**
- First sentence becomes the synopsis
- Start with the name of the element being documented
- Use complete sentences

### Example: Source Code Representation

```go
package main

import "fmt"

// Greeting returns a localized greeting.
// Supported languages: en, ja, es
func Greeting(lang string) string {
    greetings := map[string]string{
        "en": "Hello",
        "ja": "こんにちは",  // UTF-8 string literal
        "es": "¡Hola!",
    }
    
    if g, ok := greetings[lang]; ok {
        return g
    }
    return greetings["en"]
}

func main() {
    fmt.Println(Greeting("ja"))  // Output: こんにちは
}
```

---

## 1.2 Identifiers & Naming

### Identifier Grammar

**Interview Question:** *"What are the rules for valid identifiers in Go?"*

From the specification:
```
identifier = letter { letter | unicode_digit }
letter     = unicode_letter | "_"
```

**Valid identifiers:**
```go
x
_x9
ThisIsValid
αβγ
_  // Blank identifier (special)
```

**Invalid identifiers:**
```go
9x     // Cannot start with digit
my-var // No hyphens
my var // No spaces
```

### Predeclared Identifiers

**Interview Question:** *"Can you shadow built-in types or functions in Go? What are the risks?"*

Go has predeclared identifiers that are **NOT reserved** — you can shadow them (but shouldn't):

**Types:**
```go
bool, byte, complex64, complex128, error, float32, float64,
int, int8, int16, int32, int64, rune, string,
uint, uint8, uint16, uint32, uint64, uintptr
```

**Constants:**
```go
true, false, iota
```

**Zero value:**
```go
nil
```

**Functions:**
```go
append, cap, close, complex, copy, delete, imag, len,
make, new, panic, print, println, real, recover
```

**Dangerous shadowing example:**
```go
func bad() {
    // DON'T DO THIS - shadows built-in
    len := 5
    fmt.Println(len)      // Works, prints 5
    // fmt.Println(len("hello"))  // ERROR: len is int, not function!
}
```

**Best Practice:** Never shadow predeclared identifiers.

### The Blank Identifier (`_`)

**Interview Question:** *"What is the blank identifier and when would you use it?"*

The blank identifier `_` discards values:

```go
// 1. Ignore return values
_, err := strconv.Atoi("42")
if err != nil {
    return err
}

// 2. Import for side effects only
import _ "github.com/lib/pq"  // Registers PostgreSQL driver

// 3. Compile-time interface check
var _ io.Reader = (*MyType)(nil)

// 4. Ignore loop index
for _, v := range slice {
    fmt.Println(v)
}
```

### Export Rules (Visibility)

**Interview Question:** *"How does Go handle public vs. private visibility?"*

Go uses **capitalization** instead of access modifiers:

| First Letter | Visibility | Equivalent |
|--------------|------------|------------|
| Uppercase (`MyFunc`) | Exported (public) | `public` |
| Lowercase (`myFunc`) | Unexported (package-private) | `private` |

```go
package user

type User struct {
    ID   int    // Exported - accessible outside package
    name string // Unexported - only accessible within package
}

func New(name string) *User {  // Exported
    return &User{name: name}
}

func (u *User) validate() bool {  // Unexported
    return u.name != ""
}
```

**Accessing from another package:**
```go
package main

import "myapp/user"

func main() {
    u := user.New("Alice")
    fmt.Println(u.ID)      // OK - exported
    // fmt.Println(u.name)  // ERROR - unexported
}
```

### Naming Conventions

**Interview Question:** *"What naming conventions does Go follow? How are acronyms handled?"*

**Conventions (from Effective Go and style guides):**

| Type | Convention | Example |
|------|------------|---------|
| Package | Short, lowercase, no underscores | `http`, `strconv`, `bufio` |
| Variable | MixedCaps, short in small scope | `userID`, `i`, `ctx` |
| Constant | MixedCaps (not ALL_CAPS!) | `MaxRetries`, `defaultTimeout` |
| Function | MixedCaps | `HandleRequest`, `parseJSON` |
| Interface | MixedCaps, often `-er` suffix | `Reader`, `Stringer`, `Handler` |

**Acronyms — keep consistent case:**
```go
// CORRECT
var userID string
var httpClient *http.Client
var xmlParser Parser
type HTTPHandler struct{}
type XMLRPC struct{}

// WRONG
var UserId string
var HttpClient *http.Client
type HttpHandler struct{}
```

### Example: Proper Naming

```go
package repository

import "context"

// UserRepository defines the interface for user persistence.
type UserRepository interface {
    FindByID(ctx context.Context, id string) (*User, error)
    Save(ctx context.Context, user *User) error
}

// postgresUserRepo is the PostgreSQL implementation.
// Note: unexported type with exported interface.
type postgresUserRepo struct {
    db *sql.DB
}

// NewPostgresUserRepo creates a UserRepository backed by PostgreSQL.
func NewPostgresUserRepo(db *sql.DB) UserRepository {
    return &postgresUserRepo{db: db}
}

func (r *postgresUserRepo) FindByID(ctx context.Context, id string) (*User, error) {
    const query = `SELECT id, name FROM users WHERE id = $1`
    // ...
}
```

---

## 1.3 Keywords (The Reserved 25)

**Interview Question:** *"How many keywords does Go have? Name some unique ones."*

Go has exactly **25 keywords** — intentionally minimal:

### Declaration Keywords

```go
package main          // Declares package name

import "fmt"          // Imports packages

const Pi = 3.14159    // Compile-time constant

var globalVar int     // Variable declaration

type User struct {    // Type definition
    Name string
}

func Add(a, b int) int {  // Function declaration
    return a + b
}
```

### Composite Type Keywords

```go
// struct - aggregate type
type Point struct {
    X, Y float64
}

// interface - behavior contract
type Reader interface {
    Read(p []byte) (n int, err error)
}

// map - hash table
ages := map[string]int{"alice": 30}

// chan - communication channel
ch := make(chan int)
```

### Control Flow Keywords

```go
// if, else
if x > 0 {
    fmt.Println("positive")
} else {
    fmt.Println("non-positive")
}

// for, range, break, continue
for i, v := range slice {
    if v < 0 {
        continue
    }
    if v > 100 {
        break
    }
}

// switch, case, default, fallthrough
switch day {
case "Mon", "Tue", "Wed", "Thu", "Fri":
    fmt.Println("weekday")
    fallthrough  // Explicit fallthrough
case "Sat", "Sun":
    fmt.Println("or weekend")
default:
    fmt.Println("unknown")
}

// return
func double(x int) int {
    return x * 2
}

// goto (rarely used)
    goto cleanup
cleanup:
    // ...
```

### Concurrency Keywords

```go
// go - spawn goroutine
go processItem(item)

// select - multiplex channels
select {
case v := <-ch1:
    fmt.Println(v)
case ch2 <- x:
    fmt.Println("sent")
default:
    fmt.Println("no communication")
}
```

### Special Keyword

```go
// defer - schedule for function exit
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()  // Will run when function returns
    
    // ... process file
    return nil
}
```

---

## 1.4 Operators & Delimiters

### Arithmetic Operators

**Interview Question:** *"How does integer division work in Go? What happens with division by zero?"*

```go
// Addition, Subtraction, Multiplication
a := 10 + 3   // 13
b := 10 - 3   // 7
c := 10 * 3   // 30

// Integer Division (truncates toward zero)
d := 10 / 3   // 3 (not 3.333...)
e := -10 / 3  // -3 (truncates toward zero)

// Modulo (remainder)
f := 10 % 3   // 1

// Division by zero:
// - Integer: runtime panic
// - Float: +Inf, -Inf, or NaN
g := 1.0 / 0.0  // +Inf
```

### Bitwise Operators

**Interview Question:** *"What is the `&^` operator in Go? Give an example."*

```go
// AND
a := 0b1100 & 0b1010  // 0b1000 (8)

// OR
b := 0b1100 | 0b1010  // 0b1110 (14)

// XOR
c := 0b1100 ^ 0b1010  // 0b0110 (6)

// AND NOT (bit clear) - UNIQUE TO GO
// Clears bits in first operand where second operand has 1s
d := 0b1100 &^ 0b1010  // 0b0100 (4)
// Equivalent to: a & (^b)

// Left Shift
e := 1 << 3  // 8

// Right Shift
f := 16 >> 2  // 4
```

**Use case for `&^` (bit clear):**
```go
const (
    FlagRead  = 1 << iota  // 1
    FlagWrite              // 2
    FlagExec               // 4
)

permissions := FlagRead | FlagWrite | FlagExec  // 7

// Remove write permission
permissions = permissions &^ FlagWrite  // 5 (Read + Exec)
```

### Comparison Operators

**Interview Question:** *"Can you compare slices with `==` in Go?"*

```go
// Equality (only for comparable types)
a := 10 == 10    // true
b := "go" == "go" // true

// NOT comparable with ==:
// - Slices (use slices.Equal or loop)
// - Maps (use maps.Equal or loop)
// - Functions (can only compare to nil)

// Comparable:
// - Basic types (int, string, bool, etc.)
// - Pointers
// - Channels
// - Interfaces (but careful with nil!)
// - Arrays (if element type is comparable)
// - Structs (if all fields are comparable)

// Ordering
c := 5 < 10   // true
d := 5 <= 5   // true
e := "a" < "b" // true (lexicographic)
```

### Logical Operators

**Interview Question:** *"Does Go have short-circuit evaluation?"*

```go
// AND (short-circuit)
if isValid() && process() {  // process() only called if isValid() is true
    // ...
}

// OR (short-circuit)
if hasCached() || fetch() {  // fetch() only called if hasCached() is false
    // ...
}

// NOT
if !isEmpty(slice) {
    // ...
}
```

### Address and Pointer Operators

```go
x := 42
p := &x      // p is *int, points to x
fmt.Println(*p)  // 42 (dereference)
*p = 100     // x is now 100
```

### Channel Operator

```go
ch := make(chan int)

// Send
ch <- 42

// Receive
v := <-ch

// Receive with ok (closed channel check)
v, ok := <-ch
if !ok {
    fmt.Println("channel closed")
}
```

### Operator Precedence Table

**Interview Question:** *"What is the operator precedence in Go?"*

From highest to lowest:

| Precedence | Operators |
|------------|-----------|
| 5 (highest) | `*`, `/`, `%`, `<<`, `>>`, `&`, `&^` |
| 4 | `+`, `-`, `\|`, `^` |
| 3 | `==`, `!=`, `<`, `<=`, `>`, `>=` |
| 2 | `&&` |
| 1 (lowest) | `\|\|` |

```go
// Example: what does this evaluate to?
a := 2 + 3*4   // 14 (not 20)
b := 1 | 2 + 3 // 4 (2+3=5, 1|5=5, wait no: + is higher, so 1|(2+3)=1|5=5)
// Actually: + has lower precedence than |
// So: 1 | 2 + 3 = (1 | 2) + 3 = 3 + 3 = 6? 
// NO! Check table: + has precedence 4, | has precedence 4
// Same precedence: left to right
// 1 | 2 + 3 = ... actually | and + are same level, left-to-right
// (1 | 2) + 3 = 3 + 3 = 6
```

**Best Practice:** Use parentheses to make intent clear.

### Example: Operators in Action

```go
package main

import "fmt"

func main() {
    // Bitwise operations for permissions
    const (
        Read  = 1 << iota  // 1
        Write              // 2
        Exec               // 4
    )
    
    // Combine permissions
    userPerms := Read | Write  // 3
    
    // Check permission
    canWrite := userPerms&Write != 0  // true
    
    // Remove permission
    userPerms = userPerms &^ Write  // 1 (Read only)
    
    fmt.Printf("Permissions: %03b\n", userPerms)
    fmt.Printf("Can write: %v\n", canWrite)
    
    // Short-circuit evaluation
    items := []int{1, 2, 3}
    if len(items) > 0 && items[0] == 1 {  // Safe: len checked first
        fmt.Println("First item is 1")
    }
}
```

---

## 1.5 Literals & Constants

### Integer Literals

**Interview Question:** *"What numeric literal formats does Go support?"*

```go
// Decimal
a := 42

// Binary (0b or 0B prefix)
b := 0b101010  // 42

// Octal (0o or 0O prefix, or just 0)
c := 0o52      // 42
d := 052       // 42 (legacy, avoid)

// Hexadecimal (0x or 0X prefix)
e := 0x2A      // 42

// Underscores for readability (Go 1.13+)
million := 1_000_000
binary := 0b1010_1010
hex := 0xFF_FF_FF_FF
```

### Floating-Point Literals

```go
// Decimal
a := 3.14159
b := .5        // 0.5
c := 5.        // 5.0

// Scientific notation
d := 6.022e23  // 6.022 × 10²³
e := 1E-10     // 0.0000000001

// Hexadecimal float (rare, Go 1.13+)
f := 0x1.fp3   // 1.9375 × 2³ = 15.5
```

### Rune Literals

**Interview Question:** *"What is the difference between a byte and a rune in Go?"*

```go
// Single-quoted character
a := 'A'       // rune (int32), value 65
b := '日'      // rune, value 26085
c := '\n'      // newline, value 10

// Escape sequences
'\a'   // Alert (bell)
'\b'   // Backspace
'\f'   // Form feed
'\n'   // Newline
'\r'   // Carriage return
'\t'   // Tab
'\v'   // Vertical tab
'\\'   // Backslash
'\''   // Single quote

// Unicode escapes
'\u0041'      // 'A' (16-bit Unicode)
'\U00010000'  // 𐀀 (32-bit Unicode)
'\x41'        // 'A' (8-bit hex)
```

**Key Distinction:**
- `byte` = `uint8` (0-255)
- `rune` = `int32` (Unicode code point, 0 to 0x10FFFF)

### String Literals

**Interview Question:** *"What's the difference between interpreted and raw string literals?"*

```go
// Interpreted string (processes escapes)
s1 := "Hello\nWorld"  // Contains actual newline

// Raw string (literal, no escapes processed)
s2 := `Hello\nWorld`  // Contains backslash-n literally
s3 := `Line 1
Line 2
Line 3`  // Can span multiple lines

// Common use for raw strings:
regex := `\d+\.\d+`           // Regex without double escaping
json := `{"name": "Alice"}`   // JSON without escaping quotes
sql := `SELECT * FROM users
        WHERE active = true`  // Multi-line SQL
```

### Constants

**Interview Question:** *"What are untyped constants in Go? Why do they exist?"*

```go
// Typed constant
const typedPi float64 = 3.14159

// Untyped constant (has arbitrary precision!)
const untypedPi = 3.14159265358979323846264338327950288

// Untyped constants adapt to context
var f32 float32 = untypedPi  // Works!
var f64 float64 = untypedPi  // Works!
// var f32 float32 = typedPi // ERROR: cannot use float64 as float32
```

**Untyped constant precision:**
```go
const huge = 1 << 100  // Valid! Beyond int64 range

// But assignment requires fitting in target type:
// var x int64 = huge  // ERROR: overflow
var x = huge >> 90     // OK: now fits in int
```

### Iota Enumerator

**Interview Question:** *"Explain how `iota` works. Can you show some advanced patterns?"*

```go
// Basic usage - resets to 0 in each const block
const (
    Sunday = iota    // 0
    Monday           // 1
    Tuesday          // 2
    Wednesday        // 3
    Thursday         // 4
    Friday           // 5
    Saturday         // 6
)

// Skip values
const (
    _ = iota         // 0 (discarded)
    One              // 1
    Two              // 2
    _                // 3 (discarded)
    Four             // 4
)

// Bit flags
const (
    FlagRead = 1 << iota  // 1
    FlagWrite             // 2
    FlagExec              // 4
)

// Byte sizes
const (
    _  = iota
    KB = 1 << (10 * iota)  // 1 << 10 = 1024
    MB                      // 1 << 20
    GB                      // 1 << 30
    TB                      // 1 << 40
)

// Multiple constants per line
const (
    a, b = iota, iota + 10  // 0, 10
    c, d                     // 1, 11
    e, f                     // 2, 12
)
```

### Example: Constants in Practice

```go
package config

import "time"

// Configuration constants
const (
    // Server settings
    DefaultPort    = 8080
    DefaultTimeout = 30 * time.Second
    MaxConnections = 1000
    
    // Feature flags using bit positions
    FeatureLogging = 1 << iota
    FeatureMetrics
    FeatureTracing
    FeatureRateLimit
    
    // Default features
    DefaultFeatures = FeatureLogging | FeatureMetrics
)

// Byte size constants
const (
    _ = 1 << (10 * iota)
    KB
    MB
    GB
)

func HasFeature(features, flag int) bool {
    return features&flag != 0
}
```

---

## 1.6 Variables & Zero Values

### Variable Declaration Forms

**Interview Question:** *"What are the different ways to declare variables in Go? When would you use each?"*

```go
// Form 1: Full declaration
var name string = "Alice"

// Form 2: Type inference
var name = "Alice"

// Form 3: Zero value declaration
var name string  // name is ""

// Form 4: Short declaration (inside functions only)
name := "Alice"

// Form 5: Multiple variables
var x, y int = 1, 2
a, b := 1, "hello"

// Form 6: Block declaration
var (
    host     = "localhost"
    port     = 8080
    maxConns = 100
)
```

**When to use each:**

| Form | Use Case |
|------|----------|
| `var x T = v` | When type differs from literal type |
| `var x = v` | Package-level, or when type is obvious |
| `var x T` | Need zero value explicitly |
| `x := v` | Inside functions, most common |

### Short Variable Declaration Gotcha

**Interview Question:** *"What is the redeclaration gotcha with `:=`?"*

```go
func example() {
    x := 1
    
    // This shadows x in inner scope!
    if true {
        x := 2  // NEW variable, shadows outer x
        fmt.Println(x)  // 2
    }
    fmt.Println(x)  // 1 (outer x unchanged)
    
    // Redeclaration - only allowed if:
    // 1. At least one variable is new
    // 2. Same scope
    x, y := 10, 20  // x redeclared, y is new - OK
    
    // ERROR: no new variables
    // x := 100
}
```

**The shadowing bug:**
```go
func getUser() (*User, error) {
    user, err := fetchFromCache()
    if err != nil {
        user, err := fetchFromDB()  // SHADOW! Outer user unchanged
        if err != nil {
            return nil, err
        }
        // user goes out of scope here!
    }
    return user, nil  // Returns cached user, not DB user!
}

// FIX:
func getUser() (*User, error) {
    user, err := fetchFromCache()
    if err != nil {
        user, err = fetchFromDB()  // Assignment, not declaration
        if err != nil {
            return nil, err
        }
    }
    return user, nil
}
```

### Zero Value Guarantee

**Interview Question:** *"What are the zero values for different types in Go? Why is this important?"*

Go guarantees that every variable is initialized to a well-defined zero value:

| Type | Zero Value | Notes |
|------|------------|-------|
| `bool` | `false` | |
| `int`, `int8`, etc. | `0` | |
| `float32`, `float64` | `0.0` | |
| `complex64`, `complex128` | `0+0i` | |
| `string` | `""` | Empty string, NOT nil |
| `pointer` | `nil` | |
| `slice` | `nil` | `len` and `cap` are 0 |
| `map` | `nil` | Reading returns zero value, writing panics! |
| `channel` | `nil` | Send and receive block forever |
| `function` | `nil` | |
| `interface` | `nil` | |
| `struct` | All fields zeroed | Recursively |
| `array` | All elements zeroed | Recursively |

**Why this matters:**
```go
// No need to check for uninitialized values
type Counter struct {
    count int  // Automatically 0
}

func (c *Counter) Increment() {
    c.count++  // Safe to use immediately
}

// Zero value is useful
var buf bytes.Buffer  // Ready to use, no initialization needed
buf.WriteString("hello")
```

### Example: Zero Values in Action

```go
package main

import "fmt"

type Config struct {
    Host    string
    Port    int
    Timeout int
    Debug   bool
}

func main() {
    // All fields get zero values
    var cfg Config
    
    fmt.Printf("Host: %q\n", cfg.Host)      // ""
    fmt.Printf("Port: %d\n", cfg.Port)      // 0
    fmt.Printf("Timeout: %d\n", cfg.Timeout) // 0
    fmt.Printf("Debug: %v\n", cfg.Debug)    // false
    
    // Zero value slice is safe to append
    var items []string  // nil
    items = append(items, "first")  // Works!
    
    // Zero value map is NOT safe to write
    var users map[string]int  // nil
    // users["alice"] = 30  // PANIC!
    
    // Must initialize map before writing
    users = make(map[string]int)
    users["alice"] = 30  // OK
}
```

---

## 1.7 Control Flow Structures

### The `for` Loop (Go's Only Loop)

**Interview Question:** *"Go only has one loop keyword. How do you implement different loop patterns?"*

#### Three-Component Form (C-style)

```go
for i := 0; i < 10; i++ {
    fmt.Println(i)
}
```

#### Condition-Only Form (while-style)

```go
for condition {
    // ...
}

// Example: read until EOF
for scanner.Scan() {
    line := scanner.Text()
    process(line)
}
```

#### Infinite Loop

```go
for {
    // ...
    if shouldStop {
        break
    }
}
```

#### Range Loop

**Interview Question:** *"What are the different ways to use `range`? What does it return for each type?"*

| Type | Range Returns | Example |
|------|---------------|---------|
| Array/Slice | index, value copy | `for i, v := range slice` |
| String | byte index, rune | `for i, r := range str` |
| Map | key, value copy | `for k, v := range m` |
| Channel | value | `for v := range ch` |

```go
// Slice - index and value
for i, v := range []int{10, 20, 30} {
    fmt.Printf("Index %d: %d\n", i, v)
}

// Ignore index
for _, v := range slice {
    fmt.Println(v)
}

// Ignore value (just need index)
for i := range slice {
    slice[i] *= 2
}

// String - byte index and rune
for i, r := range "日本語" {
    fmt.Printf("Byte %d: %c (U+%04X)\n", i, r, r)
}
// Output:
// Byte 0: 日 (U+65E5)
// Byte 3: 本 (U+672C)
// Byte 6: 語 (U+8A9E)

// Map - random order!
for k, v := range map[string]int{"a": 1, "b": 2} {
    fmt.Println(k, v)  // Order is random
}

// Channel - until closed
for v := range ch {
    process(v)
}
```

#### Loop Variable Capture Bug (Pre-Go 1.22)

**Interview Question:** *"What is the loop variable capture bug? How was it fixed?"*

```go
// Pre-Go 1.22 BUG:
funcs := make([]func(), 3)
for i := 0; i < 3; i++ {
    funcs[i] = func() { fmt.Println(i) }
}
for _, f := range funcs {
    f()  // Prints: 3, 3, 3 (all capture same variable!)
}

// OLD Fix: capture explicitly
for i := 0; i < 3; i++ {
    i := i  // Shadow with new variable
    funcs[i] = func() { fmt.Println(i) }
}

// Go 1.22+ FIX: loop variables are per-iteration
// The original code now works correctly!
```

### Conditional Statements

#### `if` Statement

```go
// Basic
if x > 0 {
    fmt.Println("positive")
}

// With initialization (scoped variable)
if err := validate(); err != nil {
    return err
}
// err not accessible here

// if-else
if x > 0 {
    fmt.Println("positive")
} else if x < 0 {
    fmt.Println("negative")
} else {
    fmt.Println("zero")
}
```

### Switch Statement

**Interview Question:** *"How does Go's `switch` differ from C/Java? What is a type switch?"*

#### Expression Switch

```go
// No fallthrough by default
switch day {
case "Monday":
    fmt.Println("Start of week")
case "Friday":
    fmt.Println("TGIF!")
case "Saturday", "Sunday":  // Multiple values
    fmt.Println("Weekend!")
default:
    fmt.Println("Midweek")
}

// Explicit fallthrough
switch n {
case 1:
    fmt.Println("one")
    fallthrough  // Continue to next case
case 2:
    fmt.Println("one or two")
}
```

#### Tagless Switch (cleaner if-else)

```go
switch {
case x < 0:
    fmt.Println("negative")
case x == 0:
    fmt.Println("zero")
case x > 0:
    fmt.Println("positive")
}
```

#### Type Switch

```go
func describe(i interface{}) {
    switch v := i.(type) {
    case int:
        fmt.Printf("Integer: %d\n", v)
    case string:
        fmt.Printf("String: %s\n", v)
    case bool:
        fmt.Printf("Boolean: %v\n", v)
    case nil:
        fmt.Println("nil value")
    default:
        fmt.Printf("Unknown type: %T\n", v)
    }
}
```

### Break, Continue, and Labels

```go
// Break inner loop only
for i := 0; i < 10; i++ {
    for j := 0; j < 10; j++ {
        if j == 5 {
            break  // Only breaks inner loop
        }
    }
}

// Break outer loop with label
outer:
for i := 0; i < 10; i++ {
    for j := 0; j < 10; j++ {
        if i*j > 50 {
            break outer  // Breaks outer loop
        }
    }
}

// Continue with label
outer:
for i := 0; i < 3; i++ {
    for j := 0; j < 3; j++ {
        if j == 1 {
            continue outer  // Skip to next i iteration
        }
        fmt.Println(i, j)
    }
}
```

### Example: Control Flow

```go
package main

import "fmt"

func processItems(items []int) {
    // Early exit if empty
    if len(items) == 0 {
        return
    }
    
    // Process with different strategies based on size
    switch {
    case len(items) < 10:
        // Simple iteration for small sets
        for _, item := range items {
            fmt.Println(item)
        }
    case len(items) < 1000:
        // Batch processing for medium sets
        for i := 0; i < len(items); i += 10 {
            end := i + 10
            if end > len(items) {
                end = len(items)
            }
            processBatch(items[i:end])
        }
    default:
        // Parallel processing for large sets
        processParallel(items)
    }
}
```

---

## 1.8 The `defer` Statement

**Interview Question:** *"Explain how `defer` works. What is the execution order? When are arguments evaluated?"*

### Execution Timing

`defer` schedules a function call to run:
1. **After** the surrounding function returns
2. **Before** the return value is delivered to the caller

```go
func example() string {
    defer fmt.Println("deferred")
    fmt.Println("normal")
    return "result"
}
// Output:
// normal
// deferred
// (then "result" is returned)
```

### LIFO Order

Multiple defers execute in Last-In-First-Out order:

```go
func example() {
    defer fmt.Println("1")
    defer fmt.Println("2")
    defer fmt.Println("3")
}
// Output:
// 3
// 2
// 1
```

### Argument Evaluation Time

**Critical:** Arguments are evaluated when `defer` is executed, NOT when the deferred function runs:

```go
func example() {
    x := 1
    defer fmt.Println(x)  // x is evaluated NOW (1)
    x = 2
}
// Output: 1 (not 2!)

// To capture current value at execution time:
func example2() {
    x := 1
    defer func() {
        fmt.Println(x)  // Closure captures reference
    }()
    x = 2
}
// Output: 2
```

### Named Return Interaction

**Interview Question:** *"Can a deferred function modify the return value?"*

```go
func double(x int) (result int) {
    defer func() {
        result *= 2  // Modifies named return!
    }()
    return x  // result = x, then defer runs
}

fmt.Println(double(5))  // Output: 10
```

### Common Use Cases

```go
// 1. Resource cleanup
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()  // Always closes, even on error
    
    // ... process file
    return nil
}

// 2. Mutex unlock
func (s *SafeCounter) Increment() {
    s.mu.Lock()
    defer s.mu.Unlock()  // Unlocks even if panic
    s.count++
}

// 3. Recover from panic
func safeCall(f func()) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic: %v", r)
        }
    }()
    f()
    return nil
}

// 4. Timing/tracing
func processRequest() {
    defer trace("processRequest")()  // Note: double parens
    // ...
}

func trace(name string) func() {
    start := time.Now()
    fmt.Printf("Entering %s\n", name)
    return func() {
        fmt.Printf("Exiting %s (took %v)\n", name, time.Since(start))
    }
}
```

### Performance Considerations

**Interview Question:** *"Is there overhead to using `defer`? When might you avoid it?"*

Before Go 1.14, defer had noticeable overhead (~35ns). Since Go 1.14, most defers are "open-coded" with nearly zero overhead.

```go
// Still avoid defer in tight loops for absolute performance
func sum(items []int) int {
    total := 0
    for _, item := range items {
        // DON'T: defer something() in a loop
        total += item
    }
    return total
}
```

### Example: Defer Patterns

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

// Pattern 1: Resource cleanup
func readConfig(path string) ([]byte, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    
    return io.ReadAll(f)
}

// Pattern 2: Mutex with defer
type SafeMap struct {
    mu   sync.RWMutex
    data map[string]int
}

func (m *SafeMap) Get(key string) (int, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    v, ok := m.data[key]
    return v, ok
}

// Pattern 3: Timing decorator
func timed(name string) func() {
    start := time.Now()
    return func() {
        fmt.Printf("%s took %v\n", name, time.Since(start))
    }
}

func expensiveOperation() {
    defer timed("expensiveOperation")()
    
    time.Sleep(100 * time.Millisecond)
}

func main() {
    expensiveOperation()  // Output: expensiveOperation took 100.xxx ms
}
```

---

## 1.9 `goto` and Labels

**Interview Question:** *"Does Go have `goto`? When would you ever use it?"*

### Label Syntax

```go
// Labels end with colon
MyLabel:
    statement
```

### `goto` Restrictions

```go
// CANNOT jump over variable declarations
func invalid() {
    goto End
    x := 5  // Error: goto jumps over declaration
End:
    fmt.Println(x)
}

// CANNOT jump into a block
func alsoInvalid() {
    goto Inside  // Error
    if true {
    Inside:
        fmt.Println("inside")
    }
}
```

### Legitimate Use Cases

**1. Breaking out of deeply nested loops:**
```go
func findInMatrix(matrix [][]int, target int) (int, int) {
    for i, row := range matrix {
        for j, val := range row {
            if val == target {
                return i, j
            }
        }
    }
    return -1, -1
}

// Alternative with goto (less preferred, but valid)
func findWithGoto(matrix [][]int, target int) (ri, rj int) {
    for i, row := range matrix {
        for j, val := range row {
            if val == target {
                ri, rj = i, j
                goto Found
            }
        }
    }
    return -1, -1
Found:
    return ri, rj
}
```

**2. Error handling cleanup (rare):**
```go
func processWithCleanup() error {
    resource1, err := acquire1()
    if err != nil {
        return err
    }
    
    resource2, err := acquire2()
    if err != nil {
        goto Cleanup1
    }
    
    resource3, err := acquire3()
    if err != nil {
        goto Cleanup2
    }
    
    // Use resources...
    return nil

Cleanup2:
    release2(resource2)
Cleanup1:
    release1(resource1)
    return err
}
```

**Best Practice:** Prefer labeled `break` over `goto`. Use `goto` sparingly.

---

## 1.10 Environment Variables & OS Interaction

### The `os` Package Environment Functions

**Interview Question:** *"How do you read environment variables in Go? What's the difference between `Getenv` and `LookupEnv`?"*

```go
import "os"

// Get value (returns "" if not set)
port := os.Getenv("PORT")

// Get value with existence check
port, exists := os.LookupEnv("PORT")
if !exists {
    port = "8080"  // Default
}

// Set value (current process only)
os.Setenv("MY_VAR", "value")

// Remove value
os.Unsetenv("MY_VAR")

// Get all environment variables
for _, env := range os.Environ() {
    fmt.Println(env)  // KEY=value format
}

// Expand variables in string
path := os.ExpandEnv("$HOME/.config")
```

### Environment Variable Patterns

**Interview Question:** *"How do you handle configuration in a Go application following 12-Factor App principles?"*

```go
package config

import (
    "os"
    "strconv"
    "time"
)

type Config struct {
    Host         string
    Port         int
    Debug        bool
    Timeout      time.Duration
    DatabaseURL  string
}

func Load() (*Config, error) {
    cfg := &Config{
        Host:    getEnvOrDefault("HOST", "localhost"),
        Port:    getEnvAsInt("PORT", 8080),
        Debug:   getEnvAsBool("DEBUG", false),
        Timeout: getEnvAsDuration("TIMEOUT", 30*time.Second),
    }
    
    // Required variable
    dbURL, exists := os.LookupEnv("DATABASE_URL")
    if !exists {
        return nil, fmt.Errorf("DATABASE_URL is required")
    }
    cfg.DatabaseURL = dbURL
    
    return cfg, nil
}

func getEnvOrDefault(key, defaultVal string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
    if val := os.Getenv(key); val != "" {
        if i, err := strconv.Atoi(val); err == nil {
            return i
        }
    }
    return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
    if val := os.Getenv(key); val != "" {
        if b, err := strconv.ParseBool(val); err == nil {
            return b
        }
    }
    return defaultVal
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
    if val := os.Getenv(key); val != "" {
        if d, err := time.ParseDuration(val); err == nil {
            return d
        }
    }
    return defaultVal
}
```

### Command-Line Arguments

```go
import "os"

func main() {
    // os.Args[0] is the program name
    // os.Args[1:] are the arguments
    
    if len(os.Args) < 2 {
        fmt.Println("Usage: program <command>")
        os.Exit(1)
    }
    
    command := os.Args[1]
    args := os.Args[2:]
    
    switch command {
    case "run":
        run(args)
    case "help":
        printHelp()
    default:
        fmt.Printf("Unknown command: %s\n", command)
        os.Exit(1)
    }
}
```

### Testing with Environment Variables

```go
func TestConfig(t *testing.T) {
    // Go 1.17+ - automatically restored after test
    t.Setenv("PORT", "9090")
    t.Setenv("DEBUG", "true")
    
    cfg, err := Load()
    if err != nil {
        t.Fatal(err)
    }
    
    if cfg.Port != 9090 {
        t.Errorf("expected port 9090, got %d", cfg.Port)
    }
}
```

---

## Interview Questions

### Beginner Level

1. **Q:** What are the two types of string literals in Go?
   **A:** Interpreted strings (`"..."`) process escape sequences, raw strings (`` `...` ``) are literal and can span lines.

2. **Q:** What is the zero value of a string? Of a slice?
   **A:** String: `""` (empty string). Slice: `nil` (but safe to append to).

3. **Q:** How do you discard a return value in Go?
   **A:** Use the blank identifier: `_, err := function()`

### Intermediate Level

4. **Q:** Why does this code print "3 3 3"?
   ```go
   for i := 0; i < 3; i++ {
       go func() { fmt.Println(i) }()
   }
   ```
   **A:** Loop variable capture bug. The closure captures a reference to `i`, which is 3 after the loop. Fixed in Go 1.22+, or use `i := i` to shadow.

5. **Q:** What's wrong with this code?
   ```go
   var m map[string]int
   m["key"] = 1
   ```
   **A:** Writing to a nil map panics. Must initialize: `m = make(map[string]int)`.

6. **Q:** Can a deferred function modify the return value?
   **A:** Yes, if the function uses named return values. The deferred function runs after the return statement but before the value is delivered.

### Advanced Level

7. **Q:** Explain the output:
   ```go
   defer fmt.Println(1)
   defer fmt.Println(2)
   defer fmt.Println(3)
   ```
   **A:** Output is "3 2 1" - LIFO order.

8. **Q:** What's special about the `&^` operator?
   **A:** It's AND NOT (bit clear), unique to Go. `a &^ b` clears bits in `a` where `b` has 1s.

9. **Q:** Why is this problematic?
   ```go
   func getUser() (*User, error) {
       user, err := cache.Get()
       if err != nil {
           user, err := db.Get()
           // ...
       }
       return user, nil
   }
   ```
   **A:** Inner `:=` shadows outer `user`. The cached user is returned even on cache miss. Use `=` instead of `:=` in the inner scope.

---

## Summary

Phase 1 covers Go's lexical elements and fundamental constructs:

| Topic | Key Points |
|-------|------------|
| Source Code | UTF-8 required, automatic semicolons |
| Identifiers | Capitalization controls visibility |
| Keywords | Only 25, intentionally minimal |
| Operators | `&^` is unique, precedence matters |
| Constants | Untyped constants have arbitrary precision |
| Zero Values | Everything has a well-defined default |
| Control Flow | `for` is the only loop, `switch` doesn't fall through |
| `defer` | LIFO order, arguments evaluated immediately |
| Environment | `os.Getenv`, `os.LookupEnv`, `os.Args` |

**Next Phase:** [Phase 2 — The Type System Deep Dive](../Phase_2/Phase_2.md)
