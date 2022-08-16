package cursormanager

import (
	"monster/pkg/common"
	"monster/pkg/common/define/cursormanager"
	"monster/pkg/common/point"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/filesystem/logfile"
	"monster/pkg/utils"
	"monster/pkg/utils/parsing"
)

type CursorManager struct {
	showCursor        bool
	cursorNormal      common.Sprite
	cursorInteract    common.Sprite
	cursorTalk        common.Sprite
	cursorAttack      common.Sprite
	cursorLhpNormal   common.Sprite
	cursorLhpInteract common.Sprite
	cursorLhpTalk     common.Sprite
	cursorLhpAttack   common.Sprite
	offsetNormal      point.Point
	offsetInteract    point.Point
	offsetTalk        point.Point
	offsetAttack      point.Point
	offsetLhpNormal   point.Point
	offsetLhpInteract point.Point
	offsetLhpTalk     point.Point
	offsetLhpAttack   point.Point

	cursorCurrent common.Sprite
	offsetCurrent *point.Point
	lowHp         bool
}

func New(settings common.Settings, mods common.ModManager, device common.RenderDevice) *CursorManager {
	cm := &CursorManager{
		showCursor: true,
	}

	infile := fileparser.New()

	err := infile.Open("engine/mouse_cursor.txt", true, mods)
	if err != nil && utils.IsNotExist(err) {
		return cm
	} else if err != nil {
		panic(err)
	}
	defer infile.Close()

	for infile.Next(mods) {
		var offset point.Point
		var sprite common.Sprite
		strVal := ""
		first := ""

		switch infile.Key() {
		case "normal":
			fallthrough
		case "interact":
			fallthrough
		case "talk":
			fallthrough
		case "attack":
			fallthrough
		case "lowhp_normal":
			fallthrough
		case "lowhp_interact":
			fallthrough
		case "lowhp_talk":
			fallthrough
		case "lowhp_attack":
			strVal = infile.Val()
			first, strVal = parsing.PopFirstString(strVal, "")
			graphics, err := device.LoadImage(settings, mods, first)
			if err != nil {
				panic(err)
			}

			defer graphics.UnRef()
			sprite, err = graphics.CreateSprite()
			if err != nil {
				panic(err)
			}

			//graphics.UnRef()
			offset = parsing.ToPoint(strVal)
		default:
			logfile.LogError("CursorManager: '%s' is not a valid key.", infile.Key())
		}

		switch infile.Key() {
		case "normal":
			cm.cursorNormal = sprite
			cm.offsetNormal = offset
		case "interact":
			cm.cursorInteract = sprite
			cm.offsetInteract = offset
		case "talk":
			cm.cursorTalk = sprite
			cm.offsetTalk = offset
		case "attack":
			cm.cursorAttack = sprite
			cm.offsetAttack = offset
		case "lowhp_normal":
			cm.cursorLhpNormal = sprite
			cm.offsetLhpNormal = offset
		case "lowhp_interact":
			cm.cursorLhpInteract = sprite
			cm.offsetLhpInteract = offset
		case "lowhp_talk":
			cm.cursorLhpTalk = sprite
			cm.offsetLhpTalk = offset
		case "lowhp_attack":
			cm.cursorLhpAttack = sprite
			cm.offsetLhpAttack = offset
		}
	}

	return cm
}

func (this *CursorManager) Close() {
	if this.cursorNormal != nil {
		this.cursorNormal.Close()
	}
	if this.cursorInteract != nil {
		this.cursorInteract.Close()
	}
	if this.cursorTalk != nil {
		this.cursorTalk.Close()
	}
	if this.cursorAttack != nil {
		this.cursorAttack.Close()
	}
	if this.cursorLhpNormal != nil {
		this.cursorLhpNormal.Close()
	}
	if this.cursorLhpInteract != nil {
		this.cursorLhpInteract.Close()
	}
	if this.cursorLhpTalk != nil {
		this.cursorLhpTalk.Close()
	}
	if this.cursorLhpAttack != nil {
		this.cursorLhpAttack.Close()
	}
}

