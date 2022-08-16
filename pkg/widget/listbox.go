package widget

import (
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define/fontengine"
	"monster/pkg/common/define/inputstate"
	"monster/pkg/common/define/widget"
	"monster/pkg/common/define/widget/listbox"
	"monster/pkg/common/define/widget/scrollbar"
	"monster/pkg/common/labelinfo"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/common/tooltipdata"
	"monster/pkg/utils"
	"monster/pkg/widget/base"
	"sort"
)

type ListBoxItem struct {
	value    string
	tooltip  string
	selected bool
}

type ListBox struct {
	base.Widget

	filename        string
	listboxs        common.Sprite
	cursor          int // 滑块的值
	hasScrollBar    bool
	anySelected     bool
	items           []ListBoxItem // 全部内容
	vlabels         []*Label      // 能显示的每行文字
	rows            []rect.Rect   // 能进行显示的每一行的位置
	scrollbar       *ScrollBar
	posScroll       rect.Rect
	pressed         bool
	multiSelect     bool
	canDeselect     bool
	canSelect       bool
	scrollbarOffset int
	disableTextTrim bool
}

func NewListBox(modules common.Modules, height int, filename string) *ListBox {
	lb := &ListBox{}

	lb.Init(modules, height, filename)
	return lb
}

func (this *ListBox) Init(modules common.Modules, height int, filename string) common.WidgetListBox {
	render := modules.Render()
	mods := modules.Mods()
	settings := modules.Settings()

	// base
	this.Widget = base.ConstructWidget()

	// self
	this.filename = filename
	for i := 0; i < height; i++ {
		this.vlabels = append(this.vlabels, NewLabel(modules))
		this.rows = append(this.rows, rect.Construct())
	}
	this.scrollbar = NewScrollBar(modules, scrollbar.DEFAULT_FILE)
	this.posScroll = rect.Construct()
	this.canDeselect = true
	this.canSelect = true

	tmpFilename := listbox.DEFAULT_FILE
	if filename != listbox.DEFAULT_FILE {
		tmpFilename = filename
	}

	graphics, err := render.LoadImage(settings, mods, tmpFilename)
	if err != nil {
		panic(err)
	}
	defer graphics.UnRef()

	this.listboxs, err = graphics.CreateSprite()
	if err != nil {
		panic(err)
	}

	gw, err := this.listboxs.GetGraphicsWidth()
	if err != nil {
		panic(err)
	}

	gh, err := this.listboxs.GetGraphicsHeight()
	if err != nil {
		panic(err)
	}

	this.SetPosW(gw)
	this.SetPosH(gh / 3)
	this.SetScrollType(widget.SCROLL_VERTICAL)

	return this
}
func (this *ListBox) Clear() {
	for _, ptr := range this.vlabels {
		ptr.Close()
	}

	this.vlabels = nil

	if this.scrollbar != nil {
		this.scrollbar.Close()
		this.scrollbar = nil
	}

	if this.listboxs != nil {
		this.listboxs.Close()
		this.listboxs = nil
	}
}

func (this *ListBox) Close() {
	this.Widget.Close(this)
}

func (this *ListBox) Clear1(modules common.Modules) {
	this.items = nil
	this.Refresh(modules)
}

func (this *ListBox) ShiftUp(modules common.Modules) {
	this.anySelected = false

	if len(this.items) != 0 && !this.items[0].selected {
		for i := 1; i < len(this.items); i++ {
			if this.items[i].selected {
				this.anySelected = true
				tmpItem := this.items[i]
				this.items[i] = this.items[i-1]
				this.items[i-1] = tmpItem
			}
		}

		if this.anySelected {
			this.ScrollUp(modules)
		}
	}
}

func (this *ListBox) ShiftDown(modules common.Modules) {
	this.anySelected = false

	if len(this.items) != 0 && !this.items[len(this.items)-1].selected {
		for i := len(this.items) - 2; i >= 0; i-- {
			if this.items[i].selected {
				this.anySelected = true
				tmpItem := this.items[i]
				this.items[i] = this.items[i+1]
				this.items[i+1] = tmpItem
			}

			if this.anySelected {
				this.ScrollDown(modules)
			}
		}
	}
}

