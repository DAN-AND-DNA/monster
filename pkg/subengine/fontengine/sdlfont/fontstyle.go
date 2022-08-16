package sdlfont

import (
	"monster/pkg/subengine/fontengine/base"

	"github.com/veandco/go-sdl2/ttf"
)

type FontStyle struct {
	base.FontStyle
	ttfont *ttf.Font
}

func ConstructFontStyle() FontStyle {
	impl := FontStyle{}
	impl.FontStyle = base.ConstructFontStyle()

	return impl
}

func (this *FontStyle) Ttfont() interface{} {
	return this.ttfont
}
