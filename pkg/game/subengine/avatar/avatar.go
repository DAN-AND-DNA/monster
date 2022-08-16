package avatar

import (
	"fmt"
	"math"
	"monster/pkg/common"
	"monster/pkg/common/define"
	"monster/pkg/common/define/game/mapcollision"
	"monster/pkg/common/define/game/stats"
	"monster/pkg/common/define/inputstate"
	"monster/pkg/common/define/renderable"
	"monster/pkg/common/fpoint"
	"monster/pkg/common/gameres"
	"monster/pkg/common/gameres/avatar"
	"monster/pkg/common/gameres/power"
	"monster/pkg/common/gameres/statblock"
	"monster/pkg/common/timer"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/game/base"
	"monster/pkg/utils"
	"monster/pkg/utils/parsing"
)

type Avatar struct {
	base.Entity

	animsets            []common.AnimationSet
	anims               []common.Animation
	body                int16
	transformTriggered  bool // 在释放变身技能
	lastTransform       string
	mmKey               int
	setDirTimer         timer.Timer
	layerReferenceOrder []string
	layerDef            [][]uint
	setPowers           bool
	revertPowers        bool
	untransformPower    define.PowerId

	transformPos fpoint.FPoint
	transformMap string

	currentPower         define.PowerId // 当前技能
	actTarget            fpoint.FPoint
	dragWalking          bool // 按住移动
	newLevelNotification bool
	respawn              bool
	allowMovement        bool
	powerCooldownTimers  map[define.PowerId]*timer.Timer
	powerCastTimers      map[define.PowerId]*timer.Timer // 技能释法动作时间

	cursorEnemy        gameres.Entity
	lockEnemy          gameres.Entity
	timePlayed         uint64 // 游戏时间
	usingMain1         bool
	usingMain2         bool
	prevHP             int
	teleportCameraLock bool
}

func New(modules common.Modules, mapr gameres.MapRenderer, ss gameres.Stats, powers gameres.PowerManager, gresf gameres.Factory) *Avatar {
	a := &Avatar{}
	a.init(modules, mapr, ss, powers, gresf)

	return a
}

func (this *Avatar) init(modules common.Modules, mapr gameres.MapRenderer, ss gameres.Stats, powers gameres.PowerManager, gresf gameres.Factory) gameres.Avatar {
	settings := modules.Settings()
	anim := modules.Anim()
	mods := modules.Mods()
	render := modules.Render()
	mresf := modules.Resf()

	// base
	this.Entity = base.ConstructEntity(modules, gresf)

	// self
	this.mmKey = inputstate.MAIN1
	if settings.Get("mouse_move_swap").(bool) {
		this.mmKey = inputstate.MAIN2
	}
	this.actTarget = fpoint.Construct()
	this.allowMovement = true

	this.Init(modules, mapr, ss, powers)

	anim.IncreaseCount("animations/hero.txt")
	aSet, err := anim.GetAnimationSet(settings, mods, render, mresf, "animations/hero.txt")
	if err != nil {
		panic(err)
	}

	this.Entity.SetAnimationSet(aSet) // 总的配置，用来统一后面加载每个同名动画的帧数
	this.Entity.SetActiveAnimation(aSet.GetAnimation(""))
	stats := this.Entity.GetStats()

	// 配置里没有写被击中的冷却时间
	if !stats.GetCooldownHitEnabled() {
		hitAnim := aSet.GetAnimation("hit")
		if hitAnim != nil {
			// 被击中的冷却时间和动画时间一致
			stats.GetCooldownHit().SetDuration((uint)(hitAnim.GetDuration()))
			hitAnim.Close()
		} else {
			stats.GetCooldownHit().SetDuration(0)
		}
	}

	// 加载精灵图层定义
	err = this.loadLayerDefinitions(modules)
	if err != nil {
		panic(err)
	}

	// TODO
	// sound

	return this
}

