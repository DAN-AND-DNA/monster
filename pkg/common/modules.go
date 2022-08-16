package common

import (
	"monster/pkg/common/color"
	"monster/pkg/common/define/widget/slot"
	"monster/pkg/common/fpoint"
	"monster/pkg/common/labelinfo"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/common/tooltipdata"
	"monster/pkg/config/version"
)

type Modules interface {
	Settings() Settings
	NewSettings() Settings
	Platform() Platform
	NewPlatform() Platform
	Eset() EngineSettings
	NewEset() EngineSettings
	Mods() ModManager
	NewMods(Platform, Settings, []string) ModManager
	Msg() MessageEngine
	NewMsg() MessageEngine
	Font() FontEngine
	NewFont(Settings, ModManager) FontEngine
	Render() RenderDevice
	NewRender(Settings, EngineSettings) RenderDevice
	Inpt() InputState
	NewInpt(Platform, Settings, EngineSettings, ModManager, MessageEngine) InputState
	Tooltipm() Tooltipm
	NewTooltipm(Settings, ModManager, RenderDevice) Tooltipm
	Anim() AnimationManager
	NewAnim() AnimationManager
	Icons() IconManager
	NewIcons(Settings, EngineSettings, RenderDevice, ModManager) IconManager

	// 工厂方法
	Widgetf() Factory
	Resf() Factory
}

type Settings interface {
	LoadSettings(ModManager) error
	LogSettings()
	Get(string) interface{}
	Set(string, interface{})
	LOGIC_FPS() float32

	SetPathConf(string)
	SetPathUser(string)
	SetPathData(string)
	SetSafeVideo(bool)
	SetViewH(int)
	SetViewW(int)
	SetViewHHalf(int)
	SetViewWHalf(int)
	SetViewScaling(float32)

	GetPathConf() string
	GetPathData() string
	GetPathUser() string
	GetCustomPathData() string
	GetViewH() int
	GetViewW() int
	GetViewHHalf() int
	GetViewWHalf() int
	GetViewScaling() float32
	GetMouseScaled() bool
	UpdateScreenVars(EngineSettings)
	GetLoadSlot() string
	GetShowHud() bool
	SetSoftReset(bool)
	GetSoftReset() bool
	GetEncounterDist() float32
}

type Platform interface {
	GetHasExitButton() bool
	GetHasLockfile() bool
	SetPaths(Settings) error
	GetIsMobileDevice() bool
	FullscreenBypass() bool
	SetFullscreen(bool)
	GetConfigMenuType() int
	SetExitEventFilter()
	GetConfigVideo() []bool
	GetConfigAudio() []bool
	GetConfigInterface() []bool
	GetConfigInput() []bool
	GetConfigMisc() []bool
}

type Mod interface {
	GetGame() string
	GetName() string
	GetVersion() version.Version
	GetEngineMinVersion() version.Version
	GetEngineMaxVersion() version.Version
	GetLocalDescription(string) string
	GetDepends() []string
	GetDependsMax() []version.Version
	GetDependsMin() []version.Version
}

type ModManager interface {
	List(string) ([]string, error)
	Locate(Settings, string) (string, error)
	ApplyDepends() error
	ClearModList()
	GetModList() []Mod
	GetModDirs() []string
	LoadMod(string) (Mod, error)
	AddToModList(Mod)
	SaveMods(Modules) error
}

type DamageType interface {
	GetId() string
	GetName() string
	GetNameMin() string
	GetNameMax() string
	GetDescription() string
	GetMin() string
	GetMax() string
}

type Element interface {
	GetName() string
	GetId() string
}

type PrimaryStats interface {
	GetIndexById(string) (int, bool)
}

type PrimaryStat interface {
	GetName() string
	GetId() string
}

/*
type Xp interface {
	GetLevelXP(int) uint64
	GetMaxLevel() int
	GetLevelFromXP(uint64) int
}
*/

type HeroClass interface {
	GetName() string
	GetDescription() string
	GetHeroOptions() []int
}

