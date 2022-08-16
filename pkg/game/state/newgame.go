package state

import (
	"fmt"
	"math"
	"math/rand"
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define"
	"monster/pkg/common/define/fontengine"
	"monster/pkg/common/define/inputstate"
	"monster/pkg/common/define/widget/button"
	"monster/pkg/common/define/widget/checkbox"
	"monster/pkg/common/define/widget/listbox"
	"monster/pkg/common/gameres"
	"monster/pkg/common/rect"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/game/base"
	"monster/pkg/utils/parsing"
	"monster/pkg/utils/tools"
	"sort"
)

type HeroOption struct {
	base     string
	head     string
	portrait string
	name     string
}

const (
	OPTION_CURRENT = iota
	OPTION_PREV
	OPTION_NEXT
	OPTION_RANDOM
)

type NewGame struct {
	base.State

	heroOptions      []HeroOption // 可用的人物信息
	currentOption    int
	portraitImage    common.Sprite
	portraitBorder   common.Sprite
	buttonExit       common.WidgetButton
	buttonCreate     common.WidgetButton
	buttonNext       common.WidgetButton
	buttonPrev       common.WidgetButton
	buttonRandomize  common.WidgetButton
	labelPortrait    common.WidgetLabel
	labelName        common.WidgetLabel
	buttonPermadeath common.WidgetCheckBox
	labelPermadeath  common.WidgetLabel
	labelClassList   common.WidgetLabel
	classList        common.WidgetListBox

	portraitPos   rect.Rect // 头像显示位置
	showClassList bool
	showRandomize bool
	deleteItems   bool
	randomOption  bool
	randomClass   bool

	allOptions []int // 可用的人物编号
}

func NewNewGame(modules common.Modules, gameRes gameres.GameRes) *NewGame {
	new := &NewGame{}
	new.init(modules, gameRes)

	return new
}

