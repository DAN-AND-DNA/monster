package state

import (
	"fmt"
	"monster/pkg/common"
	"monster/pkg/common/define"
	"monster/pkg/common/gameres"
	"monster/pkg/common/gameres/avatar"
	"monster/pkg/common/timer"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/game/base"
	"monster/pkg/utils"
	"monster/pkg/utils/parsing"
)

// 角色称号
type RoleTitle struct {
	title             string
	level             int // 要求的等级
	power             define.PowerId
	requiresStatus    []define.StatusId
	requiresNotStatus []define.StatusId
	primaryStat1      string
	primaryStat2      string
}

func constructRoleTitle() RoleTitle {
	rt := RoleTitle{}

	return rt
}

type Play struct {
	base.State
	secondTimer    timer.Timer
	npcId          int
	titles         []RoleTitle
	isFirstMapLoad bool
}

func NewPlay(modules common.Modules, gameRes gameres.GameRes) *Play {
	play := &Play{}

	play.init(modules, gameRes)

	return play
}

func (this *Play) init(modules common.Modules, gameRes gameres.GameRes) gameres.GameStatePlay {
	settings := modules.Settings()

	items := gameRes.Items()
	ss := gameRes.Stats()
	menuf := gameRes.Menuf()
	gresf := gameRes.Resf()

	if items == nil {
		items = gameRes.NewItems(modules, ss)
	}

	camp := gameRes.NewCamp()
	_ = camp
	loot := gameRes.NewLoot(modules, items)
	_ = loot
	powers := gameRes.NewPowers(modules, ss)
	mapr := gameRes.NewMapr(modules, gresf)
	pc := gameRes.NewPc(modules, mapr, ss, powers, gresf)
	menu := gameRes.NewMenu(modules, pc, powers, menuf)
	_ = menu
	eventManager := gameRes.NewEventManager()
	_ = eventManager

	// base
	this.State = base.ConstructState(modules)

	// self
	this.secondTimer.SetDuration((uint)(settings.Get("max_fps").(int)))
	this.npcId = -1
	this.isFirstMapLoad = true

	err := this.loadTitles(modules, gameRes)
	if err != nil {
		panic(err)
	}

	return this
}

func (this *Play) Clear(modules common.Modules, gameRes gameres.GameRes) {
	camp := gameRes.Camp()
	loot := gameRes.Loot()
	menu := gameRes.Menu()
	mapr := gameRes.Mapr()
	powers := gameRes.Powers()
	eventManager := gameRes.NewEventManager()
	pc := gameRes.Pc()

	if camp != nil {
		camp.Close()
	}

	if loot != nil {
		loot.Close(modules)
	}

	if menu != nil {
		menu.Close()
	}

	if mapr != nil {
		mapr.Close()
	}

	if powers != nil {
		powers.Close()
	}

	if eventManager != nil {
		eventManager.Close()
	}

	if pc != nil {
		pc.Close(modules)
	}
}

func (this *Play) Close(modules common.Modules, gameRes gameres.GameRes) {
	this.State.Close(modules, gameRes, this)
}

func (this *Play) RefreshWidgets(modules common.Modules, gameRes gameres.GameRes) error {
	menu := gameRes.Menu()

	err := menu.AlignAll(modules)
	if err != nil {
		return err
	}

	return nil
}

func (this *Play) Logic(modules common.Modules, gameRes gameres.GameRes) error {
	inpt := modules.Inpt()

	mapr := gameRes.Mapr()
	menu := gameRes.Menu()
	pc := gameRes.Pc()
	camp := gameRes.Camp()
	powers := gameRes.Powers()

	if inpt.GetWindowResized() {
		this.RefreshWidgets(modules, gameRes)
	}

	// 顶层先
	menu.Logic(modules, pc, powers)

	if !this.isPaused() {
		pc.Logic(modules, mapr, camp)
	}

	err := this.checkTeleport(modules, gameRes)
	if err != nil {
		return err
	}

	err = this.checkEquipmentChange(modules, gameRes)
	if err != nil {
		return err
	}

	mapr.Logic(modules)

	return nil
}

func (this *Play) Render(modules common.Modules, gameRes gameres.GameRes) error {
	menu := gameRes.Menu()
	mapr := gameRes.Mapr()
	pc := gameRes.Pc()

	if mapr.GetIsSpawnMap() {
		return nil
	}

	var rens, rensDead []common.Renderable

	rens = pc.AddRenders(modules, rens)

	err := mapr.Render(modules, rens, rensDead)
	if err != nil {
		return err
	}

	err = menu.Render(modules)
	if err != nil {
		return err
	}

	return nil
}

// 重置游戏
func (this *Play) ResetGame(modules common.Modules, gameRes gameres.GameRes) {
	camp := gameRes.Camp()
	pc := gameRes.Pc()
	mapr := gameRes.Mapr()
	ss := gameRes.Stats()
	powers := gameRes.Powers()
	menu := gameRes.Menu()

	camp.ResetAllStatuses()
	pc.Init(modules, mapr, ss, powers)

	// TODO
	// menu
	menu.Get("inv").(gameres.MenuInventory).SetChangedEquipment(true)
	menu.Get("inv").(gameres.MenuInventory).SetCurrency(0)

	// 默认传送到出生点地图
	mapr.SetTeleportation(true)
	mapr.SetTeleportMapName("maps/spawn.txt")
}

