package timer

const (
	END = iota
	BEGIN
)

type Timer struct {
	current  uint
	duration uint
}

func Construct(vars ...uint) Timer {
	t := Timer{}
	if len(vars) == 1 {
		t.duration = vars[0]
	}

	return t
}

func New(vars ...uint) *Timer {
	t := Construct(vars...)
	return &t
}

func (this *Timer) GetCurrent() uint {
	return this.current
}

func (this *Timer) GetDuration() uint {
	return this.duration
}

func (this *Timer) SetCurrent(val uint) {
	this.current = val
	if this.current > this.duration {
		this.current = this.duration
	}
}

func (this *Timer) SetDuration(val uint) {
	this.duration = val
	this.current = this.duration
}

func (this *Timer) Tick() bool {
	if this.current > 0 {
		this.current--
	}

	if this.current == 0 {
		return true
	}

	return false
}

func (this *Timer) IsEnd() bool {
	return (this.current == 0)
}

func (this *Timer) IsBegin() bool {
	return this.current == this.duration
}

func (this *Timer) Reset(type1 int) {
	if type1 == END {
		this.current = 0
	} else if type1 == BEGIN {
		this.current = this.duration
	}
}

func (this *Timer) IsWholeSecond(maxFPS int) bool {
	return (this.current)%(uint)(maxFPS) == 0
}
