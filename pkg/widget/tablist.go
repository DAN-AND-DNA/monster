package widget

import (
	"monster/pkg/common"
	"monster/pkg/common/define/inputstate"
	"monster/pkg/common/define/widget"
	"monster/pkg/common/define/widget/tablist"
	"monster/pkg/common/fpoint"
	"monster/pkg/utils"
)

// 主要是针对组件按键导航使用，使得聚焦于某个组件，不创建资源
type Tablist struct {
	widgets       map[int]common.Widget
	current       int // 聚焦的组件序号
	previous      int
	locked        bool
	scrollType    int // 总体的滚动方式
	MV_LEFT       int
	MV_RIGHT      int
	ACTIVATE      int
	prevTablist   common.WidgetTablist
	nextTablist   common.WidgetTablist
	ignoreNoMouse bool
}

func NewTablist() *Tablist {
	t := &Tablist{}
	t.Init()

	return t
}

func (this *Tablist) Init() common.WidgetTablist {
	this.widgets = map[int]common.Widget{}
	this.current = -1
	this.previous = -1
	this.scrollType = widget.SCROLL_TWO_DIRECTIONS

	// 默认是方向键和回车
	this.MV_LEFT = inputstate.LEFT
	this.MV_RIGHT = inputstate.RIGHT
	this.ACTIVATE = inputstate.ACCEPT
	return this
}

func (this *Tablist) Close() {
	this.widgets = nil
	this.prevTablist = nil
	this.nextTablist = nil
}

func (this *Tablist) CurrentIsValid() bool {
	return this.current >= 0 && this.current < len(this.widgets)
}

func (this *Tablist) PreviousIsValid() bool {
	return this.previous >= 0 && this.previous < len(this.widgets)
}

func (this *Tablist) Lock() {
	this.locked = true
	if this.CurrentIsValid() {
		this.widgets[this.current].Defocus()
	}
}

func (this *Tablist) Unlock() {
	this.locked = false
	if this.CurrentIsValid() {
		this.widgets[this.current].Focus()
	}
}

func (this *Tablist) Add(widget common.Widget) {
	if widget == nil {
		return
	}

	found := false
	for _, ptr := range this.widgets {
		if widget == ptr {
			found = true
			break
		}
	}

	if !found {
		this.widgets[len(this.widgets)] = widget
	}
}

func (this *Tablist) Remove(widget common.Widget) {
	if widget == nil {
		return
	}

	for index, ptr := range this.widgets {
		if widget == ptr {
			delete(this.widgets, index)
			break
		}
	}
}

func (this *Tablist) Clear() {
	this.widgets = map[int]common.Widget{}
}

// 将该widget设为当前选中的
func (this *Tablist) SetCurrent(widget common.Widget) bool {
	if widget == nil {
		this.current = -1
		return false
	}

	found := false
	for index, ptr := range this.widgets {
		if widget == ptr {
			this.current = index
			found = true
		} else {
			// 其他都取消选中
			ptr.Defocus()
		}
	}

	return found
}

func (this *Tablist) GetCurrent() int {
	return this.current
}

// 通过序号获得widget
func (this *Tablist) GetWidgetByIndex(index int) (common.Widget, bool) {
	if ptr, ok := this.widgets[index]; ok {
		return ptr, true
	}

	return nil, false
}

func (this *Tablist) Size() int {
	return len(this.widgets)
}

// 全部不聚焦，当前focus的widget清除
func (this *Tablist) Defocus() {
	for _, ptr := range this.widgets {
		ptr.Defocus()
	}

	this.current = -1
}