// 加载玩家称号
func (this *Play) loadTitles(modules common.Modules, gameRes gameres.GameRes) error {
	mods := modules.Mods()

	camp := gameRes.Camp()

	infile := fileparser.New()
	err := infile.Open("engine/titles.txt", true, mods)
	if err != nil {
		return err
	}
	defer infile.Close()

	for infile.Next(mods) {
		key := infile.Key()
		val := infile.Val()

		if infile.IsNewSection() && infile.GetSection() == "title" {
			t := constructRoleTitle()
			this.titles = append(this.titles, t)
		}

		if len(this.titles) == 0 {
			continue
		}

		// 获得的要求
		switch key {
		case "title":
			this.titles[len(this.titles)-1].title = val
		case "level":
			this.titles[len(this.titles)-1].level = parsing.ToInt(val, 0)
		case "power":
			this.titles[len(this.titles)-1].power = (define.PowerId)(parsing.ToInt(val, 0))
		case "requires_status":
			var repeatVal string
			repeatVal, val = parsing.PopFirstString(val, "")
			for repeatVal != "" {

				this.titles[len(this.titles)-1].requiresStatus = append(this.titles[len(this.titles)-1].requiresStatus, camp.RegisterStatus(repeatVal))
				repeatVal, val = parsing.PopFirstString(val, "")
			}

		case "requires_not_status":
			var repeatVal string
			repeatVal, val = parsing.PopFirstString(val, "")
			for repeatVal != "" {

				this.titles[len(this.titles)-1].requiresNotStatus = append(this.titles[len(this.titles)-1].requiresNotStatus, camp.RegisterStatus(repeatVal))
				repeatVal, val = parsing.PopFirstString(val, "")
			}

		case "primary_stat":
			var first string
			first, val = parsing.PopFirstString(val, "")
			this.titles[len(this.titles)-1].primaryStat1 = first
			first, val = parsing.PopFirstString(val, "")
			this.titles[len(this.titles)-1].primaryStat2 = first
		default:
			return fmt.Errorf("GameStatePlay: '%s' is not a valid key.\n", key)
		}
	}

	return nil
}

func (this *Play) checkTeleport(modules common.Modules, gameRes gameres.GameRes) error {
	inpt := modules.Inpt()

	mapr := gameRes.Mapr()
	loot := gameRes.Loot()
	camp := gameRes.Camp()
	eventManager := gameRes.EventManager()
	gresf := gameRes.Resf()
	pc := gameRes.Pc()

	onLoadTeleport := false

	if mapr.GetTeleportation() {

		if mapr.GetTeleportation() {
			// TODO
			pc.GetStats().SetPos(mapr.GetTeleportDestination())
		} else {
		}

		if mapr.GetTeleportMapName() == "" {
			// TODO
			// 改变位置
		}

		if mapr.GetTeleportation() && mapr.GetTeleportMapName() != "" {
			mapr.GetCam().WarpTo(pc.GetStats().GetPos())
			teleportMapName := mapr.GetTeleportMapName()
			mapr.SetTeleportMapName("")
			inpt.SetLockAll(mapr.GetTeleportMapName() == "maps/spawn.txt")
			err := this.ShowLoading(modules)
			if err != nil {
				return err
			}

			err = mapr.Load(modules, loot, camp, eventManager, gresf, teleportMapName)
			if err != nil {
				return err
			}
		}

		// 清空请求换图的状态
		mapr.SetTeleportation(false)

		// 处理地图加载事件
		mapr.ExecuteOnLoadEvent(modules, eventManager, camp)

		if mapr.GetTeleportation() {
			onLoadTeleport = true
		}
	}

	if !onLoadTeleport && mapr.GetTeleportMapName() == "" {
		mapr.SetTeleportation(false)
	}

	return nil
}

func (this *Play) checkEquipmentChange(modules common.Modules, gameRes gameres.GameRes) error {
	mods := modules.Mods()
	settings := modules.Settings()

	menu := gameRes.Menu()
	pc := gameRes.Pc()

	inv := menu.Get("inv").(gameres.MenuInventory)

	// 第一次肯定会触发图片加载
	if inv.GetChangedEquipment() {
		feetIndex := -1
		_ = feetIndex
		var imgGfx []avatar.LayerGfx
		layerOrder := pc.GetLayerReferenceOrder()

		for _, val := range layerOrder {
			gfx := avatar.ConstructLayerGfx()
			gfx.Type = val

			//TODO
			// menu inv

			// 没有头盔时用默认头展现
			if gfx.Gfx == "" && val == "head" {
				gfx.Gfx = pc.GetStats().GetGfxHead()
				gfx.Type = "head"
			}

			// 其他位置没有装备
			if gfx.Gfx == "" {
				_, err := mods.Locate(settings, "animations/avatar/"+pc.GetStats().GetGfxBase()+"/default_"+gfx.Type+".txt")
				if err != nil && !utils.IsNotExist(err) {
					return err
				} else if err == nil {
					gfx.Gfx = "default_" + gfx.Type
				}
			}
			imgGfx = append(imgGfx, gfx)
		}

		err := pc.LoadGraphics(modules, imgGfx)
		if err != nil {
			return err
		}

		// TODO
		// sound
	}

	inv.SetChangedEquipment(false)

	return nil
}

func (this *Play) isPaused() bool {
	return false
}