func (this *Avatar) Init(modules common.Modules, mapr gameres.MapRenderer, ss gameres.Stats, powers gameres.PowerManager) {
	eset := modules.Eset()

	// 清空图片
	this.Entity.SetSprites(nil)

	stats := this.Entity.GetStats()
	stats.SetCurState(statblock.ENTITY_STANCE)

	if mapr.GetHeroPosEnabled() {
		// 设置配置文件里英雄的起始位置
		stats.SetPos(mapr.GetHeroPos())
	}

	this.currentPower = 0
	this.newLevelNotification = false

	// 作为英雄
	stats.SetHero(true)
	stats.SetHumanoid(true)
	stats.SetLevel(1)
	stats.SetXp(0)

	// 属性
	pList := eset.Get("primary_stats", "list").([]common.PrimaryStat)
	for index, _ := range pList {
		stats.SetPrimary(index, 1)
		stats.SetPrimaryStarting(index, 1)
		stats.SetPrimaryAdditional(index, 0)
	}

	stats.SetSpeed(0.2)
	// 重新计算英雄的等级和状态
	stats.Recalc(modules, ss)

	// TODO
	// log msg

	this.respawn = false

	// 攻击间隔
	stats.GetCooldown().Reset(timer.END)
	this.body = -1
	this.transformTriggered = false
	this.setPowers = false
	this.revertPowers = false
	this.lastTransform = ""
	this.powerCooldownTimers = map[define.PowerId]*timer.Timer{}
	this.powerCastTimers = map[define.PowerId]*timer.Timer{}

	allPowers := powers.GetPowers()
	this.untransformPower = 0
	for id, ptr := range allPowers {
		if this.untransformPower == 0 && len(ptr.RequiredItems) == 0 && ptr.SpawnType == "untransform" {
			this.untransformPower = id
		}

		this.powerCooldownTimers[id] = timer.New()
		this.powerCastTimers[id] = timer.New()
	}
}

func (this *Avatar) Clear(modules common.Modules) {

	anim := modules.Anim()

	anim.DecreaseCount("animations/hero.txt")

	for i, ptr := range this.animsets {
		if ptr != nil {
			anim.DecreaseCount(ptr.GetName())
		}

		if this.anims[i] != nil {
			this.anims[i].Close()
		}
	}

	anim.CleanUp()
}

func (this *Avatar) Close(modules common.Modules) {
	this.Entity.Close(modules, this)
}

func (this *Avatar) SetAnimation(name string) {
	if name == this.Entity.GetActiveAnimation().GetName() {
		return
	}

	this.Entity.SetAnimation(name)
	for i := 0; i < len(this.animsets); i++ {
		if this.anims[i] != nil {
			this.anims[i].Close()
		}

		if this.animsets[i] != nil {
			this.anims[i] = this.animsets[i].GetAnimation(name)
		} else {
			this.anims[i] = nil
		}
	}
}

func (this *Avatar) LoadGraphics(modules common.Modules, imageGfx []avatar.LayerGfx) error {
	settings := modules.Settings()
	mods := modules.Mods()
	render := modules.Render()
	anim := modules.Anim()
	mresf := modules.Resf()

	stats := this.Entity.GetStats()
	for i, ptr := range this.animsets {
		if ptr != nil {
			anim.DecreaseCount(ptr.GetName())
		}
		this.anims[i].Close()
	}

	this.animsets = nil
	this.anims = nil

	defer anim.CleanUp()

	for _, val := range imageGfx {
		if val.Gfx != "" {
			// 每个txt包含一堆动画
			name := "animations/avatar/" + stats.GetGfxBase() + "/" + val.Gfx + ".txt"
			anim.IncreaseCount(name)
			aSet, err := anim.GetAnimationSet(settings, mods, render, mresf, name)
			if err != nil {
				return err
			}

			aSet.SetParent(this.Entity.GetAnimationSet())
			this.animsets = append(this.animsets, aSet)
			this.anims = append(this.anims, aSet.GetAnimation(this.Entity.GetActiveAnimation().GetName()))
			this.SetAnimation("stance")

			if !this.anims[len(this.anims)-1].SyncTo(this.Entity.GetActiveAnimation()) {
				return fmt.Errorf("Avatar: Error syncing animation in '%s' to 'animations/hero.txt'.\n", this.animsets[len(this.animsets)-1].GetName())
			}

		} else {
			this.animsets = append(this.animsets, nil)
			this.anims = append(this.anims, nil)
		}
	}

	return nil
}

