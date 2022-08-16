package statblock

import (
	"math"
	"monster/pkg/common/define"
	"monster/pkg/common/timer"
)

type EntityState uint8
type CombatStyle uint8

var (
	M_SQRT2           = (float32)(math.Sqrt(2))
	DIRECTION_DELTA_X = []float32{-1, -1, -1, 0, 1, 1, 1, 0}
	DIRECTION_DELTA_Y = []float32{1, 0, -1, -1, -1, 0, 1, 1}
	SPEED_MULTIPLIER  = []float32{(1 / M_SQRT2), 1, (1 / M_SQRT2), 1, (1 / M_SQRT2), 1, (1 / M_SQRT2), 1}
)

const (
	AI_POWER_MELEE = iota
	AI_POWER_RANGED
	AI_POWER_BEACON
	AI_POWER_HIT
	AI_POWER_DEATH
	AI_POWER_HALF_DEAD
	AI_POWER_JOIN_COMBAT
	AI_POWER_DEBUFF
	AI_POWER_PASSIVE_POST
)

const (
	ENTITY_STANCE EntityState = iota
	ENTITY_MOVE
	ENTITY_POWER
	ENTITY_SPAWN
	ENTITY_BLOCK
	ENTITY_HIT
	ENTITY_DEAD
	ENTITY_CRITDEAD
)

const (
	COMBAT_DEFAULT CombatStyle = iota
	COMBAT_AGGRESSIVE
	COMBAT_PASSIVE
)

type AIPower struct {
	Type     int // 技能触发方式
	Id       define.PowerId
	Chance   int
	Cooldown timer.Timer
}

func ConstructAIPower() AIPower {
	return AIPower{
		Type:     AI_POWER_MELEE, // 近战
		Cooldown: timer.Construct(),
	}
}
