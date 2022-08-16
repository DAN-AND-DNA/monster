package sdlhardware

import (
	"monster/pkg/allocs"
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define/renderable"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/filesystem/logfile"
	"monster/pkg/subengine/cursormanager"
	"monster/pkg/subengine/render/base"
	"monster/pkg/utils"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type RenderDevice struct {
	base.RenderDevice
	window          *sdl.Window
	renderer        *sdl.Renderer
	titlebarIcon    *sdl.Surface
	title           string
	backgroundColor color.Color
	gammaR          *[256]uint16
	gammaG          *[256]uint16
	gammaB          *[256]uint16
	curs            common.CursorManager
	texture         *sdl.Texture
}

func NewRenderDevice(settings common.Settings, eset common.EngineSettings) *RenderDevice {
	impl := ConstructRenderDevice(settings, eset)
	ptr := &impl
	_ = (common.RenderDevice)(ptr)

	return ptr
}

func ConstructRenderDevice(settings common.Settings, eset common.EngineSettings) RenderDevice {
	impl := RenderDevice{
		backgroundColor: color.Construct(0, 0, 0),
		gammaR:          &[256]uint16{},
		gammaG:          &[256]uint16{},
		gammaB:          &[256]uint16{},
	}

	// base
	impl.RenderDevice = base.ConstructRenderDevice()

	strDriver, err := sdl.GetCurrentVideoDriver()
	if err != nil {
		panic(err)
	}

	logfile.LogInfo("Using Render Device: SDLHardwareRenderDevice (hardware, SDL 2, %s)", strDriver)

	// self
	impl.Fullscreen = settings.Get("fullscreen").(bool)
	impl.Hwsurface = settings.Get("hwsurface").(bool)
	impl.Vsync = settings.Get("vsync").(bool)
	impl.TextureFilter = settings.Get("texture_filter").(bool)
	impl.MinScreen.X = eset.Get("resolutions", "required_width").(int)  // mod要求的最小宽
	impl.MinScreen.Y = eset.Get("resolutions", "required_height").(int) // mod要求的最小高

	desktop, err := sdl.GetDesktopDisplayMode(0)
	if err != nil {
		panic(err)
	}

	num, err := sdl.GetNumVideoDisplays()
	if err != nil {
		panic(err)
	}

	logfile.LogInfo("RenderDevice: %d display(s), using display 0 (%dx%d @ %dhz)", num, desktop.W, desktop.H, desktop.RefreshRate)

	return impl
}

func (this *RenderDevice) CreateContext(settings common.Settings, eset common.EngineSettings, msg common.MessageEngine, mods common.ModManager) error {
	return this.RenderDevice.CreateContext(this, settings, eset, msg, mods)
}

// 初始化或者修改了配置则会反复调用该函数，调整配置和渲染，窗口等
func (this *RenderDevice) CreateContextInternal(settings common.Settings, eset common.EngineSettings, msg common.MessageEngine, mods common.ModManager) error {

	// if is windows then set opengl
	setRenderDriver()

	s := settings
	settingsChanged := ((this.Fullscreen != s.Get("fullscreen").(bool)) ||
		this.DestructiveFullscreen ||
		this.Hwsurface != s.Get("hwsurface").(bool) ||
		this.TextureFilter != s.Get("texture_filter").(bool) ||
		this.IgnoreTextureFilter != eset.Get("resolutions", "ignore_texture_filter").(bool))

	wFlags := 0
	rFlags := 0
	windowW := s.Get("resolution_w").(int)
	windowH := s.Get("resolution_h").(int)
	var err error

	if s.Get("fullscreen").(bool) {
		wFlags = wFlags | sdl.WINDOW_FULLSCREEN_DESKTOP

		// 全屏时获取桌面的信息
		// make the window the same size as the desktop resolution
		desktop, err := sdl.GetDesktopDisplayMode(0)
		if err != nil {
			return err
		}
		windowW = (int)(desktop.W)
		windowH = (int)(desktop.H)
	} else if this.Fullscreen && this.IsInitialized {

		// 配置文件修改变成窗口模式
		// if the game was previously in fullscreen, resize the window when returning to windowed mode
		windowW = eset.Get("resolutions", "required_width").(int) // mod要求的最小宽度
		windowH = eset.Get("resolutions", "required_height").(int)
		wFlags = wFlags | sdl.WINDOW_SHOWN

	} else {
		wFlags = wFlags | sdl.WINDOW_SHOWN
	}

	wFlags = wFlags | sdl.WINDOW_RESIZABLE

	if s.Get("hwsurface").(bool) {
		// 配置文件配置硬件加速
		rFlags = sdl.RENDERER_ACCELERATED | sdl.RENDERER_TARGETTEXTURE
	} else {
		rFlags = sdl.RENDERER_SOFTWARE | sdl.RENDERER_TARGETTEXTURE
		s.Set("vsync", false) // can't have software mode & vsync at the same time
	}

	if s.Get("vsync").(bool) {
		// 垂直同步
		rFlags = rFlags | sdl.RENDERER_PRESENTVSYNC
	}

	if settingsChanged || !this.IsInitialized {
		// 配置被修改或第一次创建

		// 清理
		this.DestroyContext()

		// 创建窗口
		//this.window, err = sdl.CreateWindow("", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, (int32)(windowW), (int32)(windowH), (uint32)(wFlags))
		this.window, err = allocs.SdlCreateWindow("", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, (int32)(windowW), (int32)(windowH), (uint32)(wFlags))
		if err != nil {
			return err
		}

		// 创建渲染
		//this.renderer, err = sdl.CreateRenderer(this.window, -1, (uint32)(rFlags))
		this.renderer, err = allocs.SdlCreateRenderer(this.window, -1, (uint32)(rFlags))
		if err != nil {
			return err
		}

		// 纹理过滤
		if s.Get("texture_filter").(bool) && !eset.Get("resolutions", "ignore_texture_filter").(bool) {
			sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")
		} else {
			sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "0")
		}

		this.WindowResize(s, eset)

		// 配置窗口为mod最小的大小
		this.window.SetMinimumSize((int32)(eset.Get("resolutions", "required_width").(int)), (int32)(eset.Get("resolutions", "required_height").(int)))

		// setting minimum size might move the window, so set position again
		this.window.SetPosition(sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED)

		if !this.IsInitialized {

			// save the system gamma levels if we just created the window
			this.gammaR, this.gammaG, this.gammaB, err = this.window.GetGammaRamp()
			if err != nil {
				return err
			}
			logfile.LogInfo("RenderDevice: Window size is %dx%d", s.Get("resolution_w").(int), s.Get("resolution_h").(int))

		}

		this.Fullscreen = s.Get("fullscreen").(bool)
		this.Hwsurface = s.Get("hwsurface").(bool)
		this.Vsync = s.Get("vsync").(bool)
		this.TextureFilter = s.Get("texture_filter").(bool)
		this.IgnoreTextureFilter = eset.Get("resolutions", "ignore_texture_filter").(bool)
		this.IsInitialized = true

		logfile.LogInfo("RenderDevice: Fullscreen=%v, Hardware surfaces=%v, Vsync=%v, Texture Filter=%v", this.Fullscreen, this.Hwsurface, this.Vsync, this.TextureFilter)
		ddpi, _, _, err := sdl.GetDisplayDPI(0)
		if err != nil {
			return err
		}
		logfile.LogInfo("RenderDevice: Display DPI is %f", ddpi)

	}

	// 以mod的最小值作为窗口最小
	// update minimum window size if it has changed
	if this.MinScreen.X != eset.Get("resolutions", "required_width").(int) || this.MinScreen.Y != eset.Get("resolutions", "required_height").(int) {
		this.MinScreen.X = eset.Get("resolutions", "required_width").(int)
		this.MinScreen.Y = eset.Get("resolutions", "required_height").(int)
		this.window.SetMinimumSize((int32)(this.MinScreen.X), (int32)(this.MinScreen.Y))
		this.window.SetPosition(sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED)
	}

	this.WindowResize(s, eset)

	// update title bar text and icon
	this.UpdateTitleBar(s, eset, msg, mods)

	// TODO
	// icons

	if this.curs != nil {
		this.curs.Close()
	}
	this.curs = cursormanager.New(s, mods, this)

	// 修改伽马值
	if s.Get("change_gamma").(bool) {
		this.SetGamma(s.Get("gamma").(float32))
	} else {
		this.ResetGamma()
		s.Set("change_gamma", false)
		s.Set("gamma", float32(1))
	}

	return nil
}

