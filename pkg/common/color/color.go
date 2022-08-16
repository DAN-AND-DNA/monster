package color

type Color struct {
	R uint8
	G uint8
	B uint8
	A uint8
}

func New(vars ...uint8) *Color {
	c := Construct(vars...)
	return &c
}

func Construct(vars ...uint8) Color {
	c := Color{R: 0, G: 0, B: 0, A: 255}
	switch len(vars) {
	case 4:
		c.A = vars[3]
		fallthrough
	case 3:
		c.B = vars[2]
		fallthrough
	case 2:
		c.G = vars[1]
		fallthrough
	case 1:
		c.R = vars[0]
	}

	return c

}

func (this *Color) RGBA() (uint32, uint32, uint32, uint32) {
	return (uint32)(this.R), uint32(this.G), uint32(this.B), uint32(this.A)
}

func (this Color) EncodeRGBA() uint32 {
	result := (uint32)(this.A)
	result |= (uint32)(this.R) << 24
	result |= (uint32)(this.G) << 16
	result |= (uint32)(this.B) << 8
	return result
}

func (this *Color) DecodeRGBA(encoded uint32) {
	this.A = (uint8)(encoded)
	this.R = (uint8)(encoded >> 24)
	this.G = (uint8)(encoded >> 16)
	this.B = (uint8)(encoded >> 8)
}
