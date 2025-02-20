// Copyright © 2022-25 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package diff

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/mark-summerfield/utext"
)

func testStrings(a, b string, expected []string, t *testing.T) {
	x := strings.Fields(a)
	y := strings.Fields(b)
	diff := New(x, y)
	actual := []string{}
	for _, block := range diff.Blocks() {
		actual = append(actual, fmt.Sprintf("%s %s", block.Tag,
			strings.Join(block.Items(), " ")))
	}
	for i, line := range actual {
		if line != expected[i] {
			t.Errorf("expected %q, got %q", expected[i], line)
		}
	}
}

func ExampleDiff_Blocks_common() {
	a := strings.Fields("foo\nbar\nbaz\nquux")
	b := strings.Fields("foo\nbaz\nbar\nquux")
	diff := New(a, b)
	blocks := diff.Blocks()
	for _, block := range blocks {
		fmt.Println(block.Tag, strings.Join(block.Items(), " "))
	}
	// Output:
	// = foo
	// + baz
	// = bar
	// - baz
	// = quux
}

func ExampleDiff_Blocks_strings() {
	a := strings.Fields("the quick brown fox jumped over the lazy dogs")
	b := strings.Fields("a quick red fox jumped over some lazy hogs")
	diff := New(a, b)
	blocks := diff.Blocks()
	for _, block := range blocks {
		fmt.Println(block.Tag, strings.Join(block.Items(), " "))
	}
	// Output:
	// % a
	// = quick
	// % red
	// = fox jumped over
	// % some
	// = lazy
	// % hogs
}

func ExampleDiff_Blocks_changesonly() {
	a := strings.Fields("the quick brown fox jumped over the lazy dogs")
	b := strings.Fields("a quick red fox jumped over some lazy hogs")
	diff := New(a, b)
	blocks := diff.Blocks()
	for _, block := range blocks {
		if block.Tag != Equal {
			fmt.Println(block.Tag, strings.Join(block.Items(), " "))
		}
	}
	// Output:
	// % a
	// % red
	// % some
	// % hogs
}

func ExampleDiff_Blocks_changesnoreplace() {
	a := strings.Fields("the quick brown fox jumped over the lazy dogs")
	b := strings.Fields("a quick red fox jumped over some lazy hogs")
	diff := New(a, b)
	blocks := diff.Blocks()
	for _, block := range blocks {
		if block.Tag != Equal {
			if block.Tag == Replace {
				fmt.Println(Delete, strings.Join(block.Aitems, ""))
				fmt.Println(Insert, strings.Join(block.Bitems, ""))
			} else {
				fmt.Println(block.Tag, strings.Join(block.Items(), " "))
			}
		}
	}
	// Output:
	// - the
	// + a
	// - brown
	// + red
	// - the
	// + some
	// - dogs
	// + hogs
}

func ExampleDiff_Blocks_ints() {
	a := []int{1, 2, 3, 4, 5, 6}
	b := []int{2, 3, 5, 7}
	diff := New(a, b)
	blocks := diff.Blocks()
	for _, block := range blocks {
		fmt.Println(block.Tag, utext.StringForSlice(block.Items()))
	}
	// Output:
	// - 1
	// = 2 3
	// - 4
	// = 5
	// % 7
}

type Place struct {
	x    int
	y    int
	name string
}

func newPlace(x, y int, name string) Place {
	return Place{x: x, y: y, name: name}
}

func (me *Place) String() string {
	return fmt.Sprintf("Place{%d,%d,%q}", me.x, me.y, me.name)
}

func ExampleDiffKeyFn_Blocks_namekey() {
	/*
		type Place struct {
		x    int
		y    int
		name string
		}

		func newPlace(x, y int, name string) Place {
		return Place{x: x, y: y, name: name}
		}

		func (me *Place) String() string {
		return fmt.Sprintf("Place{%d,%d,%q}", me.x, me.y, me.name)
		}
	*/
	a := []Place{newPlace(1, 2, "foo"), newPlace(3, 4, "bar"),
		newPlace(5, 6, "baz"), newPlace(7, 8, "quux")}
	b := []Place{newPlace(1, 2, "foo"), newPlace(6, 2, "baz"),
		newPlace(3, 4, "bar"), newPlace(7, 8, "quux")}
	diff := NewKeyFn(a, b, func(p Place) string { return p.name })
	blocks := diff.Blocks()
	for _, block := range blocks {
		for _, item := range block.Items() {
			fmt.Println(block.Tag, item.String())
		}
	}
	// Output:
	// = Place{1,2,"foo"}
	// + Place{6,2,"baz"}
	// = Place{3,4,"bar"}
	// - Place{5,6,"baz"}
	// = Place{7,8,"quux"}
}

