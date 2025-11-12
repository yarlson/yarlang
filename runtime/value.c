#include "value.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdarg.h>
#include <gc.h>  // Boehm GC

// Memory allocation using Boehm GC
static void* yar_alloc(size_t size) {
    return GC_malloc(size);
}

// Value constructors

Value* yar_nil() {
    Value* v = yar_alloc(sizeof(Value));
    v->type = VAL_NIL;
    return v;
}

Value* yar_bool(bool value) {
    Value* v = yar_alloc(sizeof(Value));
    v->type = VAL_BOOL;
    v->as.boolean = value;
    return v;
}

Value* yar_number(double value) {
    Value* v = yar_alloc(sizeof(Value));
    v->type = VAL_NUMBER;
    v->as.number = value;
    return v;
}

Value* yar_string(const char* str) {
    Value* v = yar_alloc(sizeof(Value));
    v->type = VAL_STRING;
    size_t len = strlen(str);
    v->as.string = yar_alloc(len + 1);
    strcpy(v->as.string, str);
    return v;
}

Value* yar_function(void* ptr, bool is_native, int arity) {
    Value* v = yar_alloc(sizeof(Value));
    v->type = VAL_FUNCTION;
    v->as.function = yar_alloc(sizeof(FunctionValue));
    v->as.function->ptr = ptr;
    v->as.function->is_native = is_native;
    v->as.function->arity = arity;
    return v;
}

// Type checks

bool yar_is_nil(Value* v) {
    return v->type == VAL_NIL;
}

bool yar_is_bool(Value* v) {
    return v->type == VAL_BOOL;
}

bool yar_is_number(Value* v) {
    return v->type == VAL_NUMBER;
}

bool yar_is_string(Value* v) {
    return v->type == VAL_STRING;
}

bool yar_is_function(Value* v) {
    return v->type == VAL_FUNCTION;
}

bool yar_is_truthy(Value* v) {
    if (yar_is_nil(v)) return false;
    if (yar_is_bool(v)) return v->as.boolean;
    return true;  // numbers, strings, functions are truthy
}

const char* yar_type_name(Value* v) {
    switch (v->type) {
        case VAL_NIL: return "nil";
        case VAL_BOOL: return "bool";
        case VAL_NUMBER: return "number";
        case VAL_STRING: return "string";
        case VAL_FUNCTION: return "function";
        default: return "unknown";
    }
}

// Operators

Value* yar_add(Value* a, Value* b) {
    if (yar_is_number(a) && yar_is_number(b)) {
        return yar_number(a->as.number + b->as.number);
    }
    if (yar_is_string(a) && yar_is_string(b)) {
        size_t len = strlen(a->as.string) + strlen(b->as.string);
        char* result = yar_alloc(len + 1);
        strcpy(result, a->as.string);
        strcat(result, b->as.string);
        Value* v = yar_alloc(sizeof(Value));
        v->type = VAL_STRING;
        v->as.string = result;
        return v;
    }
    yar_error("Cannot add %s and %s", yar_type_name(a), yar_type_name(b));
    return yar_nil();
}

Value* yar_subtract(Value* a, Value* b) {
    if (yar_is_number(a) && yar_is_number(b)) {
        return yar_number(a->as.number - b->as.number);
    }
    yar_error("Cannot subtract %s and %s", yar_type_name(a), yar_type_name(b));
    return yar_nil();
}

Value* yar_multiply(Value* a, Value* b) {
    if (yar_is_number(a) && yar_is_number(b)) {
        return yar_number(a->as.number * b->as.number);
    }
    yar_error("Cannot multiply %s and %s", yar_type_name(a), yar_type_name(b));
    return yar_nil();
}

Value* yar_divide(Value* a, Value* b) {
    if (yar_is_number(a) && yar_is_number(b)) {
        if (b->as.number == 0) {
            yar_error("Division by zero");
            return yar_nil();
        }
        return yar_number(a->as.number / b->as.number);
    }
    yar_error("Cannot divide %s and %s", yar_type_name(a), yar_type_name(b));
    return yar_nil();
}

