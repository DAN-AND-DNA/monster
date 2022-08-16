package widget

import (
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define/inputstate"
	"monster/pkg/common/define/widget/slot"
	"monster/pkg/common/labelinfo"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/utils"
	"monster/pkg/utils/tools"
	"monster/pkg/widget/base"
)

type Slot struct {
	base.Widget

	slotSelected  common.Sprite // 选择框
	slotChecked   common.Sprite // 勾选框
	labelAmountBg common.Sprite // 数量标题
	labelHotkeyBg common.Sprite // 快捷键标题
	labelAmount   *Label
	labelHotkey   *Label
	iconId        int // 当前 slot 对应的图标
	overlayId     slot.CLICK_TYPE
	amount        int
	maxAmount     int
	activateKey   int
	hotkey        int
	enabled       bool
	checked       bool // 是否在使用
	pressed       bool // 被按压，（左右鼠标键）
	continuous    bool // 运行按住来持续启动
	visible       bool
}

func NewSlot(modules common.Modules, iconId, activate int) *Slot {
	sl := &Slot{}
	sl.Init(modules, iconId, activate)

	return sl
}

// activate: 键盘按键
func (this *Slot) Init(modules common.Modules, iconId, activate int) common.WidgetSlot {
	eset := modules.Eset()
	render := modules.Render()
	settings := modules.Settings()
	mods := modules.Mods()

	// base
	this.Widget = base.ConstructWidget()

	// self
	this.iconId = iconId
	this.overlayId = slot.NO_CLICK
	this.amount = 1
	this.maxAmount = 1
	this.activateKey = activate
	this.hotkey = -1
	this.enabled = true
	this.visible = true
	this.Widget.SetFocusable(true)
	this.labelAmount = NewLabel(modules)
	this.labelAmount.SetFromLabelInfo(eset.Get("widgets", "quantity_label").(labelinfo.LabelInfo))
	this.labelAmount.SetColor(eset.Get("widgets", "quantity_color").(color.Color))

	this.labelHotkey = NewLabel(modules)
	this.labelHotkey.SetFromLabelInfo(eset.Get("widgets", "hotkey_label").(labelinfo.LabelInfo))
	this.labelHotkey.SetColor(eset.Get("widgets", "hotkey_color").(color.Color))
	this.labelHotkey.SetMaxWidth(eset.Get("resolutions", "icon_size").(int))
	this.SetPosX(eset.Get("resolutions", "icon_size").(int))
	this.SetPosY(eset.Get("resolutions", "icon_size").(int))
	src := rect.Construct(0, 0, eset.Get("resolutions", "icon_size").(int), eset.Get("resolutions", "icon_size").(int))

	selectedFilename := "images/menus/slot_selected.png"
	checkedFilename := "images/menus/slot_checked.png"

	graphics, err := render.LoadImage(settings, mods, selectedFilename)
	if err != nil {
		panic(err)
	}
	defer graphics.UnRef()

	this.slotSelected, err = graphics.CreateSprite()
	if err != nil {
		panic(err)
	}

	this.slotSelected.SetClipFromRect(src)

	graphics1, err := render.LoadImage(settings, mods, checkedFilename)
	if err != nil {
		panic(err)
	}
	defer graphics1.UnRef()

	this.slotChecked, err = graphics1.CreateSprite()
	if err != nil {
		panic(err)
	}

	this.slotChecked.SetClipFromRect(src)

	return this
}

func (this *Slot) Clear() {
	if this.slotSelected != nil {
		this.slotSelected.Close()
		this.slotSelected = nil
	}

	if this.slotChecked != nil {
		this.slotChecked.Close()
		this.slotChecked = nil
	}

	if this.labelAmountBg != nil {
		this.labelAmountBg.Close()
	}

	if this.labelHotkeyBg != nil {
		this.labelHotkeyBg.Close()
	}

	if this.labelAmount != nil {
		this.labelAmount.Close()
		this.labelAmount = nil
	}

	if this.labelHotkey != nil {
		this.labelHotkey.Close()
		this.labelHotkey = nil
	}

}

func (this *Slot) Close() {
	this.Widget.Close(this)
}

func (this *Slot) Activate() {
	this.pressed = true
}

func (this *Slot) Deactivate() {
	this.pressed = false
	this.checked = false
}

func (this *Slot) GetNext(common.Modules) bool {
	this.pressed = false
	this.checked = false
	return false
}

func (this *Slot) GetPrev(common.Modules) bool {
	this.pressed = false
	this.checked = false

	return false
}

func (this *Slot) Defocus() {
	this.Widget.Defocus()
	this.pressed = false
	this.checked = false
}

