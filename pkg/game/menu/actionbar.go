package menu

import (
	"fmt"
	"monster/pkg/common"
	"monster/pkg/common/define"
	"monster/pkg/common/define/fontengine"
	"monster/pkg/common/define/game/menu/actionbar"
	"monster/pkg/common/define/game/menu/powers"
	"monster/pkg/common/define/inputstate"
	"monster/pkg/common/gameres"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/game/base"
	"monster/pkg/utils/parsing"
)

type ActionBar struct {
	base.Menu

	spriteEmptySlot   common.Sprite
	spriteDisabled    common.Sprite
	spriteAttention   common.Sprite
	src               rect.Rect
	labels            []string
	menuLabels        []string
	slotFailCooldown  []int
	tooltipLength     int
	slotsCount        int              // 槽的数量
	hotkeys           []define.PowerId // 当前技能槽里的技能
	hotkeysTemp       []define.PowerId
	hotkeysMod        []define.PowerId
	locked            []bool
	preventChanging   []bool
	slots             []common.WidgetSlot // 槽的渲染和位置
	menus             []common.WidgetSlot
	menuTitles        []string
	slotItemCount     []int  // 槽里的数量
	slotEnabled       []bool // 槽里的技能或物品是否可用，比如冷却中，数量不够等
	requiresAttention []bool
	slotActivated     []bool // 槽里是否有技能或物品
	slotCooldownSize  []int

	dragPrevSlot int
	twoStepSlot  int
}

func NewActionBar(modules common.Modules, powerManager gameres.PowerManager) *ActionBar {
	ab := &ActionBar{}
	ab.Init(modules, powerManager)

	return ab
}