func (this *RenderDevice) DestroyContext() {
	this.ResetGamma()

	this.RenderDevice.CacheRemoveAll()
	this.IsReloadGraphics = true // 设置已经重置过渲染系统

	//TODO
	// icons

	if this.curs != nil {
		this.curs.Close()
		this.curs = nil
	}

	if this.titlebarIcon != nil {
		//this.titlebarIcon.Free()
		allocs.Delete(this.titlebarIcon)
		this.titlebarIcon = nil
	}

	if this.texture != nil {
		//this.texture.Destroy()
		allocs.Delete(this.texture)
		this.texture = nil
	}

	if this.renderer != nil {
		//this.renderer.Destroy()
		allocs.Delete(this.renderer)
		this.renderer = nil
	}

	if this.window != nil {
		//this.window.Destroy()
		allocs.Delete(this.window)
		this.window = nil
	}

	this.title = ""
}

func (this *RenderDevice) Clear() {
	this.DestroyContext()
}

func (this *RenderDevice) Close() {
	this.RenderDevice.Close(this)
}

func (this *RenderDevice) CreateContextError() {
	logfile.LogError("SDLHardwareRenderDevice: createContext() failed: %s", sdl.GetError())
	logfile.LogErrorDialog("SDLHardwareRenderDevice: createContext() failed: %s", sdl.GetError())
}

