package enginesettings

import (
	"fmt"
	"math"
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define"
	"monster/pkg/common/define/enginesettings"
	"monster/pkg/common/define/fontengine"
	"monster/pkg/common/labelinfo"
	"monster/pkg/common/point"
	"monster/pkg/config/menu/actionbar"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/utils"
	"monster/pkg/utils/parsing"
	"sort"
)

type EngineSettings struct {
	Misc
	Resolutions
	Gameplay
	Combat
	Elements
	EquipFlags
	PrimaryStats
	HeroClasses
	DamageTypes
	DeathPenalty
	Tooltips
	Loot
	Tileset
	Widgets
	XPTable
}

type ConfigEntry struct {
	name    string
	storage interface{}
}

func New() *EngineSettings {
	e := &EngineSettings{}
	e.HeroClasses.eset = e
	e.Loot.eset = e
	return e
}

func (this *EngineSettings) Load(settings common.Settings, mods common.ModManager, msg common.MessageEngine, font common.FontEngine) error {
	if err := this.Misc.load(settings, mods); err != nil {
		return err
	}

	if err := this.Resolutions.load(settings, mods); err != nil {
		return err
	}

	if err := this.Gameplay.load(settings, mods); err != nil {
		return err
	}

	if err := this.Combat.load(settings, mods); err != nil {
		return err
	}

	if err := this.Elements.load(settings, mods); err != nil {
		return err
	}

	if err := this.EquipFlags.load(settings, mods); err != nil {
		return err
	}

	if err := this.PrimaryStats.load(settings, mods); err != nil {
		return err
	}

	if err := this.HeroClasses.load(settings, mods, msg); err != nil {
		return err
	}

	if err := this.DamageTypes.load(settings, mods, msg); err != nil {
		return err
	}

	if err := this.DeathPenalty.load(settings, mods); err != nil {
		return err
	}

	if err := this.Tooltips.load(settings, mods); err != nil {
		return err
	}

	if err := this.Loot.load(settings, mods, msg); err != nil {
		return err
	}

	if err := this.Tileset.load(settings, mods); err != nil {
		return err
	}

	if err := this.Widgets.load(settings, mods, font); err != nil {
		return err
	}

	if err := this.XPTable.load(settings, mods); err != nil {
		return err
	}
	return nil
}

func (this *EngineSettings) Get(key string, subKey string) interface{} {
	switch key {
	case "resolutions":
		return this.Resolutions.get(subKey)
	case "tileset":
		return this.Tileset.get(subKey)
	case "misc":
		return this.Misc.get(subKey)
	case "tooltips":
		return this.Tooltips.get(subKey)
	case "gameplay":
		return this.Gameplay.get(subKey)
	case "widgets":
		return this.Widgets.get(subKey)
	case "damage_types":
		return this.DamageTypes.get(subKey)
	case "elements":
		return this.Elements.get(subKey)
	case "primary_stats":
		return this.PrimaryStats.get(subKey)
	case "death_penalty":
		return this.DeathPenalty.get(subKey)
	case "hero_classes":
		return this.HeroClasses.get(subKey)
	case "loot":
		return this.Loot.get(subKey)
	}

	return nil
}

type Misc struct {
	misc map[string]*ConfigEntry
}

