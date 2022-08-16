package widget

import (
	"math"
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define/fontengine"
	"monster/pkg/common/define/inputstate"
	"monster/pkg/common/define/widget"
	"monster/pkg/common/define/widget/slider"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/common/tooltipdata"
	"monster/pkg/utils"
	"monster/pkg/widget/base"
	"strconv"
)

type Slider struct {
	base.Widget

	posKnob             rect.Rect
	enabled             bool
	sl                  common.Sprite
	pressed             bool
	changedWithoutMouse bool
	minimum             int
	maximum             int
	value               int
	debug               bool
}

func NewSlider(modules common.Modules, fname string) *Slider {
	s := &Slider{}
	s.Init(modules, fname)

	return s
}

func (this *Slider) Init(modules common.Modules, fname string) common.WidgetSlider {
	render := modules.Render()
	settings := modules.Settings()
	mods := modules.Mods()

	// base
	this.Widget = base.ConstructWidget()

	// self
	this.enabled = true

	tmpFname := slider.DEFAULT_FILE
	if fname != slider.DEFAULT_FILE {
		tmpFname = fname
	}

	graphics, err := render.LoadImage(settings, mods, tmpFname)
	if err != nil {
		panic(err)
	}
	defer graphics.UnRef()

	this.sl, err = graphics.CreateSprite()
	if err != nil {
		panic(err)
	}

	gw, err := this.sl.GetGraphicsWidth()
	if err != nil {
		panic(err)
	}

	gh, err := this.sl.GetGraphicsHeight()
	if err != nil {
		panic(err)
	}

	this.SetPosW(gw)
	this.SetPosH(gh / 2)
	this.posKnob.W = gw / 8
	this.posKnob.H = gh / 2

	this.SetScrollType(widget.SCROLL_HORIZONTAL)

	return this
}

func (this *Slider) Clear() {
	if this.sl != nil {
		this.sl.Close()
	}
}

func (this *Slider) Close() {
	this.Widget.Close(this)
}

// 设置滑块的位置
func (this *Slider) Set(min, max, val int) {
	this.minimum = min
	this.maximum = max
	this.value = val

	pos := this.GetPos()

	if this.maximum-this.minimum != 0 {
		this.posKnob.X = pos.X + ((val-min)*pos.W)/(max-min) - (this.posKnob.W / 2)
		this.posKnob.Y = pos.Y
	}
}

func (this *Slider) SetPos1(modules common.Modules, offsetX, offsetY int) error {
	this.Widget.SetPos1(modules, offsetX, offsetY)
	this.Set(this.minimum, this.maximum, this.value)

	return nil
}

func (this *Slider) CheckClick() {
}

func (this *Slider) CheckClickAt(modules common.Modules, x, y int) bool {
	inpt := modules.Inpt()

	this.SetEnableTablistNav(this.enabled)

	if !this.enabled {
		return false
	}

	mouse := point.Construct(x, y)

	// 处理鼠标左键
	if !this.pressed && inpt.GetPressing(inputstate.MAIN1) && !inpt.GetLock(inputstate.MAIN1) {
		if utils.IsWithinRect(this.posKnob, mouse) {
			this.pressed = true
			inpt.SetLock(inputstate.MAIN1, true)
			return true
		}
		return false
	}

	// 处理键盘
	if this.changedWithoutMouse {
		this.changedWithoutMouse = false
		return true
	}

	// 正在使用
	if inpt.GetLock(inputstate.UP) {
		return false
	}

	if inpt.GetLock(inputstate.DOWN) {
		return false
	}

	pos := this.GetPos()
	if this.pressed {
		if !inpt.GetLock(inputstate.MAIN1) {
			this.pressed = false
		}

		// 鼠标距离坐标起点的差值
		tmp := (int)(math.Max(0, math.Min((float64)(mouse.X-pos.X), (float64)(pos.W))))
		this.posKnob.X = pos.X + tmp - (this.posKnob.W / 2)

		// 换算成滑块的值
		this.value = this.minimum + (tmp*(this.maximum-this.minimum))/pos.W

		return true
	}

	return false
}

