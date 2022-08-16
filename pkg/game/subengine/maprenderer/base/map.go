package base

import (
	"fmt"
	"math"
	"math/rand"
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define"
	"monster/pkg/common/define/game/stats"
	"monster/pkg/common/event"
	"monster/pkg/common/fpoint"
	"monster/pkg/common/gameres"
	"monster/pkg/common/point"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/utils/parsing"
	"strings"
)

type MapGroup struct {
	type1                  string
	category               string
	pos                    point.Point // 敌人的位置
	area                   point.Point
	levelMin               int
	levelMax               int
	numberMin              int // 怪物数
	numberMax              int
	chance                 float32 // 怪物的出现概率
	direction              int
	wayPoints              []fpoint.FPoint // 怪物的路径（和闲逛半径2选1）
	wanderRadius           int             // 怪物的闲逛半径（和路径2选1）
	requirements           []event.Component
	invincibleRequirements []event.Component
}

func constructMapGroup() MapGroup {
	return MapGroup{
		pos:          point.Construct(),
		area:         point.Construct(1, 1),
		numberMin:    1,
		numberMax:    1,
		chance:       1,
		direction:    -1,
		wanderRadius: 4,
	}
}

type MapNPC struct {
	type1        string
	id           string
	pos          fpoint.FPoint
	requirements []event.Component
}

func constructMapNPC() MapNPC {
	return MapNPC{
		pos: fpoint.Construct(),
	}
}

type MapEnemy struct {
	type1                  string
	pos                    fpoint.FPoint
	direction              int
	wayPoints              []fpoint.FPoint
	wanderRadius           int
	heroAlly               bool
	enemyAlly              bool
	summonPowerIndex       define.PowerId
	summoner               gameres.StatBlock
	requirements           []event.Component
	invincibleRequirements []event.Component
}

func constructMapEnemy(type1 string, pos fpoint.FPoint) MapEnemy {
	return MapEnemy{
		type1:        type1,
		pos:          pos,
		direction:    rand.Intn(100) % 8,
		wanderRadius: 4,
	}
}

type Map struct {
	statBlocks             []gameres.StatBlock // 地图上的东西都是状态块
	filename               string
	tileset                string // 瓷砖定义文件
	musicFilename          string
	layers                 []([][]uint16)
	layerNames             []string
	enemies                []MapEnemy
	enemyGroups            []MapGroup // 敌人
	npcs                   []MapNPC
	events                 []event.Event // 地图事件
	delayedEvents          []event.Event // 推迟执行的地图事件
	intermapRandomFilename string
	intermapRandomQueue    []event.Component
	title                  string
	w                      uint16 // 地图宽
	h                      uint16 // 地图高
	heroPosEnabled         bool
	heroPos                fpoint.FPoint // 默认主角出生位置
	parallaxFilename       string        // 视差图层文件定义
	backgroundColor        color.Color
}

func ConstructMap() Map {
	m := Map{}
	m.init()

	return m
}

func (this *Map) init() {
	this.w = 1
	this.h = 1
	this.heroPos = fpoint.Construct()
	this.backgroundColor = color.Construct()
}

func (this *Map) Close(impl gameres.MapRenderer) {
	impl.Clear()

	this.clear()
}

func (this *Map) clear() {
}

func (this *Map) ClearLayers() {
	this.layers = nil
	this.layerNames = nil
}

func (this *Map) ClearQueues() {
	this.enemies = nil
	this.npcs = nil
}

func (this *Map) ClearEvents() {
	this.events = nil
	this.delayedEvents = nil
	for _, ptr := range this.statBlocks {
		ptr.Close()
	}
	this.statBlocks = nil
}

func (this *Map) RemoveLayer(index int) {
	leftLn := this.layerNames[index+1:]
	this.layerNames = this.layerNames[:len(this.layerNames)-1]
	for i, val := range leftLn {
		this.layerNames[index+i] = val
	}

	leftL := this.layers[index+1:]
	this.layers = this.layers[:len(this.layers)-1]
	for i, val := range leftL {
		this.layers[index+i] = val
	}
}

