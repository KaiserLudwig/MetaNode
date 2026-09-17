package tasks

import "fmt"

// addTen 接收一个整数指针，把指针指向的值增加 10。
//
// 传指针相当于引用传递：函数内通过 *p 修改，会直接作用到调用方的变量上。
func addTen(p *int) {
	*p += 10
}

// RunPointerDemo 演示指针参数修改外部变量。
func RunPointerDemo() {
	n := 5
	fmt.Printf("修改前: n = %d\n", n)
	addTen(&n) // &n 取出 n 的地址传给函数
	fmt.Printf("修改后: n = %d  （指针修改生效，这就是引用传递）\n", n)
}