type EngineSettings interface {
	Load(Settings, ModManager, MessageEngine, FontEngine) error
	Get(string, string) interface{}
	XPGetLevelXP(int) uint64
	XPGetMaxLevel() int
	XPGetLevelFromXP(uint64) int
	PrimaryStatsGetIndexById(string) (int, bool)
}

type Animation interface {
	Close()
	DeepCopy() Animation
	GetName() string
	GetCurFrame() uint16
	GetCurFrameIndex() uint16
	GetCurFrameIndexF() float32
	GetTimesPlayed() uint
	GetAdditionalData() int16
	GetElapsedFrames() uint16
	SyncTo(Animation) bool
	AdvanceFrame()
	GetCurrentFrame(modules Modules, kind int) Renderable
	IsLastFrame() bool
	IsCompleted() bool
	GetDuration() int
}

type AnimationSet interface {
	Init(Settings, ModManager, RenderDevice, string) AnimationSet
	GetAnimation(string) Animation
	GetName() string
	GetAnimationFrameCount(name string) uint16
	SetParent(AnimationSet)
	Close()
}

type AnimationManager interface {
	Close()
	CleanUp()
	IncreaseCount(string)
	DecreaseCount(string)
	GetAnimationSet(Settings, ModManager, RenderDevice, Factory, string) (AnimationSet, error)
}

type MessageEngine interface {
	Get(string) string
}
type CombatText interface {
}

type Renderable interface {
	SetImage(Image)
	SetSrc(rect.Rect)
	SetType(uint8)
	SetMapPos(fpoint.FPoint)
	SetOffset(point.Point)
	SetPrio(uint64)
	SetBlendMode(uint8)
	SetColorMod(color.Color)
	SetAlphaMod(uint8)
	GetImage() Image
	GetSrc() rect.Rect
	GetType() uint8
	GetMapPos() fpoint.FPoint
	GetOffset() point.Point
	GetPrio() uint64
	GetBlendMode() uint8
	GetColorMod() color.Color
	GetAlphaMod() uint8
}

type Sprite interface {
	Close()
	KeepAlive()
	SetClip(x, y, w, h int) error
	SetClipFromRect(rect.Rect) error
	GetClip() rect.Rect
	GetGraphicsWidth() (int, error)
	GetGraphicsHeight() (int, error)
	SetOffset(point.Point)
	GetOffset() point.Point
	SetDestFromRect(rect.Rect)
	SetDestFromPoint(point.Point)
	SetDest(int, int)
	GetDest() point.Point
	GetSrc() rect.Rect
	GetGraphics() Image
	SetLocalFrame(rect.Rect)
	GetLocalFrame() rect.Rect
	ColorMod() color.Color
	SetAlphaMod(uint8)
	AlphaMod() uint8
}

type Image interface {
	GetFilename() string
	GetWidth() (int, error)
	GetHeight() (int, error)
	CreateSprite() (Sprite, error)
	Ref()
	UnRef()
	Close()
	Clear()
	GetRefCount() uint64
	FillWithColor(color.Color) error
	DrawPixel(x int, y int, color color.Color) error
	DrawLine(x0, y0, x1, y1 int, color color.Color) error
	GetDevice() RenderDevice
	BeginPixelBatch() error
	EndPixelBatch() error
	Resize(w, h int) (Image, error)
	Surface() interface{}
	SetSurface(interface{})
}

type CursorManager interface {
	Close()
	SetShowCursor(bool)
	Logic(Settings, InputState) error
	Render(Settings, RenderDevice, InputState) error
}