// 加载图层定义和地图触发事件
func (this *Map) Load(modules common.Modules, loot gameres.LootManager, camp gameres.CampaignManager, eventManager gameres.EventManager, gresf gameres.Factory, fname string) error {
	mods := modules.Mods()

	infile := fileparser.New()

	this.ClearEvents()
	this.ClearLayers()
	this.ClearQueues()

	this.musicFilename = ""
	this.parallaxFilename = ""
	this.backgroundColor = color.Construct(0, 0, 0, 0)
	this.w = 1
	this.h = 1
	this.heroPosEnabled = false
	this.heroPos.X = 0
	this.heroPos.Y = 0

	err := infile.Open(fname, true, mods)
	if err != nil {
		return err
	}
	defer infile.Close()

	fmt.Printf("Map: Loading map '%s'\n", fname)
	this.filename = fname

	for infile.Next(mods) {
		if infile.IsNewSection() {
			switch infile.GetSection() {
			case "enemy":
				this.enemyGroups = append(this.enemyGroups, constructMapGroup())
			case "npc":
				this.npcs = append(this.npcs, constructMapNPC())
			case "event":
				this.events = append(this.events, event.Construct())
			}
		}

		switch infile.GetSection() {
		case "header":
			err := this.loadHeader(modules, infile.Key(), infile.Val())
			if err != nil {
				panic("header")
				return err
			}
		case "layer":
			// 图层定义
			err := this.loadLayer(modules, infile)
			if err != nil {
				panic("layer")
				return err
			}
		case "enemy":
			err := this.loadEnemyGroup(camp, infile.Key(), infile.Val())
			if err != nil {
				return err
			}
		case "npc":
			err := this.loadNPC(camp, infile.Key(), infile.Val())
			if err != nil {
				return err
			}

		case "event":
			// 地图事件
			err := eventManager.LoadEvent(modules, loot, camp, infile.Key(), infile.Val(), &(this.events[len(this.events)-1]))
			if err != nil {
				return err
			}
		}
	}

	for _, val := range this.events {
		ecPower, ok := val.GetComponent(event.POWER)
		if ok {
			// 保存状态块的序号
			ecPower.X = this.AddEventStatBlock(modules, gresf, val)
		}
	}

	found := false
	for _, val := range this.layerNames {
		if val == "collision" {
			found = true
			break
		}
	}

	// 保证一定有碰撞图层
	if !found {
		this.layerNames = append(this.layerNames, "collision")
		tmp := make([][]uint16, this.w)
		for index, _ := range tmp {
			tmp[index] = make([]uint16, this.h)
		}

		this.layers = append(this.layers, tmp)
	}

	// TODO
	// fog of war

	return nil
}

// 头部元数据，尺寸，音乐，瓷砖信息
func (this *Map) loadHeader(modules common.Modules, key, val string) error {
	msg := modules.Msg()

	switch key {
	case "title":
		this.title = msg.Get(val)
	case "width":
		this.w = (uint16)(math.Max((float64)(parsing.ToInt(val, 0)), 1))
	case "height":
		this.h = (uint16)(math.Max((float64)(parsing.ToInt(val, 0)), 1))
	case "tileset":
		// 瓷砖定义
		this.tileset = val
	case "music":
		this.musicFilename = val
	case "hero_pos":
		var x, y int
		x, val = parsing.PopFirstInt(val, "")
		this.heroPos.X = float32(x) + 0.5
		y, val = parsing.PopFirstInt(val, "")
		this.heroPos.Y = float32(y) + 0.5
		this.heroPosEnabled = true
	case "parallax_layers":
		this.parallaxFilename = val
	case "background_color":
		this.backgroundColor = parsing.ToRGBA(val)
	case "fogofwar":
		// TODO
	case "save_fogofwar":
		// TODO
	case "tilewidth":

	case "tileheight":

	case "orientation":
	default:
		return fmt.Errorf("Map: '%s' is not a valid key.\n", key)
	}

	return nil
}

func (this *Map) loadLayer(modules common.Modules, infile *fileparser.FileParser) error {
	switch infile.Key() {
	case "type":
		tmp := make([][]uint16, this.w)
		for i, _ := range tmp {
			tmp[i] = make([]uint16, this.h)
		}

		this.layers = append(this.layers, tmp)
		this.layerNames = append(this.layerNames, infile.Val())
	case "format":
		if infile.Val() != "dec" {
			return fmt.Errorf("Map: The format of a layer must be 'dec'!")
		}
	case "data":
		for j := (uint16)(0); j < this.h; j++ {
			val := infile.GetRawLine()
			infile.IncrementLineNum()
			if val != "" && !strings.HasSuffix(val, ",") {
				val += ","
			}
			commaCount := strings.Count(val, ",")
			if commaCount != (int)(this.w) {
				return fmt.Errorf("Map: A row of layer data has a width not equal to %d.\n", this.w)
			}

			var first int
			for i := (uint16)(0); i < this.w; i++ {
				first, val = parsing.PopFirstInt(val, "")
				this.layers[len(this.layers)-1][i][j] = (uint16)(first)
			}
		}
	default:
		return fmt.Errorf("Map: '%s' is not a valid key.\n", infile.Key())
	}

	return nil
}

