package statblock

import (
	"fmt"
	"math"
	"monster/pkg/common"
	"monster/pkg/common/define"
	"monster/pkg/common/define/game/mapcollision"
	"monster/pkg/common/define/game/stats"
	"monster/pkg/common/event"
	"monster/pkg/common/fpoint"
	"monster/pkg/common/gameres"
	"monster/pkg/common/gameres/effect"
	"monster/pkg/common/gameres/power"
	"monster/pkg/common/gameres/statblock"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/common/timer"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/utils"
	"monster/pkg/utils/parsing"
)

type StatBlock struct {
	statsLoaded        bool
	alive              bool // 是否活的
	corpse             bool // 是否尸体和结束的动画
	corpseTimer        *timer.Timer
	hero               bool // 是否作为主角
	heroAlly           bool // 是主角的盟友
	enemyAlly          bool
	npc                bool // 是否作为npc
	humanoid           bool // 是否作为人
	lifeform           bool // 是否有生命
	permadeath         bool // 是否永久死亡
	transformed        bool // 是否处于变身状态
	refreshStats       bool // 是否刷新过状态
	summoned           bool // 是否召唤的
	converted          bool // 是转变成盟友的怪物
	summonedPowerIndex define.PowerId
	encountered        bool //  作为敌人，是否和英雄遭遇

	// hero专属
	targetCorpse            gameres.StatBlock
	targetNearest           gameres.StatBlock
	targetNearestCorpse     gameres.StatBlock
	targetNearestDist       float32
	targetNearestCorpseDist float32
	blockPower              define.PowerId // 格挡技能id

	movementType        int      // 移动类型
	flying              bool     // 是否能飞行
	intangible          bool     // 是否透明
	facing              bool     // 是否可以转向面向目标
	categories          []string // 本实体属于的分类
	name                string
	level               int
	xp                  uint64 // 作为怪物被打死后获得的经验
	checkTitle          bool
	statPointsPerLevel  int // 每个等级的获得属性点
	powerPointsPerLevel int // 每个等级的获得的技能点

	// 基础属性只是用来加点，来影响子属性
	primary         []int // 基础属性
	primaryStarting []int

	// 通过基础属性加点和等级提升，子属性提升，才带来提升

	// 战斗属性
	starting          []int   // 各属性点和伤害的起始计算值, 偶数为min，奇数为max
	base              []int   // 加成效果生效前
	current           []int   // 加成效果生效后
	perLevel          []int   // 玩家每次升级的子属性的增量
	perPrimary        [][]int // 基础属性加点对子属性的增量
	primaryAdditional []int   // 基础属性额外加成，效果加成
	characterClass    string
	characterSubclass string
	hp                int
	hpF               float32
	mp                int
	mpF               float32
	speedDefault      float32
	dmgMinAdd         []int
	dmgMaxAdd         []int
	absorbMinAdd      int
	absorbMaxAdd      int
	speed             float32 // 移动速度
	chargeSpeed       float32
	vulnerable        []int // 对某种元素的受伤程度
	vulnerableBase    []int

	// buff
	transformDuration int // buf持续时间

	effects            gameres.EffectManager // 效果加成管理器
	blocking           bool
	pos                fpoint.FPoint
	knockbackSpeed     fpoint.FPoint
	knockbackSrcPos    fpoint.FPoint
	knockbackDestPos   fpoint.FPoint
	direction          uint8        // 当前移动方向
	cooldownHit        *timer.Timer // 被击中的间隔
	cooldownHitEnabled bool
	curState           statblock.EntityState // 当前该对象的状态
	stateTimer         timer.Timer
	waypoints          []fpoint.FPoint
	waypointTimer      timer.Timer // 在每个路点的等待
	wanderArea         rect.Rect

	chancePursue    int // 追逐目标的概率
	chanceFlee      int // 逃跑的概率
	powersList      []define.PowerId
	powersListItems []define.PowerId
	powersPassive   []define.PowerId    // 被动技能
	powersAI        []statblock.AIPower // 特定情况触发的技能

	meleeRange     float32 // 攻击范围
	threatRange    float32 // 开始追逐玩家半径
	threatRangeFar float32 // 停止追逐玩家的半径
	fleeRange      float32 // 逃跑距离

	combatStyle   statblock.CombatStyle
	turnDelay     int          // 转向目标的延迟，以帧为单位
	inCombat      bool         // 是否正在战斗
	cooldown      *timer.Timer // 攻击间隔
	halfDeadPower bool         // 是否有半血触发的技能
	suppressHP    bool         // 是否隐藏血条

	fleeTimer         timer.Timer // 逃跑时间
	fleeCooldownTimer timer.Timer // 再次逃跑的冷却时间
	perfectAccuracy   bool
	lootTable         []event.Component // 该对象死亡时掉落的物品
	lootCount         point.Point       // 掉了物品的最大和最小数量

	teleportation       bool
	teleportDestination fpoint.FPoint

	currency                   int
	defeatStatus               define.StatusId // 死亡时设置的属性
	convertStatus              define.StatusId // 被玩家转化成盟友时的属性
	questLootRequiresStatus    define.StatusId
	questLootRequiresNotStatus define.StatusId
	questLootId                define.ItemId
	firstDefeatLoot            define.ItemId // 首次被击杀掉落物品

	gfxBase     string // /images/avatar中的文件
	gfxHead     string // png in /images/avatar/[base]
	gfxPortrait string // png in /images/portraits
	animations  string // 动画定义文件
	sfxStep     int

	maxSpendableStatPoints int
	maxPointsPerStat       int // 每个基础属性最大点数
	prevMaxHP              int
	prevMaxMP              int
	prevHP                 int
	prevMP                 int

	partyBuffs  []define.PowerId // 团队buf
	powerFilter []define.PowerId // 免疫
}

