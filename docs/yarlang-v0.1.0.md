# YarLang v0.1.0 — Minimal Compiler-Facing Spec

## 0) Design stance (what to actually implement)

- **Go-look, Rust-core**: simple surface, real ownership/borrows, `Result<T,E>` + `?`, RAII + `defer`.
- **Semicolons optional** via ASI (rules below). Braces define blocks.
- **Generics use `<T>`**, not `[]`.
- **No macros, no async/await, no pattern-matching** (v0.4).
- **Concurrency**: library, not syntax.

---

## 1) Lexical structure

**Encoding**: UTF-8 (BOM ignored).
**Whitespace**: space, tab, newline.
**Comments**: `//...` line, `/* ... */` block (no nesting).
**Identifiers**: `[A-Za-z_][A-Za-z0-9_]*` (ASCII for v0.4).

**Keywords (reserved)**

```
as, break, const, continue, defer, else, enum, extern, false, fn, for, if,
impl, let, module, mut, nil, pub, return, struct, trait, true, type,
unsafe, use, void
```

**Literals**

- **Int**: `123`, `0xFF`, `0b1010`, `0o755` (underscores allowed).
- **Float**: `1.0`, `.5`, `5.`, `1e9`, `3.14e-2` (underscores allowed).
- **Char**: `'a'`, `'\n'`, `'\u00A9'`.
- **String**: `"..."` with escapes `\\ \" \n \r \t \0 \xHH \uHHHH`.

**Punctuators / operators**

```
( ) { } [ ] , . ; : :: -> =
+ - * / % & | ^ ~ !
< > <= >= == !=
&& ||
<< >>
+= -= *= /= %= &= |= ^= <<= >>=
?                  // postfix error propagation
& *                // prefix address-of / deref
```

---

## 2) Automatic Semicolon Insertion (ASI)

Insert a virtual `;` **at a newline** if the token before the newline is one of:

- `IDENT`, literals (`int/float/char/string/true/false/nil`), `)`, `]`, `}`
- `break`, `continue`, `return`

**Do not insert** if the next token begins with: `, . :: : ? ) ] } else` or any binary operator.

(You may also allow explicit `;` anywhere; treat it as a statement terminator.)

---

## 3) Types (surface)

**Built-ins**
`i8 i16 i32 i64 isize  u8 u16 u32 u64 usize  f32 f64  bool char void`

**Constructed**

- Array: `[T; N]`
- Slice (borrowed view): `[]T` (lowered as `{*T,len}`)
- Tuple: `(T1, T2, ...)` (unit = `()`)
- References: `&T` (shared), `&mut T` (exclusive)
- Raw pointer: `*T` (unsafe)
- Paths: `pkg::Type`
- Generics: `Name<T, U>`

**User-defined**: `struct`, `enum`, `type` alias, `trait`, `impl`.

---

## 4) Declarations & modules

**Modules**

```
module_decl  := "module" path
path         := IDENT { "::" IDENT }
```

**Imports**

```
use_decl     := "use" path [ "as" IDENT ]
```

**Type alias / const**

```
type_alias   := "type" IDENT "=" type
const_decl   := "const" IDENT ":" type "=" expr
```

**Struct / enum / trait / impl**
(Require commas between fields/variants in v0.4 to keep parser trivial.)

```
struct_decl  := ["pub"] "struct" IDENT [ "<" tparams ">" ] "{" [ field { "," field } [ "," ] ] "}"
field        := IDENT ":" type

enum_decl    := ["pub"] "enum" IDENT [ "<" tparams ">" ] "{" [ variant { "," variant } [ "," ] ] "}"
variant      := IDENT [ "(" type { "," type } ")" ]

trait_decl   := ["pub"] "trait" IDENT [ "<" tparams ">" ] "{" { trait_sig ";" } "}"
trait_sig    := "fn" IDENT "(" params ")" [ "->" type ]

impl_block   := "impl" [ trait_ref "for" ] type "{" { function } "}"
trait_ref    := path [ "<" type { "," type } ">" ]
tparams      := IDENT { "," IDENT }
```

---

## 5) Functions (Go-look params)

**Header**

```
function     := ["pub"] "fn" IDENT [ "<" tparams ">" ] "(" params ")" [ ret_type ] block
params       := [ param { "," param } [ "," ] ]
param        := ["mut"] IDENT type_as
type_as      := IDENT | ":" type           // allow Go-look (name Type) or Rust-look (name: Type)
ret_type     := type | "->" type           // allow Go-look (after parens) or Rust-look arrow
```

Recommended consistent style for v0.4: **Go-look**: `fn f(a T, b U) R`.

---

## 6) Statements (including `let` + `:=`)

```
block        := "{" { stmt } "}"

stmt         := let_stmt
             | short_decl
             | assign_stmt
             | const_stmt
             | expr_stmt
             | return_stmt
             | if_stmt
             | while_stmt
             | for_stmt
             | break_stmt
             | continue_stmt
             | defer_stmt
             | unsafe_block

let_stmt     := "let" ["mut"] IDENT [ type_as ] "=" expr
short_decl   := IDENT ":=" expr
assign_stmt  := IDENT "=" expr
const_stmt   := "const" IDENT ":" type "=" expr
expr_stmt    := expr
return_stmt  := "return" [ expr ]
break_stmt   := "break"
continue_stmt:= "continue"
defer_stmt   := "defer" expr
unsafe_block := "unsafe" block
```

