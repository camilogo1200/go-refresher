# Go - Identifiers

Identifiers name variables and types. An identifier is a sequence of one or more letters and digits.
**The first character in and _Identifier_ must be a letter**

```go
//definition
identifier = letter { letter | unicode_digit } .
```
## example
````go
var a int = 0;
type _myStructure struct{}
var ThisVariableIsExported = " "
αβ ="Hi"
````

# Keywords

- The following keywords are reserved and may not be used as identifiers

```markdown

| break | default | func | interface | select | 
| case  | defer   | go   | map       | struct |
| chan  |

```