Below is the **complete YarLang v0 specification**, rewritten **from scratch** as if **Chris Lattner** and **Rob Pike** collaborated on a new systems language:

* Lattner brings precision, implementability, and compiler/IR rigor.
* Pike brings simplicity, clarity, and minimalism.

The result is a **small, coherent, practical v0 language**—easy to bootstrap, predictable to reason about, and designed to grow without painting future versions into a corner.

This is **the authoritative YarLang v0 Surface Language Specification**.

---

# **YarLang v0 – Surface Language Specification**

### *Chris Lattner × Rob Pike Edition*

*(A minimal, implementable systems language with modern ergonomics.)*

---

# **0. Design Philosophy**

YarLang v0 is intentionally small. Complexity belongs in future versions, not in a bootstrap compiler.

Its guiding principles:

1. **Make the common case simple.**
   Syntax is minimal, predictable, and direct.

2. **Avoid cleverness.**
   No magic. No overloaded semantics. No hidden machinery.

3. **Define behavior precisely.**
   A compiler can be written from this document alone.

4. **Build only what is needed for v0.**
   Concurrency, traits, async, incremental GC, and advanced type features belong in v1+.

v0 is a foundation: a language that can be implemented, used, and extended.

---

# **1. Lexical Structure**

### 1.1 Identifiers

```
[a-zA-Z_][a-zA-Z0-9_]*
```

Identifiers are case-sensitive.
`ok` and `err` are ordinary identifiers, not keywords.

### 1.2 Keywords

```
module  use  pub  fn  struct  enum
let  return  if  else  while  for  in
loop  break  continue  defer  when
```

No contextual keywords.

### 1.3 Literals

* Integer: decimal only in v0 (`42`, `9001`)
* Float: `1.0`, `3.14`
* String: UTF-8, `"`…"`, no raw strings in v0
* Boolean: `true`, `false`

### 1.4 Comments

* Line: `// …`
* Block comments are not in v0.

---

# **2. Modules**

### 2.1 Declaration

A file begins with exactly one module declaration:

```
module a.b.c
```

The path maps to the directory hierarchy under the build root.

### 2.2 Imports

```
use pkg.symbol
use pkg.*
pub use pkg.symbol
```

Rules:

* Imports apply to the file scope.
* Imports do not re-export unless prefixed with `pub`.
* Import order has no semantic meaning.

---

# **3. Types**

YarLang v0 has a minimal, orthogonal type system.

### 3.1 Primitive Types

```
bool, byte
i32, i64
u32, u64
f32, f64
string
```

### 3.2 Composite Types

```
T[n]        // array
T[]         // slice
map[K]V     // hash map
struct      // user-defined
enum        // user-defined, tuple variants only
fn(...) T   // function type
```

### 3.3 Values & Semantics

* Values are passed by copy unless they contain pointers.
* Pointer types are written `*T`.

---

# **4. Memory Model**

YarLang v0 uses a **precise stop-the-world mark-and-sweep GC**.

### 4.1 Traced Heap Objects

* Slices and their backing arrays
* Maps
* Strings
* Heap-allocated structs
* Closure environments

### 4.2 Roots

A GC cycle begins at:

* Local variables
* Global/module-level bindings
* Closure environments reachable from roots

No concurrency → no races → simpler GC invariants.

### 4.3 Pointer Rules

* `&expr` takes the address of an lvalue
* `*ptr` dereferences
* Pointer copies duplicate the address

No pointer arithmetic in v0.

---

# **5. Bindings**

### 5.1 Declaration

```
let name = expr
```

### 5.2 Reassignment

Bindings are mutable:

```
name = expr
```

### 5.3 Type Annotations

```
let n: i32 = 10
```

Type annotations are required only when inference is insufficient.

---

# **6. Structs**

### 6.1 Declaration

```
struct Point {
    x: i32
    y: i32
}
```

### 6.2 Literals

```
Point { x: 1, y: 2 }
Point { x: 1, ..p }   // copy remaining fields
```

Fields omitted default to zero values.

---

# **7. Enums**

### 7.1 Declaration

```
enum Result<T, E> {
    ok(T),
    err(E)
}
```

### 7.2 Constructors

Fully-qualified:

```
Result.ok
Result.err
```

Unqualified forms appear via `use Result.ok`.

### 7.3 Variant Types

Only tuple-like variants exist in v0.

---

# **8. Generics**

### 8.1 Syntax

