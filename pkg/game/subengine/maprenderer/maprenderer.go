package maprenderer

import (
	"fmt"
	"math"
	"monster/pkg/common"
	"monster/pkg/common/define/enginesettings"
	"monster/pkg/common/event"
	"monster/pkg/common/fpoint"
	"monster/pkg/common/gameres"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/common/tooltipdata"
	"monster/pkg/game/subengine/maprenderer/base"
	"monster/pkg/utils"
	"sort"
)

type MapRenderer struct {
	base.Map
	tip                common.WidgetTooltip
	tipBuf             tooltipdata.TooltipData
	tipPos             point.Point
	tset               *TileSet
	mapParallax        *MapParallax
	entityHiddenNormal common.Sprite
	entityHiddenEnemy  common.Sprite
	cam                *Camera
	mapChange          bool
	collider           gameres.MapCollision
	loot               []event.Component

	teleportation       bool // 传送
	teleportDestination fpoint.FPoint
	teleportMapName     string
	indexObjectLayer    uint // 层级关系：背景、对象、碰撞，碰撞被删除，故先背景后对象
	isSpawnMap          bool // 初始地图为maps/spawn.txt，里面包含要跳转的实际地图，所以不需要一开始就渲染
}

func New(modules common.Modules, resf common.Factory) *MapRenderer {
	mapr := &MapRenderer{}
	mapr.init(modules, resf)

	return mapr
}

func (this *MapRenderer) init(modules common.Modules, resf common.Factory) gameres.MapRenderer {
	widgetf := modules.Widgetf()
	render := modules.Render()
	settings := modules.Settings()
	mods := modules.Mods()

	// base
	this.Map = base.ConstructMap()

	// self
	this.tip = widgetf.New("tooltip").(common.WidgetTooltip).Init(modules)
	this.tipBuf = tooltipdata.Construct()
	this.tipPos = point.Construct()
	this.tset = newTileSet()
	this.mapParallax = newMapParallax()
	this.cam = newCamera(modules)
	this.collider = resf.New("mapcollision").(gameres.MapCollision).Init()

	gfx, err := render.LoadImage(settings, mods, "images/menus/entity_hidden.png")
	if err != nil {
		panic(err)
	}
	defer gfx.UnRef()

	gfxW, err := gfx.GetWidth()
	if err != nil {
		panic(err)
	}
	gfxH, err := gfx.GetHeight()
	if err != nil {
		panic(err)
	}

	this.entityHiddenNormal, err = gfx.CreateSprite()
	if err != nil {
		panic(err)
	}
	this.entityHiddenNormal.SetClip(0, 0, gfxW, gfxH/2)

	this.entityHiddenEnemy, err = gfx.CreateSprite()
	if err != nil {
		panic(err)
	}
	this.entityHiddenEnemy.SetClip(0, gfxH/2, gfxW, gfxH/2)

	this.teleportDestination = fpoint.Construct()

	return this
}

func (this *MapRenderer) Clear() {
	if this.tip != nil {
		this.tip.Close()
		this.tip = nil
	}

	if this.tset != nil {
		this.tset.Close()
		this.tset = nil
	}

	if this.mapParallax != nil {
		this.mapParallax.Close()
		this.mapParallax = nil
	}
}

func (this *MapRenderer) Close() {
	this.Map.Close(this)
}

