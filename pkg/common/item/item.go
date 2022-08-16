package item

import (
	"math"
	"monster/pkg/common/define"
)

const (
	NO_STASH_NULL    = 0
	NO_STASH_IGNORE  = 1
	NO_STASH_PRIVATE = 2
	NO_STASH_SHARED  = 3
	NO_STASH_ALL     = 4
)

type ReplacePowerPair struct {
	First  define.PowerId
	Second define.PowerId
}

type Item struct {
	Name           string
	HasName        bool
	Flavor         string // 可选描述文字
	Level          int
	Set            define.ItemSetId
	Quality        string   // 品质
	Type           string   // 装备类型
	EquipFlags     []string // 装备物品时设置的标签
	Icon           int
	Book           string
	BookIsReadable bool
	DmgMin         []int // 武器伤害范围
	DmgMax         []int
	AbsMin         int // 吸收伤害范围
	AbsMax         int
	RequiresLevel  int   // 物品要求玩家等级
	ReqStat        []int // 物品要求属性
	ReqVal         []int
	RequiresClass  string
	Bonus          []BonusData
	//TODO sfx
	Gfx           string // 装备了该物品的动画文件
	LootAnimation []LootAnimation
	Power         define.PowerId
	ReplacePower  []ReplacePowerPair // 装备了该物品，技能替换
	PowerDesc     string
	Price         int
	PricePerLevel int
	PriceSell     int
	MaxQuantity   int // 最大数量
	PickupStatus  string
	DisableSlots  []string
	QuestItem     bool // 任务道具
	NoStash       int
	Script        string
}

func New(damageTypeNum int) *Item {
	t := Construct(damageTypeNum)
	return &t
}

func Construct(damageTypeNum int) Item {
	return Item{
		BookIsReadable: true,
		MaxQuantity:    math.MaxInt,
		DmgMin:         make([]int, damageTypeNum),
		DmgMax:         make([]int, damageTypeNum),
	}

}
