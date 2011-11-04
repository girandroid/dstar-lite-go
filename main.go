package main

import (
	"fmt"
	"rand"
)

func main() {
	d := NewDsl(0, 0, 1000, 100)
	fmt.Println("new DSL")
	for i := 0; i < 1001; i++ {
		for j := 0; j < 101; j++ {
			r := rand.Float64()
			d.UpdateCell(int32(i), int32(j), r)
			fmt.Printf("Updated %d %d, %f\n", i, j, r)
		}
	}
	d.Replan()

	path := d.Path()
	for i, p := range path {
		fmt.Printf("%d: %d %d\n", i, p.X, p.Y)
	}
}
