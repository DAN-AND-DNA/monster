package power

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