func New(modules common.Modules, gresf common.Factory) *StatBlock {
	sb := &StatBlock{}
	sb.Init(modules, gresf)

	return sb
}

func (this *StatBlock) Init(modules common.Modules, gresf gameres.Factory) gameres.StatBlock {
	eset := modules.Eset()
	settings := modules.Settings()

	this.alive = true
	this.corpseTimer = timer.New()
	this.lifeform = true
	this.summoned = false
	this.movementType = mapcollision.MOVE_NORMAL
	this.facing = true
	this.statPointsPerLevel = 1
	this.powerPointsPerLevel = 1
	this.starting = make([]int, stats.COUNT+eset.Get("damage_types", "count").(int))
	this.base = make([]int, stats.COUNT+eset.Get("damage_types", "count").(int))
	this.current = make([]int, stats.COUNT+eset.Get("damage_types", "count").(int))
	this.perLevel = make([]int, stats.COUNT+eset.Get("damage_types", "count").(int))
	this.speedDefault = 0.1
	dtList := eset.Get("damage_types", "list").([]common.DamageType)
	this.dmgMinAdd = make([]int, len(dtList))
	this.dmgMaxAdd = make([]int, len(dtList))
	this.speed = 0.1
	eList := eset.Get("elements", "list").([]common.Element)
	this.vulnerable = make([]int, len(eList))
	this.vulnerableBase = make([]int, len(eList))
	for i := 0; i < len(eList); i++ {
		this.vulnerable[i] = 100
		this.vulnerableBase[i] = 100
	}
	this.effects = gresf.New("effectmanager").(gameres.EffectManager).Init(modules)
	this.pos = fpoint.Construct()
	this.knockbackSpeed = fpoint.Construct()
	this.knockbackSrcPos = fpoint.Construct()
	this.knockbackDestPos = fpoint.Construct()
	this.cooldownHit = timer.New()
	this.curState = statblock.ENTITY_STANCE
	this.stateTimer = timer.Construct()
	this.waypointTimer = timer.Construct((uint)(settings.Get("max_fps").(int)))
	this.wanderArea = rect.Construct()
	this.meleeRange = 1
	this.combatStyle = statblock.COMBAT_DEFAULT
	this.cooldown = timer.New()
	this.fleeTimer = timer.Construct((uint)(settings.Get("max_fps").(int)))
	this.fleeCooldownTimer = timer.Construct((uint)(settings.Get("max_fps").(int)))
	this.lootCount = point.Construct()
	this.teleportDestination = fpoint.Construct()
	this.gfxBase = "male"
	this.gfxHead = "head_short"
	this.gfxPortrait = ""

	// 就4个，物理，魔法...
	psList := eset.Get("primary_stats", "list").([]common.PrimaryStat)
	this.primary = make([]int, len(psList))
	this.primaryStarting = make([]int, len(psList))
	this.primaryAdditional = make([]int, len(psList))
	this.perPrimary = make([][]int, len(psList))
	for i, _ := range this.perPrimary {
		this.perPrimary[i] = make([]int, stats.COUNT+eset.Get("damage_types", "count").(int))
	}

	this.cooldown.Reset(timer.END)
	return this
}

func (this *StatBlock) Close() {
	if this.effects != nil {
		this.effects.Close()
		this.effects = nil
	}
}

func (this *StatBlock) GetDirection() uint8 {
	return this.direction
}

func (this *StatBlock) SetDirection(val uint8) {
	this.direction = val
}

func (this *StatBlock) Get(stat stats.STAT) int {
	return this.current[stat]
}

// 当前某种伤害值
func (this *StatBlock) GetDamageMin(dmgType int) int {
	return this.current[stats.COUNT+dmgType*2]
}

func (this *StatBlock) GetDamageMax(dmgType int) int {
	return this.current[stats.COUNT+dmgType*2+1]
}

