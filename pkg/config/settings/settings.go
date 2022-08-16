package settings

import (
	"math"
	"monster/pkg/common"
	"monster/pkg/common/define/enginesettings"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/utils/parsing"

	"monster/pkg/filesystem/logfile"
	"reflect"
	"strconv"
)

type ConfigEntry struct {
	name       string
	defaultVal string
	storage    interface{}
	comment    string
}

type Settings struct {
	settings map[string]*ConfigEntry
	config   map[int]*ConfigEntry

	pathConf       string // user configure
	pathUser       string // user data
	pathData       string // game data
	customPathData string

	// 命令行设置
	loadSlot string

	safeVideo        bool
	renderDeviceName string

	viewW       int
	viewH       int
	viewWHalf   int
	viewHHalf   int
	viewScaling float32

	mouseScaled   bool
	showHud       bool
	encounterDist float32 // 相遇距离
	softReset     bool
}

func New() *Settings {
	s := &Settings{
		safeVideo:   false,
		settings:    map[string]*ConfigEntry{},
		config:      map[int]*ConfigEntry{},
		viewW:       0,
		viewH:       0,
		viewWHalf:   0,
		viewHHalf:   0,
		viewScaling: 1.0,
		mouseScaled: true,
		showHud:     true,
	}

	s.setConfigDefault(0, &ConfigEntry{"move_type_dimissed", "0", false, "One time flag for initial movement type dialog | 0 = show dialog, 1 = no dialog"})
	s.setConfigDefault(1, &ConfigEntry{"fullscreen", "0", false, "Fullscreen mode | 0 = disable, 1 = enable"})
	s.setConfigDefault(2, &ConfigEntry{"resolution_w", "640", int(0), "Window size"})
	s.setConfigDefault(3, &ConfigEntry{"resolution_h", "480", int(0), ""})
	s.setConfigDefault(4, &ConfigEntry{"music_volume", "96", int(0), "Music and sound volume | 0 = silent, 128 = maximum"})
	s.setConfigDefault(5, &ConfigEntry{"sound_volume", "128", int(0), ""})
	s.setConfigDefault(6, &ConfigEntry{"combat_text", "1", false, "Display floating damage text | 0 = disable, 1 = enable"})
	s.setConfigDefault(7, &ConfigEntry{"mouse_move", "0", false, "Use mouse to move | 0 = disable, 1 = enable"})
	s.setConfigDefault(8, &ConfigEntry{"hwsurface", "1", false, "Hardware surfaces & V-sync. Try disabling for performance. | 0 = disable, 1 = enable"})
	s.setConfigDefault(9, &ConfigEntry{"vsync", "1", false, ""})
	s.setConfigDefault(10, &ConfigEntry{"texture_filter", "1", false, "Texture filter quality | 0 = nearest neighbor (worst), 1 = linear (best)"})
	s.setConfigDefault(11, &ConfigEntry{"dpi_scaling", "0", false, "DPI-based render scaling | 0 = disable, 1 = enable"})
	s.setConfigDefault(12, &ConfigEntry{"parallax_layers", "1", false, "Rendering of parallax map layers | 0 = disable, 1 = enable"})
	s.setConfigDefault(13, &ConfigEntry{"max_fps", "60", int(0), "Maximum frames per second | 60 = default"})
	s.setConfigDefault(14, &ConfigEntry{"renderer", "sdl_hardware", "", "Default render device. | sdl_hardware = default, Try sdl for compatibility"})
	s.setConfigDefault(15, &ConfigEntry{"enable_joystick", "0", false, "Joystick settings."})
	s.setConfigDefault(16, &ConfigEntry{"joystick_device", "-1", int(0), ""})
	s.setConfigDefault(17, &ConfigEntry{"joystick_deadzone", "100", int(0), ""})
	s.setConfigDefault(18, &ConfigEntry{"language", "en", "", "2-letter language code."})
	s.setConfigDefault(19, &ConfigEntry{"change_gamma", "0", false, "Allow changing screen gamma (experimental) | 0 = disable, 1 = enable"})
	s.setConfigDefault(20, &ConfigEntry{"gamma", "1.0", float32(0), "Screen gamma. Requires change_gamma=1 | 0.5 = darkest, 2.0 = lightest"})
	s.setConfigDefault(21, &ConfigEntry{"mouse_aim", "1", false, "Use mouse to aim | 0 = disable, 1 = enable"})
	s.setConfigDefault(22, &ConfigEntry{"no_mouse", "0", false, "Make using mouse secondary, give full control to keyboard | 0 = disable, 1 = enable"})
	s.setConfigDefault(23, &ConfigEntry{"show_fps", "0", false, "Show frames per second | 0 = disable, 1 = enable"})
	s.setConfigDefault(24, &ConfigEntry{"colorblind", "0", false, "Enable colorblind help text | 0 = disable, 1 = enable"})
	s.setConfigDefault(25, &ConfigEntry{"hardware_cursor", "0", false, "Use the system mouse cursor | 0 = disable, 1 = enable"})
	s.setConfigDefault(26, &ConfigEntry{"dev_mode", "0", false, "Developer mode | 0 = disable, 1 = enable"})
	s.setConfigDefault(27, &ConfigEntry{"dev_hud", "1", false, "Show additional information on-screen when dev_mode=1 | 0 = disable, 1 = enable"})
	s.setConfigDefault(28, &ConfigEntry{"loot_tooltips", "0", int(0), "Loot tooltip mode | 0 = normal, 1 = show all, 2 = hide all"})
	s.setConfigDefault(29, &ConfigEntry{"statbar_labels", "0", false, "Always show labels on HP/MP/XP bars | 0 = disable, 1 = enable"})
	s.setConfigDefault(30, &ConfigEntry{"statbar_autohide", "1", false, "Allow the HP/MP/XP bars to auto-hide on inactivity | 0 = disable, 1 = enable"})
	s.setConfigDefault(31, &ConfigEntry{"auto_equip", "1", false, "Automatically equip items | 0 = enable, 1 = enable"})
	s.setConfigDefault(32, &ConfigEntry{"subtitles", "0", false, "Subtitles | 0 = disable, 1 = enable"})
	s.setConfigDefault(33, &ConfigEntry{"minimap_mode", "0", int(0), "Mini-map display mode | 0 = normal, 1 = 2x zoom, 2 = hidden"})
	s.setConfigDefault(34, &ConfigEntry{"mouse_move_swap", "0", false, "Use 'Main2' as the movement action when mouse_move=1 | 0 = disable, 1 = enable"})
	s.setConfigDefault(35, &ConfigEntry{"mouse_move_attack", "1", false, "Allow attacking with the mouse movement button if an enemy is targeted and in range | 0 = disable, 1 = enable"})
	s.setConfigDefault(36, &ConfigEntry{"entity_markers", "1", false, "Shows a marker above entities that are hidden behind tall tiles | 0 = disable, 1 = enable"})
	s.setConfigDefault(37, &ConfigEntry{"prev_save_slot", "-1", int(0), "Index of the last used save slot"})
	s.setConfigDefault(38, &ConfigEntry{"low_hp_warning_type", "1", int(0), "Low health warning type settings | 0 = disable, 1 = all, 2 = message & cursor, 3 = message & sound, 4 = cursor & sound , 5 = message, 6 = cursor, 7 = sound"})
	s.setConfigDefault(39, &ConfigEntry{"low_hp_threshold", "20", int(0), "Low HP warning threshold percentage"})
	s.setConfigDefault(40, &ConfigEntry{"item_compare_tips", "1", false, "Show comparison tooltips for equipped items of the same type | 0 = disable, 1 = enable"})
	s.setConfigDefault(41, &ConfigEntry{"max_render_size", "0", int(0), "Overrides the maximum height (in pixels) of the internal render surface | 0 = ignore this setting"})
	s.setConfigDefault(42, &ConfigEntry{"touch_controls", "0", false, "Enables touch screen controls | 0 = disable, 1 = enable"})
	s.setConfigDefault(43, &ConfigEntry{"touch_scale", "1.0", float32(0), "Factor used to scale the touch controls | 1.0 = 100 percent scale"})

	return s
}

