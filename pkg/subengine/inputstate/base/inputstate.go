package base

import (
	"monster/pkg/common"
	"monster/pkg/common/define/inputstate"
	"monster/pkg/common/define/platform"
	"monster/pkg/common/point"
	"monster/pkg/common/timer"
	"monster/pkg/config/version"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/filesystem/logfile"
	"monster/pkg/utils"
	"monster/pkg/utils/parsing"
	"os"
	"strconv"
)

type InputState struct {
	binding         map[int]int
	code2binding    map[int]int
	bindingName     map[int]string
	mouseButton     map[int]string
	pressing        map[int]bool
	lock            map[int]bool
	slowRepeat      map[int]bool
	repeatCooldown  map[int]*timer.Timer
	done            bool
	mouse           point.Point
	inKeys          string
	lastKey         int
	lastButton      int
	scrollUp        bool
	scrollDown      bool
	lockScroll      bool
	unPress         map[int]bool
	dumpEvent       bool
	fileVersion     version.Version
	fileVersionMin  version.Version
	refreshHotkeys  bool
	lockAll         bool
	windowMinimized bool
	windowRestored  bool
	windowResized   bool
}

func ConstructInputState() InputState {
	inpt := InputState{
		binding:        map[int]int{},
		code2binding:   map[int]int{},
		bindingName:    map[int]string{},
		mouseButton:    map[int]string{},
		pressing:       map[int]bool{},
		lock:           map[int]bool{},
		slowRepeat:     map[int]bool{},
		repeatCooldown: map[int]*timer.Timer{},
		unPress:        map[int]bool{},
		fileVersion:    version.Construct(),
		fileVersionMin: version.Construct(1, 9, 20),
	}

	for i := 0; i < 31; i++ {
		inpt.binding[i] = 0
		inpt.pressing[i] = false
		inpt.unPress[i] = false
		inpt.lock[i] = false
		inpt.slowRepeat[i] = false
		inpt.repeatCooldown[i] = timer.New()

		if i < 7 {
			inpt.mouseButton[i] = ""
		}
	}

	return inpt
}

// 清理自己
func (this *InputState) Clear() {
	// pass
}

func (this *InputState) Close(impl common.InputState) {
	// 子类清理
	impl.Clear()

	// 自己清理
	this.Clear()
}

