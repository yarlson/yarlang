# YarLang Standard Library

## Core Types

- `Result<T, E>`: Error handling with Ok/Err variants
- `Option<T>`: Optional values with Some/None variants

## Built-in Functions

- `println(msg: []u8)`: Print message to stdout
- `panic(msg: []u8)`: Abort with error message
- `len<T>(xs: []T) usize`: Get length of slice

## Usage

These types are automatically available in all YarLang programs.
