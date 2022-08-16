package gameslotpreview

import (
	"fmt"
	"monster/pkg/common"
	"monster/pkg/common/gameres"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/utils/parsing"
)

type GameSlotPreview struct {
	stats    gameres.StatBlock
	pos      point.Point
	animSets []common.AnimationSet // 子动画集
	anims    []common.Animation    // 每个子动画集正在运行的动画，和主动画保持一致，名字，帧数等

	animationSet        common.AnimationSet // 主动画集，其实就一个主配置，限制子动画集的帧数与其一致
	activeAnimation     common.Animation    // 正在运行的主动画，其实就是一个状态，配置
	layerReferenceOrder []string
	layerDef            [][]uint // 图层进行覆盖
}

func New(modules common.Modules, gameRes gameres.GameRes) *GameSlotPreview {
	g := &GameSlotPreview{}
	g.Init(modules, gameRes)

	return g
}

func (this *GameSlotPreview) Init(modules common.Modules, gameRes gameres.GameRes) gameres.GameSlotPreview {
	settings := modules.Settings()
	mods := modules.Mods()
	render := modules.Render()
	anim := modules.Anim()
	mresf := modules.Resf()

	gresf := gameRes.Resf()

	var err error
	this.stats = gresf.New("statblock").(gameres.StatBlock).Init(modules, gresf)
	this.pos = point.Construct()
	anim.IncreaseCount("animations/hero.txt") // +1

	this.animationSet, err = anim.GetAnimationSet(settings, mods, render, mresf, "animations/hero.txt") // 最高层的一个配置，不包含具体的图片文件
	if err != nil {
		panic(err)
	}

	this.activeAnimation = this.animationSet.GetAnimation("") // 设置当前动画

	err = this.loadLayerDefinitions(modules)
	if err != nil {
		panic(err)
	}
	return this
}

func (this *GameSlotPreview) Close(modules common.Modules) {
	anim := modules.Anim()

	anim.DecreaseCount("animations/hero.txt") // -1 // 清理主动画集
	this.activeAnimation.Close()
	this.activeAnimation = nil

	for i, ptr := range this.animSets {
		if ptr != nil {
			anim.DecreaseCount(ptr.GetName())
		}
		if this.anims[i] != nil {
			this.anims[i].Close()
		}
	}

	this.animSets = nil
	this.anims = nil
	anim.CleanUp() // 清理无用的动画集
}

// 加载方向的图层定义和顺序
func (this *GameSlotPreview) loadLayerDefinitions(modules common.Modules) error {
	mods := modules.Mods()

	infile := fileparser.New()

	this.layerDef = nil
	this.layerReferenceOrder = nil

	this.layerDef = make([][]uint, 8)

	err := infile.Open("engine/hero_layers.txt", true, mods)
	if err != nil {
		return err
	}
	defer infile.Close()

	for infile.Next(mods) {
		switch infile.Key() {
		case "layer":
			var first, strVal, layer string
			first, strVal = parsing.PopFirstString(infile.Val(), "")
			dir := parsing.ToDirection(first)
			if dir > 7 {
				return fmt.Errorf("GameSlotPreview: Hero layer direction must be in range [0,7]\n")
			}

			layer, strVal = parsing.PopFirstString(strVal, "")
			for layer != "" {
				refPos := (uint)(0)
				found := false

				// 名字去重
				for ; refPos < (uint)(len(this.layerReferenceOrder)); refPos++ {
					if layer == this.layerReferenceOrder[refPos] {
						found = true
						break
					}
				}

				// 添加名字
				if !found {
					this.layerReferenceOrder = append(this.layerReferenceOrder, layer)
				}

				// 添加序号
				this.layerDef[dir] = append(this.layerDef[dir], refPos)

				layer, strVal = parsing.PopFirstString(strVal, "")
			}
		default:
			return fmt.Errorf("GameSlotPreview: '%s' is not a valid key.\n", infile.Key())
		}
	}

	return nil
}

func (this *GameSlotPreview) SetStatBlock(stats gameres.StatBlock) {
	this.stats = stats
}