func (this *CursorManager) Logic(settings common.Settings, inpt common.InputState) error {
	if !this.showCursor {
		err := inpt.HideCursor()
		if err != nil {
			return err
		}
		return nil
	}

	if settings.Get("hardware_cursor").(bool) {
		// 使用操作系统鼠标
		err := inpt.ShowCursor()
		if err != nil {
			return err
		}
		return nil
	}

	this.cursorCurrent = nil
	this.offsetCurrent = nil
	return nil
}

func (this *CursorManager) Render(settings common.Settings, device common.RenderDevice, inpt common.InputState) error {
	// 使用系统的鼠标或不启用光标
	if settings.Get("hardware_cursor").(bool) || !this.showCursor {
		return nil
	}

	if this.cursorCurrent != nil {
		mouse := inpt.GetMouse()
		if this.offsetCurrent != nil {
			this.cursorCurrent.SetDest(mouse.X+this.offsetCurrent.X, mouse.Y+this.offsetCurrent.Y)
		} else {
			this.cursorCurrent.SetDest(mouse.X, mouse.Y)
		}

		err := device.Render(this.cursorCurrent)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *CursorManager) SetLowHP(val bool) {
	this.lowHp = val
}

func (this *CursorManager) SetShowCursor(val bool) {
	this.showCursor = val
}

func (this *CursorManager) SetCursor(settings common.Settings, inpt common.InputState, type1 int) error {
	// 使用系统的鼠标
	if settings.Get("hardware_cursor").(bool) {
		return nil
	}

	hide := false
	if type1 == cursormanager.CURSOR_INTERACT {
		if this.lowHp && this.cursorLhpInteract != nil {
			hide = true
			this.cursorCurrent = this.cursorLhpInteract
			this.offsetCurrent = &this.offsetLhpInteract
		} else if this.cursorInteract != nil {
			hide = true
			this.cursorCurrent = this.cursorInteract
			this.offsetCurrent = &this.offsetInteract
		}

		if hide {
			if err := inpt.HideCursor(); err != nil {
				return err
			}
		}
	} else if type1 == cursormanager.CURSOR_TALK {
		if this.lowHp && this.cursorLhpTalk != nil {
			hide = true
			this.cursorCurrent = this.cursorLhpTalk
			this.offsetCurrent = &this.offsetLhpTalk
		} else if this.cursorTalk != nil {
			hide = true
			this.cursorCurrent = this.cursorTalk
			this.offsetCurrent = &this.offsetTalk
		}

		if hide {
			if err := inpt.HideCursor(); err != nil {
				return err
			}
		}
	} else if type1 == cursormanager.CURSOR_ATTACK {
		if this.lowHp && this.cursorLhpAttack != nil {
			hide = true
			this.cursorCurrent = this.cursorLhpAttack
			this.offsetCurrent = &this.offsetLhpAttack
		} else if this.cursorAttack != nil {
			hide = true
			this.cursorCurrent = this.cursorAttack
			this.offsetCurrent = &this.offsetAttack
		}

		if hide {
			if err := inpt.HideCursor(); err != nil {
				return err
			}
		}
	} else if this.cursorNormal != nil || (this.cursorLhpNormal != nil && this.lowHp) {
		if this.lowHp && this.cursorLhpNormal != nil {
			hide = true
			this.cursorCurrent = this.cursorLhpNormal
			this.offsetCurrent = &this.offsetLhpNormal
		} else if this.cursorAttack != nil {
			hide = true
			this.cursorCurrent = this.cursorNormal
			this.offsetCurrent = &this.offsetNormal
		}

		if hide {
			if err := inpt.HideCursor(); err != nil {
				return err
			}
		}
	} else {
		this.cursorCurrent = nil
		if err := inpt.ShowCursor(); err != nil {
			return err
		}
	}

	return nil
}
