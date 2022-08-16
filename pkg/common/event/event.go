package event

import (
	"monster/pkg/common/fpoint"
	"monster/pkg/common/rect"
	"monster/pkg/common/timer"
)

const (
	ACTIVATE_ON_TRIGGER  = 0
	ACTIVATE_ON_INTERACT = 1
	ACTIVATE_ON_MAPEXIT  = 2
	ACTIVATE_ON_LEAVE    = 3
	ACTIVATE_ON_LOAD     = 4
	ACTIVATE_ON_CLEAR    = 5
	ACTIVATE_STATIC      = 6
)

type Event struct {
	Type             string
	ActivateType     int // 激活类型
	Components       []Component
	Location         rect.Rect   // 范围
	Hotspot          rect.Rect   // 触发点
	Cooldown         timer.Timer // 冷却
	Delay            timer.Timer // 执行前的延长
	KeepAfterTrigger bool
	Center           fpoint.FPoint // 范围的中心（范围和触发点2选1）
	ReachableFrom    rect.Rect
}

func New() *Event {
	e := Construct()
	return &e
}

func Construct() Event {
	return Event{
		Location:      rect.Construct(),
		Hotspot:       rect.Construct(),
		Cooldown:      timer.Construct(),
		Delay:         timer.Construct(),
		Center:        fpoint.Construct(-1, -1),
		ReachableFrom: rect.Construct(),
	}
}

func (this *Event) GetComponent(type1 int) (*Component, bool) {
	for _, val := range this.Components {
		if val.Type == type1 {
			tmp := val
			return &tmp, true
		}
	}

	return nil, false
}

func (this *Event) DeleteAllComponent(type1 int) {
	del := map[int]struct{}{}

	for index, val := range this.Components {
		if val.Type == type1 {
			del[index] = struct{}{}
		}
	}

	var tmp []Component
	for index, val := range this.Components {
		if _, ok := del[index]; ok {
			continue
		}

		tmp = append(tmp, val)
	}

	this.Components = tmp
}

func (this *Event) DeepCopy() *Event {
	e := *this
	e.Components = nil
	e.Components = make([]Component, len(this.Components))

	for i := 0; i < len(this.Components); i++ {
		e.Components[i] = this.Components[i]
	}

	return &e
}