func (this *Misc) load(settings common.Settings, mods common.ModManager) error {
	s := settings

	if this.misc == nil {
		this.misc = map[string]*ConfigEntry{}
	}

	maxFps := s.Get("max_fps").(int)
	// reset to defaults
	this.misc["save_hpmp"] = &ConfigEntry{"save_hpmp", false}
	this.misc["corpse_timeout"] = &ConfigEntry{"corpse_timeout", 60 * maxFps}
	this.misc["sell_without_vendor"] = &ConfigEntry{"sell_without_vendor", true}
	this.misc["aim_assist"] = &ConfigEntry{"aim_assist", 0}
	this.misc["window_title"] = &ConfigEntry{"window_title", "Flare"}
	this.misc["save_prefix"] = &ConfigEntry{"save_prefix", ""}
	this.misc["sound_falloff"] = &ConfigEntry{"sound_falloff", (int)(15)}
	this.misc["party_exp_percentage"] = &ConfigEntry{"party_exp_percentage", (int)(100)}
	this.misc["enable_ally_collision"] = &ConfigEntry{"enable_ally_collision", true}
	this.misc["enable_ally_collision_ai"] = &ConfigEntry{"enable_ally_collision_ai", true}
	this.misc["currency_id"] = &ConfigEntry{"currency_id", 1}
	this.misc["interact_range"] = &ConfigEntry{"interact_range", 3}
	this.misc["menus_pause"] = &ConfigEntry{"menus_pause", false}
	this.misc["save_onload"] = &ConfigEntry{"save_onload", true}
	this.misc["save_onexit"] = &ConfigEntry{"save_onexit", true}
	this.misc["save_pos_onexit"] = &ConfigEntry{"save_pos_onexit", false}
	this.misc["save_oncutscene"] = &ConfigEntry{"save_oncutscene", true}
	this.misc["save_onstash"] = &ConfigEntry{"save_onstash", enginesettings.SAVE_ONSTASH_ALL}
	this.misc["save_anywhere"] = &ConfigEntry{"save_anywhere", false}
	this.misc["camera_speed"] = &ConfigEntry{"camera_speed", (float32)(10 * ((float32)(maxFps) / s.LOGIC_FPS()))}
	this.misc["save_buyback"] = &ConfigEntry{"save_buyback", true}
	this.misc["keep_buyback_on_map_change"] = &ConfigEntry{"keep_buyback_on_map_change", true}
	this.misc["sfx_unable_to_cast"] = &ConfigEntry{"sfx_unable_to_cast", ""}
	this.misc["combat_aborts_npc_interact"] = &ConfigEntry{"combat_aborts_npc_interact", true}

	infile := fileparser.Construct()

	if err := infile.Open("engine/misc.txt", true, mods); err != nil && utils.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	defer infile.Close()

	for infile.Next(mods) {
		if entry, ok := this.misc[infile.Key()]; ok {
			switch infile.Key() {
			case "corpse_timeout":
				entry.storage = parsing.ToDuration(infile.Val(), maxFps)
			case "save_onstash":
				if infile.Val() == "private" {
					entry.storage = enginesettings.SAVE_ONSTASH_PRIVATE
				} else if infile.Val() == "shared" {
					entry.storage = enginesettings.SAVE_ONSTASH_SHARED
				} else {
					if parsing.ToBool(infile.Val()) {
						entry.storage = enginesettings.SAVE_ONSTASH_ALL
					} else {
						entry.storage = enginesettings.SAVE_ONSTASH_NONE
					}
				}
			default:
				parsing.TryParseValue(infile.Val(), &(entry.storage))

				if infile.Key() == "currency_id" {
					if entry.storage.(int) < 0 {
						entry.storage = 1
					}
				}

				if infile.Key() == "camera_speed" {
					if entry.storage.(float32) < 0 {
						entry.storage = (float32)(1)
					}
				}
			}
		} else {
			fmt.Printf("EngineSettings: '%s' is not a valid key.\n", infile.Key())
			return common.Err_bad_key_in_enginesettings
		}
	}

	if this.misc["save_prefix"].storage.(string) == "" {
		fmt.Println("EngineSettings: save_prefix not found in engine/misc.txt, setting to 'default'. This may cause save file conflicts between games that have no save_prefix.")
		this.misc["save_prefix"].storage = "default"
	}

	if this.misc["save_buyback"].storage.(bool) && !this.misc["keep_buyback_on_map_change"].storage.(bool) {
		fmt.Println("EngineSettings: Warning, save_buyback=true is ignored when keep_buyback_on_map_change=false.")
		this.misc["save_buyback"].storage = false
	}

	return nil
}

func (this *Misc) get(key string) interface{} {
	if this.misc == nil {
		return nil
	}

	if val, ok := this.misc[key]; ok {
		return val.storage
	}

	return nil
}

type Resolutions struct {
	resolutions map[string]*ConfigEntry
}

func (this *Resolutions) load(settings common.Settings, mods common.ModManager) error {
	s := settings

	if this.resolutions == nil {
		this.resolutions = map[string]*ConfigEntry{}
	}

	// reset to defaults
	this.resolutions["menu_frame_width"] = &ConfigEntry{"menu_frame_width", int(0)}
	this.resolutions["menu_frame_height"] = &ConfigEntry{"menu_frame_height", int(0)}
	this.resolutions["icon_size"] = &ConfigEntry{"icon_size", int(0)}
	this.resolutions["required_width"] = &ConfigEntry{"required_width", int(640)}   // 最小宽
	this.resolutions["required_height"] = &ConfigEntry{"required_height", int(480)} // 最小高
	this.resolutions["virtual_height"] = &ConfigEntry{"virtual_height", []int{}}    // 可选高
	this.resolutions["virtual_dpi"] = &ConfigEntry{"virtual_dpi", float32(0)}
	this.resolutions["ignore_texture_filter"] = &ConfigEntry{"ignore_texture_filter", false}

	infile := fileparser.Construct()

	if err := infile.Open("engine/resolutions.txt", true, mods); err != nil && utils.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	defer infile.Close()

	for infile.Next(mods) {
		if entry, ok := this.resolutions[infile.Key()]; ok {
			switch infile.Key() {
			case "virtual_height":
				strVal := infile.Val()
				v_height, strVal := parsing.PopFirstString(strVal, "")
				tmp := []int{}
				for {
					if v_height == "" {
						break
					}

					test_v_height := parsing.ToInt(v_height, 0)
					if test_v_height <= 0 {
						fmt.Println("EngineSettings: virtual_height must be greater than zero.")
						return common.Err_bad_val_in_enginesettings
					} else {
						tmp = append(tmp, test_v_height)
					}

					v_height, strVal = parsing.PopFirstString(strVal, "")
				}

				if len(tmp) != 0 {
					sort.Slice(tmp, func(i, j int) bool { return tmp[i] < tmp[j] }) // 递增
					s.SetViewH(tmp[len(tmp)-1])
					s.SetViewHHalf(tmp[len(tmp)-1] / 2)
				}

				entry.storage = tmp

			default:
				parsing.TryParseValue(infile.Val(), &(entry.storage))
			}
		} else {
			fmt.Printf("EngineSettings: '%s' is not a valid key.\n", infile.Key())
			return common.Err_bad_key_in_enginesettings
		}
	}

	// prevent the window from being too small
	required_width := this.resolutions["required_width"].storage.(int)
	if s.Get("resolution_w").(int) < required_width {
		s.Set("resolution_w", required_width)
	}

	required_height := this.resolutions["required_height"].storage.(int)
	if s.Get("resolution_h").(int) < required_height {
		s.Set("resolution_h", required_height)
	}

	// icon size can not be zero, so we set a default of 32x32, which is fantasycore's icon size
	if this.resolutions["icon_size"].storage.(int) == 0 {
		fmt.Println("EngineSettings: icon_size is undefined. Setting it to 32.")
		this.resolutions["icon_size"].storage = 32
	}

	return nil
}