// 加载 conf_path/settings
func (this *Settings) LoadSettings(mods common.ModManager) error {

	for _, config := range this.config {
		parsing.TryParseValue(config.defaultVal, &(config.storage))
	}

	foundSettings := false
	infile := fileparser.Construct()
	if err := infile.Open(this.pathConf+"settings.txt", false, mods); err != nil {
		return err
	}

	defer infile.Close()
	foundSettings = true

	if !foundSettings {
		this.saveSettings()
	} else {
		for infile.Next(mods) {
			if entry, ok := this.settings[infile.Key()]; ok {
				parsing.TryParseValue(infile.Val(), &(entry.storage))
			} else {
				logfile.LogError("Settings: '%s' is not a valid configuration key.", infile.Key())
				return common.Err_bad_key_in_settings
			}
		}
	}

	this.loadMobileDefault()

	if this.safeVideo {
		this.renderDeviceName = "sdl"
	}

	return nil
}

func (this *Settings) saveSettings() {
	//TODO
}

func (this *Settings) loadMobileDefault() {
	//TODO
}

func (this *Settings) LogSettings() {
	i := 0
	for {
		conf, ok := this.config[i]
		if !ok {
			break
		}

		logfile.LogInfo("Settings: %s=%s", conf.name, this.configValueToString(conf.storage))
		i++
	}

}