// 加载按键绑定，若mod的按键捆绑文件存在则在用户配置文件目录加载用户的改动，如不存在用户的改动则加载mod的，如果前面都不存在则加载全局的配置
func (this *InputState) LoadKeyBindings(impl common.InputState, plat common.Platform, settings common.Settings, eset common.EngineSettings, mods common.ModManager) error {
	infile := fileparser.Construct()
	openedFile := false

	// 定位mod里的文件
	if _, err := mods.Locate(settings, "engine/default_keybindings.txt"); err == nil {
		// 先mod
		// 先用户或mod默认
		if err := infile.Open(settings.GetPathUser()+"saves/"+eset.Get("misc", "save_prefix").(string), false, mods); err == nil {
			openedFile = true
		} else if err != nil && utils.IsNotExist(err) {
			if err := infile.Open("engine/default_keybindings.txt", true, mods); err == nil {
				openedFile = true
			} else if err != nil && !utils.IsNotExist(err) {
				return err
			}
		} else {
			return err
		}

	} else if utils.IsNotExist(err) {
		// 后全局
		if err := infile.Open(settings.GetPathConf()+"keybindings.txt", false, mods); err == nil {
			openedFile = true

			if ok, err := utils.FileExists(settings.GetPathUser() + "saves/" + eset.Get("misc", "save_prefix").(string)); err == nil {
				if ok {
					logfile.LogInfo("InputState: Found unexpected save prefix keybinding file. Removing it now.")
					utils.RemoveFile(settings.GetPathUser() + "saves/" + eset.Get("misc", "save_prefix").(string))
				}

			} else if err != nil {
				return err
			}

		} else if err != nil && utils.IsNotExist(err) {
			openedFile = false
		} else {
			return err
		}
	} else {
		return err
	}

	// 上述文件都不存在则创建
	if !openedFile {
		err := this.SaveKeyBindings(settings, eset, mods)
		if err != nil {
			return err
		}
		return nil
	}

	defer infile.Close()

	// 解析按键捆绑文件
	for infile.Next(mods) {
		// 基础信息检查
		if infile.GetSection() == "" {
			if infile.Key() == "file_version" {
				this.fileVersion.SetFromString(infile.Val())
			}

			// 版本太小
			if version.Compare(this.fileVersion, this.fileVersionMin) < 0 {
				logfile.LogError("InputState: Keybindings configuration file is out of date (%s < %s). Resetting to engine defaults.", this.fileVersion.GetString(), this.fileVersionMin.GetString())

				if plat.GetConfigMenuType() != platform.CONFIG_MENU_TYPE_BASE {
					logfile.LogErrorDialog("InputState: Keybindings configuration file is out of date. Resetting to engine defaults.", this.fileVersion.GetString(), this.fileVersionMin.GetString())
				}

				err := this.SaveKeyBindings(settings, eset, mods)
				if err != nil {
					return err
				}
				break
			}

			continue
		}

		strVal := infile.Val()
		str1 := ""
		key1 := -1

		// 开始解析
		switch infile.GetSection() {
		case "user":
			// 引擎提供
			str1, strVal = parsing.PopFirstString(strVal, "")
			key1 = parsing.ToInt(str1, -1)
		case "default":
			// mod提供的
			str1, strVal = parsing.PopFirstString(strVal, "")
			if len(str1) > 0 && str1[0:6] == "mouse_" {
				// mouse_
				key1 = -1 * (parsing.ToInt(str1[6:], 0) + 1 + inputstate.MOUSE_BIND_OFFSET)
			} else if str1 != "-1" {
				key1 = impl.GetKeyFromName(str1)
			}

		default:
			continue
		}

		// 加载到内存
		cursor := -1
		switch infile.Key() {
		case "cancel":
			cursor = inputstate.CANCEL
		case "accept":
			cursor = inputstate.ACCEPT
		case "up":
			cursor = inputstate.UP
		case "down":
			cursor = inputstate.DOWN
		case "right":
			cursor = inputstate.RIGHT
		case "bar1":
			cursor = inputstate.BAR_1
		case "bar2":
			cursor = inputstate.BAR_2
		case "bar3":
			cursor = inputstate.BAR_3
		case "bar4":
			cursor = inputstate.BAR_4
		case "bar5":
			cursor = inputstate.BAR_5
		case "bar6":
			cursor = inputstate.BAR_6
		case "bar7":
			cursor = inputstate.BAR_7
		case "bar8":
			cursor = inputstate.BAR_8
		case "bar9":
			cursor = inputstate.BAR_9
		case "bar0":
			cursor = inputstate.BAR_0
		case "main1":
			cursor = inputstate.MAIN1
		case "main2":
			cursor = inputstate.MAIN2
		case "character":
			cursor = inputstate.CHARACTER
		case "inventory":
			cursor = inputstate.INVENTORY
		case "powers":
			cursor = inputstate.POWERS
		case "log":
			cursor = inputstate.LOG
		case "ctrl":
			cursor = inputstate.CTRL
		case "shift":
			cursor = inputstate.SHIFT
		case "alt":
			cursor = inputstate.ALT
		case "delete":
			cursor = inputstate.DEL
		case "actionbar":
			cursor = inputstate.ACTIONBAR
		case "actionbar_back":
			cursor = inputstate.ACTIONBAR_BACK
		case "actionbar_forward":
			cursor = inputstate.ACTIONBAR_FORWARD
		case "actionbar_use":
			cursor = inputstate.ACTIONBAR_USE
		case "developer_menu":
			cursor = inputstate.DEVELOPER_MENU
		}

		if cursor != -1 {
			this.binding[cursor] = key1

		}
	}

	// 恢复几个固定的按键，不让其被覆盖
	impl.SetFixedKeyBinding()

	return nil
}

