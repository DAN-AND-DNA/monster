package point

type Point struct {
	X int
	Y int
}

func New(vars ...int) *Point {
	p := Construct(vars...)
	return &p
}

func Construct(vars ...int) Point {
	p := Point{}
	switch len(vars) {
	case 2:
		p.Y = vars[1]
		fallthrough
	case 1:
		p.X = vars[0]
	}

	return p
}
