package menu

import (
	"monster/pkg/common"
	"monster/pkg/common/define/fontengine"
	"monster/pkg/common/define/widget/button"
	"monster/pkg/common/gameres"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/game/base"
	"monster/pkg/utils"
)

type Confirm struct {
	base.Menu
	buttonConfirm    common.WidgetButton
	buttonClose      common.WidgetButton
	label            common.WidgetLabel
	boxMsg           string
	hasConfirmButton bool
	confirmClicked   bool
	cancelClicked    bool
	isWithinButtons  bool // 是否悬停在关闭按钮上
}

func NewConfirm(modules common.Modules, buttonMsg, boxMsg string) *Confirm {
	c := &Confirm{}
	c.Init(modules, buttonMsg, boxMsg)
	return c
}

func (this *Confirm) Init(modules common.Modules, buttonMsg, boxMsg string) gameres.MenuConfirm {
	mods := modules.Mods()
	widgetf := modules.Widgetf()

	// base
	this.Menu = base.ConstructMenu(modules)

	// self
	this.label = widgetf.New("label").(common.WidgetLabel).Init(modules)
	infile := fileparser.Construct()
	err := infile.Open("menus/confirm.txt", true, mods)
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	for infile.Next(mods) {
		if this.ParseMenuKey(infile.Key(), infile.Val()) {
			continue
		}
	}

	if buttonMsg != "" {
		this.hasConfirmButton = true
	}

	this.boxMsg = boxMsg
	this.GetTablist().SetIngoreNoMouse(true)

	if this.hasConfirmButton {
		this.buttonConfirm = widgetf.New("button").(common.WidgetButton).Init(modules, button.DEFAULT_FILE)
		this.buttonConfirm.SetLabel(modules, buttonMsg) // 按钮大小和标题文字大小一致
		this.GetTablist().Add(this.buttonConfirm)
	}

	this.buttonClose = widgetf.New("button").(common.WidgetButton).Init(modules, "images/menus/buttons/button_x.png")
	this.GetTablist().Add(this.buttonClose)
	this.SetBackground(modules, "images/menus/confirm_bg.png")

	return this
}

func (this *Confirm) Clear() {
	if this.buttonConfirm != nil {
		this.buttonConfirm.Close()
	}

	if this.buttonClose != nil {
		this.buttonClose.Close()
	}

	if this.label != nil {
		this.label.Close()
	}
}

func (this *Confirm) Close() {
	this.Menu.Close(this)
}

func (this *Confirm) Align(modules common.Modules) error {
	font := modules.Font()

	this.Menu.Align(modules)

	this.label.SetJustify(fontengine.JUSTIFY_CENTER)
	this.label.SetText(this.boxMsg)
	this.label.SetColor(font.GetColor(fontengine.COLOR_MENU_NORMAL))

	windowArea := this.GetWindowArea() // 绘制位置
	if this.hasConfirmButton {
		// 确认按钮居中
		this.buttonConfirm.SetPosX(windowArea.X + windowArea.W/2 - this.buttonConfirm.GetPos().W/2)
		this.buttonConfirm.SetPosY(windowArea.Y + windowArea.H/2)
		this.buttonConfirm.Refresh(modules)

		// 弹窗文字居中 在按钮上面
		this.label.SetPos1(modules, windowArea.X+windowArea.W/2, windowArea.Y+windowArea.H-(this.buttonConfirm.GetPos().H/2))
	} else {
		// 文字居中
		this.label.SetPos1(modules, windowArea.X+windowArea.W/2, windowArea.Y+windowArea.H/4)
	}

	// 右上角关闭按钮
	this.buttonClose.SetPosX(windowArea.X + windowArea.W)
	this.buttonClose.SetPosY(windowArea.Y)

	return nil
}

func (this *Confirm) Logic(modules common.Modules, pc gameres.Avatar, powers gameres.PowerManager) error {
	inpt := modules.Inpt()

	if this.GetVisible() {
		// 处理键盘跳转逻辑
		this.GetTablist().Logic(modules)
		this.confirmClicked = false

		if this.hasConfirmButton && this.buttonConfirm.CheckClick(modules) {
			this.confirmClicked = true
		}

		if this.buttonClose.CheckClick(modules) {
			this.SetVisible(false)
			this.confirmClicked = true
			this.cancelClicked = true
		}

		// 是否悬停在关闭按钮上
		this.isWithinButtons = (this.buttonClose.GetInFocus() || utils.IsWithinRect(this.buttonClose.GetPos(), inpt.GetMouse())) || (this.hasConfirmButton && (this.buttonConfirm.GetInFocus() || utils.IsWithinRect(this.buttonConfirm.GetPos(), inpt.GetMouse())))
	}

	return nil
}

func (this *Confirm) Render(modules common.Modules) error {
	err := this.Menu.Render(modules) // 背景
	if err != nil {
		return err
	}

	// 渲染文字
	err = this.label.Render(modules)
	if err != nil {
		return err
	}

	if this.hasConfirmButton {
		err = this.buttonConfirm.Render(modules)
		if err != nil {
			return err
		}
	}

	err = this.buttonClose.Render(modules)
	if err != nil {
		return err
	}

	return nil
}
