package stats

const (
	COUNT = 20
)

type STAT uint8

// 子属性
const (
	HP_MAX STAT = iota
	HP_REGEN
	MP_MAX
	MP_REGEN
	ACCURACY
	AVOIDANCE
	ABS_MIN
	ABS_MAX
	CRIT
	XP_GAIN
	CURRENCY_FIND
	ITEM_FIND
	STEALTH
	POISE
	REFLECT
	RETURN_DAMAGE
	HP_STEAL
	MP_STEAL
	// HP_PERCENT & MP_PERCENT aren't displayed in MenuCharacter;
	// new stats should be added above this comment.
	HP_PERCENT
	MP_PERCENT
)
