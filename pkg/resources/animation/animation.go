package animation

import (
	"fmt"
	"math"
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define/animation"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
)

type Animation struct {
	name             string
	type1            int
	sprite           *Media
	blendMode        uint8
	alphaMod         uint8
	colorMod         color.Color
	numberFrames     uint16 // 播放帧数
	curFrame         uint16 // 播放列表中的第几帧
	curFrameIndex    uint16 // 播放列表的索引，以speed前进去播放
	curFrameDuration uint16
	curFrameIndexF   float32
	maxKinds         uint16

	/*
		additionalData:
			当type为BACK_FORTH 1为前进 -1为后退 0为在最后
			当type为LOOPED 为循环次数
			当type为PLAY_ONCE 无用
	*/
	additionalData int16
	timesPlayed    uint // 动画播放了多少次
	// frame 数据
	gfx                  []common.Image // 外部指针
	gfxPos               []rect.Rect
	renderOffset         []point.Point
	frames               []uint16 // 帧的播放列表，以时间分帧，
	activeFrames         []int16
	activeFrameTriggered bool
	elapsedFrames        uint16
	frameCount           uint16
	speed                float32
}

func NewAnimation(name, type1 string, sprite *Media, blendMode, alphaMod uint8, colorMod color.Color) *Animation {
	an := &Animation{}
	an.Init(name, type1, sprite, blendMode, alphaMod, colorMod)

	return an
}

func (this *Animation) Init(name, type1 string, sprite *Media, blendMode, alphaMod uint8, colorMod color.Color) common.Animation {
	this.name = name

	switch type1 {
	case "play_once":
		this.type1 = animation.ANIMTYPE_PLAY_ONCE
	case "back_forth":
		this.type1 = animation.ANIMTYPE_BACK_FORTH
	case "looped":
		this.type1 = animation.ANIMTYPE_LOOPED
	default:
		this.type1 = animation.ANIMTYPE_NONE
	}

	this.sprite = sprite
	this.blendMode = blendMode
	this.alphaMod = alphaMod
	this.colorMod = colorMod
	this.speed = 1.0

	if this.type1 == animation.ANIMTYPE_NONE {
		panic(fmt.Sprintf("Animation: Type %s is unknown\n", type1))
	}

	return this
}

func (this *Animation) DeepCopy() common.Animation {
	obj := *this

	obj.gfx = make([]common.Image, len(this.gfx))
	for index, val := range this.gfx {
		obj.gfx[index] = val
	}

	obj.gfxPos = make([]rect.Rect, len(this.gfxPos))
	for index, val := range this.gfxPos {
		obj.gfxPos[index] = val
	}

	obj.renderOffset = make([]point.Point, len(this.renderOffset))
	for index, val := range this.renderOffset {
		obj.renderOffset[index] = val
	}

	obj.frames = make([]uint16, len(this.frames))
	for index, val := range this.frames {
		obj.frames[index] = val
	}

	obj.activeFrames = make([]int16, len(this.activeFrames))
	for index, val := range this.activeFrames {
		obj.activeFrames[index] = val
	}

	return &obj
}

func (this *Animation) Close() {
}

func (this *Animation) GetFrameCount() uint16 {
	return this.frameCount
}