**For / While / If**

```
if_stmt      := "if" expr block [ "else" ( block | if_stmt ) ]
while_stmt   := "while" expr block
for_stmt     := "for" for_head block
for_head     := IDENT "in" expr                 // x in xs   OR i in 0..n
             | IDENT "," IDENT "in" expr        // i, x in iter(expr)
```

**Semantic rules for bindings**

- `let` always declares a **new** name in current scope (shadowing allowed).
- `short_decl` (`:=`) declares a new name **only if** not already declared in current scope; otherwise error.
- `assign_stmt` requires an existing **mutable** binding.
- `mut` only valid with `let`.
- Type inference allowed when annotation omitted.

---

## 7) Expressions & precedence

**Precedence (low → high)**

1. Assignment: `= += -= *= /= %= &= |= ^= <<= >>=` (right-assoc)
2. Logical OR: `||`
3. Logical AND: `&&`
4. Bitwise OR: `|`
5. Bitwise XOR: `^`
6. Bitwise AND: `&`
7. Equality: `== !=`
8. Relational: `< > <= >=`
9. Shift: `<< >>`
10. Additive: `+ -`
11. Multiplicative: `* / %`
12. Unary: `+ - ! ~ & &mut *` (right-assoc)
13. Postfix: call `()`, index `[]`, field `.`, propagate `?` (left-to-right)

**Grammar**

```
expr         := assign
assign       := logic_or { assign_op logic_or }
assign_op    := "=" | "+=" | "-=" | "*=" | "/=" | "%=" | "&=" | "|=" | "^=" | "<<=" | ">>="

logic_or     := logic_and { "||" logic_and }
logic_and    := bit_or    { "&&" bit_or }
bit_or       := bit_xor   { "|" bit_xor }
bit_xor      := bit_and   { "^" bit_and }
bit_and      := equality  { "&" equality }
equality     := relational { ("==" | "!=") relational }
relational   := shift     { ("<" | ">" | "<=" | ">=") shift }
shift        := additive  { ("<<" | ">>") additive }
additive     := mult      { ("+" | "-") mult }
mult         := unary     { ("*" | "/" | "%") unary }

unary        := postfix
             | ("+" | "-" | "!" | "~" | "&" | "*" ) unary
             | "&" "mut" unary

postfix      := primary { call | index | field | propagate }
call         := "(" [ arg_list ] ")"
arg_list     := expr { "," expr } [ "," ]
index        := "[" expr "]"
field        := "." IDENT
propagate    := "?"                         // on Result<T,E>

primary      := literal
             | IDENT
             | path
             | "(" expr ")"
             | tuple
             | array
             | struct_lit

tuple        := "(" expr "," expr { "," expr } [ "," ] ")"
array        := "[" [ expr { "," expr } [ "," ] ] "]"
struct_lit   := path "{" [ init { "," init } [ "," ] ] "}"
init         := IDENT ":" expr
```

**Path expression**

```
path         := IDENT { "::" IDENT }
```

---

## 8) Minimal semantics (enough for HIR/MIR)

- **Ownership**: values move by default; primitives & tuples/structs of primitives are `Copy`.
- **Borrows**: `&T` shared (read-only); `&mut T` exclusive. **No lifetime syntax**; regions inferred. If not provable → **compile error**.
- **`?`** on `Result<T,E>`: if `Err(e)`, **early return** `Err(e)` from current function; else unwrap `T`.
- **`defer`**: pushed in current block; on block exit, run **all defers LIFO**, then drop locals (RAII).
- **`unsafe {}`**: allows raw pointer deref/calls; parser just marks the block.
- **FFI**: `extern "c" fn name(... ) -> Ret;` declares a symbol; `#[repr(c)]` on aggregates (parser stores attribute only).

---

## 9) Builtins the checker can assume (predeclared)

```text
type Result<T, E> = enum { Ok(T), Err(E) }
type Option<T>    = enum { Some(T), None }

fn println(msg: []u8) -> void
fn panic(msg: []u8) -> void
fn len<T>(xs: []T) -> usize
```

(And opaque std types you can stub:)

```
struct Vec<T> { /* opaque */ }
```

---

## 10) AST sketch (Go-friendly)

