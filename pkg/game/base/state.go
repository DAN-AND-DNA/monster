package base

import (
	"monster/pkg/common"
	"monster/pkg/common/define/fontengine"
	"monster/pkg/common/gameres"
	"monster/pkg/common/point"
	"monster/pkg/common/tooltipdata"
	"monster/pkg/widget"
)

type State struct {
	HasMusic               bool
	hasBackground          bool
	ReloadMusic            bool
	reloadBackgrounds      bool
	forceRefreshBackground bool
	SaveSettingsOnExit     bool
	loadCounter            int
	requestedGameState     gameres.GameState // 准备切换到该场景
	exitRequested          bool
	loadingTip             common.WidgetTooltip
	loadingTipBuf          tooltipdata.TooltipData
}

func ConstructState(modules common.Modules) State {
	font := modules.Font()
	msg := modules.Msg()

	s := State{
		hasBackground:      true,
		SaveSettingsOnExit: true,
		loadingTip:         widget.NewTooltip(modules),
		loadingTipBuf:      tooltipdata.Construct(),
	}

	err := s.loadingTipBuf.AddColorText(msg.Get("Loading..."), font.GetColor(fontengine.COLOR_WIDGET_NORMAL))
	if err != nil {
		panic(err)
	}

	return s
}

// 清理
func (this *State) clear() {
	this.CloseLoadingTip()
}

func (this *State) Close(modules common.Modules, gameRes gameres.GameRes, impl gameres.GameState) {
	impl.Clear(modules, gameRes)

	// 自己
	this.clear()
}

func (this *State) CloseLoadingTip() {
	if this.loadingTip != nil {
		this.loadingTip.Close()
		this.loadingTip = nil
	}
}

func (this *State) GetRequestedGameState() gameres.GameState {
	return this.requestedGameState
}

func (this *State) SetRequestedGameState(modules common.Modules, gameRes gameres.GameRes, newState gameres.GameState) {
	if this.requestedGameState != nil {
		this.requestedGameState.Close(modules, gameRes)
		this.requestedGameState = nil
	}

	this.requestedGameState = newState

	this.requestedGameState.SetLoadingFrame()
	this.requestedGameState.RefreshWidgets(modules, gameRes) // 刷新组件的位置和大小
}

func (this *State) SetLoadingFrame() {
	this.loadCounter = 2
}

// 渲染加载文字
func (this *State) ShowLoading(modules common.Modules) error {
	render := modules.Render()
	settings := modules.Settings()
	inpt := modules.Inpt()

	if this.loadingTip == nil {
		return nil
	}

	err := this.loadingTip.Render(
		modules,
		this.loadingTipBuf,
		point.Construct(settings.GetViewW(), settings.GetViewH()),
		tooltipdata.STYLE_FLOAT)

	if err != nil {
		return err
	}

	err = render.CommitFrame(inpt)
	if err != nil {
		return err
	}

	return nil
}

func (this *State) SetForceRefreshBackground(val bool) {
	this.forceRefreshBackground = val
}

func (this *State) GetForceRefreshBackground() bool {
	return this.forceRefreshBackground
}

func (this *State) GetHasBackground() bool {
	return this.hasBackground
}

func (this *State) SetReloadBackgrounds(val bool) {
	this.reloadBackgrounds = val
}

func (this *State) GetReloadBackgrounds() bool {
	return this.reloadBackgrounds
}

func (this *State) IncrLoadCounter() {
	this.loadCounter++
}

func (this *State) GetLoadCounter() int {
	return this.loadCounter
}

func (this *State) DecrLoadCounter() {
	this.loadCounter--
}

func (this *State) SetExitRequested(val bool) {
	this.exitRequested = val
}

func (this *State) GetExitRequested() bool {
	return this.exitRequested
}
