package main

import (
	"image"
	"image/png"
	"image/color"
	"log"
	"os"
	"rand"
	"fmt"
)

func main() {
	f, err := os.Create("new.png")
	if err != nil {
		log.Fatal(err)
	}

	d := NewDsl(0, 0, 64, 64)

	m := image.NewNRGBA(image.Rect(0, 0, 64, 64))

	for x := 1; x < 63; x+=3 {
	  ok := rand.Intn(64)
		for y := 0; y < 64; y++ {
			if y == ok{
				m.Set(int(x), int(y), color.RGBA{0, 0, 64, 255})
			} else {
				fmt.Printf("unmovable: %d %d\n", x, y)
				d.UpdateCell(int32(x), int32(y), -1)
				m.Set(int(x), int(y), color.RGBA{255, 0, 0, 255})
			}
		}
	}

	b := d.Replan()
	if !b {
		fmt.Println("No Path")
	} 

	path := d.Path()
	for i, p := range path {
		m.Set(int(p.X), int(p.Y), color.RGBA{0, 255, 0, 255})
		fmt.Printf("%d:  %d %d\n", i, p.X, p.Y)
	}

	if err = png.Encode(f, m); err != nil {
		log.Fatal(err)
	}
}
