package effect

import (
	"monster/pkg/common/color"
)

type Def struct {
	Id              string
	Type            int // 造成的效果类型，速度，伤害等
	Name            string
	Icon            int    // 状态栏显示的图标
	Animation       string // 效果动画文件
	CanStack        bool   // 是否可累加
	MaxStacks       int    // 最大累加数，-1代表无限
	GroupStack      bool   // 是否累加时使用一个图标代替
	RenderAbove     bool   // 是否效果绘制在最上面
	ColorMod        color.Color
	AlphaMod        uint8
	AttackSpeedAnim string //  类型为攻击速度时，使用该动画文件
}

func ConstructDef() Def {
	def := Def{}
	def.init()

	return def
}

func (this *Def) init() {
	this.Type = NONE
	this.CanStack = true
	this.MaxStacks = -1
	this.ColorMod = color.Construct(255, 255, 255)
	this.AlphaMod = 255

}
