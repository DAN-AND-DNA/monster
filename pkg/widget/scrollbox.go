package widget

import (
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define/widget"
	"monster/pkg/common/define/widget/scrollbar"
	"monster/pkg/common/define/widget/scrollbox"
	"monster/pkg/common/define/widget/tablist"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/utils"
	"monster/pkg/widget/base"
)

// 带滚动条的弹窗
type ScrollBox struct {
	base.Widget
	contents       common.Sprite // 空白精灵，作为绘制目标，上面一堆内容，高可以大于弹窗本身
	update         bool
	bg             color.Color
	tablist        *Tablist              // 主要处理键盘导致的组件跳转 内部是其他组件的指针，不负责创建和清理
	children       map[int]common.Widget // 其他组件的指针，不负责创建和清理
	currentChild   int
	currentChildOk bool
	cursor         float32 // 当前游标位置
	cursorTarget   float32 // 目标游标位置
	scrollbar      *ScrollBar
}

// 指定宽高的滚动窗口
func NewScrollBox(modules common.Modules, width, height int) *ScrollBox {

	sb := &ScrollBox{}
	sb.Init(modules, width, height)

	return sb
}

func (this *ScrollBox) Init(modules common.Modules, width, height int) common.WidgetScrollBox {
	// base
	this.Widget = base.ConstructWidget()

	// self
	this.children = map[int]common.Widget{}
	this.tablist = NewTablist()
	this.update = true
	this.bg = color.Construct(0, 0, 0, 0) //透明
	this.scrollbar = NewScrollBar(modules, scrollbar.DEFAULT_FILE)
	this.SetPosW(width) // 指定弹窗的大小
	this.SetPosH(height)
	this.SetScrollType(widget.SCROLL_VERTICAL) // 垂直

	// 调整绘制目标精灵的大小和位置 (高度大于等于弹窗本身)
	if err := this.Resize(modules, width, height); err != nil {
		panic(err)
	}

	this.tablist.SetScrollType(widget.SCROLL_TWO_DIRECTIONS)
	return this
}

func (this *ScrollBox) Close() {
	this.Widget.Close(this)
}

func (this *ScrollBox) Clear() {
	if this.contents != nil {
		this.contents.Close()
		this.contents = nil
	}

	if this.scrollbar != nil {
		this.scrollbar.Close()
		this.scrollbar = nil
	}

	if this.tablist != nil {
		this.tablist.Close()
		this.tablist = nil
	}
}

// 设置绘制位置
func (this *ScrollBox) SetPos1(modules common.Modules, offsetX, offsetY int) error {
	// 设置渲染位置
	this.Widget.SetPos1(modules, offsetX, offsetY)

	if this.contents != nil && this.scrollbar != nil {
		h, err := this.contents.GetGraphicsHeight()
		if err != nil {
			return err
		}

		//更新滚动条的位置和大小
		pos := this.GetPos()
		err = this.scrollbar.Refresh(modules, pos.X+pos.W, pos.Y, pos.H-this.scrollbar.PosDown().H, (int)(this.cursor), h-pos.H)
		if err != nil {
			return err
		}
	}

	return nil
}

// 根据弹窗的宽高来更新绘制目标精灵和滚动条的大小
func (this *ScrollBox) Refresh(modules common.Modules) error {
	device := modules.Render()

	if this.update {
		h := this.GetPos().H
		var err error
		if this.contents != nil {
			h, err = this.contents.GetGraphicsHeight()
			if err != nil {
				return err
			}
			this.contents.Close()
			this.contents = nil
		}

		graphics, err := device.CreateImage(this.GetPos().W, h)
		if err != nil {
			return err
		}
		defer graphics.UnRef()

		this.contents, err = graphics.CreateSprite()
		if err != nil {
			return err
		}

		err = this.contents.GetGraphics().FillWithColor(this.bg)
		if err != nil {
			return err
		}
	}

	if this.contents != nil && this.scrollbar != nil {
		gh, err := this.contents.GetGraphicsHeight()
		if err != nil {
			return err
		}

		// 更新滚动条位置
		pos := this.GetPos()

		// 滚动条的值为 cursorTarget
		// 滚动条的最大值为 渲染内容-当前已经展现的部分高度
		err = this.scrollbar.Refresh(modules, pos.X+pos.W, pos.Y, pos.H-this.scrollbar.PosDown().H, (int)(this.cursorTarget), gh-pos.H)
		if err != nil {
			return err
		}
	}

	return nil
}

