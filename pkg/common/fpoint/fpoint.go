package fpoint

import (
	"math"
)

type FPoint struct {
	X float32
	Y float32
}

func New(vars ...float32) *FPoint {
	p := Construct(vars...)
	return &p
}

func Construct(vars ...float32) FPoint {
	p := FPoint{}
	switch len(vars) {
	case 2:
		p.Y = vars[1]
		fallthrough
	case 1:
		p.X = vars[0]
	}

	return p
}

func (this *FPoint) Align() {
	// this rounds the float values to the nearest multiple of 1/(2^4)
	// 1/(2^4) was chosen because it's a "nice" floating point number, removing 99% of rounding errors
	this.X = (float32)(math.Floor((float64)(this.X/0.0625)) * 0.0625)
}
