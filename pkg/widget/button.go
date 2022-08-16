package widget

import (
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define/fontengine"
	"monster/pkg/common/define/inputstate"
	"monster/pkg/common/define/widget/button"
	"monster/pkg/common/labelinfo"
	"monster/pkg/common/point"
	"monster/pkg/common/tooltipdata"
	"monster/pkg/utils"
	"monster/pkg/widget/base"
)

type Button struct {
	base.Widget
	filename          string // 按钮背景图片
	buttons           common.Sprite
	wlabel            *Label // 文字
	activated         bool
	label             string
	textColorNormal   color.Color
	textColorPressed  color.Color
	textColorHover    color.Color
	textColorDisabled color.Color
	tooltip           string // 提示
	enabled           bool
	pressed           bool
	hover             bool
}

func NewButton(modules common.Modules, filename string) *Button {
	b := &Button{}
	_ = (common.WidgetButton)(b)

	b.Init(modules, filename)

	return b
}

func (this *Button) Init(modules common.Modules, filename string) common.WidgetButton {
	font := modules.Font()

	// base
	this.Widget = base.ConstructWidget()

	// self
	this.filename = filename
	this.wlabel = NewLabel(modules)
	this.enabled = true //按钮默认可用
	this.SetFocusable(true)
	this.LoadArt(modules) // 加载按钮图片创建精灵
	this.textColorNormal = font.GetColor(fontengine.COLOR_WIDGET_NORMAL)
	this.textColorPressed = this.textColorNormal
	this.textColorHover = this.textColorNormal
	this.textColorDisabled = font.GetColor(fontengine.COLOR_WIDGET_DISABLED)

	return this
}

// 清理自己
func (this *Button) Clear() {
	if this.buttons != nil {
		this.buttons.Close()
	}

	if this.wlabel != nil {
		this.wlabel.Close()
	}
}

func (this *Button) Close() {
	this.Widget.Close(this)
}

// 加载按钮图片创建精灵
func (this *Button) LoadArt(modules common.Modules) error {

	settings := modules.Settings()
	mods := modules.Mods()
	render := modules.Render()

	if this.filename == button.NO_FILE {
		return nil
	}

	tmpFilename := this.filename
	if tmpFilename == "" {
		tmpFilename = button.DEFAULT_FILE
	}

	graphics, err := render.LoadImage(settings, mods, tmpFilename) // +1
	if err != nil {
		return err
	}
	defer graphics.UnRef() // -1

	this.buttons, err = graphics.CreateSprite() // +1
	if err != nil {
		return err
	}

	w, err := this.buttons.GetGraphicsWidth()
	if err != nil {
		return err
	}
	this.SetPosW(w)

	h, err := this.buttons.GetGraphicsHeight()
	if err != nil {
		return err
	}

	this.SetPosH(h / 4)                                          // height of one button
	this.buttons.SetClip(0, 0, this.GetPos().W, this.GetPos().H) // 调整精灵的宽高

	return nil
}

func (this *Button) Activate() {
	this.pressed = true
	this.activated = true
}

func (this *Button) Deactivate() {
}

// 修改按钮渲染位置
func (this *Button) SetPos1(modules common.Modules, offsetX, offsetY int) error {

	this.Widget.SetPos1(modules, offsetX, offsetY)
	this.Refresh(modules)

	return nil
}

// 设置标题
func (this *Button) SetLabel(modules common.Modules, s string) {
	this.label = s
	this.Refresh(modules) // 更新文字精灵的位置为在按钮局中

	if this.buttons == nil {
		// 按钮精灵的位置和文字相同，大小也相同
		this.SetPosW(this.wlabel.GetBounds(modules).W)
		this.SetPosH(this.wlabel.GetBounds(modules).H)
		this.Refresh(modules)
		this.SetPos(this.wlabel.GetBounds(modules))
	}
}

// 设置按钮不同状态下的颜色
func (this *Button) SetTextColor(state int, c color.Color) {
	switch state {
	case button.BUTTON_NORMAL:
		this.textColorNormal = c
	case button.BUTTON_PRESSED:
		this.textColorPressed = c
	case button.BUTTON_HOVER:
		this.textColorHover = c
	case button.BUTTON_DISABLED:
		this.textColorDisabled = c
	}
}

// 检查是否按钮被左键点击
func (this *Button) CheckClick(modules common.Modules) bool {
	inpt := modules.Inpt()
	return this.CheckClickAt(modules, inpt.GetMouse().X, inpt.GetMouse().Y)
}

