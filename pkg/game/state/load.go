package state

import (
	"fmt"
	"math"
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define"
	"monster/pkg/common/define/fontengine"
	"monster/pkg/common/define/inputstate"
	"monster/pkg/common/define/widget/button"
	"monster/pkg/common/define/widget/scrollbar"
	"monster/pkg/common/gameres"
	"monster/pkg/common/labelinfo"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/common/timer"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/game/base"
	"monster/pkg/utils"
	"monster/pkg/utils/parsing"
	"monster/pkg/utils/tools"
	"sort"
	"strconv"
)

type Slot struct {
	id               uint
	stats            gameres.StatBlock
	currentMap       string // 当前地图
	timePlayed       uint64
	equipped         []int
	preview          gameres.GameSlotPreview
	previewTurnTimer timer.Timer
	labelName        common.WidgetLabel
	labelLevel       common.WidgetLabel
	labelClass       common.WidgetLabel
	labelMap         common.WidgetLabel
	labelSlotNumber  common.WidgetLabel
}

func NewSlot(modules common.Modules, gameRes gameres.GameRes) *Slot {
	s := &Slot{}
	s.init(modules, gameRes)
	return s
}

func (this *Slot) init(modules common.Modules, gameRes gameres.GameRes) {
	settings := modules.Settings()
	widgetf := modules.Widgetf()

	resf := gameRes.Resf()

	this.stats = resf.New("statblock").(gameres.StatBlock).Init(modules, resf)

	this.preview = resf.New("gameslotpreview").(gameres.GameSlotPreview).Init(modules, gameRes)

	this.previewTurnTimer = timer.Construct((uint)(settings.Get("max_fps").(int) / 2))
	this.previewTurnTimer.Reset(timer.BEGIN)

	this.labelName = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.labelLevel = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.labelClass = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.labelMap = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.labelSlotNumber = widgetf.New("label").(common.WidgetLabel).Init(modules)
}

func (this *Slot) Close(modules common.Modules) {
	if this.stats != nil {
		this.stats.Close()
		this.stats = nil
	}

	if this.preview != nil {
		this.preview.Close(modules)
		this.preview = nil
	}

	if this.labelName != nil {
		this.labelName.Close()
		this.labelName = nil
	}
	if this.labelLevel != nil {
		this.labelLevel.Close()
		this.labelLevel = nil
	}
	if this.labelClass != nil {
		this.labelClass.Close()
		this.labelClass = nil
	}
	if this.labelMap != nil {
		this.labelMap.Close()
		this.labelMap = nil
	}
	if this.labelSlotNumber != nil {
		this.labelSlotNumber.Close()
		this.labelSlotNumber = nil
	}

}

type Load struct {
	base.State

	buttonExit     common.WidgetButton
	buttonNew      common.WidgetButton
	buttonLoad     common.WidgetButton
	buttonDelete   common.WidgetButton
	labelLoading   common.WidgetLabel
	scrollbar      common.WidgetScrollBar
	confirm        gameres.MenuConfirm
	background     common.Sprite
	selection      common.Sprite
	portraitBorder common.Sprite
	portrait       common.Sprite // 头像
	slotPos        []rect.Rect   // 每个存档位置
	gameSlots      []*Slot       // 全部存档

	loadingRequested bool
	loading          bool
	loaded           bool
	deleteItems      bool
	portraitDest     rect.Rect           // 头像显示位置
	gameSlotPos      rect.Rect           // 存档显示位置
	namePos          labelinfo.LabelInfo // 名字信息 相对存档位置
	levelPos         labelinfo.LabelInfo // 等级信息 相对存档位置
	classPos         labelinfo.LabelInfo // 角色信息 相对存档位置
	mapPos           labelinfo.LabelInfo // 地图信息 相对存档位置
	slotNumberPos    labelinfo.LabelInfo // 存档编号 相对存档位置
	spritesPos       point.Point         // 存档角色预览位置
	selectedSlot     int                 // 选择的存档序号
	visibleSlots     int                 // 当前可见的存档数
	scrollOffset     int                 // 滚动值
	hasScrollBar     bool                // 是否需要滚动条
	gameSlotMax      int                 // 最大可视的存档数
	textTrimBoundary int                 // 文字变成省略号的右则宽 相对存档位置
	portraitAlign    int                 // 头像水平对齐
	gameSlotAlign    int
}

func NewLoad(modules common.Modules, gameRes gameres.GameRes) *Load {
	load := &Load{}

	load.init(modules, gameRes)

	return load
}

