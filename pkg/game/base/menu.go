package base

import (
	"monster/pkg/common"
	"monster/pkg/common/define"
	"monster/pkg/common/gameres"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/utils"
	"monster/pkg/utils/parsing"
)

type Menu struct {
	visible        bool
	windowArea     rect.Rect // 组件显示在窗口上的位置和大小
	alignment      int
	tablist        common.WidgetTablist
	background     common.Sprite
	windowAreaBase point.Point // 锚点
}

func ConstructMenu(modules common.Modules) Menu {
	widgetf := modules.Widgetf()

	m := Menu{
		alignment: define.ALIGN_TOPLEFT,
		tablist:   widgetf.New("tablist").(common.WidgetTablist).Init(),
	}
	return m
}

func (this *Menu) clear() {
	if this.background != nil {
		this.background.Close()
		this.background = nil
	}

	if this.tablist != nil {
		this.tablist.Close()
		this.tablist = nil
	}
}

func (this *Menu) Close(impl gameres.Menu) {
	impl.Clear()

	this.clear()
}

func (this *Menu) GetWindowArea() rect.Rect {
	return this.windowArea
}

// 加载背景图片
func (this *Menu) SetBackground(modules common.Modules, backgroundImage string) error {
	render := modules.Render()
	mods := modules.Mods()
	settings := modules.Settings()

	if this.background != nil {
		this.background.Close()
		this.background = nil
	}

	graphics, err := render.LoadImage(settings, mods, backgroundImage)
	if err != nil {
		return err
	}
	defer graphics.UnRef()

	this.background, err = graphics.CreateSprite()
	if err != nil {
		return err
	}

	this.background.SetClip(0, 0, this.windowArea.W, this.windowArea.H)
	this.background.SetDestFromRect(this.windowArea)

	return nil
}

func (this *Menu) SetBackgroundDest(dest rect.Rect) {
	if this.background != nil {
		this.background.SetDestFromRect(dest)
	}
}

func (this *Menu) SetBackgroundClip(clip rect.Rect) {
	if this.background != nil {
		this.background.SetClipFromRect(clip)
	}
}

func (this *Menu) Render(modules common.Modules) error {
	render := modules.Render()

	if this.background != nil {
		err := render.Render(this.background)
		if err != nil {
			return err
		}
	}

	return nil
}

// 在窗口的显示位置 对齐到某个屏幕位置
func (this *Menu) Align(modules common.Modules) error {
	settings := modules.Settings()
	eset := modules.Eset()

	this.windowArea.X = this.windowAreaBase.X
	this.windowArea.Y = this.windowAreaBase.Y

	this.windowArea = utils.AlignToScreenEdge(settings, eset, this.alignment, this.windowArea)

	if this.background != nil {
		this.background.SetClip(0, 0, this.windowArea.W, this.windowArea.H)
		this.background.SetDestFromRect(this.windowArea)
	}

	return nil
}

// 设置锚点
func (this *Menu) SetWindowPos(x, y int) {
	this.windowAreaBase.X = x
	this.windowAreaBase.Y = y
}

func (this *Menu) ParseMenuKey(key, val string) bool {
	value := val
	switch key {
	case "pos":
		value += ","
		this.windowArea = parsing.ToRect(value)
		this.SetWindowPos(this.windowArea.X, this.windowArea.Y)

	case "align":
		this.alignment = parsing.ToAlignment(value, define.ALIGN_TOPLEFT)
	case "soundfx_open":
		//TODO
		// sound
	case "soundfx_close":
	default:
		return false
	}

	return true
}

func (this *Menu) GetCurrentTabList() common.WidgetTablist {
	if this.tablist.GetCurrent() != -1 {
		return this.tablist
	}

	return nil
}

func (this *Menu) DefocusTablists() {
	this.tablist.Defocus()
}

func (this *Menu) GetTablist() common.WidgetTablist {
	return this.tablist
}

func (this *Menu) SetVisible(val bool) {
	this.visible = val
}
func (this *Menu) GetVisible() bool {
	return this.visible
}
