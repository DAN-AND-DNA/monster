package menu

import (
	"monster/pkg/common"
	"monster/pkg/common/gameres"
	"monster/pkg/game/base"
)

type Inventory struct {
	base.Menu

	currency         int
	changedEquipment bool
}

func NewInventory(modules common.Modules) *Inventory {
	i := &Inventory{}
	i.Init(modules)

	return i
}

func (this *Inventory) Init(modules common.Modules) gameres.MenuInventory {
	// base
	this.Menu = base.ConstructMenu(modules)

	// self
	this.changedEquipment = true
	return this
}

func (this *Inventory) Clear() {
}

func (this *Inventory) Close() {
	this.Menu.Close(this)
}

func (this *Inventory) Logic(common.Modules, gameres.Avatar, gameres.PowerManager) error {
	return nil
}

func (this *Inventory) SetChangedEquipment(val bool) {
	this.changedEquipment = val
}

func (this *Inventory) GetChangedEquipment() bool {
	return this.changedEquipment
}

func (this *Inventory) SetCurrency(val int) {
	this.currency = val
}

func (this *Inventory) GetCurrency() int {
	return this.currency
}
