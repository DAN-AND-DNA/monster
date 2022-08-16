package menu

import (
	"monster/pkg/common"
	"monster/pkg/common/gameres"
	"monster/pkg/game/base"
)

type Exit struct {
	base.Menu

	menuConfig  *Config
	exitClicked bool
	reloadMusic bool
}

func NewExit(modules common.Modules, pc gameres.Avatar) *Exit {
	exit := &Exit{}
	exit.Init(modules, pc)

	return exit
}

func (this *Exit) Init(modules common.Modules, pc gameres.Avatar) gameres.MenuExit {
	// base
	this.Menu = base.ConstructMenu(modules)

	// self
	this.menuConfig = NewConfig(modules, false)
	this.menuConfig.hero = pc
	this.Align(modules)

	return this
}

func (this *Exit) Close() {
	this.Menu.Close(this)

}

func (this *Exit) Clear() {

	if this.menuConfig != nil {
		this.menuConfig.Close()
		this.menuConfig = nil
	}
}

func (this *Exit) Align(modules common.Modules) error {
	this.Menu.Align(modules)
	err := this.menuConfig.RefreshWidgets(modules)
	if err != nil {
		return err
	}

	return nil
}

func (this *Exit) Logic(modules common.Modules, pc gameres.Avatar, powers gameres.PowerManager) error {
	if this.GetVisible() {
		// 菜单
		this.menuConfig.Logic(modules)
	}

	if this.menuConfig.reloadMusic {
		this.reloadMusic = true
		this.menuConfig.reloadMusic = false
	}

	if this.menuConfig.clickedPauseContinue {
		this.SetVisible(false)
		this.menuConfig.clickedPauseContinue = false
	} else if this.menuConfig.clickedPauseExit {
		this.exitClicked = true
		this.menuConfig.clickedPauseExit = false
	} else if this.menuConfig.clickedPauseSave {
		this.SetVisible(false)
		this.menuConfig.clickedPauseSave = false

		//TODO
		// save
	}

	return nil
}

func (this *Exit) Render(modules common.Modules) error {
	if this.GetVisible() {
		err := this.Menu.Render(modules)
		if err != nil {
			return err
		}

		err = this.menuConfig.Render(modules)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *Exit) DisableSave(modules common.Modules) {
	this.menuConfig.SetPauseExitText(modules, false)
	this.menuConfig.SetPauseSaveEnabled(modules, false)
}

func (this *Exit) HandleCancel(modules common.Modules) error {
	if !this.GetVisible() {
		err := this.menuConfig.ResetSelectedTab(modules)
		if err != nil {
			return err
		}
		this.SetVisible(true)
	} else {
		if !this.menuConfig.inputConfirm.GetVisible() {
			this.SetVisible(false)
		}
	}

	return nil
}
