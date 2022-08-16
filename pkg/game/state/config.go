package state

import (
	"fmt"
	"monster/pkg/common"
	"monster/pkg/common/gameres"
	"monster/pkg/game/base"
)

type Config struct {
	base.State

	menuConfig         gameres.MenuConfig
	saveSettingsOnExit bool
}

func NewConfig(modules common.Modules, gameRes gameres.GameRes) *Config {
	c := &Config{}
	c.Init(modules, gameRes)

	return c
}

func (this *Config) Init(modules common.Modules, gameRes gameres.GameRes) gameres.GameStateConfig {
	menuf := gameRes.Menuf()

	// base
	this.State = base.ConstructState(modules)

	// self
	this.menuConfig = menuf.New("config").(gameres.MenuConfig).Init(modules, true)

	return this
}

func (this *Config) Clear(common.Modules, gameres.GameRes) {
	if this.menuConfig != nil {
		this.menuConfig.Close()
		this.menuConfig = nil
	}
}

func (this *Config) Close(modules common.Modules, gameRes gameres.GameRes) {
	this.State.Close(modules, gameRes, this)
}

func (this *Config) logicAccept(modules common.Modules, gameRes gameres.GameRes) error {
	platform := modules.Platform()
	settings := modules.Settings()
	eset := modules.Eset()
	mods := modules.Mods()
	msg := modules.Msg()
	font := modules.Font()
	render := modules.Render()
	tooltipm := modules.Tooltipm()
	inpt := modules.Inpt()

	stats := gameRes.Stats()

	newRenderDevice := this.menuConfig.GetRenderDevice()

	//TODO

	changed, err := this.menuConfig.RefreshMods(modules)
	if err != nil {
		return err
	}

	// mod有改动
	if changed {

		fmt.Println("mods change")
		// 需要加载背景图片
		this.SetReloadBackgrounds(true)

		mods = modules.NewMods(platform, settings, nil)
		settings.Set("prev_save_slot", -1)
	}

	msg = modules.NewMsg()
	err = eset.Load(settings, mods, msg, font)
	if err != nil {
		return err
	}

	stats.Init(modules)
	// 刷新字体 bad
	font = modules.NewFont(settings, mods)

	// TODO
	// inpt

	this.menuConfig.Close()
	this.menuConfig = nil

	err = this.ShowLoading(modules)
	if err != nil {
		return err
	}

	// 手动清理
	this.State.CloseLoadingTip()
	tooltipm.Close()

	//TODO
	// frame limit

	if newRenderDevice != settings.Get("renderer").(string) {
		fmt.Println("break")
		settings.Set("renderer", newRenderDevice)

		// 跳出并重启循环，清理之前创建的渲染资源
		inpt.SetDone(true)
		settings.SetSoftReset(true)
	}

	err = render.CreateContext(settings, eset, msg, mods)
	if err != nil {
		return err
	}

	// 用最新的render去创建 ok
	tooltipm = modules.NewTooltipm(settings, mods, render)

	// TODO save settings

	// 请求主逻辑去更换场景，并在switcher里清理当前场景
	this.SetRequestedGameState(modules, gameRes, NewTitle(modules, gameRes))
	return nil
}

// 先逻辑后渲染
func (this *Config) Logic(modules common.Modules, gameRes gameres.GameRes) error {
	err := this.menuConfig.Logic(modules)
	if err != nil {
		return err
	}

	if this.menuConfig.GetForceRefreshBackground() {
		this.SetForceRefreshBackground(true)
		this.menuConfig.SetForceRefreshBackground(false)
	}

	if this.menuConfig.GetClickedAccept() {
		this.menuConfig.SetClickedAccept(false)
		this.logicAccept(modules, gameRes)
	} else if this.menuConfig.GetClickedCancel() {
		this.menuConfig.SetClickedCancel(false)
	}

	return nil
}

func (this *Config) RefreshWidgets(modules common.Modules, gameRes gameres.GameRes) error {
	// pass
	return nil
}

func (this *Config) Render(modules common.Modules, gameRes gameres.GameRes) error {
	if this.GetRequestedGameState() != nil {
		// 正在切换场景
		return nil
	}

	err := this.menuConfig.Render(modules)
	if err != nil {
		return err
	}
	return nil
}
