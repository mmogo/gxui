package basic

import (
	"github.com/mmogo/gxui"
	"github.com/mmogo/gxui/mixins"
)

func CreateTableLayout(theme *Theme) gxui.TableLayout {
	l := &mixins.TableLayout{}
	l.Init(l, theme)
	return l
}
