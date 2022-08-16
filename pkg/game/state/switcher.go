package state

import (
	"fmt"
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define"
	"monster/pkg/common/gameres"
	"monster/pkg/common/rect"
	"monster/pkg/common/timer"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/filesystem/logfile"
	"monster/pkg/utils"
	"monster/pkg/utils/parsing"
)

// 入口
type Switcher struct {
	gameRes         *GameRes
	background      common.Sprite
	backgroundImage common.Image
	fpsUpdate       timer.Timer
	lastFPS         float32
	currentState    gameres.GameState
	labelFPS        common.WidgetLabel
	fpsPosition     rect.Rect
	fpsColor        color.Color
	fpsCorner       int

	backgroundFilename string
	backgroundList     []string
	done               bool
}

func NewSwitcher(modules common.Modules) *Switcher {
	settings := modules.Settings()
	mods := modules.Mods()
	widgetf := modules.Widgetf()

	gs := &Switcher{
		fpsUpdate: timer.Construct(),
	}

	// 资源
	gs.gameRes = NewGameRes()
	gs.gameRes.NewStats(modules)

	gs.fpsUpdate.SetDuration((uint)(settings.Get("max_fps").(int) / 4))
	gs.currentState = NewTitle(modules, gs.gameRes)
	//gs.currentState = NewLoad(modules, gs.gameRes)

	// TODO
	gs.labelFPS = widgetf.New("label").(common.WidgetLabel).Init(modules)
	gs.LoadFPS(modules)

	gs.LoadBackgroundList(mods)
	if gs.currentState.GetHasBackground() {
		err := gs.LoadBackgroundImage(modules)
		if err != nil {
			panic(err)
		}
	}

	return gs
}

func (this *Switcher) Close(modules common.Modules) {
	if this.currentState != nil {
		this.currentState.Close(modules, this.gameRes)
		this.currentState = nil
	}

	if this.labelFPS != nil {
		this.labelFPS.Close()
		this.labelFPS = nil
	}

	this.FreeBackground()
}

func (this *Switcher) Logic(modules common.Modules) error {
	settings := modules.Settings()
	inpt := modules.Inpt()
	curs := modules.Render().Curs()
	tooltipm := modules.Tooltipm()
	render := modules.Render()
	mods := modules.Mods()
	eset := modules.Eset()

	// 光标逻辑
	err := curs.Logic(settings, inpt)
	if err != nil {
		return err
	}

	// 清空提示文字
	tooltipm.Clear()

	newState := this.currentState.GetRequestedGameState()
	if newState != nil {

		// 当前游戏状态是否需要重新加载背景列表 或 之前重新加载过渲染系统
		if this.currentState.GetReloadBackgrounds() || render.ReloadGraphics() {
			err := this.LoadBackgroundList(mods)
			if err != nil {
				fmt.Println("here")
				return err
			}
		}

		fmt.Println("new state")
		this.currentState.Close(modules, this.gameRes)
		this.currentState = newState
		this.currentState.IncrLoadCounter() // 2 ++1

		// reload fps
		err = this.LoadFPS(modules)
		if err != nil {
			return err
		}

		// TODO
		// load music

		// 需要背景图片
		if this.currentState.GetHasBackground() {
			err = this.LoadBackgroundImage(modules)
			if err != nil {
				return err
			}
		} else {
			this.FreeBackground()
		}
	}

	// 处理窗口逻辑
	if (inpt.GetWindowResized() || this.currentState.GetForceRefreshBackground()) && this.currentState.GetHasBackground() {

		fmt.Println("refresh bk")
		err = this.RefreshBackground(settings, eset) // 更新背景大小
		if err != nil {
			return err
		}

		this.currentState.SetForceRefreshBackground(false)
	}
	this.currentState.Logic(modules, this.gameRes)

	this.done = this.currentState.GetExitRequested()

	return nil
}

func (this *Switcher) ShowFPS(modules common.Modules, fps float32) error {
	settings := modules.Settings()
	eset := modules.Eset()
	widgetf := modules.Widgetf()

	if settings.Get("show_fps").(bool) && settings.GetShowHud() {
		if this.labelFPS == nil {
			//this.labelFPS = widget.NewLabel(font)
			this.labelFPS = widgetf.New("label").(common.WidgetLabel).Init(modules)

		}

		if this.fpsUpdate.IsEnd() {
			this.fpsUpdate.Reset(timer.BEGIN)
			avgFPS := (fps + this.lastFPS) / 2
			this.lastFPS = fps
			strFPS := fmt.Sprintf("%2.0f", avgFPS) + " fps"
			pos := utils.AlignToScreenEdge(settings, eset, this.fpsCorner, this.fpsPosition)
			this.labelFPS.SetPos1(modules, pos.X, pos.Y)
			this.labelFPS.SetText(strFPS)
			this.labelFPS.SetColor(this.fpsColor)
		}
		err := this.labelFPS.Render(modules)
		if err != nil {
			return err
		}
		this.fpsUpdate.Tick()
	}

	return nil
}