func (this *Load) init(modules common.Modules, gameRes gameres.GameRes) gameres.GameStateLoad {
	msg := modules.Msg()
	widgetf := modules.Widgetf()
	mods := modules.Mods()
	render := modules.Render()
	settings := modules.Settings()

	menuf := gameRes.Menuf()
	items := gameRes.Items()
	stats := gameRes.Stats()

	// base
	this.State = base.ConstructState(modules)

	// self
	this.deleteItems = true
	this.selectedSlot = -1
	this.gameSlotMax = 4
	this.portraitAlign = define.ALIGN_FRAME_TOPLEFT
	this.gameSlotAlign = define.ALIGN_FRAME_TOPLEFT

	//TODO item manager
	if items == nil {
		items = gameRes.NewItems(modules, stats)
	}

	this.labelLoading = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.confirm = menuf.New("confirm").(gameres.MenuConfirm).Init(modules, msg.Get("Delete Save"), msg.Get("Delete this save?"))
	this.buttonExit = widgetf.New("button").(common.WidgetButton).Init(modules, button.DEFAULT_FILE)
	this.buttonExit.SetLabel(modules, msg.Get("Exit to Title"))

	this.buttonNew = widgetf.New("button").(common.WidgetButton).Init(modules, button.DEFAULT_FILE)
	this.buttonNew.SetLabel(modules, msg.Get("New Game"))
	this.buttonNew.SetEnabled(true)

	this.buttonLoad = widgetf.New("button").(common.WidgetButton).Init(modules, button.DEFAULT_FILE)
	this.buttonLoad.SetLabel(modules, msg.Get("Choose a Slot"))
	this.buttonLoad.SetEnabled(false)

	this.buttonDelete = widgetf.New("button").(common.WidgetButton).Init(modules, button.DEFAULT_FILE)
	this.buttonDelete.SetLabel(modules, msg.Get("Delete Save"))
	this.buttonDelete.SetEnabled(false)

	this.scrollbar = widgetf.New("scrollbar").(common.WidgetScrollBar).Init(modules, scrollbar.DEFAULT_FILE)

	//TODO tablist

	this.buttonNew.SetAlignment(define.ALIGN_FRAME_TOPLEFT)
	this.buttonLoad.SetAlignment(define.ALIGN_FRAME_TOPLEFT)
	this.buttonDelete.SetAlignment(define.ALIGN_FRAME_TOPLEFT)

	infile := fileparser.New()

	err := infile.Open("menus/gameload.txt", true, mods)
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	for infile.Next(mods) {
		switch infile.Key() {
		case "button_new":
			fallthrough
		case "button_load":
			fallthrough
		case "button_delete":
			fallthrough
		case "button_exit":
			first, strVal := "", ""
			x, y, a := 0, 0, 0
			x, strVal = parsing.PopFirstInt(infile.Val(), "")
			y, strVal = parsing.PopFirstInt(strVal, "")
			first, strVal = parsing.PopFirstString(strVal, "")
			if infile.Key() == "button_exit" {
				a = parsing.ToAlignment(first, define.ALIGN_TOPLEFT)
				this.buttonExit.SetPosBase(x, y, a)
			} else {
				a = parsing.ToAlignment(first, define.ALIGN_FRAME_TOPLEFT)
				if infile.Key() == "button_new" {
					this.buttonNew.SetPosBase(x, y, a)
				} else if infile.Key() == "button_load" {
					this.buttonLoad.SetPosBase(x, y, a)
				} else if infile.Key() == "button_delete" {
					this.buttonDelete.SetPosBase(x, y, a)
				}
			}

		case "portrait":
			this.portraitDest = parsing.ToRect(infile.Val())
		case "gameslot":
			this.gameSlotPos = parsing.ToRect(infile.Val())
		case "name":
			this.namePos = parsing.PopLabelInfo(infile.Val())
		case "level":
			this.levelPos = parsing.PopLabelInfo(infile.Val())
		case "class":
			this.classPos = parsing.PopLabelInfo(infile.Val())
		case "map":
			this.mapPos = parsing.PopLabelInfo(infile.Val())
		case "slot_number":
			this.slotNumberPos = parsing.PopLabelInfo(infile.Val())
		case "loading_label":
			this.labelLoading.SetFromLabelInfo(parsing.PopLabelInfo(infile.Val()))
		case "sprite":
			this.spritesPos = parsing.ToPoint(infile.Val())
		case "visible_slots":
			this.gameSlotMax = parsing.ToInt(infile.Val(), 0)
			// 可展现的存档必须大于1
			this.gameSlotMax = (int)(math.Max((float64)(this.gameSlotMax), 1))
		case "text_trim_boundary":
			// 文字开始变成省略号的右侧宽
			this.textTrimBoundary = parsing.ToInt(infile.Val(), 0)
		default:
			fmt.Printf("GameStateLoad: '%s' is not a valid key.", infile.Key())
			panic("bad key")
		}
	}

	if this.textTrimBoundary == 0 || this.textTrimBoundary > this.gameSlotPos.W {
		this.textTrimBoundary = this.gameSlotPos.W
	}

	// 刷新按钮的标题
	this.buttonNew.Refresh(modules)
	this.buttonLoad.Refresh(modules)
	this.buttonDelete.Refresh(modules)

	// 加载图片创建精灵
	this.loadGraphics(modules)

	// 加载游戏存档
	err = this.readGameSlots(modules, gameRes)
	if err != nil {
		panic(err)
	}

	// 保存地址
	this.refreshSavePaths()

	// 更新头像，存档和滚动条
	this.RefreshWidgets(modules, gameRes)

	// 更新按钮状态
	this.UpdateButtons(modules, gameRes)

	// 命令行指定存档
	if settings.GetLoadSlot() != "" {
		// TODO
	} else {
		// 之前加载过的存档序号
		if settings.Get("prev_save_slot").(int) >= 0 && settings.Get("prev_save_slot").(int) < len(this.gameSlots) {
			this.SetSelectedSlot(settings.Get("prev_save_slot").(int))

			// 计算滚动量
			this.ScrollToSelected()

			// 加载头像，更新按钮和滚动条
			this.UpdateButtons(modules, gameRes)
		}
	}

	// 设置背景透明
	render.SetBackgroundColor(color.Construct(0, 0, 0, 0))

	return this
}

