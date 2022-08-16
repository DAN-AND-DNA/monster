package widget

import (
	"math"
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define/fontengine"
	"monster/pkg/common/define/widget"
	"monster/pkg/common/define/widget/button"
	"monster/pkg/common/define/widget/horizontallist"
	"monster/pkg/common/labelinfo"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/common/tooltipdata"
	"monster/pkg/utils"
	"monster/pkg/widget/base"
)

type HListItem struct {
	value   string
	tooltip string
}

type HorizontalList struct {
	base.Widget
	label               *Label
	buttonLeft          *Button
	buttonRight         *Button
	buttonAction        *Button
	cursor              int
	changedWithoutMouse bool
	actionTriggered     bool
	activated           bool
	listItems           []HListItem
	tooltipArea         rect.Rect
	enabled             bool
	hasAction           bool
}

func NewHorizontalList(modules common.Modules) *HorizontalList {
	hl := &HorizontalList{}
	hl.Init(modules)

	return hl
}

func (this *HorizontalList) Init(modules common.Modules) common.WidgetHorizontalList {
	// base
	this.Widget = base.ConstructWidget()

	// self
	this.label = NewLabel(modules)
	this.buttonLeft = NewButton(modules, horizontallist.DEFAULT_FILE_LEFT)
	this.buttonRight = NewButton(modules, horizontallist.DEFAULT_FILE_RIGHT)
	this.buttonAction = NewButton(modules, button.DEFAULT_FILE)
	this.enabled = true
	this.SetFocusable(true)
	this.SetScrollType(widget.SCROLL_HORIZONTAL)

	// 更新位置和大小
	this.Refresh(modules)

	return this
}

func (this *HorizontalList) Append(value, tooltip string) {
	hli := HListItem{value: value, tooltip: tooltip}

	this.listItems = append(this.listItems, hli)
}

func (this *HorizontalList) Clear1() {
	this.listItems = nil
}

func (this *HorizontalList) Clear() {
	this.Clear1()

	if this.label != nil {
		this.label.Close()
		this.label = nil
	}

	if this.buttonLeft != nil {
		this.buttonLeft.Close()
		this.buttonLeft = nil
	}

	if this.buttonRight != nil {
		this.buttonRight.Close()
		this.buttonRight = nil
	}

	if this.buttonAction != nil {
		this.buttonAction.Close()
		this.buttonAction = nil
	}

}

func (this *HorizontalList) Close() {
	this.Widget.Close(this)
}

func (this *HorizontalList) Activate() {
	this.activated = true
}

func (this *HorizontalList) CheckAction() bool {
	if this.actionTriggered {
		this.actionTriggered = false
		return true
	}

	return false
}

// 设置位置
func (this *HorizontalList) SetPos1(modules common.Modules, offsetX, offsetY int) error {
	this.Widget.SetPos1(modules, offsetX, offsetY)
	this.Refresh(modules)

	return nil
}

func (this *HorizontalList) CheckTooltip(modules common.Modules, mouse point.Point) {
	inpt := modules.Inpt()
	settings := modules.Settings()
	font := modules.Font()
	tooltipm := modules.Tooltipm()

	// 按钮处理提示
	if this.hasAction {
		return
	}

	if this.IsEmpty() {
		return
	}

	if inpt.UsingMouse(settings) && utils.IsWithinRect(this.tooltipArea, mouse) && this.listItems[this.cursor].tooltip != "" {
		tipData := tooltipdata.Construct()
		tipData.AddColorText(this.listItems[this.cursor].tooltip, font.GetColor(fontengine.COLOR_WIDGET_NORMAL))
		newMouse := point.Construct(mouse.X+this.GetLocalFrame().X-this.GetLocalOffset().X,
			mouse.Y+this.GetLocalFrame().Y-this.GetLocalOffset().Y)
		tooltipm.Push(tipData, newMouse, tooltipdata.STYLE_FLOAT, 0)
	}
}

