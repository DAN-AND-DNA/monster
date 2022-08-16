package base

import (
	"monster/pkg/common"
	"monster/pkg/common/define"
	"monster/pkg/common/define/widget"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/utils"
)

type Widget struct {
	inFocus          bool
	focusable        bool
	enableTablistNav bool
	tablistNavRight  bool
	scrollType       int         // 滚动类型
	pos              rect.Rect   // 当前位置
	localFrame       rect.Rect   // 父级的位置
	localOffset      point.Point // 自己的偏移或父级的偏移位置
	posBase          point.Point // 初始位置
	alignment        int
}

func NewWidget() *Widget {
	w := ConstructWidget()
	return &w
}

func ConstructWidget() Widget {
	w := Widget{
		enableTablistNav: true,
		scrollType:       widget.SCROLL_TWO_DIRECTIONS,
		alignment:        define.ALIGN_TOPLEFT,
	}

	return w
}

func (this *Widget) clear() {
}

func (this *Widget) Close(impl common.Widget) {
	impl.Clear()

	this.clear()
}

func (this *Widget) GetTablistNavRight() bool {
	return this.tablistNavRight
}

func (this *Widget) SetTablistNavRight(val bool) {
	this.tablistNavRight = val
}

func (this *Widget) GetPos() rect.Rect {
	return this.pos
}

func (this *Widget) Defocus() {
	this.inFocus = false
}

func (this *Widget) Focus() {
	this.inFocus = true
}

func (this *Widget) SetEnableTablistNav(val bool) {
	this.enableTablistNav = val
}

func (this *Widget) GetEnableTablistNav() bool {
	return this.enableTablistNav
}

func (this *Widget) GetPosBase() point.Point {
	return this.posBase
}

// 设置自己在父组件的位置
func (this *Widget) SetPosBase(x, y, a int) {
	this.posBase.X = x
	this.posBase.Y = y
	this.alignment = a
}

// 加上父组件的位置，计算在地图中的位置、偏移和对齐方式，设置渲染位置
func (this *Widget) SetPos1(modules common.Modules, offsetX, offsetY int) error {
	settings := modules.Settings()
	eset := modules.Eset()

	this.pos.X = this.posBase.X + offsetX
	this.pos.Y = this.posBase.Y + offsetY

	// 对齐到某个边，另一个边的坐标
	this.pos = utils.AlignToScreenEdge(settings, eset, this.alignment, this.pos)

	return nil
}

func (this *Widget) SetPos(pos rect.Rect) {
	this.pos = pos
}

func (this *Widget) SetPosX(x int) {
	this.pos.X = x
}

func (this *Widget) SetPosY(y int) {
	this.pos.Y = y
}

func (this *Widget) SetPosW(w int) {
	this.pos.W = w
}

func (this *Widget) SetPosH(h int) {
	this.pos.H = h
}

func (this *Widget) SetScrollType(st int) {
	this.scrollType = st
}

func (this *Widget) GetScrollType() int {
	return this.scrollType
}

func (this *Widget) GetLocalOffset() point.Point {
	return this.localOffset
}

func (this *Widget) GetLocalFrame() rect.Rect {
	return this.localFrame
}
func (this *Widget) SetLocalOffset(p point.Point) {
	this.localOffset = p
}

func (this *Widget) SetLocalFrame(r rect.Rect) {
	this.localFrame = r
}

func (this *Widget) SetFocusable(focusable bool) {
	this.focusable = focusable
}

func (this *Widget) GetInFocus() bool {
	return this.inFocus
}

func (this *Widget) SetAlignment(val int) {
	this.alignment = val
}

func (this *Widget) GetAlignment() int {
	return this.alignment
}