```
fn box<T>(v: T) Box<T> { ... }
struct Box<T> { value: T }
```

### 8.2 Semantics

Generics are **monomorphized**:

* Each type instantiation generates specialized code.
* Cross-module monomorphization is allowed.

### 8.3 No constraints or traits in v0.

---

# **9. Functions**

### 9.1 Declaration

```
fn name(params) Return { ... }
```

### 9.2 Parameters

```
fn f(x: i32, y: string) bool
```

### 9.3 Return

```
return expr
```

If omitted, returns `void`.

### 9.4 No multiple return values

Use structs or `Result`.

---

# **10. Closures**

### 10.1 Syntax

```
|x| x + 1
|a, b| a + b
```

### 10.2 Capture Rules

* Variables are captured by reference.
* Captured variables are heap-lifted.
* Closure environments are GC-traced.

No concurrency → no data race model needed.

---

# **11. Control Flow**

### 11.1 Conditionals

```
if cond { ... }
else if cond { ... }
else { ... }
```

### 11.2 Loops

```
loop { ... }
while cond { ... }
for x in iterable { ... }
for i, v in iterable { ... }
for i in a..b { ... }  // end-exclusive
```

### 11.3 Break & Continue

Both behave in the expected way.

### 11.4 Defer

Runs at scope exit in LIFO order.

### 11.5 Result Matching (`when`)

Only for `Result` in v0:

```
when expr {
    ok v => ...
    err e => ...
}
```

No general pattern matching.

---

# **12. Error Propagation**

### 12.1 The `?` Operator

Valid only inside functions returning `Result<T, E>`.

```
let x = expr?
```

Desugars to:

```
when expr {
    ok v => v
    err e => return Result.err(e)
}
```

### 12.2 Main function error semantics

The CLI runtime prints errors and exits nonzero.

---

# **13. Pipelines**

### 13.1 Basic Form

```
lhs |> rhs
```

### 13.2 Desugaring

1. Identifier:

```
lhs |> f      →  f(lhs)
```

2. Call:

```
lhs |> f(a,b) →  f(lhs, a, b)
```

3. Placeholder `_`:

```
lhs |> f(_, a) → f(lhs, a)
```

4. Do-block:

```
lhs |> do { ...it... }
```

desugars to:

```
{ let it = lhs; ... }
```

### 13.3 Precedence

* Binds tighter than `||`
* Looser than `?`
* Left-associative

---

# **14. Commands (Program Entrypoints)**

### 14.1 Declaration

```
cmd main(ctx CmdContext) Result<void, error> {
    ...
}
```

### 14.2 Semantics

* Commands are not normal functions.
* They define top-level entrypoints executable by `yar build`.
* The runtime handles:

  * flag parsing
  * environment access
  * output streams
  * process exit

### 14.3 One Entrypoint Per Build

v0 always builds a CLI binary using `cmd main`.

---

# **15. Standard Library (v0 Minimum)**

The v0 prelude exposes:

```
len
println
panic
Result
Option
map
```

And minimal helpers for slices & maps.

Excluded from v0:

* functional `map`, `filter`, `fold`
* iterators beyond builtin slice/map iteration
* concurrency
* async/await
* traits/interfaces

---

# **16. Complete v0 Example**

```yar
module courier.lab

use net.connect
use fs.load_bytes
use fmt.println
use Result.ok
use Result.err

struct Packet {
    header: Header
    payload: []byte
}

fn checksum(data: []byte) u32 {
    let acc = 0u32
    let res = fold_u32(data, acc, |x, b| rotate_left(x, 5) ^ u32(b))
    return res
}

fn send(sock: *Socket, pkt: Packet) Result<u32, error> {
    let writer = sock.writer()
    defer writer.close()

    let encoded = pkt.payload |> encode_payload()
    writer.write(encoded)?
    writer.flush()?

    return ok(checksum(pkt.payload))
}

cmd main(ctx CmdContext) Result<void, error> {
    let conn = connect("10.0.0.7", 9000)?
    let payload = load_bytes("msg.bin")?
    let pkt = Packet { header: probe_header(), payload: payload }

    when send(conn, pkt) {
        ok bytes  => println("sent", bytes, "bytes")
        err issue => println("failed:", issue)
    }

    return ok(void)
}
```

This example demonstrates the complete v0 feature set:

* modules
* imports
* structs, slices
* functions, closures
* pipelines, result propagation
* defer
* command entrypoint

No tasks. No async. No concurrency.