func (this *HorizontalList) ScrollLeft(modules common.Modules) {
	if this.IsEmpty() {
		return
	}

	if this.cursor == 0 {
		this.cursor = this.GetSize() - 1
	} else {
		this.cursor--
	}

	// 更新大小和位置
	this.Refresh(modules)
}

func (this *HorizontalList) ScrollRight(modules common.Modules) {
	if this.IsEmpty() {
		return
	}

	if this.cursor+1 >= this.GetSize() {
		this.cursor = 0
	} else {
		this.cursor++
	}

	// 更新大小和位置
	this.Refresh(modules)
}

func (this *HorizontalList) CheckClickAt(modules common.Modules, x, y int) bool {
	mouse := point.Construct(x, y)

	// 处理提示
	this.CheckTooltip(modules, mouse)

	// 处理各按钮
	if this.buttonLeft.CheckClickAt(modules, mouse.X, mouse.Y) {
		this.ScrollLeft(modules)
		return true
	} else if this.buttonRight.CheckClickAt(modules, mouse.X, mouse.Y) {
		this.ScrollRight(modules)
		return true
	} else if this.changedWithoutMouse {
		this.changedWithoutMouse = false
		return true
	} else if this.hasAction && this.buttonAction.CheckClickAt(modules, mouse.X, mouse.Y) {
		this.actionTriggered = true
		return true
	} else if this.hasAction && this.activated {
		this.activated = false
		this.buttonAction.Activate()
	}

	return false
}

func (this *HorizontalList) CheckClick(modules common.Modules) bool {
	inpt := modules.Inpt()
	mouse := inpt.GetMouse()
	return this.CheckClickAt(modules, mouse.X, mouse.Y)
}

func (this *HorizontalList) IsEmpty() bool {
	return len(this.listItems) == 0
}

func (this *HorizontalList) GetSelected() (int, bool) {
	if this.cursor < this.GetSize() {
		return this.cursor, true
	}

	return this.GetSize(), false
}

func (this *HorizontalList) Select(modules common.Modules, index int) {
	if this.IsEmpty() {
		return
	}

	if index < this.GetSize() {
		this.cursor = index
	}

	this.Refresh(modules)
}

func (this *HorizontalList) SetHasAction(val bool) {
	this.hasAction = val
}

func (this *HorizontalList) GetSize() int {
	return len(this.listItems)
}

func (this *HorizontalList) GetValue() string {
	if this.cursor < this.GetSize() {
		return this.listItems[this.cursor].value
	}

	return ""
}

func (this *HorizontalList) SetValue(index int, value string) {
	if index < this.GetSize() {
		this.listItems[index].value = value
	}
}