// 核心通用数据
func (this *StatBlock) loadCoreStat(modules common.Modules, ss gameres.Stats, key, val string) (bool, error) {
	settings := modules.Settings()
	eset := modules.Eset()

	switch key {
	case "speed":
		// 移动速度
		fValue := parsing.ToFloat(val, 0)
		this.speed = fValue / (float32)(settings.Get("max_fps").(int))
		return true, nil
	case "cooldown":
		// 攻击间隔
		this.cooldown.SetDuration((uint)(parsing.ToDuration(val, settings.Get("max_fps").(int))))
		return true, nil
	case "cooldown_hit":
		this.cooldownHit.SetDuration((uint)(parsing.ToDuration(val, settings.Get("max_fps").(int))))
		this.cooldownHitEnabled = true
		return true, nil
	case "stat":
		// 属性的初始值
		stat := ""
		value := 0
		stat, val = parsing.PopFirstString(val, "")
		value, val = parsing.PopFirstInt(val, "")

		for i := 0; i < stats.COUNT; i++ {
			if ss.GetKey((stats.STAT)(i)) == stat {
				this.starting[i] = value
				return true, nil
			}
		}

		dtList := eset.Get("damage_types", "list").([]common.DamageType)
		for i := 0; i < len(dtList); i++ {
			if dtList[i].GetMin() == stat {
				this.starting[stats.COUNT+i*2] = value
				return true, nil
			} else if dtList[i].GetMax() == stat {
				this.starting[stats.COUNT+i*2+1] = value
				return true, nil
			}
		}
	case "stat_per_level":
		// 每次玩家等级提升，子属性的都会提升
		stat := ""
		value := 0
		stat, val = parsing.PopFirstString(val, "")
		value, val = parsing.PopFirstInt(val, "")

		for i := 0; i < stats.COUNT; i++ {
			if ss.GetKey((stats.STAT)(i)) == stat {
				this.perLevel[i] = value
				return true, nil
			}
		}

		dtList := eset.Get("damage_types", "list").([]common.DamageType)
		for i := 0; i < len(dtList); i++ {
			if dtList[i].GetMin() == stat {
				this.perLevel[stats.COUNT+i*2] = value
				return true, nil
			} else if dtList[i].GetMax() == stat {
				this.perLevel[stats.COUNT+i*2+1] = value
				return true, nil
			}
		}

	case "stat_per_primary":
		// 这里是主属性点影响子属性的增量
		primStat := ""
		primStat, val = parsing.PopFirstString(val, "")
		primStatIndex, ok := eset.PrimaryStatsGetIndexById(primStat)
		if !ok {
			return false, fmt.Errorf("StatBlock: '%s' is not a valid primary stat.\n", primStat)
		}

		stat := ""
		value := 0
		stat, val = parsing.PopFirstString(val, "")
		value, val = parsing.PopFirstInt(val, "")

		for i := 0; i < stats.COUNT; i++ {
			if ss.GetKey((stats.STAT)(i)) == stat {
				this.perPrimary[primStatIndex][i] = value
				return true, nil
			}
		}

		dtList := eset.Get("damage_types", "list").([]common.DamageType)
		for i := 0; i < len(dtList); i++ {
			if dtList[i].GetMin() == stat {
				this.perPrimary[primStatIndex][stats.COUNT+i*2] = value
				return true, nil
			} else if dtList[i].GetMax() == stat {
				this.perPrimary[primStatIndex][stats.COUNT+i*2+1] = value
				return true, nil
			}
		}

	case "vulnerable":
		element := ""
		value := 0
		element, val = parsing.PopFirstString(val, "")
		value, val = parsing.PopFirstInt(val, "")

		eList := eset.Get("elements", "list").([]common.Element)

		for i := 0; i < len(eList); i++ {
			if element == eList[i].GetId() {
				this.vulnerable[i] = value
				this.vulnerableBase[i] = value
				return true, nil
			}
		}
	case "power_filter":
		powerId := ""
		powerId, val = parsing.PopFirstString(val, "")
		for powerId != "" {
			this.powerFilter = append(this.powerFilter, parsing.ToPowerId(powerId, 0))
			powerId, val = parsing.PopFirstString(val, "")
		}
		return true, nil

	case "categories":
		this.categories = nil
		cat := ""
		cat, val = parsing.PopFirstString(val, "")
		for cat != "" {
			this.categories = append(this.categories, cat)
			cat, val = parsing.PopFirstString(val, "")
		}
		return true, nil
	case "melee_range":
		this.meleeRange = parsing.ToFloat(val, 0)
		return true, nil
	}

	return false, nil
}

func (this *StatBlock) loadSfxStat() {
	// TODO
}

func (this *StatBlock) isNPCStat(infile *fileparser.FileParser) bool {
	if infile.GetSection() == "npc" {
		return true
	} else if infile.GetSection() == "dialog" {
		return true
	}

	switch infile.Key() {
	case "gfx":
		fmt.Println("StatBlock: Warning! 'gfx' is deprecated. Use 'animations' instead.")
		this.animations = infile.Val()
		return true
	case "direction":
		fallthrough
	case "talker":
		fallthrough
	case "portrait":
		fallthrough
	case "vendor":
		fallthrough
	case "vendor_requires_status":
		fallthrough
	case "vendor_requires_not_status":
		fallthrough
	case "constant_stock":
		// 库存
		fallthrough
	case "status_stock":
		fallthrough
	case "random_stock":
		fallthrough
	case "random_stock_count":
		fallthrough
	case "vox_intro":
		return true
	}

	return false
}

