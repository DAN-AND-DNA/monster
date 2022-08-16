package sdlhardware

import (
	"errors"
	"monster/pkg/allocs"
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/subengine/render/base"

	"github.com/veandco/go-sdl2/sdl"
)

var (
	Err_bad_args_in_sdlhardwareimage_resize = errors.New("bad args in sdl hardware image resize")
)

// 和sdl捆绑，重建sdl时需要销毁
type Image struct {
	base.Image
	renderer          *sdl.Renderer // 外部指针，不负责销毁
	surface           *sdl.Texture  // 代表该图片本身
	pixelBatchSurface *sdl.Surface
}

// 不允许外部使用
func newImage(device common.RenderDevice, filename string, renderer *sdl.Renderer) *Image {
	ptr := &Image{}
	ptr.init(device, filename, renderer)
	return ptr
}

func (this *Image) init(device common.RenderDevice, filename string, renderer *sdl.Renderer) common.Image {
	// 先base初始化
	this.Image = base.ConstructImage(device, filename)

	// 后子类初始化
	this.renderer = renderer

	return this

}

// 清理自己
func (this *Image) Clear() {
	this.renderer = nil

	if this.surface != nil {
		//this.surface.Destroy()
		allocs.Delete(this.surface)
	}

	if this.pixelBatchSurface != nil {
		//this.pixelBatchSurface.Free()
		allocs.Delete(this.pixelBatchSurface)
	}
}

func (this *Image) Close() {
	this.Image.Close(this)
}

func (this *Image) GetWidth() (int, error) {
	_, _, w, _, err := this.surface.Query()
	if err != nil {
		return 0, nil
	}

	return (int)(w), nil
}

func (this *Image) GetHeight() (int, error) {
	_, _, _, h, err := this.surface.Query()
	if err != nil {
		return 0, err
	}

	return (int)(h), nil
}

func (this *Image) CreateSprite() (common.Sprite, error) {
	return this.Image.CreateSprite(this)
}

func (this *Image) UnRef() {
	this.Image.UnRef(this)
}

func (this *Image) FillWithColor(color color.Color) error {
	if this.surface == nil {
		return nil
	}

	err := this.renderer.SetRenderTarget(this.surface)
	if err != nil {
		return err
	}

	err = this.surface.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		return err
	}

	err = this.renderer.SetDrawColor(color.R, color.G, color.B, color.A)
	if err != nil {
		return err
	}

	err = this.renderer.Clear()
	if err != nil {
		return err
	}

	err = this.renderer.SetRenderTarget(nil)
	if err != nil {
		return err
	}

	return nil
}

func (this *Image) DrawPixel(x int, y int, color color.Color) error {
	if this.surface == nil || x < 0 || y < 0 {
		return nil
	}

	w, err := this.GetWidth()
	if err != nil {
		return err
	}

	if x >= w {
		return nil
	}

	h, err := this.GetHeight()
	if err != nil {
		return err
	}

	if y >= h {
		return nil
	}

	if this.pixelBatchSurface != nil {
		if this.pixelBatchSurface.MustLock() {
			this.pixelBatchSurface.Lock()
		}

		this.pixelBatchSurface.Set(x, y, &color)

		if this.pixelBatchSurface.MustLock() {
			this.pixelBatchSurface.Unlock()
		}
	} else {
		err := this.renderer.SetRenderTarget(this.surface)
		if err != nil {
			return err
		}

		err = this.surface.SetBlendMode(sdl.BLENDMODE_BLEND)
		if err != nil {
			return err
		}

		err = this.renderer.SetDrawColor(color.R, color.G, color.B, color.A)
		if err != nil {
			return err
		}

		err = this.renderer.DrawPoint(int32(x), int32(y))
		if err != nil {
			return err
		}

		err = this.renderer.SetRenderTarget(nil)
		if err != nil {
			return err
		}

	}

	return nil
}

func (this *Image) DrawLine(x0, y0, x1, y1 int, color color.Color) error {
	err := this.renderer.SetRenderTarget(this.surface)
	if err != nil {
		return err
	}

	err = this.surface.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		return err
	}

	err = this.renderer.SetDrawColor(color.R, color.G, color.B, color.A)
	if err != nil {
		return err
	}

	err = this.renderer.DrawLine(int32(x0), int32(y0), int32(x1), int32(y1))
	if err != nil {
		return err
	}

	err = this.renderer.SetRenderTarget(nil)
	if err != nil {
		return err
	}

	return nil
}

func (this *Image) BeginPixelBatch() error {
	if this.surface == nil {
		return nil
	}

	if this.pixelBatchSurface != nil {
		//this.pixelBatchSurface.Free()
		allocs.Delete(this.pixelBatchSurface)
	}

	w, err := this.GetWidth()
	if err != nil {
		return err
	}

	h, err := this.GetHeight()
	if err != nil {
		return err
	}

	var rmask, gmask, bmask, amask uint32

	if sdl.BYTEORDER == sdl.BIG_ENDIAN {
		rmask = 0xff000000
		gmask = 0x00ff0000
		bmask = 0x0000ff00
		amask = 0x000000ff
	} else {
		rmask = 0x000000ff
		gmask = 0x0000ff00
		bmask = 0x00ff0000
		amask = 0xff000000
	}

	//this.pixelBatchSurface, err = sdl.CreateRGBSurface(0, (int32)(w), (int32)(h), 32, rmask, gmask, bmask, amask)
	this.pixelBatchSurface, err = allocs.SdlCreateRGBSurface(0, (int32)(w), (int32)(h), 32, rmask, gmask, bmask, amask)
	if err != nil {
		return err
	}

	return nil

}

func (this *Image) EndPixelBatch() error {
	if this.surface == nil || this.pixelBatchSurface == nil {
		return nil
	}

	pixelBatchTexture, err := this.renderer.CreateTextureFromSurface(this.pixelBatchSurface)
	if err != nil {
		return err
	}
	defer pixelBatchTexture.Destroy()

	err = this.renderer.SetRenderTarget(this.surface)
	if err != nil {
		return err
	}

	err = this.surface.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		return err
	}

	err = this.renderer.Copy(pixelBatchTexture, nil, nil)
	if err != nil {
		return err
	}

	err = this.renderer.SetRenderTarget(nil)
	if err != nil {
		return err
	}

	this.pixelBatchSurface.Free()
	this.pixelBatchSurface = nil
	return nil
}

func (this *Image) Resize(width, height int) (common.Image, error) {
	if this.surface == nil || width <= 0 || height <= 0 {
		return nil, Err_bad_args_in_sdlhardwareimage_resize
	}

	var err error
	scaled := newImage(this.GetDevice(), this.GetFilename(), this.renderer) // +1
	defer scaled.UnRef()                                                    // -1

	//scaled.surface, err = this.renderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_TARGET, (int32)(width), (int32)(height))
	scaled.surface, err = allocs.SdlCreateTexture(this.renderer, sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_TARGET, (int32)(width), (int32)(height))
	if err != nil {
		return nil, err
	}

	err = this.renderer.SetRenderTarget(scaled.surface)
	if err != nil {
		return nil, err
	}

	err = this.renderer.CopyEx(this.surface, nil, nil, 0, nil, sdl.FLIP_NONE)
	if err != nil {
		return nil, err
	}

	err = this.renderer.SetRenderTarget(nil)
	if err != nil {
		return nil, err
	}

	this.UnRef() //清理老的
	scaled.Ref() // +1
	return scaled, nil
}

func (this *Image) Surface() interface{} {
	return this.surface
}

func (this *Image) SetSurface(s interface{}) {
	this.surface = s.(*sdl.Texture)
}