func (this *Load) Clear(modules common.Modules, gameRes gameres.GameRes) {
	if this.buttonExit != nil {
		this.buttonExit.Close()
		this.buttonExit = nil
	}

	if this.buttonNew != nil {
		this.buttonNew.Close()
		this.buttonNew = nil
	}

	if this.buttonLoad != nil {
		this.buttonLoad.Close()
		this.buttonLoad = nil
	}

	if this.buttonDelete != nil {
		this.buttonDelete.Close()
		this.buttonDelete = nil
	}

	if this.scrollbar != nil {
		this.scrollbar.Close()
		this.scrollbar = nil
	}

	if this.labelLoading != nil {
		this.labelLoading.Close()
		this.labelLoading = nil
	}

	if this.confirm != nil {
		this.confirm.Close()
		this.confirm = nil
	}

	if this.background != nil {
		this.background.Close()
		this.background = nil
	}

	if this.selection != nil {
		this.selection.Close()
		this.selection = nil
	}

	if this.portraitBorder != nil {
		this.portraitBorder.Close()
		this.portraitBorder = nil
	}

	if this.portrait != nil {
		this.portrait.Close()
		this.portrait = nil
	}

	for _, ptr := range this.gameSlots {
		ptr.Close(modules)
	}
	this.gameSlots = nil
}

func (this *Load) Close(modules common.Modules, gameRes gameres.GameRes) {
	this.State.Close(modules, gameRes, this)
}

func (this *Load) SetSelectedSlot(slot int) {

	// 重置状态
	if this.selectedSlot != -1 && this.selectedSlot < len(this.gameSlots) && this.gameSlots[this.selectedSlot] != nil {
		this.gameSlots[this.selectedSlot].stats.SetDirection(6)
		this.gameSlots[this.selectedSlot].previewTurnTimer.Reset(timer.BEGIN)
		this.gameSlots[this.selectedSlot].preview.SetAnimation("stance")

	}

	// 选择
	if slot != -1 && slot < len(this.gameSlots) && this.gameSlots[slot] != nil {
		this.gameSlots[slot].stats.SetDirection(6)
		this.gameSlots[slot].previewTurnTimer.Reset(timer.BEGIN)
		this.gameSlots[slot].preview.SetAnimation("run")
	}

	this.selectedSlot = slot
}

// 计算滚动量
func (this *Load) ScrollToSelected() {
	if this.visibleSlots == 0 {
		return
	}

	// 滚动的量  =  选择的序号 - 在单次展现的visibleSlots个存档里的位置
	this.scrollOffset = this.selectedSlot - (this.selectedSlot % this.visibleSlots)

	// 可滚动范围： 最小为0, 最大为总存档数 - 可展现
	if this.scrollOffset < 0 {
		this.scrollOffset = 0
	} else if this.scrollOffset > (len(this.gameSlots) - this.visibleSlots) {

		// 过大会被下层改成滚动条的max
		this.scrollOffset = this.selectedSlot - this.visibleSlots + 1
	}
}

