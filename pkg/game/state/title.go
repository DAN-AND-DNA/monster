package state

import (
	"fmt"
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define"
	"monster/pkg/common/define/fontengine"
	"monster/pkg/common/define/inputstate"
	"monster/pkg/common/define/widget/button"
	"monster/pkg/common/gameres"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/config/version"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/filesystem/logfile"
	"monster/pkg/game/base"
	"monster/pkg/utils"
	"monster/pkg/utils/parsing"
)

type Title struct {
	base.State

	logo          common.Sprite
	buttonPlay    common.WidgetButton
	buttonExit    common.WidgetButton
	buttonCfg     common.WidgetButton
	buttonCredits common.WidgetButton
	labelVersion  common.WidgetLabel
	tablist       common.WidgetTablist
	posLogo       point.Point // logo的位置
	alignLogo     int

	//TODO
	// MenuMovementType
}

func NewTitle(modules common.Modules, gameRes gameres.GameRes) *Title {
	t := &Title{}

	t.init(modules, gameRes)
	return t
}

func (this *Title) init(modules common.Modules, gameRes gameres.GameRes) gameres.GameStateTitle {
	msg := modules.Msg()
	eset := modules.Eset()
	settings := modules.Settings()
	font := modules.Font()
	mods := modules.Mods()
	render := modules.Render()
	widgetf := modules.Widgetf()

	// base
	this.State = base.ConstructState(modules)

	// self
	this.buttonPlay = widgetf.New("button").(common.WidgetButton).Init(modules, button.DEFAULT_FILE)
	this.buttonExit = widgetf.New("button").(common.WidgetButton).Init(modules, button.DEFAULT_FILE)
	this.buttonCfg = widgetf.New("button").(common.WidgetButton).Init(modules, button.DEFAULT_FILE)
	this.buttonCredits = widgetf.New("button").(common.WidgetButton).Init(modules, button.DEFAULT_FILE)
	this.labelVersion = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.tablist = widgetf.New("tablist").(common.WidgetTablist).Init()
	this.alignLogo = define.ALIGN_CENTER

	infile := fileparser.New()

	err := infile.Open("menus/gametitle.txt", true, mods)
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	for infile.Next(mods) {
		strVal := infile.Val()
		first := ""
		var x, y int

		// 设置锚点
		switch infile.Key() {
		case "logo":
			first, strVal = parsing.PopFirstString(strVal, "")
			graphics, err := render.LoadImage(settings, mods, first)
			if err != nil {
				panic(err)
			}
			defer graphics.UnRef()

			this.logo, err = graphics.CreateSprite()
			if err != nil {
				panic(err)
			}
			this.posLogo.X, strVal = parsing.PopFirstInt(strVal, "")
			this.posLogo.Y, strVal = parsing.PopFirstInt(strVal, "")
			this.alignLogo = parsing.ToAlignment(strVal, define.ALIGN_TOPLEFT)
		case "play_pos":
			x, strVal = parsing.PopFirstInt(strVal, "")
			y, strVal = parsing.PopFirstInt(strVal, "")
			first, strVal = parsing.PopFirstString(strVal, "")
			a := parsing.ToAlignment(first, define.ALIGN_TOPLEFT)
			this.buttonPlay.SetPosBase(x, y, a)

		case "config_pos":
			x, strVal = parsing.PopFirstInt(strVal, "")
			y, strVal = parsing.PopFirstInt(strVal, "")
			first, strVal = parsing.PopFirstString(strVal, "")
			a := parsing.ToAlignment(first, define.ALIGN_TOPLEFT)
			this.buttonCfg.SetPosBase(x, y, a)
		case "credits_pos":
			x, strVal = parsing.PopFirstInt(strVal, "")
			y, strVal = parsing.PopFirstInt(strVal, "")
			first, strVal = parsing.PopFirstString(strVal, "")
			a := parsing.ToAlignment(first, define.ALIGN_TOPLEFT)
			this.buttonCredits.SetPosBase(x, y, a)
		case "exit_pos":
			x, strVal = parsing.PopFirstInt(strVal, "")
			y, strVal = parsing.PopFirstInt(strVal, "")
			first, strVal = parsing.PopFirstString(strVal, "")
			a := parsing.ToAlignment(first, define.ALIGN_TOPLEFT)
			this.buttonExit.SetPosBase(x, y, a)
		default:
			logfile.LogError("GameStateTitle: '%s' is not a valid key.", infile.Key())
		}
	}

	// 设置按钮标题
	this.buttonPlay.SetLabel(modules, msg.Get("Play Game"))
	if !eset.Get("gameplay", "enable_playgame").(bool) {
		this.buttonPlay.SetEnabled(false)
		this.buttonPlay.SetTooltip("Enable a core mod to continue") // 提示文字
	}

	// 更新按钮标题
	this.buttonPlay.Refresh(modules)

	this.buttonCfg.SetLabel(modules, msg.Get("Configuration"))
	this.buttonCfg.Refresh(modules)

	this.buttonCredits.SetLabel(modules, msg.Get("Credits"))
	this.buttonCredits.Refresh(modules)

	this.buttonExit.SetLabel(modules, msg.Get("Exist Game"))
	this.buttonExit.Refresh(modules)

	// 右上角版本号
	this.labelVersion.SetJustify(fontengine.JUSTIFY_RIGHT) //设置水平对齐
	this.labelVersion.SetText(version.CreateVersionStringFull())
	this.labelVersion.SetColor(font.GetColor(fontengine.COLOR_MENU_NORMAL))

	// TODO
	// 设置tablist

	err = this.RefreshWidgets(modules, gameRes)
	if err != nil {
		panic(err)
	}

	this.SetForceRefreshBackground(true)

	// 是否有可玩mod
	if eset.Get("gameplay", "enable_playgame").(bool) && settings.GetLoadSlot() != "" {
		err := this.ShowLoading(modules)
		if err != nil {
			panic(err)
		}
		//TODO GameStateLoad
	}

	render.SetBackgroundColor(color.Construct(0, 0, 0, 0))

	//TODO MenuMovementType
	return this
}

