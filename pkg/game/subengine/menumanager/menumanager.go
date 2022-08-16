package menumanager

import (
	"monster/pkg/common"
	"monster/pkg/common/define/game/menu/statbar"
	"monster/pkg/common/define/game/stats"
	"monster/pkg/common/gameres"
)

type MenuManager struct {
	menus map[string]gameres.Menu
}

func New(modules common.Modules, pc gameres.Avatar, powers gameres.PowerManager, menuf gameres.Factory) *MenuManager {
	mm := &MenuManager{}
	mm.init(modules, pc, powers, menuf)

	return mm
}

func (this *MenuManager) init(modules common.Modules, pc gameres.Avatar, powers gameres.PowerManager, menuf gameres.Factory) gameres.MenuManager {
	this.menus = map[string]gameres.Menu{}

	this.menus["inv"] = menuf.New("inventory").(gameres.MenuInventory).Init(modules)
	this.menus["hp"] = menuf.New("statbar").(gameres.MenuStatBar).Init(modules, statbar.TYPE_HP)
	this.menus["mp"] = menuf.New("statbar").(gameres.MenuStatBar).Init(modules, statbar.TYPE_MP)
	this.menus["xp"] = menuf.New("statbar").(gameres.MenuStatBar).Init(modules, statbar.TYPE_XP)
	this.menus["exit"] = menuf.New("exit").(gameres.MenuExit).Init(modules, pc)
	this.menus["act"] = menuf.New("actionbar").(gameres.MenuActionBar).Init(modules, powers)

	return this
}

func (this *MenuManager) Close() {
	for _, ptr := range this.menus {
		ptr.Close()
	}
}

func (this *MenuManager) AlignAll(modules common.Modules) error {
	for _, ptr := range this.menus {
		err := ptr.Align(modules)
		if err != nil {
			return err
		}
	}

	return nil

}

func (this *MenuManager) Render(modules common.Modules) error {
	for _, ptr := range this.menus {
		err := ptr.Render(modules)
		if err != nil {
			panic(err)
			return err
		}
	}

	return nil

}

func (this *MenuManager) Get(name string) gameres.Menu {
	return this.menus[name]
}

func (this *MenuManager) Logic(modules common.Modules, pc gameres.Avatar, powers gameres.PowerManager) {
	eset := modules.Eset()

	// TODO
	// sound
	this.menus["hp"].(gameres.MenuStatBar).Update(0, (uint64)(pc.GetStats().GetHP()), (uint64)(pc.GetStats().Get(stats.HP_MAX)))
	this.menus["mp"].(gameres.MenuStatBar).Update(0, (uint64)(pc.GetStats().GetMP()), (uint64)(pc.GetStats().Get(stats.MP_MAX)))

	if pc.GetStats().GetLevel() == eset.XPGetMaxLevel() {
		// TODO
	}

	this.menus["act"].Logic(modules, pc, powers)
}

func (this *MenuManager) MenuAct() gameres.MenuActionBar {
	return this.menus["act"].(gameres.MenuActionBar)
}
