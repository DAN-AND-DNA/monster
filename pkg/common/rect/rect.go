package rect

type Rect struct {
	X int
	Y int
	W int
	H int
}

func New(vars ...int) *Rect {
	r := Construct(vars...)
	return &r
}

func Construct(vars ...int) Rect {
	r := Rect{}
	switch len(vars) {
	case 4:
		r.H = vars[3]
		fallthrough
	case 3:
		r.W = vars[2]
		fallthrough
	case 2:
		r.Y = vars[1]
		fallthrough
	case 1:
		r.X = vars[0]
	}

	return r
}