Value* yar_modulo(Value* a, Value* b) {
    if (yar_is_number(a) && yar_is_number(b)) {
        return yar_number((long)a->as.number % (long)b->as.number);
    }
    yar_error("Cannot modulo %s and %s", yar_type_name(a), yar_type_name(b));
    return yar_nil();
}

Value* yar_eq(Value* a, Value* b) {
    if (a->type != b->type) return yar_bool(false);
    switch (a->type) {
        case VAL_NIL: return yar_bool(true);
        case VAL_BOOL: return yar_bool(a->as.boolean == b->as.boolean);
        case VAL_NUMBER: return yar_bool(a->as.number == b->as.number);
        case VAL_STRING: return yar_bool(strcmp(a->as.string, b->as.string) == 0);
        default: return yar_bool(false);
    }
}

Value* yar_neq(Value* a, Value* b) {
    Value* eq = yar_eq(a, b);
    return yar_bool(!eq->as.boolean);
}

Value* yar_lt(Value* a, Value* b) {
    if (yar_is_number(a) && yar_is_number(b)) {
        return yar_bool(a->as.number < b->as.number);
    }
    yar_error("Cannot compare %s and %s", yar_type_name(a), yar_type_name(b));
    return yar_bool(false);
}

Value* yar_gt(Value* a, Value* b) {
    if (yar_is_number(a) && yar_is_number(b)) {
        return yar_bool(a->as.number > b->as.number);
    }
    yar_error("Cannot compare %s and %s", yar_type_name(a), yar_type_name(b));
    return yar_bool(false);
}

Value* yar_lte(Value* a, Value* b) {
    if (yar_is_number(a) && yar_is_number(b)) {
        return yar_bool(a->as.number <= b->as.number);
    }
    yar_error("Cannot compare %s and %s", yar_type_name(a), yar_type_name(b));
    return yar_bool(false);
}

Value* yar_gte(Value* a, Value* b) {
    if (yar_is_number(a) && yar_is_number(b)) {
        return yar_bool(a->as.number >= b->as.number);
    }
    yar_error("Cannot compare %s and %s", yar_type_name(a), yar_type_name(b));
    return yar_bool(false);
}

Value* yar_and(Value* a, Value* b) {
    return yar_bool(yar_is_truthy(a) && yar_is_truthy(b));
}

Value* yar_or(Value* a, Value* b) {
    return yar_bool(yar_is_truthy(a) || yar_is_truthy(b));
}

Value* yar_not(Value* v) {
    return yar_bool(!yar_is_truthy(v));
}

Value* yar_negate(Value* v) {
    if (yar_is_number(v)) {
        return yar_number(-v->as.number);
    }
    yar_error("Cannot negate %s", yar_type_name(v));
    return yar_nil();
}

// Built-in functions

void yar_print(Value* v) {
    switch (v->type) {
        case VAL_NIL:
            printf("nil");
            break;
        case VAL_BOOL:
            printf("%s", v->as.boolean ? "true" : "false");
            break;
        case VAL_NUMBER:
            printf("%g", v->as.number);
            break;
        case VAL_STRING:
            printf("%s", v->as.string);
            break;
        case VAL_FUNCTION:
            printf("<function>");
            break;
    }
}

void yar_println(Value* v) {
    yar_print(v);
    printf("\n");
}

Value* yar_len(Value* v) {
    if (yar_is_string(v)) {
        return yar_number((double)strlen(v->as.string));
    }
    yar_error("len() requires string, got %s", yar_type_name(v));
    return yar_nil();
}

Value* yar_type(Value* v) {
    return yar_string(yar_type_name(v));
}

// Error handling

void yar_error(const char* fmt, ...) {
    va_list args;
    va_start(args, fmt);
    fprintf(stderr, "Runtime error: ");
    vfprintf(stderr, fmt, args);
    fprintf(stderr, "\n");
    va_end(args);
    exit(1);
}
