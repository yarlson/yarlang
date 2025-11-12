CC = clang
CFLAGS = -Wall -Wextra -std=c11 -fPIC -I/opt/homebrew/include
LDFLAGS = -L/opt/homebrew/lib -lgc

all: libyarrt.a

libyarrt.a: value.o
	ar rcs libyarrt.a value.o

value.o: value.c value.h
	$(CC) $(CFLAGS) -c value.c -o value.o

clean:
	rm -f *.o *.a

.PHONY: all clean