// 调整绘制目标精灵的宽高，高度大于等于弹窗
func (this *ScrollBox) Resize(modules common.Modules, w, h int) error {
	render := modules.Render()

	this.SetPosW(w)
	if this.GetPos().H > h {
		h = this.GetPos().H
	}

	if this.contents != nil {
		this.contents.Close()
		this.contents = nil
	}

	// 创建空的图片
	graphics, err := render.CreateImage(this.GetPos().W, h)
	if err != nil {
		return err
	}
	defer graphics.Close()

	this.contents, err = graphics.CreateSprite()
	if err != nil {
		return err
	}

	// 目标精灵颜色
	err = this.contents.GetGraphics().FillWithColor(this.bg)
	if err != nil {
		return err
	}

	this.cursor = 0
	this.cursorTarget = 0
	err = this.Refresh(modules)
	if err != nil {
		return err
	}

	return nil
}

// 添加子组件
func (this *ScrollBox) AddChildWidget(child common.Widget) {
	found := false
	for _, ptr := range this.children {
		if child == ptr {
			found = true
		}
	}

	if found {
		return
	}

	this.children[len(this.children)] = child
	this.tablist.Add(child)
	child.SetLocalFrame(this.GetPos()) // 设置父级的位置和大小
}

func (this *ScrollBox) ClearChildWidget() {
	this.currentChildOk = false
	this.children = map[int]common.Widget{}
	this.tablist.Clear()
}

// 滚动增量为amount
func (this *ScrollBox) Scroll(modules common.Modules, amount int) error {
	this.cursorTarget += (float32)(amount)

	// 获得渲染目标的高度，包含大量内容和组件，依靠滚动条展现
	gh, err := this.contents.GetGraphicsHeight()
	if err != nil {
		return err
	}

	if this.cursorTarget < 0 {
		this.cursorTarget = 0

	} else if this.contents != nil && this.cursorTarget > float32(gh-this.GetPos().H) {
		this.cursorTarget = float32(gh - this.GetPos().H)
	}

	// 更新全部的大小和位置
	return this.Refresh(modules)
}

// 滚动的目标位置为amount
func (this *ScrollBox) ScrollTo(modules common.Modules, amount int) error {
	this.cursor = (float32)(amount) // 更新增量

	h, err := this.contents.GetGraphicsHeight()
	if err != nil {
		return err
	}

	/*
		游标从0到H-h
		0	 _______
			|_______|  \   弹窗高h
			|	|   \
			|	|      总渲染高H
		H-h	|_______|   /
			|_______|  /

	*/

	if this.cursor < 0 {
		this.cursor = 0

	} else if this.contents != nil && this.cursor > float32(h-this.GetPos().H) {
		this.cursor = float32(h - this.GetPos().H) // 渲染目标总高 - 当前弹窗的高
	}

	// 更新滚动条位置
	this.cursorTarget = this.cursor
	return this.Refresh(modules)
}

// smooth 模式 直接前往某个游标位置
func (this *ScrollBox) ScrollToSmooth(modules common.Modules, amount int) error {
	this.cursorTarget = (float32)(amount)

	h, err := this.contents.GetGraphicsHeight()
	if err != nil {
		return err
	}

	if this.cursorTarget < 0 {
		this.cursorTarget = 0

	} else if this.contents != nil && this.cursorTarget > float32(h-this.GetPos().H) {
		this.cursorTarget = float32(h - this.GetPos().H)
	}

	// 更新滚动条位置
	return this.Refresh(modules)
}

// 一定速度滚动
func (this *ScrollBox) ScrollDown(modules common.Modules) error {
	amount := this.GetPos().H / scrollbox.SCROLL_SPEED_COARSE_MOD
	return this.Scroll(modules, amount)
}

func (this *ScrollBox) ScrollUp(modules common.Modules) error {
	amount := this.GetPos().H / scrollbox.SCROLL_SPEED_COARSE_MOD
	return this.Scroll(modules, -amount)
}

func (this *ScrollBox) ScrollToTop(modules common.Modules) error {
	return this.Scroll(modules, 0)
}

// 渲染     ==> 子组件修正到父组件下
// 鼠标检测 ==> 父组件上的坐标需要修正回子组件的坐标
func (this *ScrollBox) InputAssist(mouse point.Point) (point.Point, bool) {
	newMouse := point.Construct()

	if utils.IsWithinRect(this.GetPos(), mouse) {
		newMouse.X = mouse.X - this.GetPos().X
		newMouse.Y = mouse.Y - this.GetPos().Y + (int)(this.cursor)

		return newMouse, true
	}

	newMouse.X = mouse.X - this.GetPos().X
	newMouse.Y = -1

	return newMouse, false
}