func (this *Map) loadEnemyGroup(camp gameres.CampaignManager, key, val string) error {

	var first int
	group := &(this.enemyGroups[len(this.enemyGroups)-1])
	switch key {
	case "type":
		group.type1 = val
	case "category":
		group.category = val
	case "level":
		first, val = parsing.PopFirstInt(val, "")
		group.levelMin = (int)(math.Max(0, (float64)(first)))
		first, val = parsing.PopFirstInt(val, "")
		group.levelMax = (int)(math.Max((float64)(group.levelMin), (float64)(first)))
	case "location":
		first, val = parsing.PopFirstInt(val, "")
		group.pos.X = first
		first, val = parsing.PopFirstInt(val, "")
		group.pos.Y = first
		first, val = parsing.PopFirstInt(val, "")
		group.area.X = first
		first, val = parsing.PopFirstInt(val, "")
		group.area.Y = first
	case "number":
		first, val = parsing.PopFirstInt(val, "")
		group.numberMin = (int)(math.Max(0, (float64)(first)))
		first, val = parsing.PopFirstInt(val, "")
		group.numberMax = (int)(math.Max((float64)(group.numberMin), (float64)(first)))
	case "chance":
		first, val = parsing.PopFirstInt(val, "")
		n := math.Max(0, (float64)(first)) / 100
		group.chance = (float32)(math.Min(n, 1))
	case "direction":
		group.direction = parsing.ToDirection(val)
	case "waypoints":
		var none, a, b string
		a, val = parsing.PopFirstString(val, "")
		b, val = parsing.PopFirstString(val, "")

		for a != none {
			p := fpoint.Construct()
			p.X = (float32)(parsing.ToInt(a, 0)) + 0.5
			p.Y = (float32)(parsing.ToInt(b, 0)) + 0.5
			group.wayPoints = append(group.wayPoints, p)
			a, val = parsing.PopFirstString(val, "")
			b, val = parsing.PopFirstString(val, "")
		}
		group.wanderRadius = 0
	case "wander_radius":
		first, val = parsing.PopFirstInt(val, "")
		group.wanderRadius = (int)(math.Max(0, (float64)(first)))
		group.wayPoints = nil
	case "requires_status":
		s := ""
		s, val = parsing.PopFirstString(val, "")
		for s != "" {
			ec := event.ConstructComponent()
			ec.Type = event.REQUIRES_STATUS
			ec.Status = camp.RegisterStatus(s)
			group.requirements = append(group.requirements, ec)
			s, val = parsing.PopFirstString(val, "")
		}

	case "requires_not_status":
		s := ""
		s, val = parsing.PopFirstString(val, "")
		for s != "" {
			ec := event.ConstructComponent()
			ec.Type = event.REQUIRES_NOT_STATUS
			ec.Status = camp.RegisterStatus(s)
			group.requirements = append(group.requirements, ec)
			s, val = parsing.PopFirstString(val, "")
		}

	case "requires_level":
		ec := event.ConstructComponent()
		ec.Type = event.REQUIRES_LEVEL
		ec.X, _ = parsing.PopFirstInt(val, "")
		group.requirements = append(group.requirements, ec)
	case "requires_not_level":
		ec := event.ConstructComponent()
		ec.Type = event.REQUIRES_NOT_LEVEL
		ec.X, _ = parsing.PopFirstInt(val, "")
		group.requirements = append(group.requirements, ec)
	case "requires_currency":
		ec := event.ConstructComponent()
		ec.Type = event.REQUIRES_CURRENCY
		ec.X, _ = parsing.PopFirstInt(val, "")
		group.requirements = append(group.requirements, ec)
	case "requires_not_currency":
		ec := event.ConstructComponent()
		ec.Type = event.REQUIRES_NOT_CURRENCY
		ec.X, _ = parsing.PopFirstInt(val, "")
		group.requirements = append(group.requirements, ec)
	case "requires_item":
		s := ""
		s, val = parsing.PopFirstString(val, "")
		for s != "" {
			itemStack, _ := parsing.ToItemQuantityPair(s)
			ec := event.ConstructComponent()
			ec.Type = event.REQUIRES_ITEM
			ec.Id = (int)(itemStack.Item)
			ec.X = itemStack.Quantity
			group.requirements = append(group.requirements, ec)
			s, val = parsing.PopFirstString(val, "")
		}

	case "requires_not_item":
		s := ""
		s, val = parsing.PopFirstString(val, "")
		for s != "" {
			itemStack, _ := parsing.ToItemQuantityPair(s)
			ec := event.ConstructComponent()
			ec.Type = event.REQUIRES_NOT_ITEM
			ec.Id = (int)(itemStack.Item)
			ec.X = itemStack.Quantity
			group.requirements = append(group.requirements, ec)
			s, val = parsing.PopFirstString(val, "")
		}

	case "requires_class":
		ec := event.ConstructComponent()
		ec.Type = event.REQUIRES_CLASS
		ec.S, _ = parsing.PopFirstString(val, "")
		group.requirements = append(group.requirements, ec)

	case "requires_not_class":
		ec := event.ConstructComponent()
		ec.Type = event.REQUIRES_NOT_CLASS
		ec.S, _ = parsing.PopFirstString(val, "")
		group.requirements = append(group.requirements, ec)

	case "invincible_requires_status":
		s := ""
		s, val = parsing.PopFirstString(val, "")
		for s != "" {
			ec := event.ConstructComponent()
			ec.Type = event.REQUIRES_STATUS
			ec.Status = camp.RegisterStatus(s)
			group.invincibleRequirements = append(group.invincibleRequirements, ec)
			s, val = parsing.PopFirstString(val, "")
		}

	case "invincible_requires_not_status":
		s := ""
		s, val = parsing.PopFirstString(val, "")
		for s != "" {
			ec := event.ConstructComponent()
			ec.Type = event.REQUIRES_NOT_STATUS
			ec.Status = camp.RegisterStatus(s)
			group.invincibleRequirements = append(group.invincibleRequirements, ec)
			s, val = parsing.PopFirstString(val, "")
		}

	default:
		return fmt.Errorf("Map: '%s' is not a valid key.\n", key)
	}

	return nil
}

