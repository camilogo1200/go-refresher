# 📘 Phase 1: Go Language Specification Deep Dive

**Reference Version:** Go 1.22+
**Scope:** Syntax, Type System, Control Flow, and Built-ins.
**Objective:** Complete mastery of the language mechanics before attempting concurrency or architecture.

---

## 1. Lexical Elements
Go's source code is Unicode text (UTF-8). The language is designed to be parsed without a symbol table (context-free grammar).

### 1.1 Keywords (The Reserved 25)
Go has a minimalist set of 25 keywords. They cannot be used as identifiers.

| Category | Keywords |
| :--- | :--- |
| **Declarations** | `const`, `func`, `import`, `package`, `type`, `var` |
| **Composite Types** | `chan`, `interface`, `map`, `struct` |
| **Control Flow** | `break`, `case`, `continue`, `default`, `defer`, `else`, `fallthrough`, `for`, `goto`, `if`, `range`, `return`, `select`, `switch` |

### 1.2 Operators & Delimiters
Special attention to unique Go operators:
*   `&^` (Bit clear / AND NOT): Useful for bitmask manipulation.
*   `<-` (Receive): Used exclusively with channels.
*   `:=` (Short declaration): Declares and initializes variables inferred from the right-hand side.
*   `...` (Ellipsis): Used for variadic parameters and array unpacking.

### 1.3 Export Rules (Visibility)
Go does not use `public`, `private`, or `protected`. Visibility is controlled strictly by the **capitalization of the first letter** of the identifier.
*   **Uppercase (e.g., `User`):** Exported (Public). Visible to other packages.
*   **Lowercase (e.g., `user`):** Unexported (Private). Visible only within the *same package*.

---

## 2. The Type System
Go is statically typed and strongly typed. Explicit conversion is required between different types (even `int` and `int64`).

### 2.1 Basic Types
*   **Boolean:** `bool` (Values: `true`, `false`). *Note: Cannot be cast to/from integer (0/1).*
*   **Numeric:**
    *   **Integer (Arch-dependent):** `int`, `uint` (32-bit or 64-bit depending on CPU).
    *   **Explicit sized:** `int8`, `int16`, `int32`, `int64` (and `uint` variants).
    *   **Float:** `float32`, `float64` (Standard is `float64`).
    *   **Complex:** `complex64`, `complex128`.
    *   **Byte:** Alias for `uint8`.
    *   **Rune:** Alias for `int32`. Represents a Unicode Code Point.
*   **String:**
    *   Immutable sequence of bytes (UTF-8).
    *   Behaves like a value type but implemented as a (pointer, length) header.
    *   `len(s)` returns **byte count**, not character count.

### 2.2 Composite Types
#### Arrays
*   Fixed size. The size is **part of the type**. `[4]int` is distinct from `[5]int`.
*   Arrays are **values**. Assigning an array copies the entire content.

#### Slices (The Workhorse)
*   Dynamic view of an underlying array.
*   **Structure:** A 3-word header: `{ Pointer, Length, Capacity }`.
*   **Mechanics:** `s[i:j]` creates a new slice header sharing the underlying array.
*   **Expansion:** `append()` doubles capacity (up to a limit) when `len` exceeds `cap`, requiring a new allocation and copy.

#### Maps
*   Hash table reference type.
*   Unordered. Iteration order is randomized.
*   **Initialization:** Must use `make` or literal. Writing to a `nil` map causes a **Panic**.
*   **Retrieval:** `value, ok := map[key]`. The `ok` boolean checks for existence.

#### Structs
*   Sequence of named elements (fields).
*   **Memory Layout:** Fields are laid out contiguously. Padding is added for alignment.
*   **Anonymous Fields (Embedding):** Allows type promotion.
    ```go
    type Base struct { ID int }
    type User struct { Base; Name string } // User.ID is valid
    ```
    *Architect Note:* This is Composition, not Inheritance. `User` is *not* a `Base`.

#### Function Types
*   Functions are first-class citizens. They can be assigned to variables and passed as arguments.

#### Interfaces
*   A set of method signatures.
*   **Implementation:** Implicit. A type implements an interface if it defines the methods.
*   **The Empty Interface (`interface{}` or `any`):** Has 0 methods. All types implement it.

---

## 3. Declarations & Scope