// 检查鼠标位置, 是否要渲染提示文字
func (this *Button) CheckTooltip(modules common.Modules, mouse point.Point) bool {
	settings := modules.Settings()
	inpt := modules.Inpt()
	font := modules.Font()
	tooltipm := modules.Tooltipm()

	// 要在按钮范围内
	if inpt.UsingMouse(settings) && utils.IsWithinRect(this.GetPos(), mouse) && this.tooltip != "" {
		tipData := tooltipdata.Construct()
		tipData.AddColorText(this.tooltip, font.GetColor(fontengine.COLOR_WIDGET_NORMAL))
		newMouse := point.Construct(
			mouse.X+this.GetLocalFrame().X-this.GetLocalOffset().X,
			mouse.Y+this.GetLocalFrame().Y-this.GetLocalOffset().Y,
		)

		// 提前将要渲染的提示文字和鼠标位置压入
		tooltipm.Push(tipData, newMouse, tooltipdata.STYLE_FLOAT, 0)

	}

	return false
}

// 检查是否按钮被左键点击
func (this *Button) CheckClickAt(modules common.Modules, x, y int) bool {
	this.SetEnableTablistNav(this.enabled)
	inpt := modules.Inpt()
	settings := modules.Settings()

	mouse := point.Construct(x, y)
	this.CheckTooltip(modules, mouse) // 是否渲染提示文字

	this.hover = utils.IsWithinRect(this.GetPos(), mouse) && inpt.UsingMouse(settings)

	if !this.enabled {
		return false
	}

	// 是否左键之前已经按下
	if inpt.GetLock(inputstate.MAIN1) {
		return false
	}

	// 回车之前被按下
	if !inpt.UsingMouse(settings) && inpt.GetLock(inputstate.ACCEPT) {
		return false
	}

	// 释放左键时按钮恢复状态
	if this.pressed && !inpt.GetLock(inputstate.MAIN1) &&
		(!inpt.GetLock(inputstate.ACCEPT) || inpt.UsingMouse(settings)) &&
		(utils.IsWithinRect(this.GetPos(), mouse) || this.activated) {

		this.activated = false
		this.pressed = false
		return true
	}

	// 检查是否左键被点击
	this.pressed = false
	if inpt.GetPressing(inputstate.MAIN1) {
		if utils.IsWithinRect(this.GetPos(), mouse) {
			inpt.SetLock(inputstate.MAIN1, true)
			this.pressed = true
		}
	}

	return false
}

func (this *Button) Render(modules common.Modules) error {
	device := modules.Render()

	y := 0
	if !this.enabled {
		y = button.BUTTON_DISABLED * this.GetPos().H
		this.wlabel.SetColor(this.textColorDisabled)
	} else if this.pressed {
		y = button.BUTTON_DISABLED * this.GetPos().H
		y = button.BUTTON_PRESSED * this.GetPos().H
		this.wlabel.SetColor(this.textColorPressed)
	} else if this.hover || this.GetInFocus() {
		y = button.BUTTON_HOVER * this.GetPos().H
		this.wlabel.SetColor(this.textColorHover)
	} else {
		y = button.BUTTON_NORMAL * this.GetPos().H
		this.wlabel.SetColor(this.textColorNormal)
	}

	// 按钮精灵存在
	if this.buttons != nil {
		this.buttons.SetLocalFrame(this.GetLocalFrame())
		this.buttons.SetOffset(this.GetLocalOffset())
		clip := this.buttons.GetClip()
		clip.Y = y
		this.buttons.SetClipFromRect(clip)
		this.buttons.SetDestFromRect(this.GetPos())
		err := device.Render(this.buttons)
		if err != nil {
			return err
		}
	}

	// 渲染文字精灵
	this.wlabel.SetLocalFrame(this.GetLocalFrame())
	this.wlabel.SetLocalOffset(this.GetLocalOffset())
	this.wlabel.Render(modules)

	return nil
}

// 更新标题文字
func (this *Button) Refresh(modules common.Modules) {
	font := modules.Font()

	if this.label != "" {
		if this.buttons != nil {
			// 按钮精灵存在则文字居中显示
			this.wlabel.SetPos1(modules, this.GetPos().X+this.GetPos().W/2, this.GetPos().Y+this.GetPos().H/2) // 设置显示位置
			this.wlabel.SetJustify(fontengine.JUSTIFY_CENTER)
			this.wlabel.SetValign(labelinfo.VALIGN_CENTER)
		} else {
			// 文字左对齐显示
			this.wlabel.SetPos1(modules, this.GetPos().X, this.GetPos().Y) // 设置显示位置
			this.wlabel.SetJustify(fontengine.JUSTIFY_LEFT)
			this.wlabel.SetValign(labelinfo.VALIGN_TOP)
		}

		this.wlabel.SetText(this.label)

		if this.enabled {
			this.wlabel.SetColor(font.GetColor(fontengine.COLOR_WIDGET_NORMAL))
		} else {
			this.wlabel.SetColor(font.GetColor(fontengine.COLOR_WIDGET_DISABLED))
		}
	}
}

func (this *Button) GetEnabled() bool {
	return this.enabled
}

func (this *Button) SetEnabled(val bool) {
	this.enabled = val
}

func (this *Button) SetTooltip(val string) {
	this.tooltip = val
}

func (this *Button) GetNext(modules common.Modules) bool {
	return false
}

func (this *Button) GetPrev(modules common.Modules) bool {
	return false
}
