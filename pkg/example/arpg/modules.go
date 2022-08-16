package main

import (
	"monster/pkg/common"
	"monster/pkg/config/enginesettings"
	"monster/pkg/config/platform"
	"monster/pkg/config/settings"
	"monster/pkg/resources"
	"monster/pkg/subengine/animationmanager"
	"monster/pkg/subengine/fontengine/sdlfont"
	"monster/pkg/subengine/iconmanager"
	"monster/pkg/subengine/inputstate/sdlinput"
	"monster/pkg/subengine/messageengine"
	"monster/pkg/subengine/modmanager"
	"monster/pkg/subengine/render/sdlhardware"
	"monster/pkg/subengine/tooltipmanager"
	"monster/pkg/widget"
)

type Modules struct {
	settings common.Settings
	platform common.Platform
	eset     common.EngineSettings
	mods     common.ModManager
	msg      common.MessageEngine
	font     common.FontEngine
	render   common.RenderDevice
	inpt     common.InputState
	tooltipm common.Tooltipm
	anim     common.AnimationManager
	icons    common.IconManager
}

func NewModules() common.Modules {
	return &Modules{}
}

func (this *Modules) Settings() common.Settings {
	return this.settings
}

func (this *Modules) NewSettings() common.Settings {
	this.settings = settings.New()
	return this.settings
}

func (this *Modules) Platform() common.Platform {
	return this.platform
}

func (this *Modules) NewPlatform() common.Platform {
	this.platform = platform.New()
	return this.platform
}

func (this *Modules) Eset() common.EngineSettings {
	return this.eset
}

func (this *Modules) NewEset() common.EngineSettings {
	this.eset = enginesettings.New()
	return this.eset
}

func (this *Modules) Mods() common.ModManager {
	return this.mods
}

func (this *Modules) NewMods(platform common.Platform, settings common.Settings, modList []string) common.ModManager {
	this.mods = modmanager.New(platform, settings, modList)
	return this.mods
}

func (this *Modules) Msg() common.MessageEngine {
	return this.msg
}

func (this *Modules) NewMsg() common.MessageEngine {
	this.msg = messageengine.New()
	return this.msg
}

func (this *Modules) Font() common.FontEngine {
	return this.font
}

func (this *Modules) NewFont(settings common.Settings, mods common.ModManager) common.FontEngine {
	if this.font != nil {
		this.font.Close()
	}

	this.font = sdlfont.NewFontEngine(settings, mods)

	return this.font
}

func (this *Modules) Render() common.RenderDevice {
	return this.render
}

func (this *Modules) NewRender(settings common.Settings, eset common.EngineSettings) common.RenderDevice {
	if this.render != nil {
		this.render.Close()
	}

	this.render = sdlhardware.NewRenderDevice(settings, eset)
	return this.render
}

func (this *Modules) Inpt() common.InputState {
	return this.inpt
}

func (this *Modules) NewInpt(platform common.Platform, settings common.Settings, eset common.EngineSettings, mods common.ModManager, msg common.MessageEngine) common.InputState {

	if this.inpt != nil {
		this.inpt.Close()
	}

	this.inpt = sdlinput.New(platform, settings, eset, mods, msg)

	return this.inpt
}

func (this *Modules) Tooltipm() common.Tooltipm {
	return this.tooltipm
}

func (this *Modules) NewTooltipm(settings common.Settings, mods common.ModManager, render common.RenderDevice) common.Tooltipm {
	if this.tooltipm != nil {
		this.tooltipm.Close()
	}

	this.tooltipm = tooltipmanager.New(settings, mods, render)

	return this.tooltipm
}

func (this *Modules) Anim() common.AnimationManager {
	return this.anim
}

func (this *Modules) NewAnim() common.AnimationManager {
	if this.anim != nil {
		this.anim.Close()
	}
	this.anim = animationmanager.New()

	return this.anim
}

func (this *Modules) Icons() common.IconManager {
	return this.icons
}

func (this *Modules) NewIcons(settings common.Settings, eset common.EngineSettings, render common.RenderDevice, mods common.ModManager) common.IconManager {
	if this.icons != nil {
		this.icons.Close()
	}

	this.icons = iconmanager.New(settings, eset, render, mods)

	return this.icons
}

// 工厂
func (this *Modules) Widgetf() common.Factory {
	return widget.NewFactory()
}

func (this *Modules) Resf() common.Factory {
	return resources.NewFactory()
}