// 更新
func (this *ListBox) Refresh(modules common.Modules) {
	eset := modules.Eset()
	font := modules.Font()

	temp := ""
	rightMargin := 0

	if len(this.items) > len(this.rows) {
		// 无法完全显示，需要滚动
		this.hasScrollBar = true

		this.posScroll.X = this.GetPos().X + this.GetPos().W - this.scrollbar.GetPosUp().W - this.scrollbarOffset
		this.posScroll.Y = this.GetPos().Y + this.scrollbarOffset
		this.posScroll.W = this.scrollbar.GetPosUp().W
		this.posScroll.H = this.GetPos().H*len(this.rows) - this.scrollbar.GetPosDown().H - this.scrollbarOffset*2

		// 更新滚动条
		this.scrollbar.Refresh(modules,
			this.posScroll.X,
			this.posScroll.Y,
			this.posScroll.H,
			this.cursor,
			len(this.items)-len(this.rows))
	} else {
		// 不需要滚动
		this.hasScrollBar = false
		rightMargin = eset.Get("widgets", "text_margin").(point.Point).Y
	}

	// 能展示的行
	for i, _ := range this.rows {
		this.rows[i].X = this.GetPos().X
		this.rows[i].Y = (this.GetPos().H * i) + this.GetPos().Y

		if this.hasScrollBar {
			// 宽度不包括滚动条
			this.rows[i].W = this.GetPos().W - this.posScroll.W
		} else {
			this.rows[i].W = this.GetPos().W
		}

		this.rows[i].H = this.GetPos().H

		// 文字高
		padding := font.GetFontHeight()

		// 从cursor开始的行进行展现，填充内容
		if i+this.cursor < len(this.items) {
			if this.disableTextTrim {
				temp = this.items[i+this.cursor].value
			} else {
				temp = font.TrimTextToWidth(this.items[i+this.cursor].value,
					this.GetPos().W-rightMargin-padding, true, 0)
			}
		}

		// 每行的文字和颜色
		this.vlabels[i].SetPos1(modules,
			this.rows[i].X+eset.Get("widgets", "text_margin").(point.Point).X,
			this.rows[i].Y+(this.rows[i].H/2))

		this.vlabels[i].SetValign(labelinfo.VALIGN_CENTER)
		this.vlabels[i].SetText(temp)

		if i+this.cursor < len(this.items) && this.items[i+this.cursor].selected {
			this.vlabels[i].SetColor(font.GetColor(fontengine.COLOR_WIDGET_NORMAL))
		} else if i < len(this.items) {
			this.vlabels[i].SetColor(font.GetColor(fontengine.COLOR_WIDGET_DISABLED))
		}
	}
}

func (this *ListBox) SetPos1(modules common.Modules, offsetX, offsetY int) error {
	this.Widget.SetPos1(modules, offsetX, offsetY)

	// 更新显示
	this.Refresh(modules)

	return nil
}

func (this *ListBox) ScrollUp(modules common.Modules) {
	if this.cursor > 0 {
		this.cursor -= 1
	}

	this.Refresh(modules)
}

func (this *ListBox) ScrollDown(modules common.Modules) {
	if this.cursor+len(this.rows) < len(this.items) {
		this.cursor += 1
	}

	this.Refresh(modules)
}

