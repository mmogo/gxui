// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mixins

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/interval"
	"github.com/nelsam/gxui/math"
)

type CodeEditorLinePaintInfo struct {
	LineSpan     interval.IntData
	Runes        []rune
	GlyphOffsets []math.Point
	GlyphWidth   int
	LineHeight   int
	Font         gxui.Font
}

type CodeEditorLineOuter interface {
	DefaultTextBoxLineOuter
	PaintEditorSelections(gxui.Canvas, CodeEditorLinePaintInfo)
	PaintEditorCarets(gxui.Canvas, CodeEditorLinePaintInfo)
	PaintBackgroundSpans(gxui.Canvas, CodeEditorLinePaintInfo)
	PaintGlyphs(gxui.Canvas, CodeEditorLinePaintInfo)
	PaintBorders(gxui.Canvas, CodeEditorLinePaintInfo)
}

// CodeEditorLine
type CodeEditorLine struct {
	DefaultTextBoxLine
	outer CodeEditorLineOuter
	ce    *CodeEditor
}

func (l *CodeEditorLine) Init(outer CodeEditorLineOuter, theme gxui.Theme, ce *CodeEditor, lineIndex int) {
	l.DefaultTextBoxLine.Init(outer, theme, &ce.TextBox, lineIndex)
	l.outer = outer
	l.ce = ce
}

func (l *CodeEditorLine) RuneIndexAt(p math.Point) int {
	font := l.ce.Font()
	controller := l.ce.Controller()

	x := p.X
	i := 0
	offsets := l.offsets(font)
	for ; i < len(offsets) && x > offsets[i].X; i++ {
	}

	return controller.LineStart(l.lineIndex) + i
}

func (l *CodeEditorLine) PaintBackgroundSpans(c gxui.Canvas, info CodeEditorLinePaintInfo) {
	start, _ := info.LineSpan.Span()
	offsets := info.GlyphOffsets
	remaining := interval.IntDataList{info.LineSpan}
	for _, layer := range l.ce.layers {
		if layer != nil && layer.BackgroundColor() != nil {
			color := *layer.BackgroundColor()
			for _, span := range layer.Spans().Overlaps(info.LineSpan) {
				interval.Visit(&remaining, span, func(vs, ve uint64, _ int) {
					s, e := vs-start, ve-start
					r := math.CreateRect(offsets[s].X, 0, offsets[e-1].X+info.GlyphWidth, info.LineHeight)
					c.DrawRoundedRect(r, 3, 3, 3, 3, gxui.TransparentPen, gxui.Brush{Color: color})
				})
				interval.Remove(&remaining, span)
			}
		}
	}
}

func (l *CodeEditorLine) PaintGlyphs(c gxui.Canvas, info CodeEditorLinePaintInfo) {
	start, _ := info.LineSpan.Span()
	runes, offsets, font := info.Runes, info.GlyphOffsets, info.Font
	remaining := interval.IntDataList{info.LineSpan}
	for _, layer := range l.ce.layers {
		if layer != nil && layer.Color() != nil {
			color := *layer.Color()
			for _, span := range layer.Spans().Overlaps(info.LineSpan) {
				interval.Visit(&remaining, span, func(vs, ve uint64, _ int) {
					s, e := vs-start, ve-start
					c.DrawRunes(font, runes[s:e], offsets[s:e], color)
				})
				interval.Remove(&remaining, span)
			}
		}
	}
	for _, span := range remaining {
		s, e := span.Span()
		s, e = s-start, e-start
		c.DrawRunes(font, runes[s:e], offsets[s:e], l.ce.textColor)
	}
}

func (l *CodeEditorLine) PaintBorders(c gxui.Canvas, info CodeEditorLinePaintInfo) {
	start, _ := info.LineSpan.Span()
	offsets := info.GlyphOffsets
	for _, layer := range l.ce.layers {
		if layer != nil && layer.BorderColor() != nil {
			color := *layer.BorderColor()
			interval.Visit(layer.Spans(), info.LineSpan, func(vs, ve uint64, _ int) {
				s, e := vs-start, ve-start
				r := math.CreateRect(offsets[s].X, 0, offsets[e-1].X+info.GlyphWidth, info.LineHeight)
				c.DrawRoundedRect(r, 3, 3, 3, 3, gxui.CreatePen(0.5, color), gxui.TransparentBrush)
			})
		}
	}
}