func (this *Resolutions) get(key string) interface{} {
	if val, ok := this.resolutions[key]; ok {
		return val.storage
	}

	return nil
}

type Gameplay struct {
	gameplay map[string]*ConfigEntry
}

func (this *Gameplay) get(key string) interface{} {
	if val, ok := this.gameplay[key]; ok {
		return val.storage
	}

	return nil
}

func (this *Gameplay) load(settings common.Settings, mods common.ModManager) error {
	if this.gameplay == nil {
		this.gameplay = map[string]*ConfigEntry{}
	}

	this.gameplay["enable_playgame"] = &ConfigEntry{"enable_playgame", false}

	infile := fileparser.New()
	if err := infile.Open("engine/gameplay.txt", true, mods); err != nil && utils.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	defer infile.Close()

	for infile.Next(mods) {
		if entry, ok := this.gameplay[infile.Key()]; ok {
			parsing.TryParseValue(infile.Val(), &(entry.storage))
		} else {
			fmt.Printf("EngineSettings: '%s' is not a valid key.\n", infile.Key())
			return common.Err_bad_key_in_enginesettings
		}
	}

	return nil
}

type Combat struct {
	combat map[string]*ConfigEntry
}

func (this *Combat) load(settings common.Settings, mods common.ModManager) error {
	if this.combat == nil {
		this.combat = map[string]*ConfigEntry{}
	}

	this.combat["absorb_percent"] = &ConfigEntry{"absorb_percent", []int{0, 100}}
	this.combat["resist_percent"] = &ConfigEntry{"resist_percent", []int{0, 100}}
	this.combat["block_percent"] = &ConfigEntry{"block_percent", []int{0, 100}}
	this.combat["avoidance_percent"] = &ConfigEntry{"avoidance_percent", []int{0, 100}}
	this.combat["miss_damage_percent"] = &ConfigEntry{"miss_damage_percent", []int{0, 0}}
	this.combat["crit_damage_percent"] = &ConfigEntry{"crit_damage_percent", []int{200, 200}}
	this.combat["overhit_damage_percent"] = &ConfigEntry{"overhit_damage_percent", []int{100, 100}}

	infile := fileparser.Construct()
	if err := infile.Open("engine/combat.txt", true, mods); err != nil && utils.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	defer infile.Close()

	for infile.Next(mods) {
		if entry, ok := this.combat[infile.Key()]; ok {
			strVal := infile.Val()
			firstVal := 0
			secondVal := 0
			firstVal, strVal = parsing.PopFirstInt(strVal, "")
			secondVal, strVal = parsing.PopFirstInt(strVal, "")
			entry.storage = []int{(int)(math.Min(float64(firstVal), float64(secondVal))), (int)(math.Max(float64(firstVal), float64(secondVal)))}

		} else {
			fmt.Printf("EngineSettings: '%s' is not a valid key.\n", infile.Key())
			return common.Err_bad_key_in_enginesettings
		}

	}
	return nil
}

type Element struct {
	name string
	id   string
}

func (this *Element) GetName() string {
	return this.name
}

func (this *Element) GetId() string {
	return this.id
}

type Elements struct {
	list []Element
}

func (this *Elements) get(key string) interface{} {
	if key == "list" {
		tmpList := make([]common.Element, len(this.list))

		for i, val := range this.list {
			tmp := val
			tmpList[i] = &tmp
		}

		return tmpList
	}

	return nil
}

func (this *Elements) load(settings common.Settings, mods common.ModManager) error {
	this.list = this.list[:0]

	infile := fileparser.Construct()
	if err := infile.Open("engine/elements.txt", true, mods); err != nil && utils.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	defer infile.Close()

	for infile.Next(mods) {
		if infile.IsNewSection() {
			if infile.GetSection() == "element" {
				// check if the previous element and remove it if there is no identifier
				lenList := len(this.list)
				if lenList != 0 && this.list[lenList-1].id == "" {
					this.list = this.list[:lenList-1]
				}

				this.list = append(this.list, Element{})
			}
		}

		if len(this.list) == 0 || infile.GetSection() != "element" {
			continue
		}

		lenList := len(this.list)
		if infile.Key() == "id" {
			this.list[lenList-1].id = infile.Val()
		} else if infile.Key() == "name" {
			this.list[lenList-1].name = infile.Val()
		} else {
			fmt.Printf("EngineSettings: '%s' is not a valid key.\n", infile.Key())
			return common.Err_bad_key_in_enginesettings
		}
	}

	lenList := len(this.list)
	if lenList != 0 && this.list[lenList-1].id == "" {
		this.list = this.list[:lenList-1]
	}

	return nil
}

type EquipFlags struct {
	list []EquipFlag
}

type EquipFlag struct {
	id   string
	name string
}

