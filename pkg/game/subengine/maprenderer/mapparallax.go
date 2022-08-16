package maprenderer

import (
	"monster/pkg/common"
	"monster/pkg/common/fpoint"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/utils/parsing"
)

// 视差图层
type MapParallaxLayer struct {
	sprite      common.Sprite
	speed       float32
	fixedSpeed  fpoint.FPoint
	fixedOffset fpoint.FPoint
	mapLayer    string
}

func constructMapParallaxLayer() MapParallaxLayer {
	return MapParallaxLayer{
		fixedSpeed:  fpoint.Construct(),
		fixedOffset: fpoint.Construct(),
	}
}

type MapParallax struct {
	layers          []MapParallaxLayer
	mapCenter       fpoint.FPoint
	currentLayer    int
	loaded          bool
	currentFilename string
}

func newMapParallax() *MapParallax {
	mp := &MapParallax{}
	mp.init()
	return mp
}

func (this *MapParallax) init() {
	this.mapCenter = fpoint.Construct()
}

func (this *MapParallax) Close() {
	this.clear()
}

func (this *MapParallax) clear() {
	for i, _ := range this.layers {
		if this.layers[i].sprite != nil {
			this.layers[i].sprite.Close()
		}
	}

	this.layers = nil
	this.loaded = false
}

func (this *MapParallax) Load(modules common.Modules, filename string) error {
	mods := modules.Mods()
	render := modules.Render()
	settings := modules.Settings()

	maxFPS := settings.Get("max_fps").(int)
	this.currentFilename = filename

	if this.loaded {
		this.clear()
	}

	infile := fileparser.New()

	err := infile.Open(filename, true, mods)
	if err != nil {
		return err
	}
	defer infile.Close()

	var first string
	for infile.Next(mods) {
		key := infile.Key()
		val := infile.Val()

		if infile.IsNewSection() && infile.GetSection() == "layer" {
			this.layers = append(this.layers, constructMapParallaxLayer())
		}

		if len(this.layers) == 0 {
			continue
		}

		switch key {
		case "image":
			graphics, err := render.LoadImage(settings, mods, val)
			if err != nil {
				return err
			}

			this.layers[len(this.layers)-1].sprite, err = graphics.CreateSprite()
			if err != nil {
				graphics.UnRef()
				return err
			}
			graphics.UnRef()
		case "speed":
			this.layers[len(this.layers)-1].speed = (settings.LOGIC_FPS() * parsing.ToFloat(val, 0) / (float32)(maxFPS))
		case "fixed_speed":
			first, val = parsing.PopFirstString(val, "")
			this.layers[len(this.layers)-1].fixedSpeed.X = (settings.LOGIC_FPS() * parsing.ToFloat(first, 0) / (float32)(maxFPS))
			first, val = parsing.PopFirstString(val, "")
			this.layers[len(this.layers)-1].fixedSpeed.Y = (settings.LOGIC_FPS() * parsing.ToFloat(first, 0) / (float32)(maxFPS))

		case "map_layer":
			this.layers[len(this.layers)-1].mapLayer = val
		}
	}

	this.loaded = true

	return nil
}

func (this *MapParallax) SetMapCenter(x, y int) {
	this.mapCenter.X = float32(x) + 0.5
	this.mapCenter.Y = float32(y) + 0.5
}

func (this *MapParallax) Render(cam fpoint.FPoint, mapLayer string) error {
	// TODO
	return nil
}
