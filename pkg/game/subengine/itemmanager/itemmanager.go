package itemmanager

import (
	"fmt"
	"math"
	"monster/pkg/common"
	"monster/pkg/common/define"
	"monster/pkg/common/define/game/stats"
	"monster/pkg/common/gameres"
	"monster/pkg/common/item"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/utils/parsing"
)

type ItemManager struct {
	items         map[define.ItemId]*item.Item
	itemTypes     []item.Type
	itemSets      map[define.ItemSetId]*item.Set
	itemQualities []item.Quality
}

func New(modules common.Modules, stats gameres.Stats) *ItemManager {
	im := &ItemManager{}
	im.init(modules, stats)

	return im
}

func (this *ItemManager) init(modules common.Modules, stats gameres.Stats) gameres.ItemManager {
	this.items = map[define.ItemId]*item.Item{}
	this.itemSets = map[define.ItemSetId]*item.Set{}

	err := this.loadAll(modules, stats)
	if err != nil {
		panic(err)
	}

	return this
}

func (this *ItemManager) loadAll(modules common.Modules, stats gameres.Stats) error {
	err := this.loadItems(modules, stats, "items/items.txt")
	if err != nil {
		return err
	}

	err = this.loadType(modules, "items/types.txt")
	if err != nil {
		return err
	}

	err = this.loadSets(modules, stats, "items/sets.txt")
	if err != nil {
		return err
	}

	err = this.loadQualities(modules, "items/qualities.txt")
	if err != nil {
		return err
	}

	if len(this.items) == 0 {
		return fmt.Errorf("ItemManager: No items were found.")
	}

	return nil
}

