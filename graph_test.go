package fof

import (
	"bufio"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEdgeSet(t *testing.T) {
	es1 := EdgeSet{}

	require.Equal(t, 0, es1.Len())

	require.True(t, es1.Add(1))
	require.False(t, es1.Add(1))

	require.Equal(t, 1, es1.Len())

	es1.Add(10)
	es1.Add(7)
	es1.Add(9)
	es1.Add(4)

	require.Equal(t, 5, es1.Len())

	require.True(t, es1.Exists(10))
	require.False(t, es1.Exists(11))
	require.True(t, es1.Exists(9))

	es2 := EdgeSet{}
	es2.Replace([]int{4, 6, 10})
	require.Equal(t, 3, es2.Len())

	require.True(t, es2.Exists(4))
	require.True(t, es2.Exists(6))
	require.True(t, es2.Exists(10))
	require.False(t, es2.Exists(9))

	es3 := es1.Intersection(es2)

	require.Equal(t, 2, es3.Len())
	require.True(t, es3.Exists(4))
	require.False(t, es3.Exists(6))
	require.True(t, es3.Exists(10))
}

func TestEdgeSetMerge(t *testing.T) {
	var es1, es2, es3 EdgeSet
	es1.Replace([]int{100, 200, 300, 400})
	es2.Replace([]int{400, 500, 600, 700})
	es3.Replace([]int{100, 200, 300, 400, 500, 600, 700})

	es1.Merge(es2)
	require.Equal(t, 7, es1.Len())
	require.Equal(t, 7, es1.Intersection(es3).Len())
}

func TestEdgeSetMergeReplace(t *testing.T) {
	var es1, es2, es3 EdgeSet
	es1.Replace([]int{100, 200, 300, 400})
	es2.Replace([]int{400, 500, 600, 700})
	es3.Replace([]int{100, 200, 300, 400, 500, 600, 700})

	es1.Merge(es2)
	require.Equal(t, 7, es1.Len())
	require.Equal(t, 7, es1.Intersection(es3).Len())

	var es4, es5, es6 EdgeSet
	es4.Replace(buildSeqSet(10000, 1000, 100))
	es5.Replace(buildSeqSet(0, 1000, 200))

	es6.Replace(es4.MutableIds())
	es6.MergeReplace(es5)

	es7 := es6.Intersection(es4)
	require.Equal(t, es4.Len(), es7.Len())
	es8 := es6.Intersection(es5)
	require.Equal(t, es5.Len(), es8.Len())

	t.Logf("es4=%v\nes5=%v\nes6=%v\n", es4, es5, es6)
}

func buildSeqSet(start int, count int, incr int) []int {
	out := make([]int, count)
	for i := 0; i < count; i++ {
		out[i] = start + (i * incr)
	}
	return out
}

func BenchmarkEdgeSetMerge(b *testing.B) {
	sizes := []int{1, 10, 100, 1000, 10000}
	for _, sizeLeft := range sizes {
		leftSet := buildSeqSet(0, sizeLeft, 2)
		for _, sizeRight := range sizes {
			rightSet := buildSeqSet(sizeLeft, sizeRight, 1)
			b.Run(fmt.Sprintf("L%d_R%d_Merge", sizeLeft, sizeRight), func(b *testing.B) {
				var left, right EdgeSet
				right.Replace(rightSet)
				for i := 0; i < b.N; i++ {
					left.Replace(leftSet)
					left.Merge(right)
				}
			})
			b.Run(fmt.Sprintf("L%d_R%d_MergeReplace", sizeLeft, sizeRight), func(b *testing.B) {
				var left, right EdgeSet
				right.Replace(rightSet)
				for i := 0; i < b.N; i++ {
					left.Replace(leftSet)
					left.MergeReplace(right)
				}
			})
		}
	}
}

func BenchmarkEdgeSetAdd(b *testing.B) {
	sizes := []int{1, 10, 100, 1000, 10000}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("%d", size), func(b *testing.B) {
			var es EdgeSet
			ss := buildSeqSet(0, size, 1)
			for i := 0; i < b.N; i++ {
				es = EdgeSet{edges: ss}
				es.Add(size + 1)
			}
		})
	}
}

func getFacebookGraph() Graph {
	file, err := os.Open("facebook_combined.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	g := NewGraph()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var a, b int
		cnt, err := fmt.Sscanf(line, "%d %d", &a, &b)
		if cnt < 2 {
			panic("got < 2")
		}
		if err != nil {
			panic(err)
		}
		g.Add(a, b)
	}

	return g
}

func TestGraphMutual(t *testing.T) {
	g := getFacebookGraph()
	mutuals, weights := g.Mutual(3980)
	t.Logf("%v", mutuals.edges)
	t.Logf("%v", weights)
}

func BenchmarkGraphMutual(b *testing.B) {
	g := getFacebookGraph()
	ids := []int{}
	for id, _ := range g.Index {
		ids = append(ids, id)
	}
	for i := 0; i < b.N; i++ {
		startAt := ids[i%len(ids)]
		_, _ = g.Mutual(startAt)
	}
}
