package resources

import (
	"monster/pkg/resources/animation"
	"monster/pkg/resources/renderable"
)

type Factory struct {
	name string
}

func NewFactory() *Factory {
	return &Factory{
		name: "resources factory",
	}
}

func (obj Factory) New(type1 string) interface{} {
	switch type1 {
	case "animationset":
		return &animation.Set{}
	case "renderable":
		return &renderable.Renderable{}
	}

	panic("bad type for " + obj.name + ": " + type1)
}
