// Copyright © 2022-23 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package diff2

import "fmt"

type Tag uint8

const (
	Equal Tag = iota
	Insert
	Delete
	Replace
)

func (me Tag) String() string {
	switch me {
	case Equal:
		return "="
	case Insert:
		return "+"
	case Delete:
		return "-"
	case Replace:
		return "%"
	}
	panic("invalid tag")
}

type Block[T comparable] struct {
	Tag    Tag
	Aitems []T
	Bitems []T
}

func newBlock[T comparable](tag Tag, aitems, bitems []T) Block[T] {
	return Block[T]{Tag: tag, Aitems: aitems, Bitems: bitems}
}

func (me *Block[T]) Items() []T {
	if me.Tag == Delete {
		return me.Aitems
	}
	return me.Bitems
}

type BlockKeyFn[T any] struct {
	Tag    Tag
	Aitems []T
	Bitems []T
}

func newBlockKeyFn[T any](tag Tag, aitems, bitems []T) BlockKeyFn[T] {
	return BlockKeyFn[T]{Tag: tag, Aitems: aitems, Bitems: bitems}
}

func (me *BlockKeyFn[T]) Items() []T {
	if me.Tag == Delete {
		return me.Aitems
	}
	return me.Bitems
}

type match struct {
	astart int
	bstart int
	length int
}

func newMatch(astart, bstart, length int) match {
	return match{astart: astart, bstart: bstart, length: length}
}

func matchLess(a, b match) bool {
	if a.astart != b.astart {
		return a.astart < b.astart
	}
	if a.bstart != b.bstart {
		return a.bstart < b.bstart
	}
	return a.length < b.length
}

type Quad struct {
	Astart int
	Aend   int
	Bstart int
	Bend   int
}

func newQuad(astart, aend, bstart, bend int) Quad {
	return Quad{Astart: astart, Aend: aend, Bstart: bstart, Bend: bend}
}

type Span struct {
	Tag Tag
	Quad
}

func newSpan(tag Tag, astart, aend, bstart, bend int) Span {
	return Span{Tag: tag, Quad: newQuad(astart, aend, bstart, bend)}
}

func (me Span) String() string {
	return fmt.Sprintf("<%s [%d:%d] [%d:%d]>", me.Tag, me.Astart, me.Aend,
		me.Bstart, me.Bend)
}