func (this *Slider) Activate() {

}

func (this *Slider) Deactivate() {

}

func (this *Slider) GetNext(modules common.Modules) bool {
	if !this.enabled {
		return false
	}

	this.value += (this.maximum - this.minimum) / 10
	if this.value > this.maximum {
		this.value = this.maximum
	}

	pos := this.GetPos()
	this.posKnob.X = pos.X + ((this.value-this.minimum)*pos.W)/(this.maximum-this.minimum) - this.posKnob.W/2
	this.posKnob.Y = pos.Y

	this.changedWithoutMouse = true

	return true
}

func (this *Slider) GetPrev(modules common.Modules) bool {
	if !this.enabled {
		return false
	}

	this.value -= (this.maximum - this.minimum) / 10
	if this.value < this.minimum {
		this.value = this.minimum
	}

	pos := this.GetPos()
	this.posKnob.X = pos.X + ((this.value-this.minimum)*pos.W)/(this.maximum-this.minimum) - this.posKnob.W/2
	this.posKnob.Y = pos.Y

	this.changedWithoutMouse = true

	return true

}

func (this *Slider) Render(modules common.Modules) error {
	render := modules.Render()
	eset := modules.Eset()
	font := modules.Font()
	tooltipm := modules.Tooltipm()

	pos := this.GetPos()
	base := rect.Construct(0, 0, pos.W, pos.H)
	knob := rect.Construct(0, pos.H, this.posKnob.W, this.posKnob.H)
	_ = knob
	_ = base

	if this.sl != nil {
		this.sl.SetLocalFrame(this.GetLocalFrame())
		this.sl.SetOffset(this.GetLocalOffset())
		this.sl.SetClipFromRect(base)
		this.sl.SetDestFromRect(pos) // 多了x和y
		render.Render(this.sl)
		this.sl.SetClipFromRect(knob)
		this.sl.SetDestFromRect(this.posKnob)
		render.Render(this.sl)
	}

	if this.GetInFocus() {
		// 选择框
		topLeft := point.Construct()
		bottomRight := point.Construct()

		topLeft.X = pos.X + this.GetLocalFrame().X - this.GetLocalOffset().X
		topLeft.Y = pos.Y + this.GetLocalFrame().Y - this.GetLocalOffset().Y
		bottomRight.X = topLeft.X + pos.W
		bottomRight.Y = topLeft.Y + pos.H

		draw := true
		if this.GetLocalFrame().W > 0 &&
			(topLeft.X < this.GetLocalFrame().X || bottomRight.X > (this.GetLocalFrame().X+this.GetLocalFrame().W)) {
			draw = false
		}

		if this.GetLocalFrame().H > 0 &&
			(topLeft.Y < this.GetLocalFrame().Y || bottomRight.Y > (this.GetLocalFrame().Y+this.GetLocalFrame().H)) {
			draw = false
		}

		if draw {
			err := render.DrawRectangle(topLeft, bottomRight, eset.Get("widgets", "selection_rect_color").(color.Color))
			if err != nil {
				return err
			}
		}
	}

	if this.pressed || this.GetInFocus() {
		tipData := tooltipdata.Construct()
		tipData.AddColorText(strconv.Itoa(this.value), font.GetColor(fontengine.COLOR_WIDGET_NORMAL))
		newMouse := point.Construct()
		newMouse.X = this.posKnob.X + this.posKnob.W*2 + this.GetLocalFrame().X - this.GetLocalOffset().X
		newMouse.Y = this.posKnob.Y + this.posKnob.H/2 + this.GetLocalFrame().Y - this.GetLocalOffset().Y
		tooltipm.Push(tipData, newMouse, tooltipdata.STYLE_TOPLABEL, 0)
	}

	return nil
}

func (this *Slider) SetEnabled(val bool) {
	this.enabled = val
}
