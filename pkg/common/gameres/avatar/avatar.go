package avatar

const (
	MSG_NORMAL = 0
	MSG_UNIQUE = 1
)

type LayerGfx struct {
	Gfx  string
	Type string
}

func ConstructLayerGfx() LayerGfx {
	return LayerGfx{}
}