// 加载一个状态块，其实就是一个定义和计算内核，一般作为敌人的定义
func (this *StatBlock) Load(modules common.Modules, ss gameres.Stats, loot gameres.LootManager, camp gameres.CampaignManager, powers gameres.PowerManager, filename string) error {
	mods := modules.Mods()
	msg := modules.Msg()
	settings := modules.Settings()

	infile := fileparser.New()
	err := infile.Open(filename, true, mods)
	if err != nil {
		return err
	}
	defer infile.Close()

	clearLoot := true
	fleeRangeDefined := false

	for infile.Next(mods) {
		if infile.IsNewSection() && (infile.GetSection() == "" || infile.GetSection() == "stats") {
			clearLoot = true
		}

		key := infile.Key()
		val := infile.Val()

		num := parsing.ToInt(val, 0)
		fnum := parsing.ToFloat(val, 0)
		valid, err := this.loadCoreStat(modules, ss, key, val)
		if err != nil {
			return err
		}

		valid = valid || this.isNPCStat(infile)

		switch key {
		case "name":
			this.name = msg.Get(val)
		case "humanoid":
			this.humanoid = parsing.ToBool(val)
		case "lifeform":
			this.lifeform = parsing.ToBool(val)
		case "level":
			this.level = num
		case "xp":
			this.xp = (uint64)(num)
		case "loot":
			if clearLoot {
				this.lootTable = nil
				clearLoot = false
			}

			this.lootTable = append(this.lootTable, event.ConstructComponent())

			// 掉落物品加载
			this.lootTable = loot.ParseLoot(modules, val, &(this.lootTable[len(this.lootTable)-1]), this.lootTable)

		case "loot_count":
			this.lootCount.X, val = parsing.PopFirstInt(val, "")
			this.lootCount.Y, val = parsing.PopFirstInt(val, "")
			if this.lootCount.X != 0 || this.lootCount.Y != 0 {
				this.lootCount.X = (int)(math.Max(float64(this.lootCount.X), 1))
				this.lootCount.Y = (int)(math.Max(float64(this.lootCount.Y), float64(this.lootCount.Y)))
			}

		case "defeat_status":
			this.defeatStatus = camp.RegisterStatus(val)
		case "convert_status":
			this.convertStatus = camp.RegisterStatus(val)
		case "first_defeat_loot":
			this.firstDefeatLoot = (define.ItemId)(parsing.ToInt(val, 0))
		case "quest_loot":

			// 注册属性
			var first string
			first, val = parsing.PopFirstString(val, "")
			this.questLootRequiresStatus = camp.RegisterStatus(first)
			first, val = parsing.PopFirstString(val, "")
			this.questLootRequiresNotStatus = camp.RegisterStatus(first)
			first, val = parsing.PopFirstString(val, "")
			this.questLootId = (define.ItemId)(parsing.ToInt(val, 0))
		case "flying":
			this.flying = parsing.ToBool(val)
		case "intangible":
			this.intangible = parsing.ToBool(val)
		case "facing":
			this.facing = parsing.ToBool(val)
		case "waypoint_pause":
			this.waypointTimer.SetDuration((uint)(parsing.ToDuration(val, settings.Get("max_fps").(int))))
		case "turn_delay":
			this.turnDelay = parsing.ToDuration(val, settings.Get("max_fps").(int))
		case "chance_pursue":
			this.chancePursue = num
		case "chance_flee":
			this.chanceFlee = num
		case "power":
			aiPower := statblock.ConstructAIPower()
			var aiType, first string
			aiType, val = parsing.PopFirstString(val, "")
			first, val = parsing.PopFirstString(val, "")
			aiPower.Id = powers.VerifyId(define.PowerId(parsing.ToInt(first, 0)), false)
			if aiPower.Id == 0 {
				continue
			}

			aiPower.Chance, val = parsing.PopFirstInt(val, "")

			if aiType == "melee" {
				aiPower.Type = statblock.AI_POWER_MELEE
			} else if aiType == "ranged" {
				aiPower.Type = statblock.AI_POWER_RANGED
			} else if aiType == "beacon" {
				aiPower.Type = statblock.AI_POWER_BEACON
			} else if aiType == "on_hit" {
				aiPower.Type = statblock.AI_POWER_HIT
			} else if aiType == "on_death" {
				aiPower.Type = statblock.AI_POWER_DEATH
			} else if aiType == "on_half_dead" {
				aiPower.Type = statblock.AI_POWER_HALF_DEAD
			} else if aiType == "on_join_combat" {
				aiPower.Type = statblock.AI_POWER_JOIN_COMBAT
			} else if aiType == "on_debuff" {
				aiPower.Type = statblock.AI_POWER_DEBUFF
			} else {
				return fmt.Errorf("StatBlock: '%s' is not a valid enemy power type.\n", aiType)
			}

			if aiPower.Type == statblock.AI_POWER_HALF_DEAD {
				this.halfDeadPower = true
			}

			this.powersAI = append(this.powersAI, aiPower)
		case "passive_powers":
			this.powersPassive = nil
			var p string
			p, val = parsing.PopFirstString(val, "")
			for p != "" {
				this.powersPassive = append(this.powersPassive, (define.PowerId)(parsing.ToInt(p, 0)))
				p, val = parsing.PopFirstString(val, "")

				// 被动技有后续技能
				postPower := powers.GetPower(this.powersPassive[len(this.powersPassive)-1]).PostPower
				if postPower > 0 {
					passivePostPower := statblock.ConstructAIPower()
					passivePostPower.Type = statblock.AI_POWER_PASSIVE_POST
					passivePostPower.Id = postPower
					passivePostPower.Chance = 0
					this.powersAI = append(this.powersAI, passivePostPower)
				}

			}
		case "threat_range":
			var first string
			first, val = parsing.PopFirstString(val, "")
			this.threatRange = parsing.ToFloat(first, 0)
			first, val = parsing.PopFirstString(val, "")
			if first != "" {
				this.threatRangeFar = parsing.ToFloat(first, 0)
			} else {
				this.threatRangeFar = this.threatRange * 2
			}

		case "flee_range":
			this.fleeRange = fnum
			fleeRangeDefined = true
		case "combat_style":
			if val == "default" {
				this.combatStyle = statblock.COMBAT_DEFAULT
			} else if val == "aggressive" {
				this.combatStyle = statblock.COMBAT_AGGRESSIVE
			} else if val == "passive" {
				this.combatStyle = statblock.COMBAT_PASSIVE
			} else {
				return fmt.Errorf("StatBlock: Unknown combat style '%s'\n", val)
			}
		case "animations":
			this.animations = val
		case "suppress_hp":
			this.suppressHP = parsing.ToBool(val)
		case "flee_duration":
			this.fleeTimer.SetDuration(uint(parsing.ToDuration(val, settings.Get("max_fps").(int))))
		case "flee_cooldown":
			this.fleeCooldownTimer.SetDuration(uint(parsing.ToDuration(val, settings.Get("max_fps").(int))))
		case "rarity":
			// PASS
		default:
			if !valid {
			}
		}

	}

	this.hp = this.starting[stats.HP_MAX]
	this.mp = this.starting[stats.MP_MAX]

	if !fleeRangeDefined {
		// 给默认值
		this.fleeRange = this.threatRange / 2
	}

	this.applyEffects(modules)

	return nil
}