// endOfChar takes a position of a character and returns the position
// of its end.
func (l *CodeEditorLine) endOfChar(position math.Point) math.Point {
	return position.AddX(l.ce.Font().GlyphMaxSize().W / 2)
}

func (l *CodeEditorLine) PaintEditorCarets(c gxui.Canvas, info CodeEditorLinePaintInfo) {
	controller := l.textbox.controller
	for i, count := 0, controller.SelectionCount(); i < count; i++ {
		caret := controller.Caret(i)
		line := controller.LineIndex(caret)
		if line == l.lineIndex {
			var offset math.Point
			start := controller.LineStart(line)
			if len(info.GlyphOffsets) > 0 && caret > start {
				caretOffsetIndex := caret - start - 1
				offset = l.endOfChar(info.GlyphOffsets[caretOffsetIndex])
			}
			top := math.Point{X: l.caretWidth + offset.X, Y: 0}
			bottom := top.Add(math.Point{X: 0, Y: l.Size().H})
			l.outer.PaintCaret(c, top, bottom)
		}
	}
}

func (l *CodeEditorLine) PaintEditorSelections(c gxui.Canvas, info CodeEditorLinePaintInfo) {
	controller := l.textbox.controller

	ls, le := uint64(controller.LineStart(l.lineIndex)), uint64(controller.LineEnd(l.lineIndex))

	selections := controller.Selections()
	if l.textbox.selectionDragging {
		interval.Replace(&selections, l.textbox.selectionDrag)
	}
	interval.Visit(&selections, gxui.CreateTextSelection(int(ls), int(le), false), func(s, e uint64, _ int) {
		if s < e {
			var startOffset math.Point
			if s > ls {
				startOffset = l.endOfChar(info.GlyphOffsets[s-ls-1])
			}
			var endOffset math.Point
			if e > ls {
				endOffset = l.endOfChar(info.GlyphOffsets[e-ls-1])
			}

			width := endOffset.X - startOffset.X
			m := l.outer.MeasureRunes(int(s), int(e))
			m.W = width
			top := math.Point{X: l.caretWidth + startOffset.X, Y: 0}
			bottom := top.Add(m.Point())
			l.outer.PaintSelection(c, top, bottom)
		}
	})
}

func (l *CodeEditorLine) offsets(font gxui.Font) []math.Point {
	rect := l.Size().Rect().OffsetX(l.caretWidth)
	runes := l.ce.Controller().LineRunes(l.lineIndex)
	offsets := font.Layout(&gxui.TextBlock{
		Runes:     runes,
		AlignRect: rect,
		H:         gxui.AlignLeft,
		V:         gxui.AlignMiddle,
	})
	l.applyTabWidth(runes, offsets, font)
	for i, offset := range offsets {
		offsets[i] = offset.AddX(-l.offset)
	}
	return offsets
}

// DefaultTextBoxLine overrides
func (l *CodeEditorLine) Paint(c gxui.Canvas) {
	font := l.ce.Font()
	controller := l.ce.Controller()
	start := controller.LineStart(l.lineIndex)
	end := controller.LineEnd(l.lineIndex)

	var info CodeEditorLinePaintInfo
	if start != end {
		info = CodeEditorLinePaintInfo{
			LineSpan:     interval.CreateIntData(start, end, nil),
			Runes:        controller.LineRunes(l.lineIndex),
			GlyphOffsets: l.offsets(font),
			GlyphWidth:   font.GlyphMaxSize().W,
			LineHeight:   l.Size().H,
			Font:         font,
		}

		l.outer.PaintBackgroundSpans(c, info)

		if l.textbox.HasFocus() {
			l.outer.PaintEditorSelections(c, info)
		}

		l.outer.PaintGlyphs(c, info)

		l.outer.PaintBorders(c, info)
	}

	if l.textbox.HasFocus() {
		l.outer.PaintEditorCarets(c, info)
	}
}

func (l *CodeEditorLine) applyTabWidth(runes []rune, offsets []math.Point, font gxui.Font) {
	tabWidth := font.Measure(&gxui.TextBlock{Runes: []rune{'\t'}}).W
	spaceWidth := font.Measure(&gxui.TextBlock{Runes: []rune{' '}}).W
	tabExtra := spaceWidth*l.ce.TabWidth() - tabWidth
	for i, r := range runes {
		if r == '\t' {
			for j := i; j < len(offsets); j++ {
				offsets[j].X += tabExtra
			}
		}
	}
	return
}