func (this *Settings) setConfigDefault(index int, c *ConfigEntry) {
	if c == nil {
		return
	}
	if index > len(this.config) {
		logfile.LogError("Settings: Can't set default config value; %d is not a valid index.", index)
		return
	}

	this.config[index] = c
	this.settings[c.name] = c
}

func (this *Settings) configValueToString(output interface{}) string {
	if output == nil {
		return ""
	}

	vv := reflect.ValueOf(output)
	switch vv.Kind() {
	case reflect.Bool:
		rawVal := vv.Bool()
		if rawVal == true {
			return "true"
		} else {
			return "false"
		}
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int:
		return strconv.FormatInt(vv.Int(), 10)
	case reflect.Uint:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint8:
		return strconv.FormatUint(vv.Uint(), 10)
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		return strconv.FormatFloat(vv.Float(), 'f', 5, 64)
	case reflect.String:
		return vv.String()
	}

	return ""
}

func (this *Settings) SetPathUser(s string) {
	this.pathUser = s
}

func (this *Settings) GetPathUser() string {
	return this.pathUser
}

func (this *Settings) GetPathData() string {
	return this.pathData
}

func (this *Settings) GetPathConf() string {
	return this.pathConf
}

func (this *Settings) SetPathConf(s string) {
	this.pathConf = s
}

func (this *Settings) SetPathData(s string) {
	this.pathData = s
}

func (this *Settings) GetCustomPathData() string {
	return this.customPathData
}

func (this *Settings) SetSafeVideo(val bool) {
	this.safeVideo = true
	return
}

func (this *Settings) Get(name string) interface{} {
	conf, ok := this.settings[name]
	if ok {
		return conf.storage
	}

	return nil
}

func (this *Settings) LOGIC_FPS() float32 {
	return 60.0
}

func (this *Settings) SetViewW(val int) {
	this.viewW = val
}

func (this *Settings) GetViewW() int {
	return this.viewW
}

func (this *Settings) SetViewH(val int) {
	this.viewH = val
}

func (this *Settings) GetViewH() int {
	return this.viewH
}

func (this *Settings) SetViewHHalf(val int) {
	this.viewHHalf = val
}

func (this *Settings) GetViewHHalf() int {
	return this.viewHHalf
}

func (this *Settings) SetViewWHalf(val int) {
	this.viewWHalf = val
}

func (this *Settings) GetViewWHalf() int {
	return this.viewWHalf
}

func (this *Settings) SetViewScaling(val float32) {
	this.viewScaling = val
}

func (this *Settings) GetViewScaling() float32 {
	return this.viewScaling
}

func (this *Settings) GetShowHud() bool {
	return this.showHud
}

func (this *Settings) Set(key string, val interface{}) {
	this.settings[key].storage = val
}

func (this *Settings) UpdateScreenVars(eset common.EngineSettings) {
	tileSize := eset.Get("tileset", "tile_size").([]int)
	tileW := tileSize[0]
	tileH := tileSize[1]
	tileHHalf := eset.Get("tileset", "tile_h_half").(int)
	if tileW > 0 && tileH > 0 {
		if eset.Get("tileset", "orientation").(int) == enginesettings.TILESET_ISOMETRIC {
			// 等距

			// ( (view宽/瓷砖宽)^2 + ((view高/瓷砖半高)^2) ) 的根号
			this.encounterDist = (float32)(math.Sqrt(math.Pow(float64(this.viewW/tileW), 2)+math.Pow(float64(this.viewH/tileHHalf), 2))) / 2
		} else if eset.Get("tileset", "orientation").(int) == enginesettings.TILESET_ORTHOGONAL {
			// 正交

			// ( (view宽/瓷砖宽)^2 + ((view高/瓷砖高)^2) ) 的根号
			this.encounterDist = (float32)(math.Sqrt(math.Pow(float64(this.viewW/tileW), 2)+math.Pow(float64(this.viewH/tileH), 2))) / 2
		}

	}

}

func (this *Settings) GetMouseScaled() bool {
	return this.mouseScaled
}

func (this *Settings) GetLoadSlot() string {
	return this.loadSlot
}

func (this *Settings) SetSoftReset(val bool) {
	this.softReset = val
}

func (this *Settings) GetSoftReset() bool {
	return this.softReset
}

func (this *Settings) GetEncounterDist() float32 {
	return this.encounterDist
}
