package state

import (
	"monster/pkg/common"
	"monster/pkg/common/gameres"
	"monster/pkg/game/menu"
	"monster/pkg/game/resources"
	"monster/pkg/game/subengine/avatar"
	"monster/pkg/game/subengine/campaignmanager"
	"monster/pkg/game/subengine/eventmanager"
	"monster/pkg/game/subengine/itemmanager"
	"monster/pkg/game/subengine/lootmanager"
	"monster/pkg/game/subengine/maprenderer"
	"monster/pkg/game/subengine/menumanager"
	"monster/pkg/game/subengine/powermanager"
	"monster/pkg/game/subengine/stats"
)

type GameRes struct {
	stats        gameres.Stats
	items        gameres.ItemManager
	camp         gameres.CampaignManager
	eventManager gameres.EventManager
	loot         gameres.LootManager
	menu         gameres.MenuManager
	pc           gameres.Avatar
	mapr         gameres.MapRenderer
	powers       gameres.PowerManager
	menuAct      gameres.MenuActionBar
}

func NewGameRes() *GameRes {
	return &GameRes{}
}

func (this *GameRes) Stats() gameres.Stats {
	return this.stats
}

func (this *GameRes) NewStats(modules common.Modules) gameres.Stats {
	this.stats = stats.New(modules)
	return this.stats
}

func (this *GameRes) Items() gameres.ItemManager {
	return this.items
}

func (this *GameRes) NewItems(modules common.Modules, stats gameres.Stats) gameres.ItemManager {
	this.items = itemmanager.New(modules, stats)
	return this.items
}

func (this *GameRes) Camp() gameres.CampaignManager {
	return this.camp
}

func (this *GameRes) NewCamp() gameres.CampaignManager {
	this.camp = campaignmanager.New()
	return this.camp
}

func (this *GameRes) EventManager() gameres.EventManager {
	return this.eventManager
}

func (this *GameRes) NewEventManager() gameres.EventManager {
	this.eventManager = eventmanager.New()
	return this.eventManager
}

func (this *GameRes) Loot() gameres.LootManager {
	return this.loot
}

func (this *GameRes) NewLoot(modules common.Modules, items gameres.ItemManager) gameres.LootManager {
	this.loot = lootmanager.New(modules, items)
	return this.loot
}

func (this *GameRes) Menu() gameres.MenuManager {
	return this.menu
}

func (this *GameRes) NewMenu(modules common.Modules, pc gameres.Avatar, powers gameres.PowerManager, menuf gameres.Factory) gameres.MenuManager {
	this.menu = menumanager.New(modules, pc, powers, menuf)
	return this.menu
}

func (this *GameRes) Pc() gameres.Avatar {
	return this.pc
}

func (this *GameRes) NewPc(modules common.Modules, mapr gameres.MapRenderer, ss gameres.Stats, powers gameres.PowerManager, gresf gameres.Factory) gameres.Avatar {
	this.pc = avatar.New(modules, mapr, ss, powers, gresf)
	return this.pc
}

func (this *GameRes) Mapr() gameres.MapRenderer {
	return this.mapr
}

func (this *GameRes) NewMapr(modules common.Modules, resf gameres.Factory) gameres.MapRenderer {
	this.mapr = maprenderer.New(modules, resf)
	return this.mapr
}

func (this *GameRes) Powers() gameres.PowerManager {
	return this.powers
}

func (this *GameRes) NewPowers(modules common.Modules, ss gameres.Stats) gameres.PowerManager {
	this.powers = powermanager.New(modules, ss)
	return this.powers
}

// 工厂
func (this *GameRes) Menuf() gameres.Factory {
	return menu.NewFactory()
}

func (this *GameRes) Resf() gameres.Factory {
	return resources.NewFactory()
}