type RenderDevice interface {
	CreateContext(Settings, EngineSettings, MessageEngine, ModManager) error
	CreateContextInternal(Settings, EngineSettings, MessageEngine, ModManager) error
	CreateContextError()
	DestroyContext()
	SetGamma(float32) error
	UpdateTitleBar(Settings, EngineSettings, MessageEngine, ModManager) error
	BlankScreen() error
	WindowResize(Settings, EngineSettings) error
	CommitFrame(InputState) error
	ResetGamma() error
	GetWindowSize() (int, int)
	Clear()
	Close()
	ReloadGraphics() bool
	CreateImage(width, height int) (Image, error)
	FreeImage(string)
	LoadImage(Settings, ModManager, string) (Image, error)
	DrawRectangle(p0, p1 point.Point, color color.Color) error
	Render(Sprite) error
	Render1(r Renderable, dest rect.Rect) error
	RenderToImage(srcImage Image, src rect.Rect, destImage Image, dest rect.Rect) (rect.Rect, error)
	RenderTextToImage(fontStyle FontStyle, text string, color color.Color, blended bool) (Image, error)
	FillRect() error
	Curs() CursorManager
	SetBackgroundColor(color.Color)
	GetRefreshRate() int
	CreateRenderDeviceList(MessageEngine) ([]string, []string)
}

type SDLHardwareRenderDevice interface {
	RenderDevice
}

type FontStyle interface {
	Ttfont() interface{}
}

type FontEngine interface {
	Close()
	SetFont(string)
	GetColor(int) color.Color
	CalcWidth(string) int
	GetFontHeight() int
	CalcSize(textWithNewlines string, width int) point.Point
	CursorY() int
	GetLineHeight() int
	TrimTextToWidth(text string, width int, useEllipsis bool, leftPos int) string
	Render(RenderDevice, string, int, int, int, Image, int, color.Color) error
	RenderInternal(RenderDevice, string, int, int, int, Image, color.Color) error
	RenderShadowed(renderDevice RenderDevice, text string, x, y, justify int, target Image, width int, color color.Color) error
}

type Factory interface {
	New(string) interface{}
}

type Widget interface {
	Close()
	Clear()
	GetLocalFrame() rect.Rect
	SetLocalFrame(rect.Rect)
	GetLocalOffset() point.Point
	SetLocalOffset(point.Point)
	GetTablistNavRight() bool
	SetTablistNavRight(bool)
	SetFocusable(bool)
	SetScrollType(int)
	GetScrollType() int
	SetPosX(int)
	SetPosY(int)
	SetPosW(int)
	SetPosH(int)
	Render(Modules) error
	SetPos1(Modules, int, int) error
	SetPos(rect.Rect)
	GetPos() rect.Rect
	GetPosBase() point.Point
	SetPosBase(x, y, a int)
	GetEnableTablistNav() bool
	GetInFocus() bool
	Focus()
	Deactivate()
	Activate()
	Defocus()
	GetNext(Modules) bool
	GetPrev(Modules) bool
	GetAlignment() int
	SetAlignment(int)
}

type WidgetTablist interface {
	Close()
	Init() WidgetTablist
	Unlock()
	GetNext(m Modules, inner bool, dir int) (Widget, bool)
	GetPrev(m Modules, inner bool, dir int) (Widget, bool)
	SetIngoreNoMouse(bool)
	Add(Widget)
	Logic(Modules) error
	GetCurrent() int
	Defocus()
}

type WidgetTooltip interface {
	Close()
	Init(Modules) WidgetTooltip
	Render(Modules, tooltipdata.TooltipData, point.Point, int) error
	Render1(Settings, EngineSettings, FontEngine, RenderDevice, tooltipdata.TooltipData, point.Point, int) error
	SetParent(WidgetTooltip)
	GetBounds() rect.Rect
}

type WidgetLabel interface {
	Widget
	Init(Modules) WidgetLabel
	Init1(FontEngine) WidgetLabel
	SetJustify(int)
	SetValign(int)
	SetText(string)
	SetColor(color.Color)
	SetHidden(bool)
	SetMaxWidth(int)
	GetBounds(Modules) rect.Rect
	SetFromLabelInfo(labelinfo.LabelInfo)
}

type WidgetScrollBar interface {
	Widget
	Init(modules Modules, filename string) WidgetScrollBar
	GetPosDown() rect.Rect
	Refresh(modules Modules, x, y, h, val, max int) error
}

type WidgetScrollBox interface {
	Widget
	Init(Modules, int, int) WidgetScrollBox
	SetBg(color.Color)
	Resize(Modules, int, int) error
	AddChildWidget(Widget)
	Refresh(Modules) error
	Logic(Modules) error
	InputAssist(point.Point) (point.Point, bool)
	ScrollToTop(Modules) error
}

