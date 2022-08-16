package base

import (
	"fmt"
	"monster/pkg/common"
	"monster/pkg/common/fpoint"
	"monster/pkg/common/gameres"
	"monster/pkg/common/gameres/statblock"
)

type Entity struct {
	sprites common.Image // 外部钩子
	// TODO sound
	activeAnimation common.Animation
	animationSet    common.AnimationSet
	stats           gameres.StatBlock
	typeFilename    string
	behavior        *EntityBehavior
}

func ConstructEntity(modules common.Modules, gresf gameres.Factory) Entity {
	entity := Entity{}
	entity.init(modules, gresf)

	return entity
}

func (this *Entity) init(modules common.Modules, gresf gameres.Factory) {
	this.behavior = newEntityBehavior(modules, this)
	this.stats = gresf.New("statblock").(gameres.StatBlock).Init(modules, gresf)
}

func (this *Entity) clear() {
	if this.activeAnimation != nil {
		this.activeAnimation.Close()
		this.activeAnimation = nil
	}

	if this.behavior != nil {
		this.behavior.Close()
		this.behavior = nil
	}
}

func (this *Entity) Close(modules common.Modules, impl gameres.Entity) {
	impl.Clear(modules)

	this.clear()
}

func (this *Entity) Logic(modules common.Modules, pc gameres.Avatar, camp gameres.CampaignManager, powers gameres.PowerManager) {
	this.behavior.logic(modules, pc, camp, powers)
}

func (this *Entity) moveFromOffendingTile() {
	// TODO
}

func (this *Entity) Move(modules common.Modules, mapr gameres.MapRenderer) bool {
	this.moveFromOffendingTile()

	if this.stats.GetEffects().GetKnockbackSpeed() != 0 {
		// 击退无法移动
		return false
	}

	if this.stats.GetEffects().GetStun() || this.stats.GetEffects().GetSpeed() == 0 {
		// 被晕住或者效果加成的速度为0
		return false
	}

	if this.stats.GetChargeSpeed() != 0 {
		return false
	}

	// 方向，单帧移动距离
	speed := this.stats.GetSpeed() * statblock.SPEED_MULTIPLIER[this.stats.GetDirection()] * (float32)(this.stats.GetEffects().GetSpeed()) / 100
	dx := speed * statblock.DIRECTION_DELTA_X[this.stats.GetDirection()]
	dy := speed * statblock.DIRECTION_DELTA_Y[this.stats.GetDirection()]

	newX, newY, fullmove := mapr.GetCollider().Move(modules, this.stats.GetPos().X, this.stats.GetPos().Y, dx, dy, this.stats.GetMovementType(), mapr.GetCollider().GetCollideType(this.stats.GetHero()))

	this.stats.SetPos(fpoint.Construct(newX, newY))

	return fullmove
}

func (this *Entity) GetStats() gameres.StatBlock {
	return this.stats
}

func (this *Entity) SetSprites(val common.Image) {
	this.sprites = val
}

func (this *Entity) SetAnimationSet(val common.AnimationSet) {
	this.animationSet = val
}

func (this *Entity) GetAnimationSet() common.AnimationSet {
	return this.animationSet
}

func (this *Entity) SetActiveAnimation(val common.Animation) {
	this.activeAnimation = val
}

func (this *Entity) GetActiveAnimation() common.Animation {
	return this.activeAnimation
}

func (this *Entity) SetAnimation(name string) bool {
	if this.activeAnimation != nil && this.activeAnimation.GetName() == name {
		return true
	}

	this.activeAnimation.Close()
	this.activeAnimation = this.animationSet.GetAnimation(name)

	if this.activeAnimation == nil {
		panic(fmt.Sprintf("Entity::setAnimation(%s): not found", name))
	}

	return this.activeAnimation == nil
}
