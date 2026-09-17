package tasks

import (
	"fmt"
	"math"
)

// Shape 几何图形接口：面积与周长。
type Shape interface {
	Area() float64
	Perimeter() float64
}

// Rectangle 矩形。
type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64      { return r.Width * r.Height }
func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }

// Circle 圆。
type Circle struct {
	Radius float64
}

func (c Circle) Area() float64      { return math.Pi * c.Radius * c.Radius }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.Radius }

// RunShapeDemo 以接口类型统一调用不同实现（多态）。
func RunShapeDemo() {
	shapes := []Shape{
		Rectangle{Width: 5, Height: 3},
		Circle{Radius: 4},
	}
	for _, s := range shapes {
		fmt.Printf("%-12T 面积=%-8.2f 周长=%.2f\n", s, s.Area(), s.Perimeter())
	}
}
