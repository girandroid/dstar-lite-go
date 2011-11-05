package main

import (
	"testing"
	"rand"
)

func BenchmarkUpdateCellRandom(b *testing.B) {
	d := NewDsl(0, 0, 800, 600)
	for j := 0; j < b.N; j++ {
		x := int32(rand.Intn(800))
		y := int32(rand.Intn(600))
		cost := rand.Float64() * 100.0

		d.UpdateCell(x, y, cost)
	}
	d.Replan()
}
func BenchmarkUpdateCellOne(b *testing.B) {
	d := NewDsl(0, 0, 800, 600)

	x := int32(rand.Intn(800))
	y := int32(rand.Intn(600))
	cost := rand.Float64() * 100.0

	for j := 0; j < b.N; j++ {
		d.UpdateCell(x, y, cost)
	}
	d.Replan()
}

func BenchmarkUpdateGoal(b *testing.B) {
	d := NewDsl(0, 0, 800, 600)
	for j := 0; j < b.N; j++ {
		x := int32(rand.Intn(800))
		y := int32(rand.Intn(600))

		d.UpdateGoal(x, y)
	}
	d.Replan()
}

func BenchmarkUpdateStart(b *testing.B) {
	d := NewDsl(0, 0, 800, 600)
	for j := 0; j < b.N; j++ {
		x := int32(rand.Intn(800))
		y := int32(rand.Intn(600))

		d.UpdateStart(x, y)
	}
	d.Replan()
}

func BenchmarkReplan(b *testing.B) {
	for i := 0; i < b.N; i++ {
		d := NewDsl(0, 0, 256, 512)
		b.StopTimer()
		for j := 0; j < 99; j++ {
			x := int32(rand.Intn(256))
			y := int32(rand.Intn(512))
			cost := rand.Float64() * 100.0

			d.UpdateCell(x, y, cost)
		}
		b.StartTimer()
		d.Replan()
	}
}
