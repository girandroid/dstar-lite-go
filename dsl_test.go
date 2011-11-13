package main

import (
	"testing"
	"rand"
)

func TestOpenList(t *testing.T) {
	t.Log("new")
	ol := NewOpenList()

	s1 := &State{1, 1, 1, 1}
	s2 := &State{2, 1, 1, 2}
	s3 := &State{3, 1, 2, 2}
	s4 := &State{4, 1, 2, 3}
	s5 := &State{5, 1, 2, 4}
	s6 := &State{6, 1, 6, 1}

	ol.Add(s1)
	ol.Add(s4)
	ol.Add(s3)
	ol.Add(s6)
	ol.Add(s2)
  ol.Add(s5)

/*	for i, x := range ol.queue {
		if x != nil {
			fmt.Printf("%d : %d\n", i, x.x)
		} else {
			fmt.Printf("%d : empty\n", i)
		}
	}*/

	sX := &State{0, 1, 1, 0}
	ol.Add(sX)
	if sX != ol.queue[0] {
		t.Error("X")
	}

	sY := &State{0, 0, 0, 0}
	ol.Add(sY)
	if sY != ol.queue[0] {
		t.Error("Y")
	}

	if sY != ol.Peek() {
		t.Error("Peek")
	}

	if sY != ol.Poll() {
		t.Error("Poll")
	}
	if sX != ol.Poll() {
		t.Error("Poll")
	}
	if s1 != ol.Poll() {
		t.Error("Poll")
	}
	if s2 != ol.Poll() {
		t.Error("Poll")
	}
	if s3 != ol.Poll() {
		t.Error("Poll")
	}
	if s4 != ol.Poll() {
		t.Error("Poll")
	}
	if s5 != ol.Poll() {
		t.Error("Poll")
	}
	if s6 != ol.Poll() {
		t.Error("Poll")
	}
	if nil != ol.Poll() {
		t.Error("NIL")
	}
}

func BenchmarkUpdateCellRandom(b *testing.B) {
	b.StopTimer()
	d := NewDsl(0, 0, 800, 600)
	for j := 0; j < b.N; j++ {
		x := int32(rand.Intn(800))
		y := int32(rand.Intn(600))
		cost := rand.Float64() * 100.0

		b.StartTimer()
		d.UpdateCell(x, y, cost)
		b.StopTimer()
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
		for j := 0; j < 512; j++ {
			x := int32(rand.Intn(256))
			y := int32(rand.Intn(512))
			cost := rand.Float64() * 100.0

			d.UpdateCell(x, y, cost)
		}
		b.StartTimer()
		d.Replan()
	}
}
