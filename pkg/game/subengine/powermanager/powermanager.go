package powermanager

import (
	"fmt"
	"math"
	"monster/pkg/common"
	"monster/pkg/common/define"
	"monster/pkg/common/define/game/mapcollision"
	"monster/pkg/common/define/game/stats"
	"monster/pkg/common/fpoint"
	"monster/pkg/common/gameres"
	"monster/pkg/common/gameres/effect"
	"monster/pkg/common/gameres/power"
	"monster/pkg/common/point"
	"monster/pkg/common/timer"
	"monster/pkg/filesystem/fileparser"
	misceffect "monster/pkg/game/misc/effect"
	"monster/pkg/utils"
	"monster/pkg/utils/parsing"
	"monster/pkg/utils/tools"
)

type PowerManager struct {
	collider gameres.MapCollision

	powerAnimations  map[define.PowerId]common.Animation
	effectAnimations []common.Animation

	effects           []effect.Def
	powers            map[define.PowerId]*power.Power
	usedItems         []define.ItemId // 技能已经消耗掉的仓库道具
	usedEquippedItems []define.ItemId // 技能已经消耗掉的已经装备的道具
}

func New(modules common.Modules, ss gameres.Stats) *PowerManager {
	pm := &PowerManager{}

	pm.init(modules, ss)

	return pm
}

func (this *PowerManager) init(modules common.Modules, ss gameres.Stats) gameres.PowerManager {
	this.powerAnimations = map[define.PowerId]common.Animation{}
	this.powers = map[define.PowerId]*power.Power{}

	err := this.loadEffects(modules, ss)
	if err != nil {
		panic(err)
	}

	err = this.loadPowers(modules, ss)
	if err != nil {
		panic(err)
	}

	return this
}

func (this *PowerManager) clear() {
	for _, ptr := range this.powerAnimations {
		ptr.Close()
	}
}

func (this *PowerManager) Close() {
	this.clear()
}

func (this *PowerManager) loadEffects(modules common.Modules, ss gameres.Stats) error {
	mods := modules.Mods()
	anim := modules.Anim()
	settings := modules.Settings()
	render := modules.Render()
	mresf := modules.Resf()

	infile := fileparser.New()

	err := infile.Open("powers/effects.txt", true, mods)
	if err != nil {
		return err
	}
	defer infile.Close()

	for infile.Next(mods) {

		if infile.IsNewSection() {
			if infile.GetSection() == "effect" {
				if len(this.effects) != 0 && this.effects[len(this.effects)-1].Id == "" {
					this.effects = this.effects[:len(this.effects)-1]
				}

				this.effects = append(this.effects, effect.ConstructDef())
				this.effectAnimations = append(this.effectAnimations, nil)
			}
		}

		if len(this.effects) == 0 || infile.GetSection() != "effect" {
			continue
		}

		key := infile.Key()
		val := infile.Val()
		ptr := &(this.effects[len(this.effects)-1])

		switch key {
		case "id":
			ptr.Id = val
		case "type":
			ptr.Type = misceffect.GetTypeFromString(modules, ss, val)
		case "name":
			ptr.Name = val
		case "icon":
			ptr.Icon = parsing.ToInt(val, 0)
		case "animation":
			ptr.Animation = val
			anim.IncreaseCount(val)
			a, err := anim.GetAnimationSet(settings, mods, render, mresf, val)
			if err != nil {
				return err
			}
			this.effectAnimations[len(this.effectAnimations)-1] = a.GetAnimation("")
		case "can_stack":
			ptr.CanStack = parsing.ToBool(val)
		case "max_stacks":
			ptr.MaxStacks = parsing.ToInt(val, 0)
		case "group_stack":
			ptr.GroupStack = parsing.ToBool(val)
		case "render_above":
			ptr.RenderAbove = parsing.ToBool(val)
		case "color_mod":
			ptr.ColorMod = parsing.ToRGB(val)
		case "alpha_mod":
			ptr.AlphaMod = (uint8)(parsing.ToInt(val, 0))
		case "attack_speed_anim":
			ptr.AttackSpeedAnim = val
		default:
			return fmt.Errorf("PowerManager: '%s' is not a valid key.\n", key)
		}

	}

	if len(this.effects) != 0 && this.effects[len(this.effects)-1].Id == "" {
		this.effects = this.effects[:len(this.effects)-1]
	}

	return nil
}