// 加载物品属性
func (this *ItemManager) loadItems(modules common.Modules, stats gameres.Stats, filename string) error {
	mods := modules.Mods()
	eset := modules.Eset()
	msg := modules.Msg()

	infile := fileparser.New()

	err := infile.Open(filename, true, mods)
	if err != nil {
		return err
	}
	defer infile.Close()

	clearReqStat := true
	clearBonus := true
	clearLootAnim := true
	clearReplacePower := true

	dtList := eset.Get("damage_types", "list").([]common.DamageType)
	idLine := false

	id := (define.ItemId)(0)
	for infile.Next(mods) {
		key := infile.Key()
		val := infile.Val()
		if key == "id" {
			idLine = true
			id = (define.ItemId)(parsing.ToInt(val, 0))
			this.items[id] = item.New(len(dtList))
			if this.items[id].MaxQuantity == math.MaxInt {
				this.items[id].MaxQuantity = 1
			}

			clearReqStat = true
			clearBonus = true
			clearLootAnim = true
			clearReplacePower = true

		} else {
			idLine = false
		}

		if id < 1 {
			if idLine {
				return fmt.Errorf("ItemManager: Item index out of bounds 1-%d, skipping item.\n", math.MaxInt)
			}
		}

		if idLine {
			continue
		}

		switch key {
		case "name":
			this.items[id].Name = msg.Get(val)
			this.items[id].HasName = true
		case "flavor":
			this.items[id].Flavor = msg.Get(val)
		case "level":
			this.items[id].Level = parsing.ToInt(val, 0)
		case "icon":
			this.items[id].Icon = parsing.ToInt(val, 0)
		case "book":
			this.items[id].Book = val
		case "book_is_readable":
			this.items[id].BookIsReadable = parsing.ToBool(val)
		case "quality":
			this.items[id].Quality = val
		case "item_type":
			this.items[id].Type = val
		case "equip_flags":
			this.items[id].EquipFlags = nil
			var flag string

			flag, val = parsing.PopFirstString(val, "")
			for flag != "" {
				this.items[id].EquipFlags = append(this.items[id].EquipFlags, flag)
				flag, val = parsing.PopFirstString(val, "")
			}
		case "dmg":
			var dmgTypeStr string

			dmgTypeStr, val = parsing.PopFirstString(val, "") // 伤害类型
			dmgType := len(dtList)

			for i, ptr := range dtList {
				if dmgTypeStr == ptr.GetId() {
					dmgType = i
					break
				}
			}

			if dmgType == len(dtList) {
				return fmt.Errorf("ItemManager: '%s' is not a known damage type id.\n", dmgTypeStr)
			} else {
				this.items[id].DmgMin[dmgType], val = parsing.PopFirstInt(val, "")
				if val != "" {
					this.items[id].DmgMin[dmgType], val = parsing.PopFirstInt(val, "")
				} else {
					this.items[id].DmgMin[dmgType] = this.items[id].DmgMax[dmgType]
				}
			}

		case "abs":
			this.items[id].AbsMin, val = parsing.PopFirstInt(val, "")
			if val != "" {
				this.items[id].AbsMax, val = parsing.PopFirstInt(val, "")
			} else {
				this.items[id].AbsMax = this.items[id].AbsMin
			}

		case "requires_level":
			this.items[id].RequiresLevel = parsing.ToInt(val, 0)
		case "requires_stat":
			if clearReqStat {
				this.items[id].ReqStat = nil
				this.items[id].ReqVal = nil
				clearReqStat = false
			}
			var s string
			s, val = parsing.PopFirstString(val, "")

			reqStatIndex, ok := eset.PrimaryStatsGetIndexById(s)
			if ok {
				this.items[id].ReqStat = append(this.items[id].ReqStat, reqStatIndex)
			} else {
				return fmt.Errorf("ItemManager: '%s' is not a valid primary stat.\n", s)
			}

			reqVal, _ := parsing.PopFirstInt(val, "")
			this.items[id].ReqVal = append(this.items[id].ReqVal, reqVal)

		case "requires_class":
			this.items[id].RequiresClass = val
		case "bonus":
			// 加成
			if clearBonus {
				this.items[id].Bonus = nil
				clearBonus = false
			}

			bdata := item.ConstructBonusData()
			_, err := this.parseBonus(modules, stats, key, val, &bdata)
			if err != nil {
				return err
			}
			this.items[id].Bonus = append(this.items[id].Bonus, bdata)
		case "bonus_power_level":

			var first string
			first, val = parsing.PopFirstString(val, "")
			bdata := item.ConstructBonusData()

			bdata.PowerId = (define.PowerId)(parsing.ToInt(first, 0))
			bdata.Value, val = parsing.PopFirstInt(val, "")
			this.items[id].Bonus = append(this.items[id].Bonus, bdata)

		case "soundfx":
			//TODO

		case "gfx":
			this.items[id].Gfx = val
		case "loot_animation":
			if clearLootAnim {
				this.items[id].LootAnimation = nil
				clearLootAnim = false
			}

			la := item.ConstructLootAnimation()
			la.Name, val = parsing.PopFirstString(val, "")
			la.Low, val = parsing.PopFirstInt(val, "")
			la.Hight, val = parsing.PopFirstInt(val, "")
			this.items[id].LootAnimation = append(this.items[id].LootAnimation, la)

		case "power":
			if parsing.ToInt(val, 0) > 0 {
				this.items[id].Power = (define.PowerId)(parsing.ToInt(val, 0))
			} else {
				return fmt.Errorf("ItemManager: Power index out of bounds 1-%d, skipping power.\n", math.MaxInt)
			}

		case "replace_power":

			if clearReplacePower {
				this.items[id].ReplacePower = nil
				clearReplacePower = false
			}

			powerIds := item.ReplacePowerPair{}

			var first string
			first, val = parsing.PopFirstString(val, "")
			powerIds.First = (define.PowerId)(parsing.ToInt(first, 0))
			first, val = parsing.PopFirstString(val, "")
			powerIds.Second = (define.PowerId)(parsing.ToInt(first, 0))

			this.items[id].ReplacePower = append(this.items[id].ReplacePower, powerIds)

		case "power_desc":
			this.items[id].PowerDesc = msg.Get(val)
		case "price":
			this.items[id].Price = parsing.ToInt(val, 0)
		case "price_per_level":
			this.items[id].PricePerLevel = parsing.ToInt(val, 0)
		case "price_sell":
			this.items[id].PriceSell = parsing.ToInt(val, 0)
		case "max_quantity":
			this.items[id].MaxQuantity = parsing.ToInt(val, 0)
		case "pickup_status":
			this.items[id].PickupStatus = val
		case "stepfx":
			//TODO
		case "disable_slots":
			this.items[id].DisableSlots = nil
			var slotType string
			slotType, val = parsing.PopFirstString(val, "")

			for slotType != "" {
				this.items[id].DisableSlots = append(this.items[id].DisableSlots, slotType)
				slotType, val = parsing.PopFirstString(val, "")
			}
		case "quest_item":
			this.items[id].QuestItem = parsing.ToBool(val)
			if this.items[id].NoStash == item.NO_STASH_NULL {
				this.items[id].NoStash = item.NO_STASH_ALL
			}

		case "no_stash":
			var temp string
			temp, val = parsing.PopFirstString(val, "")

			switch temp {
			case "ignore":
				this.items[id].NoStash = item.NO_STASH_IGNORE
			case "private":
				this.items[id].NoStash = item.NO_STASH_PRIVATE
			case "shared":
				this.items[id].NoStash = item.NO_STASH_SHARED
			case "all":
				this.items[id].NoStash = item.NO_STASH_ALL
			default:
				return fmt.Errorf("ItemManager: '%s' is not a valid value for 'no_stash'. Use 'ignore', 'private', 'shared', or 'all'.\n", temp)
			}

		case "script":
			this.items[id].Script, _ = parsing.PopFirstString(val, "")

		default:
			return fmt.Errorf("ItemManager: '%s' is not a valid key.\n", key)
		}
	}

	// 普通物品可以存储和丢弃
	for _, ptr := range this.items {
		if ptr.NoStash == item.NO_STASH_NULL {
			ptr.NoStash = item.NO_STASH_IGNORE
		}
	}

	return nil
}