func (this *ActionBar) Init(modules common.Modules, powerManager gameres.PowerManager) gameres.MenuActionBar {
	widgetf := modules.Widgetf()
	msg := modules.Msg()
	mods := modules.Mods()
	eset := modules.Eset()
	render := modules.Render()
	settings := modules.Settings()

	// base
	this.Menu = base.ConstructMenu(modules)

	// self
	this.src = rect.Construct()
	this.tooltipLength = powers.TOOLTIP_LONG_MENU
	this.dragPrevSlot = -1
	this.twoStepSlot = -1

	this.menuLabels = make([]string, actionbar.MENU_COUNT)
	this.requiresAttention = make([]bool, actionbar.MENU_COUNT)
	// TODO
	// tablist

	this.menus = make([]common.WidgetSlot, actionbar.MENU_COUNT)

	for i := 0; i < actionbar.MENU_COUNT; i++ {
		this.menus[i] = widgetf.New("slot").(common.WidgetSlot).Init(modules, -1, inputstate.ACTIONBAR)
		err := this.menus[i].SetHotkey(modules, inputstate.CHARACTER+i)
		if err != nil {
			panic(err)
		}

		this.menus[i].SetPosW(0)
		this.menus[i].SetPosH(0)
	}

	this.menuTitles = make([]string, actionbar.MENU_COUNT)
	this.menuTitles[actionbar.MENU_CHARACTER] = msg.Get("Character") // 角色
	this.menuTitles[actionbar.MENU_INVENTORY] = msg.Get("Inventory") // 仓库
	this.menuTitles[actionbar.MENU_POWERS] = msg.Get("Powers")       // 技能
	this.menuTitles[actionbar.MENU_LOG] = msg.Get("Log")             // 日志

	infile := fileparser.New()

	err := infile.Open("menus/actionbar.txt", true, mods)
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	for infile.Next(mods) {
		key := infile.Key()
		val := infile.Val()

		// 本组件的位置等
		if this.Menu.ParseMenuKey(key, val) {
			continue
		}

		var x, y int
		switch key {
		case "slot":

			// 技能槽
			var index int
			index, val = parsing.PopFirstInt(val, "")
			if index == 0 || index > 10 {
				panic("MenuActionBar: Slot index must be in range 1-10.")
			} else {

				// 在父组件的位置
				var first string
				x, val = parsing.PopFirstInt(val, "")
				y, val = parsing.PopFirstInt(val, "")
				first, val = parsing.PopFirstString(val, "")
				isLocked := false
				if first != "" {
					isLocked = parsing.ToBool(first)
				}

				this.addSlot(modules, index-1, x, y, isLocked)
			}
		case "slot_M1":

			var first string
			x, val = parsing.PopFirstInt(val, "")
			y, val = parsing.PopFirstInt(val, "")

			first, val = parsing.PopFirstString(val, "")
			isLocked := false
			if first != "" {
				isLocked = parsing.ToBool(first)
			}

			this.addSlot(modules, 10, x, y, isLocked)

		case "slot_M2":
			var first string
			x, val = parsing.PopFirstInt(val, "")
			y, val = parsing.PopFirstInt(val, "")

			first, val = parsing.PopFirstString(val, "")
			isLocked := false
			if first != "" {
				isLocked = parsing.ToBool(first)
			}

			this.addSlot(modules, 11, x, y, isLocked)
		case "char_menu":
			x, val = parsing.PopFirstInt(val, "")
			y, val = parsing.PopFirstInt(val, "")
			this.menus[actionbar.MENU_CHARACTER].SetPosBase(x, y, define.ALIGN_TOPLEFT)
			this.menus[actionbar.MENU_CHARACTER].SetPosW(eset.Get("resolutions", "icon_size").(int))
		case "inv_menu":
			x, val = parsing.PopFirstInt(val, "")
			y, val = parsing.PopFirstInt(val, "")
			this.menus[actionbar.MENU_INVENTORY].SetPosBase(x, y, define.ALIGN_TOPLEFT)
			this.menus[actionbar.MENU_INVENTORY].SetPosW(eset.Get("resolutions", "icon_size").(int))
		case "powers_menu":
			x, val = parsing.PopFirstInt(val, "")
			y, val = parsing.PopFirstInt(val, "")
			this.menus[actionbar.MENU_POWERS].SetPosBase(x, y, define.ALIGN_TOPLEFT)
			this.menus[actionbar.MENU_POWERS].SetPosW(eset.Get("resolutions", "icon_size").(int))
		case "log_menu":
			x, val = parsing.PopFirstInt(val, "")
			y, val = parsing.PopFirstInt(val, "")
			this.menus[actionbar.MENU_LOG].SetPosBase(x, y, define.ALIGN_TOPLEFT)
			this.menus[actionbar.MENU_LOG].SetPosW(eset.Get("resolutions", "icon_size").(int))
		case "tooltip_length":
			if val == "short" {
				this.tooltipLength = powers.TOOLTIP_SHORT
			} else if val == "long_menu" {
				this.tooltipLength = powers.TOOLTIP_LONG_MENU
			} else if val == "long_all" {
				this.tooltipLength = powers.TOOLTIP_LONG_ALL
			} else {
				panic(fmt.Sprintf("MenuActionBar: '%s' is not a valid tooltip_length setting.\n", val))
			}
		default:
			panic(fmt.Sprintf("MenuActionBar: '%s' is not a valid key.\n", key))
		}
	}

	// TODO
	// tablist

	this.slotsCount = len(this.slots)
	this.hotkeys = make([]define.PowerId, this.slotsCount)
	this.hotkeysTemp = make([]define.PowerId, this.slotsCount)
	this.hotkeysMod = make([]define.PowerId, this.slotsCount)
	this.locked = make([]bool, this.slotsCount)
	this.slotItemCount = make([]int, this.slotsCount)
	this.slotEnabled = make([]bool, this.slotsCount)
	this.slotActivated = make([]bool, this.slotsCount)
	this.slotCooldownSize = make([]int, this.slotsCount)
	this.slotFailCooldown = make([]int, this.slotsCount)

	this.Clear1(powerManager, false)

	// 加载图片
	err = this.Menu.SetBackground(modules, "images/menus/actionbar_trim.png")
	if err != nil {
		panic(err)
	}

	iconClip := rect.Construct()
	iconClip.W = eset.Get("resolutions", "icon_size").(int)
	iconClip.H = iconClip.W

	graphics1, err := render.LoadImage(settings, mods, "images/menus/slot_empty.png")
	if err != nil {
		panic(err)
	}
	defer graphics1.UnRef()

	this.spriteEmptySlot, err = graphics1.CreateSprite()
	if err != nil {
		panic(err)
	}

	this.spriteEmptySlot.SetClipFromRect(iconClip)

	graphics2, err := render.LoadImage(settings, mods, "images/menus/disabled.png")
	if err != nil {
		panic(err)
	}
	defer graphics2.UnRef()

	this.spriteDisabled, err = graphics2.CreateSprite()
	if err != nil {
		panic(err)
	}

	this.spriteDisabled.SetClipFromRect(iconClip)

	graphics3, err := render.LoadImage(settings, mods, "images/menus/attention_glow.png")
	if err != nil {
		panic(err)
	}
	defer graphics3.UnRef()

	this.spriteAttention, err = graphics3.CreateSprite()
	if err != nil {
		panic(err)
	}

	this.spriteAttention.SetClipFromRect(iconClip)

	// TODO
	// sound

	this.Align(modules)
	return this
}