func (this *PowerManager) loadPowers(modules common.Modules, ss gameres.Stats) error {
	mods := modules.Mods()
	msg := modules.Msg()
	settings := modules.Settings()
	anim := modules.Anim()
	render := modules.Render()
	mresf := modules.Resf()
	eset := modules.Eset()

	infile := fileparser.New()

	err := infile.Open("powers/powers.txt", true, mods)
	if err != nil {
		return err
	}
	defer infile.Close()

	clearPostEffects := true
	idLine := false
	inputId := define.PowerId(0)
	for infile.Next(mods) {
		key := infile.Key()
		val := infile.Val()

		if key == "id" {
			idLine = true
			inputId = define.PowerId(parsing.ToInt(val, 0))
			this.powers[inputId] = power.New(modules)
			clearPostEffects = true
			this.powers[inputId].IsEmpty = false
			continue
		} else {
			idLine = false
		}

		if inputId < 1 {
			return fmt.Errorf("PowerManager: Power index out of bounds 1-%d, skipping power.\n", math.MaxInt)
		}

		if idLine {
			continue
		}

		switch key {
		case "type":
			if val == "fixed" {
				this.powers[inputId].Type = power.TYPE_FIXED
			} else if val == "missile" {
				// 飞弹
				this.powers[inputId].Type = power.TYPE_MISSILE
			} else if val == "repeater" {
				// 比如几段的伤害
				this.powers[inputId].Type = power.TYPE_REPEATER
			} else if val == "spawn" {
				this.powers[inputId].Type = power.TYPE_SPAWN
			} else if val == "transform" {
				this.powers[inputId].Type = power.TYPE_TRANSFORM
			} else if val == "block" {
				this.powers[inputId].Type = power.TYPE_BLOCK
			} else {
				return fmt.Errorf("PowerManager: Unknown type '%s'\n", val)
			}
		case "name":
			this.powers[inputId].Name = msg.Get(val)
		case "description":
			this.powers[inputId].Description = msg.Get(val)
		case "icon":
			this.powers[inputId].Icon = parsing.ToInt(val, 0)
		case "new_state":
			if val == "instant" {
				this.powers[inputId].NewState = power.STATE_INSTANT
			} else {
				this.powers[inputId].NewState = power.STATE_ATTACK
				this.powers[inputId].AttackAnim = val
			}

		case "state_duration":
			this.powers[inputId].StateDuration = parsing.ToDuration(val, settings.Get("max_fps").(int))
		case "prevent_interrupt":
			this.powers[inputId].PreventInterrupt = parsing.ToBool(val)
		case "face":
			this.powers[inputId].Face = parsing.ToBool(val)
		case "source_type":
			if val == "hero" {
				this.powers[inputId].SourceType = power.SOURCE_TYPE_HERO
			} else if val == "neutral" {
				this.powers[inputId].SourceType = power.SOURCE_TYPE_NEUTRAL
			} else if val == "enemy" {
				this.powers[inputId].SourceType = power.SOURCE_TYPE_ENEMY
			} else {
				return fmt.Errorf("PowerManager: Unknown source_type '%s'\n", val)
			}
		case "beacon":
			this.powers[inputId].Beacon = parsing.ToBool(val)
		case "count":
			this.powers[inputId].Count = parsing.ToInt(val, 0)
		case "passive":
			this.powers[inputId].Passive = parsing.ToBool(val)
		case "passive_trigger":
			if val == "on_block" {
				// 格挡时
				this.powers[inputId].PassiveTrigger = power.TRIGGER_BLOCK
			} else if val == "on_hit" {
				this.powers[inputId].PassiveTrigger = power.TRIGGER_HIT
			} else if val == "on_halfdeath" {
				this.powers[inputId].PassiveTrigger = power.TRIGGER_HALFDEATH
			} else if val == "on_joincombat" {
				this.powers[inputId].PassiveTrigger = power.TRIGGER_JOINCOMBAT
			} else if val == "on_death" {
				this.powers[inputId].PassiveTrigger = power.TRIGGER_DEATH
			} else {
				return fmt.Errorf("PowerManager: Unknown passive trigger '%s'\n", val)
			}
		case "meta_power":
			this.powers[inputId].MetaPower = parsing.ToBool(val)
		case "no_actionbar":
			this.powers[inputId].NoActionbar = parsing.ToBool(val)
		case "requires_flags":
			this.powers[inputId].RequiresFlags = map[string]struct{}{}
			var flag string
			flag, val = parsing.PopFirstString(val, "")

			for flag != "" {
				this.powers[inputId].RequiresFlags[flag] = struct{}{}
				flag, val = parsing.PopFirstString(val, "")
			}
		case "requires_mp":
			this.powers[inputId].RequiresMP = parsing.ToInt(val, 0)
		case "requires_hp":
			this.powers[inputId].RequiresHP = parsing.ToInt(val, 0)
		case "sacrifice":
			this.powers[inputId].Sacrifice = parsing.ToBool(val)
		case "requires_los":
			this.powers[inputId].RequiresLos = parsing.ToBool(val)
			this.powers[inputId].RequiresLosDefault = false
		case "requires_empty_target":
			this.powers[inputId].RequiresEmptyTarget = parsing.ToBool(val)
		case "requires_item":
			// 在仓库
			pri := power.ConstructRequiredItem()
			var first string
			first, val = parsing.PopFirstString(val, "")
			pri.Id = (define.ItemId)(parsing.ToInt(first, 0))
			first, val = parsing.PopFirstString(val, "")
			pri.Quantity = parsing.ToInt(first, 0) // 扣除的数量
			pri.Equipped = false
			this.powers[inputId].RequiredItems = append(this.powers[inputId].RequiredItems, pri)
		case "requires_equipped_item":
			// 需要装备
			pri := power.ConstructRequiredItem()
			var first string
			first, val = parsing.PopFirstString(val, "")
			pri.Id = (define.ItemId)(parsing.ToInt(first, 0))
			pri.Quantity, val = parsing.PopFirstInt(val, "") // 扣除的数量
			pri.Equipped = true
			if pri.Quantity > 1 {
				fmt.Println("PowerManager: Only 1 equipped item can be consumed at a time.")
				pri.Quantity = (int)(math.Min((float64)(pri.Quantity), 1))
			}

			this.powers[inputId].RequiredItems = append(this.powers[inputId].RequiredItems, pri)
		case "requires_targeting":
			this.powers[inputId].RequiresTargeting = parsing.ToBool(val)
		case "requires_spawns":
			this.powers[inputId].RequiresSpawns = parsing.ToInt(val, 0)
		case "cooldown":
			this.powers[inputId].Cooldown = parsing.ToDuration(val, settings.Get("max_fps").(int))
		case "requires_hpmp_state":
			var mode, stateHP, stateHPVal, stateMP, stateMPVal string
			mode, val = parsing.PopFirstString(val, "")
			stateHP, val = parsing.PopFirstString(val, "")
			stateHPVal, val = parsing.PopFirstString(val, "")
			stateMP, val = parsing.PopFirstString(val, "")
			stateMPVal, val = parsing.PopFirstString(val, "")

			if stateHPVal == "" {
				this.powers[inputId].RequiresMaxHPMP.HP = -1
			} else {
				this.powers[inputId].RequiresMaxHPMP.HP = parsing.ToInt(stateHPVal, 0)
			}

			if stateMPVal == "" {
				this.powers[inputId].RequiresMaxHPMP.MP = -1
			} else {
				this.powers[inputId].RequiresMaxHPMP.MP = parsing.ToInt(stateMPVal, 0)
			}

			if stateHP == "percent" {
				this.powers[inputId].RequiresMaxHPMP.HPState = power.HPMPSTATE_PERCENT
			} else if stateHP == "not_percent" {
				this.powers[inputId].RequiresMaxHPMP.HPState = power.HPMPSTATE_NOT_PERCENT
			} else if stateHP == "ignore" || stateHP == "" {
				this.powers[inputId].RequiresMaxHPMP.HPState = power.HPMPSTATE_IGNORE
				this.powers[inputId].RequiresMaxHPMP.HP = -1
			} else {
				return fmt.Errorf("PowerManager: '%s' is not a valid hp/mp state. Use 'percent', 'not_percent', or 'ignore'.\n", stateHP)
			}

			if stateMP == "percent" {
				this.powers[inputId].RequiresMaxHPMP.MPState = power.HPMPSTATE_PERCENT
			} else if stateMP == "not_percent" {
				this.powers[inputId].RequiresMaxHPMP.MPState = power.HPMPSTATE_NOT_PERCENT
			} else if stateMP == "ignore" || stateMP == "" {
				this.powers[inputId].RequiresMaxHPMP.MPState = power.HPMPSTATE_IGNORE
				this.powers[inputId].RequiresMaxHPMP.MP = -1
			} else {
				return fmt.Errorf("PowerManager: '%s' is not a valid hp/mp state. Use 'percent', 'not_percent', or 'ignore'.\n", stateMP)
			}

			if mode == "any" {
				this.powers[inputId].RequiresMaxHPMP.Mode = power.HPMPSTATE_ANY
			} else if mode == "all" {
				this.powers[inputId].RequiresMaxHPMP.Mode = power.HPMPSTATE_ALL
			} else if mode == "hp" {
				fmt.Println("PowerManager: 'hp' has been deprecated. Use 'all' or 'any'.")

				this.powers[inputId].RequiresMaxHPMP.Mode = power.HPMPSTATE_ALL
				this.powers[inputId].RequiresMaxHPMP.MPState = power.HPMPSTATE_IGNORE
				this.powers[inputId].RequiresMaxHPMP.MP = -1

			} else if mode == "mp" {
				fmt.Println("PowerManager: 'mp' has been deprecated. Use 'all' or 'any'.")
				this.powers[inputId].RequiresMaxHPMP.Mode = power.HPMPSTATE_ALL
				this.powers[inputId].RequiresMaxHPMP.MPState = this.powers[inputId].RequiresMaxHPMP.HPState
				this.powers[inputId].RequiresMaxHPMP.MP = this.powers[inputId].RequiresMaxHPMP.HP
				this.powers[inputId].RequiresMaxHPMP.HPState = power.HPMPSTATE_IGNORE
				this.powers[inputId].RequiresMaxHPMP.HP = -1
			} else {
				return fmt.Errorf("PowerManager: Please specify 'any' or 'all'.\n")
			}

		case "animation":
			if this.powers[inputId].AnimationName != "" {
				anim.DecreaseCount(this.powers[inputId].AnimationName)
				this.powers[inputId].AnimationName = ""
			}

			if val != "" {
				this.powers[inputId].AnimationName = val
				anim.IncreaseCount(this.powers[inputId].AnimationName)
				aset, err := anim.GetAnimationSet(settings, mods, render, mresf, val)
				if err != nil {
					return err
				}
				this.powerAnimations[inputId] = aset.GetAnimation("")
			}
		case "soundfx":
			//TODO
		case "soundfx_hit":
			//TODO
		case "directional":
			this.powers[inputId].Directional = parsing.ToBool(val)
		case "visual_random":
			this.powers[inputId].VisualRandom = parsing.ToInt(val, 0)
		case "visual_option":
			this.powers[inputId].VisualOption = parsing.ToInt(val, 0)
		case "aim_assist":
			this.powers[inputId].AimAssist = parsing.ToBool(val)
		case "speed":
			// 飞弹的移动速度
			this.powers[inputId].Speed = parsing.ToFloat(val, 0) / (float32)(settings.Get("max_fps").(int))
		case "lifespan":
			this.powers[inputId].Lifespan = parsing.ToDuration(val, settings.Get("max_fps").(int))
		case "floor":
			this.powers[inputId].OnFloor = parsing.ToBool(val)
		case "complete_animation":
			this.powers[inputId].CompleteAnimation = parsing.ToBool(val)
		case "charge_speed":
			this.powers[inputId].ChargeSpeed = parsing.ToFloat(val, 0) / (float32)(settings.Get("max_fps").(int))
		case "attack_speed":
			this.powers[inputId].AttackSpeed = float32(parsing.ToInt(val, 0))
			if this.powers[inputId].AttackSpeed < 100 {
				fmt.Println("PowerManager: Attack speeds less than 100 are unsupported.")
				this.powers[inputId].AttackSpeed = 100
			}
		case "use_hazard":
			this.powers[inputId].UseHazard = parsing.ToBool(val)
		case "no_attack":
			this.powers[inputId].NoAttack = parsing.ToBool(val)
		case "no_aggro":
			this.powers[inputId].NoAggro = parsing.ToBool(val)
		case "radius":
			this.powers[inputId].Radius = parsing.ToFloat(val, 0)
		case "base_damage":
			dtList := eset.Get("damage_types", "list").([]common.DamageType)
			for i, ptr := range dtList {
				if val == ptr.GetId() {
					this.powers[inputId].BaseDamage = i
					break
				}
			}
		case "starting_pos":
			if val == "source" {
				this.powers[inputId].StartingPos = power.STARTING_POS_SOURCE
			} else if val == "target" {
				this.powers[inputId].StartingPos = power.STARTING_POS_TARGET
			} else if val == "melee" {
				this.powers[inputId].StartingPos = power.STARTING_POS_MELEE
			} else {
				return fmt.Errorf("PowerManager: Unknown starting_pos '%s'\n", val)
			}

		case "relative_pos":
			this.powers[inputId].RelativePos = parsing.ToBool(val)
		case "multitarget":
			this.powers[inputId].Multitarget = parsing.ToBool(val)
		case "multihit":
			this.powers[inputId].Multihit = parsing.ToBool(val)
		case "expire_with_caster":
			this.powers[inputId].ExpireWithCaster = parsing.ToBool(val)
		case "ignore_zero_damage":
			this.powers[inputId].IgnoreZeroDamage = parsing.ToBool(val)
		case "lock_target_to_direction":
			this.powers[inputId].LockTargetToDirection = parsing.ToBool(val)
		case "movement_type":
			if val == "ground" {
				this.powers[inputId].MovementType = mapcollision.MOVE_NORMAL
			} else if val == "flying" {
				this.powers[inputId].MovementType = mapcollision.MOVE_FLYING
			} else if val == "intangible" {
				this.powers[inputId].MovementType = mapcollision.MOVE_INTANGIBLE
			} else {
				return fmt.Errorf("PowerManager: Unknown movement_type '%s'\n", val)
			}
		case "trait_armor_penetration":
			this.powers[inputId].TraitArmorPenetration = parsing.ToBool(val)
		case "trait_avoidance_ignore":
			this.powers[inputId].TraitAvoidanceIgnore = parsing.ToBool(val)
		case "trait_crits_impaired":
			this.powers[inputId].TraitCritsImpaired = parsing.ToInt(val, 0)
		case "trait_elemental":
			eList := eset.Get("elements", "list").([]common.Element)
			for i, ptr := range eList {
				if val == ptr.GetId() {
					this.powers[inputId].TraitElemental = i
					break
				}
			}

		case "target_range":
			var first string
			first, val = parsing.PopFirstString(val, "")
			this.powers[inputId].TargetRange = parsing.ToFloat(first, 0)
		case "hp_steal":
			this.powers[inputId].HPSteal = parsing.ToInt(val, 0)
		case "mp_steal":
			this.powers[inputId].MPSteal = parsing.ToInt(val, 0)
		case "missile_angle":
			this.powers[inputId].MissileAngle = parsing.ToInt(val, 0)
		case "angle_variance":
			this.powers[inputId].AngleVariance = parsing.ToInt(val, 0)
		case "speed_variance":
			this.powers[inputId].SpeedVariance = parsing.ToFloat(val, 0)
		case "delay":
			this.powers[inputId].Delay = parsing.ToDuration(val, settings.Get("max_fps").(int))
		case "transform_duration":
			this.powers[inputId].TransformDuration = parsing.ToDuration(val, settings.Get("max_fps").(int))
		case "manual_untransform":
			this.powers[inputId].ManualUnTransform = parsing.ToBool(val)
		case "keep_equipment":
			this.powers[inputId].KeepEquipment = parsing.ToBool(val)
		case "untransform_on_hit":
			this.powers[inputId].UntransformOnHit = parsing.ToBool(val)
		case "buff":
			this.powers[inputId].Buff = parsing.ToBool(val)
		case "buff_teleport":
			this.powers[inputId].BuffTeleport = parsing.ToBool(val)
		case "buff_party":
			this.powers[inputId].BuffParty = parsing.ToBool(val)
		case "buff_party_power_id":
			this.powers[inputId].BuffPartyPowerId = (define.PowerId)(parsing.ToInt(val, 0))
		case "post_effect":
			fallthrough
		case "post_effect_src":
			if clearPostEffects {
				this.powers[inputId].PostEffects = nil
				clearPostEffects = false
			}

			pe := power.ConstructPostEffect()
			pe.Id, val = parsing.PopFirstString(val, "")
			if !this.isValidEffect(modules, ss, pe.Id) {
				return fmt.Errorf("PowerManager: Unknown effect '%s'\n", pe.Id)
			} else {
				if key == "post_effect_src" {
					pe.TargetSrc = true
				}

				pe.Magnitude, val = parsing.PopFirstInt(val, "")
				var first string
				first, val = parsing.PopFirstString(val, "")
				pe.Duration = parsing.ToDuration(first, settings.Get("max_fps").(int))
				first, val = parsing.PopFirstString(val, "")
				if first != "" {
					pe.Chance = parsing.ToInt(first, 0)
				}

				this.powers[inputId].PostEffects = append(this.powers[inputId].PostEffects, pe)
			}
		case "pre_power":
			var first int
			first, val = parsing.PopFirstInt(val, "")
			this.powers[inputId].PrePower = (define.PowerId)(first)
			var chance string
			chance, val = parsing.PopFirstString(val, "")
			if chance != "" {
				this.powers[inputId].PrePowerChance = parsing.ToInt(chance, 0)
			}
		case "post_power":
			var first int
			first, val = parsing.PopFirstInt(val, "")
			this.powers[inputId].PostPower = (define.PowerId)(first)
			var chance string
			chance, val = parsing.PopFirstString(val, "")
			if chance != "" {
				this.powers[inputId].PostPowerChance = parsing.ToInt(chance, 0)
			}
		case "wall_power":
			var first int
			first, val = parsing.PopFirstInt(val, "")
			this.powers[inputId].WallPower = (define.PowerId)(first)
			var chance string
			chance, val = parsing.PopFirstString(val, "")
			if chance != "" {
				this.powers[inputId].WallPowerChance = parsing.ToInt(chance, 0)
			}
		case "wall_reflect":
			this.powers[inputId].WallReflect = parsing.ToBool(val)
		case "spawn_type":
			this.powers[inputId].SpawnType = val
		case "target_neighbor":
			this.powers[inputId].TargetNeighbor = parsing.ToInt(val, 0)
		case "spawn_limit":
			var mode string
			mode, val = parsing.PopFirstString(val, "")
			if mode == "fixed" {
				this.powers[inputId].SpawnLimitMode = power.SPAWN_LIMIT_MODE_FIXED
			} else if mode == "stat" {
				this.powers[inputId].SpawnLimitMode = power.SPAWN_LIMIT_MODE_STAT
			} else if mode == "unlimited" {
				this.powers[inputId].SpawnLimitMode = power.SPAWN_LIMIT_MODE_UNLIMITED
			} else {
				return fmt.Errorf("PowerManager: Unknown spawn_limit_mode '%s'\n", mode)
			}

			if this.powers[inputId].SpawnLimitMode != power.SPAWN_LIMIT_MODE_UNLIMITED {
				this.powers[inputId].SpawnLimitQty, val = parsing.PopFirstInt(val, "")

				if this.powers[inputId].SpawnLimitMode == power.SPAWN_LIMIT_MODE_STAT {
					var stat string
					stat, val = parsing.PopFirstString(val, "")
					if primStatIndex, ok := eset.PrimaryStatsGetIndexById(stat); ok {
						this.powers[inputId].SpawnLimitStat = primStatIndex
					} else {
						return fmt.Errorf("PowerManager: '%s' is not a valid primary stat.\n", stat)
					}
				}
			}

		case "spawn_level":
			var mode string
			mode, val = parsing.PopFirstString(val, "")
			if mode == "default" {
				this.powers[inputId].SpawnLevelMode = power.SPAWN_LEVEL_MODE_DEFAULT
			} else if mode == "fixed" {
				this.powers[inputId].SpawnLevelMode = power.SPAWN_LEVEL_MODE_FIXED
			} else if mode == "stat" {
				this.powers[inputId].SpawnLevelMode = power.SPAWN_LEVEL_MODE_STAT
			} else if mode == "level" {
				this.powers[inputId].SpawnLevelMode = power.SPAWN_LEVEL_MODE_LEVEL
			} else {
				return fmt.Errorf("PowerManager: Unknown spawn_level_mode '%s'\n", mode)
			}

			if this.powers[inputId].SpawnLevelMode != power.SPAWN_LEVEL_MODE_DEFAULT {
				this.powers[inputId].SpawnLevelQty, val = parsing.PopFirstInt(val, "")

				if this.powers[inputId].SpawnLevelMode != power.SPAWN_LEVEL_MODE_FIXED {
					this.powers[inputId].SpawnLevelEvery, val = parsing.PopFirstInt(val, "")
				}

				if this.powers[inputId].SpawnLevelMode == power.SPAWN_LEVEL_MODE_STAT {
					var stat string
					stat, val = parsing.PopFirstString(val, "")
					if primStatIndex, ok := eset.PrimaryStatsGetIndexById(stat); ok {
						this.powers[inputId].SpawnLevelStat = primStatIndex
					} else {
						return fmt.Errorf("PowerManager: '%s' is not a valid primary stat.\n", stat)
					}
				}
			}

		case "target_party":
			this.powers[inputId].TargetParty = parsing.ToBool(val)
		case "target_categories":
			this.powers[inputId].TargetCategories = nil
			var cat string
			cat, val = parsing.PopFirstString(val, "")
			for cat != "" {
				this.powers[inputId].TargetCategories = append(this.powers[inputId].TargetCategories, cat)
				cat, val = parsing.PopFirstString(val, "")
			}
		case "modifier_accuracy":
			var mode string
			mode, val = parsing.PopFirstString(val, "")
			if mode == "multiply" {
				// 比例
				this.powers[inputId].ModAccuracyMode = power.STAT_MODIFIER_MODE_MULTIPLY
			} else if mode == "add" {
				this.powers[inputId].ModAccuracyMode = power.STAT_MODIFIER_MODE_ADD
			} else if mode == "absolute" {
				this.powers[inputId].ModAccuracyMode = power.STAT_MODIFIER_MODE_ABSOLUTE
			} else {
				return fmt.Errorf("PowerManager: Unknown stat_modifier_mode '%s'\n", mode)
			}

			this.powers[inputId].ModAccuracyValue, val = parsing.PopFirstInt(val, "")

		case "modifier_damage":
			var mode string
			mode, val = parsing.PopFirstString(val, "")
			if mode == "multiply" {
				// 比例
				this.powers[inputId].ModDamageMode = power.STAT_MODIFIER_MODE_MULTIPLY
			} else if mode == "add" {
				this.powers[inputId].ModDamageMode = power.STAT_MODIFIER_MODE_ADD
			} else if mode == "absolute" {
				this.powers[inputId].ModDamageMode = power.STAT_MODIFIER_MODE_ABSOLUTE
			} else {
				return fmt.Errorf("PowerManager: Unknown stat_modifier_mode '%s'\n", mode)
			}

			this.powers[inputId].ModDamageValueMin, val = parsing.PopFirstInt(val, "")
			this.powers[inputId].ModDamageValueMax, val = parsing.PopFirstInt(val, "")
		case "modifier_critical":
			var mode string
			mode, val = parsing.PopFirstString(val, "")
			if mode == "multiply" {
				// 比例
				this.powers[inputId].ModCritMode = power.STAT_MODIFIER_MODE_MULTIPLY
			} else if mode == "add" {
				this.powers[inputId].ModCritMode = power.STAT_MODIFIER_MODE_ADD
			} else if mode == "absolute" {
				this.powers[inputId].ModCritMode = power.STAT_MODIFIER_MODE_ABSOLUTE
			} else {
				return fmt.Errorf("PowerManager: Unknown stat_modifier_mode '%s'\n", mode)
			}

			this.powers[inputId].ModCritValue, val = parsing.PopFirstInt(val, "")
		case "target_movement_normal":
			this.powers[inputId].TargetMovementNormal = parsing.ToBool(val)
		case "target_movement_flying":
			this.powers[inputId].TargetMovementFlying = parsing.ToBool(val)
		case "target_movement_intangible":
			this.powers[inputId].TargetMovementIntangible = parsing.ToBool(val)
		case "walls_block_aoe":
			this.powers[inputId].WallsBlockAoe = parsing.ToBool(val)
		case "script":
			var trigger string
			trigger, val = parsing.PopFirstString(val, "")
			if trigger == "on_cast" {
				this.powers[inputId].ScriptTrigger = power.SCRIPT_TRIGGER_CAST
			} else if trigger == "on_hit" {
				this.powers[inputId].ScriptTrigger = power.SCRIPT_TRIGGER_HIT
			} else if trigger == "on_wall" {
				this.powers[inputId].ScriptTrigger = power.SCRIPT_TRIGGER_WALL
			} else {
				return fmt.Errorf("PowerManager: Unknown script trigger '%s'\n", trigger)
			}

			this.powers[inputId].Script, val = parsing.PopFirstString(val, "")

		case "remove_effect":
			var first string
			var second int
			first, val = parsing.PopFirstString(val, "")
			second, val = parsing.PopFirstInt(val, "")

			this.powers[inputId].RemoveEffects = append(this.powers[inputId].RemoveEffects, power.RemoveEffectPair{first, second})

		case "replace_by_effect":
			prbe := power.ConstructReplaceByEffect()
			prbe.PowerId, val = parsing.PopFirstInt(val, "")
			prbe.EffectId, val = parsing.PopFirstString(val, "")
			prbe.Count, val = parsing.PopFirstInt(val, "")
			this.powers[inputId].ReplaceByEffect = append(this.powers[inputId].ReplaceByEffect, prbe)
		case "requires_corpse":
			if val == "consume" {
				this.powers[inputId].RequiresCorpse = true
				this.powers[inputId].RemoveCorpse = true
			} else {
				this.powers[inputId].RequiresCorpse = parsing.ToBool(val)
				this.powers[inputId].RemoveCorpse = false
			}
		case "target_nearest":
			this.powers[inputId].TargetNearest = parsing.ToFloat(val, 0)
		case "disable_equip_slots":
			this.powers[inputId].DisableEquipSlots = nil
			var slotType string
			slotType, val = parsing.PopFirstString(val, "")

			for slotType != "" {
				this.powers[inputId].DisableEquipSlots = append(this.powers[inputId].DisableEquipSlots, slotType)
				slotType, val = parsing.PopFirstString(val, "")
			}

		default:
			return fmt.Errorf("PowerManager: '%s' is not a valid key\n", key)
		}
	}

	for _, ptr := range this.powers {
		ptr.WallPower = this.VerifyId(ptr.WallPower, true)
		ptr.PostPower = this.VerifyId(ptr.PostPower, true)

		if !((!ptr.UseHazard && ptr.Type == power.TYPE_FIXED) || ptr.NoAttack) {
			if ptr.Type == power.TYPE_FIXED {
				if ptr.RelativePos {
					// 从施法者的发出
					ptr.CombatRange += ptr.ChargeSpeed * (float32)(ptr.Lifespan)
				}

				if ptr.StartingPos == power.STARTING_POS_TARGET {
					ptr.CombatRange = math.MaxFloat32 - ptr.Radius
				}

			} else if ptr.Type == power.TYPE_MISSILE {
				ptr.CombatRange += (ptr.Speed * (float32)(ptr.Lifespan))

			} else if ptr.Type == power.TYPE_REPEATER {
				ptr.CombatRange += (ptr.Speed * (float32)(ptr.Count))
			}

			ptr.CombatRange += ptr.Radius / 2
		}

	}

	return nil
}