// 修改分辨率和视口，重建素材, 更新配置
func (this *RenderDevice) WindowResize(settings common.Settings, eset common.EngineSettings) error {
	this.RenderDevice.WindowResizeInternal(this, settings, eset)
	this.renderer.SetLogicalSize((int32)(settings.GetViewW()), (int32)(settings.GetViewH()))
	if this.texture != nil {
		//this.texture.Destroy()
		allocs.Delete(this.texture)
		this.texture = nil
	}

	var err error
	//this.texture, err = this.renderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_TARGET, (int32)(settings.GetViewW()), (int32)(settings.GetViewH()))
	this.texture, err = allocs.SdlCreateTexture(this.renderer, sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_TARGET, (int32)(settings.GetViewW()), (int32)(settings.GetViewH()))
	if err != nil {
		return err
	}

	if err := this.renderer.SetRenderTarget(this.texture); err != nil {
		return err
	}

	settings.UpdateScreenVars(eset)
	return nil
}

func (this *RenderDevice) UpdateTitleBar(settings common.Settings, eset common.EngineSettings, msg common.MessageEngine, mods common.ModManager) error {
	if this.title != "" {
		this.title = ""
	}

	if this.titlebarIcon != nil {
		//this.titlebarIcon.Free()
		allocs.Delete(this.titlebarIcon)
		this.titlebarIcon = nil
	}

	if this.window == nil {
		return nil
	}

	this.title = msg.Get(eset.Get("misc", "window_title").(string))
	loc, err := mods.Locate(settings, "images/logo/icon.png")
	if err != nil && utils.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	//this.titlebarIcon, err = img.Load(loc)
	this.titlebarIcon, err = allocs.ImgLoad(loc)
	if err != nil {
		return err
	}

	if this.title != "" {
		this.window.SetTitle(this.title)
	}

	if this.titlebarIcon != nil {
		this.window.SetIcon(this.titlebarIcon)
	}

	return nil

}

func (this *RenderDevice) SetGamma(g float32) error {
	ramp := [256]uint16{}
	sdl.CalculateGammaRamp(g, &ramp)
	return this.window.SetGammaRamp(&ramp, &ramp, &ramp)
}

func (this *RenderDevice) ResetGamma() error {
	return this.window.SetGammaRamp(this.gammaR, this.gammaG, this.gammaB)
}

func (this *RenderDevice) GetWindowSize() (int, int) {
	w, h := this.window.GetSize()
	return int(w), int(h)
}

