package menu

import (
	"errors"
	"fmt"
	"math"
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define"
	"monster/pkg/common/define/fontengine"
	"monster/pkg/common/define/game/menu/config"
	"monster/pkg/common/define/inputstate"
	"monster/pkg/common/define/modmanager"
	"monster/pkg/common/define/platform"
	"monster/pkg/common/define/widget/button"
	"monster/pkg/common/define/widget/checkbox"
	"monster/pkg/common/define/widget/listbox"
	"monster/pkg/common/define/widget/slider"
	"monster/pkg/common/gameres"
	"monster/pkg/common/labelinfo"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/common/timer"
	"monster/pkg/config/version"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/utils"
	"monster/pkg/utils/parsing"
	"monster/pkg/utils/tools"
	"sort"
	"strconv"
	"strings"
)

// ================== option ===================
// 每个选项
type ConfigOption struct {
	enabled bool
	label   common.WidgetLabel
	widget  common.Widget
}

func ConstructConfigOption() ConfigOption {
	co := ConfigOption{}
	co.Init()

	return co
}

func (this *ConfigOption) Init() *ConfigOption {
	return this
}

func (this *ConfigOption) Close() {
}

// ================= tab ===================
// 一个标签代表一堆组件
type ConfigTab struct {
	scrollbox    common.WidgetScrollBox // 滚动盒子
	enabledCount int
	options      []ConfigOption
}

func ConstructConfigTab() ConfigTab {
	ct := ConfigTab{}
	ct.Init()

	return ct
}

func (this *ConfigTab) Init() *ConfigTab {
	return this
}

func (this *ConfigTab) Close() {
	if this.scrollbox != nil {
		this.scrollbox.Close()
	}

	for _, ptr := range this.options {
		ptr.Close()
	}
}

func (this *ConfigTab) SetOptionWidgets(index int, lb common.WidgetLabel, w common.Widget, lbText string) {
	if !this.options[index].enabled {
		this.options[index].enabled = true
		this.enabledCount++
	}

	this.options[index].label = lb
	this.options[index].label.SetText(lbText)
	this.options[index].widget = w
	this.options[index].widget.SetTablistNavRight(true)
}

func (this *ConfigTab) SetOptionEnabled(index int, enable bool) {
	if this.options[index].enabled && !enable {
		this.options[index].enabled = false
		if this.enabledCount > 0 {
			this.enabledCount--
		}
	} else if !this.options[index].enabled && enable {
		this.options[index].enabled = true
		this.enabledCount++
	}
}

// 最大enabled序号
func (this *ConfigTab) GetEnabledIndex(optionIndex int) int {
	r := -1
	for i := 0; i < len(this.options); i++ {
		if this.options[i].enabled {
			r++
		}

		if i == optionIndex {
			break
		}
	}

	if r != -1 {
		return r
	}

	return 0
}

// 菜单
// ==================== config =====================
type Config struct {
	cfgTabs                []ConfigTab
	isGameState            bool
	enableGameStateButtons bool
	hero                   gameres.Avatar // 外部钩子
	frameLimits            []int
	virtualHeights         []int

	// 确认弹窗
	inputConfirm    *Confirm
	defaultsConfirm *Confirm

	// 组件组织
	optionTab   []int
	childWidget []common.Widget

	// 组件
	tabControl      common.WidgetTabControl // 标签选择控制器
	buttons         map[string]common.WidgetButton
	labels          map[string]common.WidgetLabel
	checkboxs       map[string]common.WidgetCheckBox
	sliders         map[string]common.WidgetSlider
	horizontalLists map[string]common.WidgetHorizontalList
	listboxs        map[string]common.WidgetListBox

	// 背景
	background common.Sprite // 菜单的背景

	// 其他
	activeTab                int
	frame                    point.Point
	frameOffset              point.Point
	tabOffset                point.Point // 标签区域的偏移坐标
	scrollpane               rect.Rect   // 滚动盒子的位置和子组件的位置和大小
	scrollpaneColor          color.Color // 滚动盒子的颜色
	scrollpanePadding        point.Point // 滚动盒子里的各组件的间距
	scrollpaneSeparatorColor color.Color
	secondaryOffset          point.Point
	languageISO              []string
	newRenderDevice          string
	videoModes               []rect.Rect
	keybindsLb               []common.WidgetLabel
	keybindsLstb             []common.WidgetHorizontalList
	inputConfirmTimer        timer.Timer
	keybindTipTimer          timer.Timer
	keybindTip               common.WidgetTooltip
	clickedAccept            bool
	clickedCancel            bool
	forceRefreshBackground   bool
	reloadMusic              bool
	clickedPauseContinue     bool
	clickedPauseExit         bool
	clickedPauseSave         bool
}

func NewConfig(modules common.Modules, isGameState bool) *Config {
	c := &Config{}
	c.Init(modules, isGameState)

	return c
}