func ExampleDiffKeyFn_Blocks_xkey() {
	// a: [1 3 5 7]
	// b: [1 6 3 7]
	a := []Place{newPlace(1, 2, "foo"), newPlace(3, 4, "bar"),
		newPlace(5, 6, "baz"), newPlace(7, 8, "quux")}
	b := []Place{newPlace(1, 2, "foo"), newPlace(6, 2, "bar"),
		newPlace(3, 4, "baz"), newPlace(7, 8, "quux")}
	diff := NewKeyFn(a, b,
		func(p Place) string { return strconv.Itoa(p.x) })
	blocks := diff.Blocks()
	for _, block := range blocks {
		for _, item := range block.Items() {
			fmt.Println(block.Tag, item.String())
		}
	}
	// Output:
	// = Place{1,2,"foo"}
	// + Place{6,2,"bar"}
	// = Place{3,4,"baz"}
	// - Place{5,6,"baz"}
	// = Place{7,8,"quux"}
}

func Test001(t *testing.T) {
	expected := []string{"= foo", "+ baz", "= bar", "- baz", "= quux"}
	testStrings("foo\nbar\nbaz\nquux", "foo\nbaz\nbar\nquux", expected, t)
}

func Test002(t *testing.T) {
	a := strings.Split("foo\nbar\nbaz\nquux", "\n")
	b := strings.Split("foo\nbaz\nbar\nquux", "\n")
	diff := New(a, b)
	actual := []string{}
	for _, span := range diff.Spans() {
		switch span.Tag {
		case Equal:
			actual = append(actual, "= "+
				strings.Join(a[span.Astart:span.Aend], "|"))
		case Insert:
			actual = append(actual, "+ "+
				strings.Join(b[span.Bstart:span.Bend], "|"))
		case Delete:
			actual = append(actual, "- "+
				strings.Join(a[span.Astart:span.Aend], "|"))
		case Replace:
			actual = append(actual, "% "+
				strings.Join(b[span.Bstart:span.Bend], "|"))
		}
	}
	expected := []string{"= foo", "+ baz", "= bar", "- baz", "= quux"}
	for i, line := range actual {
		if line != expected[i] {
			t.Errorf("expected %q, got %q", expected[i], line)
		}
	}
}

func Test003(t *testing.T) {
	expected := []string{"= the quick", "% red", "= fox jumped over the",
		"% very busy", "= dogs"}
	testStrings(
		"the quick brown fox jumped over the lazy dogs",
		"the quick red fox jumped over the very busy dogs",
		expected, t)
}

func Test004(t *testing.T) {
	expected := []string{"- q", "= a b", "% y", "= c d", "+ f"}
	a := "q a b x c d"
	b := "a b y c d f"
	testStrings(a, b, expected, t)
}

func Test005(t *testing.T) {
	expected := []string{"= private", "+ volatile",
		"= Thread currentThread;"}
	a := "private Thread currentThread;"
	b := "private volatile Thread currentThread;"
	testStrings(a, b, expected, t)
}

func Test006(t *testing.T) {
	expected := []Block[int]{
		newBlock(Delete, []int{1}, []int{}),
		newBlock(Equal, []int{}, []int{2, 3}),
		newBlock(Delete, []int{4}, []int{}),
		newBlock(Equal, []int{}, []int{5}),
		newBlock(Replace, []int{}, []int{7}),
	}
	a := []int{1, 2, 3, 4, 5, 6}
	b := []int{2, 3, 5, 7}
	diff := New(a, b)
	actual := diff.Blocks()
	for i, block := range actual {
		if !isEqualBlock(block, expected[i]) {
			t.Errorf("expected %s[%s], got %s[%s]", expected[i].Tag,
				utext.StringForSlice(expected[i].Items()), block.Tag,
				utext.StringForSlice(block.Items()))
		}
	}
}