func (this *PowerManager) VerifyId(powerId define.PowerId, allowZero bool) define.PowerId {
	if !allowZero && powerId == 0 {
		panic(fmt.Sprintf("PowerManager: %d is not a valid power id.\n", powerId))
	}

	return powerId
}

func (this *PowerManager) getEffectDef(id string) *effect.Def {
	for i, _ := range this.effects {
		if this.effects[i].Id == id {
			return &(this.effects[i])
		}
	}

	return nil
}

func (this *PowerManager) isValidEffect(modules common.Modules, ss gameres.Stats, type1 string) bool {
	eset := modules.Eset()

	if type1 == "speed" {
		return true
	}

	if type1 == "attack_speed" {
		return true
	}

	pList := eset.Get("primary_stats", "list").([]common.PrimaryStat)

	for _, ptr := range pList {
		if type1 == ptr.GetId() {
			return true
		}
	}

	dtList := eset.Get("damage_types", "list").([]common.DamageType)
	for _, ptr := range dtList {
		if type1 == ptr.GetMin() {
			return true
		}

		if type1 == ptr.GetMax() {
			return true
		}
	}

	for i := 0; i < stats.COUNT; i++ {
		if type1 == ss.GetKey((stats.STAT)(i)) {
			return true
		}
	}

	eList := eset.Get("elements", "list").([]common.Element)

	for _, ptr := range eList {
		if type1 == ptr.GetId()+"_resist" {
			return true
		}
	}

	if this.getEffectDef(type1) != nil {
		return true
	}

	if misceffect.GetTypeFromString(modules, ss, type1) != effect.NONE {
		return true
	}

	return false
}

