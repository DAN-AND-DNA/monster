package widget

import (
	"math"
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define/inputstate"
	"monster/pkg/common/define/widget/scrollbar"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/utils"
	"monster/pkg/widget/base"
)

// 滚动条
type ScrollBar struct {
	base.Widget
	filename    string
	scrollbars  common.Sprite
	value       int // 滚动条的值
	barHeight   int // 滚动条的总高
	maximum     int
	lockMain1   bool // 鼠标左键按下
	dragging    bool // 拖拽
	bg          common.Sprite
	upToKnob    rect.Rect // 向上按钮---滑块
	knobToDown  rect.Rect // 按钮---向下按钮
	posUp       rect.Rect // 向上按钮的位置
	posDown     rect.Rect // 向下按钮的位置
	posKnob     rect.Rect // 滑块当前位置
	pressedUp   bool
	pressedDown bool
	pressedKnob bool
}

func NewScrollBar(modules common.Modules, filename string) *ScrollBar {
	sb := &ScrollBar{}
	sb.Init(modules, filename)

	return sb
}

func (this *ScrollBar) Init(modules common.Modules, filename string) common.WidgetScrollBar {
	settings := modules.Settings()
	mods := modules.Mods()
	render := modules.Render()

	// base
	this.Widget = base.ConstructWidget()

	// me
	this.filename = filename

	this.LoadArt(settings, mods, render)
	var err error

	// 设置长宽
	this.posUp.W, err = this.scrollbars.GetGraphicsWidth()
	if err != nil {
		panic(err)
	}
	this.posDown.W = this.posUp.W
	this.posKnob.W = this.posUp.W

	this.posUp.H, err = this.scrollbars.GetGraphicsHeight()
	if err != nil {
		panic(err)
	}

	this.posUp.H = this.posUp.H / 5
	this.posDown.H = this.posUp.H
	this.posKnob.H = this.posUp.H

	return this
}

func (this *ScrollBar) Close() {
	this.Widget.Close(this)
}

func (this *ScrollBar) Clear() {
	if this.scrollbars != nil {
		this.scrollbars.Close()
	}

	if this.bg != nil {
		this.bg.Close()
	}
}

func (this *ScrollBar) LoadArt(settings common.Settings, mods common.ModManager, device common.RenderDevice) error {
	tmpFilename := scrollbar.DEFAULT_FILE
	if this.filename != "" {
		tmpFilename = this.filename
	}

	graphics, err := device.LoadImage(settings, mods, tmpFilename)
	if err != nil {
		return err
	}
	defer graphics.UnRef()

	this.scrollbars, err = graphics.CreateSprite()
	if err != nil {
		return err
	}

	return nil
}

func (this *ScrollBar) CheckClick(modules common.Modules) int {
	inpt := modules.Inpt()

	return this.CheckClickAt(modules, inpt.GetMouse().X, inpt.GetMouse().Y)
}

func (this *ScrollBar) CheckClickAt(modules common.Modules, x, y int) int {
	inpt := modules.Inpt()

	mouse := point.Construct(x, y)

	inBounds := utils.IsWithinRect(this.GetBounds(), mouse)                                         // 自己的范围内
	inUp := utils.IsWithinRect(this.posUp, mouse) || utils.IsWithinRect(this.upToKnob, mouse)       // 上
	inDown := utils.IsWithinRect(this.posDown, mouse) || utils.IsWithinRect(this.knobToDown, mouse) //下
	inKnob := utils.IsWithinRect(this.posKnob, mouse)                                               // 滑块位置

	// 范围内且之前的左键已释放或拖拽
	if inBounds && (!this.lockMain1 || this.dragging) {
		this.lockMain1 = false
		this.dragging = false

		// 处理按左键
		if inpt.GetPressing(inputstate.MAIN1) {
			inpt.SetLock(inputstate.MAIN1, true)
			if inUp && !this.pressedKnob {
				this.pressedUp = true
			} else if inDown && !this.pressedKnob {
				this.pressedDown = true
			} else if inKnob && !this.pressedUp && !this.pressedDown {
				this.pressedKnob = true // 按滑块
				this.dragging = true    // 准备进行拖拽
			} else if inBounds && this.pressedKnob {
				this.dragging = true // 保持拖拽
			}
		}
	} else {
		// 不再范围内或左键仍然按着
		this.lockMain1 = inpt.GetPressing(inputstate.MAIN1)
	}

	ret := scrollbar.NONE

	// 处理释放左键
	if this.pressedUp && !inpt.GetPressing(inputstate.MAIN1) {
		this.pressedUp = false
		if inUp {
			ret = scrollbar.UP
		}
	} else if this.pressedDown && !inpt.GetPressing(inputstate.MAIN1) {
		this.pressedDown = false
		if inDown {
			ret = scrollbar.DOWN
		}
	} else if this.pressedKnob && this.dragging {
		tmp := mouse.Y - this.posUp.Y - this.posUp.H
		if this.barHeight < 1 {
			this.barHeight = 1
		}

		// 计算滑块的值
		this.value = (tmp * this.maximum) / this.barHeight

		// 修改滑块位置
		this.Set()

		ret = scrollbar.DRAGGING
	}

	// 释放左键清理状态
	if !inpt.GetPressing(inputstate.MAIN1) {
		this.dragging = false
		this.pressedKnob = false
		this.pressedUp = false
		this.pressedDown = false
	}

	return ret

}

