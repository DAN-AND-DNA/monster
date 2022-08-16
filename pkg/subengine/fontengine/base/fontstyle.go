package base

type FontStyle struct {
	Name       string
	Path       string
	PtSize     int
	Blend      bool
	LineHeight int
	FontHeight int
}

func ConstructFontStyle() FontStyle {
	fs := FontStyle{
		Blend: true,
	}
	return fs
}
