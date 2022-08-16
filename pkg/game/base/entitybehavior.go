package base

import (
	"monster/pkg/common"
	"monster/pkg/common/fpoint"
	"monster/pkg/common/gameres"
	"monster/pkg/common/timer"
	"monster/pkg/utils"
)

const (
	PATH_FOUND_FAIL_THRESHOLD    int = 1
	PATH_FOUND_FAIL_WAIT_SECONDS int = 2
)

const (
	ALLY_FLEE_DISTANCE        float32 = 2
	ALLY_FOLLOW_DISTANCE_WALK float32 = 5.5
	ALLY_FOLLOW_DISTANCE_STOP float32 = 5
	ALLY_TELEPORT_DISTANCE    float32 = 40
)

type EntityBehavior struct {
	e                  *Entity
	path               []fpoint.FPoint
	prevTarget         fpoint.FPoint
	pathFoundFailTimer timer.Timer
	pursuePos          fpoint.FPoint
	turnTimer          timer.Timer
}

func newEntityBehavior(modules common.Modules, e *Entity) *EntityBehavior {
	eb := &EntityBehavior{}
	eb.init(modules, e)
	return eb
}

func (this *EntityBehavior) init(modules common.Modules, e *Entity) {
	settings := modules.Settings()

	this.e = e
	this.pathFoundFailTimer = timer.Construct()
	this.pursuePos = fpoint.Construct(-1, -1)
	this.turnTimer = timer.Construct()

	this.pathFoundFailTimer.SetDuration((uint)(settings.Get("max_fps").(int) * PATH_FOUND_FAIL_WAIT_SECONDS))
	this.pathFoundFailTimer.Reset(timer.END)
}

func (this *EntityBehavior) Close() {
}

func (this *EntityBehavior) logic(modules common.Modules, pc gameres.Avatar, camp gameres.CampaignManager, powers gameres.PowerManager) {
	eset := modules.Eset()
	settings := modules.Settings()

	// 作为敌人，已死亡则跳过
	if this.e.stats.GetCorpse() {
		if eset.Get("misc", "corpse_timeout_enabled").(bool) {
			this.e.stats.GetCorpseTimer().Tick()
		}

		return
	}

	// 作为敌人，和英雄是否相遇
	if !this.e.stats.GetHeroAlly() {
		if utils.CalcDist(this.e.stats.GetPos(), pc.GetStats().GetPos()) <= settings.GetEncounterDist() {
			this.e.stats.SetEncountered(true)
		}

		// 作为敌人没有和主角想接触，无逻辑
		if !this.e.stats.GetEncountered() {
			return
		}

	}

	this.doUpKeep(modules, pc, camp, powers)
}

// 更新自己的属性
func (this *EntityBehavior) doUpKeep(modules common.Modules, pc gameres.Avatar, camp gameres.CampaignManager, powers gameres.PowerManager) {
	if this.e.stats.GetHP() > 0 || this.e.stats.GetEffects().GetTriggeredDeath() {
		// TODO
		// 激活被动技能,
		// 激活对应效果
	}

	// 重新计算自己的属性状态
	this.e.stats.Logic(modules, pc, camp)

	if this.e.stats.GetTeleportation() {
		// 缴活了传送技能
		// TODO collider

		this.e.stats.SetPos(this.e.stats.GetTeleportDestination())
		this.e.stats.SetTeleportation(false)
	}
}
