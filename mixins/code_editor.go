// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mixins

import (
	"fmt"
	"strings"

	"github.com/mmogo/gxui"
	"github.com/mmogo/gxui/math"
)

type CodeEditorOuter interface {
	TextBoxOuter
	CreateSuggestionList() gxui.List
}

type CodeEditor struct {
	TextBox
	outer              CodeEditorOuter
	layers             gxui.CodeSyntaxLayers
	suggestionAdapter  *SuggestionAdapter
	suggestionList     gxui.List
	suggestionProvider gxui.CodeSuggestionProvider
	tabWidth           int
	tabSpaces          bool
	theme              gxui.Theme
}

func (e *CodeEditor) updateSpans(edits []gxui.TextBoxEdit) {
	runeCount := len(e.controller.TextRunes())
	for _, l := range e.layers {
		l.UpdateSpans(runeCount, edits)
	}
}

func (e *CodeEditor) Init(outer CodeEditorOuter, driver gxui.Driver, theme gxui.Theme, font gxui.Font) {
	e.outer = outer
	e.tabWidth = 2
	e.theme = theme

	e.suggestionAdapter = &SuggestionAdapter{}
	e.suggestionList = e.outer.CreateSuggestionList()
	e.suggestionList.SetAdapter(e.suggestionAdapter)

	e.TextBox.Init(outer, driver, theme, font)
	e.controller.OnTextChanged(e.updateSpans)

	e.SetTabSpaces(false)
}

func (e *CodeEditor) ItemSize(theme gxui.Theme) math.Size {
	return math.Size{W: math.MaxSize.W, H: e.font.GlyphMaxSize().H}
}

func (e *CodeEditor) CreateSuggestionList() gxui.List {
	l := e.theme.CreateList()
	l.SetBackgroundBrush(gxui.DefaultBrush)
	l.SetBorderPen(gxui.DefaultPen)
	return l
}

func (e *CodeEditor) SyntaxLayers() gxui.CodeSyntaxLayers {
	return e.layers
}

func (e *CodeEditor) SetSyntaxLayers(layers gxui.CodeSyntaxLayers) {
	e.layers = layers
	e.onRedrawLines.Fire()
}

func (e *CodeEditor) TabWidth() int {
	return e.tabWidth
}

func (e *CodeEditor) SetTabWidth(tabWidth int) {
	e.tabWidth = tabWidth
	if e.TabSpaces() {
		e.Controller().SetIndent(strings.Repeat(" ", e.tabWidth))
	}
}

func (e *CodeEditor) TabSpaces() bool {
	return e.tabSpaces
}

func (e *CodeEditor) SetTabSpaces(useSpaces bool) {
	e.tabSpaces = useSpaces
	if e.tabSpaces {
		e.Controller().SetIndent(strings.Repeat(" ", e.tabWidth))
	}
	e.Controller().SetIndent("\t")
}

func (e *CodeEditor) SuggestionProvider() gxui.CodeSuggestionProvider {
	return e.suggestionProvider
}

func (e *CodeEditor) SetSuggestionProvider(provider gxui.CodeSuggestionProvider) {
	if e.suggestionProvider != provider {
		e.suggestionProvider = provider
		if e.IsSuggestionListShowing() {
			e.ShowSuggestionList() // Update list
		}
	}
}

func (e *CodeEditor) IsSuggestionListShowing() bool {
	return e.outer.Children().Find(e.suggestionList) != nil
}

func (e *CodeEditor) SortSuggestionList() {
	caret := e.controller.LastCaret()
	partial := e.controller.TextRange(e.controller.WordAt(caret))
	e.suggestionAdapter.Sort(partial)
}