func (this *PowerManager) HandleNewMap(collider gameres.MapCollision) {
	this.collider = collider
}

// 某个技能对目标来说是否有效
func (this *PowerManager) HasValidTarget(powerIndex define.PowerId, srcStats gameres.StatBlock, target fpoint.FPoint) bool {
	if this.collider == nil {
		return false
	}

	// 调整目标坐标到范围内
	limitTarget := utils.ClampDistance(this.powers[powerIndex].TargetRange, srcStats.GetPos(), target)

	if !this.collider.IsEmpty(limitTarget.X, limitTarget.Y) || this.collider.IsWall(limitTarget.X, limitTarget.Y) {
		// 非空或墙
		if this.powers[powerIndex].BuffTeleport {
			// 传送技能失败
			return false
		}

	}

	return true
}

func (this *PowerManager) InitHazard(powerIndex define.PowerId, srcStats gameres.StatBlock, target fpoint.FPoint) {
	//TODO
	// harzard
}

func (this *PowerManager) Buff(modules common.Modules, ss gameres.Stats, powerIndex define.PowerId, srcStats gameres.StatBlock, target fpoint.FPoint) {

	if this.powers[powerIndex].BuffTeleport {
		// 传送技能
		limitTarget := utils.ClampDistance(this.powers[powerIndex].TargetRange, srcStats.GetPos(), target)

		if this.powers[powerIndex].TargetNeighbor > 0 {
			// 必须在目标半径内
			newTarget := this.collider.GetRandomNeighbor(modules, point.Construct(int(limitTarget.X), int(limitTarget.Y)), this.powers[powerIndex].TargetNeighbor, false)

			if math.Floor((float64)(newTarget.X)) == math.Floor((float64)(limitTarget.X)) &&
				math.Floor(float64(newTarget.Y)) == math.Floor(float64(limitTarget.Y)) {
				// 未找到合适的
				srcStats.SetTeleportation(false)
			} else {
				// 不可以在目标附近
				srcStats.SetTeleportation(true)
				srcStats.SetTeleportDestination(newTarget)
			}

		} else {
			// 只能是目标
			srcStats.SetTeleportation(true)
			srcStats.SetTeleportDestination(limitTarget)
		}
	}

	// 给自己的buf，或者是团队buf
	if this.powers[powerIndex].Buff ||
		(this.powers[powerIndex].BuffParty && (srcStats.GetHeroAlly() || srcStats.GetEnemyAlly())) {

		sourceType := power.SOURCE_TYPE_ENEMY // 来自敌人
		if srcStats.GetHero() {
			sourceType = power.SOURCE_TYPE_HERO // 来自hero
		} else if srcStats.GetHeroAlly() {
			sourceType = power.SOURCE_TYPE_ALLY // 来自盟友
		}

		// 技能引起每一个效果，更新状态
		this.effect(modules, ss, srcStats, srcStats, powerIndex, sourceType)
	}

	// 非被动的团队buf
	if this.powers[powerIndex].BuffParty && !this.powers[powerIndex].Passive {
		srcStats.AddPartyBuff(powerIndex) // 更新状态
	}

	if !this.powers[powerIndex].UseHazard {
		// 技能无伤害，清理掉效果，激活后续技能
		srcStats.GetEffects().RemoveEffectId(this.powers[powerIndex].RemoveEffects)

		if !this.powers[powerIndex].Passive {

			if tools.PercentChance(this.powers[powerIndex].PostPowerChance) {
				this.activate(modules, ss, this.powers[powerIndex].PostPower, srcStats, srcStats.GetPos())
			}
		}
	}
}

