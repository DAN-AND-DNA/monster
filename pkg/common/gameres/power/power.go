package power

import (
	"monster/pkg/common"
	"monster/pkg/common/define"
	"monster/pkg/common/define/game/mapcollision"
	"monster/pkg/common/define/game/power"
)

const (
	HPMPSTATE_ANY = 0
	HPMPSTATE_ALL = 1
)

const (
	HPMPSTATE_IGNORE      = 0
	HPMPSTATE_PERCENT     = 1
	HPMPSTATE_NOT_PERCENT = 2
)

const (
	TYPE_FIXED     = 0
	TYPE_MISSILE   = 1
	TYPE_REPEATER  = 2
	TYPE_SPAWN     = 3
	TYPE_TRANSFORM = 4
	TYPE_EFFECT    = 5
	TYPE_BLOCK     = 6
)

const (
	STATE_INSTANT = 1
	STATE_ATTACK  = 2
)

const (
	STARTING_POS_SOURCE = 0
	STARTING_POS_TARGET = 1
	STARTING_POS_MELEE  = 2
)

const (
	TRIGGER_BLOCK      = 0
	TRIGGER_HIT        = 1
	TRIGGER_HALFDEATH  = 2
	TRIGGER_JOINCOMBAT = 3
	TRIGGER_DEATH      = 4
)

const (
	SPAWN_LIMIT_MODE_FIXED     = 0
	SPAWN_LIMIT_MODE_STAT      = 1
	SPAWN_LIMIT_MODE_UNLIMITED = 2
)

const (
	SPAWN_LEVEL_MODE_DEFAULT = 0
	SPAWN_LEVEL_MODE_FIXED   = 1
	SPAWN_LEVEL_MODE_STAT    = 2
	SPAWN_LEVEL_MODE_LEVEL   = 3
)

const (
	STAT_MODIFIER_MODE_MULTIPLY = 0
	STAT_MODIFIER_MODE_ADD      = 1
	STAT_MODIFIER_MODE_ABSOLUTE = 2
)

const (
	SOURCE_TYPE_HERO    = 0
	SOURCE_TYPE_NEUTRAL = 1
	SOURCE_TYPE_ENEMY   = 2
	SOURCE_TYPE_ALLY    = 3
)

const (
	SCRIPT_TRIGGER_CAST = 0
	SCRIPT_TRIGGER_HIT  = 1
	SCRIPT_TRIGGER_WALL = 2
)

type PostEffect struct {
	Id        string
	Magnitude int // 级别
	Duration  int
	Chance    int
	TargetSrc bool // 效果作用于施法者
}

func ConstructPostEffect() PostEffect {
	pe := PostEffect{
		Chance: 100,
	}
	return pe
}

type ReplaceByEffect struct {
	PowerId  int
	Count    int
	EffectId string
}

func ConstructReplaceByEffect() ReplaceByEffect {
	p := ReplaceByEffect{}
	return p
}

type RequiredItem struct {
	Id       define.ItemId
	Quantity int
	Equipped bool
}

func ConstructRequiredItem() RequiredItem {
	return RequiredItem{}
}

type HPMPState struct {
	Mode    int
	HPState int
	MPState int
	HP      int
	MP      int
}

func ConstructHPMPState() HPMPState {
	hms := HPMPState{
		Mode:    power.HPMPSTATE_ANY,
		HPState: power.HPMPSTATE_IGNORE,
		MPState: power.HPMPSTATE_IGNORE,
		HP:      -1,
		MP:      -1,
	}

	return hms
}

type RemoveEffectPair struct {
	First  string
	Second int
}

