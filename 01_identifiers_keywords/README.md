# Go - Identifiers

Identifiers name variables and types. An identifier is a sequence of one or more letters and digits.
**The first character in and _Identifier_ must be a letter**

```go
//definition
identifier = letter { letter | unicode_digit }.
```

## example

````go
var a int = 0;
type _myStructure struct{}
var ThisVariableIsExported = " "
αβ = "Hi"
````

# Keywords

- The following keywords are reserved and may not be used as identifiers

| Category            | Keywords                                                                                               |
|:--------------------|:-------------------------------------------------------------------------------------------------------|
| **Declarations**    | const, var, func, import, package, type                                                                | 
| **Composite types** | chan, interface, map, struct                                                                           |
| **Control Flow**    | break, case, continue, default, defer, if, else, for, range, goto, return, select, switch, fallthrough |
