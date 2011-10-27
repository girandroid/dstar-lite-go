package main

import (
	"math"
	"sort"
	"fmt"
)

var (
	maxSteps uint32  = 800 // node extensions before we give up
	C1       float64 = 1.0 // cost of unseen cell
	eps      float64 = 0.00001
	M_SQRT2  float64 = math.Sqrt(2.0)
)

type PQ struct {
	states []State
}

func (pq *PQ) IsEmpty() bool {
	return len(pq.states) == 0
}

func (pq *PQ) Clear() {
	n := make([]State, 0, 1000)
	*pq = *NewPQ(n)
}

func (pq *PQ) Top() *State {
	t := pq.states[0]
	pq.states = pq.states[1:]
	return &t
}

func (pq *PQ) Peek() *State {
	return &pq.states[0]
}

func (pq *PQ) Len() int {
	return len(pq.states)
}

func (pq *PQ) Less(i, j int) bool {
	s1 := pq.states[i]
	s2 := pq.states[j]

	return s2.Lt(s1)
}

func (pq *PQ) Swap(i, j int) {
	pq.states[i], pq.states[j] = pq.states[j], pq.states[i]
}

func (pq *PQ) Sort() {
	sort.Sort(pq)
	for i, l := 0, len(pq.states); i < l/2; i++ {
		pq.states[i], pq.states[l-i-1] = pq.states[l-i-1], pq.states[i]
	}
}

func (pq *PQ) Add(u State) {
	pq.states = append(pq.states, u)
	pq.Sort()
}

func NewPQ(s []State) *PQ {
	pq := new(PQ)
	pq.states = s
	pq.Sort()
	return pq
}

type State struct {
	x, y              int32
	k_first, k_second float64
}

func NewState(x, y int32, k_first, k_second float64) *State {
	s := new(State)

	s.x = x
	s.y = y
	s.k_first = k_first
	s.k_second = k_second

	return s
}

func (s State) Eq(s2 State) bool {
	return ((s.x == s2.x) && (s.y == s2.y))
}

func (s State) Neq(s2 State) bool {
	return ((s.x != s2.x) || (s.y != s2.y))
}

func (s State) Gt(s2 State) bool {
	if s.k_first-eps > s2.k_first {
		return true
	} else if s.k_first < s2.k_first-eps {
		return false
	}
	return s.k_second > s2.k_second
}

func (s State) Lte(s2 State) bool {
	if s.k_first < s2.k_first {
		return true
	} else if s.k_first > s2.k_second {
		return false
	}
	return s.k_second < s2.k_second+eps
}

func (s State) Lt(s2 State) bool {
	if s.k_first+eps < s2.k_first {
		return true
	} else if s.k_first-eps > s2.k_first {
		return false
	}
	return s.k_second < s2.k_second
}

func (s State) Hash() int32 {
	return s.x + 34245*s.y
}

type Ipoint2 struct {
	x, y int32
}

func NewIpoint2(x, y int32) *Ipoint2 {
	p := new(Ipoint2)
	p.x = x
	p.y = y
	return p
}

type CellInfo struct {
	g, rhs, cost float64
}

type Dsl struct {
	path []State

	C1, k_m float64

	start, goal, last State

	maxSteps uint32

	openList PQ
	cellHash map[*State]CellInfo
	openHash map[*State]float64
}

func NewDsl(sX, sY, gX, gY int32) *Dsl {
	d := new(Dsl)

	d.maxSteps = 80000
	d.C1 = 1.0

	d.cellHash = make(map[*State]CellInfo)
	d.openHash = make(map[*State]float64)
	d.path = make([]State, 0, 1000)

	d.k_m = 0.0

	d.start.x = sX
	d.start.y = sY

	d.goal.x = gX
	d.goal.y = gY

	var tmp CellInfo
	tmp.g = 0
	tmp.rhs = 0
	tmp.cost = d.C1

	d.cellHash[&d.goal] = tmp

	var tmp2 CellInfo
	heuristic := d.heuristic(d.start, d.goal)
	tmp2.g = heuristic
	tmp2.rhs = heuristic
	tmp2.cost = d.C1

	d.cellHash[&d.start] = tmp2
	d.start = d.calculateKey(d.start)
	d.last = d.start

	return d
}

func (d *Dsl) heuristic(a, b State) float64 {
	return d.eightCondist(a, b) * d.C1
}

func (d *Dsl) eightCondist(a, b State) float64 {
	var min float64 = math.Fabs(float64(a.x) - float64(b.x))
	var max float64 = math.Fabs(float64(a.y) - float64(b.y))

	if min > max {
		min, max = max, min
	}

	return ((math.Sqrt(2)-1.0)*min + max)
}

func (d *Dsl) trueDist(a, b State) float64 {
	var x float64 = float64(a.x) - float64(b.x)
	var y float64 = float64(a.y) - float64(b.y)

	return math.Sqrt(x*x + y*y)
}