// 更新位置和大小
func (this *HorizontalList) Refresh(modules common.Modules) {
	eset := modules.Eset()
	font := modules.Font()

	contentWidth := eset.Get("widgets", "text_width").(int) // 显示文字的宽度
	isEnabled := !this.IsEmpty() && this.enabled

	this.buttonLeft.SetEnabled(isEnabled)
	this.buttonRight.SetEnabled(isEnabled)
	this.buttonAction.SetEnabled(isEnabled)

	this.buttonLeft.SetPos1(modules, this.GetPos().X, this.GetPos().Y)

	if this.hasAction {
		contentWidth = this.buttonAction.GetPos().W
		this.buttonAction.SetPos1(modules, this.GetPos().X+this.buttonLeft.GetPos().W, this.GetPos().Y+this.buttonLeft.GetPos().H/2-this.buttonAction.GetPos().H/2)
		this.buttonAction.SetLabel(modules, this.GetValue())
		if this.cursor < this.GetSize() {
			this.buttonAction.SetTooltip(this.listItems[this.cursor].tooltip)
		}
		this.SetPosH((int)(math.Max((float64)(this.buttonLeft.GetPos().H), (float64)(this.buttonAction.GetPos().H))))
	} else {
		this.label.SetText(this.GetValue())

		// label位置在左右按钮的之间居中
		this.label.SetPos1(modules, this.GetPos().X+this.buttonLeft.GetPos().W+contentWidth/2, this.GetPos().Y+this.buttonLeft.GetPos().H/2)
		this.label.SetMaxWidth(contentWidth)
		this.label.SetJustify(fontengine.JUSTIFY_CENTER) // 文字水平居中
		this.label.SetValign(labelinfo.VALIGN_CENTER)    // 文字垂直居中
		if isEnabled {
			this.label.SetColor(font.GetColor(fontengine.COLOR_WIDGET_NORMAL))
		} else {
			this.label.SetColor(font.GetColor(fontengine.COLOR_WIDGET_DISABLED))
		}

		// 组件高度更新为最大的一个
		this.SetPosH((int)(math.Max((float64)(this.buttonLeft.GetPos().H), (float64)(this.label.GetBounds(modules).H))))

		// 提示文字范围
		this.tooltipArea.X = this.GetPos().X + this.buttonLeft.GetPos().W
		this.tooltipArea.Y = (int)(math.Min((float64)(this.GetPos().Y), (float64)(this.label.GetBounds(modules).Y)))
		this.tooltipArea.W = contentWidth
		this.tooltipArea.H = (int)(math.Max((float64)(this.buttonLeft.GetPos().H), (float64)(this.label.GetBounds(modules).H)))
	}

	// 右按钮
	this.buttonRight.SetPos1(modules, this.GetPos().X+this.buttonLeft.GetPos().W+contentWidth, this.GetPos().Y)
	// 组件的总宽 = 左按钮宽 + 文字宽 + 右按钮宽
	this.SetPosW(this.buttonLeft.GetPos().W + this.buttonRight.GetPos().W + contentWidth)
}

func (this *HorizontalList) Deactivate() {
}

func (this *HorizontalList) GetNext(modules common.Modules) bool {
	if !this.IsEmpty() && this.enabled {
		this.ScrollRight(modules)
		this.changedWithoutMouse = true
	}

	return true

}

func (this *HorizontalList) GetPrev(modules common.Modules) bool {
	if !this.IsEmpty() && this.enabled {
		this.ScrollLeft(modules)
		this.changedWithoutMouse = true
	}

	return true
}

func (this *HorizontalList) Render(modules common.Modules) error {
	render := modules.Render()
	eset := modules.Eset()

	this.buttonLeft.SetLocalFrame(this.GetLocalFrame())
	this.buttonLeft.SetLocalOffset(this.GetLocalOffset())
	this.buttonRight.SetLocalFrame(this.GetLocalFrame())
	this.buttonRight.SetLocalOffset(this.GetLocalOffset())
	err := this.buttonLeft.Render(modules)
	if err != nil {
		return err
	}

	err = this.buttonRight.Render(modules)
	if err != nil {
		return err
	}

	if this.hasAction {
		this.buttonAction.SetLocalFrame(this.GetLocalFrame())
		this.buttonAction.SetLocalOffset(this.GetLocalOffset())
		this.buttonAction.Render(modules)
	} else {
		this.label.SetLocalFrame(this.GetLocalFrame())
		this.label.SetLocalOffset(this.GetLocalOffset())
		this.label.Render(modules)
	}

	if this.GetInFocus() {
		// 选择框
		pos := this.GetPos()
		topLeft := point.Construct()
		bottomRight := point.Construct()

		topLeft.X = pos.X + this.GetLocalFrame().X - this.GetLocalOffset().X
		topLeft.Y = pos.Y + this.GetLocalFrame().Y - this.GetLocalOffset().Y
		bottomRight.X = topLeft.X + pos.X
		bottomRight.Y = topLeft.Y + pos.H

		draw := true

		// 需要在父级范围内
		if this.GetLocalFrame().W > 0 &&
			(topLeft.X < this.GetLocalFrame().X || bottomRight.X > this.GetLocalFrame().X+this.GetLocalFrame().W) {
			draw = false
		}

		if this.GetLocalFrame().H > 0 &&
			(topLeft.Y < this.GetLocalFrame().Y || bottomRight.Y > this.GetLocalFrame().Y+this.GetLocalFrame().H) {
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
