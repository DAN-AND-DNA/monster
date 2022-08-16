package widget

import (
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define/fontengine"
	"monster/pkg/common/define/inputstate"
	"monster/pkg/common/define/widget/checkbox"
	"monster/pkg/common/point"
	"monster/pkg/common/tooltipdata"
	"monster/pkg/utils"
	"monster/pkg/widget/base"
)

type CheckBox struct {
	base.Widget
	enabled   bool
	tooltip   string
	cb        common.Sprite
	checked   bool
	pressed   bool
	activated bool
}

func NewCheckBox(modules common.Modules, fname string) *CheckBox {
	cb := &CheckBox{}
	cb.Init(modules, fname)

	return cb
}

func (this *CheckBox) Init(modules common.Modules, fname string) common.WidgetCheckBox {
	render := modules.Render()
	mods := modules.Mods()
	settings := modules.Settings()

	// base
	this.Widget = base.ConstructWidget()

	// self
	this.enabled = true
	this.SetFocusable(true)

	tmpfname := checkbox.DEFAULT_FILE
	if fname != checkbox.DEFAULT_FILE {
		tmpfname = fname
	}

	graphics, err := render.LoadImage(settings, mods, tmpfname)
	if err != nil {
		panic(err)
	}
	defer graphics.UnRef()

	this.cb, err = graphics.CreateSprite()
	if err != nil {
		panic(err)
	}

	gw, err := this.cb.GetGraphicsWidth()
	if err != nil {
		panic(err)
	}

	gh, err := this.cb.GetGraphicsHeight()
	if err != nil {
		panic(err)
	}

	this.SetPosW(gw)
	this.SetPosH(gh / 2)
	this.cb.SetClip(0, 0, this.GetPos().W, this.GetPos().H)

	return this
}

func (this *CheckBox) Clear() {
	if this.cb != nil {
		this.cb.Close()
		this.cb = nil
	}
}

func (this *CheckBox) Close() {
	this.Widget.Close(this)
}

func (this *CheckBox) Activate() {
	this.pressed = true
	this.activated = true
}

func (this *CheckBox) Deactivate() {
	// pass
}

func (this *CheckBox) SetChecked(status bool) {
	this.checked = status
	if this.cb != nil {

		// 更换渲染的来源图片
		pos := this.GetPos()
		if this.checked {
			this.cb.SetClip(0, pos.H, pos.W, pos.H)
		} else {
			this.cb.SetClip(0, 0, pos.W, pos.H)
		}
	}
}

func (this *CheckBox) GetChecked() bool {
	return this.checked
}

func (this *CheckBox) CheckTooltip(modules common.Modules, mouse point.Point) {
	font := modules.Font()
	inpt := modules.Inpt()
	tooltipm := modules.Tooltipm()
	settings := modules.Settings()

	tipData := tooltipdata.Construct()
	pos := this.GetPos()

	if inpt.UsingMouse(settings) && utils.IsWithinRect(pos, mouse) && this.tooltip != "" {
		tipData.AddColorText(this.tooltip, font.GetColor(fontengine.COLOR_WIDGET_NORMAL))
	}

	if !tipData.IsEmpty() {
		newMouse := point.Construct(
			mouse.X+this.GetLocalFrame().X-this.GetLocalOffset().X,
			mouse.Y+this.GetLocalFrame().Y-this.GetLocalOffset().Y,
		)
		tooltipm.Push(tipData, newMouse, tooltipdata.STYLE_FLOAT, 0)
	}
}

func (this *CheckBox) CheckClickAt(modules common.Modules, x, y int) bool {
	inpt := modules.Inpt()
	settings := modules.Settings()

	this.SetEnableTablistNav(this.enabled)

	if !this.enabled {
		return false
	}

	mouse := point.Construct(x, y)
	this.CheckTooltip(modules, mouse)

	// 左键已经在使用，无法再使用
	if inpt.GetLock(inputstate.MAIN1) {
		return false
	}

	// 键盘
	if !inpt.UsingMouse(settings) && inpt.GetLock(inputstate.ACCEPT) {
		return false
	}

	// 左键释放或使用键盘
	if this.pressed && !inpt.GetLock(inputstate.MAIN1) &&
		(!inpt.GetLock(inputstate.ACCEPT) || inpt.UsingMouse(settings)) &&
		(utils.IsWithinRect(this.GetPos(), mouse) || this.activated) {

		this.activated = false
		this.pressed = false
		this.SetChecked(!this.checked)
		return true
	}

	this.pressed = false

	// 处理左键按下
	if inpt.GetPressing(inputstate.MAIN1) {
		if utils.IsWithinRect(this.GetPos(), mouse) {
			this.pressed = true
			inpt.SetLock(inputstate.MAIN1, true)
		}
	}

	return false
}

func (this *CheckBox) CheckClick(modules common.Modules) {
	inpt := modules.Inpt()
	mouse := inpt.GetMouse()

	this.CheckClickAt(modules, mouse.X, mouse.Y)
}

func (this *CheckBox) Render(modules common.Modules) error {
	render := modules.Render()
	eset := modules.Eset()

	if this.cb != nil {
		this.cb.SetLocalFrame(this.GetLocalFrame())
		this.cb.SetOffset(this.GetLocalOffset())
		this.cb.SetDestFromRect(this.GetPos()) // 绘制位置为
		err := render.Render(this.cb)
		if err != nil {
			return err
		}
	}

	// 选择边框
	if this.GetInFocus() {
		topLeft := point.Construct()
		bottomRight := point.Construct()

		topLeft.X = this.GetPos().X + this.GetLocalFrame().X - this.GetLocalOffset().X
		topLeft.Y = this.GetPos().Y + this.GetLocalFrame().Y - this.GetLocalOffset().Y
		bottomRight.X = topLeft.X + this.GetPos().W
		bottomRight.Y = topLeft.Y + this.GetPos().H

		// 在范围内绘制
		draw := true
		if this.GetLocalFrame().W > 0 &&
			topLeft.X < this.GetLocalFrame().X ||
			(bottomRight.X > this.GetLocalFrame().X+this.GetLocalFrame().W) {
			draw = false
		}

		if this.GetLocalFrame().H > 0 &&
			topLeft.Y < this.GetLocalFrame().Y ||
			(bottomRight.Y > this.GetLocalFrame().Y+this.GetLocalFrame().H) {
			draw = false
		}

		if draw {
			err := render.DrawRectangle(topLeft, bottomRight, eset.Get("widgets", "selection_rect_color").(color.Color))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (this *CheckBox) GetNext(modules common.Modules) bool {
	return false
}

func (this *CheckBox) GetPrev(modules common.Modules) bool {
	return false
}

func (this *CheckBox) SetTooltip(val string) {
	this.tooltip = val
}

func (this *CheckBox) SetEnabled(val bool) {
	this.enabled = val
}
