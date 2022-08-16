package base

import (
	"monster/pkg/common"
)

// 和sdl捆绑，重建sdl时需要销毁
type Image struct {
	filename   string
	device     common.RenderDevice
	refCounter uint64
}

func ConstructImage(device common.RenderDevice, filename string) Image {
	image := Image{
		filename:   filename,
		device:     device,
		refCounter: 1,
	}

	//fmt.Println("image: ", image.GetFilename(), "new", image.GetRefCount())
	return image
}

func (this *Image) Close(impl common.Image) {
	// 先子类清理
	impl.Clear()

	// 后base清理
	this.clear()
}

// 清理自己
func (this *Image) clear() {
	this.device.FreeImage(this.filename) // 只清理缓存
}

func (this *Image) GetFilename() string {
	return this.filename
}

// 创建和图片等大的精灵
func (this *Image) CreateSprite(impl common.Image) (common.Sprite, error) {
	sprite := NewSprite(impl) // +1
	defer sprite.Close()

	w, err := impl.GetWidth()
	if err != nil {
		return nil, err
	}

	h, err := impl.GetHeight()
	if err != nil {
		return nil, err
	}

	sprite.SetClip(0, 0, w, h)
	sprite.KeepAlive()
	return sprite, nil
}

func (this *Image) Ref() {
	this.refCounter++
}

func (this *Image) UnRef(impl common.Image) {
	if this.refCounter <= 0 {
		return
	}

	this.refCounter--
	if this.refCounter == 0 {
		this.Close(impl)
	}
}

func (this *Image) GetRefCount() uint64 {
	return this.refCounter
}

func (this *Image) SetRefCount(rc uint64) {
	this.refCounter = rc
}

func (this *Image) GetDevice() common.RenderDevice {
	return this.device
}