// 加成
func (this *ItemManager) parseBonus(modules common.Modules, ss gameres.Stats, key, val string, bdata *item.BonusData) (string, error) {
	eset := modules.Eset()
	dtList := eset.Get("damage_types", "list").([]common.DamageType)
	eList := eset.Get("elements", "list").([]common.Element)
	pList := eset.Get("primary_stats", "list").([]common.PrimaryStat)

	var bonusStr string
	bonusStr, val = parsing.PopFirstString(val, "")
	bdata.Value, val = parsing.PopFirstInt(val, "")

	if bonusStr == "speed" {
		bdata.IsSpeed = true
		return val, nil
	} else if bonusStr == "attack_speed" {
		bdata.IsAttackSpeed = true
		return val, nil
	}

	// 子属性
	for i := 0; i < stats.COUNT; i++ {
		if bonusStr == ss.GetKey((stats.STAT)(i)) {
			bdata.StatIndex = i
			return val, nil
		}
	}

	// 攻击
	for i, ptr := range dtList {
		if bonusStr == ptr.GetMin() {
			bdata.DamageIndexMin = i
			return val, nil
		} else if bonusStr == ptr.GetMax() {
			bdata.DamageIndexMax = i
			return val, nil
		}
	}

	// 抗性
	for i, ptr := range eList {
		if bonusStr == ptr.GetId()+"_resist" {
			bdata.ResistIndex = i
			return val, nil
		}
	}

	// 主属性

	for i, ptr := range pList {
		if bonusStr == ptr.GetId() {
			bdata.BaseIndex = i
			return val, nil
		}
	}

	fmt.Println("-----" + bonusStr)
	return val, fmt.Errorf("ItemManager: Unknown bonus type '%s'.\n", bonusStr)
}

// 加载物品类型，装在哪个位置
func (this *ItemManager) loadType(modules common.Modules, filename string) error {
	mods := modules.Mods()

	infile := fileparser.New()

	err := infile.Open(filename, true, mods)
	if err != nil {
		return err
	}
	defer infile.Close()

	for infile.Next(mods) {
		key := infile.Key()
		val := infile.Val()

		if infile.IsNewSection() {
			if infile.GetSection() == "type" {
				if len(this.itemTypes) != 0 && this.itemTypes[len(this.itemTypes)-1].Id == "" {
					this.itemTypes = this.itemTypes[:len(this.itemTypes)-1]
				}

				this.itemTypes = append(this.itemTypes, item.ConstructType())
			}
		}

		if len(this.itemTypes) == 0 || infile.GetSection() != "type" {
			continue
		}

		switch key {
		case "id":
			this.itemTypes[len(this.itemTypes)-1].Id = val
		case "name":
			this.itemTypes[len(this.itemTypes)-1].Name = val
		default:
			return fmt.Errorf("ItemManager: '%s' is not a valid key.\n", key)
		}
	}

	if len(this.itemTypes) != 0 && this.itemTypes[len(this.itemTypes)-1].Id == "" {
		this.itemTypes = this.itemTypes[:len(this.itemTypes)-1]
	}

	return nil
}