func (this *Load) loadGraphics(modules common.Modules) error {
	mods := modules.Mods()
	render := modules.Render()
	settings := modules.Settings()

	// 背景
	graphics1, err := render.LoadImage(settings, mods, "images/menus/game_slots.png")
	if err != nil {
		return err
	}
	defer graphics1.UnRef()

	this.background, err = graphics1.CreateSprite()
	if err != nil {
		return err
	}

	// 选择框
	graphics2, err := render.LoadImage(settings, mods, "images/menus/game_slot_select.png")
	if err != nil {
		return err
	}
	defer graphics2.UnRef()

	this.selection, err = graphics2.CreateSprite()
	if err != nil {
		return err
	}

	// 角色边框
	graphics3, err := render.LoadImage(settings, mods, "images/menus/portrait_border.png")
	if err != nil {
		return err
	}
	defer graphics3.UnRef()

	this.portraitBorder, err = graphics3.CreateSprite()
	if err != nil {
		return err
	}

	// 设置源位置
	this.portraitBorder.SetClip(0, 0, this.portraitDest.W, this.portraitDest.H)
	return nil
}

func (this *Load) Logic(modules common.Modules, gameRes gameres.GameRes) error {
	inpt := modules.Inpt()

	if inpt.GetWindowResized() {
		// 更新组件
		this.RefreshWidgets(modules, gameRes)
	}

	for i, ptr := range this.gameSlots {
		if ptr == nil {
			continue
		}

		if i == this.selectedSlot {
			ptr.previewTurnTimer.Tick()
			if ptr.previewTurnTimer.IsEnd() {
				ptr.previewTurnTimer.Reset(timer.BEGIN)
				dir := ptr.stats.GetDirection()
				dir++
				if dir > 7 {
					ptr.stats.SetDirection(0)
				} else {
					ptr.stats.SetDirection(dir)
				}
			}
		}

		ptr.preview.Logic()
	}

	if this.confirm.GetVisible() {
	} else {
		if this.buttonExit.CheckClick(modules) ||
			inpt.GetPressing(inputstate.CANCEL) && !inpt.GetLock(inputstate.CANCEL) {
			inpt.SetLock(inputstate.CANCEL, true)
			this.ShowLoading(modules)
			this.SetRequestedGameState(modules, gameRes, NewTitle(modules, gameRes))
			//this.SetRequestedGameState(modules, NewLoad(modules, gameRes))
		}

		if this.buttonNew.CheckClick(modules) {
			this.ShowLoading(modules)
			newGame := NewNewGame(modules, gameRes)
			_ = newGame
			//TODO

			this.SetRequestedGameState(modules, gameRes, newGame)

		} else if this.buttonLoad.CheckClick(modules) {
		} else if this.buttonDelete.CheckClick(modules) {
		} else if len(this.gameSlots) > 0 {
			scrollArea := this.slotPos[0]
			scrollArea.H = this.slotPos[0].H * this.gameSlotMax

			if utils.IsWithinRect(scrollArea, inpt.GetMouse()) {
				if inpt.GetPressing(inputstate.MAIN1) && !inpt.GetLock(inputstate.MAIN1) {
					for i := 0; i < this.visibleSlots; i++ {
						if utils.IsWithinRect(this.slotPos[i], inpt.GetMouse()) {
							inpt.SetLock(inputstate.MAIN1, true)
							this.SetSelectedSlot(i + this.scrollOffset)
							err := this.UpdateButtons(modules, gameRes)
							if err != nil {
								return err
							}
							break
						}
					}
				}
			}
		}
	}

	return nil
}