// 根据方向选择对应的图片
func (this *Avatar) AddRenders(modules common.Modules, r []common.Renderable) []common.Renderable {
	stats := this.Entity.GetStats()

	if !stats.GetTransformed() {
		// 不是变身状态

		for i, index := range this.layerDef[stats.GetDirection()] {
			if this.anims[index] != nil {
				ren := this.anims[index].GetCurrentFrame(modules, (int)(stats.GetDirection()))
				ren.SetMapPos(stats.GetPos())
				ren.SetPrio(uint64(i) + 1)

				// 颜色同最后添加的效果和透明度
				ren.SetColorMod(stats.GetEffects().GetCurrentColor(ren.GetColorMod()))
				ren.SetAlphaMod(stats.GetEffects().GetCurrentAlpha(ren.GetAlphaMod()))
				if stats.GetHP() > 0 {
					ren.SetType(renderable.TYPE_HERO)
				}
				r = append(r, ren)
			}
		}

	} else {

	}

	return r
}

// 加载精灵图层定义
func (this *Avatar) loadLayerDefinitions(modules common.Modules) error {
	mods := modules.Mods()

	this.layerDef = make([][]uint, 8)
	this.layerReferenceOrder = nil

	infile := fileparser.New()
	err := infile.Open("engine/hero_layers.txt", true, mods)
	if err != nil {
		return err
	}
	defer infile.Close()

	for infile.Next(mods) {
		key := infile.Key()
		val := infile.Val()

		switch key {
		case "layer":
			var first string
			first, val = parsing.PopFirstString(val, "")
			dir := parsing.ToDirection(first)

			if dir > 7 {
				return fmt.Errorf("Avatar: Hero layer direction must be in range [0,7]\n")
			}

			first, val = parsing.PopFirstString(val, "")
			for first != "" {
				refPos := 0
				for ; refPos < len(this.layerReferenceOrder); refPos++ {
					if first == this.layerReferenceOrder[refPos] {
						break
					}
				}

				if refPos == len(this.layerReferenceOrder) {
					this.layerReferenceOrder = append(this.layerReferenceOrder, first)
				}

				this.layerDef[dir] = append(this.layerDef[dir], (uint)(refPos))

				first, val = parsing.PopFirstString(val, "")
			}
		default:
			return fmt.Errorf("Avatar: '%s' is not a valid key.\n", key)
		}
	}

	return nil
}

func (this *Avatar) pressingMove(modules common.Modules) bool {
	settings := modules.Settings()
	inpt := modules.Inpt()

	if !this.allowMovement || this.teleportCameraLock {
		return false
	} else if this.GetStats().GetEffects().GetKnockbackSpeed() != 0 {
		return false
	} else if settings.Get("mouse_move").(bool) {
		return inpt.GetPressing(this.mmKey) && !inpt.GetPressing(inputstate.SHIFT)
	}

	// 方向键
	return (inpt.GetPressing(inputstate.UP) && !inpt.GetLock(inputstate.UP)) ||
		(inpt.GetPressing(inputstate.DOWN) && !inpt.GetLock(inputstate.DOWN)) ||
		(inpt.GetPressing(inputstate.LEFT) && !inpt.GetLock(inputstate.LEFT)) ||
		(inpt.GetPressing(inputstate.RIGHT) && !inpt.GetLock(inputstate.RIGHT))
}

func (this *Avatar) SetDirection(modules common.Modules, mapr gameres.MapRenderer) {
	settings := modules.Settings()
	inpt := modules.Inpt()
	eset := modules.Eset()

	//
	if this.teleportCameraLock || !this.setDirTimer.IsEnd() {
		return
	}

	oldPos := this.GetStats().GetPos()
	oldDir := this.GetStats().GetDirection()

	if settings.Get("mouse_move").(bool) {
		mouse := inpt.GetMouse()
		camPos := mapr.GetCam().GetPos()

		target := utils.ScreenToMap(settings, eset, mouse.X, mouse.Y, camPos.X, camPos.Y)
		this.GetStats().SetDirection(utils.CalcDirection(oldPos.X, oldPos.Y, target.X, target.Y))
	} else {
		// TODO
		// 方向键行走
	}

	if this.GetStats().GetDirection() != oldDir {
		// 每次方向变化 都需要100ms的冷却时间
		this.setDirTimer.SetDuration((uint)(settings.Get("max_fps").(int)) / 10)
	}
}

