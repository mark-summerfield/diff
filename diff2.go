// Copyright © 2022-23 Mark Summerfield. All rights reserved.
// License: Apache-2.0

// Diff2 is a package for finding the differences between two sequences.
//
// Diff2 uses a sequence matcher based on a slightly simplified version of
// Python's difflib sequence matcher.
//
// Thanks to generics Diff2 can compare any two slices of comparables.
// (Although comparing float sequences is not recommended.) And using a key
// function it can also compare two slices of structs.
//
// See [New] for how to create a Diff value and [NewKeyFn] for how to
// create a DiffKeyFn value which can be used to compare sequences of
// structs.
package diff2

import (
	"cmp"
	_ "embed"
	"math"
	"slices"

	"github.com/mark-summerfield/set"
)

//go:embed Version.dat
var Version string

type b2jmap[T cmp.Ordered] map[T][]int

type Diff[T cmp.Ordered] struct {
	A   []T
	B   []T
	b2j b2jmap[T]
}

// New returns a Diff value based on the provided a and b slices. These
// slices are only ever read and may be accessed as .A and .B. After
// creating a Diff, call [Blocks] (or [Spans]) to see the differences.
func New[T cmp.Ordered](a, b []T) *Diff[T] {
	diff := &Diff[T]{A: a, B: b, b2j: b2jmap[T]{}}
	diff.chainBseq()
	return diff
}

func (me *Diff[T]) chainBseq() {
	for i, x := range me.B {
		indexes, ok := me.b2j[x]
		if !ok {
			indexes = []int{}
		}
		me.b2j[x] = append(indexes, i)
	}
	length := len(me.B)
	if length >= 200 { // remove most popular
		popular := set.New[T]()
		limit := 1 + int(math.Floor((float64(length) / 100.0)))
		for x, indexes := range me.b2j {
			if len(indexes) > limit {
				popular.Add(x)
			}
		}
		for x := range popular.All() {
			delete(me.b2j, x)
		}
	}
}

// Blocks returns a sequence of Block values representing how to go from a
// to b. Each block has a [Tag] and a sequence of A's and B's items.
// This is the easiest method for seeing the differences in two sequences.
// See also [Spans].
func (me *Diff[T]) Blocks() []Block[T] {
	matches := me.matches()
	blocks := []Block[T]{}
	for _, span := range spansForMatches(matches) {
		var aitems, bitems []T
		if span.Aend <= len(me.A) {
			aitems = me.A[span.Astart:span.Aend]
		}
		if span.Bend <= len(me.B) {
			bitems = me.B[span.Bstart:span.Bend]
		}
		blocks = append(blocks, newBlock(span.Tag, aitems, bitems))
	}
	return blocks
}

// Spans returns a sequence of Span values representing how to go from a
// to b. Each span has a [Tag] and a sequence of [Quad]s. Each [Quad] holds
// a pair of start/end indexes into a and b.
// The easiest method for seeing the differences in two sequences is
// [Blocks].
func (me *Diff[T]) Spans() []Span {
	matches := me.matches()
	return spansForMatches(matches)
}

func (me *Diff[T]) matches() []match {
	aLength := len(me.A)
	bLength := len(me.B)
	queue := []Quad{newQuad(0, aLength, 0, bLength)}
	matches := []match{}
	for len(queue) > 0 {
		end := len(queue) - 1
		quad := queue[end]
		queue = queue[:end]
		match := me.longestMatch(quad)
		i := match.astart
		j := match.bstart
		k := match.length
		if k > 0 {
			matches = append(matches, match)
			if quad.Astart < i && quad.Bstart < j {
				queue = append(queue,
					newQuad(quad.Astart, i, quad.Bstart, j))
			}
			if i+k < quad.Aend && j+k < quad.Bend {
				queue = append(queue,
					newQuad(i+k, quad.Aend, j+k, quad.Bend))
			}
		}
	}
	slices.SortFunc(matches, matchCompare)
	aStart := 0
	bStart := 0
	length := 0
	nonAdjacent := []match{}
	for _, match := range matches {
		if aStart+length == match.astart && bStart+length == match.bstart {
			length += match.length
		} else {
			if length != 0 {
				nonAdjacent = append(nonAdjacent,
					newMatch(aStart, bStart, length))
			}
			aStart = match.astart
			bStart = match.bstart
			length = match.length
		}
	}
	if length != 0 {
		nonAdjacent = append(nonAdjacent, newMatch(aStart, bStart, length))
	}
	nonAdjacent = append(nonAdjacent, newMatch(aLength, bLength, 0))
	return nonAdjacent
}

func (me *Diff[T]) longestMatch(quad Quad) match {
	aStart := quad.Astart
	aEnd := quad.Aend
	bStart := quad.Bstart
	bEnd := quad.Bend
	bestI := aStart
	bestJ := bStart
	bestSize := 0
	j2len := map[int]int{}
	for i := aStart; i < aEnd; i++ {
		newJ2len := map[int]int{}
		indexes, ok := me.b2j[me.A[i]]
		if ok {
			for _, j := range indexes {
				if j < bStart {
					continue
				}
				if j >= bEnd {
					break
				}
				k := j2len[j-1]
				k++
				newJ2len[j] = k
				if k > bestSize {
					bestI = i - k + 1
					bestJ = j - k + 1
					bestSize = k
				}
			}
		}
		j2len = newJ2len
	}
	for bestI > aStart && bestJ > bStart &&
		me.A[bestI-1] == me.B[bestJ-1] {
		bestI--
		bestJ--
		bestSize++
	}
	for bestI+bestSize < aEnd && bestJ+bestSize < bEnd &&
		me.A[bestI+bestSize] == me.B[bestJ+bestSize] {
		bestSize++
	}
	return newMatch(bestI, bestJ, bestSize)
}

func spansForMatches(matches []match) []Span {
	spans := []Span{}
	i := 0
	j := 0
	for _, match := range matches {
		var tag Tag
		if i < match.astart {
			if j < match.bstart {
				tag = Replace
			} else {
				tag = Delete
			}
		} else if j < match.bstart {
			tag = Insert
		}
		if tag != Equal {
			spans = append(spans,
				newSpan(tag, i, match.astart, j, match.bstart))
		}
		i = match.astart + match.length
		j = match.bstart + match.length
		if match.length != 0 {
			spans = append(spans,
				newSpan(Equal, match.astart, i, match.bstart, j))
		}
	}
	return spans
}