func (this *Load) Render(modules common.Modules, gameRes gameres.GameRes) error {
	render := modules.Render()
	font := modules.Font()
	msg := modules.Msg()
	settings := modules.Settings()
	eset := modules.Eset()

	src := rect.Construct()
	dest := rect.Construct()

	// 头像
	if this.selectedSlot >= 0 && this.portrait != nil && this.portraitBorder != nil {
		err := render.Render(this.portrait)
		if err != nil {
			return err
		}
		dest.X = this.portrait.GetDest().X
		dest.Y = this.portrait.GetDest().Y
		this.portraitBorder.SetDestFromRect(dest)
		err = render.Render(this.portraitBorder)
		if err != nil {
			return err
		}
	}

	// 加载文字
	if this.loadingRequested || this.loading || this.loaded {
		if this.loaded {
			this.labelLoading.SetText(msg.Get("Entering game world..."))
		} else {
			this.labelLoading.SetText(msg.Get("Loading saved game..."))
		}

		this.labelLoading.SetPos1(modules,
			settings.GetViewW()-eset.Get("resolutions", "menu_frame_width").(int)/2,
			settings.GetViewH()-eset.Get("resolutions", "menu_frame_height").(int)/2)
		this.labelLoading.SetColor(font.GetColor(fontengine.COLOR_MENU_NORMAL))
		err := this.labelLoading.Render(modules)
		if err != nil {
			return err
		}
	}

	for slot := 0; slot < this.visibleSlots; slot++ {
		offSlot := slot + this.scrollOffset
		if this.background != nil {
			src.X = 0
			src.Y = (offSlot % 4) * this.gameSlotPos.H
			src.W = this.gameSlotPos.W
			src.H = this.gameSlotPos.H

			dest.X = this.slotPos[slot].X
			dest.Y = this.slotPos[slot].Y
			dest.W = 0
			dest.H = 0

			this.background.SetClipFromRect(src)
			this.background.SetDestFromRect(dest)
			err := render.Render(this.background)
			if err != nil {
				return err
			}
		}
		slotDest := this.background.GetDest()

		if this.gameSlots[offSlot] == nil {
			// TODO 存档不存在
		}

		// 角色名
		nameColor := color.Construct()
		if this.gameSlots[offSlot].stats.GetPermadeath() {
			nameColor = font.GetColor(fontengine.COLOR_HARDCORE_NAME)
		} else {
			nameColor = font.GetColor(fontengine.COLOR_MENU_NORMAL)
		}

		this.gameSlots[offSlot].labelName.SetPos1(modules, this.slotPos[slot].X, this.slotPos[slot].Y)
		this.gameSlots[offSlot].labelName.SetText(this.gameSlots[offSlot].stats.GetName())
		this.gameSlots[offSlot].labelName.SetColor(nameColor)

		lnBounds := this.gameSlots[offSlot].labelName.GetBounds(modules)
		if this.textTrimBoundary > 0 && lnBounds.X+lnBounds.W >= this.textTrimBoundary+slotDest.X {
			this.gameSlots[offSlot].labelName.SetMaxWidth(this.textTrimBoundary - (lnBounds.X - slotDest.X))
		}

		err := this.gameSlots[offSlot].labelName.Render(modules)
		if err != nil {
			return err
		}

		// 等级
		levelStr := msg.Get(fmt.Sprintf("Level %d", this.gameSlots[offSlot].stats.GetLevel())) +
			" / " + strconv.FormatInt(int64(this.gameSlots[offSlot].timePlayed), 10)

		if this.gameSlots[offSlot].stats.GetPermadeath() {
			levelStr += " / +"
		}

		this.gameSlots[offSlot].labelLevel.SetPos1(modules, this.slotPos[slot].X, this.slotPos[slot].Y)
		this.gameSlots[offSlot].labelLevel.SetText(levelStr)
		this.gameSlots[offSlot].labelLevel.SetColor(font.GetColor(fontengine.COLOR_MENU_NORMAL))

		lvBounds := this.gameSlots[offSlot].labelLevel.GetBounds(modules)
		if this.textTrimBoundary > 0 && lvBounds.X+lvBounds.W >= this.textTrimBoundary+slotDest.X {
			this.gameSlots[offSlot].labelLevel.SetMaxWidth(this.textTrimBoundary - (lvBounds.X - slotDest.X))
		}

		err = this.gameSlots[offSlot].labelLevel.Render(modules)
		if err != nil {
			return err
		}

		// 职业
		this.gameSlots[offSlot].labelClass.SetPos1(modules, this.slotPos[slot].X, this.slotPos[slot].Y)
		this.gameSlots[offSlot].labelClass.SetText(this.gameSlots[offSlot].stats.GetLongClass(modules))
		this.gameSlots[offSlot].labelClass.SetColor(font.GetColor(fontengine.COLOR_MENU_NORMAL))

		lcBounds := this.gameSlots[offSlot].labelClass.GetBounds(modules)
		if this.textTrimBoundary > 0 && lcBounds.X+lcBounds.W >= this.textTrimBoundary+slotDest.X {
			this.gameSlots[offSlot].labelClass.SetMaxWidth(this.textTrimBoundary - (lcBounds.X - slotDest.X))
		}

		err = this.gameSlots[offSlot].labelClass.Render(modules)
		if err != nil {
			return err
		}

		// 地图
		this.gameSlots[offSlot].labelMap.SetPos1(modules, this.slotPos[slot].X, this.slotPos[slot].Y)
		this.gameSlots[offSlot].labelMap.SetText(this.gameSlots[offSlot].currentMap)
		this.gameSlots[offSlot].labelMap.SetColor(font.GetColor(fontengine.COLOR_MENU_NORMAL))

		lmBounds := this.gameSlots[offSlot].labelMap.GetBounds(modules)
		if this.textTrimBoundary > 0 && lmBounds.X+lmBounds.W >= this.textTrimBoundary+slotDest.X {
			this.gameSlots[offSlot].labelMap.SetMaxWidth(this.textTrimBoundary - (lmBounds.X - slotDest.X))
		}

		err = this.gameSlots[offSlot].labelMap.Render(modules)
		if err != nil {
			return err
		}

		// 角色预览
		dest.X = this.slotPos[slot].X + this.spritesPos.X
		dest.Y = this.slotPos[slot].Y + this.spritesPos.Y
		this.gameSlots[offSlot].preview.SetPos(point.Construct(dest.X, dest.Y))
		err = this.gameSlots[offSlot].preview.Render(modules)
		if err != nil {
			return err
		}

		// 存档号
		slotNumberStr := "#" + strconv.FormatInt(int64(offSlot+1), 10)

		this.gameSlots[offSlot].labelSlotNumber.SetPos1(modules, this.slotPos[slot].X, this.slotPos[slot].Y)
		this.gameSlots[offSlot].labelSlotNumber.SetText(slotNumberStr)
		this.gameSlots[offSlot].labelSlotNumber.SetColor(font.GetColor(fontengine.COLOR_MENU_NORMAL))

		lsnBounds := this.gameSlots[offSlot].labelSlotNumber.GetBounds(modules)
		if this.textTrimBoundary > 0 && lsnBounds.X+lsnBounds.W >= this.textTrimBoundary+slotDest.X {
			this.gameSlots[offSlot].labelSlotNumber.SetMaxWidth(this.textTrimBoundary - (lsnBounds.X - slotDest.X))
		}

		err = this.gameSlots[offSlot].labelSlotNumber.Render(modules)
		if err != nil {
			return err
		}
	}

	// 选择框
	if this.selectedSlot >= this.scrollOffset && this.selectedSlot < this.visibleSlots && this.selection != nil {
		this.selection.SetDestFromRect(this.slotPos[this.selectedSlot-this.scrollOffset])
		render.Render(this.selection)
	}

	// 滚动条
	if this.hasScrollBar {
		err := this.scrollbar.Render(modules)
		if err != nil {
			return err
		}
	}

	if this.confirm.GetVisible() {
		//TODO
	}

	err := this.buttonExit.Render(modules)
	if err != nil {
		return err
	}

	err = this.buttonNew.Render(modules)
	if err != nil {
		return err
	}

	err = this.buttonLoad.Render(modules)
	if err != nil {
		return err
	}

	err = this.buttonDelete.Render(modules)
	if err != nil {
		return err
	}

	return nil
}