func (this *ActionBar) Clear() {
	if this.spriteEmptySlot != nil {
		this.spriteEmptySlot.Close()
	}

	if this.spriteDisabled != nil {
		this.spriteDisabled.Close()
	}

	if this.spriteAttention != nil {
		this.spriteAttention.Close()
	}

	for _, ptr := range this.menus {
		ptr.Close()
	}

	for _, ptr := range this.slots {
		ptr.Close()
	}
}

func (this *ActionBar) Close() {
	this.Menu.Close(this)
}

func (this *ActionBar) Logic(modules common.Modules, pc gameres.Avatar, powers gameres.PowerManager) error {

	eset := modules.Eset()

	// TODO
	// tablist

	if pc.GetPowerCastTimersSize() == 0 {
		// 没有技能
		return nil
	}

	for i := 0; i < this.slotsCount; i++ {
		if this.slots[i] == nil {
			continue
		}

		if this.hotkeys[i] > 0 {
			power := powers.GetPower(this.hotkeysMod[i])

			if len(power.RequiredItems) == 0 {

				// 技能
				this.SetItemCount(modules, i, -1, false)
			} else {

				// 消耗物品, 吃药
				for j := 0; j < len(power.RequiredItems); j++ {
					// TODO
					// menu inventory

					if power.RequiredItems[j].Quantity > 0 {
						break
					}
				}

			}

			// TODO
			// pc statsblock
			this.slotEnabled[i] = pc.GetPowerCooldownTimer(this.hotkeysMod[i]).IsEnd() &&
				pc.GetPowerCastTimer(this.hotkeysMod[i]).IsEnd() &&
				(this.twoStepSlot == -1 || this.twoStepSlot == i)

			this.slots[i].SetIcon(power.Icon, -1)

		} else {
			this.slotEnabled[i] = true
		}

		// 处理技能释放，技能冷却定时器
		esetIconSize := eset.Get("resolutions", "icon_size").(int)
		if this.hotkeysMod[i] != 0 && !pc.GetPowerCastTimer(this.hotkeysMod[i]).IsEnd() && pc.GetPowerCastTimer(this.hotkeysMod[i]).GetDuration() > 0 {
			this.slotCooldownSize[i] = esetIconSize * (int)(pc.GetPowerCastTimer(this.hotkeysMod[i]).GetDuration()) / (int)(pc.GetPowerCastTimer(this.hotkeysMod[i]).GetDuration())
		} else if this.hotkeysMod[i] != 0 && !pc.GetPowerCooldownTimer(this.hotkeysMod[i]).IsEnd() && pc.GetPowerCooldownTimer(this.hotkeysMod[i]).GetDuration() > 0 {
			this.slotCooldownSize[i] = esetIconSize * (int)(pc.GetPowerCooldownTimer(this.hotkeysMod[i]).GetDuration()) / (int)(pc.GetPowerCooldownTimer(this.hotkeysMod[i]).GetDuration())
		} else {
			this.slotCooldownSize[i] = esetIconSize

			if this.slotEnabled[i] {
				this.slotCooldownSize[i] = 0
			}
		}

		if this.slotFailCooldown[i] > 0 {
			this.slotFailCooldown[i]--
		}
	}

	return nil
}