```go
type File struct {
    Module *Path
    Items  []Item
}

type Item interface{ isItem() }
type (
    UseDecl    struct{ Path Path; Alias *Ident }
    ConstDecl  struct{ Name Ident; Ty Type; Value Expr }
    TypeAlias  struct{ Name Ident; Ty Type }
    StructDecl struct{ Pub bool; Name Ident; TParams []Ident; Fields []Field }
    EnumDecl   struct{ Pub bool; Name Ident; TParams []Ident; Variants []Variant }
    TraitDecl  struct{ Pub bool; Name Ident; TParams []Ident; Sigs []FnSig }
    ImplBlock  struct{ Trait *TypeRef; For Type; Fns []Func }
    Func       struct{ Pub bool; Name Ident; TParams []Ident; Params []Param; Ret Type; Body *Block }
)

type Type interface{ isType() }
type (
    TypePath  struct{ Path Path; Args []Type }
    RefType   struct{ Mut bool; Elem Type }     // &T / &mut T
    PtrType   struct{ Elem Type }               // *T
    SliceType struct{ Elem Type }               // []T
    ArrayType struct{ Elem Type; Len Expr }     // [T; N]
    TupleType struct{ Elems []Type }            // (A,B,...)
    VoidType  struct{}                          // void
)

type Stmt interface{ isStmt() }
type (
    LetStmt     struct{ Mut bool; Name Ident; Ty Type; Value Expr }
    ShortDecl   struct{ Name Ident; Value Expr }          // x := expr
    AssignStmt  struct{ Name Ident; Value Expr }          // x = expr
    ConstStmt   struct{ Name Ident; Ty Type; Value Expr }
    ExprStmt    struct{ X Expr }
    ReturnStmt  struct{ X Expr }                          // nil if bare
    IfStmt      struct{ Cond Expr; Then *Block; Else Stmt } // Else: nil|*Block|*IfStmt
    WhileStmt   struct{ Cond Expr; Body *Block }
    ForStmt     struct{ Key *Ident; Val *Ident; Iter Expr; Body *Block }
    BreakStmt   struct{}
    ContinueStmt struct{}
    DeferStmt   struct{ X Expr }
    UnsafeBlock struct{ Body *Block }
)

type Expr interface{ isExpr() }
type (
    NameRef    struct{ Name Ident }
    PathExpr   struct{ Path Path }
    Lit        struct{ Kind LitKind; Value any }
    TupleExpr  struct{ Elems []Expr }
    ArrayExpr  struct{ Elems []Expr }
    StructExpr struct{ Path Path; Inits []StructInit }

    CallExpr   struct{ Callee Expr; Args []Expr }
    IndexExpr  struct{ X Expr; Idx Expr }
    FieldExpr  struct{ X Expr; Field Ident }

    UnaryExpr  struct{ Op Token; X Expr }
    BinaryExpr struct{ Op Token; X, Y Expr }
    AssignExpr struct{ Op Token; LHS, RHS Expr } // only for compound ops if you include them

    Propagate  struct{ X Expr } // X?
    RangeExpr  struct{ Low, High Expr } // a..b  (use in for-head only)
)

type Block struct{ Stmts []Stmt }
```

---

## 11) MIR/IR (tiny starter set)

Emits SSA-ish ops:

```
Alloca, Load, Store, AddrOf, Call, CallExtern, Ret, Br, CondBr, Phi,
BinOp(Add/Sub/Mul/Div/Mod/And/Or/Xor/Shl/Shr),
Cmp(Eq/Ne/Lt/Le/Gt/Ge),
DeferPush, DeferRunAll, Panic
```

Lowering conventions:

- `[]T` → `{ ptr: *T, len: usize }`
- strings → `[]u8`
- `X?` on `Result<T,E>`:

  ```
  t = X
  if is_err(t) { return Err(extract_err(t)) }
  v = extract_ok(t)
  ```

- Run `DeferRunAll` before block epilogue, then drop locals.

---

## 12) Example (v0.4-valid; Go-look params/returns + ASI)

```yar
module demo

use core::io::File
use core::fmt::println

struct Point { x: f64, y: f64 }

impl Point {
    fn len(&self) f64 {
        (self.x*self.x + self.y*self.y).sqrt()
    }
}

fn hyp(a f64, b f64) f64 {
    let p: Point = Point{ x: a, y: b }
    return p.len()
}

fn copy(src []u8, dst []u8) Result<usize, IoErr> {
    let s := File::open(src)?
    defer s.close()

    let d := File::create(dst)?
    defer d.close()

    let buf := [0u8; 4096]
    let mut total usize = 0

    while true {
        let n := s.read(&buf)?
        if n == 0 { break }
        d.write_all(&buf[0..n])?
        total = total + n
    }
    return Ok(total)
}

fn main() {
    println("hello")
    let v := [1, 2, 3]
    for i in 0..len(v) {
        println("tick")
    }
}
```

---

## 13) Test grid (to ship the parser)

1. **Lexing**: all operators (incl. `?` postfix), literals, comments, `&mut`.
2. **ASI**: statement ends before `}` / after literals / after `return`.
3. **Decls**: `fn/struct/enum/trait/impl/use/module`, generics `<T>`.
4. **Statements**: `let`, `:=`, `=`, `if/while/for`, `defer`, `unsafe`.
5. **Expr precedence**: chained postfix (`call.index.field?`) and binary ops.
6. **Struct/array/tuple** literals and trailing commas.
7. **`?` propagation** inside expressions.
8. **Error recovery** at `;` and `}`.

That’s it. You can build a lexer, a parser, the AST above, and start lowering to MIR/LLVM today.
