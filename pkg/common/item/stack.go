package item

import (
	"monster/pkg/common/define"
	"monster/pkg/common/point"
)

type Stack struct {
	Item       define.ItemId
	Quantity   int // 数量
	CanBuyback bool
}

func ConstructStack() Stack {
	return Stack{}
}

func ConstructStack1(item define.ItemId, quantity int) Stack {
	return Stack{
		Item:     item,
		Quantity: quantity,
	}
}

func ConstructStack2(p point.Point) Stack {
	return Stack{
		Item:     (define.ItemId)(p.X),
		Quantity: p.Y,
	}
}

func (this *Stack) Empty() bool {
	if this.Item != 0 && this.Quantity > 0 {
		return false
	} else if this.Item == 0 && this.Quantity != 0 {
		this.Clear()
	} else if this.Item != 0 && this.Quantity == 0 {
		this.Clear()
	}

	return true
}

func (this *Stack) Clear() {
	this.Item = 0
	this.Quantity = 0
	this.CanBuyback = false
}
