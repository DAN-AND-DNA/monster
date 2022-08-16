package widget

import (
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define/fontengine"
	"monster/pkg/common/define/inputstate"
	"monster/pkg/common/define/widget"
	"monster/pkg/common/labelinfo"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/utils"
	"monster/pkg/widget/base"
)

// 标签控制器
type TabControl struct {
	base.Widget
	activeTabSurface   common.Sprite // 单个选中的标签精灵
	inactiveTabSurface common.Sprite // 单个未选中的标签精灵
	titles             []string
	tabs               []rect.Rect // 每个标签的大小和位置 宽度 = 文字宽 + 2 个间隔宽
	activeLabels       []*Label
	inactiveLabels     []*Label
	enabled            []bool    // 是否已经设置过标题，进行启用
	activeTab          uint      // 选中的标签序号
	tabsArea           rect.Rect // 标签范围区域
	lockMain1          bool
	dragging           bool
}

func NewTabControl(modules common.Modules) *TabControl {
	tc := &TabControl{}
	tc.Init(modules)

	return tc
}

func (this *TabControl) Init(modules common.Modules) common.WidgetTabControl {
	// base
	this.Widget = base.ConstructWidget()

	// self
	err := this.LoadGraphics(modules)
	if err != nil {
		panic(err)
	}

	this.SetScrollType(widget.SCROLL_HORIZONTAL)

	return this
}

func (this *TabControl) Close() {
	this.Widget.Close(this)
}

func (this *TabControl) Clear() {
	if this.activeTabSurface != nil {
		this.activeTabSurface.Close()
		this.activeTabSurface = nil
	}

	if this.inactiveTabSurface != nil {
		this.inactiveTabSurface.Close()
		this.inactiveTabSurface = nil
	}

	lenal := len(this.activeLabels)
	for i := 0; i < lenal; i++ {
		this.activeLabels[i].Close()
	}
	this.activeLabels = nil

	lenial := len(this.inactiveLabels)
	for i := 0; i < lenial; i++ {
		this.inactiveLabels[i].Close()
	}
	this.inactiveLabels = nil
}

// 标签创建精灵
func (this *TabControl) LoadGraphics(modules common.Modules) error {
	render := modules.Render()
	settings := modules.Settings()
	mods := modules.Mods()

	graphics1, err := render.LoadImage(settings, mods, "images/menus/tab_active.png")
	if err != nil {
		return err
	}
	defer graphics1.UnRef()

	this.activeTabSurface, err = graphics1.CreateSprite()
	if err != nil {
		return err
	}

	graphics2, err := render.LoadImage(settings, mods, "images/menus/tab_inactive.png")
	if err != nil {
		return err
	}
	defer graphics2.UnRef()

	this.inactiveTabSurface, err = graphics2.CreateSprite()
	if err != nil {
		return err
	}

	return nil
}

// 帧逻辑
func (this *TabControl) Logic(modules common.Modules) error {
	inpt := modules.Inpt()
	mouse := inpt.GetMouse()
	return this.Logic1(modules, mouse.X, mouse.Y)
}

func (this *TabControl) Logic1(modules common.Modules, x, y int) error {
	inpt := modules.Inpt()

	mouse := point.Construct(x, y)

	if utils.IsWithinRect(this.tabsArea, mouse) && (!this.lockMain1 || this.dragging) {
		this.lockMain1 = false
		this.dragging = false

		if inpt.GetPressing(inputstate.MAIN1) {
			inpt.SetLock(inputstate.MAIN1, true)
			this.dragging = true

			for i := 0; i < len(this.tabs); i++ {

				if utils.IsWithinRect(this.tabs[i], mouse) && this.enabled[i] {
					this.activeTab = (uint)(i)
					return nil
				}
			}
		}
	} else {
		this.lockMain1 = inpt.GetPressing(inputstate.MAIN1)
	}

	if !inpt.GetPressing(inputstate.MAIN1) {
		this.dragging = false
	}
	return nil
}