// 技能
type Power struct {
	IsEmpty          bool
	Type             int // 激活类型
	Name             string
	Description      string
	Icon             int    // 技能图标id
	NewState         int    // 使用技能得到的新状态
	StateDuration    int    // 状态持续时间，时间过长，状态动画会暂停在最后一帧
	PreventInterrupt bool   // 是否可以打断技能
	AttackAnim       string // 使用技能时的动画
	Face             bool   // 是否释放时需要面向目标位置
	SourceType       int    // 技能影响什么类型
	Beacon           bool   // 是否敌人会呼唤盟友
	Count            int    // 技能造成效果、伤害或召唤的数量
	Passive          bool   // 被动技能
	PassiveTrigger   int    // 被动技能的触发条件
	MetaPower        bool   // 道具自己释放的技能
	NoActionbar      bool

	// 技能的要求
	RequiresFlags       map[string]struct{} // 需要的装备标签，比如需要弓
	RequiresMP          int                 // 需要蓝等级
	RequiresHP          int                 // 红等级
	Sacrifice           bool                // 技能需要消耗红
	RequiresLos         bool                // 需要视野
	RequiresLosDefault  bool
	RequiresEmptyTarget bool           // 技能释放需要目标为空
	RequiredItems       []RequiredItem // 需要仓库里有该物品，可作为技能消耗道具或使用物品
	RequiresTargeting   bool           // 需要释放到目标
	RequiresSpawns      int            // 需要召唤物
	Cooldown            int            // cd  时间
	RequiresMaxHPMP     HPMPState

	// 动画信息
	AnimationName     string
	Directional       bool
	VisualRandom      int
	VisualOption      int
	AimAssist         bool
	Speed             float32 // 飞弹类的移动速度
	Lifespan          int     // 动画和伤害持续时间
	OnFloor           bool    // 是否绘制在对象和背景之间
	CompleteAnimation bool    // 是否命中了也要播完动画
	ChargeSpeed       float32 // 施法者冲锋速度
	AttackSpeed       float32 // 技能的攻击动画速度

	// 伤害相关的
	UseHazard             bool     // 是否使用伤害
	NoAttack              bool     // 是否是攻击技能
	NoAggro               bool     // 技能不吸引仇恨，目标不会进行战斗
	Radius                float32  // 伤害的像素半径
	BaseDamage            int      // 伤害类型
	StartingPos           int      // 伤害的开始位置，直接在对方身上还是，飞过去的。。。
	RelativePos           bool     // 伤害是否相对于施法者移动
	Multitarget           bool     // 是否伤害多个目标
	Multihit              bool     // 是否伤害可以多次命中同一个目标
	ExpireWithCaster      bool     // 施法者死亡，技能是否失效
	IgnoreZeroDamage      bool     //
	LockTargetToDirection bool     // 只能按8个标准方向移动
	MovementType          int      // 伤害技能在地图上的碰撞类型
	TargetRange           float32  // 目标距离施放者的距离
	TargetParty           bool     // 伤害只会影响自己人
	TargetCategories      []string // 伤害会影响到的敌人类型
	CombatRange           float32

	ModAccuracyMode   int // 修改技能命中率模式
	ModAccuracyValue  int // 技能命中率值
	ModCritMode       int // 修改技能导致暴击概率的模式
	ModCritValue      int
	ModDamageMode     int // 修改技能的伤害模式
	ModDamageValueMin int
	ModDamageValueMax int

	// 偷取类
	HPSteal int // 伤害的百分比
	MPSteal int

	// 飞弹
	MissileAngle  int     // 角度
	AngleVariance int     // 角度增量
	SpeedVariance float32 // 速度增量

	Delay int

	TraitElemental        int  // 造成的伤害是何种元素
	TraitArmorPenetration bool // 是否无视目标伤害吸收属性
	TraitAvoidanceIgnore  bool // 是否无视目标闪避属性
	TraitCritsImpaired    int  // 对无法移动或者减速的目标的额外暴击概率

	// buff 和 debuff的持续时间
	TransformDuration int
	ManualUnTransform bool
	KeepEquipment     bool
	UntransformOnHit  bool

	Buff             bool           // 技能是给自己的buff
	BuffTeleport     bool           // 技能是传送技能
	BuffParty        bool           // 技能是团队buff
	BuffPartyPowerId define.PowerId // 团队buff只影响某个技能的召唤物

	PostEffects     []PostEffect
	PrePower        define.PowerId
	PrePowerChance  int
	PostPower       define.PowerId // 后续在触发其他技能
	PostPowerChance int            // 后续技能触发概率
	WallPower       define.PowerId // 技能撞强之后还会触发什么技能
	WallPowerChance int
	WallReflect     bool

	// 技能展现信息
	SpawnType       string // 是召唤还是自己变身
	TargetNeighbor  int
	SpawnLimitMode  int // 召唤物限制类型，数量还是属性
	SpawnLimitQty   int // 对应的数值
	SpawnLimitEvery int
	SpawnLimitStat  int

	SpawnLevelMode  int
	SpawnLevelQty   int
	SpawnLevelEvery int
	SpawnLevelStat  int

	// 接受技能的目标 移动信息
	TargetMovementNormal     bool // 该技能能影响普通地图移动对象
	TargetMovementFlying     bool
	TargetMovementIntangible bool
	WallsBlockAoe            bool // 是否墙后单位会受影响

	ScriptTrigger int    // 触发条件达成执行某些脚本
	Script        string //

	RemoveEffects   []RemoveEffectPair // 技能会对应清理掉其他效果，指定id和数量
	ReplaceByEffect []ReplaceByEffect

	RequiresCorpse bool // 技能释放需要尸体
	RemoveCorpse   bool // 尸体是否在技能释放后被删除

	TargetNearest     float32 // 半径范围，最近的目标
	DisableEquipSlots []string
}

func New(modules common.Modules) *Power {
	p := Construct(modules)
	return &p
}

func Construct(modules common.Modules) Power {
	p := Power{}
	p.init(modules)
	return p
}

func (this *Power) init(modules common.Modules) {
	eset := modules.Eset()

	this.RequiresFlags = map[string]struct{}{}
	this.IsEmpty = true
	this.Type = -1
	this.Icon = -1
	this.NewState = -1
	this.Count = 1
	this.PassiveTrigger = -1
	this.RequiresLosDefault = true
	this.AttackSpeed = 100
	dtList := eset.Get("damage_types", "list").([]common.DamageType)
	this.BaseDamage = len(dtList)
	this.StartingPos = power.STARTING_POS_SOURCE
	this.MovementType = mapcollision.MOVE_FLYING
	this.ModAccuracyMode = -1
	this.ModAccuracyValue = 100
	this.ModCritMode = -1
	this.ModCritValue = 100
	this.ModDamageValueMin = 100
	this.ModDamageValueMax = 0
	this.TraitElemental = -1
	this.PrePowerChance = 100
	this.PostPowerChance = 100
	this.WallPowerChance = 100
	this.SpawnLimitMode = power.SPAWN_LIMIT_MODE_UNLIMITED
	this.SpawnLimitQty = 1
	this.SpawnLimitEvery = 1
	this.SpawnLevelMode = power.SPAWN_LEVEL_MODE_DEFAULT
	this.TargetMovementNormal = true
	this.TargetMovementFlying = true
	this.TargetMovementIntangible = true
	this.ScriptTrigger = -1

}