### 3.1 Variables
*   **Standard:** `var x int = 10`
*   **Inferred:** `var x = 10`
*   **Short (Inside functions only):** `x := 10`
*   **Grouping:**
    ```go
    var (
        a = 1
        b = 2
    )
    ```

### 3.2 Constants & Iota
*   Constants are evaluated at **compile time**.
*   **Untyped Constants:** Numeric constants have arbitrary precision until assigned to a variable.
*   **Iota:** The boolean enumerator. Resets to 0 in every `const` block.
    ```go
    const (
        Red = iota // 0
        Blue       // 1
        Green      // 2
    )
    ```

### 3.3 Zero Values (Default Initialization)
Variables declared without an explicit initial value are given their zero value:
*   `0` for numeric types.
*   `false` for booleans.
*   `""` (empty string) for strings.
*   `nil` for pointers, functions, interfaces, slices, channels, and maps.

---

## 4. Control Flow Statements

### 4.1 The `for` Loop
Go has only one looping keyword: `for`.
*   **C-Style:** `for i := 0; i < 10; i++ { ... }`
*   **While-Style:** `for i < 10 { ... }`
*   **Infinite:** `for { ... }`
*   **Range:** Iterates over Slice, Map, String, or Channel.
    ```go
    for index, value := range mySlice { ... }
    ```
    *Spec Detail:* Range on a map is random order. Range on a string iterates over Runes (Unicode points), not bytes.

### 4.2 `if` / `else`
*   No parentheses around condition.
*   **Initialization Statement:** Supports a statement before the condition, scoping the variable to the block.
    ```go
    if err := doSomething(); err != nil {
        return err
    } // 'err' is undefined here
    ```

### 4.3 `switch`
*   **No Fallthrough:** Unlike C/Java, case execution stops automatically. Use `fallthrough` keyword to override.
*   **Tagless Switch:** Acts as a cleaner if-else chain.
    ```go
    switch {
    case x > 0: ...
    case x < 0: ...
    }
    ```
*   **Type Switch:** Specifically for interfaces.
    ```go
    switch v := i.(type) {
    case int: ...
    case string: ...
    }
    ```

### 4.4 `defer`
*   Schedules a function call to be run immediately **after** the surrounding function returns (but before the result is returned to the caller).
*   **LIFO Order:** Last-In-First-Out (Stack based).
*   **Evaluation:** Arguments are evaluated **immediately** when the `defer` statement is hit, not when the call executes.

### 4.5 `goto` and Labels
*   Go supports `goto`.
*   **Break/Continue with Labels:** Essential for breaking out of nested loops.
    ```go
    OuterLoop:
        for i := 0; i < 10; i++ {
            for j := 0; j < 10; j++ {
                if condition { break OuterLoop }
            }
        }
    ```

---

## 5. Functions & Methods
Functions are the central building block in Go. They are **first-class citizens**, meaning they can be assigned to variables, passed as arguments, and returned from other functions.

### 5.1 Function Declaration & Signatures
*   **Syntax:** `func Name(parameter-list) (result-list) { body }`
*   **Parameter/Result Typing:** Consecutive parameters of the same type can share a type identifier.
    *   `func add(a, b int) int` is valid.
*   **Discarding Returns:** If a function returns values, the caller must treat them. They can be explicitly discarded using the blank identifier `_`.

### 5.2 Multiple Return Values
Go functions can return multiple values, a feature often used to return a result alongside an `error`.
*   **Spec:** The return values are treated as a tuple for assignment but are not a formal tuple type in the language.
*   **Usage:**
    ```go
    func swap(x, y string) (string, string) {
        return y, x
    }
    ```

### 5.3 Named Return Values (Naked Returns)
Return parameters can be named in the function signature.
*   **Behavior:** They are initialized to their **zero values** at the start of the function.
*   **Naked Return:** A `return` statement without arguments returns the current values of the named results.
*   **Shadowing Warning:** Be careful of shadowing named return variables within the function scope.
    ```go
    func split(sum int) (x, y int) { // x and y are 0
        x = sum * 4 / 9
        y = sum - x
        return // Returns x, y
    }
    ```