func (this *Map) loadNPC(camp gameres.CampaignManager, key, val string) error {
	npc := &(this.npcs[len(this.npcs)-1])

	switch key {
	case "type":
		npc.type1 = val
	case "filename":
		npc.id = val
	case "location":
		var first int
		first, val = parsing.PopFirstInt(val, "")
		npc.pos.X = (float32)(first)
		first, val = parsing.PopFirstInt(val, "")
		npc.pos.Y = (float32)(first)
	case "requires_status":
		s := ""
		s, val = parsing.PopFirstString(val, "")
		for s != "" {
			ec := event.ConstructComponent()
			ec.Type = event.REQUIRES_STATUS
			ec.Status = camp.RegisterStatus(s)
			npc.requirements = append(npc.requirements, ec)
			s, val = parsing.PopFirstString(val, "")
		}

	case "requires_not_status":
		s := ""
		s, val = parsing.PopFirstString(val, "")
		for s != "" {
			ec := event.ConstructComponent()
			ec.Type = event.REQUIRES_NOT_STATUS
			ec.Status = camp.RegisterStatus(s)
			npc.requirements = append(npc.requirements, ec)
			s, val = parsing.PopFirstString(val, "")
		}

	case "requires_level":
		ec := event.ConstructComponent()
		ec.Type = event.REQUIRES_LEVEL
		ec.X, _ = parsing.PopFirstInt(val, "")
		npc.requirements = append(npc.requirements, ec)

	case "requires_not_level":
		ec := event.ConstructComponent()
		ec.Type = event.REQUIRES_NOT_LEVEL
		ec.X, _ = parsing.PopFirstInt(val, "")
		npc.requirements = append(npc.requirements, ec)

	case "requires_currency":
		ec := event.ConstructComponent()
		ec.Type = event.REQUIRES_CURRENCY
		ec.X, _ = parsing.PopFirstInt(val, "")
		npc.requirements = append(npc.requirements, ec)

	case "requires_not_currency":
		ec := event.ConstructComponent()
		ec.Type = event.REQUIRES_NOT_CURRENCY
		ec.X, _ = parsing.PopFirstInt(val, "")
		npc.requirements = append(npc.requirements, ec)

	case "requires_item":
		s := ""
		s, val = parsing.PopFirstString(val, "")
		for s != "" {
			itemStack, _ := parsing.ToItemQuantityPair(s)
			ec := event.ConstructComponent()
			ec.Type = event.REQUIRES_ITEM
			ec.Id = (int)(itemStack.Item)
			ec.X = itemStack.Quantity
			npc.requirements = append(npc.requirements, ec)
			s, val = parsing.PopFirstString(val, "")
		}

	case "requires_not_item":
		s := ""
		s, val = parsing.PopFirstString(val, "")
		for s != "" {
			itemStack, _ := parsing.ToItemQuantityPair(s)
			ec := event.ConstructComponent()
			ec.Type = event.REQUIRES_NOT_ITEM
			ec.Id = (int)(itemStack.Item)
			ec.X = itemStack.Quantity
			npc.requirements = append(npc.requirements, ec)
			s, val = parsing.PopFirstString(val, "")
		}

	case "requires_class":
		ec := event.ConstructComponent()
		ec.Type = event.REQUIRES_CLASS
		ec.S, _ = parsing.PopFirstString(val, "")
		npc.requirements = append(npc.requirements, ec)

	case "requires_not_class":
		ec := event.ConstructComponent()
		ec.Type = event.REQUIRES_NOT_CLASS
		ec.S, _ = parsing.PopFirstString(val, "")
		npc.requirements = append(npc.requirements, ec)

	default:
		return fmt.Errorf("Map: '%s' is not a valid key.\n", key)
	}

	return nil
}