func (this *ScrollBox) Logic(modules common.Modules) error {
	inpt := modules.Inpt()

	// 处理鼠标逻辑
	err := this.Logic1(modules, inpt.GetMouse().X, inpt.GetMouse().Y)
	if err != nil {
		return err
	}

	if this.GetInFocus() {
		if !this.currentChildOk && len(this.children) != 0 {
			this.GetNext(modules)
		}

		// 处理键盘逻辑
		this.tablist.Logic(modules)
	} else {
		this.tablist.Defocus()
		this.currentChildOk = false
	}

	return nil
}

func (this *ScrollBox) Logic1(modules common.Modules, x, y int) error {
	inpt := modules.Inpt()
	settings := modules.Settings()

	mouse := point.Construct(x, y)

	if utils.IsWithinRect(this.GetPos(), mouse) {
		inpt.SetLockScroll(true) // 锁住滚动
		if inpt.GetScrollUp() {
			this.ScrollUp(modules)
		}

		if inpt.GetScrollDown() {
			this.ScrollDown(modules)
		}
	} else {
		inpt.SetLockScroll(false)
	}

	gh, err := this.contents.GetGraphicsHeight()
	if err != nil {
		return err
	}

	if this.contents != nil && gh > this.GetPos().H && this.scrollbar != nil {
		switch this.scrollbar.CheckClickAt(modules, mouse.X, mouse.Y) {
		case scrollbar.UP:
			// 向上
			this.ScrollUp(modules)
		case scrollbar.DOWN:
			// 向下
			this.ScrollDown(modules)
		case scrollbar.DRAGGING:
			// 拖动滑块
			this.cursorTarget = (float32)(this.scrollbar.GetValue())
			this.cursor = this.cursorTarget
		}
	}

	// 处理 smooth mod
	maxFPS := settings.Get("max_fps").(int)
	if this.cursorTarget < this.cursor {

		// 向上
		// 计算当前游标
		this.cursor -= ((float32)(this.GetPos().H*scrollbox.SCROLL_SPEED_SMOOTH_MOD) + (this.cursor - this.cursorTarget)) / (float32)(maxFPS)

		if this.cursor < this.cursorTarget {
			this.cursor = this.cursorTarget
		}
	} else if this.cursorTarget > this.cursor {

		// 向下
		// 计算当前游标
		this.cursor += ((float32)(this.GetPos().H*scrollbox.SCROLL_SPEED_SMOOTH_MOD) + (this.cursorTarget - this.cursor)) / (float32)(maxFPS)

		if this.cursor > this.cursorTarget {
			this.cursor = this.cursorTarget
		}

	}

	return nil
}