### 5.4 Variadic Functions
*   **Syntax:** `...T` (must be the final parameter).
*   **Internal Representation:** The variadic argument is converted into a **slice** `[]T` within the function.
*   **Slice Unpacking:** You can pass an existing slice into a variadic function using the suffix `...`.
    ```go
    nums := []int{1, 2, 3}
    sum(nums...) // Unpacks slice into individual arguments
    ```

### 5.5 Function Literals (Closures)
*   **Anonymous Functions:** Functions defined inline without a name.
*   **Closures:** Function literals **close over** variables from their surrounding scope. They retain access to those variables even after the surrounding function returns.
    ```go
    func adder() func(int) int {
        sum := 0
        return func(x int) int {
            sum += x // 'sum' is trapped in the closure
            return sum
        }
    }
    ```

### 5.6 Methods & Receivers
A method is a function with a special *receiver* argument.
*   **Constraint:** You can only declare a method on a type defined in the **same package**. You cannot define methods on built-in types (like `int`) directly; you must `type` alias them first.

#### Value Receivers `func (v T)`
*   Operates on a **copy** of the value.
*   Mutations inside the method **do not** affect the original value.
*   Generally safe for concurrent read access.

#### Pointer Receivers `func (v *T)`
*   Operates on the **actual memory address**.
*   Mutations **do** affect the original value.
*   **Nil Receivers:** Unlike Java/C++, a method can be called on a `nil` pointer. The method body must handle the check: `if v == nil { return }`.

#### Method Sets (Crucial for Interfaces)
*   The method set of type `T` consists of all methods with receiver `T`.
*   The method set of type `*T` consists of all methods with receiver `*T` **AND** `T`.
*   *Implication:* If an interface requires a pointer-receiver method, you **cannot** assign a value of type `T` to that interface; it must be `*T`.

---

## 6. Built-in Functions
These functions are predeclared in the `universe` block and are available without imports.

### 6.1 Allocation & Initialization
*   **`new(T)`**:
    *   Allocates zeroed storage for type `T`.
    *   Returns `*T` (a pointer).
    *   Used for Structs and Primitives.
*   **`make(T, args)`**:
    *   Creates slices, maps, and channels **only**.
    *   Returns `T` (an initialized value, not a pointer).
    *   Initializes internal data structures (e.g., hash buckets for maps).

### 6.2 Container Manipulation
*   **`len(v)`**: Returns the length (int).
    *   String: Number of bytes.
    *   Slice/Array: Number of elements.
    *   Channel: Number of elements currently in buffer.
*   **`cap(v)`**: Returns the capacity (int).
    *   Slice: Size of underlying array.
    *   Channel: Buffer size.
*   **`append(s, ...elems)`**:
    *   Appends elements to the end of a slice.
    *   Returns the updated slice.
    *   **Memory Management:** If the backing array is too small, a larger array is allocated, data is copied, and the new slice points to the new array.
*   **`copy(dst, src)`**:
    *   Copies elements from `src` to `dst`.
    *   Returns the number of elements copied (minimum of `len(dst)` and `len(src)`).
*   **`delete(m, key)`**:
    *   Removes the element with `key` from map `m`.
    *   If `key` doesn't exist or `m` is nil, it is a no-op (safe).

### 6.3 Channel Management
*   **`close(c)`**:
    *   Closes a channel.
    *   Sends on a closed channel cause a **Panic**.
    *   Receives from a closed channel return the zero value immediately.

### 6.4 Handling Panics
Go does not have standard exceptions. It uses a Panic/Recover mechanism for unrecoverable errors.
*   **`panic(v)`**:
    *   Aborts execution of the current function.
    *   Runs any `defer` functions in LIFO order.
    *   Propagates up the stack until the program crashes.
*   **`recover()`**:
    *   Must be called **inside a deferred function**.
    *   Regains control of a panicking goroutine.
    *   Returns the value passed to `panic()`. If normal execution, returns `nil`.

---

## 7. Zero Values Reference
It is vital to know the default state of types when declared without assignment.

| Type | Zero Value |
| :--- | :--- |
| `bool` | `false` |
| `int`, `float`, etc. | `0` |
| `string` | `""` (Empty string, not null) |
| `pointer` | `nil` |
| `slice` | `nil` (Has no underlying array) |
| `map` | `nil` (Read is safe, Write panics) |
| `channel` | `nil` (Read/Write blocks forever) |
| `interface` | `nil` |
| `struct` | Recursively zeroed fields |