func (this *Switcher) LoadFPS(modules common.Modules) error {
	mods := modules.Mods()
	font := modules.Font()

	infile := fileparser.Construct()

	err := infile.Open("menus/fps.txt", true, mods)
	if err != nil && !utils.IsNotExist(err) {
		return err
	}
	defer infile.Close()

	for infile.Next(mods) {
		if infile.Key() == "position" {
			firstInt, strVal := parsing.PopFirstInt(infile.Val(), "")
			this.fpsPosition.X = firstInt
			firstInt, strVal = parsing.PopFirstInt(strVal, "")
			this.fpsPosition.Y = firstInt
			firstStr, strVal := parsing.PopFirstString(strVal, "")
			this.fpsCorner = parsing.ToAlignment(firstStr, define.ALIGN_TOPLEFT) // 在哪个角落
		} else if infile.Key() == "color" {
			this.fpsColor = parsing.ToRGB(infile.Val())
		} else {
			logfile.LogError("GameSwitcher: '%s' is not a valid key.", infile.Key())
		}
	}

	font.SetFont("font_regular")
	this.fpsPosition.W = font.CalcWidth("00 fps")
	this.fpsPosition.H = font.GetLineHeight()

	if this.labelFPS != nil {
		this.labelFPS.Close()
		this.labelFPS = nil
	}

	return nil
}

func (this *Switcher) LoadBackgroundList(mods common.ModManager) error {
	this.backgroundList = this.backgroundList[:0]
	this.FreeBackground()

	infile := fileparser.New()
	err := infile.Open("engine/menu_backgrounds.txt", true, mods)
	if err != nil {
		return err
	}
	defer infile.Close()

	for infile.Next(mods) {
		if infile.Key() == "background" {
			this.backgroundList = append(this.backgroundList, infile.Val())
		} else {
			return common.Err_bad_key_in_gameswitcher
		}
	}
	return nil
}

func (this *Switcher) LoadBackgroundImage(modules common.Modules) error {
	settings := modules.Settings()
	eset := modules.Eset()
	mods := modules.Mods()
	render := modules.Render()

	if len(this.backgroundList) == 0 || this.backgroundFilename != "" {
		return nil
	}

	// TODO
	index := 1
	this.backgroundFilename = this.backgroundList[index]
	var err error
	this.backgroundImage, err = render.LoadImage(settings, mods, this.backgroundFilename) // +1
	if err != nil {
		return err
	}

	this.RefreshBackground(settings, eset)

	return nil
}

func (this *Switcher) RefreshBackground(settings common.Settings, eset common.EngineSettings) error {
	if this.backgroundImage == nil {
		return nil
	}

	this.backgroundImage.Ref() // old +1 保留

	bkw, err := this.backgroundImage.GetWidth()
	if err != nil {
		return err
	}
	bkh, err := this.backgroundImage.GetHeight()
	if err != nil {
		return err
	}

	dest := utils.ResizeToScreen(settings, eset, bkw, bkh, true, define.ALIGN_CENTER)

	resized, err := this.backgroundImage.Resize(dest.W, dest.H) // old -1 //  new +1
	if err != nil {
		return err
	}
	defer resized.UnRef() // new -1

	if this.background != nil {
		this.background.Close()
		this.background = nil
	}
	this.background, err = resized.CreateSprite() // new +1
	if err != nil {
		return err
	}

	this.background.SetDestFromRect(dest)

	return nil
}

func (this *Switcher) FreeBackground() {
	if this.background != nil {
		this.background.Close()
		this.background = nil
	}

	if this.backgroundImage != nil {
		this.backgroundImage.UnRef()
		this.backgroundImage = nil
	}

	this.backgroundFilename = ""
}

func (this *Switcher) Render(modules common.Modules) error {
	render := modules.Render()
	tooltipm := modules.Tooltipm()
	curs := render.Curs()
	settings := modules.Settings()
	eset := modules.Eset()
	font := modules.Font()
	inpt := modules.Inpt()

	if this.background != nil && this.currentState.GetHasBackground() {
		err := render.Render(this.background)
		if err != nil {
			return err
		}
	}
	this.currentState.Render(modules, this.gameRes)

	err := tooltipm.Render(settings, eset, font, render)
	if err != nil {
		return err
	}

	err = curs.Render(settings, render, inpt)
	if err != nil {
		return err
	}

	return nil
}

func (this *Switcher) IsLoadingFrame() bool {
	if this.currentState == nil {
		return false
	}

	if this.currentState.GetLoadCounter() > 0 {
		this.currentState.DecrLoadCounter()
		return true
	}

	return false
}

func (this *Switcher) GetDone() bool {
	return this.done
}
