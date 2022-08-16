package labelinfo

import (
	"monster/pkg/common/define/fontengine"
)

// labelinfo
const (
	VALIGN_CENTER = iota
	VALIGN_TOP
	VALIGN_BOTTOM
)

type LabelInfo struct {
	X         int
	Y         int
	Justify   int
	Valign    int
	Hidden    bool
	FontStyle string
}

func New() *LabelInfo {
	li := Construct()
	return &li
}

func Construct() LabelInfo {
	return LabelInfo{
		Justify:   fontengine.JUSTIFY_LEFT,
		Valign:    VALIGN_TOP,
		Hidden:    false,
		FontStyle: "font_regular",
	}
}