func (this *ListBox) CheckClickAt(modules common.Modules, x, y int) bool {
	inpt := modules.Inpt()
	mouse := point.Construct(x, y)
	this.Refresh(modules)

	// 处理提示
	this.CheckTooltip(modules, mouse)

	// 可显示的滚动区域
	scrollArea := rect.Construct()
	scrollArea.X = this.rows[0].X
	scrollArea.Y = this.rows[0].Y
	scrollArea.W = this.rows[0].W
	scrollArea.H = this.rows[0].H * (len(this.rows))

	if utils.IsWithinRect(scrollArea, mouse) {
		inpt.SetLockScroll(true) // 滚动
		if inpt.GetScrollUp() {
			this.ScrollUp(modules)
		}

		if inpt.GetScrollDown() {
			this.ScrollDown(modules)
		}
	} else {
		inpt.SetLockScroll(false)
	}

	// 处理滚动条点击
	if this.hasScrollBar {
		switch this.scrollbar.CheckClickAt(modules, mouse.X, mouse.Y) {
		case scrollbar.UP:
			this.ScrollUp(modules)
		case scrollbar.DOWN:
			this.ScrollDown(modules)
		case scrollbar.DRAGGING:
			this.cursor = this.scrollbar.GetValue()
			this.Refresh(modules)
		}
	}

	if inpt.GetLock(inputstate.MAIN1) {
		return false
	}

	// 处理左键释放
	if this.pressed && !inpt.GetLock(inputstate.MAIN1) && this.canSelect {
		this.pressed = false

		// 可显示的行
		for i, _ := range this.rows {
			if i < len(this.items) {
				if utils.IsWithinRect(this.rows[i], mouse) && this.items[i+this.cursor].value != "" {

					// 无法多选 取消其他行
					if !this.multiSelect {
						for j, _ := range this.items {

							// 从cursor开始的显示内容
							if j != i+this.cursor {
								this.items[j].selected = false
							}
						}
					}

					// 选择上
					if this.items[i+this.cursor].selected {
						if this.canDeselect {
							this.items[i+this.cursor].selected = false
						}
					} else {
						this.items[i+this.cursor].selected = true
					}

					// 更新
					this.Refresh(modules)
					return true
				}
			}
		}
	}

	this.pressed = false

	// 处理点击
	if inpt.GetPressing(inputstate.MAIN1) {
		for i, _ := range this.rows {
			if utils.IsWithinRect(this.rows[i], mouse) {
				inpt.SetLock(inputstate.MAIN1, true) // 锁住
				this.pressed = true
			}
		}
	}

	return false
}

func (this *ListBox) CheckClick(modules common.Modules) bool {
	inpt := modules.Inpt()
	mouse := inpt.GetMouse()

	return this.CheckClickAt(modules, mouse.X, mouse.Y)
}

// 处理提示
func (this *ListBox) CheckTooltip(modules common.Modules, mouse point.Point) {
	inpt := modules.Inpt()
	settings := modules.Settings()
	font := modules.Font()
	tooltipm := modules.Tooltipm()

	if !inpt.UsingMouse(settings) {
		return
	}

	tipData := tooltipdata.Construct()

	for i, _ := range this.rows {
		if i < len(this.items) {
			if utils.IsWithinRect(this.rows[i], mouse) && this.items[i+this.cursor].tooltip != "" {
				tipData.AddColorText(this.items[i+this.cursor].tooltip, font.GetColor(fontengine.COLOR_WIDGET_NORMAL))
				break
			}
		}
	}

	if !tipData.IsEmpty() {
		newMouse := point.Construct(mouse.X+this.GetLocalFrame().X-this.GetLocalOffset().X, mouse.Y+this.GetLocalFrame().Y-this.GetLocalOffset().Y)
		tooltipm.Push(tipData, newMouse, tooltipdata.STYLE_FLOAT, 0)
	}
}

func (this *ListBox) Append(modules common.Modules, value, tooltip string) {
	if value == "" {
		return
	}
	this.items = append(this.items, ListBoxItem{value: value, tooltip: tooltip})
	this.Refresh(modules)
}

func (this *ListBox) Remove(modules common.Modules, index int) {
	if index >= len(this.items) || index < 0 {
		return
	}

	// 删除最后一个
	if index == len(this.items)-1 {
		this.items = this.items[0 : len(this.items)-1]
		return
	}

	old := this.items
	this.items = make([]ListBoxItem, len(old)-1)

	// 跳过中间某个
	i := 0
	for _, val := range old[:index] {
		this.items[i] = val
		i++
	}

	old = old[index+1:]
	for _, val := range old {
		this.items[i] = val
		i++
	}

	this.ScrollUp(modules)
	this.Refresh(modules)
}

func (this *ListBox) Activate() {
}

func (this *ListBox) Deactivate() {
}

func (this *ListBox) GetNext(modules common.Modules) bool {
	return false
}