func (this *EquipFlags) load(settings common.Settings, mods common.ModManager) error {
	this.list = this.list[:0]

	infile := fileparser.Construct()
	if err := infile.Open("engine/equip_flags.txt", true, mods); err != nil && utils.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	defer infile.Close()

	for infile.Next(mods) {
		if infile.IsNewSection() {
			if infile.GetSection() == "flag" {
				// check if the previous flag and remove it if there is no identifier
				lenList := len(this.list)
				if lenList != 0 && this.list[lenList-1].id == "" {
					this.list = this.list[:lenList-1]
				}

				this.list = append(this.list, EquipFlag{})
			}
		}

		if len(this.list) == 0 || infile.GetSection() != "flag" {
			continue
		}

		lenList := len(this.list)
		if infile.Key() == "id" {
			this.list[lenList-1].id = infile.Val()
		} else if infile.Key() == "name" {
			this.list[lenList-1].name = infile.Val()
		} else {
			fmt.Printf("EngineSettings: '%s' is not a valid key.\n", infile.Key())
			return common.Err_bad_key_in_enginesettings
		}
	}

	lenList := len(this.list)
	if lenList != 0 && this.list[lenList-1].id == "" {
		this.list = this.list[:lenList-1]
	}

	return nil
}

type PrimaryStats struct {
	list []PrimaryStat
}

type PrimaryStat struct {
	id   string
	name string
}

func (this *PrimaryStat) GetName() string {
	return this.name
}

func (this *PrimaryStat) GetId() string {
	return this.id
}

func (this *PrimaryStats) get(key string) interface{} {
	if key == "list" {
		tmpList := make([]common.PrimaryStat, len(this.list))

		for i, val := range this.list {
			tmp := val
			tmpList[i] = &tmp
		}

		return tmpList
	}

	return nil
}

func (this *PrimaryStats) load(settings common.Settings, mods common.ModManager) error {
	this.list = this.list[:0]

	infile := fileparser.Construct()
	if err := infile.Open("engine/primary_stats.txt", true, mods); err != nil && utils.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	defer infile.Close()

	for infile.Next(mods) {
		if infile.IsNewSection() {
			if infile.GetSection() == "stat" {
				// check if the previous stat and remove it if there is no identifier
				lenList := len(this.list)
				if lenList != 0 && this.list[lenList-1].id == "" {
					this.list = this.list[:lenList-1]
				}

				this.list = append(this.list, PrimaryStat{})
			}
		}

		if len(this.list) == 0 || infile.GetSection() != "stat" {
			continue
		}

		lenList := len(this.list)
		if infile.Key() == "id" {
			this.list[lenList-1].id = infile.Val()
		} else if infile.Key() == "name" {
			this.list[lenList-1].name = infile.Val()
		} else {
			fmt.Printf("EngineSettings: '%s' is not a valid key.\n", infile.Key())
			return common.Err_bad_key_in_enginesettings
		}
	}

	lenList := len(this.list)
	if lenList != 0 && this.list[lenList-1].id == "" {
		this.list = this.list[:lenList-1]
	}

	return nil
}

func (this *PrimaryStats) PrimaryStatsGetIndexById(id string) (int, bool) {
	return this.GetIndexById(id)

}

func (this *PrimaryStats) GetIndexById(id string) (int, bool) {
	for i := 0; i < len(this.list); i++ {
		if id == this.list[i].id {
			return i, true
		}
	}
	return len(this.list), false
}

func (this *PrimaryStats) getSize() int {
	return len(this.list)
}

type HeroClasses struct {
	eset *EngineSettings
	list []HeroClass
}

type HeroClass struct {
	name            string
	description     string
	currency        int
	equipment       string
	carried         string
	primary         []int
	actionbar       []define.PowerId
	powers          []define.PowerId
	campaign        []string
	powerTree       string
	defaultPowerTab int
	heroOptions     []int
}

func (this *HeroClass) GetName() string {
	return this.name
}

func (this *HeroClass) GetDescription() string {
	return this.description
}

func (this *HeroClass) GetHeroOptions() []int {
	tmp := make([]int, len(this.heroOptions))
	for index, val := range this.heroOptions {
		tmp[index] = val
	}

	return tmp
}

func (this *HeroClasses) get(key string) []common.HeroClass {
	if key == "list" {
		tmpList := make([]common.HeroClass, len(this.list))

		for i, val := range this.list {
			tmp := val
			tmpList[i] = &tmp
		}

		return tmpList
	}

	return nil
}