func (this *ScrollBox) Render(modules common.Modules) error {
	eset := modules.Eset()
	device := modules.Render()
	src := rect.Construct(0, (int)(this.cursor), this.GetPos().W, this.GetPos().H)
	dest := this.GetPos()

	// 渲染自己的背景
	if this.contents != nil {
		this.contents.SetLocalFrame(this.GetLocalFrame())
		this.contents.SetOffset(this.GetLocalOffset())
		this.contents.SetClipFromRect(src)
		this.contents.SetDestFromRect(dest)
		err := device.Render(this.contents)
		if err != nil {
			return err
		}
	}

	// 渲染子组件
	for _, ptr := range this.children {
		ptr.SetLocalFrame(this.GetPos())
		localOffset := ptr.GetLocalOffset()
		localOffset.Y = (int)(this.cursor)
		ptr.SetLocalOffset(localOffset)
		if err := ptr.Render(modules); err != nil {
			return err
		}
	}

	// 渲染自己的滚动条
	h, err := this.contents.GetGraphicsHeight()
	if err != nil {
		return err
	}

	if this.contents != nil && h > this.GetPos().H && this.scrollbar != nil {
		this.scrollbar.SetLocalFrame(this.GetLocalFrame())
		this.scrollbar.SetLocalOffset(this.GetLocalOffset())
		this.scrollbar.Render(modules)
	}

	this.update = false

	// 子组件都不存在时绘制一个矩形
	if this.GetInFocus() && len(this.children) == 0 {
		topLeft := point.Construct()
		bottomRight := point.Construct()

		sbRect := this.scrollbar.GetBounds() // 滚动条的大小和范围
		topLeft.X = sbRect.X + this.GetLocalFrame().X - this.GetLocalOffset().X
		topLeft.Y = sbRect.Y + this.GetLocalFrame().Y - this.GetLocalOffset().Y
		bottomRight.X = topLeft.X + sbRect.W
		bottomRight.Y = topLeft.Y + sbRect.H

		draw := true

		// 在父级local frame 范围内才绘制
		if this.GetLocalFrame().W > 0 && (topLeft.X < this.GetLocalFrame().X || (bottomRight.X > this.GetLocalFrame().X+this.GetLocalFrame().W)) {
			draw = false
		}

		if this.GetLocalFrame().H > 0 && (topLeft.Y < this.GetLocalFrame().Y || bottomRight.Y > (this.GetLocalFrame().Y+this.GetLocalFrame().H)) {
			draw = false
		}

		if draw {
			err := device.DrawRectangle(topLeft, bottomRight, eset.Get("widgets", "selection_rect_color").(color.Color))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// 处理键盘的前后组件切换
func (this *ScrollBox) GetNext(modules common.Modules) bool {
	if len(this.children) == 0 {
		prevCursor := (int)(this.cursor)
		bottom := 0

		if this.contents != nil {
			gh, err := this.contents.GetGraphicsHeight()
			if err != nil {
				panic(err)
			}

			bottom = gh - this.GetPos().H // 底部游标
		}

		this.ScrollDown(modules) // 往下滚动内容
		if (int)(this.cursor) == bottom && prevCursor == bottom {
			return false
		}
		return true
	}

	if this.currentChildOk {
		this.children[this.currentChild].Defocus()
		this.currentChild, this.currentChildOk = this.tablist.GetNextRelativeIndex(tablist.WIDGET_SELECT_DOWN)
		if this.currentChildOk {
			this.tablist.SetCurrent(this.children[this.currentChild])
		}
	} else {
		//TODO
		this.currentChild = 0
		this.tablist.SetCurrent(this.children[this.currentChild])
		this.currentChild, this.currentChildOk = this.tablist.GetNextRelativeIndex(tablist.WIDGET_SELECT_DOWN)
		if this.currentChildOk {
			this.tablist.SetCurrent(this.children[this.currentChild])
		}
		this.currentChild, this.currentChildOk = this.tablist.GetNextRelativeIndex(tablist.WIDGET_SELECT_UP)
		if this.currentChildOk {
			this.tablist.SetCurrent(this.children[this.currentChild])
		}
	}

	if this.currentChildOk {
		this.children[this.currentChild].Focus()
		//直接跳转到
		this.ScrollToSmooth(modules, this.children[this.currentChild].GetPos().Y)
	} else {
		return false
	}

	return true

}

// 处理键盘的前后组件切换
func (this *ScrollBox) GetPrev(modules common.Modules) bool {

	// 没有子组件
	if len(this.children) == 0 {
		prevCursor := (int)(this.cursor)
		this.ScrollUp(modules)

		if this.cursor == 0 && prevCursor == 0 {
			return false
		}

		return true
	}

	// 存在子组件
	if this.currentChildOk {
		this.children[this.currentChild].Defocus()
		this.currentChild, this.currentChildOk = this.tablist.GetNextRelativeIndex(tablist.WIDGET_SELECT_UP)
		if this.currentChildOk {
			this.tablist.SetCurrent(this.children[this.currentChild])
		}
	} else {
		this.currentChild = 0
		this.tablist.SetCurrent(this.children[this.currentChild])
		this.currentChild, this.currentChildOk = this.tablist.GetNextRelativeIndex(tablist.WIDGET_SELECT_DOWN)
		if this.currentChildOk {
			this.tablist.SetCurrent(this.children[this.currentChild])
		}
		this.currentChild, this.currentChildOk = this.tablist.GetNextRelativeIndex(tablist.WIDGET_SELECT_UP)
		if this.currentChildOk {
			this.tablist.SetCurrent(this.children[this.currentChild])
		}
	}

	if this.currentChildOk {
		this.children[this.currentChild].Focus()
		this.ScrollToSmooth(modules, this.children[this.currentChild].GetPos().Y)
	} else {
		return false
	}

	return true
}

func (this *ScrollBox) Activate() {
	if this.currentChildOk {
		this.children[this.currentChild].Activate()
	}
}

func (this *ScrollBox) Deactivate() {
}

// 设置背景颜色
func (this *ScrollBox) SetBg(bg color.Color) {
	this.bg = bg
}
