package effect

import (
	"fmt"
	"monster/pkg/common"
	"monster/pkg/common/define/game/stats"
	"monster/pkg/common/gameres"
	"monster/pkg/common/gameres/effect"
)

func GetTypeFromString(modules common.Modules, ss gameres.Stats, typeStr string) int {
	eset := modules.Eset()

	if typeStr == "" {
		return effect.NONE
	}

	switch typeStr {
	case "damage":
		return effect.DAMAGE
	case "damage_percent":
		return effect.DAMAGE_PERCENT
	case "hpot":
		return effect.HPOT
	case "hpot_percent":
		return effect.HPOT_PERCENT
	case "mpot":
		return effect.MPOT
	case "mpot_percent":
		return effect.MPOT_PERCENT
	case "speed":
		return effect.SPEED
	case "attack_speed":
		return effect.ATTACK_SPEED
	case "immunity":
		return effect.IMMUNITY
	case "immunity_damage":
		return effect.IMMUNITY_DAMAGE
	case "immunity_slow":
		return effect.IMMUNITY_SLOW
	case "immunity_stun":
		return effect.IMMUNITY_STUN
	case "immunity_hp_steal":
		return effect.IMMUNITY_HP_STEAL
	case "immunity_mp_steal":
		return effect.IMMUNITY_MP_STEAL
	case "immunity_knockback":
		return effect.IMMUNITY_KNOCKBACK
	case "immunity_damage_reflect":
		return effect.IMMUNITY_DAMAGE_REFLECT
	case "immunity_stat_debuff":
		return effect.IMMUNITY_STAT_DEBUFF
	case "stun":
		return effect.STUN
	case "revive":
		return effect.REVIVE
	case "convert":
		return effect.CONVERT
	case "fear":
		return effect.FEAR
	case "death_sentence":
		return effect.DEATH_SENTENCE
	case "shield":
		return effect.SHIELD
	case "heal":
		return effect.HEAL
	case "knockback":
		return effect.KNOCKBACK
	default:
		for i := 0; i < stats.COUNT; i++ {
			if typeStr == ss.GetKey((stats.STAT)(i)) {
				return effect.TYPE_COUNT + i
			}
		}

		dtList := eset.Get("damage_types", "list").([]common.DamageType)
		for index, ptr := range dtList {
			if typeStr == ptr.GetMin() {
				return effect.TYPE_COUNT + stats.COUNT + index*2
			} else if typeStr == ptr.GetMax() {
				return effect.TYPE_COUNT + stats.COUNT + index*2 + 1
			}
		}

		eList := eset.Get("elements", "list").([]common.Element)
		for index, ptr := range eList {
			if typeStr == ptr.GetId()+"_resist" {
				return effect.TYPE_COUNT + stats.COUNT + eset.Get("damage_types", "count").(int) + index
			}
		}

		psList := eset.Get("primary_stats", "list").([]common.PrimaryStat)
		for index, ptr := range psList {
			if typeStr == ptr.GetId() {
				return effect.TYPE_COUNT + stats.COUNT + eset.Get("damage_types", "count").(int) + len(eList) + index
			}
		}
	}

	panic(fmt.Sprintf("EffectManager: '%s' is not a valid effect type.\n", typeStr))
}

func TypeIsStat(t int) bool {
	return t >= effect.TYPE_COUNT && t < effect.TYPE_COUNT+stats.COUNT
}

func TypeIsDmgMin(modules common.Modules, t int) bool {
	eset := modules.Eset()

	return t >= effect.TYPE_COUNT+stats.COUNT &&
		t < effect.TYPE_COUNT+stats.COUNT+eset.Get("damage_types", "count").(int) &&
		(t-stats.COUNT-effect.TYPE_COUNT)%2 == 0
}

func TypeIsDmgMax(modules common.Modules, t int) bool {
	eset := modules.Eset()

	return t >= effect.TYPE_COUNT+stats.COUNT &&
		t < effect.TYPE_COUNT+stats.COUNT+eset.Get("damage_types", "count").(int) &&
		(t-stats.COUNT-effect.TYPE_COUNT)%2 == 1
}

func TypeIsResist(modules common.Modules, t int) bool {
	eset := modules.Eset()
	eList := eset.Get("elements", "list").([]common.Element)

	return t > effect.TYPE_COUNT+stats.COUNT+eset.Get("damage_types", "count").(int) &&
		t < effect.TYPE_COUNT+stats.COUNT+eset.Get("damage_types", "count").(int)+len(eList)
}

func TypeIsPrimary(modules common.Modules, t int) bool {
	eset := modules.Eset()
	eList := eset.Get("elements", "list").([]common.Element)
	psList := eset.Get("primary_stats", "list").([]common.PrimaryStat)

	return t > effect.TYPE_COUNT+stats.COUNT+eset.Get("damage_types", "count").(int)+len(eList) &&
		t < effect.TYPE_COUNT+stats.COUNT+eset.Get("damage_types", "count").(int)+len(eList)+len(psList)
}

func GetStatFromType(t int) int {
	return t - effect.TYPE_COUNT
}

func GetDmgFromType(t int) int {
	return t - effect.TYPE_COUNT - stats.COUNT
}

func GetResistFromType(modules common.Modules, t int) int {
	eset := modules.Eset()

	return t - effect.TYPE_COUNT - stats.COUNT - eset.Get("damage_types", "count").(int)
}

func GetPrimaryFromType(modules common.Modules, t int) int {
	eset := modules.Eset()
	eList := eset.Get("elements", "list").([]common.Element)

	return t - effect.TYPE_COUNT - stats.COUNT - eset.Get("damage_types", "count").(int) - len(eList)
}