func (this *NewGame) init(modules common.Modules, gameRes gameres.GameRes) gameres.GameStateNewGame {
	widgetf := modules.Widgetf()
	msg := modules.Msg()
	eset := modules.Eset()
	font := modules.Font()
	mods := modules.Mods()
	render := modules.Render()
	settings := modules.Settings()

	// base
	this.State = base.ConstructState(modules)

	// self
	this.showClassList = true
	this.showRandomize = true
	this.deleteItems = true

	this.buttonExit = widgetf.New("button").(common.WidgetButton).Init(modules, button.DEFAULT_FILE)
	this.buttonExit.SetLabel(modules, msg.Get("Cancel"))

	this.buttonCreate = widgetf.New("button").(common.WidgetButton).Init(modules, button.DEFAULT_FILE)
	this.buttonCreate.SetLabel(modules, msg.Get("Create"))
	this.buttonCreate.SetEnabled(false)
	this.buttonCreate.Refresh(modules)

	this.buttonPrev = widgetf.New("button").(common.WidgetButton).Init(modules, "images/menus/buttons/left.png")
	this.buttonNext = widgetf.New("button").(common.WidgetButton).Init(modules, "images/menus/buttons/right.png")

	this.buttonRandomize = widgetf.New("button").(common.WidgetButton).Init(modules, button.DEFAULT_FILE)
	this.buttonRandomize.SetLabel(modules, msg.Get("Randomize"))

	// TODO
	// input name

	this.buttonPermadeath = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	if eset.Get("death_penalty", "permadeath").(bool) {
		this.buttonPermadeath.SetEnabled(false)
		this.buttonPermadeath.SetChecked(true)
	}

	this.classList = widgetf.New("listbox").(common.WidgetListBox).Init(modules, 12, listbox.DEFAULT_FILE)
	this.classList.SetCanDeselect(false)

	this.labelPortrait = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.labelPortrait.SetText(msg.Get("Choose a Portrait"))
	this.labelPortrait.SetColor(font.GetColor(fontengine.COLOR_MENU_NORMAL))

	this.labelName = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.labelName.SetText(msg.Get("Choose a Name"))
	this.labelName.SetColor(font.GetColor(fontengine.COLOR_MENU_NORMAL))

	this.labelPermadeath = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.labelPermadeath.SetText(msg.Get("Permadeath?"))
	this.labelPermadeath.SetColor(font.GetColor(fontengine.COLOR_MENU_NORMAL))

	this.labelClassList = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.labelClassList.SetText(msg.Get("Choose a Class"))
	this.labelClassList.SetColor(font.GetColor(fontengine.COLOR_MENU_NORMAL))

	this.buttonPrev.SetAlignment(define.ALIGN_FRAME_TOPLEFT)
	this.buttonNext.SetAlignment(define.ALIGN_FRAME_TOPLEFT)
	this.buttonPermadeath.SetAlignment(define.ALIGN_FRAME_TOPLEFT)
	this.buttonRandomize.SetAlignment(define.ALIGN_FRAME_TOPLEFT)

	// TODO
	// input name

	this.classList.SetAlignment(define.ALIGN_FRAME_TOPLEFT)

	infile := fileparser.New()

	err := infile.Open("menus/gamenew.txt", true, mods)
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	var x, y, a int
	var first, strVal string
	for infile.Next(mods) {
		switch infile.Key() {
		case "button_prev":
			x, strVal = parsing.PopFirstInt(infile.Val(), "")
			y, strVal = parsing.PopFirstInt(strVal, "")
			first, strVal = parsing.PopFirstString(strVal, "")
			a = parsing.ToAlignment(first, define.ALIGN_FRAME_TOPLEFT)
			this.buttonPrev.SetPosBase(x, y, a)
		case "button_next":
			x, strVal = parsing.PopFirstInt(infile.Val(), "")
			y, strVal = parsing.PopFirstInt(strVal, "")
			first, strVal = parsing.PopFirstString(strVal, "")
			a = parsing.ToAlignment(first, define.ALIGN_FRAME_TOPLEFT)
			this.buttonNext.SetPosBase(x, y, a)
		case "button_exit":
			x, strVal = parsing.PopFirstInt(infile.Val(), "")
			y, strVal = parsing.PopFirstInt(strVal, "")
			first, strVal = parsing.PopFirstString(strVal, "")
			a = parsing.ToAlignment(first, define.ALIGN_TOPLEFT)
			this.buttonExit.SetPosBase(x, y, a)
		case "button_create":
			x, strVal = parsing.PopFirstInt(infile.Val(), "")
			y, strVal = parsing.PopFirstInt(strVal, "")
			first, strVal = parsing.PopFirstString(strVal, "")
			a = parsing.ToAlignment(first, define.ALIGN_TOPLEFT)
			this.buttonCreate.SetPosBase(x, y, a)
		case "button_permadeath":
			x, strVal = parsing.PopFirstInt(infile.Val(), "")
			y, strVal = parsing.PopFirstInt(strVal, "")
			first, strVal = parsing.PopFirstString(strVal, "")
			a = parsing.ToAlignment(first, define.ALIGN_FRAME_TOPLEFT)
			this.buttonPermadeath.SetPosBase(x, y, a)
		case "button_randomize":
			x, strVal = parsing.PopFirstInt(infile.Val(), "")
			y, strVal = parsing.PopFirstInt(strVal, "")
			first, strVal = parsing.PopFirstString(strVal, "")
			a = parsing.ToAlignment(first, define.ALIGN_FRAME_TOPLEFT)
			this.buttonRandomize.SetPosBase(x, y, a)
		case "name_input":
			// TODO
		case "portrait_label":
			this.labelPortrait.SetFromLabelInfo(parsing.PopLabelInfo(infile.Val()))
		case "name_label":
			this.labelName.SetFromLabelInfo(parsing.PopLabelInfo(infile.Val()))
		case "permadeath_label":
			this.labelPermadeath.SetFromLabelInfo(parsing.PopLabelInfo(infile.Val()))
		case "classlist_label":
			this.labelClassList.SetFromLabelInfo(parsing.PopLabelInfo(infile.Val()))
		case "classlist_height":
			x, strVal = parsing.PopFirstInt(infile.Val(), "")
			this.classList.SetHeight(modules, x)
		case "portrait":
			this.portraitPos = parsing.ToRect(infile.Val())
		case "class_list":
			x, strVal = parsing.PopFirstInt(infile.Val(), "")
			y, strVal = parsing.PopFirstInt(strVal, "")
			first, strVal = parsing.PopFirstString(strVal, "")
			a = parsing.ToAlignment(first, define.ALIGN_FRAME_TOPLEFT)
			this.classList.SetPosBase(x, y, a)
		case "show_classlist":
			this.showClassList = parsing.ToBool(infile.Val())
		case "show_randomize":
			this.showRandomize = parsing.ToBool(infile.Val())
		case "random_option":
			this.randomOption = parsing.ToBool(infile.Val())
		case "random_class":
			this.randomClass = parsing.ToBool(infile.Val())
		default:
			panic(fmt.Sprintf("GameStateNew: '%s' is not a valid key.\n", infile.Key()))
		}
	}

	hcList := eset.Get("hero_classes", "list").([]common.HeroClass)
	for _, hc := range hcList {
		this.classList.Append(modules, msg.Get(hc.GetName()), msg.Get(hc.GetDescription()))
	}

	if len(hcList) != 0 {
		classIndex := 0
		if this.randomClass {
			classIndex = rand.Intn(30) % len(hcList)
		}

		// 随机选择一个职业
		this.classList.Select(classIndex)
	}

	// 加载图片
	graphics, err := render.LoadImage(settings, mods, "images/menus/portrait_border.png")
	if err != nil {
		panic(err)
	}
	defer graphics.UnRef()
	this.portraitBorder, err = graphics.CreateSprite()
	if err != nil {
		panic(err)
	}

	// 加载可用人物
	err = this.loadOptions(modules, "hero_options.txt")
	if err != nil {
		panic(err)
	}

	if this.randomOption {
		this.setHeroOption(modules, OPTION_RANDOM)
	} else {
		this.setHeroOption(modules, OPTION_CURRENT)
	}

	err = this.RefreshWidgets(modules, gameRes)
	if err != nil {
		panic(err)
	}

	render.SetBackgroundColor(color.Construct(0, 0, 0, 0))
	return this
}

