package base

import (
	"monster/pkg/common"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/filesystem/logfile"
)

type RenderDevice struct {
	Fullscreen            bool
	Hwsurface             bool
	Vsync                 bool
	TextureFilter         bool
	IgnoreTextureFilter   bool
	MinScreen             point.Point
	DestructiveFullscreen bool
	IsInitialized         bool
	IsReloadGraphics      bool
	ddpi                  float32
	cache                 map[string]common.Image
	MClip                 rect.Rect
	MDest                 rect.Rect
}

func ConstructRenderDevice() RenderDevice {
	r := RenderDevice{
		MinScreen: point.Construct(640, 480),
		cache:     map[string]common.Image{},
	}

	return r
}

func (this *RenderDevice) Clear() {
}

func (this *RenderDevice) Close(impl common.RenderDevice) {
	//
	impl.Clear()

	// 自己
	this.Clear()
}

func (this *RenderDevice) CreateContext(impl common.RenderDevice, settings common.Settings, eset common.EngineSettings, msg common.MessageEngine, mods common.ModManager) error {
	if err := impl.CreateContextInternal(settings, eset, msg, mods); err != nil {
		impl.CreateContextError()
		return err
	}

	return nil
}

func (this *RenderDevice) CacheLookup(filename string) (common.Image, bool) {
	if val, ok := this.cache[filename]; ok {
		val.Ref()
		return val, true
	}

	return nil, false
}

func (this *RenderDevice) CacheStore(filename string, image common.Image) {
	if image == nil || filename == "" {
		return
	}

	this.cache[filename] = image
}

func (this *RenderDevice) CacheRemove(filename string) {
	if filename == "" {
		return
	}
	delete(this.cache, filename)
}

func (this *RenderDevice) CacheRemoveAll() {
	this.cache = map[string]common.Image{}
}

func (this *RenderDevice) LocalToGlobal(r common.Sprite) bool {
	this.MClip = r.GetSrc()
	/*
		0-----------------> x
		|
		|
		|
		|
		|
		|
		|
		v
		y
	*/

	// 计算在全局显示的坐标
	left := r.GetDest().X - r.GetOffset().X // 地图位置的X - sprite换算地图左边的偏移的X = 左边位置 ( <0 左移  >0 右移 )
	right := left + this.MClip.W
	up := r.GetDest().Y - r.GetOffset().Y
	down := up + this.MClip.H

	// 是子组件
	if r.GetLocalFrame().W > 0 {

		// 向右边渲染
		if left > r.GetLocalFrame().W {
			return false //太宽
		}

		if right < 0 {
			return false
		}

		// 向左边渲染
		if left < 0 {
			this.MClip.X = this.MClip.X - left // 裁剪窗口右移 （左边部分消失）
			left = 0                           // x坐标变为0 ( < 0部分不可见 )
		}

		if right >= r.GetLocalFrame().W {
			right = r.GetLocalFrame().W // 保持<=父组件范围
		}

		this.MClip.W = right - left // 计算宽
	}

	// 是子组件
	if r.GetLocalFrame().H > 0 {

		if up > r.GetLocalFrame().H {
			return false
		}

		if down < 0 {
			return false
		}

		if up < 0 {
			this.MClip.Y = this.MClip.Y - up
			up = 0
		}

		if down >= r.GetLocalFrame().H {
			down = r.GetLocalFrame().H
		}

		this.MClip.H = down - up
	}

	this.MDest.X = left + r.GetLocalFrame().X // 目标位置更新
	this.MDest.Y = up + r.GetLocalFrame().Y

	return true
}

// 之前是否已经重置过渲染系统，之后并重置状态
func (this *RenderDevice) ReloadGraphics() bool {
	if this.IsReloadGraphics {
		this.IsReloadGraphics = false
		return true
	}

	return false
}

func (this *RenderDevice) FreeImage(filename string) {
	this.CacheRemove(filename)
}

// 创建窗口后调整分辨率和视口
func (this *RenderDevice) WindowResizeInternal(impl common.RenderDevice, settings common.Settings, eset common.EngineSettings) error {
	s := settings

	// 配置的视口 (来自mod的engine配置或者来自之前的调用)
	oldViewW := s.GetViewW()
	oldViewH := s.GetViewH()

	// 配置里的分辨率
	oldScreenW := s.Get("resolution_w").(int)
	oldScreenH := s.Get("resolution_h").(int)
	w, h := impl.GetWindowSize()

	s.Set("resolution_w", w)
	s.Set("resolution_h", h)

	tmpScreenH := 0

	if s.Get("dpi_scaling").(bool) && this.ddpi > 0 && eset.Get("resolutions", "virtual_dpi").(float32) > 0 {
		// 配置了启用dpi缩放
		tmpScreenH = (int)((float32)(s.Get("resolution_h").(int)) * (eset.Get("resolutions", "virtual_dpi").(float32) / this.ddpi))
	} else {
		tmpScreenH = s.Get("resolution_h").(int)
	}

	s.SetViewH(tmpScreenH) // 配置视口高 默认等于分辨率

	maxRenderSize := 0
	tmp := eset.Get("resolutions", "virtual_height").([]int)
	if s.Get("max_render_size").(int) == 0 {
		// 遵从mod的配置的最大范围
		if len(tmp) > 0 {
			maxRenderSize = tmp[len(tmp)-1]
		}
	} else {
		// 遵从主配置
		maxRenderSize = s.Get("max_render_size").(int)
	}

	// 处理窗口高度(tmpScreenH)不在要求的范围内的情况
	// scale virtual height when outside of VIRTUAL_HEIGHTS range
	if len(tmp) != 0 {
		if tmpScreenH < tmp[0] {
			s.SetViewH(tmp[0])
		} else if tmpScreenH >= maxRenderSize {
			s.SetViewH(maxRenderSize)
		}
	}

	s.SetViewHHalf(s.GetViewH() / 2)
	s.SetViewScaling((float32)(s.GetViewH()) / (float32)(s.Get("resolution_h").(int)))
	s.SetViewW((int)((float32)(s.Get("resolution_w").(int)) * s.GetViewScaling()))

	// 处理比mod的要求的最小宽要小的情况
	// letterbox if too tall
	minScreenW := eset.Get("resolutions", "required_width").(int)
	if s.GetViewW() < minScreenW {
		s.SetViewW(minScreenW)
		s.SetViewScaling((float32)(s.GetViewW()) / (float32)(s.Get("resolution_w").(int)))
	}

	s.SetViewWHalf(s.GetViewW() / 2)

	if s.GetViewW() != oldViewW || s.GetViewH() != oldViewH {
		logfile.LogInfo("RenderDevice: Internal render size is %dx%d", s.GetViewW(), s.GetViewH())
	}

	if s.Get("resolution_w").(int) != oldScreenW || s.Get("resolution_h").(int) != oldScreenH {
		logfile.LogInfo("RenderDevice: Window size changed to %dx%d", s.Get("resolution_w").(int), s.Get("resolution_h").(int))
	}

	return nil
}

func (this *RenderDevice) CreateRenderDeviceList(msg common.MessageEngine) ([]string, []string) {
	s1 := []string{
		"sdl",
		"sdl_hardware",
	}

	s2 := []string{
		msg.Get("SDL software renderer\n\nOften slower, but less likely to have issues."),
		msg.Get("SDL hardware renderer\n\nThe default renderer that is often faster than the SDL software renderer."),
	}

	return s1, s2
}