func (this *Map) AddEventStatBlock(modules common.Modules, gresf gameres.Factory, evnt event.Event) int {
	eset := modules.Eset()

	this.statBlocks = append(this.statBlocks, gresf.New("statblock").(gameres.StatBlock).Init(modules, gresf))
	statb := this.statBlocks[len(this.statBlocks)-1]

	statb.SetPerfectAccuracy(true)

	ecPath, ok := evnt.GetComponent(event.POWER_PATH) // 技能路径，起始到终点
	if ok {
		statb.SetPos(fpoint.Construct(float32(ecPath.X)+0.5, float32(ecPath.X)+0.5)) // 起点
	} else {
		statb.SetPos(fpoint.Construct(float32(evnt.Location.X)+0.5, float32(evnt.Location.X)+0.5)) // 事件位置作为启动
	}

	// 事件配置是否要覆盖配置默认值
	ecDamage, ok := evnt.GetComponent(event.POWER_DAMAGE)
	if ok {
		dtCount := eset.Get("damage_types", "count").(int)
		for i := 0; i < dtCount; i++ {
			if i%2 == 0 {
				statb.SetStarting(stats.COUNT+i, ecDamage.X) // min
			} else {
				statb.SetStarting(stats.COUNT+i, ecDamage.Y) // max
			}
		}
	}

	// TODO
	// effect
	// statblock

	return len(this.statBlocks) - 1
}

func (this *Map) GetLayers() []([][]uint16) {
	return this.layers
}

func (this *Map) GetLayer(index int) [][]uint16 {
	return this.layers[index]
}

func (this *Map) GetLayerName(index int) string {
	return this.layerNames[index]
}

func (this *Map) GetTileSet() string {
	return this.tileset
}

func (this *Map) GetBackgroundColor() color.Color {
	return this.backgroundColor
}

func (this *Map) GetH() uint16 {
	return this.h
}

func (this *Map) GetW() uint16 {
	return this.w
}

func (this *Map) RegisterDelayedEvent(e event.Event) {
	this.delayedEvents[len(this.delayedEvents)] = e
}

func (this *Map) GetEvents() []event.Event {
	return this.events
}

func (this *Map) SetEvents(events []event.Event) {
	this.events = events
}

func (this *Map) GetHeroPosEnabled() bool {
	return this.heroPosEnabled
}

func (this *Map) GetHeroPos() fpoint.FPoint {
	return this.heroPos
}

func (this *Map) GetFilename() string {
	return this.filename
}
