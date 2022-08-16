package sdlinput

import (
	"fmt"
	"math"
	"monster/pkg/common"
	"monster/pkg/common/define/inputstate"
	input "monster/pkg/common/define/inputstate"
	"monster/pkg/common/timer"
	"monster/pkg/subengine/inputstate/base"

	"github.com/veandco/go-sdl2/sdl"
)

type InputState struct {
	base.InputState
	resizeCooldown timer.Timer
	textInput      bool
}

func New(platform common.Platform, settings common.Settings, eset common.EngineSettings, mods common.ModManager, msg common.MessageEngine) *InputState {
	is := &InputState{}

	_ = (common.InputState)(is)

	// base
	is.InputState = base.ConstructInputState()

	// self
	is.resizeCooldown = timer.Construct()
	platform.SetExitEventFilter()
	is.DefaultQwertyKeyBindings(platform) // 默认键盘绑定按键

	// 清空状态
	keyCount := is.GetKeyCount()
	for key := 0; key < keyCount; key++ {
		is.SetPressing(key, false)
		is.SetUnPress(key, false)
		is.SetLock(key, false)
	}

	is.LoadKeyBindings(platform, settings, eset, mods) // 加载文件或写入默认按键绑定
	is.SetKeybindNames(msg)                            // 更换名字
	is.GenCode2Binding()                               // code 和 key 映射
	return is
}

// 清理自己
func (this *InputState) Clear() {
}

func (this *InputState) Close() {
	this.InputState.Close(this)
}

func (this *InputState) LoadKeyBindings(platform common.Platform, settings common.Settings, eset common.EngineSettings, mods common.ModManager) {
	this.InputState.LoadKeyBindings(this, platform, settings, eset, mods)
}

// 默认键盘按键绑定
func (this *InputState) DefaultQwertyKeyBindings(platform common.Platform) {
	if platform.GetIsMobileDevice() {
		panic("is mobile device")
	} else {
		this.SetBinding(input.CANCEL, sdl.K_AC_BACK)
	}

	this.SetBinding(input.ACCEPT, sdl.K_RETURN)
	this.SetBinding(input.UP, sdl.K_w)
	this.SetBinding(input.DOWN, sdl.K_s)
	this.SetBinding(input.LEFT, sdl.K_a)
	this.SetBinding(input.RIGHT, sdl.K_d)

	this.SetBinding(input.BAR_1, sdl.K_q)
	this.SetBinding(input.BAR_2, sdl.K_e)
	this.SetBinding(input.BAR_3, sdl.K_r)
	this.SetBinding(input.BAR_4, sdl.K_f)
	this.SetBinding(input.BAR_5, sdl.K_1)
	this.SetBinding(input.BAR_6, sdl.K_2)
	this.SetBinding(input.BAR_7, sdl.K_3)
	this.SetBinding(input.BAR_8, sdl.K_4)
	this.SetBinding(input.BAR_9, sdl.K_5)
	this.SetBinding(input.BAR_0, sdl.K_6)

	this.SetBinding(input.CHARACTER, sdl.K_c)
	this.SetBinding(input.INVENTORY, sdl.K_i)
	this.SetBinding(input.POWERS, sdl.K_p)
	this.SetBinding(input.LOG, sdl.K_l)

	this.SetBinding(input.MAIN1, (-1)*(sdl.BUTTON_LEFT+inputstate.MOUSE_BIND_OFFSET))
	this.SetBinding(input.MAIN2, (-1)*(sdl.BUTTON_RIGHT+inputstate.MOUSE_BIND_OFFSET))

	this.SetBinding(input.CTRL, sdl.K_LCTRL)
	this.SetBinding(input.SHIFT, sdl.K_LSHIFT)
	this.SetBinding(input.DEL, sdl.K_DELETE)
	this.SetBinding(input.ALT, sdl.K_LALT)

	this.SetBinding(input.ACTIONBAR, sdl.K_b)
	this.SetBinding(input.ACTIONBAR_BACK, sdl.K_z)
	this.SetBinding(input.ACTIONBAR_FORWARD, sdl.K_x)
	this.SetBinding(input.ACTIONBAR_USE, sdl.K_n)

	this.SetBinding(input.DEVELOPER_MENU, sdl.K_F5)

	// 转化 SDL_Keycode 到 SDL_Scancode
	keyCount := this.GetKeyCount()
	for key := 0; key < keyCount; key++ {
		scanCode := sdl.GetScancodeFromKey((sdl.Keycode)(this.GetBinding(key)))
		if scanCode > 0 {
			this.SetBinding(key, (int)(scanCode))
		}
	}
}