type WidgetButton interface {
	Widget
	Init(Modules, string) WidgetButton
	SetLabel(Modules, string)
	GetEnabled() bool
	SetEnabled(bool)
	SetTooltip(string)
	Refresh(Modules)
	CheckClick(Modules) bool
	CheckClickAt(Modules, int, int) bool
}

type WidgetTabControl interface {
	Widget
	Init(Modules) WidgetTabControl
	SetTabTitle(Modules, int, string)
	GetTabHeight() (int, error)
	SetEnabled(uint, bool)
	SetMainArea(Modules, int, int) error
	GetActiveTab() uint
	SetActiveTab(uint)
	Logic(Modules) error
}

type WidgetCheckBox interface {
	Widget
	Init(Modules, string) WidgetCheckBox
	SetTooltip(string)
	SetEnabled(bool)
	SetChecked(bool)
	CheckClick(Modules)
}

type WidgetSlider interface {
	Widget
	Init(Modules, string) WidgetSlider
	Set(int, int, int)
	SetEnabled(bool)
}

type WidgetHorizontalList interface {
	Widget
	Init(Modules) WidgetHorizontalList
	SetHasAction(bool)
	Append(string, string)
	Clear1()
	Select(Modules, int)
	GetValue() string
	CheckClickAt(Modules, int, int) bool
}

type WidgetListBox interface {
	Widget
	Init(Modules, int, string) WidgetListBox
	SetScrollbarOffset(int)
	SetMultiSelect(bool)
	Append(Modules, string, string)
	SetHeight(Modules, int)
	Sort()
	Refresh(Modules)
	CheckClick(Modules) bool
	ShiftUp(Modules)
	ShiftDown(Modules)
	IsSelected(int) bool
	Remove(Modules, int)
	GetValue(int) (string, bool)
	GetTooltip(int) (string, bool)
	GetSize() int
	SetCanDeselect(bool)
	GetSelected() (int, bool)
	Select(int)
}

type WidgetSlot interface {
	Widget

	Init(modules Modules, iconId, activate int) WidgetSlot
	SetHotkey(modules Modules, key int) error
	SetContinuous(bool)
	RenderSelection(modules Modules) error
	SetAmount(modules Modules, amount, maxAmount int) error
	SetIcon(iconId int, overlayId slot.CLICK_TYPE)
}

type InputState interface {
	Clear()
	Close()
	ResetScroll()
	SetDone(bool)
	GetDone() bool
	GetWindowMinimized() bool
	SetWindowMinimized(bool)
	GetWindowRestored() bool
	SetWindowRestored(bool)
	GetWindowResized() bool
	SetWindowResized(bool)
	GetKeyFromName(string) int
	SetFixedKeyBinding()
	HideCursor() error
	ShowCursor() error
	GetMouse() point.Point
	UsingMouse(Settings) bool
	GetScrollUp() bool
	SetScrollUp(bool)
	GetScrollDown() bool
	SetScrollDown(bool)
	GetLockScroll() bool
	SetLockScroll(bool)
	GetLock(int) bool
	SetLock(int, bool)
	SetLockAll(bool)
	GetPressing(int) bool
	Handle(Modules) error
	GetBindingName(int) string
	GetBindingString(msg MessageEngine, key int, getShortString bool) string
	GetRefreshHotkeys() bool
}

type Tooltipm interface {
	Close()
	Clear()
	IsEmpty() bool
	Push(tooltipdata.TooltipData, point.Point, int, int)
	Render(Settings, EngineSettings, FontEngine, RenderDevice) error
}

type IconManager interface {
	Close()
	SetIcon(eset EngineSettings, iconId int, destPos point.Point)
	RenderToImage(RenderDevice, Image) error
	Render(RenderDevice) error
	GetTextOffset() point.Point
}

//  衔接

type GameSwitcher interface {
	Render(Modules) error
	IsLoadingFrame() bool
	ShowFPS(Modules, float32) error
	Logic(Modules) error
	GetDone() bool
}