// 更新滚动条
func (this *Load) refreshScrollBar(modules common.Modules) {
	this.hasScrollBar = len(this.gameSlots) > this.gameSlotMax

	if this.hasScrollBar {
		scrollPos := rect.Construct()

		scrollPos.X = this.slotPos[0].X + this.slotPos[0].W
		scrollPos.Y = this.slotPos[0].Y

		// 按钮之间的距离 = 最大展示高度 - 向下按钮的高
		scrollPos.H = (this.slotPos[0].H * this.gameSlotMax) - this.scrollbar.GetPosDown().H
		this.scrollbar.Refresh(modules, scrollPos.X, scrollPos.Y, scrollPos.H, this.scrollOffset, len(this.gameSlots)-this.visibleSlots)
	}
}

// 更新头像，存档和滚动条
func (this *Load) RefreshWidgets(modules common.Modules, gameRes gameres.GameRes) error {
	settings := modules.Settings()
	eset := modules.Eset()

	// 不偏移
	this.buttonExit.SetPos1(modules, 0, 0)
	this.buttonNew.SetPos1(modules, 0, 0)
	this.buttonLoad.SetPos1(modules, 0, 0)
	this.buttonDelete.SetPos1(modules, 0, 0)

	// 调整头像位置
	if this.portrait != nil {
		// 限制在menu范围内
		portraitRect := utils.AlignToScreenEdge(settings, eset, this.portraitAlign, this.portraitDest)
		this.portrait.SetDest(portraitRect.X, portraitRect.Y)
	}

	// 每个存档的位置
	this.slotPos = make([]rect.Rect, this.visibleSlots)
	for i, _ := range this.slotPos {
		tmpRect := rect.Construct(
			this.gameSlotPos.X,
			this.gameSlotPos.Y+(i*this.gameSlotPos.H),
			this.gameSlotPos.W,
			this.gameSlotPos.H,
		)

		// 限制在menu范围内 (menu_frame)
		this.slotPos[i] = utils.AlignToScreenEdge(settings, eset, this.gameSlotAlign, tmpRect)
	}

	// 更新滚动条
	this.refreshScrollBar(modules)
	this.confirm.Align(modules)
	return nil
}

