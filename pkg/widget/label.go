package widget

import (
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define/fontengine"
	"monster/pkg/common/define/widget/label"
	"monster/pkg/common/labelinfo"
	"monster/pkg/common/rect"
	"monster/pkg/widget/base"
)

// 表示文字
type Label struct {
	base.Widget

	justify          int // 水平对齐
	valign           int // 垂向对齐
	maxWidth         int // 最大行像素宽
	updateFlag       int
	hidden           bool
	windowResizeFlag bool
	alpha            uint8
	label            common.Sprite // 展现内容
	text             string
	fontStyle        string
	color            color.Color
	bounds           rect.Rect // 内容大小也即精灵的大小
}

func NewLabel(modules common.Modules) *Label {
	l := &Label{}

	l.Init(modules)
	return l
}

func (this *Label) Init(modules common.Modules) common.WidgetLabel {
	font := modules.Font()
	return this.Init1(font)
}

func (this *Label) Init1(font common.FontEngine) common.WidgetLabel {
	// base
	this.Widget = base.ConstructWidget()

	// 自己
	this.justify = fontengine.JUSTIFY_LEFT
	this.valign = labelinfo.VALIGN_TOP
	this.updateFlag = label.UPDATE_NONE
	this.alpha = 255
	this.fontStyle = "font_regular"
	this.color = font.GetColor(fontengine.COLOR_WIDGET_NORMAL)

	return this
}

func (this *Label) Clear() {
	if this.label != nil {
		this.label.Close()
		this.label = nil
	}
}

func (this *Label) Close() {
	this.Widget.Close(this)
}

func (this *Label) SetMaxWidth(width int) {
	if width != this.maxWidth {
		this.maxWidth = width
		this.SetUpdateFlag(label.UPDATE_RECACHE)
	}
}

//设置渲染位置
func (this *Label) SetPos1(modules common.Modules, offsetX, offsetY int) error {
	oldPos := this.GetPos()
	this.Widget.SetPos1(modules, offsetX, offsetY)

	if oldPos.X != this.GetPos().X || oldPos.Y != this.GetPos().Y {
		this.SetUpdateFlag(label.UPDATE_POS)
	}

	return nil
}

// 设置水平对齐
func (this *Label) SetJustify(justify int) {
	if this.justify != justify {
		this.justify = justify
		this.SetUpdateFlag(label.UPDATE_RECACHE) // 重建精灵
	}
}

func (this *Label) SetText(text string) {
	if this.text != text {
		this.text = text
		this.SetUpdateFlag(label.UPDATE_RECACHE)
	}
}

func (this *Label) SetValign(valign int) {
	if this.valign != valign {
		this.valign = valign
		this.SetUpdateFlag(label.UPDATE_RECACHE)
	}
}

func (this *Label) SetColor(color color.Color) {
	if this.color.R != color.R || this.color.G != color.G || this.color.B != this.color.B {
		this.color = color
		this.SetUpdateFlag(label.UPDATE_RECACHE)
	}
}

func (this *Label) SetFont(font string) {
	if this.fontStyle != font {
		this.fontStyle = font
		this.SetUpdateFlag(label.UPDATE_RECACHE)
	}
}

func (this *Label) SetAlpha(alpha uint8) {
	if this.alpha != alpha {
		this.alpha = alpha
		this.SetUpdateFlag(label.UPDATE_POS)
	}
}

func (this *Label) SetFromLabelInfo(labelInfo labelinfo.LabelInfo) {
	if this.GetPosBase().X != labelInfo.X || this.GetPosBase().Y != labelInfo.Y {
		this.SetUpdateFlag(label.UPDATE_POS)
	}

	this.SetPosBase(labelInfo.X, labelInfo.Y, this.GetAlignment())
	this.SetJustify(labelInfo.Justify)
	this.SetValign(labelInfo.Valign)
	this.SetFont(labelInfo.FontStyle)
	this.SetHidden(labelInfo.Hidden)
}

func (this *Label) GetText() string {
	return this.text
}

// 更新精灵的大小和位置，重建精灵缓存
func (this *Label) GetBounds(modules common.Modules) rect.Rect {
	this.Update(modules)

	return this.bounds
}