func (this *RenderDevice) Render1(r common.Renderable, dest rect.Rect) error {
	dest.W = r.GetSrc().W
	dest.H = r.GetSrc().H

	var src, _dest sdl.Rect
	src.X = (int32)(r.GetSrc().X)
	src.Y = (int32)(r.GetSrc().Y)
	src.W = (int32)(r.GetSrc().W)
	src.H = (int32)(r.GetSrc().H)

	_dest.X = (int32)(dest.X)
	_dest.Y = (int32)(dest.Y)
	_dest.W = (int32)(dest.W)
	_dest.H = (int32)(dest.H)

	var err error

	err = this.renderer.SetRenderTarget(this.texture)
	if err != nil {
		return err
	}

	surface := r.GetImage().Surface().(*sdl.Texture)
	if r.GetBlendMode() == renderable.BLEND_ADD {
		err = surface.SetBlendMode(sdl.BLENDMODE_ADD)
		if err != nil {
			return err
		}
	} else {
		err = surface.SetBlendMode(sdl.BLENDMODE_BLEND)
		if err != nil {
			return err
		}
	}

	err = surface.SetColorMod(r.GetColorMod().R, r.GetColorMod().G, r.GetColorMod().B)
	if err != nil {
		return err
	}

	err = surface.SetAlphaMod(r.GetAlphaMod())
	if err != nil {
		return err
	}

	err = this.renderer.Copy(surface, &src, &_dest)
	if err != nil {
		return err
	}

	return nil
}

func (this *RenderDevice) Render(r common.Sprite) error {
	if r == nil || !this.LocalToGlobal(r) {
		return nil
	}

	if this.MClip.X < 0 {
		this.MClip.W += this.MClip.X // 宽度缩短x
		this.MDest.X -= this.MClip.X // 目标右移动x
		this.MClip.X = 0
	}

	if this.MClip.Y < 0 {
		this.MClip.H += this.MClip.Y // 高度缩短y
		this.MDest.Y -= this.MClip.Y // 目标向下移动y
		this.MClip.Y = 0
	}

	this.MDest.W = this.MClip.W
	this.MDest.H = this.MClip.H

	var src, dest sdl.Rect
	src.X = (int32)(this.MClip.X)
	src.Y = (int32)(this.MClip.Y)
	src.W = (int32)(this.MClip.W)
	src.H = (int32)(this.MClip.H)

	dest.X = (int32)(this.MDest.X)
	dest.Y = (int32)(this.MDest.Y)
	dest.W = (int32)(this.MDest.W)
	dest.H = (int32)(this.MDest.H)

	surface := r.GetGraphics().Surface().(*sdl.Texture)
	err := surface.SetColorMod(r.ColorMod().R, r.ColorMod().G, r.ColorMod().B)
	if err != nil {
		return err
	}

	err = surface.SetAlphaMod(r.AlphaMod())
	if err != nil {
		return err
	}

	err = this.renderer.SetRenderTarget(this.texture)
	if err != nil {
		return err
	}

	// src 和 dest为空就是全部，然后就是拉伸来填充
	// 文档： the texture will be stretched to fill the given rectangle
	err = this.renderer.Copy(surface, &src, &dest)
	if err != nil {
		return err
	}

	return nil
}

// 把src图片绘制到dest图片上
func (this *RenderDevice) RenderToImage(srcImage common.Image, src rect.Rect, destImage common.Image, dest rect.Rect) (rect.Rect, error) {
	if srcImage == nil || destImage == nil {
		panic("nil")
		return dest, nil
	}

	destSurface := destImage.Surface().(*sdl.Texture)
	err := this.renderer.SetRenderTarget(destSurface)
	if err != nil {
		return dest, err
	}

	dest.W = src.W
	dest.H = src.H

	var _src, _dest sdl.Rect
	_src.X = (int32)(src.X)
	_src.Y = (int32)(src.Y)
	_src.W = (int32)(src.W)
	_src.H = (int32)(src.H)

	_dest.X = (int32)(dest.X)
	_dest.Y = (int32)(dest.Y)
	_dest.W = (int32)(dest.W)
	_dest.H = (int32)(dest.H)

	srcSurface := srcImage.Surface().(*sdl.Texture)

	err = destSurface.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		return dest, err
	}

	// src 和 dest为空就是全部，然后就是拉伸来填充
	// 文档： the texture will be stretched to fill the given rectangle
	err = this.renderer.Copy(srcSurface, &_src, &_dest)
	if err != nil {
		return dest, err
	}

	err = this.renderer.SetRenderTarget(nil)
	if err != nil {
		return dest, err
	}

	return dest, nil
}