// 加载存档对应的头像
func (this *Load) loadPortrait(modules common.Modules, slot int) error {
	render := modules.Render()
	settings := modules.Settings()
	mods := modules.Mods()

	if this.portrait != nil {
		this.portrait.Close()
		this.portrait = nil
	}

	if slot < 0 || slot >= len(this.gameSlots) || this.gameSlots[slot] == nil {
		return nil
	}

	if this.gameSlots[slot].stats.GetName() == "" {
		return nil
	}

	graphics, err := render.LoadImage(settings, mods, this.gameSlots[slot].stats.GetGfxPortrait())
	if err != nil {
		return err
	}
	defer graphics.UnRef()
	this.portrait, err = graphics.CreateSprite()
	if err != nil {
		return err
	}
	this.portrait.SetClip(0, 0, this.portraitDest.W, this.portraitDest.H)

	return nil
}

//
func (this *Load) getMapName(modules common.Modules, mapFilename string) (string, error) {
	mods := modules.Mods()
	msg := modules.Msg()

	infile := fileparser.New()
	err := infile.Open(mapFilename, true, mods)
	if err != nil {
		return "", err
	}
	defer infile.Close()

	for infile.Next(mods) {
		if infile.Key() == "title" {
			return msg.Get(infile.Val()), nil
		}
	}

	return "", nil
}

// 加载预览
func (this *Load) loadPreview(modules common.Modules, slot *Slot) error {
	mods := modules.Mods()
	settings := modules.Settings()

	if slot == nil {
		return nil
	}

	var imgGfx []string

	// 每个图层的名字
	previewLayer := slot.preview.GetLayerReferenceOrder()
	for _, val := range previewLayer {
		_, err := mods.Locate(settings, "animations/avatar/"+slot.stats.GetGfxBase()+"/default_"+val+".txt")
		if err != nil && !utils.IsNotExist(err) {
			return err
		} else if err == nil {
			imgGfx = append(imgGfx, "default_"+val)
		} else if val == "head" {
			imgGfx = append(imgGfx, slot.stats.GetGfxHead())
		} else {
			imgGfx = append(imgGfx, "")
		}
	}

	// TODO
	// 加载装备

	slot.preview.LoadGraphics(modules, imgGfx)

	return nil
}

