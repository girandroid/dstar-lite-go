package main

import (
	"image"
	"image/png"
	"image/color"
	"log"
	"os"
	"fmt"
	"time"
	"math"
)

func main() {
	f, err := os.Create("new.png")
	if err != nil {
		log.Fatal(err)
	}

	d := NewDsl(1, 1, 63, 63)

	m := image.NewNRGBA(image.Rect(0, 0, 64, 64))
	ok := 20
	for x := 0; x < 64; x+=16 {
		for y := 0; y < 64; y++ {
			if math.Abs(float64(y - ok))  < 3 {
				m.Set(int(x), int(y), color.RGBA{0, 0, 64, 255})
			} else {
				d.UpdateCell(int32(x), int32(y), -1)
				m.Set(int(x), int(y), color.RGBA{255, 0, 0, 255})
			}
		}
	}
  
	for x:=0; x <= 64; x++ {
		d.UpdateCell(int32(x), 0, -1)
				m.Set(int(x), 0, color.RGBA{255, 0, 0 , 255})
		d.UpdateCell(0, int32(x), -1)
				m.Set(0, int(x), color.RGBA{255,0, 0, 255})
		d.UpdateCell(int32(x), 64, -1)
				m.Set(int(x), 64, color.RGBA{255,0, 0, 255})
		d.UpdateCell(64,int32(x), -1)
				m.Set(64, int(x), color.RGBA{255,0, 0, 255})
  }
	t1 := time.Nanoseconds()
	b := d.Replan()
	t2 := time.Nanoseconds()
  
	fmt.Printf("Time :%d ns", t2-t1)
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