// 创建资源和初始化
// 一般8行即maxKinds代表一个方向的动画
// renderSize 代表网格大小
// frameCount 代表一个方向帧数
func (this *Animation) Setup(frameCount, duration, maxKinds uint16) {
	this.frameCount = frameCount
	this.frames = nil

	if frameCount > 0 && duration%frameCount == 0 {
		// 时间能等分
		divided := duration / frameCount          // 每帧的持续时间
		for i := uint16(0); i < frameCount; i++ { // 3
			for j := uint16(0); j < divided; j++ { // 4
				this.frames = append(this.frames, i) // 相当于[0000,1111,2222]
			}
		}
	} else {
		// Bresenham 算法来补偿
		x0, y0 := (uint16)(0), (uint16)(0)
		x1, y1 := duration-1, frameCount-1
		dx, dy := x1-x0, y1-y0 // 起点和终点的差

		d := 2*dy - dx
		this.frames = append(this.frames, y0)

		x, y := x0+1, y0

		for x <= x1 {
			if d > 0 {
				y++
				this.frames = append(this.frames, y)
				d = d + (2*dy - 2*dx)
			} else {
				this.frames = append(this.frames, y)
				d = d + 2*dy
			}
			x++
		}
	}

	if len(this.frames) != 0 {
		this.numberFrames = this.frames[len(this.frames)-1] + 1
	}

	if this.type1 == animation.ANIMTYPE_PLAY_ONCE {
		this.additionalData = 0
	} else if this.type1 == animation.ANIMTYPE_LOOPED {
		this.additionalData = 0
	} else if this.type1 == animation.ANIMTYPE_BACK_FORTH {
		this.numberFrames = 2 * this.numberFrames
		this.additionalData = 1
	}

	// 清零
	this.curFrame = 0
	this.curFrameIndex = 0
	this.curFrameIndexF = 0
	this.maxKinds = maxKinds
	this.timesPlayed = 0

	// 取中间作为激活帧(可视)
	this.activeFrames = append(this.activeFrames, (int16)(this.numberFrames-1)/2)

	//总数 =  每个动画方向 * 每个方向的帧数
	i := maxKinds * frameCount

	// 扩容
	oldGfx := this.gfx
	this.gfx = make([]common.Image, i)
	for index, val := range oldGfx {
		this.gfx[index] = val
	}

	oldGfxPos := this.gfxPos
	this.gfxPos = make([]rect.Rect, i)
	for index, val := range oldGfxPos {
		this.gfxPos[index] = val
	}

	oldRenderOffset := this.renderOffset
	this.renderOffset = make([]point.Point, i)
	for index, val := range oldRenderOffset {
		this.renderOffset[index] = val
	}
}

// 从网格创建动画
// maxKinds代表一个方向的动画，从上到下，每一行是其中一个帧
// renderSize 代表网格大小和渲染大小
// position代表从每一行的哪个位置开始是帧动画
// frameCount代表每一行多少帧
// renderOffser从图片转化到地图坐标偏移
func (this *Animation) SetupUncompressed(renderSize, renderOffset point.Point, position, frameCount, duration, maxKinds uint16) {

	// 创建资源和初始化
	this.Setup(frameCount, duration, maxKinds)
	// 一个方向的总帧数
	for i := (uint16)(0); i < frameCount; i++ {
		baseIndex := maxKinds * i

		// 每个方向
		for kind := (uint16)(0); kind < maxKinds; kind++ {
			this.gfx[baseIndex+kind], _ = this.sprite.GetImageFromKey("")
			this.gfxPos[baseIndex+kind].X = renderSize.X * (int)(position+i)
			this.gfxPos[baseIndex+kind].Y = renderSize.Y * (int)(kind)
			this.gfxPos[baseIndex+kind].W = renderSize.X
			this.gfxPos[baseIndex+kind].H = renderSize.Y
			this.renderOffset[baseIndex+kind].X = renderOffset.X
			this.renderOffset[baseIndex+kind].Y = renderOffset.Y
		}
	}
}

// 添加帧
// kind: 某个方向
// index: 某个方向里的索引
func (this *Animation) AddFrame(index, kind uint16, rect rect.Rect, renderOffset point.Point, key string) bool {
	if index >= (uint16)(len(this.gfx))/this.maxKinds || kind > this.maxKinds-1 {
		// 范围外的
		return false
	}

	i := this.maxKinds*index + kind // 计算索引
	this.gfx[i], _ = this.sprite.GetImageFromKey(key)
	this.gfxPos[i] = rect
	this.renderOffset[i] = renderOffset
	return true
}

