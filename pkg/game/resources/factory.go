package resources

import (
	"monster/pkg/game/resources/effect"
	"monster/pkg/game/resources/gameslotpreview"
	"monster/pkg/game/resources/mapcollision"
	"monster/pkg/game/resources/statblock"
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
	case "statblock":
		return &statblock.StatBlock{}
	case "effectmanager":
		return &effect.Manager{}
	case "gameslotpreview":
		return &gameslotpreview.GameSlotPreview{}
	case "mapcollision":
		return &mapcollision.MapCollision{}
	}

	panic("bad type for " + obj.name + ": " + type1)
}
