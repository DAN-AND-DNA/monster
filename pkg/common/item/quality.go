package item

import (
	"monster/pkg/common/color"
)

type Quality struct {
	Id          string
	Name        string
	Color       color.Color
	OverlayIcon int
}

func ConstructQuality() Quality {
	iq := Quality{
		Color:       color.Construct(255, 255, 255),
		OverlayIcon: -1,
	}

	return iq
}