func (e *CodeEditor) ShowSuggestionList() {
	if e.suggestionProvider == nil || e.IsSuggestionListShowing() {
		return
	}

	caret := e.controller.LastCaret()
	s, _ := e.controller.WordAt(caret)

	suggestions := e.suggestionProvider.SuggestionsAt(s)
	if len(suggestions) == 0 {
		e.HideSuggestionList()
		return
	}

	e.suggestionAdapter.SetSuggestions(suggestions)
	e.SortSuggestionList()
	child := e.AddChild(e.suggestionList)

	// Position the suggestion list below the last caret
	lineIdx := e.controller.LineIndex(caret)
	// TODO: What if the last caret is not visible?
	bounds := e.Size().Rect().Contract(e.Padding())
	line := e.Line(lineIdx)
	lineOffset := gxui.ChildToParent(math.ZeroPoint, line, e.outer)
	target := line.PositionAt(caret).Add(lineOffset)
	cs := e.suggestionList.DesiredSize(math.ZeroSize, bounds.Size())
	e.suggestionList.Select(e.suggestionList.Adapter().ItemAt(0))
	e.suggestionList.SetSize(cs)
	child.Layout(cs.Rect().Offset(target).Intersect(bounds))
}

func (e *CodeEditor) HideSuggestionList() {
	if e.IsSuggestionListShowing() {
		e.RemoveChild(e.suggestionList)
	}
}

func (e *CodeEditor) Line(idx int) TextBoxLine {
	return gxui.FindControl(e.ItemControl(idx).(gxui.Parent), func(c gxui.Control) bool {
		_, b := c.(TextBoxLine)
		return b
	}).(TextBoxLine)
}

// mixins.List overrides
func (e *CodeEditor) Click(ev gxui.MouseEvent) (consume bool) {
	e.HideSuggestionList()
	return e.TextBox.Click(ev)
}

func (e *CodeEditor) KeyPress(ev gxui.KeyboardEvent) (consume bool) {
	switch ev.Key {
	case gxui.KeyTab:
		replace := true
		for _, sel := range e.controller.Selections() {
			start, end := sel.Range()
			if e.controller.LineIndex(start) != e.controller.LineIndex(end) {
				replace = false
				break
			}
		}
		switch {
		case replace:
			e.controller.ReplaceAll(strings.Repeat(" ", e.tabWidth))
			e.controller.Deselect(false)
		case ev.Modifier.Shift():
			e.controller.UnindentSelection()
		default:
			e.controller.IndentSelection()
		}
		return true
	case gxui.KeySpace:
		if ev.Modifier.Control() {
			e.ShowSuggestionList()
			return
		}
	case gxui.KeyUp:
		fallthrough
	case gxui.KeyDown:
		if e.IsSuggestionListShowing() {
			return e.suggestionList.KeyPress(ev)
		}
	case gxui.KeyLeft:
		e.HideSuggestionList()
	case gxui.KeyRight:
		e.HideSuggestionList()
	case gxui.KeyEnter:
		controller := e.controller
		if e.IsSuggestionListShowing() {
			text := e.suggestionAdapter.Suggestion(e.suggestionList.Selected()).Code()
			start, end := controller.WordAt(e.controller.LastCaret())
			controller.SetSelection(gxui.CreateTextSelection(start, end, false))
			controller.ReplaceAll(text)
			controller.Deselect(false)
			e.HideSuggestionList()
		} else {
			e.controller.ReplaceWithNewlineKeepIndent()
		}
		return true
	case gxui.KeyEscape:
		if e.IsSuggestionListShowing() {
			e.HideSuggestionList()
			return true
		}
	}
	return e.TextBox.KeyPress(ev)
}

func (e *CodeEditor) KeyStroke(ev gxui.KeyStrokeEvent) (consume bool) {
	consume = e.TextBox.KeyStroke(ev)
	if e.IsSuggestionListShowing() {
		e.SortSuggestionList()
	}
	return
}

// mixins.TextBox overrides
func (e *CodeEditor) CreateLine(theme gxui.Theme, index int) (TextBoxLine, gxui.Control) {
	lineNumber := theme.CreateLabel()
	lineNumber.SetText(fmt.Sprintf("%.4d", index+1)) // Displayed lines start at 1

	line := &CodeEditorLine{}
	line.Init(line, theme, e, index)

	layout := theme.CreateLinearLayout()
	layout.SetDirection(gxui.LeftToRight)
	layout.AddChild(lineNumber)
	layout.AddChild(line)

	return line, layout
}