func (this *Slot) Render(modules common.Modules) error {
	icons := modules.Icons()
	eset := modules.Eset()
	render := modules.Render()
	inpt := modules.Inpt()

	if !this.visible {
		return nil
	}

	if this.iconId != -1 && icons != nil {
		// 绘制代表slot自己的图标
		icons.SetIcon(eset, this.iconId, point.Construct(this.GetPos().X, this.GetPos().Y))
		err := icons.Render(render)
		if err != nil {
			return err
		}

		if this.amount > 1 || this.maxAmount > 1 {
			if this.labelAmountBg != nil {
				err := render.Render(this.labelAmountBg)
				if err != nil {
					return err
				}
			}

			err := this.labelAmount.Render(modules)
			if err != nil {
				return err
			}
		}
	}

	if this.hotkey != -1 {
		if inpt.GetRefreshHotkeys() {
			// 重建
			err := this.SetHotkey(modules, this.hotkey)
			if err != nil {
				return err
			}
		}

		if this.labelHotkeyBg != nil {
			err := render.Render(this.labelHotkeyBg)
			if err != nil {
				return err
			}
		}

		err := this.labelHotkey.Render(modules)
		if err != nil {
			return err
		}
	}

	// 绘制使用或者点击框
	err := this.RenderSelection(modules)
	if err != nil {
		return err
	}

	return nil
}