func (this *Tablist) GetNextRelativeIndex(dir int) (int, bool) {
	if this.current == -1 {
		return -1, false
	}

	next := this.current
	minDistance := float32(-1)

	for index, ptr := range this.widgets {
		if this.current == index {
			continue
		}

		if !ptr.GetEnableTablistNav() {
			continue
		}

		cPos := this.widgets[this.current].GetPos()
		iPos := ptr.GetPos()

		wDiv := 2
		if ptr.GetTablistNavRight() {
			wDiv = 1
		}

		p1 := fpoint.Construct((float32)(cPos.X+cPos.W/wDiv), (float32)(cPos.Y+cPos.H/2))
		p2 := fpoint.Construct((float32)(iPos.X+iPos.W/wDiv), (float32)(iPos.Y+iPos.H/2))

		if dir == tablist.WIDGET_SELECT_LEFT && p1.X < p2.X {
			continue
		} else if dir == tablist.WIDGET_SELECT_RIGHT && p1.X >= p2.X {
			continue
		} else if dir == tablist.WIDGET_SELECT_UP && p1.Y < p2.Y {
			continue
		} else if dir == tablist.WIDGET_SELECT_DOWN && p1.Y >= p2.Y {
			continue
		}

		dist := utils.CalcDist(p1, p2)

		// 获得最小距离
		if minDistance == -1 || dist < minDistance {
			minDistance = dist
			next = index
		}
	}

	if next == this.current {
		return -1, false
	}

	return next, true
}

// 获取下一个widget, dir为寻找方式
func (this *Tablist) GetNext(modules common.Modules, inner bool, dir int) (common.Widget, bool) {
	if len(this.widgets) == 0 {
		// widget为空，则跳到下个列表
		if this.nextTablist != nil {
			this.Defocus()     // 清理自己的状态
			this.locked = true // 锁定自己
			this.nextTablist.Unlock()
			return this.nextTablist.GetNext(modules, false, tablist.WIDGET_SELECT_AUTO) // 自增模式
		}
		return nil, false
	}

	if this.CurrentIsValid() {
		if inner && this.widgets[this.current].GetNext(modules) {
			return nil, false
		}

		this.widgets[this.current].Defocus()
	}

	if dir == tablist.WIDGET_SELECT_AUTO {
		// 自增模式
		this.current++

		if this.current >= len(this.widgets) {
			this.current = 0
		}

	} else {
		// 最小距离模式
		if next, ok := this.GetNextRelativeIndex(dir); ok {
			this.current = next
		} else {
			if this.nextTablist == nil {
				this.current++
				if this.current >= len(this.widgets) {
					this.current = 0
				}
			} else {
				this.Defocus()
				this.locked = true
				this.nextTablist.Unlock()
				return this.nextTablist.GetNext(modules, false, tablist.WIDGET_SELECT_AUTO)
			}
		}
	}

	this.widgets[this.current].Focus()
	return this.widgets[this.current], true
}

func (this *Tablist) GetPrev(modules common.Modules, inner bool, dir int) (common.Widget, bool) {
	if len(this.widgets) == 0 {
		if this.prevTablist != nil {
			this.Defocus()                                                              // 清理自己的状态
			this.locked = true                                                          // 锁定自己
			this.prevTablist.Unlock()                                                   // 解锁前一个列表
			return this.prevTablist.GetPrev(modules, false, tablist.WIDGET_SELECT_AUTO) // 自增模式
		}
		return nil, false
	}

	if this.CurrentIsValid() {
		if inner && this.widgets[this.current].GetPrev(modules) {
			return nil, false
		}

		this.widgets[this.current].Defocus()
	}

	if this.current == -1 {
		this.current = 0
	} else if dir == tablist.WIDGET_SELECT_AUTO {
		//自增模式
		this.current--

		if this.current <= -1 {
			this.current = len(this.widgets) - 1
		}
	} else {
		if next, ok := this.GetNextRelativeIndex(dir); ok {
			this.current = next
		} else {
			if this.prevTablist == nil {
				this.current--

				if this.current <= -1 {
					this.current = len(this.widgets) - 1
				}
			} else {
				// 清理自己，准备跳转到其他list
				this.Defocus()
				this.locked = true
				this.prevTablist.Unlock()
				return this.prevTablist.GetPrev(modules, false, tablist.WIDGET_SELECT_AUTO)
			}
		}
	}

	// 找到合适的了
	this.widgets[this.current].Focus()
	return this.widgets[this.current], true
}

func (this *Tablist) DeactivatePrevious() {
	if this.PreviousIsValid() && this.previous != this.current {
		this.widgets[this.previous].Deactivate()
	}
}

func (this *Tablist) Activate() {
	if this.CurrentIsValid() {
		this.widgets[this.current].Activate()
		this.previous = this.current
	}
}

func (this *Tablist) SetPrevTablist(tl common.WidgetTablist) {
	this.prevTablist = tl
}

func (this *Tablist) SetNextTablist(tl common.WidgetTablist) {
	this.nextTablist = tl
}

