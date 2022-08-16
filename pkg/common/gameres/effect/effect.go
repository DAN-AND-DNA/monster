package effect

import (
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/gameres/power"
	"monster/pkg/common/timer"
)

const (
	TYPE_COUNT = 26
)

const (
	NONE                    = 0
	DAMAGE                  = 1
	DAMAGE_PERCENT          = 2
	HPOT                    = 3
	HPOT_PERCENT            = 4
	MPOT                    = 5
	MPOT_PERCENT            = 6
	SPEED                   = 7
	ATTACK_SPEED            = 8
	IMMUNITY                = 9
	IMMUNITY_DAMAGE         = 10
	IMMUNITY_SLOW           = 11
	IMMUNITY_STUN           = 12
	IMMUNITY_HP_STEAL       = 13
	IMMUNITY_MP_STEAL       = 14
	IMMUNITY_KNOCKBACK      = 15
	IMMUNITY_DAMAGE_REFLECT = 16
	IMMUNITY_STAT_DEBUFF    = 17
	STUN                    = 18
	REVIVE                  = 19
	CONVERT                 = 20
	FEAR                    = 21
	DEATH_SENTENCE          = 22
	SHIELD                  = 23
	HEAL                    = 24
	KNOCKBACK               = 25
)

type Effect struct {
	Id              string
	Name            string
	Icon            int
	Timer           timer.Timer // 持续时间
	Type            int
	Magnitude       int // 当前效果的数值，比如护盾
	MagnitudeMax    int // 最大
	AnimationName   string
	Animation       common.Animation
	Item            bool
	Trigger         int // 触发方式
	RenderAbove     bool
	PassiveId       int
	SourceType      int
	GrounpStack     bool
	ColorMod        uint32
	AlphaMod        uint8
	AttackSpeedAnim string
}

func Construct() Effect {
	effect := Effect{}
	effect.init()

	return effect
}

func (this *Effect) init() {
	this.Icon = -1
	this.Timer = timer.Construct()
	this.Type = NONE
	this.Trigger = -1
	this.SourceType = power.SOURCE_TYPE_HERO
	this.ColorMod = color.Construct(255, 255, 255).EncodeRGBA()
	this.AlphaMod = 255
}

func (this *Effect) LoadAnimation(modules common.Modules, s string) error {
	settings := modules.Settings()
	mods := modules.Mods()
	render := modules.Render()
	mresf := modules.Resf()
	anim := modules.Anim()

	if s == "" {
		return nil
	}

	this.AnimationName = s
	anim.IncreaseCount(s)
	animationSet, err := anim.GetAnimationSet(settings, mods, render, mresf, s)
	if err != nil {
		return err
	}
	this.Animation = animationSet.GetAnimation("")

	return nil
}

func (this *Effect) UnloadAnimation(modules common.Modules) {
	anim := modules.Anim()

	if this.Animation != nil {
		if this.AnimationName != "" {
			anim.DecreaseCount(this.AnimationName)
		}
		this.Animation.Close()
		this.Animation = nil
	}
}
