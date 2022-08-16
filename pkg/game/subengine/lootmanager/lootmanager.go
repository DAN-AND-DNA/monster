package lootmanager

import (
	"fmt"
	"math"
	"monster/pkg/common"
	"monster/pkg/common/define"
	"monster/pkg/common/event"
	"monster/pkg/common/gameres"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/utils"
	"monster/pkg/utils/parsing"
)

type LootManager struct {
	lootTables map[string]([]event.Component)
	animations map[define.ItemId]([]common.Animation)
}

func New(modules common.Modules, items gameres.ItemManager) *LootManager {
	lm := &LootManager{}

	lm.init(modules, items)

	return lm
}

func (this *LootManager) init(modules common.Modules, items gameres.ItemManager) gameres.LootManager {
	this.lootTables = map[string]([]event.Component){}
	this.animations = map[define.ItemId]([]common.Animation){}

	err := this.loadGraphics(modules, items)
	if err != nil {
		panic(err)
	}

	err = this.loadLootTables(modules)
	if err != nil {
		panic(err)
	}

	return this
}

func (this *LootManager) Close(modules common.Modules) {
	anim := modules.Anim()

	for _, val := range this.animations {
		if len(val) == 0 {
			continue
		}

		for _, ptr := range val {
			anim.DecreaseCount(ptr.GetName())
		}
	}

	anim.CleanUp()
}

func (this *LootManager) loadGraphics(modules common.Modules, items gameres.ItemManager) error {
	settings := modules.Settings()
	mods := modules.Mods()
	render := modules.Render()
	mresf := modules.Resf()
	anim := modules.Anim()

	allItems := items.GetItems()

	for key, item := range allItems {
		if len(item.LootAnimation) == 0 {
			continue
		}

		this.animations[key] = make([]common.Animation, len(item.LootAnimation))
		for i, la := range item.LootAnimation {
			anim.IncreaseCount(la.Name)
			tmpSet, err := anim.GetAnimationSet(settings, mods, render, mresf, la.Name)
			if err != nil {
				return err
			}
			this.animations[key][i] = tmpSet.GetAnimation("")
		}
	}

	return nil
}

// 加载战利品的掉落信息
func (this *LootManager) loadLootTables(modules common.Modules) error {
	mods := modules.Mods()
	eset := modules.Eset()

	filenames, err := mods.List("loot")
	if err != nil {
		return err
	}

	for _, filename := range filenames {
		infile := fileparser.New()

		err := infile.Open(filename, true, mods)
		if err != nil && utils.IsNotExist(err) {
			continue
		} else if err != nil {
			return err
		}

		defer infile.Close()

		ecList := this.lootTables[filename]
		skipToNext := false
		var ec *event.Component

		for infile.Next(mods) {
			if infile.GetSection() == "" {

				// 行加载
				if infile.Key() == "loot" {
					ecList = append(ecList, event.ConstructComponent())
					ec = &(ecList[len(ecList)-1])
					ecList = this.ParseLoot(modules, infile.Val(), ec, ecList)
				}
			} else if infile.GetSection() == "loot" {

				// 节加载
				if infile.IsNewSection() {
					ecList = append(ecList, event.ConstructComponent())
					ec = &(ecList[len(ecList)-1])
					ec.Type = event.LOOT
					skipToNext = false
				}

				if skipToNext || ec == nil {
					continue
				}

				switch infile.Key() {
				case "id":
					ec.S = infile.Val()

					if ec.S == "currency" {
						ec.Id = eset.Get("misc", "currency_id").(int)
					} else if parsing.ToInt(ec.S, -1) != -1 {
						ec.Id = parsing.ToInt(ec.S, 0)
					} else {
						return fmt.Errorf("LootManager: Invalid item id for loot.")
					}
				case "chance":
					var chance string
					chance, _ = parsing.PopFirstString(infile.Val(), "")

					// 掉落概率
					if chance == "fixed" {
						ec.F = 0
					} else {
						ec.F = parsing.ToFloat(chance, 0)
					}

				case "quantity":
					// 品质的范围
					var min, max int
					var val string
					min, val = parsing.PopFirstInt(infile.Val(), "")
					ec.A = (int)(math.Max((float64)(min), 1))
					max, val = parsing.PopFirstInt(val, "")
					ec.B = (int)(math.Max((float64)(max), (float64)(ec.A)))

				}
			}
		}
	}

	return nil
}

func (this *LootManager) getLootTable(filename string, ecList []event.Component) []event.Component {
	for key, val := range this.lootTables {
		if key == filename {
			for _, ec := range val {
				ecList = append(ecList, ec)
			}
			break
		}
	}

	return ecList
}

// 单个或直接加载某个文件
// 翻译战利品的id和掉落信息
func (this *LootManager) ParseLoot(modules common.Modules, val string, e *event.Component, ecList []event.Component) []event.Component {
	eset := modules.Eset()

	if e == nil {
		return ecList
	}

	firstIsFilename := false

	e.S, val = parsing.PopFirstString(val, "")

	if e.S == "currency" {
		e.Id = eset.Get("misc", "currency_id").(int)
	} else if parsing.ToInt(e.S, -1) != -1 {
		e.Id = parsing.ToInt(e.S, 0)
	} else if ecList != nil {

		// 加载整个文件的loot
		filename := e.S

		if e == &(ecList[len(ecList)-1]) {
			ecList = ecList[:len(ecList)-1]
		}

		ecList = this.getLootTable(filename, ecList)
		firstIsFilename = true
	}

	var chance string
	if !firstIsFilename {
		e.Type = event.LOOT

		chance, val = parsing.PopFirstString(val, "")

		// 掉落概率
		if chance == "fixed" {
			e.F = 0
		} else {
			e.F = parsing.ToFloat(chance, 0)
		}

		// 品质的范围
		var min, max int
		min, val = parsing.PopFirstInt(val, "")
		e.A = (int)(math.Max((float64)(min), 1))
		max, val = parsing.PopFirstInt(val, "")
		e.B = (int)(math.Max((float64)(max), (float64)(e.A)))
	}

	// 同行的重复战利品
	if ecList != nil {
		var repeatVal string
		repeatVal, val = parsing.PopFirstString(val, "")
		for repeatVal != "" {
			ecList = append(ecList, event.ConstructComponent())
			ec := &(ecList[len(ecList)-1])
			ec.Type = event.LOOT
			ec.S = repeatVal
			if ec.S == "currency" {
				ec.Id = eset.Get("misc", "currency_id").(int)
			} else if parsing.ToInt(ec.S, -1) != -1 {
				ec.Id = parsing.ToInt(ec.S, 0)
			} else {
				ecList = ecList[:len(ecList)-1]
				ecList = this.getLootTable(repeatVal, ecList)
				repeatVal, val = parsing.PopFirstString(val, "")
				continue
			}

			chance, val = parsing.PopFirstString(val, "")

			// 掉落概率
			if chance == "fixed" {
				ec.F = 0
			} else {
				ec.F = parsing.ToFloat(chance, 0)
			}

			// 品质的范围
			var min, max int
			min, val = parsing.PopFirstInt(val, "")
			ec.A = (int)(math.Max((float64)(min), 1))
			max, val = parsing.PopFirstInt(val, "")
			ec.B = (int)(math.Max((float64)(max), (float64)(ec.A)))

			repeatVal, val = parsing.PopFirstString(val, "")
		}
	}

	return ecList
}