func (this *InputState) ValidateFixedKeyBinding(action, key int) {
	scanKey := 0

	if key < 0 {
		scanKey = key // 处理鼠标
	} else {
		scanKey = (int)(sdl.GetScancodeFromKey((sdl.Keycode)(key))) // 转成 scancode
	}

	keyCount := this.GetKeyCount()
	for i := 0; i < keyCount; i++ {
		// 设置该key
		if i == action {
			this.SetBinding(action, scanKey)
			continue
		}

		// 清理其他重复的
		if this.GetBinding(i) == scanKey {
			this.SetBinding(i, -1)
		}
	}
}

// 固定绑定几个键盘按钮
func (this *InputState) SetFixedKeyBinding() {
	this.ValidateFixedKeyBinding(input.MAIN1, (-1)*(sdl.BUTTON_LEFT+input.MOUSE_BIND_OFFSET))
	this.ValidateFixedKeyBinding(input.CTRL, sdl.K_LCTRL)
	this.ValidateFixedKeyBinding(input.SHIFT, sdl.K_LSHIFT)
	this.ValidateFixedKeyBinding(input.DEL, sdl.K_DELETE)
	this.ValidateFixedKeyBinding(input.ALT, sdl.K_LALT)
}

func (this *InputState) Handle(modules common.Modules) error {
	settings := modules.Settings()
	eset := modules.Eset()
	device := modules.Render()
	curs := device.Curs()

	this.InputState.Handle() // 清除
	bindButton := 0

	for {
		rawEvent := sdl.PollEvent()
		if rawEvent == nil {
			break
		}

		if this.DumpEvent() {
			fmt.Println(rawEvent)
		}

		switch rawEvent.GetType() {
		case sdl.TEXTINPUT:
			event := rawEvent.(*sdl.TextInputEvent)
			this.SetInKeys(this.GetInKeys() + event.GetText())
		case sdl.MOUSEMOTION:
			event := rawEvent.(*sdl.MouseMotionEvent)
			this.SetMouse(this.ScaleMouse(settings, (uint)(event.X), (uint)(event.Y)))
			curs.SetShowCursor(true) // 显示
		case sdl.MOUSEWHEEL:
			event := rawEvent.(*sdl.MouseWheelEvent)
			if event.Y > 0 {
				this.SetScrollUp(true)
			} else if event.Y < 0 {
				this.SetScrollDown(true)
			}
		case sdl.MOUSEBUTTONDOWN:

			event := rawEvent.(*sdl.MouseButtonEvent)
			this.SetMouse(this.ScaleMouse(settings, (uint)(event.X), (uint)(event.Y)))
			bindButton = (-1) * ((int)(event.Button) + inputstate.MOUSE_BIND_OFFSET)
			if key, ok := this.GetCode2Binding(bindButton); ok {
				this.SetPressing(key, true)
				this.SetUnPress(key, false)
			}
		case sdl.MOUSEBUTTONUP:

			event := rawEvent.(*sdl.MouseButtonEvent)
			this.SetMouse(this.ScaleMouse(settings, (uint)(event.X), (uint)(event.Y)))
			bindButton = (-1) * ((int)(event.Button) + inputstate.MOUSE_BIND_OFFSET)
			if key, ok := this.GetCode2Binding(bindButton); ok {
				this.SetUnPress(key, true)
			}
			this.SetLastButton(bindButton)

		case sdl.WINDOWEVENT:
			event := rawEvent.(*sdl.WindowEvent)
			if event.Event == sdl.WINDOWEVENT_SIZE_CHANGED {
				fmt.Println("size change")
				this.resizeCooldown.SetDuration((uint)(settings.Get("max_fps").(int)) / 4)
			} else if event.Event == sdl.WINDOWEVENT_MINIMIZED {
				fmt.Println("mini")
				this.SetWindowMinimized(true)

				//TODO
				// sound
				// menu

			} else if event.Event == sdl.WINDOWEVENT_FOCUS_GAINED {
				// bugfix
				fmt.Println("focus")
				this.SetWindowMinimized(false)
			} else if event.Event == sdl.WINDOWEVENT_RESTORED {
				fmt.Println("restore")
				this.SetWindowRestored(true)
				//TODO
				// sound
			}

		case sdl.KEYDOWN:
		case sdl.KEYUP:
		case sdl.QUIT:
			this.SetDone(true)
			keyCount := this.GetKeyCount()
			for key := 0; key < keyCount; key++ {
				this.SetPressing(key, false)
				this.SetUnPress(key, false)
				this.SetLock(key, false)
			}
		default:
			break
		}
	}

	if this.resizeCooldown.GetDuration() > 0 {
		this.resizeCooldown.Tick()

		if this.resizeCooldown.IsEnd() {
			this.resizeCooldown.SetDuration(0)
			this.SetWindowResized(true)
			if err := device.WindowResize(settings, eset); err != nil {
				return err
			}
		}
	}

	keyCount := this.GetKeyCount()
	for i := 0; i < keyCount; i++ {
		if this.GetSlowRepeat(i) {
			if !this.GetPressing(i) {
				// 未按下，则重置延长
				this.GetRepeatCooldown(i).SetDuration((uint)(settings.Get("max_fps").(int)))
			} else if this.GetPressing(i) && !this.GetLock(i) {
				// 按下了，则锁住并设置延长
				this.SetLock(i, true)
				prevDuration := this.GetRepeatCooldown(i).GetDuration()

				this.GetRepeatCooldown(i).SetDuration((uint)(math.Max((float64)(settings.Get("max_fps").(int)/10), (float64)((int)(prevDuration)-settings.Get("max_fps").(int)/2))))

			} else if this.GetPressing(i) && this.GetLock(i) {
				this.GetRepeatCooldown(i).Tick()
				if this.GetRepeatCooldown(i).IsEnd() {
					this.SetLock(i, false)
				}
			}
		}
	}

	return nil
}