func (this *MapRenderer) Render(modules common.Modules, r []common.Renderable, rDead []common.Renderable) error {
	eset := modules.Eset()

	err := this.mapParallax.Render(this.cam.shake, "")
	if err != nil {
		return err
	}

	if eset.Get("tileset", "orientation").(int) == enginesettings.TILESET_ORTHOGONAL {
		// TODO
	} else {
		r = this.calculatePriosIso(r)
		rDead = this.calculatePriosIso(rDead)
		sort.Slice(r, func(i, j int) bool { return r[i].GetPrio() < r[j].GetPrio() })
		sort.Slice(rDead, func(i, j int) bool { return r[i].GetPrio() < r[j].GetPrio() })
		err := this.renderIso(modules, r, rDead)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *MapRenderer) calculatePriosIso(r []common.Renderable) []common.Renderable {
	for _, ptr := range r {
		tileX := math.Floor((float64)(ptr.GetMapPos().X))
		tileY := math.Floor((float64)(ptr.GetMapPos().Y))

		commax := (int)((ptr.GetMapPos().X - (float32)(tileX)) * (1 << 10))
		commay := (int)((ptr.GetMapPos().Y - (float32)(tileY)) * (1 << 10))
		oldPrio := ptr.GetPrio()
		oldPrio += ((uint64)(tileX+tileY) << 37) + ((uint64)(tileX) << 20) + ((uint64)(commax+commay) << 8)
		ptr.SetPrio(oldPrio)
	}

	return r
}

func (this *MapRenderer) renderIso(modules common.Modules, r []common.Renderable, rDead []common.Renderable) error {
	index := 0

	// 背景层
	for index < (int)(this.indexObjectLayer) {
		err := this.renderIsoLayer(modules, this.GetLayer(index), this.tset)
		if err != nil {
			return err
		}
		err = this.mapParallax.Render(this.cam.shake, this.GetLayerName(index))
		if err != nil {
			return err
		}
		index++
	}

	err := this.renderIsoBackObjects(modules, rDead)
	if err != nil {
		return err
	}

	err = this.renderIsoFrontObjects(modules, r)
	if err != nil {
		return err
	}

	// 对象层
	index++

	// 战争迷雾
	layers := this.Map.GetLayers()
	for index < len(layers) {

		if this.GetLayerName(index) != "fow_dark" && this.GetLayerName(index) != "fow_fog" {
			err := this.renderIsoLayer(modules, this.GetLayer(index), this.tset)
			if err != nil {
				return err
			}
		}

		err := this.mapParallax.Render(this.cam.shake, this.GetLayerName(index))
		if err != nil {
			return err
		}

		index++
	}

	return nil

}

func (this *MapRenderer) centerTile(eset common.EngineSettings, p point.Point) point.Point {
	r := p

	if eset.Get("tileset", "orientation").(int) == enginesettings.TILESET_ORTHOGONAL {
		r.X += eset.Get("tileset", "tile_w_half").(int)
		r.Y += eset.Get("tileset", "tile_h_half").(int)
	} else {
		r.Y += eset.Get("tileset", "tile_h_half").(int)
	}

	return r

}

func (this *MapRenderer) renderIsoLayer(modules common.Modules, layerData [][]uint16, tileSet *TileSet) error {
	eset := modules.Eset()
	settings := modules.Settings()
	render := modules.Render()

	// 屏幕左上角(0, 0)在世界地图的位置
	upperLeft := utils.ScreenToMap(settings, eset, 0, 0, this.cam.shake.X, this.cam.shake.Y)
	tileSize := eset.Get("tileset", "tile_size").([]int)

	// 屏幕可以塞多少个瓷砖   (等距，瓷砖的高是宽的一半)
	maxTilesWidth := settings.GetViewW()/tileSize[0] + 2*this.tset.maxSizeX
	maxTilesHeight := 2*settings.GetViewH()/tileSize[1] + 2*(this.tset.maxSizeY+1)

	i := (int16)(upperLeft.X - (float32)(this.tset.maxSizeY)/2 - (float32)(this.tset.maxSizeX))
	j := (int16)(upperLeft.Y - (float32)(this.tset.maxSizeY)/2 + (float32)(this.tset.maxSizeX))

	dest := point.Construct()

	for y := maxTilesHeight; y > 0; y-- {
		tilesWidth := (int16)(0)

		if i < -1 {
			j = j + i + 1
			tilesWidth = tilesWidth - (i + 1)
			i = -1
		}

		d := j - (int16)(this.Map.GetH())
		if d >= 0 {
			j = j - d
			tilesWidth = tilesWidth + d
			i = i + d
		}

		jEnd := (int16)(math.Max((float64)(j+i-(int16)(this.GetW())+1), math.Max((float64)(j-(int16)(maxTilesWidth)), 0)))

		p := utils.MapToScreen(settings, eset, (float32)(i), (float32)(j), this.cam.shake.X, this.cam.shake.Y)
		p = this.centerTile(eset, p) // 瓷砖中心位置的地图坐标

		// 水平画
		for j > jEnd {
			j--
			i++
			tilesWidth++
			p.X += tileSize[0]

			currentTile := layerData[i][j]

			tile := tileSet.tiles[currentTile]

			dest.X = p.X - tile.offset.X
			dest.Y = p.Y - tile.offset.Y

			tile.tile.SetDestFromPoint(dest)
			err := render.Render(tile.tile)
			if err != nil {
				return err
			}
		}

		j += tilesWidth
		i -= tilesWidth

		if y%2 > 0 {
			i++
		} else {
			j++
		}
	}

	return nil
}

func (this *MapRenderer) SetTeleportation(val bool) {
	this.teleportation = val
}

func (this *MapRenderer) SetTeleportMapName(val string) {
	this.teleportMapName = val
}

func (this *MapRenderer) GetTeleportation() bool {
	return this.teleportation
}

func (this *MapRenderer) GetTeleportMapName() string {
	return this.teleportMapName
}

func (this *MapRenderer) SetTeleportDestination(val fpoint.FPoint) {
	this.teleportDestination = val
}

func (this *MapRenderer) GetTeleportDestination() fpoint.FPoint {
	return this.teleportDestination
}

func (this *MapRenderer) Load(modules common.Modules, loot gameres.LootManager, camp gameres.CampaignManager, event gameres.EventManager, gresf gameres.Factory, fname string) error {

	render := modules.Render()

	// TODO
	// reset all

	if fname == "maps/spawn.txt" {
		this.isSpawnMap = true
	} else {
		this.isSpawnMap = false
	}

	// 加载图层定义，怪物定义和地图触发事件等
	err := this.Map.Load(modules, loot, camp, event, gresf, fname)
	if err != nil {
		panic(err)
		return err
	}

	// TODO
	// load music

	layers := this.Map.GetLayers()

	// 拷贝并移除碰撞图层
	for i, layer := range layers {
		if this.Map.GetLayerName(i) == "collision" {
			width := (uint16)(len(layer))
			if width == 0 {
				return fmt.Errorf("MapRenderer: Map width is 0. Can't set collision layer.")
			}
			height := (uint16)(len(layer[0]))
			this.collider.SetMap(layer, width, height) // 拷贝
			this.Map.RemoveLayer(i)                    // 移除
		}
	}

	layers = this.Map.GetLayers()
	for i, _ := range layers {
		if this.Map.GetLayerName(i) == "object" {
			this.indexObjectLayer = (uint)(i)
		}
	}

	// TODO enemy group
	// 加载瓷砖
	this.tset.Load(modules, this.Map.GetTileSet())

	// TODO
	// fog of war
	// parallax map

	render.SetBackgroundColor(this.Map.GetBackgroundColor())

	return nil
}

func (this *MapRenderer) Logic(modules common.Modules) {
	this.tset.Logic()
	this.cam.Logic(modules)
}

func (this *MapRenderer) ExecuteOnLoadEvent(modules common.Modules, eventManager gameres.EventManager, camp gameres.CampaignManager) {
	events := this.Map.GetEvents()

	if len(events) == 0 {
		return
	}

	del := map[int]struct{}{}

	for i := len(events); i > 0; {
		i--

		if events[i].ActivateType == event.ACTIVATE_ON_LOAD {
			if eventManager.ExecuteEvent(modules, this, camp, &(events[i])) {
				del[i] = struct{}{}
			}
		}
	}

	for i := len(events); i > 0; {
		i--

		if events[i].ActivateType == event.ACTIVATE_STATIC {
			if eventManager.ExecuteEvent(modules, this, camp, &(events[i])) {
				del[i] = struct{}{}
			}

		}
	}

	var tmp []event.Event

	for id, val := range events {
		if _, ok := del[id]; ok {
			continue
		}

		tmp = append(tmp, val)
	}

	this.Map.SetEvents(tmp)
}

func (this *MapRenderer) renderIsoBackObjects(modules common.Modules, r []common.Renderable) error {
	//TODO
	return nil
}

// 地图上的瓷砖位置(x,y) 转成 屏幕上坐标和大小
func (this *MapRenderer) getTileBounds(modules common.Modules, x, y int16, layerData [][]uint16) (rect.Rect, point.Point) {
	settings := modules.Settings()
	eset := modules.Eset()

	bounds := rect.Construct()
	center := point.Construct()

	if x >= 0 && x < (int16)(this.GetW()) && y >= 0 && y < (int16)(this.GetH()) {
		tileIndex := layerData[x][y]       // 指定的瓷砖序号
		tile := this.tset.tiles[tileIndex] // 拿瓷砖信息
		if tile.tile == nil {
			return bounds, center
		}

		center = this.centerTile(eset, utils.MapToScreen(settings, eset, float32(x), float32(y), this.cam.shake.X, this.cam.shake.Y))
		bounds.X = center.X - tile.offset.X
		bounds.Y = center.Y - tile.offset.Y
		bounds.W = tile.tile.GetClip().W
		bounds.H = tile.tile.GetClip().H
	}

	return bounds, center
}

func (this *MapRenderer) drawRenderable(modules common.Modules, r []common.Renderable, index int) error {
	settings := modules.Settings()
	eset := modules.Eset()
	render := modules.Render()

	rCursor := r[index]

	if rCursor.GetImage() != nil {
		p := utils.MapToScreen(settings, eset, rCursor.GetMapPos().X, rCursor.GetMapPos().Y, this.cam.shake.X, this.cam.shake.Y)
		dest := rect.Construct()
		dest.X = p.X - rCursor.GetOffset().X
		dest.Y = p.Y - rCursor.GetOffset().Y
		err := render.Render1(rCursor, dest)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *MapRenderer) renderIsoFrontObjects(modules common.Modules, r []common.Renderable) error {
	settings := modules.Settings()
	eset := modules.Eset()
	render := modules.Render()

	// 屏幕左上
	upperLeft := utils.ScreenToMap(settings, eset, 0, 0, this.cam.shake.X, this.cam.shake.Y)

	tileSize := eset.Get("tileset", "tile_size").([]int)

	// 瓷砖的最高和最宽坐标(等距)
	maxTilesWidth := settings.GetViewW()/tileSize[0] + 2*this.tset.maxSizeX
	maxTilesHeight := 2 * (settings.GetViewH()/tileSize[1] + 2*this.tset.maxSizeY)

	i := (int16)(upperLeft.X - (float32)(this.tset.maxSizeY) - (float32)(this.tset.maxSizeX))
	j := (int16)(upperLeft.Y - (float32)(this.tset.maxSizeY) + (float32)(this.tset.maxSizeX))

	dest := point.Construct()
	rCursor := 0
	rEnd := len(r)
	for _, ptr := range r {
		mp := ptr.GetMapPos()
		if int16(mp.X)+int16(mp.Y) < i+j || int16(mp.X) < i {
			rCursor++
			continue
		} else {
			break
		}
	}

	layers := this.Map.GetLayers()
	if this.indexObjectLayer >= (uint)(len(layers)) {
		return nil
	}

	var renderBehindSW []int
	var renderBehindNE []int
	var renderBehindNone []int

	drawnTiles := make([][]uint16, this.GetW())
	for index, _ := range drawnTiles {
		drawnTiles[index] = make([]uint16, this.GetH())
	}

	for y := maxTilesHeight; y > 0; y-- {
		tilesWidth := int16(0)

		if i < -1 {
			j = j + i + 1
			tilesWidth = tilesWidth - (i + 1)
			i = -1
		}

		d := j - (int16)(this.GetH())
		if d >= 0 {
			j = j - d
			tilesWidth = tilesWidth + d
			i = i + d
		}

		jEnd := (int16)(math.Max((float64)(j+i-(int16)(this.GetW())+1), math.Max((float64)(j-(int16)(maxTilesWidth)), 0)))
		p := utils.MapToScreen(settings, eset, (float32)(i), (float32)(j), this.cam.shake.X, this.cam.shake.Y)
		p = this.centerTile(eset, p)

		currentLayer := layers[this.indexObjectLayer]
		isLastNETile := false

		for j > jEnd {
			j--
			i++
			tilesWidth++
			p.X += tileSize[0]
			drawTile := true

			rPreCursor := rCursor
			for rPreCursor < rEnd {
				tmpMp := r[rPreCursor].GetMapPos()
				rCursorX := (int16)(tmpMp.X)
				rCursorY := (int16)(tmpMp.Y)

				if rCursorX-1 == i && rCursorY+1 == j || rCursorX+1 == i && rCursorY-1 == j {
					drawTile = false
					break
				} else if rCursorX+1 > i || rCursorY+1 > j {
					break
				}

				rPreCursor++
			}

			if drawTile && drawnTiles[i][j] == 0 {
				currentTile := currentLayer[i][j]
				tile := this.tset.tiles[currentTile]
				dest.X = p.X - tile.offset.X
				dest.Y = p.Y - tile.offset.Y

				tile.tile.SetDestFromPoint(dest)

				//TODO
				// 战争迷雾

				this.checkHiddenEntities(i, j, currentLayer, r)

				err := render.Render(tile.tile)
				if err != nil {
					return err
				}
				drawnTiles[i][j] = 1
			}

			if rCursor == rEnd {
				continue
			}

		do_last_NE_tile:
			// 获得地图x，y序号在当前屏幕上的坐标
			tileSWBounds, tileSWCenter := this.getTileBounds(modules, i-2, j+2, currentLayer)
			tileSBounds, tileSCenter := this.getTileBounds(modules, i-1, j+2, currentLayer)
			_ = tileSCenter
			tileNEBounds, tileNECenter := this.getTileBounds(modules, i, j, currentLayer)
			tileEBounds, tileECenter := this.getTileBounds(modules, i, j+1, currentLayer)
			_ = tileECenter

			drawSWTile := false
			drawNETile := false

			for rCursor != rEnd {
				rCursorX := (int)(r[rCursor].GetMapPos().X)
				rCursorY := (int)(r[rCursor].GetMapPos().Y)

				if rCursorX+1 == (int)(i) && rCursorY-1 == (int)(j) {
					drawSWTile = true
					drawNETile = !isLastNETile

					rCursorLeft := utils.MapToScreen(settings, eset, r[rCursor].GetMapPos().X, r[rCursor].GetMapPos().Y, this.cam.shake.X, this.cam.shake.Y)
					rCursorLeft.Y -= r[rCursor].GetOffset().Y
					rCursorRight := rCursorLeft
					rCursorLeft.X -= r[rCursor].GetOffset().X
					rCursorRight.X += r[rCursor].GetSrc().W - r[rCursor].GetOffset().X

					isBehindSW := false
					isBehindNE := false

					if utils.IsWithinRect(tileSBounds, rCursorRight) && utils.IsWithinRect(tileSWBounds, rCursorLeft) {
						isBehindSW = true
					}

					if drawNETile && utils.IsWithinRect(tileEBounds, rCursorLeft) && utils.IsWithinRect(tileNEBounds, rCursorRight) {
						isBehindNE = true
					}

					if isBehindSW {
						renderBehindSW = append(renderBehindSW, rCursor)
					} else if isBehindNE {
						renderBehindNE = append(renderBehindNE, rCursor)
					} else {
						renderBehindNone = append(renderBehindNone, rCursor)
					}

					rCursor++
				} else {
					break
				}
			}

			for _, index := range renderBehindSW {
				err := this.drawRenderable(modules, r, index)
				if err != nil {
					return err
				}
			}

			renderBehindSW = nil

			if drawSWTile && i-2 >= 0 && uint16(j)+2 < this.GetH() && drawnTiles[i-2][j+2] == 0 {
				currentTile := currentLayer[i-2][j+2]
				tile := this.tset.tiles[currentTile]

				dest.X = tileSWCenter.X - tile.offset.X
				dest.Y = tileSWCenter.Y - tile.offset.Y
				tile.tile.SetDestFromPoint(dest)
				this.checkHiddenEntities(i, j, currentLayer, r)
				// TODO
				// 战争迷雾
				err := render.Render(tile.tile)
				if err != nil {
					return err
				}

				drawnTiles[i-2][j+2] = 1
			}

			for _, index := range renderBehindNE {
				err := this.drawRenderable(modules, r, index)
				if err != nil {
					return err
				}
			}
			renderBehindNE = nil

			if drawNETile && !drawTile && drawnTiles[i][j] == 0 {
				currentTile := currentLayer[i][j]
				tile := this.tset.tiles[currentTile]

				dest.X = tileNECenter.X - tile.offset.X
				dest.Y = tileNECenter.Y - tile.offset.Y

				tile.tile.SetDestFromPoint(dest)
				this.checkHiddenEntities(i, j, currentLayer, r)
				// TODO
				// 战争迷雾
				err := render.Render(tile.tile)
				if err != nil {
					return err
				}
				drawnTiles[i][j] = 1
			}

			for _, index := range renderBehindNone {
				err := this.drawRenderable(modules, r, index)
				if err != nil {
					return err
				}
			}

			renderBehindNone = nil

			if isLastNETile {
				j++
				i--
				isLastNETile = false
			} else if i == (int16)(this.GetW())-1 || j == 0 {
				j--
				i++
				isLastNETile = true
				goto do_last_NE_tile
			}
		}

		j += tilesWidth
		i -= tilesWidth

		if y%2 != 0 {
			i++
		} else {
			j++
		}

		for {
			if rCursor != rEnd {
				mp := r[rCursor].GetMapPos()

				if (int16)(mp.X)+(int16)(mp.Y) < i+j || (int16)(mp.X) <= i {
					rCursor++
				} else {
					break
				}
			} else {
				break
			}
		}
	}

	return nil
}

func (this *MapRenderer) checkHiddenEntities(x, y int16, layerData [][]uint16, r []common.Renderable) {
	//TODO
}

func (this *MapRenderer) GetIsSpawnMap() bool {
	return this.isSpawnMap
}

func (this *MapRenderer) GetCollider() gameres.MapCollision {
	return this.collider
}

func (this *MapRenderer) GetCam() gameres.MapCamera {
	return this.cam
}