func (this *NewGame) Clear(modules common.Modules, gameRes gameres.GameRes) {
	if this.portraitImage != nil {
		this.portraitImage.Close()
		this.portraitImage = nil
	}

	if this.portraitBorder != nil {
		this.portraitBorder.Close()
		this.portraitBorder = nil
	}

	if this.buttonExit != nil {
		this.buttonExit.Close()
		this.buttonExit = nil
	}

	if this.buttonCreate != nil {
		this.buttonCreate.Close()
		this.buttonCreate = nil
	}

	if this.buttonNext != nil {
		this.buttonNext.Close()
		this.buttonNext = nil
	}

	if this.buttonPrev != nil {
		this.buttonPrev.Close()
		this.buttonPrev = nil
	}

	if this.buttonRandomize != nil {
		this.buttonRandomize.Close()
		this.buttonRandomize = nil
	}

	if this.labelPortrait != nil {
		this.labelPortrait.Close()
		this.labelPortrait = nil
	}

	if this.labelName != nil {
		this.labelName.Close()
		this.labelName = nil
	}

	if this.buttonPermadeath != nil {
		this.buttonPermadeath.Close()
		this.buttonPermadeath = nil
	}

	if this.labelPermadeath != nil {
		this.labelPermadeath.Close()
		this.labelPermadeath = nil
	}

	if this.labelClassList != nil {
		this.labelClassList.Close()
		this.labelClassList = nil
	}

	if this.classList != nil {
		this.classList.Close()
		this.classList = nil
	}
}

func (this *NewGame) Close(modules common.Modules, gameRes gameres.GameRes) {
	this.State.Close(modules, gameRes, this)
}

