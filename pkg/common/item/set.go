package item

import (
	"monster/pkg/common/color"
	"monster/pkg/common/define"
)

type Set struct {
	Name  string
	Items []define.ItemId
	Bonus []SetBonusData
	Color color.Color
}

func NewSet() *Set {
	s := ConstructSet()
	return &s
}

func ConstructSet() Set {
	return Set{
		Color: color.Construct(255, 255, 255),
	}
}
