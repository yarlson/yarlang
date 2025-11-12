#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

void println(const char *msg) {
    printf("%s\n", msg);
}

void println_i32(int32_t value) {
    printf("%d\n", value);
}

void println_bool(bool value) {
    printf(value ? "true\n" : "false\n");
}

void panic(const char *msg) {
    fprintf(stderr, "panic: %s\n", msg);
    exit(1);
}
