package iconmanager

import (
	"monster/pkg/common"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/utils/parsing"
)

type IconSet struct {
	gfx     common.Sprite
	idBegin int // 起始id
	idEnd   int
	columns int // 列
}

func constructIconSet() IconSet {
	return IconSet{
		columns: 1,
	}
}

type IconManager struct {
	textOffset  point.Point
	iconSets    []IconSet
	currentSet  *IconSet
	currentSrc  rect.Rect
	currentDest rect.Rect
}

func New(settings common.Settings, eset common.EngineSettings, render common.RenderDevice, mods common.ModManager) *IconManager {
	im := &IconManager{}
	im.init(settings, eset, render, mods)

	return im
}

func (this *IconManager) init(settings common.Settings, eset common.EngineSettings, render common.RenderDevice, mods common.ModManager) common.IconManager {
	this.textOffset = point.Construct()
	this.currentSrc = rect.Construct()
	this.currentDest = rect.Construct()

	infile := fileparser.New()

	err := infile.Open("engine/icons.txt", true, mods)
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	for infile.Next(mods) {
		key := infile.Key()
		val := infile.Val()

		switch key {
		case "icon_set":
			var firstId int
			var filename string

			firstId, val = parsing.PopFirstInt(val, "")
			filename, val = parsing.PopFirstString(val, "")

			newIset, ok, err := this.loadIconSet(settings, eset, render, mods, filename, firstId)
			if err != nil {
				panic(err)
			}

			if ok {
				this.iconSets = append(this.iconSets, newIset)
			}

		case "text_offset":
			this.textOffset = parsing.ToPoint(val)
		}

	}

	if len(this.iconSets) == 0 {
		newIset, ok, err := this.loadIconSet(settings, eset, render, mods, "images/icons/icons.png", 0)
		if err != nil {
			panic(err)
		}

		if ok {
			this.iconSets = append(this.iconSets, newIset)
		}
	}

	return this
}

func (this *IconManager) Close() {
	for i := 0; i < len(this.iconSets); i++ {
		this.iconSets[i].gfx.Close()
	}

	this.iconSets = nil
}

func (this *IconManager) loadIconSet(settings common.Settings, eset common.EngineSettings, render common.RenderDevice, mods common.ModManager, filename string, firstId int) (IconSet, bool, error) {
	iset := constructIconSet()

	iconSize := eset.Get("resolutions", "icon_size").(int)

	if render == nil || iconSize == 0 {
		return iset, false, nil
	}

	graphics, err := render.LoadImage(settings, mods, filename)
	if err != nil {
		return iset, false, err
	}
	defer graphics.UnRef()

	iset.gfx, err = graphics.CreateSprite()
	if err != nil {
		return iset, false, err
	}
	defer iset.gfx.Close()

	gh, err := iset.gfx.GetGraphicsHeight()
	if err != nil {
		return iset, false, err
	}

	gw, err := iset.gfx.GetGraphicsWidth()
	if err != nil {
		return iset, false, err
	}

	rows := gh / iconSize        // 行
	iset.columns = gw / iconSize // 列

	if iset.columns == 0 {
		iset.columns = 1
	}

	iset.idBegin = firstId
	iset.idEnd = firstId + iset.columns*rows - 1
	iset.gfx.KeepAlive()
	return iset, true, nil
}

func (this *IconManager) SetIcon(eset common.EngineSettings, iconId int, destPos point.Point) {
	if len(this.iconSets) == 0 {
		this.currentSet = nil
		return
	}

	for i := len(this.iconSets); i > 0; i-- {
		if iconId >= this.iconSets[i-1].idBegin && iconId <= this.iconSets[i-1].idEnd {
			this.currentSet = &(this.iconSets[i-1])
			break
		} else if i-1 == 0 {
			this.currentSet = nil
			return
		}
	}

	iconSize := eset.Get("resolutions", "icon_size").(int)

	offsetId := iconId - this.currentSet.idBegin
	this.currentSrc.X = offsetId % this.currentSet.columns * iconSize
	this.currentSrc.Y = offsetId % this.currentSet.columns * iconSize
	this.currentSrc.W = iconSize
	this.currentSrc.H = iconSize
	this.currentSet.gfx.SetClipFromRect(this.currentSrc)

	this.currentDest.X = destPos.X
	this.currentDest.Y = destPos.Y
	this.currentSet.gfx.SetDestFromRect(this.currentDest)
}

func (this *IconManager) RenderToImage(render common.RenderDevice, img common.Image) error {
	if this.currentSet == nil {
		return nil
	}

	if img != nil {
		_, err := render.RenderToImage(this.currentSet.gfx.GetGraphics(), this.currentSrc, img, this.currentDest)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *IconManager) Render(render common.RenderDevice) error {
	if this.currentSet == nil {
		return nil
	}

	err := render.Render(this.currentSet.gfx)
	if err != nil {
		return err
	}

	return nil
}

func (this *IconManager) GetTextOffset() point.Point {
	return this.textOffset
}