func (this *Slot) RenderSelection(modules common.Modules) error {
	render := modules.Render()

	if !this.visible {
		return nil
	}

	if this.GetInFocus() {
		if this.slotChecked != nil && this.checked {
			// 有使用 优先级高
			this.slotChecked.SetLocalFrame(this.GetLocalFrame())
			this.slotChecked.SetOffset(this.GetLocalOffset())
			this.slotChecked.SetDestFromRect(this.GetPos())
			err := render.Render(this.slotChecked)
			if err != nil {
				return err
			}
		} else if this.slotSelected != nil {
			// 有点击
			this.slotSelected.SetLocalFrame(this.GetLocalFrame())
			this.slotSelected.SetOffset(this.GetLocalOffset())
			this.slotSelected.SetDestFromRect(this.GetPos())
			err := render.Render(this.slotSelected)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (this *Slot) SetPos1(modules common.Modules, offsetX, offsetY int) error {
	icons := modules.Icons()

	this.Widget.SetPos1(modules, offsetX, offsetY) // 调整整体位置

	// 文字偏移位置
	this.labelAmount.SetPos1(modules, this.GetPos().X+icons.GetTextOffset().X, this.GetPos().Y+icons.GetTextOffset().Y)
	this.labelHotkey.SetPos1(modules, this.GetPos().X+icons.GetTextOffset().X, this.GetPos().Y+icons.GetTextOffset().Y)

	// 背景
	if this.labelAmountBg != nil {
		r := this.labelAmount.GetBounds(modules)
		this.labelAmountBg.SetDest(r.X, r.Y)
	}

	if this.labelHotkeyBg != nil {
		r := this.labelHotkey.GetBounds(modules)
		this.labelHotkeyBg.SetDest(r.X, r.Y)
	}

	return nil
}

func (this *Slot) CheckClick(modules common.Modules) slot.CLICK_TYPE {
	inpt := modules.Inpt()
	mouse := inpt.GetMouse()
	return this.CheckClick1(modules, mouse.X, mouse.Y)

}

func (this *Slot) CheckClick1(modules common.Modules, x, y int) slot.CLICK_TYPE {
	inpt := modules.Inpt()

	if !this.enabled {
		return slot.NO_CLICK
	}

	mouse := point.Construct(x, y)

	// 可以按住持续使用时，右键或激活键未释放，表示还在使用
	if this.continuous && this.pressed && this.checked && (inpt.GetLock(inputstate.MAIN2) || inpt.GetLock(this.activateKey)) {
		return slot.ACTIVATED
	}

	if inpt.GetLock(inputstate.MAIN1) {
		return slot.NO_CLICK
	}

	if inpt.GetLock(inputstate.MAIN2) {
		return slot.NO_CLICK
	}

	if inpt.GetLock(this.activateKey) {
		return slot.NO_CLICK
	}

	// 之前时按了，本次逻辑帧已经释放鼠标
	if this.pressed && !inpt.GetLock(inputstate.MAIN1) && !inpt.GetLock(inputstate.MAIN2) {
		this.pressed = false

		this.checked = !this.checked
		if this.checked {
			return slot.CHECKED
		} else if this.continuous {
			return slot.NO_CLICK
		} else {
			return slot.ACTIVATED
		}
	}

	// 选择
	if inpt.GetPressing(inputstate.MAIN1) {
		if utils.IsWithinRect(this.GetPos(), mouse) {
			inpt.SetLock(inputstate.MAIN1, true)
			this.pressed = true
			this.checked = false
		}
	}

	// 使用
	if inpt.GetPressing(inputstate.MAIN2) {
		if utils.IsWithinRect(this.GetPos(), mouse) {
			inpt.SetLock(inputstate.MAIN2, true)
			this.pressed = true
			this.checked = true
		}
	}

	return slot.NO_CLICK
}

func (this *Slot) GetIcon() int {
	return this.iconId
}

func (this *Slot) SetIcon(iconId int, overlayId slot.CLICK_TYPE) {
	this.iconId = iconId
	this.overlayId = overlayId
}

// 设置slot里物品的数量，内部会重建标题和背景
func (this *Slot) SetAmount(modules common.Modules, amount, maxAmount int) error {
	eset := modules.Eset()
	icons := modules.Icons()
	render := modules.Render()

	this.amount = amount
	this.maxAmount = maxAmount

	amountStr := tools.AbbreviatedKilo(amount)

	if (amount > 1 || maxAmount > 1) && !eset.Get("widgets", "quantity_label").(labelinfo.LabelInfo).Hidden {
		this.labelAmount.SetPos1(modules, this.GetPos().X+icons.GetTextOffset().X, this.GetPos().Y+icons.GetTextOffset().Y)
		this.labelAmount.SetText(amountStr)
		this.labelAmount.SetLocalFrame(this.GetLocalFrame())
		this.labelAmount.SetLocalOffset(this.GetLocalOffset())

		// 更新标题大小
		r := this.labelAmount.GetBounds(modules)
		laGw, err := this.labelAmountBg.GetGraphicsWidth()
		if err != nil {
			return err
		}

		laGh, err := this.labelAmountBg.GetGraphicsHeight()
		if err != nil {
			return err
		}

		// 更新背景大小，必要时重建
		if this.labelAmountBg == nil || laGw != r.W || laGh != r.H {
			if this.labelAmountBg != nil {
				this.labelAmountBg.Close()
				this.labelAmountBg = nil
			}

			bgColor := eset.Get("widgets", "quantity_bg_color").(color.Color)
			if bgColor.A != 0 {

				// 不透明
				temp, err := render.CreateImage(r.W, r.H)
				if err != nil {
					return err
				}
				defer temp.UnRef()

				err = temp.FillWithColor(bgColor)
				if err != nil {
					return err
				}

				this.labelAmountBg, err = temp.CreateSprite()
				if err != nil {
					return err
				}

			}

			if this.labelAmountBg != nil {
				this.labelAmountBg.SetDest(r.X, r.Y)
			}
		}
	}

	return nil
}

// 设置快捷键，内部会重建标题和背景
func (this *Slot) SetHotkey(modules common.Modules, key int) error {
	eset := modules.Eset()
	icons := modules.Icons()
	inpt := modules.Inpt()
	msg := modules.Msg()
	render := modules.Render()

	this.hotkey = key

	if this.hotkey != -1 && !eset.Get("widgets", "hotkey_label").(labelinfo.LabelInfo).Hidden {
		this.labelHotkey.SetPos1(modules, this.GetPos().X+icons.GetTextOffset().X, this.GetPos().Y+icons.GetTextOffset().Y)
		this.labelHotkey.SetText(inpt.GetBindingString(msg, this.hotkey, true))
		this.labelHotkey.SetLocalFrame(this.GetLocalFrame())
		this.labelHotkey.SetLocalOffset(this.GetLocalOffset())

		// 修正标题大小
		r := this.labelHotkey.GetBounds(modules)

		lhGw, err := this.labelHotkeyBg.GetGraphicsWidth()
		if err != nil {
			return err
		}

		lhGh, err := this.labelHotkeyBg.GetGraphicsHeight()
		if err != nil {
			return err
		}

		// 修正背景大小
		if this.labelHotkeyBg == nil || lhGw != r.W || lhGh != r.H {
			if this.labelHotkeyBg != nil {
				this.labelHotkeyBg.Close()
				this.labelHotkeyBg = nil
			}

			bgColor := eset.Get("widgets", "hotkey_bg_color").(color.Color)
			if bgColor.A != 0 {

				// 不透明

				temp, err := render.CreateImage(r.W, r.H)
				if err != nil {
					return err
				}
				defer temp.UnRef()

				err = temp.FillWithColor(bgColor)
				if err != nil {
					return err
				}

				this.labelHotkeyBg, err = temp.CreateSprite()
				if err != nil {
					return err
				}
			}

			if this.labelHotkeyBg != nil {
				this.labelHotkeyBg.SetDest(r.X, r.Y)
			}
		}
	}

	return nil
}

func (this *Slot) SetContinuous(val bool) {
	this.continuous = val
}