func (this *Title) Clear(common.Modules, gameres.GameRes) {
	if this.logo != nil {
		this.logo.Close()
	}

	if this.buttonPlay != nil {
		this.buttonPlay.Close()
	}

	if this.buttonExit != nil {
		this.buttonExit.Close()
	}

	if this.buttonCfg != nil {
		this.buttonCfg.Close()
	}

	if this.buttonCredits != nil {
		this.buttonCredits.Close()
	}

	if this.labelVersion != nil {
		this.labelVersion.Close()
	}

	if this.tablist != nil {
		this.tablist.Close()
	}
}

func (this *Title) Close(modules common.Modules, gameRes gameres.GameRes) {
	this.State.Close(modules, gameRes, this)
}

// 更新旗下子组件的位置
func (this *Title) RefreshWidgets(modules common.Modules, gameRes gameres.GameRes) error {
	settings := modules.Settings()
	eset := modules.Eset()

	var err error
	if this.logo != nil {
		r := rect.Construct()
		r.X = this.posLogo.X
		r.Y = this.posLogo.Y
		r.W, err = this.logo.GetGraphicsWidth()
		if err != nil {
			return err
		}

		r.H, err = this.logo.GetGraphicsHeight()
		if err != nil {
			return err
		}

		r = utils.AlignToScreenEdge(settings, eset, this.alignLogo, r)
		this.logo.SetDestFromRect(r)
	}

	this.buttonPlay.SetPos1(modules, 0, 0)
	this.buttonCfg.SetPos1(modules, 0, 0)
	this.buttonCredits.SetPos1(modules, 0, 0)
	this.buttonExit.SetPos1(modules, 0, 0)

	// 锚点为(0,0) X的偏移为视口宽，右对齐
	this.labelVersion.SetPos1(modules, settings.GetViewW(), 0)

	return err
}

func (this *Title) Render(modules common.Modules, gameRes gameres.GameRes) error {
	render := modules.Render()
	platform := modules.Platform()

	// TODO
	// menu_movement_type

	err := render.Render(this.logo)
	if err != nil {
		return err
	}

	err = this.buttonPlay.Render(modules)
	if err != nil {
		return err
	}

	err = this.buttonCfg.Render(modules)
	if err != nil {
		return err
	}

	err = this.buttonCredits.Render(modules)
	if err != nil {
		return err
	}

	if platform.GetHasExitButton() {
		err = this.buttonExit.Render(modules)
		if err != nil {
			return err
		}
	}
	err = this.labelVersion.Render(modules)
	if err != nil {
		return err
	}

	return nil
}

func (this *Title) Logic(modules common.Modules, gameRes gameres.GameRes) error {
	inpt := modules.Inpt()
	eset := modules.Eset()
	platform := modules.Platform()

	// 窗口大小变化则刷新组件
	if inpt.GetWindowResized() {
		this.RefreshWidgets(modules, gameRes)
	}

	this.buttonPlay.SetEnabled(eset.Get("gameplay", "enable_playgame").(bool))

	// TODO
	// snd

	if inpt.GetPressing(inputstate.CANCEL) && !inpt.GetLock(inputstate.CANCEL) {
		inpt.SetLock(inputstate.CANCEL, true)
		this.SetExitRequested(true)
	}

	//TODO
	// menu

	//TODO
	//this.tablist.Logic()

	// 检测按钮状态变化，hover，按放等
	if this.buttonPlay.CheckClick(modules) {
		fmt.Println("click play")

		this.ShowLoading(modules)
		this.SetRequestedGameState(modules, gameRes, NewLoad(modules, gameRes))
	} else if this.buttonCfg.CheckClick(modules) {
		fmt.Println("click cfg")

		this.ShowLoading(modules)
		this.SetRequestedGameState(modules, gameRes, NewConfig(modules, gameRes))
	} else if this.buttonCredits.CheckClick(modules) {
		fmt.Println("click credits")
	} else if platform.GetHasExitButton() && this.buttonExit.CheckClick(modules) {
		this.SetExitRequested(true)
	}
	return nil
}