// 前进帧，主要是时间
func (this *Animation) AdvanceFrame() {

	if len(this.frames) == 0 {
		this.curFrameIndex = 0
		this.curFrameIndexF = 0
		this.timesPlayed++
		return
	}

	lastBaseIndex := (uint16)(len(this.frames) - 1) // 最后的索引号

	switch this.type1 {
	case animation.ANIMTYPE_PLAY_ONCE:
		if this.curFrameIndex < lastBaseIndex {
			// 未到最后正常递增
			this.curFrameIndexF += this.speed
			this.curFrameIndex = (uint16)(this.curFrameIndexF)
		} else {
			this.timesPlayed = 1
		}

	case animation.ANIMTYPE_LOOPED:
		if this.curFrameIndex < lastBaseIndex {
			this.curFrameIndexF += this.speed
			this.curFrameIndex = (uint16)(this.curFrameIndexF)
		} else {
			// 到最后，在从头开始
			this.curFrameIndex = 0
			this.curFrameIndexF = 0
			this.timesPlayed++
		}
	case animation.ANIMTYPE_BACK_FORTH:
		if this.additionalData == 1 {
			// 前进
			if this.curFrameIndex < lastBaseIndex {
				this.curFrameIndexF += this.speed
				this.curFrameIndex = (uint16)(this.curFrameIndexF)
			} else {
				// 到尾部了
				this.additionalData = -1
			}

		} else if this.additionalData == -1 {
			// 后退
			if this.curFrameIndex > 0 {
				this.curFrameIndexF -= this.speed
				this.curFrameIndex = (uint16)(this.curFrameIndexF)
			} else {
				// 到头部了
				this.additionalData = 1
				this.timesPlayed++
			}

		}
	default:
		// do nothing
	}

	this.curFrameIndex = (uint16)(math.Max(0, (float64)(this.curFrameIndex)))
	if this.curFrameIndex > lastBaseIndex {
		this.curFrameIndex = lastBaseIndex
	}

	if this.curFrame != this.frames[this.curFrameIndex] {
		// 是
		this.elapsedFrames++
	}

	// 更新播放列表的帧
	this.curFrame = this.frames[this.curFrameIndex]
}

// kind: 某一个方向
func (this *Animation) GetCurrentFrame(modules common.Modules, kind int) common.Renderable {
	mresf := modules.Resf()

	r := mresf.New("renderable").(common.Renderable)

	if len(this.frames) != 0 {

		// 当前所处的网格图片信息
		index := (int)(this.maxKinds*this.frames[this.curFrameIndex]) + kind

		r.SetSrc(this.gfxPos[index])
		r.SetOffset(this.renderOffset[index])
		r.SetImage(this.gfx[index])
		r.SetBlendMode(this.blendMode)
		r.SetColorMod(this.colorMod)
		r.SetAlphaMod(this.alphaMod)
	}

	return r
}

// 重置
func (this *Animation) Rest() {
	this.curFrame = 0
	this.curFrameIndex = 0
	this.curFrameIndexF = 0
	this.timesPlayed = 0
	this.additionalData = 1 // 前
	this.elapsedFrames = 0
	this.activeFrameTriggered = false
}

func (this *Animation) GetCurFrame() uint16 {
	return this.curFrame
}

func (this *Animation) GetCurFrameIndex() uint16 {
	return this.curFrameIndex
}

func (this *Animation) GetCurFrameIndexF() float32 {
	return this.curFrameIndexF
}

func (this *Animation) GetAdditionalData() int16 {
	return this.additionalData
}

func (this *Animation) GetElapsedFrames() uint16 {
	return this.elapsedFrames
}

func (this *Animation) SyncTo(other common.Animation) bool {
	this.curFrame = other.GetCurFrame()
	this.curFrameIndex = other.GetCurFrameIndex()
	this.curFrameIndexF = other.GetCurFrameIndexF()
	this.timesPlayed = other.GetTimesPlayed()
	this.additionalData = other.GetAdditionalData()
	this.elapsedFrames = other.GetElapsedFrames()

	if this.curFrameIndex >= (uint16)(len(this.frames)) {
		if len(this.frames) == 0 {
			fmt.Printf("Animation: '%s' animation has no frames, but current frame index is greater than 0.\n", this.name)
			this.curFrameIndex = 0
			this.curFrameIndexF = 0
			return false
		} else {
			fmt.Printf("Animation: Current frame index (%d) was larger than the last frame index (%d) when syncing '%s' animation.\n", this.curFrameIndex, len(this.frames), this.name)
			this.curFrameIndex = (uint16)(len(this.frames) - 1)
			this.curFrameIndexF = (float32)(this.curFrameIndex)
			return false
		}
	}

	return true
}