func (d *Dsl) UpdateCell(x, y int32, val float64) {
	u := NewState(x, y, 0, 0)

	if u.Eq(d.start) || u.Eq(d.goal) {
		return
	}

	d.makeNewCell(*u)

	tmp, _ := d.cellHash[u]
	tmp.cost = val
	d.cellHash[u] = tmp

	d.updateVertex(*u)
}

func (d *Dsl) makeNewCell(u State) {
	_, ok := d.cellHash[&u]
	if ok {
		return
	}
	var tmp CellInfo
	h := d.heuristic(u, d.goal)
	tmp.g = h
	tmp.rhs = h
	tmp.cost = d.C1

	d.cellHash[&u] = tmp
}

func (d *Dsl) updateVertex(u State) {
	var s PQ
	if u.Neq(d.goal) {
		s = d.getSucc(u)

		var tmp float64 = math.Inf(1)
		for _, i := range s.states {
			tmp2 := d.getG(i) + d.cost(u, i)
			if tmp2 < tmp {
				tmp = tmp2
			}
		}

		if !d.Close(d.getRHS(u), tmp) {
			d.setRHS(u, tmp)
		}
	}
	insert := !d.Close(d.getG(u), d.getRHS(u))
	if insert {
		d.insert(u)
	}
}

func (d *Dsl) Close(x, y float64) bool {
	if x == math.Inf(1) && y == math.Inf(1) {
		return true
	}

	return math.Fabs(x-y) < eps
}

func (d *Dsl) getSucc(u State) PQ {
	ns := make([]State, 0, 100000)

	if d.occupied(u) {
		return *NewPQ(ns)
	}

	ns = append(ns, *NewState(u.x+1, u.y, -1, -1))
	ns = append(ns, *NewState(u.x+1, u.y+1, -1, -1))
	ns = append(ns, *NewState(u.x, u.y+1, -1, -1))
	ns = append(ns, *NewState(u.x-1, u.y+1, -1, -1))
	ns = append(ns, *NewState(u.x-1, u.y, -1, -1))
	ns = append(ns, *NewState(u.x-1, u.y-1, -1, -1))
	ns = append(ns, *NewState(u.x, u.y-1, -1, -1))
	ns = append(ns, *NewState(u.x+1, u.y-1, -1, -1))

	return *NewPQ(ns)
}

func (d *Dsl) getPred(u State) PQ {
	ns := make([]State, 0, 100000)

	if !d.occupied(u) {
		ns = append(ns, *NewState(u.x+1, u.y, -1, -1))
	}
	if !d.occupied(u) {
		ns = append(ns, *NewState(u.x+1, u.y+1, -1, -1))
	}
	if !d.occupied(u) {
		ns = append(ns, *NewState(u.x, u.y+1, -1, -1))
	}
	if !d.occupied(u) {
		ns = append(ns, *NewState(u.x-1, u.y+1, -1, -1))
	}
	if !d.occupied(u) {
		ns = append(ns, *NewState(u.x-1, u.y, -1, -1))
	}
	if !d.occupied(u) {
		ns = append(ns, *NewState(u.x-1, u.y-1, -1, -1))
	}
	if !d.occupied(u) {
		ns = append(ns, *NewState(u.x, u.y-1, -1, -1))
	}
	if !d.occupied(u) {
		ns = append(ns, *NewState(u.x+1, u.y-1, -1, -1))
	}
	return *NewPQ(ns)
}

func (d *Dsl) getG(u State) float64 {
	tmp, ok := d.cellHash[&u]
	if !ok {
		return d.heuristic(u, d.goal)
	}
	return tmp.g
}

func (d *Dsl) getRHS(u State) float64 {
	if u.Eq(d.goal) {
		return 0
	}

	tmp, ok := d.cellHash[&u]
	if !ok {
		return d.heuristic(u, d.goal)
	}

	return tmp.rhs
}

func (d *Dsl) setRHS(u State, rhs float64) {
	d.makeNewCell(u)
	tmp, _ := d.cellHash[&u]
	tmp.rhs = rhs
	d.cellHash[&u] = tmp
}

func (d *Dsl) setG(u State, g float64) {
	d.makeNewCell(u)
	tmp, _ := d.cellHash[&u]
	tmp.g = g
	d.cellHash[&u] = tmp
}

func (d *Dsl) cost(a, b State) float64 {
	xd := math.Fabs(float64(a.x - b.x))
	yd := math.Fabs(float64(a.y - b.y))

	var scale float64 = 1
	if xd+yd > 1 {
		scale = M_SQRT2
	}

	it, ok := d.cellHash[&a]
	if !ok {
		return scale * C1
	}

	return scale * it.cost
}

func (d *Dsl) insert(u State) {
	var csum float64

	u = d.calculateKey(u)
	csum = d.keyHashCode(u)

	d.openHash[&u] = csum
	d.openList.Add(u)
}

func (d *Dsl) occupied(u State) bool {
	it, ok := d.cellHash[&u]
	if !ok {
		return false
	}

	return it.cost < 0
}

func (d *Dsl) calculateKey(u State) State {
	val := math.Fmin(d.getRHS(u), d.getG(u))

	u.k_first = val + d.heuristic(u, d.start) + d.k_m
	u.k_second = val

	return u
}