// 主角自己的属性
func (this *StatBlock) loadHeroStats(modules common.Modules, ss gameres.Stats) error {
	settings := modules.Settings()
	mods := modules.Mods()
	eset := modules.Eset()

	// 66ms转成经历的总帧数
	this.cooldown.SetDuration((uint)(parsing.ToDuration("66ms", settings.Get("max_fps").(int))))

	infile := fileparser.New()
	err := infile.Open("engine/stats.txt", true, mods)
	if err != nil {
		return err
	}
	defer infile.Close()

	for infile.Next(mods) {

		// 核心通用数据
		ok, err := this.loadCoreStat(modules, ss, infile.Key(), infile.Val())
		if err != nil {
			return err
		}

		if !ok {
			value := parsing.ToInt(infile.Val(), 0)
			switch infile.Key() {
			case "max_points_per_stat":
				this.maxPointsPerStat = value
			case "sfx_step":
				this.sfxStep = value
			case "stat_points_per_level":
				this.statPointsPerLevel = value
			case "power_points_per_level":
				this.powerPointsPerLevel = value
			default:
				return fmt.Errorf("StatBlock: '%s' is not a valid key.\n", infile.Key())
			}
		}
	}

	if this.maxPointsPerStat == 0 {
		this.maxPointsPerStat = this.maxSpendableStatPoints/4 + 1
	}
	this.maxSpendableStatPoints = eset.XPGetMaxLevel() * this.statPointsPerLevel

	// done
	this.statsLoaded = true
	return nil
}

// 计算子属性的初始值
func (this *StatBlock) calcBase(modules common.Modules) {
	eset := modules.Eset()

	// 等级1无加成
	lev0 := (int)(math.Max((float64)(this.level-1), 0))

	// 1. 基础属性加点对每个子属性进行加成
	if len(this.perPrimary) == 0 {
		// 无子属性加成
		for i := 0; i < stats.COUNT+eset.Get("damage_types", "count").(int); i++ {
			this.base[i] = this.starting[i] + lev0*this.perLevel[i]
		}
	} else {
		// 有子属性加成
		for j := 0; j < len(this.perPrimary); j++ {
			// 每个基础属性，基础属性为1无加成，故只对大于1的部分进行加成
			currentPrimary := (int)(math.Max((float64)(this.GetPrimary(j)-1), 0))

			// 获得某一类基础属性下的子属性加成
			perPrimaryVec := this.perPrimary[j]
			for i := 0; i < stats.COUNT+eset.Get("damage_types", "count").(int); i++ {
				if j == 0 {
					// 物理还要加上等级增长
					this.base[i] = this.starting[i] + lev0*this.perLevel[i]
				}

				this.base[i] += (currentPrimary * perPrimaryVec[i])
			}

		}
	}

	// 2. 装备对伤害属性的加成
	dtList := eset.Get("damage_types", "list").([]common.DamageType)
	for i := 0; i < len(dtList); i++ {
		this.base[stats.COUNT+i*2] += this.dmgMinAdd[i]
		this.base[stats.COUNT+i*2+1] += this.dmgMaxAdd[i]

		// 最小
		this.base[stats.COUNT+i*2] = (int)(math.Max((float64)(this.base[stats.COUNT+i*2]), 0))
		// 最大
		this.base[stats.COUNT+i*2+1] = (int)(math.Max((float64)(this.base[stats.COUNT+i*2+1]), (float64)(this.base[stats.COUNT+i*2])))
	}

	// 伤害吸收
	this.base[stats.ABS_MIN] += this.absorbMinAdd
	this.base[stats.ABS_MAX] += this.absorbMaxAdd

	this.base[stats.ABS_MIN] = (int)(math.Max(float64(this.base[stats.ABS_MIN]), 0))
	this.base[stats.ABS_MAX] = (int)(math.Max(float64(this.base[stats.ABS_MAX]), float64(this.base[stats.ABS_MIN])))
}