func (this *Avatar) Logic(modules common.Modules, mapr gameres.MapRenderer, camp gameres.CampaignManager) {
	settings := modules.Settings()
	inpt := modules.Inpt()
	eset := modules.Eset()

	restrictPowerUse := false
	_ = restrictPowerUse

	if settings.Get("mouse_move").(bool) {
		if inpt.GetPressing(this.mmKey) && !inpt.GetPressing(inputstate.SHIFT) {
			// TODO
			// menu
			restrictPowerUse = true
		}
	}

	// 暂时清理掉当前位置的碰撞
	mapr.GetCollider().Unblock(this.GetStats().GetPos().X, this.GetStats().GetPos().Y)

	if (this.GetStats().GetHP() > 0 || this.GetStats().GetEffects().GetTriggeredDeath()) && !this.respawn && !this.transformTriggered {
		// TODO
		// powers
		// 打开被动技能
	}

	if this.transformTriggered {
		this.transformTriggered = false
	}

	// 格挡技能结束
	if this.GetStats().GetEffects().GetTriggeredBlock() && !this.GetStats().GetBlocking() {
		this.GetStats().SetCurState(statblock.ENTITY_STANCE)
		this.GetStats().GetEffects().SetTriggeredBlock(false)                 // 恢复
		this.GetStats().GetEffects().ClearTriggerEffects(power.TRIGGER_BLOCK) // 清理掉格挡的触发效果
		this.GetStats().SetRefreshStats(true)
		this.GetStats().SetBlockPower(0) // 当前的格挡技能清空
	}

	// 计算状态值
	this.GetStats().Logic(modules, this, camp)

	if this.isDroppedToLowHP(modules) {
		// TODO
		// log msg
		// sound
	}

	// TODO
	// sound

	this.prevHP = this.GetStats().GetHP()

	if this.GetStats().GetLevel() < eset.XPGetMaxLevel() && this.GetStats().GetXp() >= eset.XPGetLevelXP(this.GetStats().GetLevel()+1) {
		// TODO
		// 可升级
	}

	this.mmKey = inputstate.MAIN1
	if settings.Get("mouse_move_swap").(bool) {
		this.mmKey = inputstate.MAIN2
	}

	if !inpt.GetPressing(this.mmKey) {
		this.dragWalking = false // 清理状态
	}

	this.usingMain1 = inpt.GetPressing(inputstate.MAIN1) && !inpt.GetLock(inputstate.MAIN1)
	this.usingMain2 = inpt.GetPressing(inputstate.MAIN2) && !inpt.GetLock(inputstate.MAIN2)

	if !this.GetStats().GetEffects().GetStun() {
		this.GetActiveAnimation().AdvanceFrame()
		for _, ptr := range this.anims {
			// 不同方向都同步到同一帧
			if ptr != nil {
				ptr.AdvanceFrame()
			}
		}
	}

	// 当前处于变身状态，暂存一个合法的当前位置，方便变身恢复
	if this.GetStats().GetTransformed() &&
		mapr.GetCollider().IsValidPosition(modules, this.GetStats().GetPos().X, this.GetStats().GetPos().Y, mapcollision.MOVE_NORMAL, mapcollision.COLLIDE_HERO) {

		this.transformPos = this.GetStats().GetPos()
		this.transformMap = mapr.GetFilename()
	}

	// TODO
	// menu
	// attack id
	mmCanUsePower := true

	if settings.Get("mouse_move").(bool) {
		if !inpt.GetPressing(this.mmKey) {
			// TODO
			// enemy
		}
	}

	if this.teleportCameraLock &&
		utils.CalcDist(this.GetStats().GetPos(), mapr.GetCam().GetPos()) < 0.5 {
		this.teleportCameraLock = false
	}

	// 转向冷却
	this.setDirTimer.Tick()

	if !this.pressingMove(modules) {
		this.setDirTimer.Reset(timer.END)
	}

	if !this.GetStats().GetEffects().GetStun() {

		allowedToMove := false
		allowedToTurn := false
		_ = allowedToTurn
		allowedToUsePower := true
		_ = allowedToUsePower

		switch this.GetStats().GetCurState() {
		case statblock.ENTITY_STANCE:
			this.SetAnimation("stance")

			if settings.Get("mouse_move").(bool) {
				// 鼠标移动
				allowedToMove = restrictPowerUse && (!inpt.GetLock(this.mmKey) || this.dragWalking) && this.lockEnemy == nil
				allowedToTurn = allowedToMove
				allowedToUsePower = true

				// 攻击
				if (inpt.GetPressing(this.mmKey) && inpt.GetPressing(inputstate.SHIFT)) || this.lockEnemy != nil {
					inpt.SetLock(this.mmKey, false)
				}

			} else if settings.Get("mouse_aim").(bool) {
				// 鼠标瞄准攻击

				// 非shift攻击
				allowedToMove = !inpt.GetPressing(inputstate.SHIFT)
				allowedToTurn = true
				allowedToUsePower = true

			} else {
				allowedToMove = true
				allowedToTurn = true
				allowedToUsePower = true
			}

			if allowedToTurn {
				this.SetDirection(modules, mapr)
			}

			if this.pressingMove(modules) && allowedToMove {
				if this.Entity.Move(modules, mapr) {
					if settings.Get("mouse_move").(bool) && inpt.GetPressing(this.mmKey) {
						inpt.SetLock(this.mmKey, true)
						this.dragWalking = true
					}

					this.GetStats().SetCurState(statblock.ENTITY_MOVE)
				}
			}

			// TODO
			// 攻击需要保持站立
			if settings.Get("mouse_move").(bool) &&
				settings.Get("mouse_move_attack").(bool) &&
				this.cursorEnemy != nil &&
				!this.GetStats().GetHeroAlly() &&
				mmCanUsePower {

				this.GetStats().SetCurState(statblock.ENTITY_STANCE)
				this.lockEnemy = this.cursorEnemy
			}

		case statblock.ENTITY_MOVE:
			this.SetAnimation("run")

			// TODO
			// sound

			this.SetDirection(modules, mapr)

			if !this.pressingMove(modules) {
				this.GetStats().SetCurState(statblock.ENTITY_STANCE)
			} else if !this.Entity.Move(modules, mapr) {
				this.GetStats().SetCurState(statblock.ENTITY_STANCE)
			} else if (settings.Get("mouse_move").(bool) || !settings.Get("mouse_aim").(bool)) && inpt.GetPressing(inputstate.SHIFT) {
				this.GetStats().SetCurState(statblock.ENTITY_STANCE)
			}

			if this.Entity.GetActiveAnimation().GetName() != "run" {
				this.GetStats().SetCurState(statblock.ENTITY_STANCE)
			}

			// TODO
			// 攻击需要保持站立
			if settings.Get("mouse_move").(bool) &&
				settings.Get("mouse_move_attack").(bool) &&
				this.cursorEnemy != nil &&
				!this.GetStats().GetHeroAlly() &&
				mmCanUsePower {

				this.GetStats().SetCurState(statblock.ENTITY_STANCE)
				this.lockEnemy = this.cursorEnemy
			}
		}

		if allowedToUsePower {
			// TODO 技能使用
		}

	}

	mapr.GetCam().SetTarget(this.Entity.GetStats().GetPos())
}

func (this *Avatar) isDroppedToLowHP(modules common.Modules) bool {
	settings := modules.Settings()

	hpOnePerc := math.Max(float64(this.GetStats().Get(stats.HP_MAX)), 1) / 100

	// 初次警告
	return (float64)(this.GetStats().GetHP())/hpOnePerc < (float64)(settings.Get("low_hp_threshold").(int)) &&
		(float64)(this.prevHP)/hpOnePerc >= (float64)(settings.Get("low_hp_threshold").(int))

}

func (this *Avatar) GetTimePlayed() uint64 {
	return this.timePlayed
}

func (this *Avatar) GetLayerReferenceOrder() []string {
	return this.layerReferenceOrder
}

func (this *Avatar) GetPowerCastTimersSize() int {
	return len(this.powerCastTimers)
}

func (this *Avatar) GetPowerCastTimer(id define.PowerId) *timer.Timer {
	return this.powerCastTimers[id]
}

func (this *Avatar) GetPowerCooldownTimer(id define.PowerId) *timer.Timer {
	return this.powerCooldownTimers[id]
}