func (this *HeroClasses) load(settings common.Settings, mods common.ModManager, msg common.MessageEngine) error {
	this.list = this.list[:0]

	infile := fileparser.Construct()
	if err := infile.Open("engine/classes.txt", true, mods); err != nil && utils.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	defer infile.Close()

	for infile.Next(mods) {
		if infile.IsNewSection() {
			if infile.GetSection() == "class" {
				// check if the previous stat and remove it if there is no identifier
				lenList := len(this.list)
				if lenList != 0 && this.list[lenList-1].name == "" {
					this.list = this.list[:lenList-1]
				}

				this.list = append(this.list, HeroClass{
					primary:         make([]int, this.eset.PrimaryStats.getSize()),
					actionbar:       make([]define.PowerId, actionbar.SLOT_MAX),
					defaultPowerTab: -1,
				})
			}
		}

		if len(this.list) == 0 || infile.GetSection() != "class" {
			continue
		}

		lenList := len(this.list)
		switch infile.Key() {
		case "name":
			this.list[lenList-1].name = infile.Val()
		case "description":
			this.list[lenList-1].description = infile.Val()
		case "currency":
			this.list[lenList-1].currency = parsing.ToInt(infile.Val(), 0)
		case "equipment":
			this.list[lenList-1].equipment = infile.Val()
		case "carried":
			this.list[lenList-1].carried = infile.Val()
		case "primary":
			strVal := infile.Val()
			primStat := ""
			primStat, strVal = parsing.PopFirstString(strVal, "")
			primStatIndex, ok := this.eset.PrimaryStats.GetIndexById(primStat)
			if !ok {
				fmt.Printf("EngineSettings: '%s' is not a valid key.\n", infile.Key())
				return common.Err_bad_key_in_enginesettings
			}
			this.list[lenList-1].primary[primStatIndex] = parsing.ToInt(strVal, 0)
		case "actionbar":
			for i := 0; i < actionbar.SLOT_MAX; i++ {
				this.list[lenList-1].actionbar[i] = parsing.ToPowerId(infile.Val(), 0)
			}
		case "powers":
			power := ""
			strVal := infile.Val()

			for {
				if strVal == "" {
					break
				}

				power, strVal = parsing.PopFirstString(strVal, "")
				this.list[lenList-1].powers = append(this.list[lenList-1].powers, parsing.ToPowerId(power, 0))
			}

		case "campaign":
			status := ""
			strVal := infile.Val()
			for {
				if strVal == "" {
					break
				}

				status, strVal = parsing.PopFirstString(strVal, "")
				this.list[lenList-1].campaign = append(this.list[lenList-1].campaign, status)
			}

		case "power_tree":
			this.list[lenList-1].powerTree = infile.Val()
		case "hero_options":
			heroOption := ""
			strVal := infile.Val()
			for {
				if strVal == "" {
					break
				}

				heroOption, strVal = parsing.PopFirstString(strVal, "")
				this.list[lenList-1].heroOptions = append(this.list[lenList-1].heroOptions, parsing.ToInt(heroOption, 0))
			}

			sort.Slice(this.list[lenList-1].heroOptions, func(i, j int) bool { return this.list[lenList-1].heroOptions[i] < this.list[lenList-1].heroOptions[j] })

		case "default_power_tab":
			this.list[lenList-1].defaultPowerTab = parsing.ToInt(infile.Val(), 0)

		default:
			fmt.Printf("EngineSettings: '%s' is not a valid key.\n", infile.Key())
			return common.Err_bad_key_in_enginesettings
		}
	}

	lenList := len(this.list)
	if lenList != 0 && this.list[lenList-1].name == "" {
		this.list = this.list[:lenList-1]
	}

	// Make a default hero class if none were found
	if len(this.list) == 0 {
		msg.Get("Adventurer")
		this.list = append(this.list, HeroClass{
			name:            "Adventurer",
			primary:         make([]int, this.eset.PrimaryStats.getSize()),
			actionbar:       make([]define.PowerId, actionbar.SLOT_MAX),
			defaultPowerTab: -1,
		})
	}
	return nil
}

type DamageType struct {
	id          string
	name        string
	nameMin     string
	nameMax     string
	description string
	min         string
	max         string
}

func (this *DamageType) GetId() string {
	return this.id
}
func (this *DamageType) GetName() string {
	return this.name
}
func (this *DamageType) GetNameMin() string {
	return this.nameMin
}
func (this *DamageType) GetNameMax() string {
	return this.nameMax
}
func (this *DamageType) GetDescription() string {
	return this.description
}
func (this *DamageType) GetMin() string {
	return this.min
}
func (this *DamageType) GetMax() string {
	return this.max
}

type DamageTypes struct {
	list  []DamageType
	count int
}

func (this *DamageTypes) get(key string) interface{} {
	switch key {
	case "list":
		tmpList := make([]common.DamageType, len(this.list))
		for i, val := range this.list {
			tmp := val
			tmpList[i] = &tmp
		}
		return tmpList
	case "count":
		return this.count
	}

	return nil
}

func (this *DamageTypes) load(settings common.Settings, mods common.ModManager, msg common.MessageEngine) error {
	this.list = this.list[:0]

	infile := fileparser.Construct()
	if err := infile.Open("engine/damage_types.txt", true, mods); err != nil && utils.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	defer infile.Close()

	for infile.Next(mods) {
		if infile.IsNewSection() {
			if infile.GetSection() == "damage_type" {
				// check if the previous stat and remove it if there is no identifier
				lenList := len(this.list)
				if lenList != 0 && this.list[lenList-1].id == "" {
					this.list = this.list[:lenList-1]
				}

				this.list = append(this.list, DamageType{})
			}
		}

		if len(this.list) == 0 || infile.GetSection() != "damage_type" {
			continue
		}

		lenList := len(this.list)
		switch infile.Key() {
		case "id":
			this.list[lenList-1].id = infile.Val()
		case "name":
			this.list[lenList-1].name = msg.Get(infile.Val())
		case "name_min":
			this.list[lenList-1].nameMin = msg.Get(infile.Val())
		case "name_max":
			this.list[lenList-1].nameMax = msg.Get(infile.Val())
		case "description":
			this.list[lenList-1].description = msg.Get(infile.Val())
		case "min":
			this.list[lenList-1].min = infile.Val()
		case "max":
			this.list[lenList-1].max = infile.Val()

		default:
			fmt.Printf("EngineSettings: '%s' is not a valid key.\n", infile.Key())
			return common.Err_bad_key_in_enginesettings
		}
	}

	lenList := len(this.list)
	this.count = lenList * 2 // 伤害的范围故*2

	// use the IDs if the damage type doesn't have printable names
	for i := 0; i < lenList; i++ {
		if this.list[i].name == "" {
			this.list[i].name = this.list[i].id
		}

		if this.list[i].nameMin == "" {
			this.list[i].nameMin = this.list[i].min
		}

		if this.list[i].nameMax == "" {
			this.list[i].nameMax = this.list[i].max
		}
	}

	return nil
}