// 设置主动画和子动画
func (this *GameSlotPreview) SetAnimation(name string) {
	if name == this.activeAnimation.GetName() {
		return
	}

	if this.activeAnimation != nil {
		this.activeAnimation.Close()
		this.activeAnimation = nil
	}

	this.activeAnimation = this.animationSet.GetAnimation(name)

	for i, ptr := range this.animSets {

		if this.anims[i] != nil {
			this.anims[i].Close()
			this.anims[i] = nil
		}

		if ptr != nil {
			this.anims[i] = ptr.GetAnimation(name)
		}
	}
}

// 加载动画图片
func (this *GameSlotPreview) LoadGraphics(modules common.Modules, imgGfx []string) error {
	settings := modules.Settings()
	mods := modules.Mods()
	render := modules.Render()
	anim := modules.Anim()
	mresf := modules.Resf()

	if this.stats == nil {
		return nil
	}

	// 加载每个方向的图层定义和顺序
	this.loadLayerDefinitions(modules)

	for i, ptr := range this.animSets {
		if ptr != nil {
			anim.DecreaseCount(ptr.GetName())
		}

		this.anims[i].Close()
	}

	// 清空状态
	this.animSets = nil
	this.anims = nil

	for _, val := range imgGfx {
		if val != "" {

			// 创建新的动画集
			name := "animations/avatar/" + this.stats.GetGfxBase() + "/" + val + ".txt"
			anim.IncreaseCount(name)
			newAnimSet, err := anim.GetAnimationSet(settings, mods, render, mresf, name)
			if err != nil {
				return err
			}

			newAnimSet.SetParent(this.animationSet) // 限制后续的动画集帧数与主动画集配置相同
			this.animSets = append(this.animSets, newAnimSet)
			this.anims = append(this.anims, newAnimSet.GetAnimation(this.activeAnimation.GetName()))
			this.SetAnimation("stance")

			// 同步主配置动画配置给每个子动画
			if !this.anims[len(this.anims)-1].SyncTo(this.activeAnimation) {
				return fmt.Errorf("GameSlotPreview: Error syncing animation in '%s' to 'animations/hero.txt'.", newAnimSet.GetName())
			}
		} else {
			this.animSets = append(this.animSets, nil)
			this.anims = append(this.anims, nil)
		}
	}

	// 清理无用的动画集
	anim.CleanUp()
	this.SetAnimation("stance")

	return nil
}

func (this *GameSlotPreview) Logic() {
	this.activeAnimation.AdvanceFrame()
	for _, ptr := range this.anims {
		if ptr != nil {
			ptr.AdvanceFrame()
		}
	}
}

func (this *GameSlotPreview) SetPos(pos point.Point) {
	this.pos = pos
}

func (this *GameSlotPreview) AddRenders(modules common.Modules, r []common.Renderable) []common.Renderable {
	if this.stats == nil {
		return r
	}

	//fmt.Println(this.stats.GetDirection())
	//fmt.Println(this.layerDef)
	// 不同方向，逐层覆盖的顺序不同
	for i, index := range this.layerDef[this.stats.GetDirection()] {
		if this.anims[index] != nil {
			// 获取当前帧的信息
			ren := this.anims[index].GetCurrentFrame(modules, (int)(this.stats.GetDirection()))
			ren.SetPrio((uint64)(i + 1)) // 设置优先级
			r = append(r, ren)
		}
	}

	return r
}

func (this *GameSlotPreview) Render(modules common.Modules) error {
	render := modules.Render()

	var r []common.Renderable
	var err error

	r = this.AddRenders(modules, r)

	for _, ptr := range r {
		if ptr.GetImage() != nil {
			dest := rect.Construct(this.pos.X-ptr.GetOffset().X, this.pos.Y-ptr.GetOffset().Y)
			err = render.Render1(ptr, dest)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (this *GameSlotPreview) GetLayerReferenceOrder() []string {
	tmp := make([]string, len(this.layerReferenceOrder))
	for i, val := range this.layerReferenceOrder {
		tmp[i] = val
	}

	return tmp
}