func (this *Tablist) SetScrollType(scrollType int) {
	this.scrollType = scrollType
}

func (this *Tablist) SetInput(left, right, activate int) {
	this.MV_LEFT = left
	this.MV_RIGHT = right
	this.ACTIVATE = activate
}

// 只关心键盘逻辑
func (this *Tablist) Logic(modules common.Modules) error {
	inpt := modules.Inpt()
	settings := modules.Settings()

	// 自己被锁住则退出
	if this.locked {
		return nil
	}

	// 使用键盘时
	if !inpt.UsingMouse(settings) || this.ignoreNoMouse {

		// 组件自己的滚动方式
		innerScrollType := widget.SCROLL_VERTICAL // 垂直方向
		_ = innerScrollType

		if this.CurrentIsValid() && this.widgets[this.current].GetScrollType() != widget.SCROLL_TWO_DIRECTIONS {
			innerScrollType = this.widgets[this.current].GetScrollType()
		}

		// 总体的滚动方式 垂直方向或者both
		if this.scrollType == widget.SCROLL_VERTICAL || this.scrollType == widget.SCROLL_TWO_DIRECTIONS {
			if inpt.GetPressing(inputstate.DOWN) && !inpt.GetLock(inputstate.DOWN) {
				inpt.SetLock(inputstate.DOWN, true) // 占用

				if innerScrollType == widget.SCROLL_VERTICAL {
					// 向下找距离最小
					this.GetNext(modules, true, tablist.WIDGET_SELECT_DOWN)
				} else if innerScrollType == widget.SCROLL_HORIZONTAL {
					this.GetNext(modules, false, tablist.WIDGET_SELECT_DOWN)
				}

			} else if inpt.GetPressing(inputstate.UP) && !inpt.GetLock(inputstate.UP) {
				inpt.SetLock(inputstate.UP, true) //占用

				if innerScrollType == widget.SCROLL_VERTICAL {
					// 向下找距离最小
					this.GetPrev(modules, true, tablist.WIDGET_SELECT_UP)
				} else if innerScrollType == widget.SCROLL_HORIZONTAL {
					this.GetPrev(modules, false, tablist.WIDGET_SELECT_UP)
				}
			}
		}

		// 总体的滚动方式 水平方向或者both
		if this.scrollType == widget.SCROLL_HORIZONTAL || this.scrollType == widget.SCROLL_TWO_DIRECTIONS {
			if inpt.GetPressing(this.MV_LEFT) && !inpt.GetLock(this.MV_LEFT) {
				inpt.SetLock(this.MV_LEFT, true) // 占用

				// 向左找距离最小
				if innerScrollType == widget.SCROLL_VERTICAL {
					this.GetPrev(modules, false, tablist.WIDGET_SELECT_LEFT)
				} else if innerScrollType == widget.SCROLL_HORIZONTAL {
					this.GetPrev(modules, true, tablist.WIDGET_SELECT_LEFT)
				}

			} else if inpt.GetPressing(this.MV_RIGHT) && !inpt.GetLock(this.MV_RIGHT) {
				inpt.SetLock(this.MV_RIGHT, true) //占用

				// 向右找距离最小
				if innerScrollType == widget.SCROLL_VERTICAL {
					this.GetNext(modules, false, tablist.WIDGET_SELECT_RIGHT)
				} else if innerScrollType == widget.SCROLL_HORIZONTAL {
					this.GetNext(modules, true, tablist.WIDGET_SELECT_RIGHT)
				}
			}
		}

		if inpt.GetPressing(this.ACTIVATE) && !inpt.GetLock(this.ACTIVATE) {
			inpt.SetLock(this.ACTIVATE, true)
			this.DeactivatePrevious()
			this.Activate()
		}
	}

	// 使用鼠标时，全部组件都不聚焦，取消前面的键盘按键导致的聚焦
	if inpt.GetPressing(inputstate.MAIN1) &&
		!inpt.GetLock(inputstate.MAIN1) &&
		this.CurrentIsValid() &&
		!utils.IsWithinRect(this.widgets[this.GetCurrent()].GetPos(), inpt.GetMouse()) {
		this.Defocus()
	}

	return nil
}

func (this *Tablist) SetIngoreNoMouse(val bool) {
	this.ignoreNoMouse = val
}