func (this *ActionBar) Align(modules common.Modules) error {
	msg := modules.Msg()
	inpt := modules.Inpt()
	settings := modules.Settings()

	this.Menu.Align(modules) // 背景，总体位置

	for i := 0; i < this.slotsCount; i++ {
		if this.slots[i] != nil {
			this.slots[i].SetPos1(modules, this.GetWindowArea().X, this.GetWindowArea().Y)
		}
	}

	for i := 0; i < actionbar.MENU_COUNT; i++ {
		this.menus[i].SetPos1(modules, this.GetWindowArea().X, this.GetWindowArea().Y)
	}

	for i := 0; i < actionbar.SLOT_MAIN1; i++ {
		if i < len(this.slots) && this.slots[i] != nil {
			this.labels[i] = msg.Get(fmt.Sprintf("Hotkey: %s", inpt.GetBindingString(msg, i+inputstate.BAR_1, false)))
		}
	}

	for i := actionbar.SLOT_MAIN1; i < actionbar.SLOT_MAX; i++ {
		if i < len(this.slots) && this.slots[i] != nil {
			if settings.Get("mouse_move").(bool) &&
				((i == actionbar.SLOT_MAIN2 && settings.Get("mouse_move_swap").(bool)) || (i == actionbar.SLOT_MAIN1 && !settings.Get("mouse_move_swap").(bool))) {
				this.labels[i] = msg.Get(fmt.Sprintf("Hotkey: %s", inpt.GetBindingString(msg, inputstate.SHIFT, false)+" + "+inpt.GetBindingString(msg, i-actionbar.SLOT_MAIN1+inputstate.MAIN1, false)))
			} else {
				this.labels[i] = msg.Get(fmt.Sprintf("Hotkey: %s", inpt.GetBindingString(msg, i-actionbar.SLOT_MAIN1+inputstate.MAIN1, false)))
			}
		}
	}

	for i := 0; i < len(this.menuLabels); i++ {
		this.menus[i].SetPos1(modules, this.GetWindowArea().X, this.GetWindowArea().Y)
		this.menuLabels[i] = msg.Get(fmt.Sprintf("Hotkey: %s", inpt.GetBindingString(msg, i+inputstate.CHARACTER, false)))
	}

	return nil
}

func (this *ActionBar) addSlot(modules common.Modules, index, x, y int, isLocked bool) {
	widgetf := modules.Widgetf()
	eset := modules.Eset()

	lenSlots := len(this.slots)

	if index >= lenSlots {
		for i := 0; i < index+1-lenSlots; i++ {
			this.labels = append(this.labels, "")
			this.slots = append(this.slots, nil)
		}
	}

	this.slots[index] = widgetf.New("slot").(common.WidgetSlot).Init(modules, -1, inputstate.ACTIONBAR)
	this.slots[index].SetPosBase(x, y, define.ALIGN_TOPLEFT)
	this.slots[index].SetPosW(eset.Get("resolutions", "icon_size").(int))
	this.slots[index].SetPosH(eset.Get("resolutions", "icon_size").(int))
	this.slots[index].SetContinuous(true) // 按住一直释放

	if index < 10 {
		this.slots[index].SetHotkey(modules, inputstate.BAR_1+index)
	} else if index < 12 {
		this.slots[index].SetHotkey(modules, inputstate.MAIN1+index-10)
	}

	this.preventChanging = make([]bool, len(this.slots))
	this.preventChanging[index] = isLocked
}