func (this *Label) IsHidden() bool {
	return this.hidden
}

func (this *Label) SetHidden(isHidden bool) {
	this.hidden = isHidden
}

// 修正显示位置
func (this *Label) ApplyOffsets() {

	// 水平对齐
	switch this.justify {
	case fontengine.JUSTIFY_LEFT:
		this.bounds.X = this.GetPos().X
	case fontengine.JUSTIFY_RIGHT:
		this.bounds.X = this.GetPos().X - this.bounds.W
	case fontengine.JUSTIFY_CENTER:
		this.bounds.X = this.GetPos().X - this.bounds.W/2
	}

	// 垂直对齐
	switch this.valign {
	case labelinfo.VALIGN_TOP:
		this.bounds.Y = this.GetPos().Y
	case labelinfo.VALIGN_BOTTOM:
		this.bounds.Y = this.GetPos().Y - this.bounds.H
	case labelinfo.VALIGN_CENTER:
		this.bounds.Y = this.GetPos().Y - this.bounds.H/2
	}

	if this.label != nil {
		this.label.SetDestFromRect(this.bounds) // 修正渲染位置
		this.label.SetAlphaMod(this.alpha)
	}
}

// 把文字渲染到空的图片上，保留该图片的精灵，相当于缓存了内容，更新精灵大小
func (this *Label) RecacheTextSprite(device common.RenderDevice, font common.FontEngine) error {
	if this.label != nil {
		this.label.Close()
		this.label = nil
	}

	if this.text == "" {
		this.bounds.W = 0
		this.bounds.H = 0
		return nil
	}

	// 设置内容大小
	tempText := this.text
	font.SetFont(this.fontStyle)
	this.bounds.W = font.CalcWidth(tempText)
	this.bounds.H = font.GetFontHeight()

	if this.maxWidth > 0 && this.bounds.W > this.maxWidth {
		tempText = font.TrimTextToWidth(this.text, this.maxWidth, true, 0)
		this.bounds.W = font.CalcWidth(tempText)
	}

	// 目标图片
	image, err := device.CreateImage(this.bounds.W, this.bounds.H) // +1
	if err != nil {
		return err
	}
	defer image.UnRef() // -1

	// 把文字渲染上去，对图片的修改内容都会保存
	err = font.RenderShadowed(device, tempText, 0, 0, fontengine.JUSTIFY_LEFT, image, 0, this.color)
	if err != nil {
		return err
	}

	// 创建精灵
	this.label, err = image.CreateSprite() // +1
	if err != nil {
		return err
	}

	return nil
}

func (this *Label) SetUpdateFlag(updateFlag int) {
	if updateFlag > this.updateFlag || updateFlag == label.UPDATE_NONE {
		this.updateFlag = updateFlag
	}
}

// 更新精灵的大小和内容的大小，根据对齐方式修改文字的显示位置
func (this *Label) Update(modules common.Modules) {
	inpt := modules.Inpt()
	device := modules.Render()
	font := modules.Font()

	if inpt.GetWindowResized() && !this.windowResizeFlag {
		this.SetUpdateFlag(label.UPDATE_RECACHE)
		this.windowResizeFlag = true
	}

	if this.updateFlag == label.UPDATE_RECACHE {
		this.RecacheTextSprite(device, font) // 重建缓存，更新精灵的大小和位置
	}

	if this.updateFlag >= label.UPDATE_POS {
		this.ApplyOffsets() // 修改显示位置
	}

	this.SetUpdateFlag(label.UPDATE_NONE)
}

func (this *Label) Render(modules common.Modules) error {
	render := modules.Render()

	if this.hidden {
		return nil
	}
	this.Update(modules)
	if this.label != nil {
		this.label.SetLocalFrame(this.GetLocalFrame())
		this.label.SetOffset(this.GetLocalOffset())
		err := render.Render(this.label)
		if err != nil {
			return err
		}
	}

	this.windowResizeFlag = false
	return nil
}

func (this *Label) Activate() {
}

func (this *Label) Deactivate() {
}

func (this *Label) GetNext(common.Modules) bool {
	return false
}

func (this *Label) GetPrev(common.Modules) bool {
	return false
}