func (this *PowerManager) checkNearestTargeting(pow *power.Power, srcStats gameres.StatBlock, checkCorpses bool) bool {

	if srcStats == nil {
		return false
	}

	if pow.TargetNearest <= 0 {
		return true
	}

	if !checkCorpses && srcStats.GetTargetNearest() != nil && pow.TargetNearest > srcStats.GetTargetNearestDist() {
		// 目标不是尸体
		// 英雄的目标在技能范围内
		return true
	} else if checkCorpses && srcStats.GetTargetNearestCorpse() != nil && pow.TargetNearest > srcStats.GetTargetNearestCorpseDist() {
		// 目标是尸体
		// 英雄的目标在技能范围内
		return true
	}

	return false
}

func (this *PowerManager) payPowerCost(powerIndex define.PowerId, srcStats gameres.StatBlock) {
	if srcStats == nil {
		return
	}

	if srcStats.GetHero() {
		// 作为英雄

		// 扣蓝
		srcStats.SetMP(srcStats.GetMP() - this.powers[powerIndex].RequiresMP)

		// 扣道具
		for _, pri := range this.powers[powerIndex].RequiredItems {
			if pri.Id > 0 {

				found := false
				if pri.Equipped {
					for _, id := range this.usedEquippedItems {
						if id == pri.Id {
							found = true
							break
						}
					}

					if found {
						// 已经存在
						continue
					}
				}

				quantity := pri.Quantity
				for quantity > 0 {
					if pri.Equipped {
						this.usedEquippedItems = append(this.usedEquippedItems, pri.Id)
					} else {
						this.usedItems = append(this.usedItems, pri.Id)
					}

					quantity--
				}

			}
		}
	}

	// 技能消耗生命
	if this.powers[powerIndex].RequiresHP > 0 {
		srcStats.TakeDamage(this.powers[powerIndex].RequiresHP, false, power.SOURCE_TYPE_NEUTRAL)
	}

	// 技能消耗尸体
	if this.powers[powerIndex].RequiresCorpse &&
		this.powers[powerIndex].RemoveCorpse &&
		srcStats.GetTargetCorpse() != nil {
		srcStats.GetTargetCorpse().GetCorpseTimer().Reset(timer.END)
		srcStats.SetTargetCorpse(nil)
	}

}

