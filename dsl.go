package main

import (
	"math"
	"sort"
)

var (
	maxSteps uint32  = 80000 // node extensions before we give up
	C1       float64 = 1.0 // cost of unseen cell
	eps      float64 = 0.00001
	M_SQRT2  float64 = math.Sqrt(2.0)
)

type PQueue struct {
	states []*State
}

func (pq *PQueue) IsEmpty() bool {
	return len(pq.states) == 0
}

func (pq *PQueue) Clear() {
	n := make([]*State, 0, maxSteps)
	*pq = *NewPQueue(n)
}

func (pq *PQueue) Top() *State {
	t := pq.states[0]
	pq.states = pq.states[1:]
	return t
}

func (pq *PQueue) Peek() *State {
	return pq.states[0]
}

func (pq *PQueue) Len() int {
	return len(pq.states)
}

func (pq *PQueue) Less(i, j int) bool {
	s1 := pq.states[i]
	s2 := pq.states[j]

	//return s2.Lt(s1)
	return s2.Gt(s1)
}

func (pq *PQueue) Swap(i, j int) {
	pq.states[i], pq.states[j] = pq.states[j], pq.states[i]
}

func (pq *PQueue) Sort() {
	sort.Sort(pq)
}

func (pq *PQueue) Add(u *State) {
	pq.states = append(pq.states, u)
	pq.Sort()
}

func NewPQueue(s []*State) *PQueue {
	pq := new(PQueue)
	pq.states = s
	pq.Sort()

	return pq
}

type State struct {
	x, y   int32
	k1, k2 float64
}

func (s *State) Eq(s2 *State) bool {
	return ((s.x == s2.x) && (s.y == s2.y))
}

func (s *State) Neq(s2 *State) bool {
	return ((s.x != s2.x) || (s.y != s2.y))
}

func (s *State) Gt(s2 *State) bool {
	if s.k1-eps > s2.k1 {
		return true
	} else if s.k1 < s2.k1-eps {
		return false
	}
	return s.k2 > s2.k2
}

func (s *State) Lte(s2 *State) bool {
	if s.k1 < s2.k1 {
		return true
	} else if s.k1 > s2.k2 {
		return false
	}
	return s.k2 < s2.k2+eps
}

func (s *State) Lt(s2 *State) bool {
	if s.k1+eps < s2.k1 {
		return true
	} else if s.k1-eps > s2.k1 {
		return false
	}
	return s.k2 < s2.k2
}

func (s *State) Hash() int32 {
	return s.x + 34245*s.y
}

type Point struct {
	X, Y int32
}

type CellInfo struct {
	g, rhs, cost float64
}

type CellHash struct {
	info  map[int32]*CellInfo
	cells map[int32]*State
}

func (ch *CellHash) Put(state *State, info *CellInfo) {
	ch.info[state.Hash()] = info
	ch.cells[state.Hash()] = state
}

func (ch *CellHash) Get(state *State) (*CellInfo, bool) {
	c, ok := ch.info[state.Hash()]
	return c, ok
}

func (ch *CellHash) Clear() {
	ch.info = make(map[int32]*CellInfo)
	ch.cells = make(map[int32]*State)
}

func NewCellHash() *CellHash {
	ch := new(CellHash)
	ch.Clear()

	return ch
}

type Dsl struct {
	path []Point

	k_m float64

	start, goal, last *State

	openList PQueue
	cellHash CellHash
	openHash map[int32]float64
}

func NewDsl(sX, sY, gX, gY int32) *Dsl {
	d := new(Dsl)

	d.cellHash = *NewCellHash()
	d.openHash = make(map[int32]float64)
	d.path = make([]Point, 0, maxSteps)

	d.k_m = 0.0

	d.start = &State{sX, sY, 0, 0}
	d.goal = &State{gX, gY, 0, 0}

	d.cellHash.Put(d.goal, &CellInfo{0, 0, C1})

	h := d.heuristic(d.start, d.goal)
	d.cellHash.Put(d.start, &CellInfo{h, h, C1})
	d.start = d.calculateKey(d.start)
	d.last = d.start

	return d
}

func (d *Dsl) heuristic(a, b *State) float64 {
	return d.eightCondist(a, b) * C1
}

