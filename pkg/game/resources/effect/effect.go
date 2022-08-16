package effect

import (
	"math"
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define"
	"monster/pkg/common/define/game/stats"
	"monster/pkg/common/gameres"
	"monster/pkg/common/gameres/effect"
	"monster/pkg/common/gameres/power"
)

type Manager struct {
	effectList            []effect.Effect
	damage                int
	damagePercent         int
	hpot                  int // 每秒治疗
	hpotPercent           int
	mpot                  int
	mpotPercent           int
	speed                 int // 速度百分比
	immunityDamage        bool
	immunitySlow          bool
	immunityStun          bool
	immunityHPSteal       bool
	immunityMPSteal       bool
	immunityKnockback     bool
	immunityDamageReflect bool
	immunityStatDebuff    bool
	stun                  bool // 晕眩
	revive                bool
	convert               bool // 被转化成盟友
	deathSentence         bool
	fear                  bool
	knockbackSpeed        float32 // 被击退速度

	bonus        []int // 子属性和伤害加成
	bonusResist  []int // 防御加成
	bonusPrimary []int // 基础属性加成

	triggeredOthers     bool
	triggeredBlock      bool
	triggeredHit        bool
	triggeredHalfdeath  bool
	triggeredJoincombat bool
	triggeredDeath      bool

	refreshStats bool
}

func New(modules common.Modules) *Manager {
	em := &Manager{}
	em.Init(modules)
	return em
}

func (this *Manager) Init(modules common.Modules) gameres.EffectManager {
	eset := modules.Eset()

	eList := eset.Get("elements", "list").([]common.Element)
	psList := eset.Get("primary_stats", "list").([]common.PrimaryStat)

	// 加成数量同基础属性
	this.bonus = make([]int, stats.COUNT+eset.Get("damage_types", "count").(int))
	this.bonusResist = make([]int, len(eList))
	this.bonusPrimary = make([]int, len(psList))

	this.clearStatus()
	return this
}

func (this *Manager) Close() {
}

func (this *Manager) clearStatus() {
	this.damage = 0
	this.damagePercent = 0
	this.hpot = 0
	this.hpotPercent = 0
	this.mpot = 0
	this.mpotPercent = 0
	this.speed = 100
	this.immunityDamage = false
	this.immunitySlow = false
	this.immunityStun = false
	this.immunityHPSteal = false
	this.immunityMPSteal = false
	this.immunityKnockback = false
	this.immunityDamageReflect = false
	this.immunityStatDebuff = false
	this.stun = false
	this.revive = false
	this.convert = false
	this.deathSentence = false
	this.fear = false
	this.knockbackSpeed = 0

	for i, _ := range this.bonus {
		this.bonus[i] = 0
	}

	for i, _ := range this.bonusResist {
		this.bonusResist[i] = 0
	}

	for i, _ := range this.bonusPrimary {
		this.bonusPrimary[i] = 0
	}
}

