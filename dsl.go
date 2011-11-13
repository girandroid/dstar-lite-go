package main

import (
	"math"
)

var (
	maxSteps uint32  = 999000 // node extensions before we give up
	C1       float64 = 1.0  // cost of unseen cell
	eps      float64 = 0.00001
	M_SQRT2  float64 = math.Sqrt(2.0)
)

type OpenList struct {
	queue []*State
	size  int
}

func NewOpenList() *OpenList {
	ol := new(OpenList)
	ol.queue = make([]*State, 64, 64)
	ol.size = 0
	return ol
}

func (ol *OpenList) IsEmpty() bool {
	return ol.size == 0
}

func (ol *OpenList) Add(s *State) bool {
	i := ol.size
	if i >= len(ol.queue) {
		ol.grow(i + 1)
	}
	ol.size = i + 1
	if i == 0 {
		ol.queue[0] = s
	} else {
		ol.siftUp(i, s)
	}
	return true
}

func (ol *OpenList) grow(minC int) {
	oldC := len(ol.queue)
	var newC int
	if oldC < 64 {
		newC = (oldC + 1) * 2
	} else {
		newC = (oldC / 2) * 3 
	}

	if newC < minC {
		newC = minC
	}
	newQueue := make([]*State, newC, newC)
	copy(newQueue, ol.queue)
	ol.queue = newQueue
}

func (ol *OpenList) Peek() *State {
	if len(ol.queue) == 0 {
		return nil
	}
	return ol.queue[0]
}

func (ol *OpenList) Poll() *State {
	if ol.size == 0 {
		return nil
	}
	ol.size--
	s := ol.size
	result := ol.queue[0]
	x := ol.queue[s]
	ol.queue[s] = nil
	if s != 0 {
		ol.siftDown(0, x)
	}
	return result
}

func (ol *OpenList) siftDown(k int, x *State) {
	half := ol.size / 2
	for k < half {
		child := (k * 2) + 1
		c := ol.queue[child]
		right := child + 1
		if right < ol.size && c.Gt(ol.queue[right]) {
			child = right
			c = ol.queue[child]
		}
		if x.Lte(c) {
			break
		}

		ol.queue[k] = c
		k = child
	}
	ol.queue[k] = x
}

func (ol *OpenList) siftUp(k int, x *State) {
	for k > 0 {
		parent := (k - 1) / 2
		e := ol.queue[parent]
		if !x.Lt(e) {
			break
		}
		ol.queue[k] = e
		k = parent
	}
	ol.queue[k] = x
}

func (ol *OpenList) Clear() {
	ol.queue = make([]*State, len(ol.queue), cap(ol.queue))
	ol.size = 0
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
	} else if s.k1+eps < s2.k1 {
		return false
	}
	return s.k2 > s2.k2
}

func (s *State) Gte(s2 *State) bool {
	if s.k1 > s2.k1 {
		return true
	} else if s.k1 < s2.k1 {
		return false
	}
	return s.k2 > s2.k2-eps
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

	openList OpenList
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
	if u.Neq(d.goal) {
		s := d.getSucc(u)

		var tmp float64 = math.Inf(1)
		for _, i := range s {
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

func (d *Dsl) getSucc(u *State) []*State {
	ns := make([]*State, 0, 8)
	if d.occupied(u) {
		return ns
	}
	ns = append(ns, &State{u.x + 1, u.y - 1, -1, -1})
	ns = append(ns, &State{u.x, u.y - 1, -1, -1})
	ns = append(ns, &State{u.x - 1, u.y - 1, -1, -1})
	ns = append(ns, &State{u.x - 1, u.y, -1, -1})
	ns = append(ns, &State{u.x - 1, u.y + 1, -1, -1})
	ns = append(ns, &State{u.x, u.y + 1, -1, -1})
	ns = append(ns, &State{u.x + 1, u.y + 1, -1, -1})
	ns = append(ns, &State{u.x + 1, u.y, -1, -1})

	return ns
}

func (d *Dsl) getPred(u *State) []*State {
	ns := make([]*State, 0, 8)

	tempState := &State{u.x + 1, u.y - 1, -1, -1}
	if !d.occupied(tempState) {
		ns = append(ns, tempState)
	}
	tempState = &State{u.x, u.y - 1, -1, -1}
	if !d.occupied(tempState) {
		ns = append(ns, tempState)
	}
	tempState = &State{u.x - 1, u.y - 1, -1, -1}
	if !d.occupied(tempState) {
		ns = append(ns, tempState)
	}
	tempState = &State{u.x - 1, u.y, -1, -1}
	if !d.occupied(tempState) {
		ns = append(ns, tempState)
	}
	tempState = &State{u.x - 1, u.y + 1, -1, -1}
	if !d.occupied(tempState) {
		ns = append(ns, tempState)
	}
	tempState = &State{u.x, u.y + 1, -1, -1}
	if !d.occupied(tempState) {
		ns = append(ns, tempState)
	}
	tempState = &State{u.x + 1, u.y, -1, -1}
	if !d.occupied(tempState) {
		ns = append(ns, tempState)
	}
	tempState = &State{u.x + 1, u.y + 1, -1, -1}
	if !d.occupied(tempState) {
		ns = append(ns, tempState)
	}

	return ns
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
		// lazy remove
		for {
			if d.openList.IsEmpty() {
				return 1
			}
			u = d.openList.Poll()
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
			for _, i := range s {
				d.updateVertex(i)
			}
		} else {
			d.setG(u, math.Inf(1))
			s := d.getPred(u)
			for _, i := range s {
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
	d.openList.Clear()

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
		if len(n) == 0 {
			return false
		}

		var cmin float64 = math.Inf(1)
		var tmin float64 = 0
		var smin = State{0, 0, 0, 0}
		for _, i := range n {
			var val float64 = d.cost(cur, i)
			var val2 float64 = d.trueDist(i, d.goal) + d.trueDist(d.start, i)
			val += d.getG(i)

			iclose := d.Close(val, cmin)
			if iclose && tmin > val2 {
				tmin = val2
				cmin = val
				smin = *i
			} else if val < cmin {
				tmin = val2
				cmin = val
				smin = *i
			}
		}
		cur = &State{smin.x, smin.y, smin.k1, smin.k2}
	}

	d.path = append(d.path, Point{d.goal.x, d.goal.y})
	return true
}