func (d *Dsl) eightCondist(a, b *State) float64 {
	var min float64 = math.Abs(float64(a.x) - float64(b.x))
	var max float64 = math.Abs(float64(a.y) - float64(b.y))

	if min > max {
		min, max = max, min
	}

	return ((M_SQRT2-1.0)*min + max)
}

func (d *Dsl) trueDist(a, b *State) float64 {
	var x float64 = float64(a.x) - float64(b.x)
	var y float64 = float64(a.y) - float64(b.y)

	return math.Sqrt(x*x + y*y)
}

func (d *Dsl) UpdateCell(x, y int32, val float64) {
	u := &State{x, y, 0, 0}

	if u.Eq(d.start) || u.Eq(d.goal) {
		return
	}

	d.makeNewCell(u)

	tmp, _ := d.cellHash.Get(u)
	tmp.cost = val
	d.updateVertex(u)
}

func (d *Dsl) makeNewCell(u *State) {
	_, ok := d.cellHash.Get(u)
	if ok {
		return
	}

	h := d.heuristic(u, d.goal)
	d.cellHash.Put(u, &CellInfo{h, h, C1})
}

func (d *Dsl) updateVertex(u *State) {
	var s *PQueue
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

	return math.Abs(x-y) < eps
}

func (d *Dsl) getSucc(u *State) *PQueue {
	ns := make([]*State, 0, 8)
	if d.occupied(u) {
		return NewPQueue(ns)
	}

	ns = append(ns, &State{u.x + 1, u.y, -1, -1})
	ns = append(ns, &State{u.x + 1, u.y + 1, -1, -1})
	ns = append(ns, &State{u.x, u.y + 1, -1, -1})
	ns = append(ns, &State{u.x - 1, u.y + 1, -1, -1})
	ns = append(ns, &State{u.x - 1, u.y, -1, -1})
	ns = append(ns, &State{u.x - 1, u.y - 1, -1, -1})
	ns = append(ns, &State{u.x, u.y - 1, -1, -1})
	ns = append(ns, &State{u.x + 1, u.y - 1, -1, -1})

	return NewPQueue(ns)
}

func (d *Dsl) getPred(u *State) *PQueue {
	ns := make([]*State, 0, 8)
	var tempState *State

	tempState = &State{u.x + 1, u.y, -1, -1}
	if !d.occupied(tempState) {
		ns = append(ns, tempState)
	}
	tempState = &State{u.x + 1, u.y + 1, -1, -1}
	if !d.occupied(tempState) {
		ns = append(ns, tempState)
	}
	tempState = &State{u.x, u.y + 1, -1, -1}
	if !d.occupied(tempState) {
		ns = append(ns, tempState)
	}
	tempState = &State{u.x - 1, u.y + 1, -1, -1}
	if !d.occupied(tempState) {
		ns = append(ns, tempState)
	}
	tempState = &State{u.x - 1, u.y, -1, -1}
	if !d.occupied(tempState) {
		ns = append(ns, tempState)
	}
	tempState = &State{u.x - 1, u.y - 1, -1, -1}
	if !d.occupied(tempState) {
		ns = append(ns, tempState)
	}
	tempState = &State{u.x, u.y - 1, -1, -1}
	if !d.occupied(tempState) {
		ns = append(ns, tempState)
	}
	tempState = &State{u.x + 1, u.y - 1, -1, -1}
	if !d.occupied(tempState) {
		ns = append(ns, tempState)
	}

	return NewPQueue(ns)
}

func (d *Dsl) getG(u *State) float64 {
	tmp, ok := d.cellHash.Get(u)
	if !ok {
		return d.heuristic(u, d.goal)
	}
	return tmp.g
}

func (d *Dsl) getRHS(u *State) float64 {
	if u.Eq(d.goal) {
		return 0
	}

	tmp, ok := d.cellHash.Get(u)
	if !ok {
		return d.heuristic(u, d.goal)
	}

	return tmp.rhs
}

func (d *Dsl) setRHS(u *State, rhs float64) {
	d.makeNewCell(u)
	tmp, _ := d.cellHash.Get(u)
	tmp.rhs = rhs
}

func (d *Dsl) setG(u *State, g float64) {
	d.makeNewCell(u)
	tmp, _ := d.cellHash.Get(u)
	tmp.g = g
}

func (d *Dsl) cost(a, b *State) float64 {
	xd := math.Abs(float64(a.x - b.x))
	yd := math.Abs(float64(a.y - b.y))

	var scale float64 = 1
	if xd+yd > 1 {
		scale = M_SQRT2
	}

	it, ok := d.cellHash.Get(a)
	if !ok {
		return scale * C1
	}

	return scale * it.cost
}

