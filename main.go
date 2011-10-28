package main

import (
	"fmt"
)

func main() {
	d := NewDsl(2, 0, 10, 10)
	fmt.Println("new DSL")
	d.UpdateCell(3, 3, -1)
	d.UpdateCell(3, 2, -1)
	d.UpdateCell(3, 1, -1)
	d.UpdateCell(2, 3, -1)
	d.UpdateCell(2, 2, 42.432)
	d.UpdateCell(1, 3, -1)
	d.Replan()

	path := d.Path()
	for i, p := range path {
		fmt.Printf("%d: %d %d\n", i, p.X, p.Y)
	}

	d.UpdateStart(0, 0)
	d.Replan()

	path = d.Path()
	for i, p := range path {
		fmt.Printf("%d: %d %d\n", i, p.X, p.Y)
	}

	d.UpdateGoal(12, 19)
	d.Replan()
	path = d.Path()
	for i, p := range path {
		fmt.Printf("%d: %d %d\n", i, p.X, p.Y)
	}
}