// 以字符串创建图片
func (this *RenderDevice) RenderTextToImage(fontStyle common.FontStyle, text string, color color.Color, blended bool) (common.Image, error) {
	image := newImage(this, "", this.renderer) // +1
	defer image.UnRef()                        // -1

	var cleanup *sdl.Surface
	var err error

	_color := sdl.Color{color.R, color.G, color.B, color.A}

	if blended {
		//cleanup, err = fontStyle.Ttfont().(*ttf.Font).RenderUTF8Blended(text, _color)
		cleanup, err = allocs.FontRenderUTF8Blended(fontStyle.Ttfont().(*ttf.Font), text, _color)
		if err != nil {
			return nil, err
		}
	} else {
		//cleanup, err = fontStyle.Ttfont().(*ttf.Font).RenderUTF8Solid(text, _color)
		cleanup, err = allocs.FontRenderUTF8Solid(fontStyle.Ttfont().(*ttf.Font), text, _color)
		if err != nil {
			return nil, err
		}
	}

	defer allocs.Delete(cleanup)
	//defer cleanup.Free()

	//surface, err := this.renderer.CreateTextureFromSurface(cleanup)
	surface, err := allocs.SdlCreateTextureFromSurface(this.renderer, cleanup)
	if err != nil {
		return nil, err
	}

	image.SetSurface(surface)
	image.Ref() // +1

	return image, nil
}

func (this *RenderDevice) DrawPixel(x, y int, color color.Color) error {
	err := this.renderer.SetDrawColor(color.R, color.G, color.B, color.A)
	if err != nil {
		return err
	}

	err = this.renderer.DrawPoint(int32(x), int32(y))
	if err != nil {
		return err
	}

	return nil
}

func (this *RenderDevice) DrawLine(x0, y0, x1, y1 int, color color.Color) error {
	err := this.renderer.SetDrawColor(color.R, color.G, color.B, color.A)
	if err != nil {
		return err
	}

	err = this.renderer.DrawLine(int32(x0), int32(y0), int32(x1), int32(y1))
	if err != nil {
		return err
	}
	return nil
}

func (this *RenderDevice) DrawRectangle(p0, p1 point.Point, color color.Color) error {
	err := this.renderer.SetDrawColor(color.R, color.G, color.B, color.A)
	if err != nil {
		return err
	}

	err = this.DrawLine(p0.X, p0.Y, p1.X, p0.Y, color)
	if err != nil {
		return err
	}
	err = this.DrawLine(p1.X, p0.Y, p1.X, p1.Y, color)
	if err != nil {
		return err
	}
	err = this.DrawLine(p0.X, p0.Y, p0.X, p1.Y, color)
	if err != nil {
		return err
	}

	err = this.DrawLine(p0.X, p1.Y, p1.X, p1.Y, color)
	if err != nil {
		return err
	}

	return nil
}

func (this *RenderDevice) BlankScreen() error {
	err := this.renderer.SetRenderTarget(nil)
	if err != nil {
		return err
	}

	err = this.renderer.SetDrawColor(0, 0, 0, 255)
	if err != nil {
		return err
	}

	err = this.renderer.Clear()
	if err != nil {
		return err
	}

	err = this.renderer.SetDrawColor(this.backgroundColor.R, this.backgroundColor.G, this.backgroundColor.B, this.backgroundColor.A)
	if err != nil {
		return err
	}

	err = this.renderer.SetRenderTarget(this.texture)
	if err != nil {
		return err
	}

	err = this.renderer.Clear()
	if err != nil {
		return err
	}

	return nil
}

func (this *RenderDevice) CommitFrame(inpt common.InputState) error {
	err := this.renderer.SetRenderTarget(nil)
	if err != nil {
		return err
	}

	err = this.renderer.Copy(this.texture, nil, nil)
	if err != nil {
		return err
	}

	this.renderer.Present()
	inpt.SetWindowResized(false)

	return nil
}