// 清理之前的状态
func (this *InputState) Handle() {
	this.refreshHotkeys = false

	if this.lockAll {
		return
	}

	if len(this.binding) != 31 {
		panic("binding is not 31")
	}

	this.inKeys = ""

	for key := 0; key < len(this.binding); key++ {
		if this.unPress[key] == true {
			this.pressing[key] = false
			this.unPress[key] = false
			this.lock[key] = false
		}
	}
}

// 保存按键绑定，若mod的按键捆绑文件存在则写用户配置文件目录，否则写到全局配置
func (this *InputState) SaveKeyBindings(settings common.Settings, eset common.EngineSettings, mods common.ModManager) error {
	outPath := ""

	if _, err := mods.Locate(settings, "engine/default_keybindings.txt"); err == nil {
		err = utils.CreateDir(settings.GetPathUser() + "saves/" + eset.Get("misc", "save_prefix").(string) + "/keybindings.txt")
		if err != nil {
			return err
		}

		outPath = settings.GetPathUser() + "saves/" + eset.Get("misc", "save_prefix").(string) + "/keybindings.txt"
	} else if err != nil && utils.IsNotExist(err) {
		outPath = settings.GetPathConf() + "keybindings.txt"
	} else {
		return err
	}

	f, err := os.OpenFile(outPath, os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString("# Keybindings\n")
	f.WriteString("# FORMAT: {ACTION}={BIND},{BIND_ALT},{BIND_JOY}\n")
	f.WriteString("# A bind value of -1 means unbound\n")
	f.WriteString("# For BIND and BIND_ALT, a value of 0 is also unbound\n")
	f.WriteString("# For BIND and BIND_ALT, any value less than -1 is a mouse button\n")
	f.WriteString("# As an example, mouse button 1 would be -3 here. Button 2 would be -4, etc.\n\n")

	// 设置最小的版本
	this.fileVersion = this.fileVersionMin

	f.WriteString("file_version=" + this.fileVersion.GetString() + "\n\n")
	f.WriteString("[user]\n")
	f.WriteString("cancel=" + strconv.Itoa(this.binding[inputstate.CANCEL]) + "\n")
	f.WriteString("accept=" + strconv.Itoa(this.binding[inputstate.ACCEPT]) + "\n")
	f.WriteString("up=" + strconv.Itoa(this.binding[inputstate.UP]) + "\n")
	f.WriteString("down=" + strconv.Itoa(this.binding[inputstate.DOWN]) + "\n")
	f.WriteString("left=" + strconv.Itoa(this.binding[inputstate.LEFT]) + "\n")
	f.WriteString("right=" + strconv.Itoa(this.binding[inputstate.RIGHT]) + "\n")
	f.WriteString("bar1=" + strconv.Itoa(this.binding[inputstate.BAR_1]) + "\n")
	f.WriteString("bar2=" + strconv.Itoa(this.binding[inputstate.BAR_2]) + "\n")
	f.WriteString("bar3=" + strconv.Itoa(this.binding[inputstate.BAR_3]) + "\n")
	f.WriteString("bar4=" + strconv.Itoa(this.binding[inputstate.BAR_4]) + "\n")
	f.WriteString("bar5=" + strconv.Itoa(this.binding[inputstate.BAR_5]) + "\n")
	f.WriteString("bar6=" + strconv.Itoa(this.binding[inputstate.BAR_6]) + "\n")
	f.WriteString("bar7=" + strconv.Itoa(this.binding[inputstate.BAR_7]) + "\n")
	f.WriteString("bar8=" + strconv.Itoa(this.binding[inputstate.BAR_8]) + "\n")
	f.WriteString("bar9=" + strconv.Itoa(this.binding[inputstate.BAR_9]) + "\n")
	f.WriteString("bar0=" + strconv.Itoa(this.binding[inputstate.BAR_0]) + "\n")
	f.WriteString("main1=" + strconv.Itoa(this.binding[inputstate.MAIN1]) + "\n")
	f.WriteString("main2=" + strconv.Itoa(this.binding[inputstate.MAIN2]) + "\n")
	f.WriteString("character=" + strconv.Itoa(this.binding[inputstate.CHARACTER]) + "\n")
	f.WriteString("inventory=" + strconv.Itoa(this.binding[inputstate.INVENTORY]) + "\n")
	f.WriteString("powers=" + strconv.Itoa(this.binding[inputstate.POWERS]) + "\n")
	f.WriteString("log=" + strconv.Itoa(this.binding[inputstate.LOG]) + "\n")
	f.WriteString("ctrl=" + strconv.Itoa(this.binding[inputstate.CTRL]) + "\n")
	f.WriteString("shift=" + strconv.Itoa(this.binding[inputstate.SHIFT]) + "\n")
	f.WriteString("alt=" + strconv.Itoa(this.binding[inputstate.ALT]) + "\n")
	f.WriteString("delete=" + strconv.Itoa(this.binding[inputstate.DEL]) + "\n")
	f.WriteString("actionbar=" + strconv.Itoa(this.binding[inputstate.ACTIONBAR]) + "\n")
	f.WriteString("actionbar_back=" + strconv.Itoa(this.binding[inputstate.ACTIONBAR_BACK]) + "\n")
	f.WriteString("actionbar_forward=" + strconv.Itoa(this.binding[inputstate.ACTIONBAR_FORWARD]) + "\n")
	f.WriteString("actionbar_use=" + strconv.Itoa(this.binding[inputstate.ACTIONBAR_USE]) + "\n")
	f.WriteString("developer_menu=" + strconv.Itoa(this.binding[inputstate.DEVELOPER_MENU]) + "\n")

	return nil
}

func (this *InputState) ResetScroll() {
	this.scrollUp = false
	this.scrollDown = false
}

func (this *InputState) LockActionBar() {
	this.pressing[inputstate.BAR_1] = false
	this.pressing[inputstate.BAR_2] = false
	this.pressing[inputstate.BAR_3] = false
	this.pressing[inputstate.BAR_4] = false
	this.pressing[inputstate.BAR_5] = false
	this.pressing[inputstate.BAR_6] = false
	this.pressing[inputstate.BAR_7] = false
	this.pressing[inputstate.BAR_8] = false
	this.pressing[inputstate.BAR_9] = false
	this.pressing[inputstate.BAR_0] = false
	this.pressing[inputstate.MAIN1] = false
	this.pressing[inputstate.MAIN2] = false
	this.pressing[inputstate.ACTIONBAR_USE] = false
	this.lock[inputstate.BAR_1] = true
	this.lock[inputstate.BAR_2] = true
	this.lock[inputstate.BAR_3] = true
	this.lock[inputstate.BAR_4] = true
	this.lock[inputstate.BAR_5] = true
	this.lock[inputstate.BAR_6] = true
	this.lock[inputstate.BAR_7] = true
	this.lock[inputstate.BAR_8] = true
	this.lock[inputstate.BAR_9] = true
	this.lock[inputstate.BAR_0] = true
	this.lock[inputstate.MAIN1] = true
	this.lock[inputstate.MAIN2] = true
	this.lock[inputstate.ACTIONBAR_USE] = true
}

func (this *InputState) UnLockActionBar() {
	this.lock[inputstate.BAR_1] = false
	this.lock[inputstate.BAR_2] = false
	this.lock[inputstate.BAR_3] = false
	this.lock[inputstate.BAR_4] = false
	this.lock[inputstate.BAR_5] = false
	this.lock[inputstate.BAR_6] = false
	this.lock[inputstate.BAR_7] = false
	this.lock[inputstate.BAR_8] = false
	this.lock[inputstate.BAR_9] = false
	this.lock[inputstate.BAR_0] = false
	this.lock[inputstate.MAIN1] = false
	this.lock[inputstate.MAIN2] = false
	this.lock[inputstate.ACTIONBAR_USE] = false
}

func (this *InputState) SetKeybindNames(msg common.MessageEngine) {
	this.bindingName[inputstate.CANCEL] = msg.Get("Cancel")
	this.bindingName[inputstate.ACCEPT] = msg.Get("Accept")
	this.bindingName[inputstate.UP] = msg.Get("Up")
	this.bindingName[inputstate.DOWN] = msg.Get("Down")
	this.bindingName[inputstate.LEFT] = msg.Get("Left")
	this.bindingName[inputstate.RIGHT] = msg.Get("Right")
	this.bindingName[inputstate.BAR_1] = msg.Get("Bar1")
	this.bindingName[inputstate.BAR_2] = msg.Get("Bar2")
	this.bindingName[inputstate.BAR_3] = msg.Get("Bar3")
	this.bindingName[inputstate.BAR_4] = msg.Get("Bar4")
	this.bindingName[inputstate.BAR_5] = msg.Get("Bar5")
	this.bindingName[inputstate.BAR_6] = msg.Get("Bar6")
	this.bindingName[inputstate.BAR_7] = msg.Get("Bar7")
	this.bindingName[inputstate.BAR_8] = msg.Get("Bar8")
	this.bindingName[inputstate.BAR_9] = msg.Get("Bar9")
	this.bindingName[inputstate.BAR_0] = msg.Get("Bar0")
	this.bindingName[inputstate.CHARACTER] = msg.Get("Character")
	this.bindingName[inputstate.INVENTORY] = msg.Get("Inventory")
	this.bindingName[inputstate.POWERS] = msg.Get("Powers")
	this.bindingName[inputstate.LOG] = msg.Get("Log")
	this.bindingName[inputstate.MAIN1] = msg.Get("Main1")
	this.bindingName[inputstate.CTRL] = msg.Get("Ctrl")
	this.bindingName[inputstate.SHIFT] = msg.Get("Shift")
	this.bindingName[inputstate.ALT] = msg.Get("Alt")
	this.bindingName[inputstate.DEL] = msg.Get("Delete")
	this.bindingName[inputstate.ACTIONBAR] = msg.Get("ActionBar Accept")
	this.bindingName[inputstate.ACTIONBAR_BACK] = msg.Get("ActionBar Left")
	this.bindingName[inputstate.ACTIONBAR_FORWARD] = msg.Get("ActionBar Right")
	this.bindingName[inputstate.ACTIONBAR_USE] = msg.Get("ActionBar Use")
	this.bindingName[inputstate.DEVELOPER_MENU] = msg.Get("Developer Menu")

	this.mouseButton[0] = msg.Get("Left Mouse")
	this.mouseButton[1] = msg.Get("Middle Mouse")
	this.mouseButton[2] = msg.Get("Right Mouse")
	this.mouseButton[3] = msg.Get("Wheel Up")
	this.mouseButton[4] = msg.Get("Wheel Down")
	this.mouseButton[5] = msg.Get("Mouse X1")
	this.mouseButton[6] = msg.Get("Mouse X2")
}

func (this *InputState) GenCode2Binding() {
	for key, val := range this.binding {
		this.code2binding[val] = key
	}
}

func (this *InputState) GetCode2Binding(code int) (int, bool) {
	key, ok := this.code2binding[code]
	if ok {
		return key, true
	}

	return 0, false
}

func (this *InputState) EnableEventlog() {
	this.dumpEvent = true
}

func (this *InputState) ScaleMouse(settings common.Settings, x, y uint) point.Point {
	if settings.GetMouseScaled() {
		return point.Construct((int)(x), (int)(y))
	}

	scaledMouse := point.Construct()
	offsetY := (int)((((float32)(settings.Get("resolution_h").(int)) - (float32)(settings.GetViewH())/settings.GetViewScaling()) / 2) * settings.GetViewScaling())
	scaledMouse.X = (int)(float32(x) * settings.GetViewScaling())
	scaledMouse.Y = (int)((float32)(y)*settings.GetViewScaling()) - offsetY

	return scaledMouse
}

// 窗口大小发生变化
func (this *InputState) GetWindowResized() bool {
	return this.windowResized
}

func (this *InputState) SetBinding(key, val int) {
	if _, ok := this.binding[key]; !ok {
		panic("binding has no such key")
	}

	this.binding[key] = val

}

func (this *InputState) GetKeyCount() int {
	return len(this.binding)
}

func (this *InputState) GetBinding(key int) int {
	return this.binding[key]
}

func (this *InputState) SetPressing(key int, val bool) {
	if _, ok := this.pressing[key]; !ok {
		panic("pressing has no such key")
	}

	this.pressing[key] = val
}

func (this *InputState) GetPressing(key int) bool {
	return this.pressing[key]
}

func (this *InputState) SetUnPress(key int, val bool) {
	if _, ok := this.unPress[key]; !ok {
		panic("unpress has no such key")
	}

	this.unPress[key] = val
}

// 业务占用该按键，锁住
func (this *InputState) SetLock(key int, val bool) {
	if _, ok := this.lock[key]; !ok {
		panic("lock has no such key")
	}

	this.lock[key] = val
}

func (this *InputState) GetLock(key int) bool {
	return this.lock[key]
}

func (this *InputState) DumpEvent() bool {
	return this.dumpEvent
}

func (this *InputState) SetInKeys(text string) {
	this.inKeys = text
}

func (this *InputState) GetInKeys() string {
	return this.inKeys
}

func (this *InputState) GetMouse() point.Point {
	return this.mouse
}

func (this *InputState) SetMouse(m point.Point) {
	this.mouse = m
}

func (this *InputState) SetScrollUp(val bool) {
	this.scrollUp = val
}

func (this *InputState) GetScrollUp() bool {
	return this.scrollUp
}

func (this *InputState) SetScrollDown(val bool) {
	this.scrollDown = val
}

func (this *InputState) GetScrollDown() bool {
	return this.scrollDown
}

func (this *InputState) SetLastButton(val int) {
	this.lastButton = val
}

func (this *InputState) GetWindowMinimized() bool {
	return this.windowMinimized
}

func (this *InputState) SetWindowMinimized(val bool) {
	this.windowMinimized = val
}

func (this *InputState) SetWindowResized(val bool) {
	this.windowResized = val
}

func (this *InputState) GetWindowRestored() bool {
	return this.windowRestored
}

func (this *InputState) SetWindowRestored(val bool) {
	this.windowRestored = val
}

func (this *InputState) GetDone() bool {
	return this.done
}

func (this *InputState) SetDone(val bool) {
	this.done = val
}

func (this *InputState) GetSlowRepeat(key int) bool {
	return this.slowRepeat[key]
}

func (this *InputState) GetRepeatCooldown(key int) *timer.Timer {
	return this.repeatCooldown[key]
}

func (this *InputState) GetMouseButtonNameCount() int {
	return len(this.mouseButton)
}

func (this *InputState) GetMouseButton(key int) string {
	return this.mouseButton[key]
}

func (this *InputState) GetBindingName(key int) string {
	return this.bindingName[key]
}

func (this *InputState) SetLockScroll(val bool) {
	this.lockScroll = val
}

func (this *InputState) GetLockScroll() bool {
	return this.lockScroll
}

func (this *InputState) SetLockAll(val bool) {
	this.lockAll = val
}

func (this *InputState) GetRefreshHotkeys() bool {
	return this.refreshHotkeys
}
