package item

import "monster/pkg/common/define"

type BonusData struct {
	StatIndex      int // 在stats里定义
	DamageIndexMin int // 在engine/damage_types.txt定义
	DamageIndexMax int // engine/damage_types.txt
	ResistIndex    int // engine/elements.txt
	BaseIndex      int
	IsSpeed        bool
	IsAttackSpeed  bool
	Value          int
	PowerId        define.PowerId
}

func ConstructBonusData() BonusData {
	return BonusData{
		StatIndex:      -1,
		DamageIndexMin: -1,
		DamageIndexMax: -1,
		ResistIndex:    -1,
		BaseIndex:      -1,
	}
}
