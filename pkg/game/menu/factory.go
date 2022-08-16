package menu

type Factory struct {
	name string
}

func NewFactory() *Factory {
	return &Factory{
		name: "menu factory",
	}
}

func (obj Factory) New(type1 string) interface{} {
	switch type1 {
	case "config":
		return &Config{}
	case "confirm":
		return &Confirm{}
	case "exit":
		return &Exit{}
	case "statbar":
		return &StatBar{}
	case "inventory":
		return &Inventory{}
	case "actionbar":
		return &ActionBar{}
	}

	panic("bad type for " + obj.name + ": " + type1)
}