// 计算 本身的属性 + 效果加成
func (this *StatBlock) applyEffects(modules common.Modules) {
	eset := modules.Eset()

	this.prevMaxHP = (int)(math.Max((float64)(this.Get(stats.HP_MAX)), 1))
	this.prevMaxHP = (int)(math.Max((float64)(this.Get(stats.MP_MAX)), 1))
	this.prevHP = this.hp
	this.prevMP = this.mp

	// 计算基础属性
	for i := 0; i < len(this.primary); i++ {

		// 基础属性 + 基础属性加成
		if this.GetPrimary(i) != this.primary[i]+this.effects.GetBonusPrimary(i) {

			// 加成发生变化需要刷新角色菜单
			this.refreshStats = true
		}

		// 更新
		this.primaryAdditional[i] = this.effects.GetBonusPrimary(i)
	}

	// 计算子属性的初始值
	this.calcBase(modules)

	for i := 0; i < stats.COUNT+eset.Get("damage_types", "count").(int); i++ {
		this.current[i] = this.base[i] + this.effects.GetBonus(i)
	}

	for i := 0; i < len(this.vulnerable); i++ {
		this.vulnerable[i] = this.vulnerableBase[i] - this.effects.GetBonusResist(i)
	}

	this.current[stats.HP_MAX] += this.Get(stats.HP_MAX) * this.Get(stats.HP_PERCENT) / 100
	this.current[stats.MP_MAX] += this.Get(stats.MP_MAX) * this.Get(stats.MP_PERCENT) / 100
	this.current[stats.HP_MAX] = (int)(math.Max(float64(this.Get(stats.HP_MAX)), 1))
	this.current[stats.MP_MAX] = (int)(math.Max(float64(this.Get(stats.MP_MAX)), 1))

	if this.hp > this.Get(stats.HP_MAX) {
		this.hp = this.Get(stats.HP_MAX)
	}

	if this.mp > this.Get(stats.MP_MAX) {
		this.mp = this.Get(stats.MP_MAX)
	}

	this.speed = this.speedDefault
}

// 本实体吃伤害
func (this *StatBlock) TakeDamage(dmg int, crit bool, sourceType int) {
	this.hp -= this.effects.DamageShields(dmg) // 护盾效果吸收了多少

	if this.hp <= 0 {
		// 处理没血了

		this.hp = 0

		this.effects.SetTriggeredDeath(true)

		if this.hero {
			// 作为英雄
			this.curState = statblock.ENTITY_DEAD
		} else {
			// 敌人
			if !this.heroAlly || this.converted {
				//TODO

				// 掉宝
				// loot
				// xp
			}

			if crit {
				// 暴击
				this.curState = statblock.ENTITY_CRITDEAD
			} else {
				this.curState = statblock.ENTITY_DEAD
			}

			//TODO
			// collider
		}
	}
}

// 重新计算英雄的等级和状态
func (this *StatBlock) Recalc(modules common.Modules, ss gameres.Stats) {
	eset := modules.Eset()

	if this.hero {
		if !this.statsLoaded {
			// 不重复加载主角数据和核心通用数据
			this.loadHeroStats(modules, ss)
		}

		this.refreshStats = true

		xpMax := eset.XPGetLevelXP(eset.XPGetMaxLevel())
		this.xp = (uint64)(math.Min(float64(this.xp), float64(xpMax)))
		this.level = eset.XPGetLevelFromXP(this.xp)
		if this.level != 0 {
			this.checkTitle = true
		}
	}

	if this.level < 1 {
		this.level = 1
	}

	this.applyEffects(modules)

	this.hp = this.Get(stats.HP_MAX)
	this.mp = this.Get(stats.MP_MAX)
}

