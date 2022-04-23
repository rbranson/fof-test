package fof

import (
	"fmt"
	"sort"
)

type Graph struct {
	Index map[int]*EdgeSet
}

func NewGraph() Graph {
	return Graph{
		Index: make(map[int]*EdgeSet),
	}
}

func (g *Graph) GetOrCreate(id int) *EdgeSet {
	es, ok := g.Index[id]
	if !ok {
		es = &EdgeSet{}
		g.Index[id] = es
	}
	return es
}

func (g *Graph) Add(a, b int) bool {
	if a == b {
		return false
	}

	aes := g.GetOrCreate(a)
	bes := g.GetOrCreate(b)

	aok := aes.Add(b)
	bok := bes.Add(a)

	if aok != bok {
		panic("graph inconsistency")
	}

	return aok
}

func addWeights(ids []int, weights map[int]int) {
	for _, id := range ids {
		weight, _ := weights[id]
		weights[id] = weight + 1
	}
}

func (g *Graph) Mutual(from int) (ret EdgeSet, weights map[int]int) {
	fromEs, ok := g.Index[from]
	if !ok {
		return
	}
	weights = make(map[int]int)
	for _, toId := range fromEs.MutableIds() {
		toEs := g.Index[toId]
		matchedEs := fromEs.Intersection(*toEs)
		addWeights(matchedEs.MutableIds(), weights)
		// MergeReplace is almost always faster until we get a b-tree in herr
		ret.MergeReplace(matchedEs)
	}
	return
}

type EdgeSet struct {
	edges []int
}

func intSliceInsert(s []int, v int, idx int) []int {
	if idx > len(s) {
		panic(fmt.Sprintf("idx > %v", len(s)))
	}
	if idx == 0 {
		preamble := []int{v}
		if len(s) == 0 {
			return preamble
		}
		return append(preamble, s...)
	}
	if idx == len(s) {
		return append(s, v)
	}
	ret := append(s[:idx+1], s[idx:]...)
	ret[idx] = v
	return ret
}

func (s *EdgeSet) Add(a int) (ok bool) {
	var idx int
	if len(s.edges) == 0 {
		idx = 0
	} else if s.edges[len(s.edges)-1] < a {
		// optimize add highest case
		idx = len(s.edges)
	} else {
		idx, ok = s.index(a)
		if ok {
			return false
		}
	}
	s.edges = intSliceInsert(s.edges, a, idx)
	return true
}

func (s *EdgeSet) Exists(a int) bool {
	_, ok := s.index(a)
	return ok
}

func (s *EdgeSet) index(a int) (int, bool) {
	idx := sort.SearchInts(s.edges, a)
	if idx < len(s.edges) && s.edges[idx] == a {
		return -1, true
	}
	return idx, false
}

func (s *EdgeSet) Intersection(o EdgeSet) EdgeSet {
	si, oi := 0, 0
	var outSet []int

	for si < len(s.edges) && oi < len(o.edges) {
		if s.edges[si] == o.edges[oi] {
			outSet = append(outSet, s.edges[si])
			si++
			oi++
		} else if s.edges[si] < o.edges[oi] {
			si++
		} else if s.edges[si] > o.edges[oi] {
			oi++
		} else {
			panic("inconceivable")
		}
	}

	return EdgeSet{edges: outSet}
}

func (s *EdgeSet) Replace(ids []int) {
	if len(ids) == 0 {
		s.edges = nil
		return
	}

	edges := ids[:]
	sort.Ints(edges)
	s.edges = edges
}

func (s EdgeSet) Len() int {
	return len(s.edges)
}

func (s *EdgeSet) MutableIds() []int {
	return s.edges
}

func (s *EdgeSet) Merge(o EdgeSet) {
	for _, id := range o.MutableIds() {
		s.Add(id)
	}
}

func (s *EdgeSet) MergeReplace(o EdgeSet) {
	si, oi := 0, 0
	// 5 is magic, find common cases to be long
	outSet := make([]int, maxInt(s.Len(), o.Len())+5)
	outIdx := 0

	for si < len(s.edges) && oi < len(o.edges) {
		if outIdx >= len(outSet) {
			extendo := make([]int, maxInt(s.Len()-si, o.Len()-oi))
			outSet = append(outSet, extendo...)
		}
		if s.edges[si] == o.edges[oi] {
			outSet[outIdx] = s.edges[si]
			outIdx++
			si++
			oi++
		} else if s.edges[si] < o.edges[oi] {
			outSet[outIdx] = s.edges[si]
			outIdx++
			si++
		} else if s.edges[si] > o.edges[oi] {
			outSet[outIdx] = o.edges[oi]
			outIdx++
			oi++
		} else {
			panic("inconceivable")
		}
	}

	for si < len(s.edges) {
		if outIdx >= len(outSet) {
			extendo := make([]int, len(s.edges)-si)
			outSet = append(outSet, extendo...)
		}
		outSet[outIdx] = s.edges[si]
		outIdx++
		si++
	}

	for oi < len(o.edges) {
		if outIdx >= len(outSet) {
			extendo := make([]int, len(o.edges)-oi)
			outSet = append(outSet, extendo...)
		}
		outSet[outIdx] = o.edges[oi]
		outIdx++
		oi++
	}

	s.edges = outSet[:outIdx]
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