func (this *Manager) Logic(modules common.Modules) {
	settings := modules.Settings()
	eset := modules.Eset()

	this.clearStatus()

	maxFPS := settings.Get("max_fps").(int)
	dtCount := eset.Get("damage_types", "count").(int)
	eList := eset.Get("elements", "list").([]common.Element)

	// 遍历已拥有的效果
	for i := 0; i < len(this.effectList); i++ {
		ei := &(this.effectList[i])

		if ei.Timer.GetDuration() > 0 {
			// 效果时间结束
			if ei.Timer.IsEnd() {
				if ei.Type == effect.DEATH_SENTENCE {
					this.deathSentence = true
				}

				this.removeEffect(i) // 删除本次的序号
				i--                  // 回前一个序号
				continue
			}
		}

		// 每秒效果，跟着时间
		doTimedEffect := ei.Timer.IsWholeSecond(maxFPS) ||
			(ei.Timer.GetDuration() < (uint)(maxFPS) &&
				ei.Timer.IsBegin())

			// 每秒伤害
		if ei.Type == effect.DAMAGE && doTimedEffect {
			this.damage += ei.Magnitude
		} else if ei.Type == effect.DAMAGE_PERCENT && doTimedEffect {
			this.damagePercent += ei.Magnitude
		} else if ei.Type == effect.HPOT && doTimedEffect {
			this.hpot += ei.Magnitude
		} else if ei.Type == effect.HPOT_PERCENT && doTimedEffect {
			this.hpotPercent += ei.Magnitude
		} else if ei.Type == effect.MPOT && doTimedEffect {
			this.mpot += ei.Magnitude
		} else if ei.Type == effect.MPOT_PERCENT && doTimedEffect {
			this.mpotPercent += ei.Magnitude
		} else if ei.Type == effect.SPEED {
			// 速度加成百分比
			this.speed = (ei.Magnitude * this.speed) / 100
		} else if ei.Type == effect.IMMUNITY {
			this.immunityDamage = true
			this.immunitySlow = true
			this.immunityStun = true
			this.immunityHPSteal = true
			this.immunityMPSteal = true
			this.immunityKnockback = true
			this.immunityDamageReflect = true
			this.immunityStatDebuff = true
		} else if ei.Type == effect.IMMUNITY_DAMAGE {
			this.immunityDamage = true
		} else if ei.Type == effect.IMMUNITY_SLOW {
			this.immunitySlow = true
		} else if ei.Type == effect.IMMUNITY_STUN {
			this.immunityStun = true
		} else if ei.Type == effect.IMMUNITY_HP_STEAL {
			this.immunityHPSteal = true
		} else if ei.Type == effect.IMMUNITY_MP_STEAL {
			this.immunityMPSteal = true
		} else if ei.Type == effect.IMMUNITY_KNOCKBACK {
			this.immunityKnockback = true
		} else if ei.Type == effect.IMMUNITY_DAMAGE_REFLECT {
			this.immunityDamageReflect = true
		} else if ei.Type == effect.IMMUNITY_STAT_DEBUFF {
			this.immunityStatDebuff = true
		} else if ei.Type == effect.STUN {
			this.stun = true
		} else if ei.Type == effect.REVIVE {
			this.revive = true
		} else if ei.Type == effect.CONVERT {
			this.convert = true
		} else if ei.Type == effect.FEAR {
			this.fear = true
		} else if ei.Type == effect.KNOCKBACK {
			this.knockbackSpeed = float32(ei.Magnitude) / (float32)(maxFPS)
		} else if ei.Type >= effect.TYPE_COUNT && ei.Type < effect.TYPE_COUNT+stats.COUNT+dtCount {
			//  基础属性加成
			this.bonus[ei.Type-effect.TYPE_COUNT] += ei.Magnitude
		} else if ei.Type >= effect.TYPE_COUNT+stats.COUNT+dtCount && ei.Type < effect.TYPE_COUNT+stats.COUNT+dtCount+len(eList) {
			// 抗性加成
			this.bonusResist[ei.Type-effect.TYPE_COUNT-stats.COUNT-dtCount] += ei.Magnitude
		} else if ei.Type >= effect.TYPE_COUNT {
			this.bonusResist[ei.Type-effect.TYPE_COUNT-stats.COUNT-dtCount-len(eList)] += ei.Magnitude
		}

		ei.Timer.Tick()

		// 伤害吸收到极限
		if ei.MagnitudeMax > 0 && ei.Magnitude == 0 {
			if ei.Type == effect.SHIELD {
				this.removeEffect(i)
				i--
				continue
			}
		}

		// 根据动画来结束效果
		if (ei.Animation != nil && ei.Animation.IsLastFrame()) || ei.Animation == nil {
			if ei.Type == effect.HEAL {
				this.removeEffect(i)
				i--
				continue
			}
		}

		// 播放效果动画
		if ei.Animation != nil {
			if !ei.Animation.IsCompleted() {
				ei.Animation.AdvanceFrame()
			}
		}

	}
}

// 清理掉指定id和数量的效果
func (this *Manager) RemoveEffectId(removeEffects []power.RemoveEffectPair) {
	for _, val := range removeEffects {
		count := val.Second
		removeAll := false
		if count == 0 {
			removeAll = true
		}

		for j := len(this.effectList); j > 0; j-- {
			if !removeAll && count <= 0 {
				break
			}

			if this.effectList[j-1].Id == val.First {
				this.removeEffect(j - 1)
				count--
			}
		}
	}
}

func (this *Manager) removeEffect(id int) {
	leftEffectList := this.effectList[id+1:]
	this.effectList = this.effectList[:len(this.effectList)-1]

	for index, val := range leftEffectList {
		this.effectList[index+id] = val
	}

	this.refreshStats = true
}

func (this *Manager) AddEffect(modules common.Modules, def effect.Def, duration, magnitude, sourceType int, powerId define.PowerId) {
	this.addEffectInternal(modules, def, duration, magnitude, sourceType, false, powerId)
}

