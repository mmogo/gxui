// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gxui

import test "github.com/mmogo/gxui/testing"
import (
	"testing"

	"github.com/mmogo/gxui/interval"
)

func TestTextSelectionMergeOne(t *testing.T) {
	s := CreateTextSelection(5, 10, true)
	l := TextSelectionList{}
	interval.Merge(&l, s)
	test.AssertEquals(t, TextSelectionList{s}, l)
}

func TestTextSelectionMergeInner(t *testing.T) {
	s1 := CreateTextSelection(5, 10, true)
	s2 := CreateTextSelection(6, 9, false)
	l := TextSelectionList{s1}
	interval.Merge(&l, s2)
	test.AssertEquals(t, TextSelectionList{
		CreateTextSelection(5, 10, false),
	}, l)
}

func TestTextSelectionMergeAtStart(t *testing.T) {
	s1 := CreateTextSelection(6, 9, true)
	s2 := CreateTextSelection(6, 7, false)
	l := TextSelectionList{s1}
	interval.Merge(&l, s2)
	test.AssertEquals(t, TextSelectionList{
		CreateTextSelection(6, 9, false),
	}, l)
}

func TestTextSelectionMergeAtEnd(t *testing.T) {
	s1 := CreateTextSelection(6, 9, true)
	s2 := CreateTextSelection(8, 9, false)
	l := TextSelectionList{s1}
	interval.Merge(&l, s2)
	test.AssertEquals(t, TextSelectionList{
		CreateTextSelection(6, 9, false),
	}, l)
}

func TestTextSelectionMergeEncompass(t *testing.T) {
	s1 := CreateTextSelection(6, 9, false)
	s2 := CreateTextSelection(5, 10, true)
	l := TextSelectionList{s1}
	interval.Merge(&l, s2)
	test.AssertEquals(t, TextSelectionList{
		CreateTextSelection(5, 10, true),
	}, l)
}

func TestTextSelectionMergeDuplicate(t *testing.T) {
	s1 := CreateTextSelection(2, 6, false)
	s2 := CreateTextSelection(2, 6, true)
	l := TextSelectionList{s1}
	interval.Merge(&l, s2)
	test.AssertEquals(t, TextSelectionList{
		CreateTextSelection(2, 6, true),
	}, l)
}

func TestTextSelectionMergeDuplicate0Len(t *testing.T) {
	s1 := CreateTextSelection(2, 2, false)
	s2 := CreateTextSelection(2, 2, true)
	l := TextSelectionList{s1}
	interval.Merge(&l, s2)
	test.AssertEquals(t, TextSelectionList{
		CreateTextSelection(2, 2, true),
	}, l)
}

func TestTextSelectionMergeExtendStart(t *testing.T) {
	s1 := CreateTextSelection(6, 9, false)
	s2 := CreateTextSelection(1, 7, true)
	l := TextSelectionList{s1}
	interval.Merge(&l, s2)
	test.AssertEquals(t, TextSelectionList{
		CreateTextSelection(1, 9, true),
	}, l)
}

func TestTextSelectionMergeExtendEnd(t *testing.T) {
	s1 := CreateTextSelection(6, 9, true)
	s2 := CreateTextSelection(8, 15, false)
	l := TextSelectionList{s1}
	interval.Merge(&l, s2)
	test.AssertEquals(t, TextSelectionList{
		CreateTextSelection(6, 15, false),
	}, l)
}

func TestTextSelectionMergeBeforeStart(t *testing.T) {
	s1 := CreateTextSelection(6, 9, true)
	s2 := CreateTextSelection(2, 6, false)
	l := TextSelectionList{s1}
	interval.Merge(&l, s2)
	test.AssertEquals(t, TextSelectionList{
		CreateTextSelection(2, 6, false),
		CreateTextSelection(6, 9, true),
	}, l)
}

func TestTextSelectionMergeAfterEnd(t *testing.T) {
	s1 := CreateTextSelection(2, 6, false)
	s2 := CreateTextSelection(6, 9, true)
	l := TextSelectionList{s1}
	interval.Merge(&l, s2)
	test.AssertEquals(t, TextSelectionList{
		CreateTextSelection(2, 6, false),
		CreateTextSelection(6, 9, true),
	}, l)
}
