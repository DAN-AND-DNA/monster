package stats

import (
	"fmt"
	"monster/pkg/common"
	"monster/pkg/common/define/game/stats"
)

type Stats struct {
	keys     []string
	names    []string
	descs    []string
	percents []bool
}

func New(modules common.Modules) *Stats {
	ss := &Stats{
		keys:     make([]string, stats.COUNT),
		names:    make([]string, stats.COUNT),
		descs:    make([]string, stats.COUNT),
		percents: make([]bool, stats.COUNT),
	}

	ss.Init(modules)
	return ss
}

func (this *Stats) Init(modules common.Modules) {
	msg := modules.Msg()
	eset := modules.Eset()

	this.keys[stats.HP_MAX] = "hp"
	this.names[stats.HP_MAX] = msg.Get("Max HP")
	this.descs[stats.HP_MAX] = msg.Get("Total amount of HP.")
	this.percents[stats.HP_MAX] = false

	this.keys[stats.HP_REGEN] = "hp_regen"
	this.names[stats.HP_REGEN] = msg.Get("HP Regen")
	this.descs[stats.HP_REGEN] = msg.Get("Ticks of HP regen per minute.")
	this.percents[stats.HP_REGEN] = false

	this.keys[stats.MP_MAX] = "mp"
	this.names[stats.MP_MAX] = msg.Get("Max MP")
	this.descs[stats.MP_MAX] = msg.Get("Total amount of MP.")
	this.percents[stats.MP_MAX] = false

	this.keys[stats.MP_REGEN] = "mp_regen"
	this.names[stats.MP_REGEN] = msg.Get("MP Regen")
	this.descs[stats.MP_REGEN] = msg.Get("Ticks of MP regen per minute.")
	this.percents[stats.MP_REGEN] = false

	this.keys[stats.ACCURACY] = "accuracy"
	this.names[stats.ACCURACY] = msg.Get("Accuracy")
	this.descs[stats.ACCURACY] = msg.Get("Accuracy rating. The enemy's Avoidance rating is subtracted from this value to calculate your likeliness to land a direct hit.")
	this.percents[stats.ACCURACY] = true

	this.keys[stats.AVOIDANCE] = "avoidance"
	this.names[stats.AVOIDANCE] = msg.Get("Avoidance")
	this.descs[stats.AVOIDANCE] = msg.Get("Avoidance rating. This value is subtracted from the enemy's Accuracy rating to calculate their likeliness to land a direct hit.")
	this.percents[stats.AVOIDANCE] = true

	// 伤害吸收
	this.keys[stats.ABS_MIN] = "absorb_min"
	this.names[stats.ABS_MIN] = msg.Get("Absorb Min")
	this.descs[stats.ABS_MIN] = msg.Get("Reduces the amount of damage taken.")
	this.percents[stats.ABS_MIN] = false

	this.keys[stats.ABS_MAX] = "absorb_max"
	this.names[stats.ABS_MAX] = msg.Get("Absorb Max")
	this.descs[stats.ABS_MAX] = msg.Get("Reduces the amount of damage taken.")
	this.percents[stats.ABS_MAX] = false

	this.keys[stats.CRIT] = "crit"
	this.names[stats.CRIT] = msg.Get("Critical Hit Chance")
	this.descs[stats.CRIT] = msg.Get("Chance for an attack to do extra damage.")
	this.percents[stats.CRIT] = true

	this.keys[stats.XP_GAIN] = "xp_gain"
	this.names[stats.XP_GAIN] = msg.Get("Bonus XP")
	this.descs[stats.XP_GAIN] = msg.Get("Increases the XP gained per kill.")
	this.percents[stats.XP_GAIN] = true

	// 掉钱
	this.keys[stats.CURRENCY_FIND] = "currency_find"
	this.names[stats.CURRENCY_FIND] = msg.Get(fmt.Sprintf("Bonus %s", eset.Get("loot", "currency_name").(string)))
	this.descs[stats.CURRENCY_FIND] = msg.Get("Increases the XP gained per kill.")
	this.percents[stats.CURRENCY_FIND] = true

	// 掉宝
	this.keys[stats.ITEM_FIND] = "item_find"
	this.names[stats.ITEM_FIND] = msg.Get("Item Find Chance")
	this.descs[stats.ITEM_FIND] = msg.Get("Increases the chance that an enemy will drop an item.")
	this.percents[stats.ITEM_FIND] = true

	// 不容易被怪物察觉
	this.keys[stats.STEALTH] = "stealth"
	this.names[stats.STEALTH] = msg.Get("Stealth")
	this.descs[stats.STEALTH] = msg.Get("Increases your ability to move undetected.")
	this.percents[stats.STEALTH] = true

	this.keys[stats.POISE] = "poise"
	this.names[stats.POISE] = msg.Get("Poise")
	this.descs[stats.POISE] = msg.Get("Reduces your chance of stumbling when hit.")
	this.percents[stats.POISE] = true

	this.keys[stats.REFLECT] = "reflect_chance"
	this.names[stats.REFLECT] = msg.Get("Missile Reflect Chance")
	this.descs[stats.REFLECT] = msg.Get("Increases your chance of reflecting missiles back at enemies.")
	this.percents[stats.REFLECT] = true

	this.keys[stats.RETURN_DAMAGE] = "return_damage"
	this.names[stats.RETURN_DAMAGE] = msg.Get("Damage Reflection")
	this.descs[stats.RETURN_DAMAGE] = msg.Get("Deals a percentage of damage taken back to the attacker.")
	this.percents[stats.RETURN_DAMAGE] = true

	this.keys[stats.HP_STEAL] = "hp_steal"
	this.names[stats.HP_STEAL] = msg.Get("HP Steal")
	this.descs[stats.HP_STEAL] = msg.Get("Percentage of HP stolen per hit.")
	this.percents[stats.HP_STEAL] = true

	this.keys[stats.MP_STEAL] = "mp_steal"
	this.names[stats.MP_STEAL] = msg.Get("MP Steal")
	this.descs[stats.MP_STEAL] = msg.Get("Percentage of MP stolen per hit.")
	this.percents[stats.MP_STEAL] = true

	this.keys[stats.HP_PERCENT] = "hp_percent"
	this.names[stats.HP_PERCENT] = msg.Get("Base HP")
	this.descs[stats.HP_PERCENT] = ""
	this.percents[stats.HP_PERCENT] = true

	this.keys[stats.MP_PERCENT] = "mp_percent"
	this.names[stats.MP_PERCENT] = msg.Get("Base MP")
	this.descs[stats.MP_PERCENT] = ""
	this.percents[stats.MP_PERCENT] = true

}

func (this *Stats) GetKey(key stats.STAT) string {
	return this.keys[key]
}