func (this *Manager) addEffectInternal(modules common.Modules, def effect.Def, duration, magnitude, sourceType int, item bool, powerId define.PowerId) {

	this.refreshStats = true

	if this.immunityDamage && def.Type == effect.DAMAGE || def.Type == effect.DAMAGE_PERCENT {
		return
	} else if this.immunitySlow && def.Type == effect.SPEED && magnitude < 100 {
		return
	} else if this.immunityStun && def.Type == effect.STUN {
		return
	} else if this.immunityKnockback && def.Type == effect.KNOCKBACK {
		return
	} else if this.immunityStatDebuff && def.Type > effect.TYPE_COUNT && magnitude < 0 {
		return
	}

	// 单次只允许一种击退
	if def.Type == effect.KNOCKBACK && this.knockbackSpeed != 0 {
		return
	}

	insertEffect := false
	insertPos := 0
	stacksApplied := 0
	trigger := -1

	if powerId > 0 {
		// TODO
		// 技能主动触发
	}

	passiveId := 0
	if powerId > 0 {
		// TODO
		// 主动技能
	}

	for i := len(this.effectList); i > 0; i-- {
		ei := &(this.effectList[i-1])

		// 找到对应的效果
		if ei.Type == def.Type && ei.Id == def.Id {
			if trigger > -1 && ei.Trigger == trigger {
				// 同种触发方式类型的效果单次只能一个
				return
			}

			if !def.CanStack {
				this.removeEffect(i - 1)
			} else {

				// 单独处理护盾吸收
				if def.Type == effect.SHIELD && def.GroupStack {
					ei.Magnitude += magnitude

					if def.MaxStacks == -1 || (magnitude != 0 && ei.MagnitudeMax/magnitude < def.MaxStacks) {
						ei.MagnitudeMax += magnitude

					}

					if ei.Magnitude > ei.MagnitudeMax {
						ei.Magnitude = ei.MagnitudeMax
					}

					return
				}

				// 叠加效果
				if insertEffect == false && def.MaxStacks != -1 {
					insertEffect = true
					insertPos = i
				}

				stacksApplied++
			}
		}

		if def.Type == effect.IMMUNITY {
			this.clearNegativeEffects(-1)
		} else if def.Type == effect.IMMUNITY_DAMAGE {
			this.clearNegativeEffects(effect.IMMUNITY_DAMAGE)
		} else if def.Type == effect.IMMUNITY_SLOW {
			this.clearNegativeEffects(effect.IMMUNITY_SLOW)
		} else if def.Type == effect.IMMUNITY_STUN {
			this.clearNegativeEffects(effect.IMMUNITY_STUN)
		} else if def.Type == effect.IMMUNITY_KNOCKBACK {
			this.clearNegativeEffects(effect.IMMUNITY_KNOCKBACK)
		}
	}

	e := effect.Construct()
	e.Id = def.Id
	e.Name = def.Name
	e.Icon = def.Icon
	e.Type = def.Type
	e.RenderAbove = def.RenderAbove
	e.GrounpStack = def.GroupStack
	e.ColorMod = def.ColorMod.EncodeRGBA()
	e.AlphaMod = def.AlphaMod
	e.AttackSpeedAnim = def.AttackSpeedAnim
	if def.Animation != "" {
		e.LoadAnimation(modules, def.Animation)
	}

	e.Timer.SetDuration((uint)(duration))
	e.Magnitude = magnitude
	e.MagnitudeMax = magnitude
	e.Item = item
	e.Trigger = trigger
	e.PassiveId = passiveId
	e.SourceType = sourceType

	if insertEffect {

		// 和最近插入的同个效果进行整合
		if def.MaxStacks != -1 && stacksApplied >= def.MaxStacks {
			this.removeEffect(insertPos - stacksApplied)
			insertPos--
		}

		leftList := this.effectList[insertPos:]
		this.effectList = this.effectList[:insertPos]
		this.effectList = append(this.effectList, e)
		this.effectList = append(this.effectList, leftList...)

	} else {
		this.effectList = append(this.effectList, e)
	}

}

func (this *Manager) clearNegativeEffects(type1 int) {
	for i := len(this.effectList); i > 0; i-- {
		if (type1 == -1 || type1 == effect.IMMUNITY_DAMAGE) && this.effectList[i-1].Type == effect.DAMAGE {
			this.removeEffect(i - 1)
		} else if (type1 == -1 || type1 == effect.IMMUNITY_DAMAGE) && this.effectList[i-1].Type == effect.DAMAGE_PERCENT {
			this.removeEffect(i - 1)
		} else if (type1 == -1 || type1 == effect.IMMUNITY_SLOW) && this.effectList[i-1].Type == effect.SPEED && this.effectList[i-1].MagnitudeMax < 100 {
			this.removeEffect(i - 1)
		} else if (type1 == -1 || type1 == effect.IMMUNITY_STUN) && this.effectList[i-1].Type == effect.STUN {
			this.removeEffect(i - 1)
		} else if (type1 == -1 || type1 == effect.IMMUNITY_KNOCKBACK) && this.effectList[i-1].Type == effect.KNOCKBACK {
			this.removeEffect(i - 1)
		} else if (type1 == -1 || type1 == effect.IMMUNITY_STAT_DEBUFF) && this.effectList[i-1].Type > effect.TYPE_COUNT && this.effectList[i-1].MagnitudeMax < 0 {
			this.removeEffect(i - 1)
		}
	}
}