func Test007(t *testing.T) {
	expected := []Block[rune]{
		newBlock(Delete, []rune{'q'}, []rune{}),
		newBlock(Equal, []rune{}, []rune{'a', 'b'}),
		newBlock(Replace, []rune{}, []rune{'y'}),
		newBlock(Equal, []rune{}, []rune{'c', 'd'}),
		newBlock(Insert, []rune{}, []rune{'f'}),
	}
	a := []rune("qabxcd")
	b := []rune("abycdf")
	diff := New(a, b)
	actual := diff.Blocks()
	for i, block := range actual {
		if !isEqualBlock(block, expected[i]) {
			t.Errorf("expected %s[%s], got %s[%s]", expected[i].Tag,
				utext.StringForSlice(expected[i].Items()), block.Tag,
				utext.StringForSlice(block.Items()))
		}
	}
}

func Test008(t *testing.T) {
	expected := []string{"- the quick brown fox jumped over the lazy dogs"}
	a := "the quick brown fox jumped over the lazy dogs"
	b := ""
	testStrings(a, b, expected, t)
}

func Test009(t *testing.T) {
	expected := []string{"+ the quick brown fox jumped over the lazy dogs"}
	a := ""
	b := "the quick brown fox jumped over the lazy dogs"
	testStrings(a, b, expected, t)
}

func Test010(t *testing.T) {
	expected := []Block[string]{
		newBlock(Delete, []string{"1"}, []string{}),
		newBlock(Equal, []string{}, []string{"2", "3"}),
		newBlock(Delete, []string{"4"}, []string{}),
		newBlock(Equal, []string{}, []string{"5"}),
		newBlock(Replace, []string{}, []string{"7"}),
	}
	a := []string{"1", "2", "3", "4", "5", "6"}
	b := []string{"2", "3", "5", "7"}
	diff := New(a, b)
	actual := diff.Blocks()
	for i, block := range actual {
		if !isEqualBlock(block, expected[i]) {
			t.Errorf("expected %s[%s], got %s[%s]", expected[i].Tag,
				utext.StringForSlice(expected[i].Items()), block.Tag,
				utext.StringForSlice(block.Items()))
		}
	}
}

func Test011(t *testing.T) {
	expected := []Span{
		newSpan(Delete, 0, 1, 0, 0),
		newSpan(Equal, 1, 3, 0, 2),
		newSpan(Delete, 3, 4, 2, 2),
		newSpan(Equal, 4, 5, 2, 3),
		newSpan(Replace, 5, 6, 3, 4),
	}
	a := []int{1, 2, 3, 4, 5, 6}
	b := []int{2, 3, 5, 7}
	diff := New(a, b)
	actual := diff.Spans()
	for i, span := range actual {
		if !isEqualSpan(span, expected[i]) {
			t.Errorf("expected %q, got %q", expected[i], span)
		}
	}
}

func Test012(t *testing.T) {
	expectedMatch := newMatch(1, 0, 2)
	expected := []Block[int]{
		newBlock(Delete, []int{1}, []int{}),
		newBlock(Equal, []int{}, []int{2, 3}),
		newBlock(Delete, []int{4}, []int{}),
		newBlock(Equal, []int{}, []int{5}),
		newBlock(Replace, []int{}, []int{7}),
	}
	a := []int{1, 2, 3, 4, 5, 6}
	b := []int{2, 3, 5, 7}
	diff := New(a, b)
	match := diff.longestMatch(newQuad(0, len(a), 0, len(b)))
	if match != expectedMatch {
		t.Errorf("expected %q, got %q", expectedMatch, match)
	}
	actual := diff.Blocks()
	for i, block := range actual {
		if !isEqualBlock(block, expected[i]) {
			t.Errorf("expected %s[%s], got %s[%s]", expected[i].Tag,
				utext.StringForSlice(expected[i].Items()), block.Tag,
				utext.StringForSlice(block.Items()))
		}
	}
}

func isEqualSpan(a, b Span) bool {
	if a.Tag != b.Tag {
		return false
	}
	if a.Astart != b.Astart {
		return false
	}
	if a.Aend != b.Aend {
		return false
	}
	if a.Bstart != b.Bstart {
		return false
	}
	if a.Bend != b.Bend {
		return false
	}
	return true
}

func isEqualBlock[T comparable](a, b Block[T]) bool {
	if a.Tag != b.Tag {
		return false
	}
	if len(a.Items()) != len(b.Items()) {
		return false
	}
	aitems := a.Items()
	bitems := b.Items()
	for i := 0; i < len(aitems); i++ {
		if aitems[i] != bitems[i] {
			return false
		}
	}
	return true
}
