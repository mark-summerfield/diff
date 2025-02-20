// Copyright © 2022-25 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package diff

import "slices"

type KeyFn[T any] func(x T) string

type b2jmapkeyfn map[string][]int

type DiffKeyFn[T any] struct {
	A     []T
	B     []T
	b2j   b2jmapkeyfn
	keyfn KeyFn[T]
}

// NewKeyFn returns a DiffKeyFn value based on the provided a and b
// slices. These slices are only ever read and may be accessed as .A and .B.
// After creating a DiffKeyFn, call [Blocks] (or [Spans]) to see the
// differences.
func NewKeyFn[T any](a, b []T, keyfn KeyFn[T]) *DiffKeyFn[T] {
	diff := &DiffKeyFn[T]{A: a, B: b, b2j: b2jmapkeyfn{}, keyfn: keyfn}
	diff.chainBseq()
	return diff
}

func (me *DiffKeyFn[T]) chainBseq() {
	for i, x := range me.B {
		key := me.keyfn(x)
		indexes, ok := me.b2j[key]
		if !ok {
			indexes = []int{}
		}
		me.b2j[key] = append(indexes, i)
	}
}

// Blocks returns a sequence of BlockKeyFn values representing how to go
// from a to b. Each block has a [Tag] and a sequence of A's and B's items.
// This is the easiest method for seeing the differences in two sequences.
// See also [Spans].
func (me *DiffKeyFn[T]) Blocks() []BlockKeyFn[T] {
	matches := me.matches()
	blocks := []BlockKeyFn[T]{}
	for _, span := range spansForMatches(matches) {
		var aitems, bitems []T
		if span.Aend <= len(me.A) {
			aitems = me.A[span.Astart:span.Aend]
		}
		if span.Bend <= len(me.B) {
			bitems = me.B[span.Bstart:span.Bend]
		}
		blocks = append(blocks, newBlockKeyFn(span.Tag, aitems, bitems))
	}
	return blocks
}

// Spans returns a sequence of Span values representing how to go from a
// to b. Each span has a [Tag] and a sequence of [Quad]s. Each [Quad] holds
// a pair of start/end indexes into a and b.
// The easiest method for seeing the differences in two sequences is
// [Blocks].
func (me *DiffKeyFn[T]) Spans() []Span {
	matches := me.matches()
	return spansForMatches(matches)
}

func (me *DiffKeyFn[T]) matches() []match {
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

func (me *DiffKeyFn[T]) longestMatch(quad Quad) match {
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
		indexes, ok := me.b2j[me.keyfn(me.A[i])]
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
		me.keyfn(me.A[bestI-1]) == me.keyfn(me.B[bestJ-1]) {
		bestI--
		bestJ--
		bestSize++
	}
	for bestI+bestSize < aEnd && bestJ+bestSize < bEnd &&
		me.keyfn(me.A[bestI+bestSize]) == me.keyfn(me.B[bestJ+bestSize]) {
		bestSize++
	}
	return newMatch(bestI, bestJ, bestSize)
}