func (this *ListBox) GetPrev(modules common.Modules) bool {
	return false
}

func (this *ListBox) Render(modules common.Modules) error {
	render := modules.Render()
	eset := modules.Eset()

	src := rect.Construct()
	src.X = 0
	src.W = this.GetPos().W
	src.H = this.GetPos().H

	if this.listboxs != nil {
		this.listboxs.SetLocalFrame(this.GetLocalFrame())
		this.listboxs.SetOffset(this.GetLocalOffset())
	}

	for i, _ := range this.rows {
		if i == 0 {
			src.Y = 0
		} else if i == len(this.rows)-1 {
			src.Y = this.GetPos().H * 2
		} else {
			src.Y = this.GetPos().H
		}

		if this.listboxs != nil {
			this.listboxs.SetClipFromRect(src)
			this.listboxs.SetDestFromRect(this.rows[i])
			err := render.Render(this.listboxs)
			if err != nil {
				return err
			}
		}

		if i < len(this.items) {
			this.vlabels[i].SetLocalFrame(this.GetLocalFrame())
			this.vlabels[i].SetLocalOffset(this.GetLocalOffset())
			err := this.vlabels[i].Render(modules)
			if err != nil {
				return err
			}
		}
	}

	if this.GetInFocus() {
		// 选择框
		topLeft := point.Construct()
		bottomRight := point.Construct()

		topLeft.X = this.rows[0].X + this.GetLocalFrame().X - this.GetLocalOffset().X
		topLeft.Y = this.rows[0].Y + this.GetLocalFrame().Y - this.GetLocalOffset().Y
		bottomRight.X = this.rows[len(this.rows)-1].X + this.rows[0].W + this.GetLocalFrame().X - this.GetLocalOffset().X
		bottomRight.Y = this.rows[len(this.rows)-1].Y + this.rows[0].H + this.GetLocalFrame().Y - this.GetLocalOffset().Y

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

	// 滚动条
	if this.hasScrollBar {
		this.scrollbar.SetLocalFrame(this.GetLocalFrame())
		this.scrollbar.SetLocalOffset(this.GetLocalOffset())
		err := this.scrollbar.Render(modules)
		if err != nil {
			return err
		}

	}

	return nil
}

func (this *ListBox) SetScrollbarOffset(val int) {
	this.scrollbarOffset = val
}

func (this *ListBox) SetMultiSelect(val bool) {
	this.multiSelect = true
}

func (this *ListBox) SetHeight(modules common.Modules, val int) {
	if val < 2 {
		val = 2
	}

	this.vlabels = nil
	this.rows = nil
	for i := 0; i < val; i++ {
		this.vlabels = append(this.vlabels, NewLabel(modules))
		this.rows = append(this.rows, rect.Construct())
	}

	// 更新
	this.Refresh(modules)
}

func (this *ListBox) Sort() {
	sort.Slice(this.items, func(i, j int) bool { return this.items[i].value < this.items[j].value })
}

func (this *ListBox) IsSelected(index int) bool {
	if len(this.items) == 0 {
		return false
	}

	if index >= len(this.items) {
		return false
	}

	return this.items[index].selected
}

func (this *ListBox) GetTooltip(index int) (string, bool) {
	if len(this.items) == 0 {
		return "", false
	}

	if index >= len(this.items) {
		return "", false
	}

	return this.items[index].tooltip, true
}

func (this *ListBox) GetValue(index int) (string, bool) {
	if len(this.items) == 0 {
		return "", false
	}

	if index >= len(this.items) {
		return "", false
	}

	return this.items[index].value, true
}

func (this *ListBox) GetSize() int {
	return len(this.items)
}

func (this *ListBox) SetCanDeselect(val bool) {
	this.canDeselect = val
}

func (this *ListBox) GetSelected() (int, bool) {
	for i, val := range this.items {
		if val.selected {
			return i, true
		}
	}

	return -1, false
}

func (this *ListBox) Select(index int) {
	if len(this.items) == 0 {
		return
	}

	sel, ok := this.GetSelected()
	if !this.multiSelect && ok {
		this.items[sel].selected = false
	}

	this.items[index].selected = true

}