// 套装
func (this *ItemManager) loadSets(modules common.Modules, stats gameres.Stats, filename string) error {
	mods := modules.Mods()
	msg := modules.Msg()

	infile := fileparser.New()

	err := infile.Open(filename, true, mods)
	if err != nil {
		return err
	}
	defer infile.Close()

	clearBonus := true
	_ = clearBonus

	id := define.ItemSetId(0)
	idLine := false

	for infile.Next(mods) {
		key := infile.Key()
		val := infile.Val()

		if key == "id" {
			idLine = true
			id = (define.ItemSetId)(parsing.ToInt(val, 0))

			if id > 0 {
				this.itemSets[id] = item.NewSet()
			}

			clearBonus = true

		} else {
			idLine = false
		}

		if id < 1 {
			return fmt.Errorf("ItemManager: Item set index out of bounds 1-%d, skipping set.\n", math.MaxInt)
		}

		if idLine {
			continue
		}

		switch key {
		case "name":
			this.itemSets[id].Name = msg.Get(val)
		case "items":
			this.itemSets[id].Items = nil
			var first string
			first, val = parsing.PopFirstString(val, "")
			for first != "" {
				itemId := (define.ItemId)(parsing.ToInt(first, 0))
				//fmt.Println(itemId, len(this.items))
				this.items[itemId].Set = id
				this.itemSets[id].Items = append(this.itemSets[id].Items, itemId)
				first, val = parsing.PopFirstString(val, "")
			}
		case "color":
			this.itemSets[id].Color = parsing.ToRGB(val)
		case "bonus":
			if clearBonus {
				this.itemSets[id].Bonus = nil
				clearBonus = false
			}

			bonus := item.ConstructSetBonusData()
			bonus.Requirement, val = parsing.PopFirstInt(val, "")
			this.parseBonus(modules, stats, key, val, &(bonus.BonusData))
			this.itemSets[id].Bonus = append(this.itemSets[id].Bonus, bonus)

		case "bonus_power_level":
			bonus := item.ConstructSetBonusData()
			bonus.Requirement, val = parsing.PopFirstInt(val, "")

			var first string
			first, val = parsing.PopFirstString(val, "")
			bonus.PowerId = (define.PowerId)(parsing.ToInt(first, 0))
			bonus.Value, val = parsing.PopFirstInt(val, "")

			this.itemSets[id].Bonus = append(this.itemSets[id].Bonus, bonus)
		default:
			return fmt.Errorf("ItemManager: '%s' is not a valid key.\n", key)
		}
	}

	return nil
}

// 加载品质
func (this *ItemManager) loadQualities(modules common.Modules, filename string) error {
	mods := modules.Mods()

	infile := fileparser.New()

	err := infile.Open(filename, true, mods)
	if err != nil {
		return err
	}
	defer infile.Close()

	for infile.Next(mods) {
		key := infile.Key()
		_ = key
		val := infile.Val()
		_ = val

		if infile.IsNewSection() {
			if infile.GetSection() == "quality" {
				if len(this.itemQualities) != 0 && this.itemQualities[len(this.itemQualities)-1].Id == "" {
					this.itemQualities = this.itemQualities[:len(this.itemQualities)-1]
				}
				this.itemQualities = append(this.itemQualities, item.ConstructQuality())
			}
		}

		if len(this.itemQualities) == 0 || infile.GetSection() != "quality" {
			continue
		}

		switch key {
		case "id":
			this.itemQualities[len(this.itemQualities)-1].Id = val
		case "name":
			this.itemQualities[len(this.itemQualities)-1].Name = val
		case "color":
			this.itemQualities[len(this.itemQualities)-1].Color = parsing.ToRGB(val)
		case "overlay_icon":
			this.itemQualities[len(this.itemQualities)-1].OverlayIcon = parsing.ToInt(val, 0)
		default:
			return fmt.Errorf("ItemManager: '%s' is not a valid key.\n", key)
		}
	}

	if len(this.itemQualities) != 0 && this.itemQualities[len(this.itemQualities)-1].Id == "" {
		this.itemQualities = this.itemQualities[:len(this.itemQualities)-1]
	}

	return nil
}

func (this *ItemManager) GetItems() map[define.ItemId]item.Item {
	tmp := make(map[define.ItemId]item.Item, len(this.items))

	for k, ptr := range this.items {
		item := *ptr
		tmp[k] = item
	}

	return tmp
}