func (this *TabControl) Render(modules common.Modules) error {
	render := modules.Render()
	eset := modules.Eset()

	// 每个标签
	for i := 0; i < len(this.tabs); i++ {
		this.RenderTab(modules, i)
	}

	// 选择框
	if this.GetInFocus() {
		topLeft := point.Construct()
		bottomRight := point.Construct()

		topLeft.X = this.tabs[this.activeTab].X
		topLeft.Y = this.tabs[this.activeTab].Y

		bottomRight.X = topLeft.X + this.tabs[this.activeTab].W
		bottomRight.Y = topLeft.Y + this.tabs[this.activeTab].H

		err := render.DrawRectangle(topLeft, bottomRight, eset.Get("widgets", "selection_rect_color").(color.Color))
		if err != nil {
			return err
		}
	}

	return nil
}

// 绘制单个标签
func (this *TabControl) RenderTab(modules common.Modules, i int) error {
	eset := modules.Eset()
	render := modules.Render()

	if !this.enabled[i] {
		return nil
	}

	// 获得精灵的宽
	gfxWidth, err := this.activeTabSurface.GetGraphicsWidth()
	if err != nil {
		return err
	}

	// 每个标签的间隔
	tabPadding := eset.Get("widgets", "padding").(point.Point)

	// 每个标签宽
	widthToRender := this.tabs[i].W - tabPadding.X
	renderCursor := 0

	src := rect.Construct()
	dest := rect.Construct()

	src.Y = 0
	src.H = this.tabs[i].H
	dest.Y = this.tabs[i].Y

	// 除了左边间距的宽度
	for renderCursor < widthToRender {
		dest.X = this.tabs[i].X + renderCursor

		// 1. 绘制精灵的左中
		if renderCursor == 0 {
			src.X = 0
			src.W = this.tabs[i].W - tabPadding.X

			// 精灵的宽小于标签的宽
			if src.W > gfxWidth-tabPadding.X {
				src.W = gfxWidth - tabPadding.X
			}
		} else {
			// 2. 当精灵的宽小于标签的宽，则绘制精灵的中 (可选)
			src.X = tabPadding.X
			src.W = this.tabs[i].W - tabPadding.X*2

			if src.W >= gfxWidth-tabPadding.X*2 {
				src.W = gfxWidth - tabPadding.X*2
			}
		}

		renderCursor += src.W

		if renderCursor > this.tabs[i].W {
			src.W = this.tabs[i].W - (renderCursor - src.W)
		}

		if (uint)(i) == this.activeTab {
			this.activeTabSurface.SetClipFromRect(src)
			this.activeTabSurface.SetDestFromRect(dest)
			render.Render(this.activeTabSurface)
		} else {
			this.inactiveTabSurface.SetClipFromRect(src)
			this.inactiveTabSurface.SetDestFromRect(dest)
			render.Render(this.inactiveTabSurface)
		}
	}

	// 3. 绘制右则线条
	src.X = gfxWidth - tabPadding.X
	src.W = tabPadding.X
	dest.X = this.tabs[i].X + this.tabs[i].W - tabPadding.X // 绘制位置

	if (uint)(i) == this.activeTab {
		this.activeTabSurface.SetClipFromRect(src)
		this.activeTabSurface.SetDestFromRect(dest)
		render.Render(this.activeTabSurface)
	} else {
		this.inactiveTabSurface.SetClipFromRect(src)
		this.inactiveTabSurface.SetDestFromRect(dest)
		render.Render(this.inactiveTabSurface)
	}

	// 4. 绘制文字
	if (uint)(i) == this.activeTab {
		this.activeLabels[i].Render(modules)
	} else {
		this.inactiveLabels[i].Render(modules)
	}
	return nil
}

// 设置标签的标题, 并设置标签为可用
func (this *TabControl) SetTabTitle(modules common.Modules, index int, title string) {
	lenT := len(this.titles)

	if index+1 > lenT {
		for i := 0; i < index+1-lenT; i++ {
			this.titles = append(this.titles, "")
			this.tabs = append(this.tabs, rect.Construct())
			this.activeLabels = append(this.activeLabels, NewLabel(modules))
			this.inactiveLabels = append(this.inactiveLabels, NewLabel(modules))
			this.enabled = append(this.enabled, false)
		}
	}

	this.titles[index] = title
	this.enabled[index] = true
}

func (this *TabControl) GetActiveTab() uint {
	return this.activeTab
}

