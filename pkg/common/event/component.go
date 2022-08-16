package event

import "monster/pkg/common/define"

// 组件类型
const (
	NONE                     = 0
	TOOLTIP                  = 1
	POWER_PATH               = 2
	POWER_DAMAGE             = 3
	INTERMAP                 = 4
	INTRAMAP                 = 5
	MAPMOD                   = 6
	SOUNDFX                  = 7
	LOOT                     = 8
	LOOT_COUNT               = 9
	MSG                      = 10
	SHAKYCAM                 = 11
	REQUIRES_STATUS          = 12
	REQUIRES_NOT_STATUS      = 13
	REQUIRES_LEVEL           = 14
	REQUIRES_NOT_LEVEL       = 15
	REQUIRES_CURRENCY        = 16
	REQUIRES_NOT_CURRENCY    = 17
	REQUIRES_ITEM            = 18
	REQUIRES_NOT_ITEM        = 19
	REQUIRES_CLASS           = 20
	REQUIRES_NOT_CLASS       = 21
	SET_STATUS               = 22
	UNSET_STATUS             = 23
	REMOVE_CURRENCY          = 24
	REMOVE_ITEM              = 25
	REWARD_XP                = 26
	REWARD_CURRENCY          = 27
	REWARD_ITEM              = 28
	REWARD_LOOT              = 29
	REWARD_LOOT_COUNT        = 30
	RESTORE                  = 31
	POWER                    = 32
	SPAWN                    = 33
	STASH                    = 34
	NPC                      = 35
	MUSIC                    = 36
	CUTSCENE                 = 37
	REPEAT                   = 38
	SAVE_GAME                = 39
	BOOK                     = 40
	SCRIPT                   = 41
	CHANCE_EXEC              = 42
	RESPEC                   = 43
	SHOW_ON_MINIMAP          = 44
	PARALLAX_LAYERS          = 45
	NPC_ID                   = 46
	NPC_HOTSPOT              = 47
	NPC_DIALOG_THEM          = 48
	NPC_DIALOG_YOU           = 49
	NPC_VOICE                = 50
	NPC_DIALOG_TOPIC         = 51
	NPC_DIALOG_GROUP         = 52
	NPC_DIALOG_ID            = 53
	NPC_DIALOG_RESPONSE      = 54
	NPC_DIALOG_RESPONSE_ONLY = 55
	NPC_ALLOW_MOVEMENT       = 56
	NPC_PORTRAIT_THEM        = 57
	NPC_PORTRAIT_YOU         = 58
	QUEST_TEXT               = 59
	WAS_INSIDE_EVENT_AREA    = 60
	NPC_TAKE_A_PARTY         = 61
)

type Component struct {
	S                      string
	Status                 define.StatusId
	Type, X, Y, Z, A, B, C int
	F                      float32
	Id                     int
}

func ConstructComponent() Component {
	return Component{}
}
