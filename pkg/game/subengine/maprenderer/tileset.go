package maprenderer

import (
	"fmt"
	"math"
	"monster/pkg/common"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/utils/parsing"
)

type TileDef struct {
	offset point.Point
	tile   common.Sprite
}

func constructTileDef() TileDef {
	td := TileDef{}

	return td
}

type TileAnim struct {
	frames        uint16 // 帧数
	currentFrame  uint16
	duration      uint16        // 当前帧在 本次循环里已经展现的时间
	pos           []point.Point // 每帧的位置
	frameDuration []uint16      // 每帧每次循环里需要展现时间
}

func constructTileAnim() TileAnim {
	ta := TileAnim{}

	return ta
}

type TileSet struct {
	currentFilename string
	sprites         []common.Sprite
	anim            []TileAnim
	tiles           []TileDef // 瓷砖的定义
	maxSizeX        int       // 比例：多少个eset定义大小的瓷砖
	maxSizeY        int
}

func newTileSet() *TileSet {
	ts := &TileSet{}
	ts.init()

	return ts
}

func (this *TileSet) init() {
	this.Reset()
}

func (this *TileSet) Reset() {
	for _, ptr := range this.sprites {
		if ptr != nil {
			ptr.Close()
		}
	}

	this.sprites = nil

	for i, _ := range this.tiles {
		if this.tiles[i].tile != nil {
			this.tiles[i].tile.Close()
		}
	}

	this.tiles = nil
	this.anim = nil

	this.maxSizeX = 0
	this.maxSizeY = 0

}

func (this *TileSet) loadGraphics(modules common.Modules, filenames []string) error {
	settings := modules.Settings()
	mods := modules.Mods()
	render := modules.Render()

	for i, filename := range filenames {
		if this.sprites[i] != nil {
			this.sprites[i].Close()
			this.sprites[i] = nil
		}

		if filename == "" {
			continue
		}

		graphics, err := render.LoadImage(settings, mods, filename)
		if err != nil {
			return err
		}

		this.sprites[i], err = graphics.CreateSprite()
		if err != nil {
			graphics.UnRef()
			return err
		}

		graphics.UnRef()

	}

	return nil
}

func (this *TileSet) Load(modules common.Modules, filename string) error {
	mods := modules.Mods()
	eset := modules.Eset()

	if this.currentFilename == filename {
		return nil
	}

	this.Reset()
	infile := fileparser.New()

	err := infile.Open(filename, true, mods)
	if err != nil {
		return err
	}
	defer infile.Close()

	var imageFilenames []string // 瓷砖图片文件
	var tileImages []int        // 瓷砖定义对应的图片文件
	var tileClips []rect.Rect
	var tileOffsets []point.Point
	var index int
	for infile.Next(mods) {
		key := infile.Key()
		val := infile.Val()

		if (infile.IsNewSection() && infile.GetSection() == "tileset") ||
			(len(this.sprites) == 0 && infile.GetSection() == "") {
			imageFilenames = append(imageFilenames, "")
			this.sprites = append(this.sprites, nil)
		}

		switch key {
		case "img":
			// 对应的图片
			imageFilenames[len(imageFilenames)-1] = val
		case "tile":
			// 瓷砖的定义
			index, val = parsing.PopFirstInt(val, "")

			if index >= len(this.tiles) {
				add := index + 1 - len(this.tiles)

				for i := 0; i < add; i++ {
					this.tiles = append(this.tiles, constructTileDef())
					tileImages = append(tileImages, 0)
					tileClips = append(tileClips, rect.Construct())
					tileOffsets = append(tileOffsets, point.Construct())
				}
			}

			clip := rect.Construct()
			clip.X, val = parsing.PopFirstInt(val, "")
			clip.Y, val = parsing.PopFirstInt(val, "")
			clip.W, val = parsing.PopFirstInt(val, "")
			clip.H, val = parsing.PopFirstInt(val, "")

			offset := point.Construct()
			offset.X, val = parsing.PopFirstInt(val, "")
			offset.Y, val = parsing.PopFirstInt(val, "")

			tileImages[index] = len(imageFilenames) - 1
			tileClips[index] = clip
			tileOffsets[index] = offset
		case "animation":
			var frame uint16
			index, val = parsing.PopFirstInt(val, "")

			if index >= len(this.anim) {
				add := index + 1 - len(this.anim)
				for i := 0; i < add; i++ {
					this.anim = append(this.anim, constructTileAnim())
				}
			}
			var repeatVal, strDuration string
			repeatVal, val = parsing.PopFirstString(val, "")
			for repeatVal != "" {
				this.anim[index].frames++
				this.anim[index].pos = append(this.anim[index].pos, point.Construct())
				this.anim[index].frameDuration = append(this.anim[index].frameDuration, 0)
				this.anim[index].pos[frame].X = parsing.ToInt(repeatVal, 0)
				this.anim[index].pos[frame].Y, val = parsing.PopFirstInt(val, "")
				strDuration, val = parsing.PopFirstString(val, "")
				this.anim[index].frameDuration[frame] = (uint16)(parsing.ToDuration(strDuration, 0))
				frame++
				repeatVal, val = parsing.PopFirstString(val, "")
			}
		default:
			return fmt.Errorf("TileSet: '%s' is not a valid key.\n", key)
		}
	}

	err = this.loadGraphics(modules, imageFilenames)
	if err != nil {
		return err
	}

	tileW := eset.Get("tileset", "tile_size").([]int)[0]
	tileH := eset.Get("tileset", "tile_size").([]int)[1]

	// 每个瓷砖的定义
	for i, _ := range this.tiles {
		// 瓷砖定义对应的精灵
		ptr := this.sprites[tileImages[i]]
		if ptr == nil {
			continue
		}

		// 创建独立的一份
		this.tiles[i].tile, err = ptr.GetGraphics().CreateSprite()
		if err != nil {
			return err
		}

		this.tiles[i].tile.SetClipFromRect(tileClips[i])
		this.tiles[i].offset = tileOffsets[i]

		this.maxSizeX = (int)(math.Max(float64(this.maxSizeX), float64(this.tiles[i].tile.GetClip().W/tileW)+1))
		this.maxSizeY = (int)(math.Max(float64(this.maxSizeY), float64(this.tiles[i].tile.GetClip().H/tileH)+1))
	}

	this.currentFilename = filename

	return nil
}

func (this *TileSet) Close() {
	this.Reset()
}

func (this *TileSet) Logic() {
	for i, _ := range this.anim {
		an := &(this.anim[i])

		if an.frames == 0 {
			// 静态
			continue
		}

		if this.tiles[i].tile != nil && an.duration >= an.frameDuration[an.currentFrame] {
			// play下一帧
			clip := this.tiles[i].tile.GetClip()
			clip.X = an.pos[an.currentFrame].X
			clip.Y = an.pos[an.currentFrame].Y
			this.tiles[i].tile.SetClipFromRect(clip)
			an.duration = 0
			an.currentFrame = (an.currentFrame + 1) % an.frames // 反复
		}

		an.duration++
	}
}