type DeathPenalty struct {
	deathpenalty map[string]*ConfigEntry
}

func (this *DeathPenalty) get(key string) interface{} {
	if val, ok := this.deathpenalty[key]; ok {
		return val.storage
	}

	return nil
}

func (this *DeathPenalty) load(settings common.Settings, mods common.ModManager) error {
	if this.deathpenalty == nil {
		this.deathpenalty = map[string]*ConfigEntry{}
	}

	this.deathpenalty["enable"] = &ConfigEntry{"enable", true}
	this.deathpenalty["permadeath"] = &ConfigEntry{"permadeath", true}
	this.deathpenalty["currency"] = &ConfigEntry{"currency", 50}
	this.deathpenalty["xp_total"] = &ConfigEntry{"xp_total", 0}
	this.deathpenalty["xp_current_level"] = &ConfigEntry{"xp_current_level", 0}
	this.deathpenalty["random_item"] = &ConfigEntry{"random_item", false}

	infile := fileparser.Construct()

	if err := infile.Open("engine/death_penalty.txt", true, mods); err != nil && utils.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	defer infile.Close()

	for infile.Next(mods) {
		if entry, ok := this.deathpenalty[infile.Key()]; ok {
			parsing.TryParseValue(infile.Val(), &(entry.storage))
		} else {
			fmt.Printf("EngineSettings: '%s' is not a valid key.\n", infile.Key())
			return common.Err_bad_key_in_enginesettings
		}
	}

	return nil
}

type Tooltips struct {
	tooltips map[string]*ConfigEntry
}

func (this *Tooltips) load(settings common.Settings, mods common.ModManager) error {
	if this.tooltips == nil {
		this.tooltips = map[string]*ConfigEntry{}
	}

	this.tooltips["tooltip_offset"] = &ConfigEntry{"tooltip_offset", 0}
	this.tooltips["tooltip_width"] = &ConfigEntry{"tooltip_width", 1}
	this.tooltips["tooltip_margin"] = &ConfigEntry{"tooltip_margin", 0}
	this.tooltips["npc_tooltip_margin"] = &ConfigEntry{"npc_tooltip_margin", 0}
	this.tooltips["tooltip_background_border"] = &ConfigEntry{"tooltip_background_border", 0}

	infile := fileparser.Construct()

	if err := infile.Open("engine/tooltips.txt", true, mods); err != nil && utils.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	defer infile.Close()

	for infile.Next(mods) {
		if entry, ok := this.tooltips[infile.Key()]; ok {
			parsing.TryParseValue(infile.Val(), &(entry.storage))
		} else {
			fmt.Printf("EngineSettings: '%s' is not a valid key.\n", infile.Key())
			return common.Err_bad_key_in_enginesettings
		}
	}

	return nil
}

func (this *Tooltips) get(key string) interface{} {
	if val, ok := this.tooltips[key]; ok {
		return val.storage
	}

	return nil
}

type Loot struct {
	loot map[string]*ConfigEntry
	eset *EngineSettings
}

func (this *Loot) get(key string) interface{} {
	if val, ok := this.loot[key]; ok {
		return val.storage
	}

	return nil
}

func (this *Loot) load(settings common.Settings, mods common.ModManager, msg common.MessageEngine) error {
	if this.loot == nil {
		this.loot = map[string]*ConfigEntry{}
	}

	this.loot["tooltip_margin"] = &ConfigEntry{"tooltip_margin", 0}
	this.loot["autopickup_currency"] = &ConfigEntry{"autopickup_currency", false}
	this.loot["autopickup_range"] = &ConfigEntry{"autopickup_range", (float32)(this.eset.Misc.get("interact_range").(int))}
	this.loot["currency_name"] = &ConfigEntry{"currency_name", "Gold"}
	this.loot["vendor_ratio"] = &ConfigEntry{"vendor_ratio", (float32)(0.25)}
	this.loot["vendor_ratio_buyback"] = &ConfigEntry{"vendor_ratio_buyback", (float32)(0)}
	this.loot["sfx_loot"] = &ConfigEntry{"sfx_loot", ""}
	this.loot["drop_max"] = &ConfigEntry{"drop_max", 1}
	this.loot["drop_radius"] = &ConfigEntry{"drop_radius", 1}
	this.loot["hide_radius"] = &ConfigEntry{"hide_radius", (float32)(3.0)}

	infile := fileparser.Construct()

	if err := infile.Open("engine/loot.txt", true, mods); err != nil && utils.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	defer infile.Close()

	for infile.Next(mods) {
		if entry, ok := this.loot[infile.Key()]; ok {
			parsing.TryParseValue(infile.Val(), &(entry.storage))

			switch infile.Key() {
			case "currency_name":
				entry.storage = msg.Get(entry.storage.(string))
			case "vendor_ratio":
				fallthrough
			case "vendor_ratio_buyback":
				entry.storage = entry.storage.(float32) / 100.0
			case "drop_max":
				fallthrough
			case "drop_radius":
				entry.storage = (int)(math.Max((float64)(entry.storage.(int)), 1))
			}
		} else {
			fmt.Printf("EngineSettings: '%s' is not a valid key.\n", infile.Key())
			return common.Err_bad_key_in_enginesettings
		}
	}

	return nil
}

