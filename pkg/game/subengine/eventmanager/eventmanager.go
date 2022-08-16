package eventmanager

import (
	"fmt"
	"math"
	"monster/pkg/common"
	"monster/pkg/common/event"
	"monster/pkg/common/fpoint"
	"monster/pkg/common/gameres"
	"monster/pkg/common/timer"
	"monster/pkg/utils/parsing"
	"monster/pkg/utils/tools"
)

type EventManager struct {
}

func New() *EventManager {
	return &EventManager{}
}

func (this *EventManager) Close() {
}

func (this *EventManager) LoadEvent(modules common.Modules, loot gameres.LootManager, camp gameres.CampaignManager, key, val string, evnt *event.Event) error {
	settings := modules.Settings()

	// 事件的元数据和具体效果内容
	switch key {
	case "type":
		evnt.Type = val
	case "activate":
		// 事件激活类型
		switch val {
		case "on_trigger":
			evnt.ActivateType = event.ACTIVATE_ON_TRIGGER
		case "on_interact":
			evnt.ActivateType = event.ACTIVATE_ON_INTERACT
		case "on_mapexit":
			evnt.ActivateType = event.ACTIVATE_ON_MAPEXIT
		case "on_leave":
			evnt.ActivateType = event.ACTIVATE_ON_LEAVE
		case "on_load":
			evnt.ActivateType = event.ACTIVATE_ON_LOAD
			evnt.KeepAfterTrigger = false
		case "static":
			evnt.ActivateType = event.ACTIVATE_STATIC
		default:
			return fmt.Errorf("EventManager: Event activation type '%s' unknown. Defaulting to 'on_trigger'.\n", val)
		}
	case "location":
		var first int
		first, val = parsing.PopFirstInt(val, "")
		evnt.Location.X = first
		first, val = parsing.PopFirstInt(val, "")
		evnt.Location.Y = first
		first, val = parsing.PopFirstInt(val, "")
		evnt.Location.W = first
		first, val = parsing.PopFirstInt(val, "")
		evnt.Location.H = first

		if evnt.Center.X == -1 && evnt.Center.Y == -1 {
			evnt.Center.X = float32(evnt.Location.X) + (float32)(evnt.Location.W)/2
			evnt.Center.Y = float32(evnt.Location.Y) + (float32)(evnt.Location.H)/2
		}
	case "hotspot":
		if val == "location" {
			evnt.Hotspot.X = evnt.Location.X
			evnt.Hotspot.Y = evnt.Location.Y
			evnt.Hotspot.W = evnt.Location.W
			evnt.Hotspot.H = evnt.Location.H
		} else {
			var first int
			first, val = parsing.PopFirstInt(val, "")
			evnt.Hotspot.X = first
			first, val = parsing.PopFirstInt(val, "")
			evnt.Hotspot.Y = first
			first, val = parsing.PopFirstInt(val, "")
			evnt.Hotspot.W = first
			first, val = parsing.PopFirstInt(val, "")
			evnt.Hotspot.H = first
		}

		evnt.Center.X = float32(evnt.Hotspot.X) + (float32)(evnt.Hotspot.W)/2
		evnt.Center.Y = float32(evnt.Hotspot.Y) + (float32)(evnt.Hotspot.H)/2

	case "cooldown":
		evnt.Cooldown.SetDuration((uint)(parsing.ToDuration(val, settings.Get("max_fps").(int))))
		evnt.Cooldown.Reset(timer.END)
	case "delay":
		evnt.Delay.SetDuration((uint)(parsing.ToDuration(val, settings.Get("max_fps").(int))))
		evnt.Delay.Reset(timer.END)
	case "reachable_from":
		var first int
		first, val = parsing.PopFirstInt(val, "")
		evnt.ReachableFrom.X = first
		first, val = parsing.PopFirstInt(val, "")
		evnt.ReachableFrom.Y = first
		first, val = parsing.PopFirstInt(val, "")
		evnt.ReachableFrom.W = first
		first, val = parsing.PopFirstInt(val, "")
		evnt.ReachableFrom.H = first

	default:
		//  具体事件效果和限制内容
		err := this.loadEventComponent(modules, loot, camp, key, val, evnt, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

// 具体事件效果内容
func (this *EventManager) loadEventComponent(modules common.Modules, loot gameres.LootManager, camp gameres.CampaignManager, key, val string, evnt *event.Event, ec *event.Component) error {
	msg := modules.Msg()
	settings := modules.Settings()

	var e *event.Component
	if evnt != nil {
		evnt.Components = append(evnt.Components, event.ConstructComponent())
		e = &(evnt.Components[len(evnt.Components)-1])
	} else if ec != nil {
		e = ec
	}

	e.Type = event.NONE

	switch key {
	case "tooltip":
		e.Type = event.TOOLTIP
		e.S = msg.Get(val)
	case "power_path":
		// 技能路径
		var first int
		var dest string

		// 起点
		e.Type = event.POWER_PATH
		first, val = parsing.PopFirstInt(val, "")
		e.X = first
		first, val = parsing.PopFirstInt(val, "")
		e.Y = first

		// 终点
		dest, val = parsing.PopFirstString(val, "")
		if dest == "hero" {
			e.S = "hero"
		} else {
			e.A = parsing.ToInt(dest, 0)
			e.B, _ = parsing.PopFirstInt(val, "")
		}

	case "power_damage":
		var first int

		e.Type = event.POWER_DAMAGE
		first, val = parsing.PopFirstInt(val, "")
		e.X = first
		first, val = parsing.PopFirstInt(val, "")
		e.Y = first

	case "intermap":
		// 跨地图传送

		var testX string
		e.Type = event.INTERMAP
		e.S, val = parsing.PopFirstString(val, "")
		e.X = -1
		e.Y = -1

		testX, val = parsing.PopFirstString(val, "")
		if testX != "" {
			e.X = parsing.ToInt(testX, 0)
			e.Y, _ = parsing.PopFirstInt(val, "")
		}

	case "intermap_random":
		// 随机跨地图传送

		e.Type = event.INTERMAP
		e.S, val = parsing.PopFirstString(val, "")
		e.Z = 1
	case "intramap":
		// 当前地图传送

		var first int

		e.Type = event.INTRAMAP
		first, val = parsing.PopFirstInt(val, "")
		e.X = first
		first, val = parsing.PopFirstInt(val, "")
		e.Y = first

	case "mapmod":
		// 事件导致的地图瓷砖变化
		var first int
		e.Type = event.MAPMOD
		e.S, val = parsing.PopFirstString(val, "")
		first, val = parsing.PopFirstInt(val, "")
		e.X = first
		first, val = parsing.PopFirstInt(val, "")
		e.Y = first
		first, val = parsing.PopFirstInt(val, "")
		e.Z = first

		if evnt != nil {
			var repeatVal string
			repeatVal, val = parsing.PopFirstString(val, "")
			for repeatVal != "" {
				evnt.Components = append(evnt.Components, event.ConstructComponent())
				e = &(evnt.Components[len(evnt.Components)-1])
				e.Type = event.MAPMOD
				e.S = repeatVal
				first, val = parsing.PopFirstInt(val, "")
				e.X = first
				first, val = parsing.PopFirstInt(val, "")
				e.Y = first
				first, val = parsing.PopFirstInt(val, "")
				e.Z = first

				repeatVal, val = parsing.PopFirstString(val, "")
			}
		}

	case "soundfx":
		// TODO
	case "loot":
		// 事件导致战利品的掉落信息变化
		e.Type = event.LOOT
		evnt.Components = loot.ParseLoot(modules, val, e, evnt.Components)
	case "loot_count":
		e.Type = event.LOOT_COUNT
		e.X, val = parsing.PopFirstInt(val, "")
		e.Y, val = parsing.PopFirstInt(val, "")
		if e.X != 0 || e.Y != 0 {
			e.X = (int)(math.Max((float64)(e.X), 1))
			e.Y = (int)(math.Max((float64)(e.Y), (float64)(e.X)))
		}

	case "msg":
		e.Type = event.MSG
		e.S = msg.Get(val)
	case "shakycam":
		// 事件导致摄像头抖动
		e.Type = event.SHAKYCAM
		maxFps := settings.Get("max_fps").(int)
		e.X = parsing.ToDuration(val, maxFps)

	case "requires_status":
		// 事件需要的状态
		e.Type = event.REQUIRES_STATUS

		var first string
		first, val = parsing.PopFirstString(val, "")
		e.Status = camp.RegisterStatus(first)

		if evnt != nil {
			first, val = parsing.PopFirstString(val, "")

			for first != "" {
				evnt.Components = append(evnt.Components, event.ConstructComponent())
				e = &(evnt.Components[len(evnt.Components)-1])
				e.Type = event.REQUIRES_STATUS
				e.Status = camp.RegisterStatus(first)

				first, val = parsing.PopFirstString(val, "")
			}
		}

	case "requires_not_status":
		e.Type = event.REQUIRES_NOT_STATUS

		var first string
		first, val = parsing.PopFirstString(val, "")
		e.Status = camp.RegisterStatus(first)

		if evnt != nil {
			first, val = parsing.PopFirstString(val, "")

			for first != "" {
				evnt.Components = append(evnt.Components, event.ConstructComponent())
				e = &(evnt.Components[len(evnt.Components)-1])
				e.Type = event.REQUIRES_NOT_STATUS
				e.Status = camp.RegisterStatus(first)

				first, val = parsing.PopFirstString(val, "")
			}
		}
	case "requires_level":
		e.Type = event.REQUIRES_LEVEL
		e.X, val = parsing.PopFirstInt(val, "")
	case "requires_not_level":
		e.Type = event.REQUIRES_NOT_LEVEL
		e.X, val = parsing.PopFirstInt(val, "")

	case "requires_currency":
		// 事件需要至少多少钱
		e.Type = event.REQUIRES_CURRENCY
		e.X, val = parsing.PopFirstInt(val, "")
	case "requires_not_currency":
		e.Type = event.REQUIRES_NOT_CURRENCY
		e.X, val = parsing.PopFirstInt(val, "")

	case "requires_item":
		// 事件需要的物品
		e.Type = event.REQUIRES_ITEM

		var first string
		first, val = parsing.PopFirstString(val, "")
		itemStack, _ := parsing.ToItemQuantityPair(first)
		e.Id = (int)(itemStack.Item)
		e.X = (int)(itemStack.Quantity)

		if evnt != nil {
			first, val = parsing.PopFirstString(val, "")

			for first != "" {
				evnt.Components = append(evnt.Components, event.ConstructComponent())
				e = &(evnt.Components[len(evnt.Components)-1])
				e.Type = event.REQUIRES_ITEM
				itemStack, _ = parsing.ToItemQuantityPair(first)
				e.Id = (int)(itemStack.Item)
				e.X = (int)(itemStack.Quantity)

				first, val = parsing.PopFirstString(val, "")
			}

		}

	case "requires_not_item":
		e.Type = event.REQUIRES_NOT_ITEM

		var first string
		first, val = parsing.PopFirstString(val, "")
		itemStack, _ := parsing.ToItemQuantityPair(first)
		e.Id = (int)(itemStack.Item)
		e.X = (int)(itemStack.Quantity)

		if evnt != nil {
			first, val = parsing.PopFirstString(val, "")

			for first != "" {
				evnt.Components = append(evnt.Components, event.ConstructComponent())
				e = &(evnt.Components[len(evnt.Components)-1])
				e.Type = event.REQUIRES_NOT_ITEM
				itemStack, _ = parsing.ToItemQuantityPair(first)
				e.Id = (int)(itemStack.Item)
				e.X = (int)(itemStack.Quantity)

				first, val = parsing.PopFirstString(val, "")
			}

		}
	case "requires_class":
		// 需要职业
		e.Type = event.REQUIRES_CLASS
		e.S, val = parsing.PopFirstString(val, "")
	case "requires_not_class":
		e.Type = event.REQUIRES_NOT_CLASS
		e.S, val = parsing.PopFirstString(val, "")

	case "set_status":
		// 事件设置的状态
		e.Type = event.SET_STATUS

		var first string
		first, val = parsing.PopFirstString(val, "")
		e.Status = camp.RegisterStatus(first)

		if evnt != nil {
			first, val = parsing.PopFirstString(val, "")

			for first != "" {
				evnt.Components = append(evnt.Components, event.ConstructComponent())
				e = &(evnt.Components[len(evnt.Components)-1])
				e.Type = event.SET_STATUS
				e.Status = camp.RegisterStatus(first)

				first, val = parsing.PopFirstString(val, "")
			}
		}
	case "unset_status":
		e.Type = event.UNSET_STATUS

		var first string
		first, val = parsing.PopFirstString(val, "")
		e.Status = camp.RegisterStatus(first)

		if evnt != nil {
			first, val = parsing.PopFirstString(val, "")

			for first != "" {
				evnt.Components = append(evnt.Components, event.ConstructComponent())
				e = &(evnt.Components[len(evnt.Components)-1])
				e.Type = event.UNSET_STATUS
				e.Status = camp.RegisterStatus(first)

				first, val = parsing.PopFirstString(val, "")
			}
		}

	case "remove_currency":
		// 事件导致扣钱
		e.Type = event.REMOVE_CURRENCY
		e.X = (int)(math.Max((float64)(parsing.ToInt(val, 0)), 0))

	case "remove_item":
		// 事件导致移除物品
		e.Type = event.REMOVE_ITEM

		var first string
		first, val = parsing.PopFirstString(val, "")
		itemStack, _ := parsing.ToItemQuantityPair(first)
		e.Id = (int)(itemStack.Item)
		e.X = (int)(itemStack.Quantity)

		if evnt != nil {
			first, val = parsing.PopFirstString(val, "")

			for first != "" {
				evnt.Components = append(evnt.Components, event.ConstructComponent())
				e = &(evnt.Components[len(evnt.Components)-1])
				e.Type = event.REMOVE_ITEM
				itemStack, _ = parsing.ToItemQuantityPair(first)
				e.Id = (int)(itemStack.Item)
				e.X = (int)(itemStack.Quantity)

				first, val = parsing.PopFirstString(val, "")
			}
		}

	case "reward_xp":
		e.Type = event.REWARD_XP
		e.X = (int)(math.Max((float64)(parsing.ToInt(val, 0)), 0))
	case "reward_currency":
		e.Type = event.REWARD_CURRENCY
		e.X = (int)(math.Max((float64)(parsing.ToInt(val, 0)), 0))
	case "reward_item":
		e.Type = event.REWARD_ITEM

		var first string
		first, val = parsing.PopFirstString(val, "")
		itemStack, checkPair := parsing.ToItemQuantityPair(first)

		if !checkPair {
			e.Id = (int)(itemStack.Item)
			rawX, _ := parsing.PopFirstInt(val, "")
			e.X = (int)(math.Max((float64)(rawX), 1))

		} else {
			e.Id = (int)(itemStack.Item)
			e.X = (int)(itemStack.Quantity)

			if evnt != nil {
				first, val = parsing.PopFirstString(val, "")

				for first != "" {
					evnt.Components = append(evnt.Components, event.ConstructComponent())
					e = &(evnt.Components[len(evnt.Components)-1])
					e.Type = event.REWARD_ITEM
					itemStack, _ = parsing.ToItemQuantityPair(first)
					e.Id = (int)(itemStack.Item)
					e.X = (int)(itemStack.Quantity)

					first, val = parsing.PopFirstString(val, "")
				}
			}
		}

	case "reward_loot":
		e.Type = event.REWARD_LOOT
		e.S = val
	case "reward_loot_count":
		e.Type = event.REWARD_LOOT_COUNT
		var first int
		first, val = parsing.PopFirstInt(val, "")
		e.X = (int)(math.Max((float64)(first), 1))
		first, val = parsing.PopFirstInt(val, "")
		e.Y = (int)(math.Max((float64)(first), (float64)(e.X)))
	case "restore":
		e.Type = event.RESTORE
		e.S = val
	case "power":
		// 事件触发指定的技能
		e.Type = event.POWER
		e.Id = parsing.ToInt(val, 0)
	case "spawn":
		// 刷出怪物
		e.Type = event.SPAWN
		e.S, val = parsing.PopFirstString(val, "")
		e.X, val = parsing.PopFirstInt(val, "")
		e.Y, val = parsing.PopFirstInt(val, "")

		if evnt != nil {
			var first string
			first, val = parsing.PopFirstString(val, "")

			for first != "" {
				evnt.Components = append(evnt.Components, event.ConstructComponent())
				e = &(evnt.Components[len(evnt.Components)-1])
				e.Type = event.SPAWN
				e.S = first
				e.X, val = parsing.PopFirstInt(val, "")
				e.Y, val = parsing.PopFirstInt(val, "")

				first, val = parsing.PopFirstString(val, "")
			}
		}

	case "stash":
		e.Type = event.STASH
		if parsing.ToBool(val) {
			e.X = 1
		} else {
			e.X = 0
		}
	case "npc":
		e.Type = event.NPC
		e.S = val
	case "music":
		// TODO
	case "cutscene":
		e.Type = event.CUTSCENE
		e.S = val

	case "repeat":
		// 事件触发后会重复执行
		e.Type = event.REPEAT
		if parsing.ToBool(val) {
			e.X = 1
		} else {
			e.X = 0
		}

	case "save_game":
		// 事件导致保存游戏
		e.Type = event.SAVE_GAME
		if parsing.ToBool(val) {
			e.X = 1
		} else {
			e.X = 0
		}

	case "book":
		e.Type = event.BOOK
		e.S = val
	case "script":
		e.Type = event.SCRIPT
		e.S = val
	case "chance_exec":
		// 事件执行概率
		e.Type = event.CHANCE_EXEC
		e.X, _ = parsing.PopFirstInt(val, "")
	case "respec":
		// 重置玩家
		e.Type = event.RESPEC
		var mode, useEngineDefaults string
		mode, val = parsing.PopFirstString(val, "")
		useEngineDefaults, val = parsing.PopFirstString(val, "")

		if mode == "xp" {
			e.X = 3
		} else if mode == "stats" {
			e.X = 2
		} else if mode == "powers" {
			e.X = 1
		}

		// 使用默认值
		if useEngineDefaults != "" {
			if parsing.ToBool(useEngineDefaults) {
				e.Y = 1
			} else {
				e.Y = 0
			}
		}
	case "show_on_minimap":
		e.Type = event.SHOW_ON_MINIMAP
		if parsing.ToBool(val) {
			e.X = 1
		} else {
			e.X = 0
		}

	case "parallax_layers":
		e.Type = event.PARALLAX_LAYERS
		e.S = val
	default:
		fmt.Errorf("EventManager: '%s' is not a valid key.\n", key)
	}

	return nil
}

func (this *EventManager) ExecuteEvent(modules common.Modules, mapr gameres.MapRenderer, camp gameres.CampaignManager, e *event.Event) bool {
	return this.executeEventInternal(modules, mapr, camp, e, false)
}

func (this *EventManager) ExecteDelayedEvent(modules common.Modules, mapr gameres.MapRenderer, camp gameres.CampaignManager, e *event.Event) bool {
	return this.executeEventInternal(modules, mapr, camp, e, true)
}

// 事件已经触发，在此执行该事件的全部组件
// 返回true表示事件结束不再运行
func (this *EventManager) executeEventInternal(modules common.Modules, mapr gameres.MapRenderer, camp gameres.CampaignManager, ev *event.Event, skipDelay bool) bool {

	settings := modules.Settings()
	mods := modules.Mods()

	// 跳过还在冷却的
	if !ev.Delay.IsEnd() || !ev.Cooldown.IsEnd() {
		return false
	}

	// 触发后重复执行
	ecRepeat, ok := ev.GetComponent(event.REPEAT)
	if ok {
		if ecRepeat.X == 0 {
			ev.KeepAfterTrigger = false
		} else {
			ev.KeepAfterTrigger = true
		}
	}

	// 注册延迟事件
	if ev.Delay.GetDuration() > 0 && !skipDelay {
		ev.Delay.Reset(timer.BEGIN)
		// 深拷贝一个
		mapr.RegisterDelayedEvent(*ev.DeepCopy()) //
		ev.Cooldown.Reset(timer.BEGIN)
		return !ev.KeepAfterTrigger
	}

	ev.Cooldown.Reset(timer.BEGIN)
	ecChanceExec, ok := ev.GetComponent(event.CHANCE_EXEC)

	// 概率失败不执行
	if ok && !tools.PercentChance(ecChanceExec.X) {
		return !ev.KeepAfterTrigger
	}

	for index, _ := range ev.Components {
		ec := &(ev.Components[index])

		switch ec.Type {
		case event.SET_STATUS:
			// 需要启动状态
			camp.SetStatus(ec.Status)
		case event.INTERMAP:
			if ec.Z == 1 {
				// TODO
				// 随机位置
			}

			_, err := mods.Locate(settings, ec.S)
			if err != nil {
				panic(err)
			}

			mapr.SetTeleportation(true)
			mapr.SetTeleportMapName(ec.S)

			if ec.X == -1 && ec.Y == -1 {
				mapr.SetTeleportDestination(fpoint.Construct(-1, -1))
			} else {
				mapr.SetTeleportDestination(fpoint.Construct(float32(ec.X)+0.5, float32(ec.Y)+0.5))
			}
		}
	}

	return !ev.KeepAfterTrigger
}