// 选中标签
func (this *TabControl) SetActiveTab(tab uint) {
	lenTabs := (uint)(len(this.tabs))
	if tab > lenTabs {
		tab = 0
	} else if tab == lenTabs {
		tab = lenTabs - 1
	}

	for i := tab; tab < lenTabs; i++ {
		if this.enabled[i] {
			this.activeTab = i
			return
		}
	}

	for i := uint(0); i < tab; i++ {
		if this.enabled[i] {
			this.activeTab = i
			return
		}
	}

	// 没有启动的tab
	this.activeTab = tab
}

// 获取标签高
func (this *TabControl) GetTabHeight() (int, error) {
	if this.activeTabSurface != nil {
		h, err := this.activeTabSurface.GetGraphicsHeight()
		if err != nil {
			return 0, err
		}

		return h, nil
	}

	return 0, nil
}

// 设置tab的位置
func (this *TabControl) SetMainArea(modules common.Modules, x, y int) error {
	eset := modules.Eset()
	font := modules.Font()

	// 全部的标签的大小和位置
	this.tabsArea.X = x
	this.tabsArea.Y = y
	this.tabsArea.W = 0 // 根据文字进行计算宽度
	var err error
	this.tabsArea.H, err = this.GetTabHeight()
	if err != nil {
		return err
	}

	xOffset := this.tabsArea.X

	tabPadding := eset.Get("widgets", "padding").(point.Point)

	// 设置激活和为激活时候标签文字信息
	for i := 0; i < len(this.tabs); i++ {
		this.tabs[i].Y = this.tabsArea.Y
		this.tabs[i].H = this.tabsArea.H
		this.tabs[i].X = xOffset

		this.activeLabels[i].SetPos1(modules, this.tabs[i].X+tabPadding.X, this.tabs[i].Y+tabPadding.Y+this.tabs[i].H/2)
		this.activeLabels[i].SetValign(labelinfo.VALIGN_CENTER) // 垂直对齐
		this.activeLabels[i].SetText(this.titles[i])
		this.activeLabels[i].SetColor(font.GetColor(fontengine.COLOR_WIDGET_NORMAL))

		this.inactiveLabels[i].SetPos1(modules, this.tabs[i].X+tabPadding.X, this.tabs[i].Y+tabPadding.Y+this.tabs[i].H/2)
		this.inactiveLabels[i].SetValign(labelinfo.VALIGN_CENTER) // 垂直对齐
		this.inactiveLabels[i].SetText(this.titles[i])
		this.inactiveLabels[i].SetColor(font.GetColor(fontengine.COLOR_WIDGET_NORMAL))

		// 若设置了标题则宽度可知
		if this.enabled[i] {
			this.tabs[i].W = this.activeLabels[i].GetBounds(modules).W + tabPadding.X*2
			this.tabsArea.W += this.tabs[i].W
			xOffset += this.tabs[i].W // 下个标签的x坐标
		}
	}

	if !this.enabled[this.activeTab] {
		this.GetNext(modules)
	}

	return nil
}

// 获得下一个可用的标签
func (this *TabControl) GetNextEnabledTab(tab uint) uint {
	for i := tab + 1; i < (uint)(len(this.tabs)); i++ {
		if this.enabled[i] {
			return i
		}
	}

	return tab
}

// 获取前一个可用的标签
func (this *TabControl) GetPrevEnabledTab(tab uint) uint {
	for i := tab - 1; i < (uint)(len(this.tabs)); i-- {
		if this.enabled[i] {
			return i
		}
	}

	return tab
}

// 激活下一个可用标签
func (this *TabControl) GetNext(common.Modules) bool {
	this.SetActiveTab(this.GetNextEnabledTab(this.activeTab))
	return true
}

// 激活前一个可用标签
func (this *TabControl) GetPrev(common.Modules) bool {
	this.SetActiveTab(this.GetPrevEnabledTab(this.activeTab))
	return true
}

func (this *TabControl) SetEnabled(index uint, val bool) {
	if index > (uint)(len(this.enabled)) {
		return
	}

	this.enabled[index] = val
}

func (this *TabControl) Activate() {
}

func (this *TabControl) Deactivate() {
}