// 读档
func (this *Load) readGameSlots(modules common.Modules, gameRes gameres.GameRes) error {
	settings := modules.Settings()
	eset := modules.Eset()
	mods := modules.Mods()

	ss := gameRes.Stats()

	saveRoot := settings.GetPathUser() + "saves/" + eset.Get("misc", "save_prefix").(string) + "/"
	var saveDirs []string
	var err error

	//
	saveDirs, err = utils.GetDirList(settings.GetPathUser()+"saves/"+eset.Get("misc", "save_prefix").(string), saveDirs)
	if err != nil {
		return err
	}

	// 存档文件名为从1开始的数字，所以可排序
	sort.Slice(saveDirs, func(i, j int) bool { return parsing.ToInt(saveDirs[i], 0) < parsing.ToInt(saveDirs[j], 0) })
	for i := len(saveDirs); i > 0; i-- {
		if parsing.ToInt(saveDirs[i-1], 0) < 1 {
			saveDirs = tools.EraseStr(saveDirs, i-1)
		}
	}

	// 分配内存
	this.gameSlots = make([]*Slot, len(saveDirs))
	if len(this.gameSlots) < this.gameSlotMax {
		this.visibleSlots = len(this.gameSlots)
	} else {
		this.visibleSlots = this.gameSlotMax
	}

	infile := fileparser.New()

	// 解析存档
	for i, dir := range saveDirs {
		filename := saveRoot + dir + "/avatar.txt"

		err := infile.Open(filename, false, mods)
		if err != nil {
			return err
		}

		defer infile.Close()

		this.gameSlots[i] = NewSlot(modules, gameRes)

		this.gameSlots[i].id = (uint)(parsing.ToInt(saveDirs[i], 0))
		this.gameSlots[i].stats.SetHero(true)
		this.gameSlots[i].labelName.SetFromLabelInfo(this.namePos)             // 名字
		this.gameSlots[i].labelLevel.SetFromLabelInfo(this.levelPos)           // 等级
		this.gameSlots[i].labelClass.SetFromLabelInfo(this.classPos)           // 职业
		this.gameSlots[i].labelMap.SetFromLabelInfo(this.mapPos)               // 地图信息
		this.gameSlots[i].labelSlotNumber.SetFromLabelInfo(this.slotNumberPos) // 存档编号

		for infile.Next(mods) {
			switch infile.Key() {
			case "name":
				this.gameSlots[i].stats.SetName(infile.Val())
			case "class":
				first, strVal := parsing.PopFirstString(infile.Val(), "")
				this.gameSlots[i].stats.SetCharacterClass(first)
				first, strVal = parsing.PopFirstString(strVal, "")
				this.gameSlots[i].stats.SetCharacterSubclass(first)
			case "xp":
				this.gameSlots[i].stats.SetXp(parsing.ToUnsignedLong(infile.Val(), 0))
			case "build":
				lenList := len(eset.Get("primary_stats", "list").([]common.PrimaryStat))
				strVal := infile.Val()
				first := 0

				for j := 0; j < lenList; j++ {
					first, strVal = parsing.PopFirstInt(strVal, "")
					this.gameSlots[i].stats.SetPrimary(j, first)
				}
			case "equipped":
				repeatVal, strVal := parsing.PopFirstString(infile.Val(), "")
				for repeatVal != "" {
					this.gameSlots[i].equipped = append(this.gameSlots[i].equipped, parsing.ToInt(repeatVal, 0))
					repeatVal, strVal = parsing.PopFirstString(strVal, "")
				}
			case "option":
				first, strVal := parsing.PopFirstString(infile.Val(), "")
				this.gameSlots[i].stats.SetGfxBase(first)
				first, strVal = parsing.PopFirstString(strVal, "")
				this.gameSlots[i].stats.SetGfxHead(first)
				first, strVal = parsing.PopFirstString(strVal, "")
				this.gameSlots[i].stats.SetGfxPortrait(first)
			case "spawn":
				// 复活点
				first, _ := parsing.PopFirstString(infile.Val(), "")
				var err error
				this.gameSlots[i].currentMap, err = this.getMapName(modules, first)
				if err != nil {
					return err
				}

			case "permadeath":
				this.gameSlots[i].stats.SetPermadeath(parsing.ToBool(infile.Val()))
			case "time_played":
				this.gameSlots[i].timePlayed = parsing.ToUnsignedLong(infile.Val(), 0)
			}
		}

		this.gameSlots[i].stats.Recalc(modules, ss)
		this.gameSlots[i].stats.SetDirection(6)
		this.gameSlots[i].preview.SetStatBlock(this.gameSlots[i].stats)

		// 加载预览
		this.loadPreview(modules, this.gameSlots[i])
	}

	return nil
}

func (this *Load) refreshSavePaths() {
	//TODO
}

// 加载选择存档的头像，更新按钮状态、文字及滚动条
func (this *Load) UpdateButtons(modules common.Modules, gameRes gameres.GameRes) error {
	mods := modules.Mods()
	settings := modules.Settings()
	msg := modules.Msg()

	err := this.loadPortrait(modules, this.selectedSlot) // 加载头像
	if err != nil {
		return err
	}

	_, err = mods.Locate(settings, "maps/spawn.txt")
	if err != nil && utils.IsNotExist(err) {

		// 需要加载可玩mod
		this.buttonNew.SetEnabled(false)
		this.buttonNew.SetTooltip(msg.Get("Enable a story mod to continue"))
	} else if err != nil {
		return err
	}

	// 选择了一个的存档
	if this.selectedSlot >= 0 && this.gameSlots[this.selectedSlot] != nil {
		if !this.buttonLoad.GetEnabled() {
			this.buttonLoad.SetEnabled(true)
		}
		this.buttonLoad.SetTooltip("")

		if !this.buttonDelete.GetEnabled() {
			this.buttonDelete.SetEnabled(true)
		}

		// 更改为加载
		this.buttonLoad.SetLabel(modules, msg.Get("Load Game"))
		if this.gameSlots[this.selectedSlot].currentMap == "" {
			_, err := mods.Locate(settings, "maps/spawn.txt")
			if err != nil && utils.IsNotExist(err) {

				// 需要加载可玩mod
				this.buttonLoad.SetEnabled(false)
				this.buttonLoad.SetTooltip(msg.Get("Enable a story mod to continue"))
			} else if err != nil {
				return err
			}
		}
	} else {
		// 未选择存档
		this.buttonLoad.SetLabel(modules, msg.Get("Choose a Slot"))
		this.buttonLoad.SetEnabled(false)
		this.buttonDelete.SetEnabled(false)
	}

	// 更新按钮状态
	this.buttonNew.Refresh(modules)
	this.buttonLoad.Refresh(modules)
	this.buttonDelete.Refresh(modules)

	// 更新存档，滚动条和头像
	this.RefreshWidgets(modules, gameRes)

	return nil
}
