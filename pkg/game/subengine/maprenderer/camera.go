package maprenderer

import (
	"math"
	"math/rand"
	"monster/pkg/common"
	"monster/pkg/common/fpoint"
	"monster/pkg/common/gameres"
	"monster/pkg/common/timer"
	"monster/pkg/utils"
)

type Camera struct {
	pos           fpoint.FPoint
	shake         fpoint.FPoint // 抖动
	shakeTimer    timer.Timer   //
	target        fpoint.FPoint
	prevCamTarget fpoint.FPoint
	prevCamDx     float32
	prevCamDy     float32
	camThreshold  float32 // 最大速度
	shakeStrength int
}

func newCamera(modules common.Modules) *Camera {
	cam := constructCamera(modules)
	return &cam
}

func constructCamera(modules common.Modules) Camera {
	cam := Camera{}
	cam.init(modules)

	return cam
}

func (this *Camera) init(modules common.Modules) gameres.MapCamera {
	eset := modules.Eset()
	settings := modules.Settings()

	this.pos = fpoint.Construct()
	this.shake = fpoint.Construct()
	this.shakeTimer = timer.Construct()
	this.target = fpoint.Construct()
	this.prevCamTarget = fpoint.Construct()
	this.camThreshold = eset.Get("misc", "camera_speed").(float32) / settings.LOGIC_FPS() / 50
	this.shakeStrength = 8

	return this
}

// 逐渐移动摄像头往目标走
func (this *Camera) Logic(modules common.Modules) {
	eset := modules.Eset()

	camDelta := utils.CalcDist(this.pos, this.target) // 距离目标

	// 理论 x 轴的速度
	camDx := utils.CalcDist(fpoint.Construct(this.pos.X, this.target.Y), this.target) / eset.Get("misc", "camera_speed").(float32)

	// 理论 y 轴的速度
	camDy := utils.CalcDist(fpoint.Construct(this.target.X, this.pos.Y), this.target) / eset.Get("misc", "camera_speed").(float32)

	if this.prevCamTarget.X == this.target.X && this.prevCamTarget.Y == this.target.Y {
		// 目标无变动

		if camDelta == 0 || camDelta >= this.camThreshold {
			// 距离太大，加速

			// 保持最大速度
			this.prevCamDx = camDx
			this.prevCamDy = camDy
		} else if camDelta < this.camThreshold {
			// 距离太小，不加速

			if camDx < this.prevCamDx || camDy < this.prevCamDy {
				// 之前已经在运行了，保持之前的速度
				camDx = this.prevCamDx
				camDy = this.prevCamDy
			} else {
				// 否则取一个最小速度
				b := math.Abs((float64)(this.pos.X - this.target.X))
				alpha := math.Acos(b / (float64)(camDelta))
				fastDx := this.camThreshold * (float32)(math.Cos(alpha))
				fastDy := this.camThreshold * (float32)(math.Sin(alpha))
				this.prevCamDx = fastDx / eset.Get("misc", "camera_speed").(float32)
				this.prevCamDy = fastDy / eset.Get("misc", "camera_speed").(float32)
			}

		}

	} else {
		// 目标变动

		// 重置
		this.prevCamTarget = this.target
		this.prevCamDx = 0
		this.prevCamDy = 0
	}

	// 移动
	if this.pos.X < this.target.X {
		this.pos.X += camDx
		if this.pos.X > this.target.X {
			this.pos.X = this.target.X
		}
	} else if this.pos.X > this.target.X {
		this.pos.X -= camDx
		if this.pos.X < this.target.X {
			this.pos.X = this.target.X
		}
	}

	if this.pos.Y < this.target.Y {
		this.pos.Y += camDy
		if this.pos.Y > this.target.Y {
			this.pos.Y = this.target.Y
		}
	} else if this.pos.Y > this.target.Y {
		this.pos.Y -= camDy
		if this.pos.Y < this.target.Y {
			this.pos.Y = this.target.Y
		}
	}

	this.shakeTimer.Tick()

	if this.shakeTimer.IsEnd() {
		// 抖动结束
		this.shake.X = this.pos.X
		this.shake.Y = this.pos.Y
	} else {
		this.shake.X = this.pos.X + float32(rand.Intn(200)%(this.shakeStrength*2)-this.shakeStrength)*0.0078125
		this.shake.Y = this.pos.Y + float32(rand.Intn(200)%(this.shakeStrength*2)-this.shakeStrength)*0.0078125
	}

}

// 设置目标
func (this *Camera) SetTarget(target fpoint.FPoint) {
	this.target = target
}

// 直接移动到某个点
func (this *Camera) WarpTo(target fpoint.FPoint) {
	this.pos = target
	this.shake = target
	this.prevCamTarget = target
	this.shakeTimer.Reset(timer.END)
	this.prevCamDx = 0
	this.prevCamDy = 0
}

func (this *Camera) GetPos() fpoint.FPoint {
	return this.pos

}