func (this *NewGame) setHeroOption(modules common.Modules, dir int) error {
	eset := modules.Eset()
	render := modules.Render()
	mods := modules.Mods()
	settings := modules.Settings()
	// 选择的职业

	changed := false
	availableOptions := this.allOptions

	classIndex, ok := this.classList.GetSelected()
	if ok {
		// 某种职业能用的人物编号
		hcList := eset.Get("hero_classes", "list").([]common.HeroClass)
		heroOptions := hcList[classIndex].GetHeroOptions()
		if classIndex < len(hcList) && len(heroOptions) != 0 {
			availableOptions = heroOptions
			changed = true
		}
	}

	switch dir {
	case OPTION_CURRENT:
		if !tools.FindInt(availableOptions, this.currentOption) {
			if this.randomOption && changed {
				randIndex := rand.Intn(100) % len(availableOptions)
				this.currentOption = availableOptions[randIndex]
			} else {
				this.currentOption = availableOptions[0]
			}
		}
	case OPTION_NEXT:
		index := -1
		for i, val := range availableOptions {
			if val == this.currentOption {
				index = i
				break
			}
		}

		if index == -1 {
			this.currentOption = availableOptions[0]
		} else {
			index++
			if index < len(availableOptions) {
				this.currentOption = availableOptions[index]
			} else {
				this.currentOption = availableOptions[0]
			}
		}

	case OPTION_PREV:
		index := -1
		for i, val := range availableOptions {
			if val == this.currentOption {
				index = i
				break
			}
		}

		if index == 0 {
			this.currentOption = availableOptions[len(availableOptions)-1]
		} else {
			index--
			this.currentOption = availableOptions[index]
		}

	default:
		if len(availableOptions) != 0 {
			randIndex := rand.Intn(100) % len(availableOptions)
			this.currentOption = availableOptions[randIndex]
		}
	}

	// 加载对应的人物的头像
	if this.portraitImage != nil {
		this.portraitImage.Close()
		this.portraitImage = nil
	}

	graphics, err := render.LoadImage(settings, mods, this.heroOptions[this.currentOption].portrait)
	if err != nil {
		return err
	}
	defer graphics.UnRef()

	this.portraitImage, err = graphics.CreateSprite()
	if err != nil {
		return err
	}
	this.portraitImage.SetDestFromRect(this.portraitPos)

	//TODO set name

	return nil
}

// 加载全部可用人物编号
func (this *NewGame) loadOptions(modules common.Modules, filename string) error {
	mods := modules.Mods()
	msg := modules.Msg()

	infile := fileparser.New()

	err := infile.Open("engine/"+filename, true, mods)
	if err != nil {
		return err
	}
	defer infile.Close()

	var rawIndex int
	var strVal, first string
	for infile.Next(mods) {
		if infile.Key() == "option" {
			rawIndex, strVal = parsing.PopFirstInt(infile.Val(), "")
			curIndex := (int)(math.Max(0, float64(rawIndex)))

			if curIndex+1 > len(this.heroOptions) {
				oldOne := this.heroOptions
				this.heroOptions = make([]HeroOption, curIndex+1)

				// copy
				for index, val := range oldOne {
					this.heroOptions[index] = val
				}
			}

			first, strVal = parsing.PopFirstString(strVal, "")
			this.heroOptions[curIndex].base = first
			first, strVal = parsing.PopFirstString(strVal, "")
			this.heroOptions[curIndex].head = first
			first, strVal = parsing.PopFirstString(strVal, "")
			this.heroOptions[curIndex].portrait = first
			first, strVal = parsing.PopFirstString(strVal, "")
			this.heroOptions[curIndex].name = msg.Get(first)

			this.allOptions = append(this.allOptions, curIndex)
		}
	}

	if len(this.heroOptions) == 0 {
		this.heroOptions = make([]HeroOption, 1)
	}

	sort.Slice(this.allOptions, func(i, j int) bool { return this.allOptions[i] < this.allOptions[j] })

	return nil
}

func (this *NewGame) RefreshWidgets(modules common.Modules, gameRes gameres.GameRes) error {
	settings := modules.Settings()
	eset := modules.Eset()

	this.buttonExit.SetPos1(modules, 0, 0)
	this.buttonCreate.SetPos1(modules, 0, 0)
	this.buttonPrev.SetPos1(modules, 0, 0)
	this.buttonNext.SetPos1(modules, 0, 0)
	this.buttonPermadeath.SetPos1(modules, 0, 0)
	this.buttonRandomize.SetPos1(modules, 0, 0)
	this.classList.SetPos1(modules, 0, 0)

	tmpW := (settings.GetViewW() - eset.Get("resolutions", "menu_frame_width").(int)) / 2
	tmpH := (settings.GetViewH() - eset.Get("resolutions", "menu_frame_height").(int)) / 2
	this.labelPortrait.SetPos1(modules, tmpW, tmpH)
	this.labelName.SetPos1(modules, tmpW, tmpH)
	this.labelPermadeath.SetPos1(modules, tmpW, tmpH)
	this.labelClassList.SetPos1(modules, tmpW, tmpH)

	return nil
}