// 伤害打在护盾上被吸收了部分或全部，还保留多少
func (this *Manager) DamageShields(dmg int) int {
	overDmg := dmg

	for i := 0; i < len(this.effectList); i++ {
		if this.effectList[i].MagnitudeMax > 0 && this.effectList[i].Type == effect.SHIELD {
			// 护盾效果
			this.effectList[i].Magnitude -= overDmg

			if this.effectList[i].Magnitude < 0 {
				overDmg = int(math.Abs(float64(this.effectList[i].Magnitude)))
				this.effectList[i].Magnitude = 0
			} else {
				return 0
			}
		}
	}

	return overDmg
}

func (this *Manager) GetBonusPrimary(index int) int {
	if index >= len(this.bonusPrimary) || index < 0 {
		panic("bad index for bonus primary")
	}

	return this.bonusPrimary[index]
}

func (this *Manager) GetBonus(index int) int {
	if index >= len(this.bonus) || index < 0 {
		panic("bad index for bonus")
	}

	return this.bonus[index]
}

func (this *Manager) GetBonusResist(index int) int {
	if index >= len(this.bonusResist) || index < 0 {
		panic("bad index for bonus resist")
	}

	return this.bonusResist[index]
}
func (this *Manager) SetTriggeredBlock(val bool) {
	this.triggeredBlock = val
}

func (this *Manager) GetTriggeredBlock() bool {
	return this.triggeredBlock
}

func (this *Manager) SetTriggeredDeath(val bool) {
	this.triggeredDeath = val
}

func (this *Manager) GetTriggeredDeath() bool {
	return this.triggeredDeath
}

func (this *Manager) GetRevive() bool {
	return this.revive
}

func (this *Manager) GetRefreshStats() bool {
	return this.refreshStats
}

func (this *Manager) SetRefreshStats(val bool) {
	this.refreshStats = val
}

func (this *Manager) GetDamage() int {
	return this.damage
}

func (this *Manager) GetDamagePercent() int {
	return this.damagePercent
}

// 获得伤害效果来自哪个技能来源
func (this *Manager) GetDamageSourceType(dmgMod int) int {
	if !(dmgMod == effect.DAMAGE || dmgMod == effect.DAMAGE_PERCENT) {
		panic("bad type for damage source type")
	}

	sourceType := power.SOURCE_TYPE_NEUTRAL

	for _, val := range this.effectList {
		if val.Type == dmgMod {
			if val.SourceType != power.SOURCE_TYPE_ALLY {
				return val.SourceType
			}
			sourceType = val.SourceType
		}
	}

	return sourceType
}

func (this *Manager) GetDeathSentence() bool {
	return this.deathSentence
}

func (this *Manager) GetStun() bool {
	return this.stun
}

func (this *Manager) GetHPot() int {
	return this.hpot
}

func (this *Manager) GetMPot() int {
	return this.mpot
}

func (this *Manager) GetHPotPercent() int {
	return this.hpotPercent
}

func (this *Manager) GetMPotPercent() int {
	return this.mpotPercent
}

func (this *Manager) GetKnockbackSpeed() float32 {
	return this.knockbackSpeed
}

func (this *Manager) GetConvert() bool {
	return this.convert
}

func (this *Manager) GetSpeed() int {
	return this.speed
}

func (this *Manager) GetCurrentColor(colorMod color.Color) color.Color {
	defaultColor := colorMod.EncodeRGBA()
	noColor := color.Construct(255, 255, 255).EncodeRGBA()

	for i := len(this.effectList); i > 0; i-- {
		ei := this.effectList[i-1]
		if ei.ColorMod == noColor {
			continue
		}

		if ei.ColorMod != defaultColor {
			colorMod.DecodeRGBA(ei.ColorMod)
			return colorMod
		}
	}

	return colorMod
}

func (this *Manager) GetCurrentAlpha(alphaMod uint8) uint8 {
	defaultAlpha := alphaMod
	noAlpha := (uint8)(255)

	for i := len(this.effectList); i > 0; i-- {
		ei := this.effectList[i-1]
		if ei.AlphaMod == noAlpha {
			continue
		}

		if ei.AlphaMod != defaultAlpha {
			alphaMod = ei.AlphaMod
			return alphaMod
		}
	}

	return alphaMod
}

func (this *Manager) ClearTriggerEffects(trigger int) {
	for i := len(this.effectList); i > 0; i-- {
		if this.effectList[i-1].Trigger > -1 && this.effectList[i-1].Trigger == trigger {
			this.removeEffect(i - 1)
		}
	}
}