func (this *InputState) HideCursor() error {
	_, err := sdl.ShowCursor(sdl.DISABLE)
	if err != nil {
		return err
	}
	return nil
}

func (this *InputState) ShowCursor() error {
	_, err := sdl.ShowCursor(sdl.ENABLE)
	if err != nil {
		return err
	}

	return nil
}

func (this *InputState) GetKeyName(msg common.MessageEngine, key int, getShortString bool) string {
	key1 := sdl.GetKeyFromScancode((sdl.Scancode)(key))
	if getShortString {
		switch key1 {
		case sdl.K_BACKSPACE:
			return msg.Get("BkSp")
		case sdl.K_CAPSLOCK:
			return msg.Get("Caps")
		case sdl.K_DELETE:
			return msg.Get("Del")
		case sdl.K_DOWN:
			return msg.Get("Down")
		case sdl.K_END:
			return msg.Get("End")
		case sdl.K_ESCAPE:
			return msg.Get("Esc")
		case sdl.K_HOME:
			return msg.Get("Home")
		case sdl.K_INSERT:
			return msg.Get("Ins")
		case sdl.K_LALT:
			return msg.Get("LAlt")
		case sdl.K_LCTRL:
			return msg.Get("LCtrl")
		case sdl.K_LEFT:
			return msg.Get("Left")
		case sdl.K_LSHIFT:
			return msg.Get("LShft")
		case sdl.K_NUMLOCKCLEAR:
			return msg.Get("Num")
		case sdl.K_PAGEDOWN:
			return msg.Get("PgDn")
		case sdl.K_PAGEUP:
			return msg.Get("PgUp")
		case sdl.K_PAUSE:
			return msg.Get("Pause")
		case sdl.K_PRINTSCREEN:
			return msg.Get("Print")
		case sdl.K_RALT:
			return msg.Get("RAlt")
		case sdl.K_RCTRL:
			return msg.Get("RCtrl")
		case sdl.K_RETURN:
			return msg.Get("Ret")
		case sdl.K_RIGHT:
			return msg.Get("Right")
		case sdl.K_RSHIFT:
			return msg.Get("RShft")
		case sdl.K_SCROLLLOCK:
			return msg.Get("SLock")
		case sdl.K_SPACE:
			return msg.Get("Spc")
		case sdl.K_TAB:
			return msg.Get("Tab")
		case sdl.K_UP:
			return msg.Get("Up")
		}
	} else {
		switch key1 {
		case sdl.K_BACKSPACE:
			return msg.Get("Backspace")
		case sdl.K_CAPSLOCK:
			return msg.Get("CapsLock")
		case sdl.K_DELETE:
			return msg.Get("Delete")
		case sdl.K_DOWN:
			return msg.Get("Down")
		case sdl.K_END:
			return msg.Get("End")
		case sdl.K_ESCAPE:
			return msg.Get("Escape")
		case sdl.K_HOME:
			return msg.Get("Home")
		case sdl.K_INSERT:
			return msg.Get("Insert")
		case sdl.K_LALT:
			return msg.Get("Left Alt")
		case sdl.K_LCTRL:
			return msg.Get("Left Ctrl")
		case sdl.K_LEFT:
			return msg.Get("Left")
		case sdl.K_LSHIFT:
			return msg.Get("Left Shift")
		case sdl.K_NUMLOCKCLEAR:
			return msg.Get("NumLock")
		case sdl.K_PAGEDOWN:
			return msg.Get("PageDown")
		case sdl.K_PAGEUP:
			return msg.Get("PageUp")
		case sdl.K_PAUSE:
			return msg.Get("Pause")
		case sdl.K_PRINTSCREEN:
			return msg.Get("PrintScreen")
		case sdl.K_RALT:
			return msg.Get("Right Alt")
		case sdl.K_RCTRL:
			return msg.Get("Right Ctrl")
		case sdl.K_RETURN:
			return msg.Get("Return")
		case sdl.K_RIGHT:
			return msg.Get("Right")
		case sdl.K_RSHIFT:
			return msg.Get("Right Shift")
		case sdl.K_SCROLLLOCK:
			return msg.Get("ScrollLock")
		case sdl.K_SPACE:
			return msg.Get("Space")
		case sdl.K_TAB:
			return msg.Get("Tab")
		case sdl.K_UP:
			return msg.Get("Up")
		}
	}

	return sdl.GetKeyName(key1)
}

