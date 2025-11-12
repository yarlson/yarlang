fn fib(n i32) i32 {
	if n <= 1 {
		return n
	}
	return fib(n - 1) + fib(n - 2)
}

fn main() {
	let result: i32 = fib(10)
	println("fib(10) = ")
	println(result)
}