func (this *NewGame) Logic(modules common.Modules, gameRes gameres.GameRes) error {
	inpt := modules.Inpt()
	eset := modules.Eset()

	if inpt.GetWindowResized() {
		this.RefreshWidgets(modules, gameRes)
	}

	// TODO
	// input name

	this.buttonPermadeath.CheckClick(modules)
	if this.showClassList && this.classList.CheckClick(modules) {
		this.setHeroOption(modules, OPTION_CURRENT)
	}

	this.buttonCreate.SetEnabled(true)
	this.buttonCreate.Refresh(modules)

	if inpt.GetPressing(inputstate.CANCEL) && !inpt.GetLock(inputstate.CANCEL) || this.buttonExit.CheckClick(modules) {

		if inpt.GetPressing(inputstate.CANCEL) {
			inpt.SetLock(inputstate.CANCEL, true)
		}

		this.deleteItems = false
		this.ShowLoading(modules)
		this.SetRequestedGameState(modules, gameRes, NewLoad(modules, gameRes))
	}

	if this.buttonCreate.CheckClick(modules) {
		inpt.SetLockAll(true)
		this.deleteItems = false
		this.ShowLoading(modules)
		play := NewPlay(modules, gameRes)
		play.ResetGame(modules, gameRes)

		//TODO play
		this.SetRequestedGameState(modules, gameRes, play)
	}

	if this.buttonNext.CheckClick(modules) {
		this.setHeroOption(modules, OPTION_NEXT)
	} else if this.buttonPrev.CheckClick(modules) {
		this.setHeroOption(modules, OPTION_NEXT)
	}

	if this.showRandomize && this.buttonRandomize.CheckClick(modules) {
		hcList := eset.Get("hero_classes", "list").([]common.HeroClass)
		if len(hcList) != 0 {
			classIndex := rand.Intn(30) % len(hcList)
			this.classList.Select(classIndex)
		}
		this.setHeroOption(modules, OPTION_RANDOM)
	}

	return nil
}

func (this *NewGame) Render(modules common.Modules, gameRes gameres.GameRes) error {
	settings := modules.Settings()
	eset := modules.Eset()
	render := modules.Render()

	err := this.buttonExit.Render(modules)
	if err != nil {
		return err
	}

	err = this.buttonCreate.Render(modules)
	if err != nil {
		return err
	}

	err = this.buttonPrev.Render(modules)
	if err != nil {
		return err
	}

	err = this.buttonNext.Render(modules)
	if err != nil {
		return err
	}

	err = this.buttonPermadeath.Render(modules)
	if err != nil {
		return err
	}

	if this.showRandomize {
		err = this.buttonRandomize.Render(modules)
		if err != nil {
			return err
		}
	}

	src := rect.Construct()
	dest := rect.Construct()

	src.W = this.portraitPos.W
	src.H = this.portraitPos.H
	src.X = 0
	src.Y = 0

	dest.W = src.W
	dest.H = src.H
	dest.X = this.portraitPos.X + (settings.GetViewW()-eset.Get("resolutions", "menu_frame_width").(int))/2
	dest.Y = this.portraitPos.Y + (settings.GetViewH()-eset.Get("resolutions", "menu_frame_height").(int))/2

	if this.portraitImage != nil {
		this.portraitImage.SetClipFromRect(src)
		this.portraitImage.SetDestFromRect(dest)
		err := render.Render(this.portraitImage)
		if err != nil {
			return err
		}

		this.portraitBorder.SetClipFromRect(src)
		this.portraitBorder.SetDestFromRect(dest)
		err = render.Render(this.portraitBorder)
		if err != nil {
			return err
		}
	}

	err = this.labelPortrait.Render(modules)
	if err != nil {
		return err
	}

	err = this.labelName.Render(modules)
	if err != nil {
		return err
	}

	err = this.labelPermadeath.Render(modules)
	if err != nil {
		return err
	}

	if this.showClassList {
		err = this.labelClassList.Render(modules)
		if err != nil {
			return err
		}

		err = this.classList.Render(modules)
		if err != nil {
			return err
		}
	}

	return nil
}
