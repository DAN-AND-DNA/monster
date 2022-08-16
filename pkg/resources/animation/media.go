package animation

import "monster/pkg/common"

// 图片容器
type Media struct {
	images   map[string]common.Image
	firstKey string
}

func NewMedia() *Media {
	m := &Media{}

	m.Init()

	return m
}

func (this *Media) Init() {
	this.images = map[string]common.Image{}
}

func (this *Media) Close() {
	// do nothing
}

// 记载图片，标记为key，key为空字符串代表默认
func (this *Media) LoadImage(settings common.Settings, mods common.ModManager, render common.RenderDevice, path, key string) error {
	loadedImg, err := render.LoadImage(settings, mods, path)
	if err != nil {
		return err
	}

	if ptr, ok := this.images[key]; ok {
		ptr.UnRef()
	}

	this.images[key] = loadedImg

	if len(this.images) == 1 {
		this.firstKey = key
	}

	return nil
}

func (this *Media) GetImageFromKey(key string) (common.Image, bool) {
	ptr, ok := this.images[key]
	if ok {
		return ptr, true
	}

	if len(this.images) != 0 {
		return this.images[this.firstKey], true
	}

	return nil, false
}

func (this *Media) UnRef() {
	for _, ptr := range this.images {
		ptr.UnRef()
	}

	this.images = map[string]common.Image{}
}