type Tileset struct {
	tileset map[string]*ConfigEntry
}

func (this *Tileset) load(settings common.Settings, mods common.ModManager) error {
	if this.tileset == nil {
		this.tileset = map[string]*ConfigEntry{}
	}

	this.tileset["units_per_pixel_x"] = &ConfigEntry{"units_per_pixel_x", float32(2)}
	this.tileset["units_per_pixel_y"] = &ConfigEntry{"units_per_pixel_y", float32(4)}
	this.tileset["tile_size"] = &ConfigEntry{"tile_size", []int{64, 32}}
	this.tileset["tile_w_half"] = &ConfigEntry{"tile_w_half", 64}
	this.tileset["tile_h_half"] = &ConfigEntry{"tile_h_half", 32}
	this.tileset["orientation"] = &ConfigEntry{"orientation", enginesettings.TILESET_ISOMETRIC}

	infile := fileparser.Construct()

	if err := infile.Open("engine/tileset_config.txt", true, mods); err != nil && utils.IsNotExist(err) {
		fmt.Println("Unable to open engine/tileset_config.txt! Defaulting to 64x32 isometric tiles.")
		return nil
	} else if err != nil {
		return err
	}

	defer infile.Close()

	for infile.Next(mods) {
		switch infile.Key() {
		case "tile_size":
			strVal := infile.Val()
			tile_w := 0
			tile_h := 0
			tile_w, strVal = parsing.PopFirstInt(strVal, "")
			tile_h, strVal = parsing.PopFirstInt(strVal, "")
			this.tileset["tile_size"].storage = []int{tile_w, tile_h}
			this.tileset["tile_w_half"].storage = tile_w / 2
			this.tileset["tile_h_half"].storage = tile_h / 2
		case "orientation":
			if infile.Val() == "isometric" {
				this.tileset["orientation"].storage = enginesettings.TILESET_ISOMETRIC
			} else if infile.Val() == "orthogonal" {
				this.tileset["orientation"].storage = enginesettings.TILESET_ORTHOGONAL
			}
		default:

			fmt.Printf("EngineSettings: '%s' is not a valid key.\n", infile.Key())
			return common.Err_bad_key_in_enginesettings
		}
	}

	tile_size := this.tileset["tile_size"].storage.([]int)
	tile_w := tile_size[0]
	tile_h := tile_size[1]

	if tile_w > 0 && tile_h > 0 {
		if this.tileset["orientation"].storage.(int) == enginesettings.TILESET_ISOMETRIC {
			this.tileset["units_per_pixel_x"].storage = 2 / float32(tile_w)
			this.tileset["units_per_pixel_y"].storage = 2 / float32(tile_h)
		} else {
			this.tileset["units_per_pixel_x"].storage = 1 / float32(tile_w)
			this.tileset["units_per_pixel_y"].storage = 1 / float32(tile_h)
		}
	} else {
		fmt.Println("EngineSettings: Tile dimensions must be greater than 0. Resetting to the default size of 64x32.")
		this.tileset["tile_size"].storage = []int{64, 32}
	}

	units_per_pixel_x := this.tileset["units_per_pixel_x"].storage.(float32)
	units_per_pixel_y := this.tileset["units_per_pixel_y"].storage.(float32)
	if units_per_pixel_x == 0 || units_per_pixel_y == 0 {
		fmt.Printf("EngineSettings: One of UNITS_PER_PIXEL values is zero! %fx%f\n", units_per_pixel_x, units_per_pixel_y)
		//logfile.LogErrorDialog("EngineSettings: One of UNITS_PER_PIXEL values is zero! %fx%f", units_per_pixel_x, units_per_pixel_y)
		return common.Err_bad_val_in_enginesettings
	}

	return nil
}

func (this *Tileset) get(key string) interface{} {
	if val, ok := this.tileset[key]; ok {
		return val.storage
	}

	return nil
}

type Widgets struct {
	widgets map[string]*ConfigEntry
}

func (this *Widgets) get(key string) interface{} {
	if val, ok := this.widgets[key]; ok {
		return val.storage
	}

	return nil
}

