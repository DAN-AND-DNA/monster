package widget

type Factory struct {
	name string
}

func NewFactory() *Factory {
	return &Factory{
		name: "widget factory",
	}
}

func (obj Factory) New(type1 string) interface{} {
	switch type1 {
	case "tooltip":
		return &Tooltip{}
	case "tabcontrol":
		return &TabControl{}
	case "button":
		return &Button{}
	case "label":
		return &Label{}
	case "tablist":
		return &Tablist{}
	case "slider":
		return &Slider{}
	case "checkbox":
		return &CheckBox{}
	case "listbox":
		return &ListBox{}
	case "horizontallist":
		return &HorizontalList{}
	case "scrollbox":
		return &ScrollBox{}
	case "scrollbar":
		return &ScrollBar{}
	case "slot":
		return &Slot{}
	}
	panic("bad type for " + obj.name + ": " + type1)
}