func (d *Dsl) insert(u *State) {
	var csum float64

	u = d.calculateKey(u)
	csum = d.keyHashCode(u)

	d.openHash[u.Hash()] = csum
	d.openList.Add(u)
}

func (d *Dsl) occupied(u *State) bool {
	it, ok := d.cellHash.Get(u)
	if !ok {
		return false
	}

	return it.cost < 0
}

func (d *Dsl) calculateKey(u *State) *State {
	val := math.Min(d.getRHS(u), d.getG(u))

	u.k1 = val + d.heuristic(u, d.start) + d.k_m
	u.k2 = val

	return u
}

func (d *Dsl) keyHashCode(u *State) float64 {
	return u.k1 + 1193*u.k2
}

func (d *Dsl) Path() []Point {
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

		var u *State
		// lazy remove1
		for {
			if d.openList.IsEmpty() {
				return 1
			}

			u = d.openList.Top()
			if !d.isValid(u) {
				continue
			}

			if !(u.Lt(d.start)) && !test {
				return 2
			}

			break
		}

		d.openHash[u.Hash()] = 0, false

		k_old := State{u.x, u.y, u.k1, u.k2}

		if k_old.Lt(d.calculateKey(u)) { // u is out of date
			d.insert(u)
		} else if d.getG(u) > d.getRHS(u) { // needs update, got better
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

func (d *Dsl) isValid(u *State) bool {
	i, ok := d.openHash[u.Hash()]
	if !ok {
		return false
	}

	if !d.Close(d.keyHashCode(u), i) {
		return false
	}

	return true
}

func (d *Dsl) UpdateGoal(x, y int32) {
	addPoints := make(map[int32]Point)
	addCosts := make(map[int32]float64)

	//info map[int32]*CellInfo
	//cells map[int32]*State
	for h, i := range d.cellHash.info {
		if !d.Close(i.cost, C1) {
			s := d.cellHash.cells[h]
			addPoints[h] = Point{s.x, s.y}
			addCosts[h] = i.cost
		}
	}

	d.cellHash = *NewCellHash()
	d.openHash = make(map[int32]float64)
	d.path = make([]Point, 0, maxSteps)

	d.k_m = 0.0

	d.goal.x = x
	d.goal.y = y

	d.cellHash.Put(d.goal, &CellInfo{0, 0, C1})

	h := d.heuristic(d.start, d.goal)
	d.cellHash.Put(d.start, &CellInfo{h, h, C1})
	d.start = d.calculateKey(d.start)
	d.last = d.start

	for i, v := range addPoints {
		d.UpdateCell(v.X, v.Y, addCosts[i])
	}
}

func (d *Dsl) UpdateStart(x, y int32) {
	d.start.x = x
	d.start.y = y

	d.k_m += d.heuristic(d.last, d.start)

	d.start = d.calculateKey(d.start)
	d.last = d.start
}

func (d *Dsl) Replan() bool {
	d.path = make([]Point, 0, maxSteps)

	res := d.computeShortestPath()
	if res < 0 {
		return false
	}

	var cur *State = d.start
	if d.getG(d.start) == math.Inf(1) {
		return false
	}

	for cur.Neq(d.goal) {
		d.path = append(d.path, Point{cur.x, cur.y})

		n := d.getSucc(cur)
		if n.IsEmpty() {
			return false
		}

		var cmin float64 = math.Inf(1)
		var tmin float64 = 0
		var smin = State{0, 0, 0, 0}

		for _, i := range n.states {
			var val float64 = d.cost(cur, i)
			var val2 float64 = d.trueDist(i, d.goal) + d.trueDist(d.start, i)
			val += d.getG(i)

			iclose := d.Close(val, cmin)
			if iclose && tmin > val2 {
				tmin = val2
				cmin = val
				smin = State{i.x, i.y, i.k1, i.k2}
			} else if val < cmin {
				tmin = val2
				cmin = val
				smin = State{i.x, i.y, i.k1, i.k2}
			}
		}
		n.Clear()
		cur = &State{smin.x, smin.y, smin.k1, smin.k2}
	}

	d.path = append(d.path, Point{d.goal.x, d.goal.y})
	return true
}