// 英雄格挡
func (this *PowerManager) block(modules common.Modules, ss gameres.Stats, powerIndex define.PowerId, srcStats gameres.StatBlock) bool {
	if srcStats.GetEffects().GetTriggeredBlock() {
		// 已经格挡，返回
		return false
	}

	srcStats.GetEffects().SetTriggeredBlock(true)
	srcStats.SetBlockPower(powerIndex)

	// 格挡触发技能
	this.powers[powerIndex].PassiveTrigger = power.TRIGGER_BLOCK
	// 更新技能效果到英雄
	this.effect(modules, ss, srcStats, srcStats, powerIndex, power.SOURCE_TYPE_HERO)

	// 技能消耗
	this.payPowerCost(powerIndex, srcStats)
	return true
}

// 激活对应的技能或者功能
func (this *PowerManager) activate(modules common.Modules, ss gameres.Stats, powerIndex define.PowerId, srcStats gameres.StatBlock, target fpoint.FPoint) bool {
	if this.powers[powerIndex].IsEmpty {
		return false
	}

	// 发起者为英雄
	if srcStats.GetHero() {
		// 蓝不够
		if this.powers[powerIndex].RequiresMP > srcStats.GetMP() {
			return false
		}

		if srcStats.GetTargetCorpse() == nil &&
			srcStats.GetTargetNearestCorpse() != nil &&
			this.checkNearestTargeting(this.powers[powerIndex], srcStats, true) {
			// 最近的尸体满足技能的要求作为目标尸体
			srcStats.SetTargetCorpse(srcStats.GetTargetNearestCorpse())
		}

		if this.powers[powerIndex].RequiresCorpse && srcStats.GetTargetCorpse() == nil {
			return false
		}
	}

	if srcStats.GetHP() > 0 && this.powers[powerIndex].Sacrifice == false && this.powers[powerIndex].RequiresHP >= srcStats.GetHP() {
		return false
	}

	if this.powers[powerIndex].Type == power.TYPE_BLOCK {
		// 激活英雄格挡
		return this.block(modules, ss, powerIndex, srcStats)
	}

	if this.powers[powerIndex].ScriptTrigger == power.SCRIPT_TRIGGER_CAST {
		// TODO
		// 执行脚本
	}

	newTarget := target
	if this.powers[powerIndex].LockTargetToDirection {
		// 从8个方向移动

		dist := utils.CalcDist(srcStats.GetPos(), newTarget)
		dir := utils.CalcDirection(srcStats.GetPos().X, srcStats.GetPos().Y, newTarget.X, newTarget.Y)
		newTarget = utils.CalcVector(srcStats.GetPos(), dir, dist)
	}

	switch this.powers[powerIndex].Type {
	case power.TYPE_FIXED:
		// 固定
		// TODO
	case power.TYPE_MISSILE:
		// 飞弹
		// TODO
	case power.TYPE_REPEATER:
		// TODO
		// 一条线上多次伤害
	case power.TYPE_SPAWN:
		// TODO
		// 召唤
	case power.TYPE_TRANSFORM:
		// TODO
		// 变身

	}

	return false
}