func (this *ScrollBar) Set() {
	if this.maximum < 1 {
		this.maximum = 1
	}
	/*
		|v| down
		| |
		| |
		|_|
		|_| knob
		| |
		| |
		| |
		| |
		|^| up
	*/

	// 计算滑块的范围
	this.value = (int)(math.Max(0, math.Min((float64)(this.maximum), (float64)(this.value))))

	// 更新滑块的位置
	this.posKnob.Y = this.posUp.Y + this.posUp.H + (this.value * (this.barHeight - this.posUp.H) / this.maximum)

	this.upToKnob.X = this.posKnob.X
	this.knobToDown.X = this.posKnob.X

	this.upToKnob.W = this.posKnob.W
	this.knobToDown.W = this.posKnob.W

	this.upToKnob.Y = this.posUp.Y + this.posUp.H
	this.upToKnob.H = this.posKnob.Y - this.upToKnob.Y

	this.knobToDown.Y = this.posKnob.Y + this.posKnob.H
	this.knobToDown.H = this.posDown.Y - this.knobToDown.Y
}

func (this *ScrollBar) GetValue() int {
	return this.value
}

func (this *ScrollBar) GetBounds() rect.Rect {
	r := rect.Construct()
	r.X = this.posUp.X
	r.Y = this.posUp.Y
	r.W = this.posUp.W
	r.H = this.posUp.H*2 + this.barHeight

	return r
}

func (this *ScrollBar) Render(modules common.Modules) error {
	device := modules.Render()
	srcUp := rect.Construct()
	srcDown := rect.Construct()
	srcKnob := rect.Construct()

	// 裁剪图片中的位置
	srcUp.X = 0
	if this.pressedUp {
		srcUp.Y = this.posUp.H
	} else {
		srcUp.Y = 0
	}
	srcUp.W = this.posUp.W
	srcUp.H = this.posUp.H

	// 裁剪图片中的位置
	srcDown.X = 0
	if this.pressedDown {
		srcDown.Y = this.posDown.H * 3
	} else {
		srcDown.Y = this.posDown.H * 2
	}
	srcDown.W = this.posDown.W
	srcDown.H = this.posDown.H

	srcKnob.X = 0
	srcKnob.Y = this.posKnob.H * 4
	srcKnob.W = this.posKnob.W
	srcKnob.H = this.posKnob.H

	if this.bg != nil {
		this.bg.SetLocalFrame(this.GetLocalFrame())
		this.bg.SetOffset(this.GetLocalOffset())
		this.bg.SetDestFromRect(this.posUp)
		err := device.Render(this.bg)
		if err != nil {
			return err
		}
	}

	if this.scrollbars != nil {
		this.scrollbars.SetLocalFrame(this.GetLocalFrame())
		this.scrollbars.SetOffset(this.GetLocalOffset())

		// 绘制各部分
		this.scrollbars.SetClipFromRect(srcUp)
		this.scrollbars.SetDestFromRect(this.posUp)
		err := device.Render(this.scrollbars)
		if err != nil {
			return err
		}

		this.scrollbars.SetClipFromRect(srcDown)
		this.scrollbars.SetDestFromRect(this.posDown)
		err = device.Render(this.scrollbars)
		if err != nil {
			return err
		}

		this.scrollbars.SetClipFromRect(srcKnob)
		this.scrollbars.SetDestFromRect(this.posKnob)
		err = device.Render(this.scrollbars)
		if err != nil {
			return err
		}
	}

	return nil
}

// 滚动条更新位置和滚动值 h为按钮之间的距离，下按钮到上按钮的差，y为向上按钮的左上y点
func (this *ScrollBar) Refresh(modules common.Modules, x, y, h, val, max int) error {
	eset := modules.Eset()
	device := modules.Render()

	before := this.GetBounds()
	this.maximum = max
	this.value = val
	this.posUp.X = x
	this.posDown.X = x
	this.posKnob.X = x
	this.posUp.Y = y
	this.posDown.Y = y + h
	this.barHeight = this.posDown.Y - (this.posUp.Y + this.posUp.H)
	this.Set()

	after := this.GetBounds()

	if before.H != after.H {
		if this.bg != nil {
			this.bg.Close()
			this.bg = nil
		}

		graphics, err := device.CreateImage(after.W, after.H)
		if err != nil {
			return err
		}
		defer graphics.UnRef()
		this.bg, err = graphics.CreateSprite()
		if err != nil {
			return err
		}

		err = this.bg.GetGraphics().FillWithColor(eset.Get("widgets", "bg_color").(color.Color))
		if err != nil {
			return err
		}

	}

	return nil
}

func (this *ScrollBar) PosDown() rect.Rect {
	return this.posDown
}

func (this *ScrollBar) Activate() {
}

func (this *ScrollBar) Deactivate() {
}

func (this *ScrollBar) GetNext(common.Modules) bool {
	return false
}

func (this *ScrollBar) GetPrev(common.Modules) bool {
	return false
}

// 向上按钮的位置和大小
func (this *ScrollBar) GetPosUp() rect.Rect {
	return this.posUp
}

// 向下按钮的位置和大小
func (this *ScrollBar) GetPosDown() rect.Rect {
	return this.posDown
}

// 滑块的位置和大小
func (this *ScrollBar) GetPosKnob() rect.Rect {
	return this.posKnob
}
