package tasks

import "fmt"

// doubleSlice 接收一个整数切片的指针，把每个元素乘以 2。
func doubleSlice(s *[]int) {
	for i := range *s {
		(*s)[i] *= 2
	}
}

// RunPointerSliceDemo 演示通过切片指针修改底层数组。
func RunPointerSliceDemo() {
	nums := []int{1, 2, 3, 4, 5}
	fmt.Printf("修改前: %v\n", nums)
	doubleSlice(&nums)
	fmt.Printf("修改后: %v\n", nums)
}