// 设置激活帧
func (this *Animation) SetActiveFrames(activeFrames []int16) {
	// 只有一个-1代表全部设为 激活帧
	if len(activeFrames) == 1 && activeFrames[0] == -1 {
		this.activeFrames = make([]int16, this.numberFrames)
		for i := uint16(0); i < this.numberFrames; i++ {
			this.activeFrames[i] = (int16)(i)
		}
	} else {
		this.activeFrames = activeFrames
	}

	// 是否存在最后一帧
	haveLastFrame := false
	for _, val := range this.activeFrames {
		if (uint16)(val) == this.numberFrames-1 {
			haveLastFrame = true
		}
	}

	// 清理序号非法的，控制在0到最后一帧的范围内
	for i, val := range this.activeFrames {
		if (uint16)(val) >= this.numberFrames {
			if haveLastFrame {
				// 删除i
				if i == len(this.activeFrames)-1 {
					// 最后一个
					this.activeFrames = this.activeFrames[:len(this.activeFrames)-1]
				} else {
					// 中间一个
					old := this.activeFrames
					this.activeFrames = make([]int16, len(old)-1)
					j := 0
					for _, val1 := range old[:i] {
						this.activeFrames[j] = val1
						j++
					}

					// 跳过i
					old = old[i+1:]
					for _, val1 := range old {
						this.activeFrames[j] = val1
						j++
					}
				}
			} else {
				// 过大的直接变成最后一帧，等下次清理其他过大的帧
				this.activeFrames[i] = int16(this.numberFrames - 1)
				haveLastFrame = true
			}
		}
	}
}

// 是否当前索引是在第一帧
func (this *Animation) IsFirstFrame() bool {
	return this.curFrameIndex == 0 && float32(this.curFrameIndex) == this.curFrameIndexF
}

// 获得对应帧的序号
func (this *Animation) GetLastFrameIndex(frame int16) uint16 {
	if len(this.frames) == 0 || frame < 0 {
		return 0
	}

	if this.type1 == animation.ANIMTYPE_BACK_FORTH && this.additionalData == -1 {
		// 后退
		for index, val := range this.frames {
			if val == (uint16)(frame) {
				return (uint16)(index)
			}
		}
		return 0
	}

	for i := len(this.frames); i > 0; i-- {
		if this.frames[i-1] == (uint16)(frame) {
			return (uint16)(i - 1)
		}
	}

	return uint16(len(this.frames) - 1)
}

// 判断当前的索引是不是最后的帧
func (this *Animation) IsLastFrame() bool {
	return this.curFrameIndex == this.GetLastFrameIndex((int16)(this.numberFrames)-1)
}

// 判断当前的索引是不是最后第二帧
func (this *Animation) IsSecondLastFrame() bool {
	return this.curFrameIndex == this.GetLastFrameIndex((int16)(this.numberFrames)-2)
}

// 是否当前是在激活帧
func (this *Animation) IsActiveFrame() bool {
	if this.type1 == animation.ANIMTYPE_BACK_FORTH {
		// 来回

		for _, val := range this.activeFrames {
			if val == (int16)(this.elapsedFrames) {
				// found
				return this.curFrameIndex == this.GetLastFrameIndex((int16)(this.curFrame)) &&
					(float32)(this.curFrameIndex) == this.curFrameIndexF
			}
		}
	} else {
		// 其他

		found := false
		for _, val := range this.activeFrames {
			if val == (int16)(this.curFrame) {
				// found
				found = true
			}
		}

		if found &&
			this.curFrameIndex == this.GetLastFrameIndex((int16)(this.curFrame)) &&
			(float32)(this.curFrameIndex) == this.curFrameIndexF {

			if this.type1 == animation.ANIMTYPE_PLAY_ONCE {
				this.activeFrameTriggered = true
			}

			return true
		}
	}

	// 最后一帧
	return this.IsLastFrame() && this.type1 == animation.ANIMTYPE_PLAY_ONCE && !this.activeFrameTriggered && len(this.activeFrames) != 0
}

func (this *Animation) GetTimesPlayed() uint {
	return this.timesPlayed
}

func (this *Animation) SetTimesPlayed(val uint) {
	this.timesPlayed = val
}

func (this *Animation) GetName() string {
	return this.name
}

// 动画时间
func (this *Animation) GetDuration() int {
	return (int)(float32(len(this.frames)) / this.speed)
}

func (this *Animation) IsCompleted() bool {
	return this.type1 == animation.ANIMTYPE_PLAY_ONCE && this.timesPlayed > 0
}

func (this *Animation) SetSpeed(val float32) {
	this.speed = val / 100
}
