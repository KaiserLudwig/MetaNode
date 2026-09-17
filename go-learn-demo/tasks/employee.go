package tasks

import "fmt"

// Person 人：姓名与年龄。
type Person struct {
	Name string
	Age  int
}

// Employee 员工：组合 Person，并扩展员工编号。
type Employee struct {
	Person     // 匿名组合：Person 的字段和方法被"提升"
	EmployeeID string
}

// PrintInfo 输出员工信息（接收者使用值类型）。
func (e Employee) PrintInfo() {
	fmt.Printf("员工信息: ID=%s 姓名=%s 年龄=%d\n", e.EmployeeID, e.Name, e.Age)
}

// RunEmployeeDemo 演示组合与字段提升。
func RunEmployeeDemo() {
	emp := Employee{
		Person:     Person{Name: "张三", Age: 28},
		EmployeeID: "E1001",
	}
	emp.PrintInfo()

	// 组合字段被提升，可直接访问和修改
	emp.Name = "李四"
	emp.Age = 30
	emp.PrintInfo()
}