func (this *Widgets) load(settings common.Settings, mods common.ModManager, font common.FontEngine) error {
	if this.widgets == nil {
		this.widgets = map[string]*ConfigEntry{}
	}

	this.widgets["selection_rect_color"] = &ConfigEntry{"selection_rect_color", color.Construct(255, 248, 220, 255)}
	this.widgets["colorblind_highlight_offset"] = &ConfigEntry{"colorblind_highlight_offset", point.Construct(2, 2)}
	this.widgets["padding"] = &ConfigEntry{"padding", point.Construct(8, 0)}
	//this.widgets["quantity_label"] = &ConfigEntry{"quantity_label", widget.ConstructLabel(font)}
	//this.widgets["quantity_label"] = &ConfigEntry{"quantity_label", widgetf.New("label").(common.WidgetLabel).Init1(font)}
	this.widgets["quantity_label"] = &ConfigEntry{"quantity_label", labelinfo.Construct()}

	this.widgets["quantity_color"] = &ConfigEntry{"quantity_color", font.GetColor(fontengine.COLOR_WIDGET_NORMAL)}
	this.widgets["quantity_bg_color"] = &ConfigEntry{"quantity_bg_color", color.Construct(0, 0, 0, 0)}
	//wl := widget.ConstructLabel(font)
	//wl := widgetf.New("label").(common.WidgetLabel).Init1(font)
	wl := labelinfo.Construct()
	wl.Hidden = true
	//wl.SetHidden(true)
	this.widgets["hotkey_label"] = &ConfigEntry{"hotkey_label", wl}
	this.widgets["hotkey_color"] = &ConfigEntry{"hotkey_color", font.GetColor(fontengine.COLOR_WIDGET_NORMAL)}
	this.widgets["hotkey_bg_color"] = &ConfigEntry{"hotkey_bg_color", color.Construct(0, 0, 0, 0)}
	this.widgets["text_margin"] = &ConfigEntry{"text_margin", point.Construct(8, 0)}
	this.widgets["text_width"] = &ConfigEntry{"text_width", 150}
	this.widgets["bg_color"] = &ConfigEntry{"bg_color", color.Construct(0, 0, 0, 64)}

	infile := fileparser.Construct()

	if err := infile.Open("engine/widget_settings.txt", true, mods); err != nil && utils.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	defer infile.Close()

	for infile.Next(mods) {
		if entry, ok := this.widgets[infile.Key()]; ok {
			switch infile.GetSection() {
			case "misc":
				if infile.Key() == "selection_rect_color" {
					entry.storage = parsing.ToRGBA(infile.Val())
				} else if infile.Key() == "colorblind_highlight_offset" {
					entry.storage = parsing.ToPoint(infile.Val())
				}
			case "tab":
				if infile.Key() == "padding" {
					entry.storage = parsing.ToPoint(infile.Val())
				}
			case "slot":
				if infile.Key() == "quantity_label" {
					entry.storage = parsing.PopLabelInfo(infile.Val())
				} else if infile.Key() == "quantity_color" {
					entry.storage = parsing.ToRGB(infile.Val())
				} else if infile.Key() == "quantity_bg_color" {
					entry.storage = parsing.ToRGBA(infile.Val())
				} else if infile.Key() == "hotkey_label" {
					entry.storage = parsing.PopLabelInfo(infile.Val())
				} else if infile.Key() == "hotkey_color" {
					entry.storage = parsing.ToRGB(infile.Val())
				} else if infile.Key() == "hotkey_bg_color" {
					entry.storage = parsing.ToRGBA(infile.Val())
				}

			case "listbox":
				if infile.Key() == "text_margin" {
					entry.storage = parsing.ToPoint(infile.Val())
				}
			case "horizontal_list":
				if infile.Key() == "text_width" {
					entry.storage = parsing.ToInt(infile.Val(), 0)
				}
			case "scrollbar":
				if infile.Key() == "bg_color" {
					entry.storage = parsing.ToRGBA(infile.Val())
				}

			}
		} else {
			fmt.Printf("EngineSettings: '%s' is not a valid key.\n", infile.Key())
			return common.Err_bad_key_in_enginesettings
		}

	}

	return nil
}

type XPTable struct {
	xpTable []uint64
}

func (this *XPTable) load(settings common.Settings, mods common.ModManager) error {
	infile := fileparser.Construct()

	if err := infile.Open("engine/xp_table.txt", true, mods); err != nil && utils.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	defer infile.Close()

	for infile.Next(mods) {
		if infile.Key() == "level" {
			strVal := infile.Val()
			strLvlXp := ""
			_, strVal = parsing.PopFirstInt(strVal, "")
			strLvlXp, strVal = parsing.PopFirstString(strVal, "")
			lvlXp := parsing.ToUnsignedLong(strLvlXp, 0)

			this.xpTable = append(this.xpTable, lvlXp)

		} else {
			fmt.Printf("EngineSettings: '%s' is not a valid key.\n", infile.Key())
			return common.Err_bad_key_in_enginesettings
		}
	}

	if len(this.xpTable) == 0 {
		this.xpTable = append(this.xpTable, 0)
	}

	return nil
}

func (this *XPTable) XPGetLevelXP(level int) uint64 {
	return this.getLevelXP(level)
}

func (this *XPTable) getLevelXP(level int) uint64 {
	if level <= 1 || len(this.xpTable) == 0 {
		return 0
	} else if level > len(this.xpTable) {
		return this.xpTable[len(this.xpTable)-1]
	}
	return this.xpTable[level-1]
}

func (this *XPTable) XPGetMaxLevel() int {
	return this.getMaxLevel()
}

func (this *XPTable) getMaxLevel() int {
	return len(this.xpTable)
}

func (this *XPTable) XPGetLevelFromXP(levelXp uint64) int {
	return this.getLevelFromXP(levelXp)
}
func (this *XPTable) getLevelFromXP(levelXp uint64) int {
	level := 0
	for i, xp := range this.xpTable {
		if levelXp >= xp {
			level = i + 1
		}
	}

	return level
}