func (this *Config) Init(modules common.Modules, isGameState bool) gameres.MenuConfig {
	widgetf := modules.Widgetf()
	msg := modules.Msg()
	render := modules.Render()
	settings := modules.Settings()
	mods := modules.Mods()
	font := modules.Font()
	eset := modules.Eset()

	this.isGameState = isGameState
	this.buttons = map[string]common.WidgetButton{}
	this.labels = map[string]common.WidgetLabel{}
	this.checkboxs = map[string]common.WidgetCheckBox{}
	this.sliders = map[string]common.WidgetSlider{}
	this.horizontalLists = map[string]common.WidgetHorizontalList{}
	this.listboxs = map[string]common.WidgetListBox{}

	// 标签控制器
	this.tabControl = widgetf.New("tabcontrol").(common.WidgetTabControl).Init(modules)

	this.buttons["ok"] = widgetf.New("button").(common.WidgetButton).Init(modules, button.DEFAULT_FILE)
	this.buttons["defaults"] = widgetf.New("button").(common.WidgetButton).Init(modules, button.DEFAULT_FILE)
	this.buttons["cancel"] = widgetf.New("button").(common.WidgetButton).Init(modules, button.DEFAULT_FILE)

	// 确认弹窗
	this.inputConfirm = NewConfirm(modules, msg.Get("Clear"), msg.Get("Assign:"))
	this.defaultsConfirm = NewConfirm(modules, msg.Get("Defaults"), msg.Get("Reset ALL settings?"))

	// 定义组件
	this.labels["pause_continue"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.buttons["pause_continue"] = widgetf.New("button").(common.WidgetButton).Init(modules, button.DEFAULT_FILE)
	this.labels["pause_exit"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.buttons["pause_exit"] = widgetf.New("button").(common.WidgetButton).Init(modules, button.DEFAULT_FILE)
	this.labels["pause_save"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.buttons["pause_save"] = widgetf.New("button").(common.WidgetButton).Init(modules, button.DEFAULT_FILE)
	this.labels["pause_time"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.labels["pause_time_text"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.labels["renderer"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.horizontalLists["renderer"] = widgetf.New("horizontallist").(common.WidgetHorizontalList).Init(modules)
	this.labels["fullscreen"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["fullscreen"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["hwsurface"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["hwsurface"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["vsync"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["vsync"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["texture_filter"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["texture_filter"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["dpi_scaling"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["dpi_scaling"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["parallax_layers"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["parallax_layers"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["change_gamma"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["change_gamma"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["gamma"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.sliders["gamma"] = widgetf.New("slider").(common.WidgetSlider).Init(modules, slider.DEFAULT_FILE)
	this.labels["frame_limit"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.horizontalLists["frame_limit"] = widgetf.New("horizontallist").(common.WidgetHorizontalList).Init(modules)
	this.labels["max_render_size"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.horizontalLists["max_render_size"] = widgetf.New("horizontallist").(common.WidgetHorizontalList).Init(modules)
	this.labels["music_volume"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.sliders["music_volume"] = widgetf.New("slider").(common.WidgetSlider).Init(modules, slider.DEFAULT_FILE)
	this.labels["sound_volume"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.sliders["sound_volume"] = widgetf.New("slider").(common.WidgetSlider).Init(modules, slider.DEFAULT_FILE)
	this.labels["show_fps"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["show_fps"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["hardware_cursor"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["hardware_cursor"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["colorblind"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["colorblind"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["dev_mod"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["dev_mod"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["subtitles"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["subtitles"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["loot_tooltip"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.horizontalLists["loot_tooltip"] = widgetf.New("horizontallist").(common.WidgetHorizontalList).Init(modules)
	this.labels["minimap"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.horizontalLists["minimap"] = widgetf.New("horizontallist").(common.WidgetHorizontalList).Init(modules)
	this.labels["statbar_labels"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["statbar_labels"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["statbar_autohide"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["statbar_autohide"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["combat_text"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["combat_text"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["auto_equip"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["auto_equip"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["entity_markers"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["entity_markers"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["low_hp_warning"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.horizontalLists["low_hp_warning"] = widgetf.New("horizontallist").(common.WidgetHorizontalList).Init(modules)
	this.labels["low_hp_threshold"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.horizontalLists["low_hp_threshold"] = widgetf.New("horizontallist").(common.WidgetHorizontalList).Init(modules)
	this.labels["item_compare_tips"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["item_compare_tips"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["joystick_device"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.horizontalLists["joystick_device"] = widgetf.New("horizontallist").(common.WidgetHorizontalList).Init(modules)
	this.labels["mouse_move"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["mouse_move"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["mouse_aim"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["mouse_aim"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["no_mouse"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["no_mouse"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["mouse_move_swap"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["mouse_move_swap"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["mouse_move_attack"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["mouse_move_attack"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["joystick_deadzone"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.sliders["joystick_deadzone"] = widgetf.New("slider").(common.WidgetSlider).Init(modules, slider.DEFAULT_FILE)
	this.labels["touch_controls"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.checkboxs["touch_controls"] = widgetf.New("checkbox").(common.WidgetCheckBox).Init(modules, checkbox.DEFAULT_FILE)
	this.labels["touch_scale"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.sliders["touch_scale"] = widgetf.New("slider").(common.WidgetSlider).Init(modules, slider.DEFAULT_FILE)
	this.labels["activemods"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.listboxs["activemods"] = widgetf.New("listbox").(common.WidgetListBox).Init(modules, 10, listbox.DEFAULT_FILE)
	this.labels["inactivemods"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.listboxs["inactivemods"] = widgetf.New("listbox").(common.WidgetListBox).Init(modules, 10, listbox.DEFAULT_FILE)
	this.labels["language"] = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.horizontalLists["language"] = widgetf.New("horizontallist").(common.WidgetHorizontalList).Init(modules)
	this.buttons["activemods_shiftup"] = widgetf.New("button").(common.WidgetButton).Init(modules, "images/menus/buttons/up.png")
	this.buttons["activemods_shiftdown"] = widgetf.New("button").(common.WidgetButton).Init(modules, "images/menus/buttons/down.png")
	this.buttons["activemods_deactivate"] = widgetf.New("button").(common.WidgetButton).Init(modules, button.DEFAULT_FILE)
	this.buttons["inactivemods_activate"] = widgetf.New("button").(common.WidgetButton).Init(modules, button.DEFAULT_FILE)

	// 其他
	this.frame = point.Construct(0, 0)
	this.frameOffset = point.Construct(11, 8)
	this.tabOffset = point.Construct(3, 0)
	this.scrollpaneColor = color.Construct(0, 0, 0, 0)
	this.scrollpanePadding = point.Construct(8, 40)
	this.scrollpaneSeparatorColor = font.GetColor(fontengine.COLOR_WIDGET_DISABLED)
	this.newRenderDevice = settings.Get("renderer").(string)
	this.inputConfirmTimer = timer.Construct((uint)(settings.Get("max_fps").(int) * 10))
	this.keybindTipTimer = timer.Construct((uint)(settings.Get("max_fps").(int) * 5))
	this.keybindTip = widgetf.New("tooltip").(common.WidgetTooltip).Init(modules)

	graphics, err := render.LoadImage(settings, mods, "images/menus/config.png")
	if err != nil {
		panic(err)
	}
	defer graphics.UnRef()

	this.background, err = graphics.CreateSprite()
	if err != nil {
		panic(err)
	}

	this.buttons["ok"].SetLabel(modules, msg.Get("OK"))
	this.buttons["defaults"].SetLabel(modules, msg.Get("Defaults"))
	this.buttons["cancel"].SetLabel(modules, msg.Get("Cancel"))
	this.buttons["pause_continue"].SetLabel(modules, msg.Get("Continue"))
	this.SetPauseExitText(modules, true)
	this.buttons["pause_save"].SetLabel(modules, msg.Get("Save Game"))
	this.SetPauseSaveEnabled(modules, true)
	this.labels["pause_time"].SetText(utils.GetTimeString(0))
	this.labels["pause_time"].SetJustify(fontengine.JUSTIFY_RIGHT)
	this.labels["pause_time"].SetValign(labelinfo.VALIGN_CENTER)

	// mods的列表框
	this.listboxs["activemods"].SetMultiSelect(true)
	modList := mods.GetModList()
	for _, mod := range modList {
		if mod.GetName() != modmanager.FALLBACK_MOD {
			this.listboxs["activemods"].Append(modules, mod.GetName(), this.CreateModTooltip(modules, mod))
		}
	}

	this.listboxs["inactivemods"].SetMultiSelect(true)
	modDirs := mods.GetModDirs() // 获取mod文件里的全部目录
	for _, dir := range modDirs {
		skipMod := false

		// 跳过已经加载的
		for _, mod := range modList {
			if dir == mod.GetName() {
				skipMod = true
				break
			}
		}

		// 未加载的mod
		if !skipMod && dir != modmanager.FALLBACK_MOD {
			tempMod, err := mods.LoadMod(dir)
			if err != nil {
				panic(err)
			}

			this.listboxs["inactivemods"].Append(modules, dir, this.CreateModTooltip(modules, tempMod))
		}
	}
	this.listboxs["inactivemods"].Sort()

	// 按键绑定
	for i := 0; i < inputstate.KEY_COUNT_USER; i++ {
		this.keybindsLb = append(this.keybindsLb, widgetf.New("label").(common.WidgetLabel).Init(modules))
		newlstb := widgetf.New("horizontallist").(common.WidgetHorizontalList).Init(modules)
		newlstb.SetHasAction(true)
		this.keybindsLstb = append(this.keybindsLstb, newlstb)
	}

	// 战利品提示
	this.horizontalLists["loot_tooltip"].Append(
		msg.Get("Default"),
		msg.Get("Show all loot tooltips, except for those that would be obscured by the player or an enemy. Temporarily show all loot tooltips with 'Alt'."),
	)

	this.horizontalLists["loot_tooltip"].Append(
		msg.Get("Show all"),
		msg.Get("Always show loot tooltips. Temporarily hide all loot tooltips with 'Alt'."),
	)

	this.horizontalLists["loot_tooltip"].Append(
		msg.Get("Hidden"),
		msg.Get("Always hide loot tooltips, except for when a piece of loot is hovered with the mouse cursor. Temporarily show all loot tooltips with 'Alt'."),
	)

	// 小地图提示
	this.horizontalLists["minimap"].Append(
		msg.Get("Visible"),
		"",
	)

	this.horizontalLists["minimap"].Append(
		msg.Get("Visible (2x zoom)"),
		"",
	)

	this.horizontalLists["minimap"].Append(
		msg.Get("Hidden"),
		"",
	)

	lhpwPrefix := msg.Get("Controls the type of warning to be activated when the player is below the low health threshold.")
	lhpwWarning1 := msg.Get("- Display a message")
	lhpwWarning2 := msg.Get("- Play a sound")
	lhpwWarning3 := msg.Get("- Change the cursor")

	this.horizontalLists["low_hp_warning"].Append(msg.Get("Disabled"), lhpwPrefix)
	this.horizontalLists["low_hp_warning"].Append(msg.Get("All"), lhpwPrefix+"\n\n"+lhpwWarning1+"\n"+lhpwWarning2+"\n"+lhpwWarning3)
	this.horizontalLists["low_hp_warning"].Append(msg.Get("Message & Cursor"), lhpwPrefix+"\n\n"+lhpwWarning1+"\n"+lhpwWarning3)
	this.horizontalLists["low_hp_warning"].Append(msg.Get("Message & Sound"), lhpwPrefix+"\n\n"+lhpwWarning1+"\n"+lhpwWarning2)
	this.horizontalLists["low_hp_warning"].Append(msg.Get("Sound & Cursor"), lhpwPrefix+"\n\n"+lhpwWarning2+"\n"+lhpwWarning3)
	this.horizontalLists["low_hp_warning"].Append(msg.Get("Message"), lhpwPrefix+"\n\n"+lhpwWarning1)
	this.horizontalLists["low_hp_warning"].Append(msg.Get("Cursor"), lhpwPrefix+"\n\n"+lhpwWarning3)
	this.horizontalLists["low_hp_warning"].Append(msg.Get("Sound"), lhpwPrefix+"\n\n"+lhpwWarning2)

	for i := 1; i < 10; i++ {
		this.horizontalLists["low_hp_threshold"].Append(strconv.Itoa(i*5)+"%", msg.Get("When the player's health drops below the given threshold, the low health notifications are triggered if one or more of them is enabled."))
	}

	// 帧上限
	this.frameLimits = append(this.frameLimits, 30)
	this.frameLimits = append(this.frameLimits, 60)
	this.frameLimits = append(this.frameLimits, 120)
	this.frameLimits = append(this.frameLimits, 240)

	if !tools.FindInt(this.frameLimits, settings.Get("max_fps").(int)) {
		this.frameLimits = append(this.frameLimits, settings.Get("max_fps").(int))
	}

	refreshRate := render.GetRefreshRate()
	if refreshRate > 0 && !tools.FindInt(this.frameLimits, refreshRate) {
		this.frameLimits = append(this.frameLimits, refreshRate)
	}

	sort.Slice(this.frameLimits, func(i, j int) bool { return this.frameLimits[i] < this.frameLimits[j] })

	for _, limit := range this.frameLimits {
		this.horizontalLists["frame_limit"].Append(strconv.Itoa(limit), msg.Get("The maximum frame rate that the game will be allowed to run at."))
	}

	maxRenderSizeTooltip := msg.Get("The render size refers to the height in pixels of the surface used to draw the game. Mods define the allowed render sizes, but this option allows overriding the maximum size.")
	this.horizontalLists["max_render_size"].Append("Default", maxRenderSizeTooltip)

	this.virtualHeights = eset.Get("resolutions", "virtual_height").([]int)

	if settings.Get("max_render_size").(int) > 0 && !tools.FindInt(this.virtualHeights, settings.Get("max_render_size").(int)) {
		this.virtualHeights = append(this.virtualHeights, settings.Get("max_render_size").(int))
	}

	sort.Slice(this.virtualHeights, func(i, j int) bool { return this.virtualHeights[i] < this.virtualHeights[j] })

	for _, height := range this.virtualHeights {
		this.horizontalLists["max_render_size"].Append(strconv.Itoa(height), maxRenderSizeTooltip)
	}

	// 组织组件
	err = this.arrange(modules)
	if err != nil {
		panic(err)
	}
	render.SetBackgroundColor(color.Construct(0, 0, 0, 0))
	return this
}

func (this *Config) clear() {
	// 背景图片精灵
	if this.background != nil {
		this.background.Close()
		this.background = nil
	}

	// 关系
	for index, _ := range this.cfgTabs {
		this.cfgTabs[index].Close()
	}

	this.cfgTabs = nil
	this.childWidget = nil

	// 弹窗
	if this.inputConfirm != nil {
		this.inputConfirm.Close()
		this.inputConfirm = nil
	}

	if this.defaultsConfirm != nil {
		this.defaultsConfirm.Close()
		this.defaultsConfirm = nil
	}

	// 标签控制器
	if this.tabControl != nil {
		this.tabControl.Close()
		this.tabControl = nil
	}

	// 组件
	for _, ptr := range this.buttons {
		ptr.Close()
	}
	this.buttons = nil

	for _, ptr := range this.labels {
		ptr.Close()
	}

	this.labels = nil
	for _, ptr := range this.checkboxs {
		ptr.Close()
	}

	this.checkboxs = nil
	for _, ptr := range this.sliders {
		ptr.Close()
	}

	this.sliders = nil
	for _, ptr := range this.horizontalLists {
		ptr.Close()
	}

	this.horizontalLists = nil
	for _, ptr := range this.listboxs {
		ptr.Close()
	}

	this.listboxs = nil
	for _, ptr := range this.keybindsLb {
		ptr.Close()
	}

	this.keybindsLb = nil
	for _, ptr := range this.keybindsLstb {
		ptr.Close()
	}
	this.keybindsLstb = nil

	if this.keybindTip != nil {
		this.keybindTip.Close()
	}
	this.keybindTip = nil
	this.languageISO = nil
}

func (this *Config) Close() {
	this.clear()
}

func (this *Config) CreateModTooltip(modules common.Modules, mod common.Mod) string {
	settings := modules.Settings()
	msg := modules.Msg()

	ret := ""

	if mod != nil {
		modVer := ""
		if mod.GetVersion() != version.MIN {
			modVer = mod.GetVersion().GetString()
		}

		engineVer := version.CreateVersionReqString(mod.GetEngineMinVersion(), mod.GetEngineMaxVersion())
		ret = mod.GetName() + "\n"

		modDescription := mod.GetLocalDescription(settings.Get("language").(string))
		if modDescription != "" {
			ret += "\n"
			ret += modDescription + "\n"
		}

		middleSection := false
		_ = middleSection

		if modVer != "" {
			middleSection = true
			ret += "\n"
			ret += msg.Get("Version:") + " " + modVer
		}

		if mod.GetGame() != "" && mod.GetGame() != modmanager.FALLBACK_GAME {
			middleSection = true
			ret += "\n"
			ret += msg.Get("Game:") + " " + mod.GetGame()
		}

		if engineVer != "" {
			middleSection = true
			ret += "\n"
			ret += msg.Get("Engine version:") + " " + engineVer
		}

		if middleSection {
			ret += "\n"
		}

		modDepends := mod.GetDepends()
		if len(modDepends) != 0 {
			ret += "\n"
			ret += msg.Get("Requires mods:") + "\n"
			dependsMin := mod.GetDependsMin()
			dependsMax := mod.GetDependsMax()
			for i, depend := range modDepends {
				ret += "-  " + depend

				dependVer := version.CreateVersionReqString(dependsMin[i], dependsMax[i])
				if dependVer != "" {
					ret += " (" + dependVer + ")"
				}

				if i < len(modDepends)-1 {
					ret += "\n"
				}

			}
		}

		if ret != "" && strings.HasSuffix(ret, "\n") {
			ret = strings.TrimSuffix(ret, "\n")
		}
	}

	return ret
}

func (this *Config) SetPauseExitText(modules common.Modules, enableSave bool) {
	eset := modules.Eset()
	msg := modules.Msg()

	if eset.Get("misc", "save_onexit").(bool) && enableSave {
		this.buttons["pause_exit"].SetLabel(modules, msg.Get("Save & Exit"))
	} else {
		this.buttons["pause_exit"].SetLabel(modules, msg.Get("Exit"))
	}
}

func (this *Config) SetPauseSaveEnabled(modules common.Modules, enableSave bool) {
	eset := modules.Eset()

	this.buttons["pause_save"].SetEnabled(enableSave && eset.Get("misc", "save_anywhere").(bool))
}

func (this *Config) ReadConfig(modules common.Modules) error {
	mods := modules.Mods()
	msg := modules.Msg()

	infile := fileparser.New()
	err := infile.Open("menus/config.txt", true, mods)
	if err != nil && !utils.IsNotExist(err) {
		return err
	}
	defer infile.Close()

	for infile.Next(mods) {
		x := 0
		y := 0
		a := 0
		strVal := ""
		first := ""
		switch infile.Key() {
		case "button_ok":
			x, strVal = parsing.PopFirstInt(infile.Val(), "")
			y, strVal = parsing.PopFirstInt(strVal, "")
			first, strVal = parsing.PopFirstString(strVal, "")
			a = parsing.ToAlignment(first, define.ALIGN_TOPLEFT)
			this.buttons["ok"].SetPosBase(x, y, a)
		case "button_defaults":
			x, strVal = parsing.PopFirstInt(infile.Val(), "")
			y, strVal = parsing.PopFirstInt(strVal, "")
			first, strVal = parsing.PopFirstString(strVal, "")
			a = parsing.ToAlignment(first, define.ALIGN_TOPLEFT)
			this.buttons["defaults"].SetPosBase(x, y, a)
		case "button_cancel":
			x, strVal = parsing.PopFirstInt(infile.Val(), "")
			y, strVal = parsing.PopFirstInt(strVal, "")
			first, strVal = parsing.PopFirstString(strVal, "")
			a = parsing.ToAlignment(first, define.ALIGN_TOPLEFT)
			this.buttons["cancel"].SetPosBase(x, y, a)
		default:
			x1, y1, x2, y2 := 0, 0, 0, 0
			x1, strVal = parsing.PopFirstInt(infile.Val(), "")
			y1, strVal = parsing.PopFirstInt(strVal, "")
			x2, strVal = parsing.PopFirstInt(strVal, "")
			y2, strVal = parsing.PopFirstInt(strVal, "")

			switch infile.Key() {
			case "listbox_scrollbar_offset":
				this.listboxs["inactivemods"].SetScrollbarOffset(x1)
				this.listboxs["activemods"].SetScrollbarOffset(x1)
			case "frame_offset":
				this.frameOffset.X = x1
				this.frameOffset.Y = y1
			case "tab_offset":
				this.tabOffset.X = x1
				this.tabOffset.Y = y1
			case "activemods":
				this.labels["activemods"].SetPosBase(x1, y1, define.ALIGN_TOPLEFT)
				this.labels["activemods"].SetText(msg.Get("Active Mods"))
				this.labels["activemods"].SetJustify(fontengine.JUSTIFY_LEFT)
				this.listboxs["activemods"].SetPosBase(x2, y2, define.ALIGN_TOPLEFT)
			case "activemods_height":
				this.listboxs["activemods"].SetHeight(modules, x1)
			case "inactivemods":
				this.labels["inactivemods"].SetPosBase(x1, y1, define.ALIGN_TOPLEFT)
				this.labels["inactivemods"].SetText(msg.Get("Active Mods"))
				this.labels["inactivemods"].SetJustify(fontengine.JUSTIFY_LEFT)
				this.listboxs["inactivemods"].SetPosBase(x2, y2, define.ALIGN_TOPLEFT)
			case "inactivemods_height":
				this.listboxs["inactivemods"].SetHeight(modules, x1)
			case "activemods_shiftup":
				this.buttons["activemods_shiftup"].SetPosBase(x1, y1, define.ALIGN_TOPLEFT)
				this.buttons["activemods_shiftup"].Refresh(modules)
			case "activemods_shiftdown":
				this.buttons["activemods_shiftdown"].SetPosBase(x1, y1, define.ALIGN_TOPLEFT)
				this.buttons["activemods_shiftdown"].Refresh(modules)
			case "activemods_deactivate":
				this.buttons["activemods_deactivate"].SetLabel(modules, msg.Get("<< Disable"))
				this.buttons["activemods_deactivate"].SetPosBase(x1, y1, define.ALIGN_TOPLEFT)
				this.buttons["activemods_deactivate"].Refresh(modules)
			case "inactivemods_activate":
				this.buttons["inactivemods_activate"].SetLabel(modules, msg.Get("Enable >>"))
				this.buttons["inactivemods_activate"].SetPosBase(x1, y1, define.ALIGN_TOPLEFT)
				this.buttons["inactivemods_activate"].Refresh(modules)
			case "scrollpane":
				this.scrollpane.X = x1
				this.scrollpane.Y = y1
				this.scrollpane.W = x2
				this.scrollpane.H = y2
			case "scrollpane_padding":
				this.scrollpanePadding.X = x1
				this.scrollpanePadding.Y = y1
			case "scrollpane_separator_color":
				this.scrollpaneSeparatorColor.R = (uint8)(x1)
				this.scrollpaneSeparatorColor.G = (uint8)(y1)
				this.scrollpaneSeparatorColor.B = (uint8)(x2)
			case "scrollpane_bg_color":
				this.scrollpaneColor.R = (uint8)(x1)
				this.scrollpaneColor.G = (uint8)(y1)
				this.scrollpaneColor.B = (uint8)(x2)
				this.scrollpaneColor.A = (uint8)(y2)
			default:
				return errors.New(fmt.Sprintf("MenuConfig: %s is not a valid key.", infile.Key()))
			}
		}
	}

	this.checkboxs["hwsurface"].SetTooltip(msg.Get("Will try to store surfaces in video memory versus system memory. The effect this has on performance depends on the renderer."))
	this.checkboxs["vsync"].SetTooltip(msg.Get("Prevents screen tearing. Disable if you experience \"stuttering\" in windowed mode or input lag."))
	this.checkboxs["dpi_scaling"].SetTooltip(msg.Get("When enabled, this uses the screen DPI in addition to the window dimensions to scale the rendering resolution. Otherwise, only the window dimensions are used."))
	this.checkboxs["parallax_layers"].SetTooltip(msg.Get("This enables parallax (non-tile) layers. Disabling this setting can improve performance in some cases."))
	this.checkboxs["change_gamma"].SetTooltip(msg.Get("Enables the below setting that controls the screen gamma level. The behavior of the gamma setting can vary between platforms."))
	this.checkboxs["colorblind"].SetTooltip(msg.Get("Provides additional text for information that is primarily conveyed through color."))
	this.checkboxs["statbar_autohide"].SetTooltip(msg.Get("Some mods will automatically hide the stat bars when they are inactive. Disabling this option will keep them displayed at all times."))
	this.checkboxs["auto_equip"].SetTooltip(msg.Get("When enabled, empty equipment slots will be filled with applicable items when they are obtained."))
	this.checkboxs["entity_markers"].SetTooltip(msg.Get("Shows a marker above enemies, allies, and the player when they are obscured by tall objects."))
	this.checkboxs["item_compare_tips"].SetTooltip(msg.Get("When enabled, tooltips for equipped items of the same type are shown next to standard item tooltips."))
	this.checkboxs["no_mouse"].SetTooltip(msg.Get("This allows the game to be controlled entirely with the keyboard (or joystick)."))
	this.checkboxs["mouse_move_swap"].SetTooltip(msg.Get("When 'Move hero using mouse' is enabled, this setting controls if 'Main1' or 'Main2' is used to move the hero. If enabled, 'Main2' will move the hero instead of 'Main1'."))
	this.checkboxs["mouse_move_attack"].SetTooltip(msg.Get("When 'Move hero using mouse' is enabled, this setting controls if the Power assigned to the movement button can be used by targeting an enemy. If this setting is disabled, it is required to use 'Shift' to access the Power assigned to the movement button."))
	this.checkboxs["mouse_aim"].SetTooltip(msg.Get("The player's attacks will be aimed in the direction of the mouse cursor when this is enabled."))
	this.checkboxs["touch_controls"].SetTooltip(msg.Get("When enabled, a virtual gamepad will be added in-game. Other interactions, such as drag-and-drop behavior, are also altered to better suit touch input."))

	return nil
}

func (this *Config) AddChildWidget(w common.Widget, tab int) {
	this.childWidget = append(this.childWidget, w)
	this.optionTab = append(this.optionTab, tab)
}

func (this *Config) arrange(modules common.Modules) error {
	msg := modules.Msg()
	eset := modules.Eset()
	inpt := modules.Inpt()
	plat := modules.Platform()
	widgetf := modules.Widgetf()

	// 标签控制器
	this.tabControl.SetTabTitle(modules, config.EXIT_TAB, msg.Get("Exit"))
	this.tabControl.SetTabTitle(modules, config.VIDEO_TAB, msg.Get("Video"))
	this.tabControl.SetTabTitle(modules, config.AUDIO_TAB, msg.Get("Audio"))
	this.tabControl.SetTabTitle(modules, config.INTERFACE_TAB, msg.Get("Interface"))
	this.tabControl.SetTabTitle(modules, config.INPUT_TAB, msg.Get("Input"))
	this.tabControl.SetTabTitle(modules, config.KEYBINDS_TAB, msg.Get("Keybindings"))
	this.tabControl.SetTabTitle(modules, config.MODS_TAB, msg.Get("Mods"))

	err := this.ReadConfig(modules)
	if err != nil {
		return err
	}

	// 分配
	this.cfgTabs = make([]ConfigTab, 6)
	this.cfgTabs[config.EXIT_TAB].options = make([]ConfigOption, 4)
	this.cfgTabs[config.VIDEO_TAB].options = make([]ConfigOption, platform.VIDEO_COUNT)
	this.cfgTabs[config.AUDIO_TAB].options = make([]ConfigOption, platform.AUDIO_COUNT)
	this.cfgTabs[config.INTERFACE_TAB].options = make([]ConfigOption, platform.INTERFACE_COUNT)
	this.cfgTabs[config.INPUT_TAB].options = make([]ConfigOption, platform.INPUT_COUNT)
	this.cfgTabs[config.KEYBINDS_TAB].options = make([]ConfigOption, inputstate.KEY_COUNT_USER)

	// ======== exit tab ========
	this.cfgTabs[config.EXIT_TAB].SetOptionWidgets(config.EXIT_OPTION_CONTINUE, this.labels["pause_continue"], this.buttons["pause_continue"], msg.Get("Paused"))
	this.cfgTabs[config.EXIT_TAB].SetOptionWidgets(config.EXIT_OPTION_SAVE, this.labels["pause_save"], this.buttons["pause_save"], msg.Get(""))
	this.cfgTabs[config.EXIT_TAB].SetOptionWidgets(config.EXIT_OPTION_EXIT, this.labels["pause_exit"], this.buttons["pause_exit"], msg.Get(""))
	this.cfgTabs[config.EXIT_TAB].SetOptionWidgets(config.EXIT_OPTION_TIME_PLAYED, this.labels["pause_time"], this.labels["pause_time_text"], msg.Get("Time Played"))

	if !eset.Get("misc", "save_anywhere").(bool) {
		this.cfgTabs[config.EXIT_TAB].SetOptionEnabled(config.EXIT_OPTION_SAVE, false)
	}

	// ======== video tab ========
	this.cfgTabs[config.VIDEO_TAB].SetOptionWidgets(platform.VIDEO_RENDERER, this.labels["renderer"], this.horizontalLists["renderer"], msg.Get("Renderer"))
	this.cfgTabs[config.VIDEO_TAB].SetOptionWidgets(platform.VIDEO_FULLSCREEN, this.labels["fullscreen"], this.checkboxs["fullscreen"], msg.Get("Full Screen Mode"))
	this.cfgTabs[config.VIDEO_TAB].SetOptionWidgets(platform.VIDEO_HWSURFACE, this.labels["hwsurface"], this.checkboxs["hwsurface"], msg.Get("Hardware surfaces"))
	this.cfgTabs[config.VIDEO_TAB].SetOptionWidgets(platform.VIDEO_VSYNC, this.labels["vsync"], this.checkboxs["vsync"], msg.Get("V-Sync"))
	this.cfgTabs[config.VIDEO_TAB].SetOptionWidgets(platform.VIDEO_TEXTURE_FILTER, this.labels["texture_filter"], this.checkboxs["texture_filter"], msg.Get("Texture Filtering"))
	this.cfgTabs[config.VIDEO_TAB].SetOptionWidgets(platform.VIDEO_DPI_SCALING, this.labels["dpi_scaling"], this.checkboxs["dpi_scaling"], msg.Get("DPI scaling"))
	this.cfgTabs[config.VIDEO_TAB].SetOptionWidgets(platform.VIDEO_PARALLAX_LAYERS, this.labels["parallax_layers"], this.checkboxs["parallax_layers"], msg.Get("Parallax Layers"))
	this.cfgTabs[config.VIDEO_TAB].SetOptionWidgets(platform.VIDEO_ENABLE_GAMMA, this.labels["change_gamma"], this.checkboxs["change_gamma"], msg.Get("Allow changing gamma"))
	this.cfgTabs[config.VIDEO_TAB].SetOptionWidgets(platform.VIDEO_GAMMA, this.labels["gamma"], this.sliders["gamma"], msg.Get("Gamma"))
	this.cfgTabs[config.VIDEO_TAB].SetOptionWidgets(platform.VIDEO_MAX_RENDER_SIZE, this.labels["max_render_size"], this.horizontalLists["max_render_size"], msg.Get("aximum Render Size"))
	this.cfgTabs[config.VIDEO_TAB].SetOptionWidgets(platform.VIDEO_FRAME_LIMIT, this.labels["frame_limit"], this.horizontalLists["frame_limit"], msg.Get("Frame Limit"))

	// ======== audio tab ========
	this.cfgTabs[config.AUDIO_TAB].SetOptionWidgets(platform.AUDIO_SFX, this.labels["sound_volume"], this.sliders["sound_volume"], msg.Get("Sound Volume"))
	this.cfgTabs[config.AUDIO_TAB].SetOptionWidgets(platform.AUDIO_MUSIC, this.labels["music_volume"], this.sliders["music_volume"], msg.Get("Music Volume"))

	// ======== interface tab ========
	this.cfgTabs[config.INTERFACE_TAB].SetOptionWidgets(platform.INTERFACE_LANGUAGE, this.labels["language"], this.horizontalLists["language"], msg.Get("Language"))
	this.cfgTabs[config.INTERFACE_TAB].SetOptionWidgets(platform.INTERFACE_SHOW_FPS, this.labels["show_fps"], this.checkboxs["show_fps"], msg.Get("Show FPS"))
	this.cfgTabs[config.INTERFACE_TAB].SetOptionWidgets(platform.INTERFACE_HARDWARE_CURSOR, this.labels["hardware_cursor"], this.checkboxs["hardware_cursor"], msg.Get("Hardware mouse cursor"))
	this.cfgTabs[config.INTERFACE_TAB].SetOptionWidgets(platform.INTERFACE_COLORBLIND, this.labels["colorblind"], this.checkboxs["colorblind"], msg.Get("Colorblind Mode"))
	this.cfgTabs[config.INTERFACE_TAB].SetOptionWidgets(platform.INTERFACE_DEV_MODE, this.labels["dev_mod"], this.checkboxs["dev_mod"], msg.Get("Developer Mode"))
	this.cfgTabs[config.INTERFACE_TAB].SetOptionWidgets(platform.INTERFACE_SUBTITLES, this.labels["subtitles"], this.checkboxs["subtitles"], msg.Get("Subtitles"))
	this.cfgTabs[config.INTERFACE_TAB].SetOptionWidgets(platform.INTERFACE_LOOT_TOOLTIPS, this.labels["loot_tooltip"], this.horizontalLists["loot_tooltip"], msg.Get("Loot tooltip visibility"))
	this.cfgTabs[config.INTERFACE_TAB].SetOptionWidgets(platform.INTERFACE_MINIMAP_MODE, this.labels["minimap"], this.horizontalLists["minimap"], msg.Get("Mini-map mode"))
	this.cfgTabs[config.INTERFACE_TAB].SetOptionWidgets(platform.INTERFACE_STATBAR_LABELS, this.labels["statbar_labels"], this.checkboxs["statbar_labels"], msg.Get("Always show stat bar labels"))
	this.cfgTabs[config.INTERFACE_TAB].SetOptionWidgets(platform.INTERFACE_STATBAR_AUTOHIDE, this.labels["statbar_autohide"], this.checkboxs["statbar_autohide"], msg.Get("Allow stat bar auto-hiding"))
	this.cfgTabs[config.INTERFACE_TAB].SetOptionWidgets(platform.INTERFACE_COMBAT_TEXT, this.labels["combat_text"], this.checkboxs["combat_text"], msg.Get("Show combat text"))
	this.cfgTabs[config.INTERFACE_TAB].SetOptionWidgets(platform.INTERFACE_AUTO_EQUIP, this.labels["auto_equip"], this.checkboxs["auto_equip"], msg.Get("Automatically equip items"))
	this.cfgTabs[config.INTERFACE_TAB].SetOptionWidgets(platform.INTERFACE_ENTITY_MARKERS, this.labels["entity_markers"], this.checkboxs["entity_markers"], msg.Get("Show hidden entity markers"))
	this.cfgTabs[config.INTERFACE_TAB].SetOptionWidgets(platform.INTERFACE_LOW_HP_WARNING_TYPE, this.labels["low_hp_warning"], this.horizontalLists["low_hp_warning"], msg.Get("Low health notification"))
	this.cfgTabs[config.INTERFACE_TAB].SetOptionWidgets(platform.INTERFACE_LOW_HP_THRESHOLD, this.labels["low_hp_threshold"], this.horizontalLists["low_hp_threshold"], msg.Get("Low health threshold"))
	this.cfgTabs[config.INTERFACE_TAB].SetOptionWidgets(platform.INTERFACE_ITEM_COMPARE_TIPS, this.labels["item_compare_tips"], this.checkboxs["item_compare_tips"], msg.Get("Show item comparison tooltips"))

	// ========= input tab =========
	this.cfgTabs[config.INPUT_TAB].SetOptionWidgets(platform.INPUT_JOYSTICK, this.labels["joystick_device"], this.horizontalLists["joystick_device"], msg.Get("Joystick"))
	this.cfgTabs[config.INPUT_TAB].SetOptionWidgets(platform.INPUT_MOUSE_MOVE, this.labels["mouse_move"], this.checkboxs["mouse_move"], msg.Get("Move hero using mouse"))
	this.cfgTabs[config.INPUT_TAB].SetOptionWidgets(platform.INPUT_MOUSE_AIM, this.labels["mouse_aim"], this.checkboxs["mouse_aim"], msg.Get("Mouse aim"))
	this.cfgTabs[config.INPUT_TAB].SetOptionWidgets(platform.INPUT_NO_MOUSE, this.labels["no_mouse"], this.checkboxs["no_mouse"], msg.Get("Do not use mouse"))
	this.cfgTabs[config.INPUT_TAB].SetOptionWidgets(platform.INPUT_MOUSE_MOVE_SWAP, this.labels["mouse_move_swap"], this.checkboxs["mouse_move_swap"], msg.Get("Swap mouse movement button"))
	this.cfgTabs[config.INPUT_TAB].SetOptionWidgets(platform.INPUT_MOUSE_MOVE_ATTACK, this.labels["mouse_move_attack"], this.checkboxs["mouse_move_attack"], msg.Get("Attack with mouse movement"))
	this.cfgTabs[config.INPUT_TAB].SetOptionWidgets(platform.INPUT_JOYSTICK_DEADZONE, this.labels["joystick_deadzone"], this.sliders["joystick_deadzone"], msg.Get("Joystick Deadzone"))
	this.cfgTabs[config.INPUT_TAB].SetOptionWidgets(platform.INPUT_TOUCH_CONTROLS, this.labels["touch_controls"], this.checkboxs["touch_controls"], msg.Get("Touch Controls"))
	this.cfgTabs[config.INPUT_TAB].SetOptionWidgets(platform.INPUT_TOUCH_SCALE, this.labels["touch_scale"], this.sliders["touch_scale"], msg.Get("Touch Gamepad Scaling"))

	// ========= keybinds tab =========
	for i, ptr := range this.keybindsLstb {
		this.cfgTabs[config.KEYBINDS_TAB].SetOptionWidgets(i, this.keybindsLb[i], ptr, inpt.GetBindingName(i))
		ptr.Append(inpt.GetBindingName(i), msg.Get(fmt.Sprintf("Primary binding: %s", inpt.GetBindingName(i))))
	}

	// ========= 默认禁用 ==========
	if !this.isGameState {
		this.cfgTabs[config.VIDEO_TAB].SetOptionEnabled(platform.VIDEO_RENDERER, false)
		this.cfgTabs[config.VIDEO_TAB].SetOptionEnabled(platform.VIDEO_HWSURFACE, false)
		this.cfgTabs[config.VIDEO_TAB].SetOptionEnabled(platform.VIDEO_VSYNC, false)
		this.cfgTabs[config.VIDEO_TAB].SetOptionEnabled(platform.VIDEO_TEXTURE_FILTER, false)
		this.cfgTabs[config.VIDEO_TAB].SetOptionEnabled(platform.VIDEO_FRAME_LIMIT, false)
		this.cfgTabs[config.INTERFACE_TAB].SetOptionEnabled(platform.INTERFACE_LANGUAGE, false)
		this.cfgTabs[config.INTERFACE_TAB].SetOptionEnabled(platform.INTERFACE_DEV_MODE, false)
		this.tabControl.SetEnabled(config.MODS_TAB, false)
		this.tabControl.SetEnabled(config.EXIT_TAB, true)
		this.enableGameStateButtons = false
	} else {
		this.cfgTabs[config.EXIT_TAB].SetOptionEnabled(config.EXIT_OPTION_CONTINUE, false)
		this.cfgTabs[config.EXIT_TAB].SetOptionEnabled(config.EXIT_OPTION_SAVE, false)
		this.cfgTabs[config.EXIT_TAB].SetOptionEnabled(config.EXIT_OPTION_EXIT, false)
		this.cfgTabs[config.EXIT_TAB].SetOptionEnabled(config.EXIT_OPTION_TIME_PLAYED, false)
		this.tabControl.SetEnabled(config.EXIT_TAB, false)
		this.enableGameStateButtons = true
	}

	for i, val := range plat.GetConfigVideo() {
		if !val {
			this.cfgTabs[config.VIDEO_TAB].SetOptionEnabled(i, false)
		}
	}

	for i, val := range plat.GetConfigAudio() {
		if !val {
			this.cfgTabs[config.AUDIO_TAB].SetOptionEnabled(i, false)
		}
	}

	for i, val := range plat.GetConfigInterface() {
		if !val {
			this.cfgTabs[config.INTERFACE_TAB].SetOptionEnabled(i, false)
		}
	}

	for i, val := range plat.GetConfigInput() {
		if !val {
			this.cfgTabs[config.INPUT_TAB].SetOptionEnabled(i, false)
		}
	}

	if !plat.GetConfigMisc()[platform.MISC_KEYBINDS] {
		for i, _ := range this.keybindsLstb {
			this.cfgTabs[config.KEYBINDS_TAB].SetOptionEnabled(i, false)
		}
	}

	if !plat.GetConfigMisc()[platform.MISC_MODS] {
		this.tabControl.SetEnabled(config.MODS_TAB, false)
	}

	// ====== ======
	for i, _ := range this.cfgTabs {
		if this.cfgTabs[i].enabledCount == 0 {
			this.tabControl.SetEnabled((uint)(i), false)
		}

		this.cfgTabs[i].scrollbox = widgetf.New("scrollbox").(common.WidgetScrollBox).Init(modules, this.scrollpane.W, this.scrollpane.H)
		this.cfgTabs[i].scrollbox.SetPosBase(this.scrollpane.X, this.scrollpane.Y, define.ALIGN_TOPLEFT)
		this.cfgTabs[i].scrollbox.SetBg(this.scrollpaneColor)
		// 调整总高
		err = this.cfgTabs[i].scrollbox.Resize(modules, this.scrollpane.W, this.cfgTabs[i].enabledCount*this.scrollpanePadding.Y)
		if err != nil {
			return err
		}

		// 设置每个组件
		for j, opt := range this.cfgTabs[i].options {
			lb := opt.label
			w := opt.widget
			enabledIndex := this.cfgTabs[i].GetEnabledIndex(j)

			// 1. 设置每个组件的位置和宽高
			if w != nil {
				yOffset := (int)(math.Max((float64)(this.scrollpanePadding.Y-w.GetPos().H), 0) / 2)
				w.SetPosBase(
					this.scrollpane.W-w.GetPos().W-this.scrollpanePadding.X,
					(enabledIndex*this.scrollpanePadding.Y)+yOffset,
					define.ALIGN_TOPLEFT,
				)
				w.SetPos1(modules, 0, 0) // 偏移为0
			}

			if lb != nil {
				lb.SetPosBase(
					this.scrollpanePadding.X,
					(enabledIndex*this.scrollpanePadding.Y)+this.scrollpanePadding.Y/2,
					define.ALIGN_TOPLEFT,
				)
				lb.SetPos1(modules, 0, 0)
				lb.SetValign(labelinfo.VALIGN_CENTER)
			}

			// 2.
			if opt.enabled {

				// 添加到滚动盒子 位置限制在滚动盒子范围里
				this.cfgTabs[i].scrollbox.AddChildWidget(w)
				this.cfgTabs[i].scrollbox.AddChildWidget(lb)
			}

			this.AddChildWidget(w, config.NO_TAB)
			this.AddChildWidget(lb, config.NO_TAB)
		}
	}

	this.AddChildWidget(this.listboxs["activemods"], config.MODS_TAB)
	this.AddChildWidget(this.labels["activemods"], config.MODS_TAB)
	this.AddChildWidget(this.listboxs["inactivemods"], config.MODS_TAB)
	this.AddChildWidget(this.labels["inactivemods"], config.MODS_TAB)
	this.AddChildWidget(this.buttons["activemods_shiftup"], config.MODS_TAB)
	this.AddChildWidget(this.buttons["activemods_shiftdown"], config.MODS_TAB)
	this.AddChildWidget(this.buttons["activemods_deactivate"], config.MODS_TAB)
	this.AddChildWidget(this.buttons["inactivemods_activate"], config.MODS_TAB)

	// TODO
	// tablist

	this.Update(modules)

	this.RefreshWidgets(modules)
	return nil
}

// 更新窗口和组件的位置和大小
func (this *Config) refreshWindowSize(modules common.Modules) {
	render := modules.Render()
	inpt := modules.Inpt()
	settings := modules.Settings()
	eset := modules.Eset()

	render.WindowResize(settings, eset) // 设备更新
	inpt.SetWindowResized(true)         // 逻辑上通知所有要更新的组件
	this.RefreshWidgets(modules)
	this.forceRefreshBackground = true
}

// 同步视频配置
func (this *Config) updateVideo(modules common.Modules) error {
	settings := modules.Settings()
	render := modules.Render()
	msg := modules.Msg()

	this.checkboxs["fullscreen"].SetChecked(settings.Get("fullscreen").(bool))
	this.checkboxs["hwsurface"].SetChecked(settings.Get("hwsurface").(bool))
	this.checkboxs["vsync"].SetChecked(settings.Get("vsync").(bool))
	this.checkboxs["texture_filter"].SetChecked(settings.Get("texture_filter").(bool))
	this.checkboxs["dpi_scaling"].SetChecked(settings.Get("dpi_scaling").(bool))
	this.checkboxs["parallax_layers"].SetChecked(settings.Get("parallax_layers").(bool))
	this.checkboxs["change_gamma"].SetChecked(settings.Get("change_gamma").(bool))

	if settings.Get("change_gamma").(bool) {
		err := render.SetGamma(settings.Get("gamma").(float32))
		if err != nil {
			return err
		}

		this.sliders["gamma"].Set(slider.GAMMA_MIN, slider.GAMMA_MAX, (int)(settings.Get("gamma").(float32)*10))
	} else {
		this.sliders["gamma"].SetEnabled(false)
		this.sliders["gamma"].Set(slider.GAMMA_MIN, slider.GAMMA_MAX, 10)
		err := render.ResetGamma()
		if err != nil {
			return err
		}
	}

	this.horizontalLists["renderer"].Clear1()
	names, descs := render.CreateRenderDeviceList(msg)

	for i, name := range names {
		this.horizontalLists["renderer"].Append(name, descs[i])
		if name == settings.Get("renderer").(string) {
			this.horizontalLists["renderer"].Select(modules, i)
		}
	}

	for i, limit := range this.frameLimits {
		if limit == settings.Get("max_fps").(int) {
			this.horizontalLists["frame_limit"].Select(modules, i)
			break
		}
	}

	if settings.Get("max_render_size").(int) == 0 {
		this.horizontalLists["max_render_size"].Select(modules, 0)
	} else {
		for i, val := range this.virtualHeights {
			if val == settings.Get("max_render_size").(int) {
				this.horizontalLists["max_render_size"].Select(modules, i+1)
				break
			}
		}
	}

	this.refreshWindowSize(modules) // 更新窗口和组件的大小和位置
	err := this.cfgTabs[config.VIDEO_TAB].scrollbox.Refresh(modules)
	if err != nil {
		return err
	}
	return nil
}

func (this *Config) updateMods(modules common.Modules) error {
	this.listboxs["activemods"].Refresh(modules)
	this.listboxs["inactivemods"].Refresh(modules)

	return nil
}

func (this *Config) Update(modules common.Modules) error {
	err := this.updateVideo(modules)
	if err != nil {
		return err
	}

	err = this.updateMods(modules)
	if err != nil {
		return err
	}

	return nil
}

// 刷新组件位置和大小
func (this *Config) RefreshWidgets(modules common.Modules) error {
	settings := modules.Settings()
	eset := modules.Eset()

	// 显示位置 = 屏幕宽 - 菜单的宽
	tmpW := settings.GetViewW() - eset.Get("resolutions", "menu_frame_width").(int)
	tmpH := settings.GetViewH() - eset.Get("resolutions", "menu_frame_height").(int)

	// 设置标签区域范围的位置
	this.tabControl.SetMainArea(modules, tmpW/2+this.tabOffset.X, tmpH/2+this.tabOffset.Y)

	this.frame.X = tmpW/2 + this.frameOffset.X
	tabHeight, err := this.tabControl.GetTabHeight() // 获取单个标签的高
	if err != nil {
		return err
	}
	this.frame.Y = tmpH/2 + tabHeight + this.frameOffset.Y

	for i, ptr := range this.childWidget {
		if this.optionTab[i] != config.NO_TAB {
			ptr.SetPos1(modules, this.frame.X, this.frame.Y)
		}
	}

	this.buttons["ok"].SetPos1(modules, 0, 0) // 偏移为0
	this.buttons["defaults"].SetPos1(modules, 0, 0)
	this.buttons["cancel"].SetPos1(modules, 0, 0)

	this.defaultsConfirm.Align(modules)

	// 设置每个滚动盒子的位置
	for i, _ := range this.cfgTabs {
		this.cfgTabs[i].scrollbox.SetPos1(modules, this.frame.X, this.frame.Y)
	}

	this.inputConfirm.Align(modules)

	return nil
}

func (this *Config) logicDefaults() {
	// TODO
}

func (this *Config) logicInput() {
	// TODO
}

//  游戏内退出逻辑
func (this *Config) logicExit(modules common.Modules) error {
	inpt := modules.Inpt()

	err := this.cfgTabs[config.EXIT_TAB].scrollbox.Logic(modules)
	if err != nil {
		return err
	}

	// 父组件的坐标转化到其子组件的坐标
	mouse, ok := this.cfgTabs[config.EXIT_TAB].scrollbox.InputAssist(inpt.GetMouse())

	// 父组件范围内
	if ok {
		if this.cfgTabs[config.EXIT_TAB].options[config.EXIT_OPTION_CONTINUE].enabled && this.buttons["pause_continue"].CheckClickAt(modules, mouse.X, mouse.Y) {
			this.clickedPauseContinue = true
		} else if this.cfgTabs[config.EXIT_TAB].options[config.EXIT_OPTION_SAVE].enabled && this.buttons["pause_save"].CheckClickAt(modules, mouse.X, mouse.Y) {
			this.clickedPauseSave = true
		} else if this.cfgTabs[config.EXIT_TAB].options[config.EXIT_OPTION_EXIT].enabled && this.buttons["pause_exit"].CheckClickAt(modules, mouse.X, mouse.Y) {
			this.clickedPauseExit = true
		}
	}

	return nil
}

func (this *Config) logicVideo(modules common.Modules) error {
	inpt := modules.Inpt()

	err := this.cfgTabs[config.VIDEO_TAB].scrollbox.Logic(modules)
	if err != nil {
		return err
	}

	// 父组件的坐标转化到其子组件的坐标
	mouse, ok := this.cfgTabs[config.VIDEO_TAB].scrollbox.InputAssist(inpt.GetMouse())

	// 父组件范围内
	if ok {
		if this.horizontalLists["renderer"].CheckClickAt(modules, mouse.X, mouse.Y) {
			this.newRenderDevice = this.horizontalLists["renderer"].GetValue()
		}
	}

	//TODO

	return nil
}

func (this *Config) logicMods(modules common.Modules) error {
	if this.listboxs["activemods"].CheckClick(modules) {

	} else if this.listboxs["inactivemods"].CheckClick(modules) {

	} else if this.buttons["activemods_shiftup"].CheckClick(modules) {
		this.listboxs["activemods"].ShiftUp(modules)

	} else if this.buttons["activemods_shiftdown"].CheckClick(modules) {
		this.listboxs["activemods"].ShiftDown(modules)

	} else if this.buttons["activemods_deactivate"].CheckClick(modules) {

		// disable

		for i := 0; i < this.listboxs["activemods"].GetSize(); i++ {
			if this.listboxs["activemods"].IsSelected(i) {
				if val, ok := this.listboxs["activemods"].GetValue(i); ok && val != modmanager.FALLBACK_MOD {
					tooltip, _ := this.listboxs["activemods"].GetTooltip(i)
					this.listboxs["inactivemods"].Append(modules, val, tooltip)
					this.listboxs["activemods"].Remove(modules, i)
				}
				i--
			}
		}

		this.listboxs["inactivemods"].Sort()
	} else if this.buttons["inactivemods_activate"].CheckClick(modules) {

		// enable
		for i := 0; i < this.listboxs["inactivemods"].GetSize(); i++ {
			if this.listboxs["inactivemods"].IsSelected(i) {
				if val, ok := this.listboxs["inactivemods"].GetValue(i); ok {
					tooltip, _ := this.listboxs["inactivemods"].GetTooltip(i)
					this.listboxs["activemods"].Append(modules, val, tooltip)
					this.listboxs["inactivemods"].Remove(modules, i)
				}
				i--
			}

		}
	}

	return nil
}

func (this *Config) logicMain(modules common.Modules) (bool, error) {
	for i, ptr := range this.childWidget {
		if ptr.GetInFocus() && this.optionTab[i] != config.NO_TAB {
			this.tabControl.SetActiveTab((uint)(this.optionTab[i]))
			break
		}
	}

	err := this.tabControl.Logic(modules)
	if err != nil {
		return false, err
	}

	if this.enableGameStateButtons {
		if this.buttons["ok"].CheckClick(modules) {
			this.clickedAccept = true
			return false, nil
		}

		if this.buttons["defaults"].CheckClick(modules) {
			//TODO
			return true, nil
		}

		if this.buttons["cancel"].CheckClick(modules) {
			this.clickedCancel = true
			return false, nil
		}

	}

	return true, nil
}

func (this *Config) Logic(modules common.Modules) error {
	inpt := modules.Inpt()

	// 窗口变化不断修正大小和位置
	if inpt.GetWindowResized() {
		this.RefreshWidgets(modules)
	}

	// 确认按钮
	if this.defaultsConfirm.GetVisible() {
		this.logicDefaults()
		return nil
	} else if this.inputConfirm.GetVisible() {
		this.logicInput()
		return nil
	} else {
		// 主逻辑
		ret, err := this.logicMain(modules)
		if err != nil {
			return err
		}

		if !ret {
			return nil
		}
	}

	this.activeTab = (int)(this.tabControl.GetActiveTab())

	switch this.activeTab {
	case config.EXIT_TAB:
		if this.hero != nil {
			this.labels["pause_time_text"].SetText(utils.GetTimeString((int)(this.hero.GetTimePlayed())))
		}

		err := this.logicExit(modules)
		if err != nil {
			return err
		}

	case config.VIDEO_TAB:
		err := this.logicVideo(modules)
		if err != nil {
			return err
		}

	case config.MODS_TAB:
		err := this.logicMods(modules)
		if err != nil {
			return err
		}
	}

	return nil
}

// 绘制菜单内容
func (this *Config) RenderTabContents(modules common.Modules) error {
	if this.activeTab == config.EXIT_TAB {
	} else if this.activeTab < config.KEYBINDS_TAB {

		// 滚动盒子
		err := this.cfgTabs[this.activeTab].scrollbox.Render(modules)
		if err != nil {
			return err
		}
	}

	// 子组件 一些按钮
	for i, ptr := range this.childWidget {
		if this.optionTab[i] == this.activeTab && this.optionTab[i] != config.NO_TAB {
			err := ptr.Render(modules)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// 绘制弹窗
func (this *Config) RenderDialogs(modules common.Modules) error {
	if this.defaultsConfirm.GetVisible() {
		err := this.defaultsConfirm.Render(modules)
		if err != nil {
			return err
		}
	}

	if this.inputConfirm.GetVisible() {
		err := this.inputConfirm.Render(modules)
		if err != nil {
			return err
		}
	}

	//TODO

	return nil
}

func (this *Config) Render(modules common.Modules) error {
	eset := modules.Eset()
	settings := modules.Settings()
	render := modules.Render()

	tabHeight, err := this.tabControl.GetTabHeight()
	if err != nil {
		return err
	}

	pos := rect.Construct()
	pos.X = (settings.GetViewW() - eset.Get("resolutions", "menu_frame_width").(int)) / 2
	pos.Y = (settings.GetViewH()-eset.Get("resolutions", "menu_frame_height").(int))/2 + tabHeight - tabHeight/16

	if this.background != nil {
		this.background.SetDestFromRect(pos)
		render.Render(this.background)
	}

	err = this.tabControl.Render(modules)
	if err != nil {
		return err
	}

	if this.enableGameStateButtons {
		err := this.buttons["ok"].Render(modules)
		if err != nil {
			return err
		}

		err = this.buttons["cancel"].Render(modules)
		if err != nil {
			return err
		}

		err = this.buttons["defaults"].Render(modules)
		if err != nil {
			return err
		}
	}

	// 每个标签的滚动盒子内容
	err = this.RenderTabContents(modules)
	if err != nil {
		return err
	}

	// 弹窗
	err = this.RenderDialogs(modules)
	if err != nil {
		return err
	}

	return nil
}

func (this *Config) SetForceRefreshBackground(val bool) {
	this.forceRefreshBackground = val
}

func (this *Config) GetForceRefreshBackground() bool {
	return this.forceRefreshBackground
}

func (this *Config) SetClickedAccept(val bool) {
	this.clickedAccept = val
}

func (this *Config) GetClickedAccept() bool {
	return this.clickedAccept
}

func (this *Config) SetClickedCancel(val bool) {
	this.clickedCancel = val
}

func (this *Config) GetClickedCancel() bool {
	return this.clickedCancel
}

func (this *Config) GetRenderDevice() string {
	return this.horizontalLists["renderer"].GetValue()
}

// 刷新mods
func (this *Config) RefreshMods(modules common.Modules) (bool, error) {
	mods := modules.Mods()

	tempList := mods.GetModList()

	mods.ClearModList() // 清理
	newMod, err := mods.LoadMod(modmanager.FALLBACK_MOD)
	if err != nil {
		return false, err
	}

	mods.AddToModList(newMod) // 添加默认mod
	for i := 0; i < this.listboxs["activemods"].GetSize(); i++ {
		if activeModName, ok := this.listboxs["activemods"].GetValue(i); ok {
			activeMod, err := mods.LoadMod(activeModName)
			if err != nil {
				return false, err
			}
			mods.AddToModList(activeMod) // 添加选择激活的mod
		}
	}

	err = mods.ApplyDepends() // 补充依赖
	if err != nil {
		return false, err
	}

	newList := mods.GetModList()
	needSave := false

	// 比较是否需要保存
	if len(newList) != len(tempList) {
		needSave = true
	} else {
		for i, val := range newList {
			if val.GetName() != tempList[i].GetName() {
				needSave = true
				break
			}
		}
	}

	if needSave {
		fmt.Println("save mods")
		// save mods to disk
		err = mods.SaveMods(modules)
		if err != nil {
			return false, err
		}

		return true, nil
	} else {
		return false, nil
	}
}

// 重置每个标签的操作
func (this *Config) ResetSelectedTab(modules common.Modules) error {
	err := this.Update(modules) // 同步配置到按钮显示
	if err != nil {
		return err
	}

	this.tabControl.SetActiveTab(0) // 回首页

	for i, _ := range this.cfgTabs {
		err := this.cfgTabs[i].scrollbox.ScrollToTop(modules) // 每个标签页回顶部
		if err != nil {
			return err
		}
	}

	this.inputConfirm.SetVisible(false)
	this.inputConfirmTimer.Reset(timer.END)
	this.keybindTipTimer.Reset(timer.END)

	return nil
}