func (d *Dsl) keyHashCode(u State) float64 {
	return u.k_first + 1193*u.k_second
}

func (d *Dsl) Path() []State {
	return d.path
}

func (d *Dsl) computeShortestPath() int32 {

	if d.openList.IsEmpty() {
		return 1
	}

	var k uint32
	for (!d.openList.IsEmpty()) && ((d.openList.Peek()).Lt((d.calculateKey(d.start)))) || (d.getRHS(d.start) != d.getG(d.start)) {

		if k > maxSteps {
			return -1
		}
		k++

		test := (d.getRHS(d.start) != d.getG(d.start))

		var u State
		for {
			if d.openList.IsEmpty() {
				return 1
			}

			u = *d.openList.Top()
			if !d.isValid(u) {
				continue
			}

			if !(u.Lt(d.start)) && !test {
				return 2
			}

			break
		}

		d.openHash[&u] = 0, false

		k_old := NewState(u.x, u.y, u.k_first, u.k_second)

		if k_old.Lt(d.calculateKey(u)) {
			d.insert(u)
		} else if d.getG(u) > d.getRHS(u) {
			d.setG(u, d.getRHS(u))
			s := d.getPred(u)
			for _, i := range s.states {
				d.updateVertex(i)
			}
		} else {
			d.setG(u, math.Inf(1))
			s := d.getPred(u)
			for _, i := range s.states {
				d.updateVertex(i)
			}
			d.updateVertex(u)
		}
	}

	return 0
}

func (d *Dsl) isValid(u State) bool {
	i, ok := d.openHash[&u]
	if !ok {
		return false
	}

	if !d.Close(d.keyHashCode(u), i) {
		return false
	}

	return true
}

func (d *Dsl) UpdateGoal(x, y int32) {
	addPoints := make(map[*State]Ipoint2)
	addCosts := make(map[*State]float64)

	for i, h := range d.cellHash {
		if !d.Close(h.cost, C1) {
			addPoints[i] = *NewIpoint2(i.x, i.y)
			addCosts[i] = h.cost
		}
	}

	d.cellHash = make(map[*State]CellInfo)
	d.openHash = make(map[*State]float64)
	d.path = make([]State, 0, 1000)

	d.k_m = 0.0

	d.goal.x = x
	d.goal.y = y

	var tmp CellInfo
	tmp.g = 0
	tmp.rhs = 0
	tmp.cost = d.C1

	d.cellHash[&d.goal] = tmp

	var tmp2 CellInfo
	heuristic := d.heuristic(d.start, d.goal)
	tmp2.g = heuristic
	tmp2.rhs = heuristic
	tmp2.cost = d.C1

	d.cellHash[&d.start] = tmp2
	d.start = d.calculateKey(d.start)
	d.last = d.start

	//addPoints := make(map[*State]Ipoint2)
	//addCosts  := make(map[*State]float64)
	for i, v := range addPoints {
		d.UpdateCell(v.x, v.y, addCosts[i])
	}
}

func (d *Dsl) UpdateStart(x, y int32) {
	d.start.x = x
	d.start.x = y

	d.k_m += d.heuristic(d.last, d.start)

	d.start = d.calculateKey(d.start)
	d.last = d.start
}

func (d *Dsl) Replan() bool {
	d.path = make([]State, 0, 1000)

	res := d.computeShortestPath()
	if res < 0 {
		return false
	}

	var cur State = d.start
	if d.getG(d.start) == math.Inf(1) {
		return false
	}

	for cur.Neq(d.goal) {
		d.path = append(d.path, cur)
		n := d.getSucc(cur)
		if n.IsEmpty() {
			return false
		}

		var cmin float64 = math.Inf(1)
		var tmin float64 = 0
		var smin = NewState(0, 0, 0, 0)

		for _, i := range n.states {
			var val float64 = d.cost(cur, i)
			var val2 float64 = d.trueDist(i, d.goal) + d.trueDist(d.start, i)
			val += d.getG(i)

			iclose := d.Close(val, cmin)
			if iclose && tmin > val2 {
				tmin = val2
				cmin = val
				smin = NewState(i.x, i.y, i.k_first, i.k_second)
			} else if val < cmin {
				tmin = val2
				cmin = val
				smin = NewState(i.x, i.y, i.k_first, i.k_second)
			}
		}
		n.Clear()
		cur = *NewState(smin.x, smin.y, smin.k_first, smin.k_second)
	}

	d.path = append(d.path, d.goal)
	return true
}

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

	path1 := d.Path()
	for i, p := range path1 {
		fmt.Printf("%d: %d %d\n", i, p.x, p.y)
	}

	d.UpdateStart(0, 0)
	d.Replan()

	path2 := d.Path()
	for i, p := range path2 {
		fmt.Printf("%d: %d %d\n", i, p.x, p.y)
	}

	d.UpdateGoal(12, 19)
	d.Replan()
	path3 := d.Path()
	for i, p := range path3 {
		fmt.Printf("%d: %d %d\n", i, p.x, p.y)
	}
}
