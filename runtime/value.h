#ifndef YARLANG_VALUE_H
#define YARLANG_VALUE_H

#include <stdint.h>
#include <stdbool.h>
#include <stddef.h>

// Value types
typedef enum {
    VAL_NIL = 0,
    VAL_BOOL,
    VAL_NUMBER,
    VAL_STRING,
    VAL_FUNCTION
} ValueType;

// Forward declare
typedef struct Value Value;
typedef struct FunctionValue FunctionValue;

// Function signature for native functions
typedef Value* (*NativeFn)(int argc, Value** args);

// Function value
struct FunctionValue {
    void* ptr;        // Function pointer (LLVM-generated or native)
    bool is_native;   // true for built-ins, false for user functions
    int arity;        // Number of parameters
};

// Runtime value
struct Value {
    ValueType type;
    union {
        bool boolean;
        double number;
        char* string;
        FunctionValue* function;
    } as;
};

// Value constructors
Value* yar_nil();
Value* yar_bool(bool value);
Value* yar_number(double value);
Value* yar_string(const char* str);
Value* yar_function(void* ptr, bool is_native, int arity);

// Type checks
bool yar_is_nil(Value* v);
bool yar_is_bool(Value* v);
bool yar_is_number(Value* v);
bool yar_is_string(Value* v);
bool yar_is_function(Value* v);
bool yar_is_truthy(Value* v);

// Type names
const char* yar_type_name(Value* v);

// Operators
Value* yar_add(Value* a, Value* b);
Value* yar_subtract(Value* a, Value* b);
Value* yar_multiply(Value* a, Value* b);
Value* yar_divide(Value* a, Value* b);
Value* yar_modulo(Value* a, Value* b);

Value* yar_eq(Value* a, Value* b);
Value* yar_neq(Value* a, Value* b);
Value* yar_lt(Value* a, Value* b);
Value* yar_gt(Value* a, Value* b);
Value* yar_lte(Value* a, Value* b);
Value* yar_gte(Value* a, Value* b);

Value* yar_and(Value* a, Value* b);
Value* yar_or(Value* a, Value* b);
Value* yar_not(Value* v);

Value* yar_negate(Value* v);

// Built-in functions
void yar_print(Value* v);
void yar_println(Value* v);
Value* yar_len(Value* v);
Value* yar_type(Value* v);

// Error handling
void yar_error(const char* fmt, ...);

#endif // YARLANG_VALUE_H