func (this *ActionBar) Clear1(powers gameres.PowerManager, skipItems bool) {
	for i := 0; i < this.slotsCount; i++ {
		if skipItems && powers != nil {
			if this.hotkeys[i] > 0 {
				power := powers.GetPower(this.hotkeysMod[i])
				if len(power.RequiredItems) != 0 {
					continue
				}
			}
		}

		this.hotkeys[i] = 0
		this.hotkeysTemp[i] = 0
		this.hotkeysMod[i] = 0
		this.slotItemCount[i] = -1
		this.slotEnabled[i] = true
		this.locked[i] = false
		this.slotActivated[i] = false
		this.slotCooldownSize[i] = 0
		this.slotFailCooldown[i] = 0
	}

	for i := 0; i < actionbar.MENU_COUNT; i++ {
		this.requiresAttention[i] = false
	}

	this.twoStepSlot = -1
}

func (this *ActionBar) Render(modules common.Modules) error {
	render := modules.Render()
	eset := modules.Eset()
	settings := modules.Settings()
	widgetf := modules.Widgetf()
	font := modules.Font()

	err := this.Menu.Render(modules)
	if err != nil {
		return err
	}

	esetIconSize := eset.Get("resolutions", "icon_size").(int)

	for i := 0; i < this.slotsCount; i++ {
		if this.slots[i] == nil {
			continue
		}

		if this.hotkeys[i] != 0 {
			// 存在技能

			err := this.slots[i].Render(modules)
			if err != nil {
				return err
			}
		} else {

			// 无技能
			if this.spriteEmptySlot != nil {
				this.spriteEmptySlot.SetDestFromRect(this.slots[i].GetPos())
				err := render.Render(this.spriteEmptySlot)
				if err != nil {
					return err
				}
			}
		}

		// 冷却或禁用图层
		if !this.slotEnabled[i] {
			clip := rect.Construct(0, 0, esetIconSize, esetIconSize)

			if this.twoStepSlot == -1 || this.twoStepSlot == i {
				clip.H = this.slotCooldownSize[i]
			}

			if this.spriteDisabled != nil && clip.H > 0 {
				this.spriteDisabled.SetClipFromRect(clip)
				this.spriteDisabled.SetDestFromRect(this.slots[i].GetPos())
				err := render.Render(this.spriteDisabled)
				if err != nil {
					return err
				}
			}
		}

		err := this.slots[i].RenderSelection(modules)
		if err != nil {
			return err
		}
	}

	for i := 0; i < actionbar.MENU_COUNT; i++ {
		err := this.menus[i].Render(modules)
		if err != nil {
			return err
		}

		if this.requiresAttention[i] && !this.menus[i].GetInFocus() {
			if this.spriteAttention != nil {
				this.spriteAttention.SetDestFromRect(this.menus[i].GetPos())
				err := render.Render(this.spriteAttention)
				if err != nil {
					return err
				}
			}

			if settings.Get("colorblind").(bool) {
				label := widgetf.New("label").(common.WidgetLabel).Init(modules)
				hOffset := eset.Get("widgets", "colorblind_highlight_offset").(point.Point)
				label.SetPos1(modules, this.menus[i].GetPos().X+hOffset.X, this.menus[i].GetPos().Y+hOffset.Y)
				label.SetText("*")
				label.SetColor(font.GetColor(fontengine.COLOR_MENU_NORMAL))
				err := label.Render(modules)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// 给某个槽添加数量
func (this *ActionBar) SetItemCount(modules common.Modules, index, count int, isEquipped bool) {
	if index >= this.slotsCount || this.slots[index] == nil {
		return
	}

	this.slotItemCount[index] = count

	if count == 0 {
		if this.slotActivated[index] {
			this.slots[index].Deactivate()
		}

		this.slotEnabled[index] = false
	}

	if isEquipped {
		this.slots[index].SetAmount(modules, count, 0)
	} else if count >= 0 {
		this.slots[index].SetAmount(modules, count, 2)
	} else {
		// 技能
		this.slots[index].SetAmount(modules, 0, 0)
	}
}