func (this *RenderDevice) CreateImage(width, height int) (common.Image, error) {
	image := newImage(this, "", this.renderer) // +1
	defer image.UnRef()                        // -1

	//FIXME
	if width > 0 && height > 0 {
		//surface, err := this.renderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_TARGET, (int32)(width), int32(height))
		surface, err := allocs.SdlCreateTexture(this.renderer, sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_TARGET, (int32)(width), int32(height))
		if err != nil {
			logfile.LogError("SDLHardwareRenderDevice: SDL_CreateTexture failed: %s", sdl.GetError())
			return nil, err
		}

		image.SetSurface(surface)

		err = this.renderer.SetRenderTarget(surface)
		if err != nil {
			return nil, err
		}

		err = surface.SetBlendMode(sdl.BLENDMODE_BLEND)
		if err != nil {
			return nil, err
		}

		err = this.renderer.SetDrawColor(0, 0, 0, 0)
		if err != nil {
			return nil, err
		}

		err = this.renderer.Clear()
		if err != nil {
			return nil, err
		}

		err = this.renderer.SetRenderTarget(nil)
		if err != nil {
			return nil, err
		}
	}

	image.Ref() // +1
	return image, nil
}

func (this *RenderDevice) LoadImage(settings common.Settings, mods common.ModManager, filename string) (common.Image, error) {
	image, ok := this.CacheLookup(filename) // +1
	if ok {
		return image, nil
	}

	image = newImage(this, filename, this.renderer) // +1
	defer image.UnRef()                             // -1

	loc, err := mods.Locate(settings, filename)
	if err != nil {
		return nil, err
	}

	//surface, err := img.LoadTexture(this.renderer, loc)
	surface, err := allocs.ImgLoadTexture(this.renderer, loc)
	if err != nil {
		logfile.LogError("SDLHardwareRenderDevice: Couldn't load image: '%s'. %s", filename, img.GetError())
		logfile.LogErrorDialog("SDLHardwareRenderDevice: Couldn't load image: '%s'.\n%s", filename, img.GetError())
		return nil, err
	}

	image.Ref() // +1
	image.SetSurface(surface)
	this.CacheStore(filename, image)
	return image, nil
}

func (this *RenderDevice) SetBackgroundColor(color color.Color) {
	this.backgroundColor = color
	this.backgroundColor.A = 255
}

func (this *RenderDevice) SetFullscreen(settings common.Settings, eset common.EngineSettings, platform common.Platform, enable bool) {
	if !this.DestructiveFullscreen {
		if enable {
			if platform.FullscreenBypass() {
				platform.SetFullscreen(true)
			} else {
				this.window.SetFullscreen(sdl.WINDOW_FULLSCREEN_DESKTOP)
			}

			this.Fullscreen = true

		} else if this.Fullscreen {

			if platform.FullscreenBypass() {
				platform.SetFullscreen(false)
			} else {
				this.window.SetFullscreen(0)

				// 恢复到默认大小
				this.window.SetMinimumSize((int32)(eset.Get("resolutions", "required_width").(int)), (int32)(eset.Get("resolutions", "required_height").(int)))
				this.window.SetSize((int32)(eset.Get("resolutions", "required_width").(int)), (int32)(eset.Get("resolutions", "required_height").(int)))
				this.WindowResize(settings, eset)
				this.window.SetPosition(sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED)
			}

			this.Fullscreen = false
		}

		this.WindowResize(settings, eset)
	}
}

func (this *RenderDevice) GetRefreshRate() int {
	mod, err := sdl.GetCurrentDisplayMode(0)
	if err != nil {
		return 0
	}

	return (int)(mod.RefreshRate)
}

func (this *RenderDevice) FillRect() error {
	surface, err := this.window.GetSurface()
	if err != nil {
		return err
	}

	//surface.FillRect(nil, 0)

	rect := sdl.Rect{0, 0, 200, 200}
	surface.FillRect(&rect, 0xffff0000)
	this.window.UpdateSurface()

	return nil
}

func (this *RenderDevice) Curs() common.CursorManager {
	return this.curs
}