func (this *InputState) GetMouseButtonName(msg common.MessageEngine, button int, getShortString bool) string {
	realButton := (button + inputstate.MOUSE_BIND_OFFSET) * (-1)

	if getShortString {
		return msg.Get(fmt.Sprintf("M%d", realButton))
	}

	if realButton > 0 && realButton <= this.GetMouseButtonNameCount() {
		return this.GetMouseButton(realButton - 1)
	} else {
		return msg.Get(fmt.Sprintf("Mouse %d", realButton))
	}
}

func (this *InputState) GetBindingString(msg common.MessageEngine, key int, getShortString bool) string {
	none := ""
	if !getShortString {
		none = msg.Get("(none)")
	}

	if this.GetBinding(key) == 0 || this.GetBinding(key) == -1 {
		return none
	} else if this.GetBinding(key) < -1 {
		return this.GetMouseButtonName(msg, this.GetBinding(key), getShortString)
	} else {
		return this.GetKeyName(msg, this.GetBinding(key), getShortString)
	}

	return none
}

func (this *InputState) GetMovementString(settings common.Settings, msg common.MessageEngine) string {
	output := "["

	if settings.Get("mouse_move").(bool) {
		if settings.Get("mouse_move_swap").(bool) {
			output += this.GetBindingString(msg, inputstate.MAIN2, false)
		} else {
			output += this.GetBindingString(msg, inputstate.MAIN1, false)
		}
	} else {
		output += this.GetBindingString(msg, inputstate.LEFT, false)
		output += this.GetBindingString(msg, inputstate.RIGHT, false)
		output += this.GetBindingString(msg, inputstate.UP, false)
		output += this.GetBindingString(msg, inputstate.DOWN, false)
	}

	output += "]"
	return output
}

func (this *InputState) GetAttackString(msg common.MessageEngine) string {
	output := "["
	output += this.GetBindingString(msg, inputstate.MAIN1, false)
	output += "]"

	return output
}

func (this *InputState) GetContinueString(msg common.MessageEngine) string {
	output := "["
	output += this.GetBindingString(msg, inputstate.ACCEPT, false)
	output += "]"

	return output
}

func (this *InputState) UsingMouse(settings common.Settings) bool {
	return !settings.Get("no_mouse").(bool)
}

func (this *InputState) StartTextInput() {
	if this.textInput {
		sdl.StartTextInput()
		this.textInput = true
	}
}

func (this *InputState) StopTextInput() {
	if this.textInput {
		sdl.StopTextInput()
		this.textInput = false
	}
}

func (this *InputState) SetKeybind(msg common.MessageEngine, key, bindingButton int) string {
	keybindMsg := ""
	if key != -1 {
		// prevent unmapping "fixed" keybinds
		if (key == (sdl.BUTTON_LEFT+input.MOUSE_BIND_OFFSET)*(-1) && bindingButton != inputstate.MAIN1) ||
			key == sdl.SCANCODE_LCTRL ||
			key == sdl.SCANCODE_RCTRL ||
			key == sdl.SCANCODE_LSHIFT ||
			key == sdl.SCANCODE_RSHIFT ||
			key == sdl.SCANCODE_LALT ||
			key == sdl.SCANCODE_RALT ||
			key == sdl.SCANCODE_DELETE ||
			key == sdl.SCANCODE_BACKSPACE {
			if key < -1 {
				keybindMsg = msg.Get(fmt.Sprintf("Can not bind: %s", this.GetMouseButtonName(msg, key, false)))
			} else {
				keybindMsg = msg.Get(fmt.Sprintf("Can not bind: %s", this.GetKeyName(msg, key, false)))
			}

			return keybindMsg
		}

		// 清理之前绑定的
		keyCount := this.GetKeyCount()
		for i := 0; i < keyCount; i++ {

			if i == bindingButton {
				continue
			}

			if this.GetBinding(i) == key && i != bindingButton {
				keybindMsg = msg.Get(fmt.Sprintf("'%s' is no longer bound to:", this.GetBindingString(msg, i, false))) + " '" + this.GetBindingName(i) + "'"
				this.SetBinding(i, -1)
			}
		}
	}

	this.SetBinding(bindingButton, key)
	return keybindMsg
}

func (this *InputState) GetKeyFromName(keyName string) int {
	return (int)(sdl.GetScancodeFromName(keyName))
}
