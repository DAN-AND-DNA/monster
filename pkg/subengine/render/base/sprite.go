package base

import (
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
)

// 0---->x
// |
// |
// v
// y

// 精灵代表一个要显示的元素
type Sprite struct {
	localFrame rect.Rect
	colorMod   color.Color
	alphaMod   uint8
	src        rect.Rect    // 裁剪图片的矩形
	offset     point.Point  // 与sprite左上角的偏移量，该点表示sprite在地图的位置
	dest       point.Point  // 绘制的位置
	image      common.Image // 外部指针
	keepalive  bool
}

func NewSprite(image common.Image) *Sprite {
	s := &Sprite{
		localFrame: rect.Construct(),
		colorMod:   color.Construct(255, 255, 255),
		alphaMod:   255,
		src:        rect.Construct(),
		offset:     point.Construct(),
		dest:       point.Construct(),
		image:      image,
	}

	_ = (common.Sprite)(s)
	image.Ref() // 图片引用+1
	return s
}

func (this *Sprite) Close() {
	if !this.keepalive {
		this.image.UnRef() // 图片引用-1
		this.image = nil
	}

	this.keepalive = false
}

func (this *Sprite) KeepAlive() {
	this.keepalive = true
}

func (this *Sprite) SetClip(x, y, w, h int) error {
	return this.SetClipFromRect(rect.Construct(x, y, w, h))
}

func (this *Sprite) SetClipFromRect(clip rect.Rect) error {
	this.src = clip

	// 保证不超过图片的范围
	targetW, err := this.GetGraphicsWidth()
	if err != nil {
		return err
	}

	targetH, err := this.GetGraphicsHeight()
	if err != nil {
		return err
	}

	if this.src.X+this.src.W > targetW {
		this.src.W = targetW - this.src.X
	}

	if this.src.Y+this.src.H > targetH {
		this.src.H = targetH - this.src.Y
	}

	return nil
}

func (this *Sprite) GetClip() rect.Rect {
	return this.src
}

func (this *Sprite) GetGraphicsWidth() (int, error) {
	return this.image.GetWidth()
}

func (this *Sprite) GetGraphicsHeight() (int, error) {
	return this.image.GetHeight()
}

func (this *Sprite) SetOffset(offset point.Point) {
	this.offset = offset
}

func (this *Sprite) GetOffset() point.Point {
	return this.offset
}

func (this *Sprite) SetDestFromRect(dest rect.Rect) {
	this.dest.X = dest.X
	this.dest.Y = dest.Y
}

func (this *Sprite) SetDestFromPoint(dest point.Point) {
	this.dest.X = dest.X
	this.dest.Y = dest.Y
}

func (this *Sprite) GetDest() point.Point {
	return this.dest
}

func (this *Sprite) SetDest(x, y int) {
	this.dest.X = x
	this.dest.Y = y
}

func (this *Sprite) GetSrc() rect.Rect {
	return this.src
}

func (this *Sprite) GetGraphics() common.Image {
	return this.image
}

func (this *Sprite) SetLocalFrame(localframe rect.Rect) {
	this.localFrame = localframe
}

func (this *Sprite) GetLocalFrame() rect.Rect {
	return this.localFrame
}

func (this *Sprite) ColorMod() color.Color {
	return this.colorMod
}

func (this *Sprite) AlphaMod() uint8 {
	return this.alphaMod
}

func (this *Sprite) SetAlphaMod(alphaMod uint8) {
	this.alphaMod = alphaMod
}