// 技能引起其定义的一堆效果, 同步技能效果到对应的对象
func (this *PowerManager) effect(modules common.Modules, ss gameres.Stats, targetStats gameres.StatBlock, casterStats gameres.StatBlock, powerIndex define.PowerId, sourceType int) bool {
	eset := modules.Eset()

	dtList := eset.Get("damage_types", "list").([]common.DamageType)

	pwr := this.powers[powerIndex]

	for i := 0; i < len(pwr.PostEffects); i++ {
		pe := &(pwr.PostEffects[i])

		// 触发后续效果
		if !tools.PercentChance(pe.Chance) {
			continue
		}

		effectPtr := this.getEffectDef(pe.Id)
		magnitude := pe.Magnitude
		duration := pe.Duration

		destStats := targetStats
		if pe.TargetSrc {
			// 效果作用到施法者
			destStats = casterStats
		}

		if destStats.GetHP() <= 0 && pe.Id != "revive" {
			// 不是复活没意义
			continue
		}

		effectData := effect.ConstructDef()
		if effectPtr != nil {
			effectData = (*effectPtr)

			switch effectData.Type {
			case effect.SHIELD:

				// 护盾
				if pwr.BaseDamage == len(dtList) {
					continue
				}

				if pwr.ModDamageMode == power.STAT_MODIFIER_MODE_MULTIPLY {
					// 比例
					magnitude = casterStats.GetDamageMax(pwr.BaseDamage) * pwr.ModDamageValueMin / 100
				} else if pwr.ModDamageMode == power.STAT_MODIFIER_MODE_ADD {
					magnitude = casterStats.GetDamageMax(pwr.BaseDamage) + pwr.ModDamageValueMin
				} else if pwr.ModDamageMode == power.STAT_MODIFIER_MODE_ABSOLUTE {
					magnitude = tools.RandBetween(pwr.ModDamageValueMin, pwr.ModDamageValueMax)
				} else {
					// 跟一般伤害一致
					magnitude = casterStats.GetDamageMax(pwr.BaseDamage)
				}

				// TODO
				// comb

			case effect.HEAL:

				// 治疗
				if pwr.BaseDamage == len(dtList) {
					continue
				}
				magnitude = tools.RandBetween(casterStats.GetDamageMax(pwr.BaseDamage), casterStats.GetDamageMin(pwr.BaseDamage))

				if pwr.ModDamageMode == power.STAT_MODIFIER_MODE_MULTIPLY {
					// 比例
					magnitude = magnitude * pwr.ModDamageValueMin / 100
				} else if pwr.ModDamageMode == power.STAT_MODIFIER_MODE_ADD {
					magnitude += pwr.ModDamageValueMin
				} else if pwr.ModDamageMode == power.STAT_MODIFIER_MODE_ABSOLUTE {
					magnitude = tools.RandBetween(pwr.ModDamageValueMin, pwr.ModDamageValueMax)
				}

				// TODO
				// comb

				destStats.SetHP(destStats.GetHP() + magnitude)
				if destStats.GetHP() > destStats.Get(stats.HP_MAX) {
					destStats.SetHP(destStats.Get(stats.HP_MAX))
				}

			case effect.KNOCKBACK:
				if destStats.GetSpeedDefault() == 0 {
					// 敌人无速度，故不可击退
					continue
				}

				if pe.TargetSrc {
					// 作用给施法者
					destStats.SetKnockbackSrcPos(targetStats.GetPos())
					destStats.SetKnockbackDestPos(casterStats.GetPos())
				} else {
					destStats.SetKnockbackSrcPos(casterStats.GetPos())
					destStats.SetKnockbackDestPos(targetStats.GetPos())
				}

			}
		} else {
			effectData.Id = pe.Id
			effectData.Type = misceffect.GetTypeFromString(modules, ss, pe.Id)
		}

		destStats.GetEffects().AddEffect(modules, effectData, duration, magnitude, sourceType, powerIndex)
	}

	return true
}

func (this *PowerManager) GetPower(powerId define.PowerId) *power.Power {
	return this.powers[powerId]
}

func (this *PowerManager) GetPowers() map[define.PowerId]*power.Power {
	return this.powers
}