// 计算自己的状态
func (this *StatBlock) Logic(modules common.Modules, pc gameres.Avatar, camp gameres.CampaignManager) {
	settings := modules.Settings()

	this.alive = !(this.hp <= 0 && !this.effects.GetTriggeredDeath() && !this.effects.GetRevive())

	// TODO
	// 团队buff
	// entitym

	// 各种伤害，回血，免疫效果的计算，动画往下播放
	this.effects.Logic(modules)

	// 把上面效果的计算值加到本状态
	this.applyEffects(modules)

	// 英雄身上的效果数量有增减，需要刷新
	if this.hero && this.effects.GetRefreshStats() {
		this.refreshStats = true
		this.effects.SetRefreshStats(false)
	}

	// 生命最大值变化，等比例缩放当前血量
	if this.prevMaxHP != this.Get(stats.HP_MAX) {
		ratio := float32(this.prevHP) / float32(this.prevMaxHP)
		this.hp = (int)(ratio * float32(this.Get(stats.HP_MAX)))
	}

	if this.prevMaxMP != this.Get(stats.MP_MAX) {
		ratio := float32(this.prevMP) / float32(this.prevMaxMP)
		this.mp = (int)(ratio * float32(this.Get(stats.MP_MAX)))
	}

	this.cooldown.Tick()

	for i, _ := range this.powersAI {
		this.powersAI[i].Cooldown.Tick()
	}

	if (int)(this.hpF) != this.hp {
		this.hpF = (float32)(this.hp)
	}

	if (int)(this.mpF) != this.mp {
		this.mpF = (float32)(this.mp)
	}

	// 自动回血
	if this.hp <= this.Get(stats.HP_MAX) && this.hp > 0 {

		var hpRegenPerFrame float32

		if !this.inCombat && !this.heroAlly && !this.hero && pc.GetStats().GetAlive() {
			// 作为小怪
			hpRegenPerFrame = float32(this.Get(stats.HP_MAX)) / 5 / (float32)(settings.Get("max_fps").(int))
		} else {
			// 作为英雄
			hpRegenPerFrame = float32(this.Get(stats.HP_REGEN)) / 60 / (float32)(settings.Get("max_fps").(int))
		}

		this.hpF += hpRegenPerFrame
		this.hp = int(math.Max(0, math.Min(float64(this.hpF), (float64)(this.Get(stats.HP_MAX)))))
	}

	// 自动回蓝
	if this.mp <= this.Get(stats.MP_MAX) && this.hp > 0 {
		mpRegenPerFrame := float32(this.Get(stats.MP_REGEN)) / 60 / (float32)(settings.Get("max_fps").(int))

		this.mpF += mpRegenPerFrame
		this.mp = int(math.Max(0, math.Min(float64(this.mpF), (float64)(this.Get(stats.MP_MAX)))))
	}

	if this.transformDuration > 0 {
		this.transformDuration--
	}

	// 效果造成的伤害
	if this.effects.GetDamage() > 0 && this.hp > 0 {
		this.TakeDamage(this.effects.GetDamage(), false, this.effects.GetDamageSourceType(effect.DAMAGE))
		//TODO
		//comb
	}

	if this.effects.GetDamagePercent() > 0 && this.hp > 0 {
		damage := this.Get(stats.HP_MAX) * this.effects.GetDamagePercent() / 100
		this.TakeDamage(damage, false, this.effects.GetDamageSourceType(effect.DAMAGE_PERCENT))
		//TODO
		//comb
	}

	if this.effects.GetDeathSentence() {
		this.TakeDamage(this.Get(stats.HP_MAX), false, power.SOURCE_TYPE_NEUTRAL)
	}

	this.cooldownHit.Tick()

	// 晕倒
	if this.effects.GetStun() {
		this.stateTimer.Reset(timer.END)
		this.chargeSpeed = 0
	}

	this.stateTimer.Tick()

	// 治疗
	if this.effects.GetHPot() > 0 {
		//TODO
		//comb

		this.hp += this.effects.GetHPot()
		if this.hp > this.Get(stats.HP_MAX) {
			this.hp = this.Get(stats.HP_MAX)
		}
	}

	if this.effects.GetHPotPercent() > 0 {
		//TODO
		//comb
		hpot := this.Get(stats.HP_MAX) * this.effects.GetHPotPercent() / 100
		this.hp += hpot
		if this.hp > this.Get(stats.HP_MAX) {
			this.hp = this.Get(stats.HP_MAX)
		}
	}

	if this.effects.GetMPot() > 0 {
		//TODO
		//comb

		this.mp += this.effects.GetMPot()
		if this.mp > this.Get(stats.MP_MAX) {
			this.mp = this.Get(stats.MP_MAX)
		}
	}

	if this.effects.GetMPotPercent() > 0 {
		//TODO
		//comb
		mpot := this.Get(stats.MP_MAX) * this.effects.GetMPotPercent() / 100
		this.mp += mpot
		if this.mp > this.Get(stats.MP_MAX) {
			this.mp = this.Get(stats.MP_MAX)
		}
	}

	// 设置地图移动，碰撞模式
	if this.intangible {
		this.movementType = mapcollision.MOVE_INTANGIBLE
	} else if this.flying {
		this.movementType = mapcollision.MOVE_FLYING
	} else {
		this.movementType = mapcollision.MOVE_NORMAL
	}

	// 没血删除召唤物
	if this.hp == 0 {
		this.removeSummons()
	}

	// 被击退,计算位置
	if this.effects.GetKnockbackSpeed() != 0 {
		// 角度
		theta := utils.CalcTheta(this.knockbackSrcPos.X, this.knockbackSrcPos.Y, this.knockbackDestPos.X, this.knockbackDestPos.Y)
		this.knockbackSrcPos.X = (float32)(this.effects.GetKnockbackSpeed() * (float32)(math.Cos(float64(theta))))
		this.knockbackSrcPos.Y = (float32)(this.effects.GetKnockbackSpeed() * (float32)(math.Sin(float64(theta))))

		// TODO
		// map collider

	} else if this.chargeSpeed != 0 {
		// TODO
		// map collider
	}

	this.waypointTimer.Tick()

	// 复活
	if this.hp <= 0 && this.effects.GetRevive() {
		this.hp = this.Get(stats.HP_MAX)
		this.alive = true
		this.corpse = false
		this.curState = statblock.ENTITY_STANCE
	}

	if !this.hero && this.effects.GetConvert() != this.converted {
		this.converted = !this.converted
		this.heroAlly = !this.heroAlly
		if this.convertStatus != 0 {
			camp.SetStatus(this.convertStatus) // 激活属性
		}
	}
}

func (this *StatBlock) removeSummons() {
	//TODO
}

func (this *StatBlock) SetName(name string) {
	this.name = name
}

func (this *StatBlock) GetName() string {
	return this.name
}

func (this *StatBlock) SetGfxHead(val string) {
	this.gfxHead = val
}

func (this *StatBlock) GetGfxHead() string {
	return this.gfxHead
}

func (this *StatBlock) SetGfxBase(val string) {
	this.gfxBase = val
}

func (this *StatBlock) GetGfxBase() string {
	return this.gfxBase
}

func (this *StatBlock) SetGfxPortrait(val string) {
	this.gfxPortrait = val
}

func (this *StatBlock) GetGfxPortrait() string {
	return this.gfxPortrait
}

func (this *StatBlock) SetHero(val bool) {
	this.hero = val
}

func (this *StatBlock) GetHero() bool {
	return this.hero
}

func (this *StatBlock) SetCharacterClass(val string) {
	this.characterClass = val
}

func (this *StatBlock) SetCharacterSubclass(val string) {
	this.characterSubclass = val
}

func (this *StatBlock) SetXp(val uint64) {
	this.xp = val
}

// 当前 基础属性 + 基础属性加成
func (this *StatBlock) GetPrimary(index int) int {
	return this.primary[index] + this.primaryAdditional[index]
}

func (this *StatBlock) SetPrimary(index, val int) {
	this.primary[index] = val
}

