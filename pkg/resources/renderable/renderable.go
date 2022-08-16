package renderable

import (
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define/renderable"
	"monster/pkg/common/fpoint"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
)

type Renderable struct {
	image     common.Image  // 只是钩子不负责销毁
	src       rect.Rect     // 定位sprite在图片的位置
	mapPos    fpoint.FPoint // 地图位置
	offset    point.Point   // 和sprite左上角坐标原点的偏移点，该点作为sprite在地图的位置
	prio      uint64        // 64-32 bit for map position, 31-16 for intertile position, 15-0 user dependent, such as Avatar
	blendMode uint8
	colorMod  color.Color
	alphaMod  uint8
	type1     uint8
}

func Init() common.Renderable {
	r := &Renderable{
		blendMode: renderable.BLEND_NORMAL,
		colorMod:  color.Construct(255, 255, 255),
		alphaMod:  255,
		type1:     renderable.TYPE_NORMAL,
	}
	return r
}

func (this *Renderable) SetImage(image common.Image) {
	this.image = image
}

func (this *Renderable) GetImage() common.Image {
	return this.image
}

func (this *Renderable) SetSrc(src rect.Rect) {
	this.src = src
}

func (this *Renderable) GetSrc() rect.Rect {
	return this.src
}

func (this *Renderable) SetMapPos(mapPos fpoint.FPoint) {
	this.mapPos = mapPos
}

func (this *Renderable) GetMapPos() fpoint.FPoint {
	return this.mapPos
}

func (this *Renderable) SetOffset(offset point.Point) {
	this.offset = offset
}

func (this *Renderable) GetOffset() point.Point {
	return this.offset
}

func (this *Renderable) SetPrio(prio uint64) {
	this.prio = prio
}

func (this *Renderable) GetPrio() uint64 {
	return this.prio
}

func (this *Renderable) SetBlendMode(blendMode uint8) {
	this.blendMode = blendMode
}

func (this *Renderable) GetBlendMode() uint8 {
	return this.blendMode
}

func (this *Renderable) SetColorMod(colorMod color.Color) {
	this.colorMod = colorMod
}

func (this *Renderable) GetColorMod() color.Color {
	return this.colorMod
}

func (this *Renderable) SetAlphaMod(alphaMod uint8) {
	this.alphaMod = alphaMod
}

func (this *Renderable) GetAlphaMod() uint8 {
	return this.alphaMod
}

func (this *Renderable) SetType(type1 uint8) {
	this.type1 = type1
}

func (this *Renderable) GetType() uint8 {
	return this.type1
}