func (this *StatBlock) SetPrimaryStarting(index, val int) {
	this.primaryStarting[index] = val
}

// 设置效果加成
func (this *StatBlock) SetPrimaryAdditional(index, val int) {
	this.primaryAdditional[index] = val
}

func (this *StatBlock) SetPermadeath(val bool) {
	this.permadeath = val
}

func (this *StatBlock) GetPermadeath() bool {
	return this.permadeath
}

func (this *StatBlock) SetLevel(val int) {
	this.level = val
}

func (this *StatBlock) GetLevel() int {
	return this.level
}

func (this *StatBlock) GetLongClass(modules common.Modules) string {
	msg := modules.Msg()

	if this.characterSubclass == "" || this.characterClass == this.characterSubclass {
		return msg.Get(this.characterClass)
	}

	return msg.Get(this.characterClass) + " / " + msg.Get(this.characterSubclass)
}

func (this *StatBlock) SetPerfectAccuracy(val bool) {
	this.perfectAccuracy = val
}

func (this *StatBlock) SetPos(pos fpoint.FPoint) {
	this.pos = pos

}

func (this *StatBlock) SetStarting(index, val int) {
	this.starting[index] = val
}

func (this *StatBlock) GetCorpse() bool {
	return this.corpse
}

func (this *StatBlock) GetCorpseTimer() *timer.Timer {
	return this.corpseTimer
}

func (this *StatBlock) GetHeroAlly() bool {
	return this.heroAlly
}

func (this *StatBlock) GetEnemyAlly() bool {
	return this.enemyAlly
}

func (this *StatBlock) GetPos() fpoint.FPoint {
	return this.pos
}

func (this *StatBlock) SetEncountered(val bool) {
	this.encountered = val
}

func (this *StatBlock) GetEncountered() bool {
	return this.encountered
}

func (this *StatBlock) GetHP() int {
	return this.hp
}

func (this *StatBlock) SetHP(val int) {
	this.hp = val
}

func (this *StatBlock) GetMP() int {
	return this.mp
}

func (this *StatBlock) SetMP(val int) {
	this.mp = val
}

func (this *StatBlock) GetEffects() gameres.EffectManager {
	return this.effects
}

func (this *StatBlock) SetTeleportation(val bool) {
	this.teleportation = val
}

func (this *StatBlock) GetTeleportation() bool {
	return this.teleportation
}

func (this *StatBlock) SetTeleportDestination(val fpoint.FPoint) {
	this.teleportDestination = val
}

func (this *StatBlock) GetTeleportDestination() fpoint.FPoint {
	return this.teleportDestination
}

func (this *StatBlock) GetSpeedDefault() float32 {
	return this.speedDefault
}

func (this *StatBlock) SetKnockbackSrcPos(val fpoint.FPoint) {
	this.knockbackSrcPos = val
}

func (this *StatBlock) SetKnockbackDestPos(val fpoint.FPoint) {
	this.knockbackDestPos = val
}

func (this *StatBlock) AddPartyBuff(powerId define.PowerId) {
	this.partyBuffs = append(this.partyBuffs, powerId)
}

func (this *StatBlock) SetTargetCorpse(val gameres.StatBlock) {
	this.targetCorpse = val
}

func (this *StatBlock) GetTargetCorpse() gameres.StatBlock {
	return this.targetCorpse
}

func (this *StatBlock) GetTargetNearest() gameres.StatBlock {
	return this.targetNearest
}

func (this *StatBlock) GetTargetNearestCorpse() gameres.StatBlock {
	return this.targetNearestCorpse
}

func (this *StatBlock) GetTargetNearestDist() float32 {
	return this.targetNearestDist
}

func (this *StatBlock) GetTargetNearestCorpseDist() float32 {
	return this.targetNearestCorpseDist
}

func (this *StatBlock) SetBlockPower(val define.PowerId) {
	this.blockPower = val
}

func (this *StatBlock) GetBlockPower() define.PowerId {
	return this.blockPower
}

func (this *StatBlock) GetAlive() bool {
	return this.alive
}

func (this *StatBlock) GetSpeed() float32 {
	return this.speed
}

func (this *StatBlock) SetSpeed(val float32) {
	this.speed = val
}

func (this *StatBlock) GetChargeSpeed() float32 {
	return this.chargeSpeed
}

func (this *StatBlock) GetMovementType() int {
	return this.movementType
}

func (this *StatBlock) SetCurState(val statblock.EntityState) {
	this.curState = val
}

func (this *StatBlock) GetCurState() statblock.EntityState {
	return this.curState
}

func (this *StatBlock) SetHumanoid(val bool) {
	this.humanoid = val
}

func (this *StatBlock) GetHumanoid() bool {
	return this.humanoid
}

func (this *StatBlock) GetCooldown() *timer.Timer {
	return this.cooldown
}

func (this *StatBlock) GetCooldownHit() *timer.Timer {
	return this.cooldownHit
}

func (this *StatBlock) GetCooldownHitEnabled() bool {
	return this.cooldownHitEnabled
}

func (this *StatBlock) GetTransformed() bool {
	return this.transformed
}

func (this *StatBlock) GetBlocking() bool {
	return this.blocking
}

func (this *StatBlock) SetRefreshStats(val bool) {
	this.refreshStats = val
}

func (this *StatBlock) GetPrevHP() int {
	return this.prevHP
}

func (this *StatBlock) GetXp() uint64 {
	return this.xp
}
