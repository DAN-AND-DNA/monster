package animation

import (
	"errors"
	"fmt"
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define/renderable"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/utils/parsing"
	"sort"
)

// 把动画打包在这个结构体
type Set struct {
	name             string // 文件名
	imageFile        string
	defaultAnimation common.Animation
	parent           common.AnimationSet
	animations       []*Animation
	sprite           *Media
}

/*
func NewSet(settings common.Settings, mods common.ModManager, render common.RenderDevice, name string) *Set {
	set := &Set{}
	set.Init(settings, mods, render, name)

	return set
}
*/

func (this *Set) Init(settings common.Settings, mods common.ModManager, render common.RenderDevice, name string) common.AnimationSet {
	this.name = name
	this.sprite = NewMedia()
	defaultAnim := NewAnimation("default",
		"play_once",
		this.sprite,
		renderable.BLEND_NORMAL,
		255,
		color.Construct(255, 255, 255))

	defaultAnim.SetupUncompressed(point.Construct(), point.Construct(), 0, 1, 0, 8)
	this.defaultAnimation = defaultAnim

	// 加载name对应的配置，加载配置文件里指定的图片
	err := this.load(settings, mods, render)
	if err != nil {
		panic(err)
	}

	return this
}

func (this *Set) Close() {
	for _, ptr := range this.animations {
		if ptr != nil {
			ptr.Close()
		}
	}

	this.animations = nil
	if this.defaultAnimation != nil {
		this.defaultAnimation.Close()
		this.defaultAnimation = nil
	}

	if this.sprite != nil {
		this.sprite.UnRef()
		this.sprite.Close()
		this.sprite = nil
	}
}

func (this *Set) SetParent(parent common.AnimationSet) {
	this.parent = parent
}

func (this *Set) GetName() string {
	return this.name
}

// 获得帧数
func (this *Set) GetAnimationFrameCount(name string) uint16 {
	for _, ptr := range this.animations {
		if ptr.GetName() == name {
			return ptr.GetFrameCount() // 一个方向的动画帧数
		}
	}

	return 0
}

// 加载name对应的配置，加载配置文件里指定的图片
func (this *Set) load(settings common.Settings, mods common.ModManager, render common.RenderDevice) error {
	infile := fileparser.New()

	if this.name == "" {
		return nil
	}

	err := infile.Open(this.name, true, mods)
	if err != nil {
		return err
	}
	defer infile.Close()

	var name, type1, startingAnimation string
	var position, frameCount, duration, parentAnimFrameCount uint16
	var blendMode, alphaMod uint8 = renderable.BLEND_NORMAL, 255
	colorMod := color.Construct(255, 255, 255)
	renderSize := point.Construct()
	renderOffset := point.Construct()
	var firstSection, compressedLoading bool = true, false
	var newAnim *Animation
	var activeFrames []int16

	for infile.Next(mods) {
		//  如果上节不存在帧数据，则本节补充上节一个空动画数据
		if infile.IsNewSection() {

			// 上节开启压缩并分配，本届则一开始不分配，直到有帧数据需要压缩加载
			// 上节不开启压缩加载，则本届一开始就分配作为上节内容
			// 第二节开始
			if !firstSection && !compressedLoading {

				// 创建新动画
				a := NewAnimation(name, type1, this.sprite, blendMode, alphaMod, colorMod)

				// 分配空间, 沿用上节的数据
				a.SetupUncompressed(renderSize, renderOffset, position, frameCount, duration, 8)

				if len(activeFrames) != 0 {
					a.SetActiveFrames(activeFrames)
				}
				activeFrames = nil
				this.animations = append(this.animations, a)
			}

			firstSection = false
			compressedLoading = false

			if this.parent != nil {
				// 获得父级帧数
				parentAnimFrameCount = this.parent.GetAnimationFrameCount(infile.GetSection())
			}
		}

		if infile.GetSection() == "" {
			// 顶部信息
			switch infile.Key() {
			case "image":
				// mod 里的图片文件名
				imgFilename, strVal := parsing.PopFirstString(infile.Val(), "")
				imgId, strVal := parsing.PopFirstString(strVal, "")
				err := this.sprite.LoadImage(settings, mods, render, imgFilename, imgId)
				if err != nil {
					return err
				}
			case "render_size":
				strVal := ""
				renderSize.X, strVal = parsing.PopFirstInt(infile.Val(), "")
				renderSize.Y, strVal = parsing.PopFirstInt(strVal, "")
			case "render_offset":
				strVal := ""
				renderOffset.X, strVal = parsing.PopFirstInt(infile.Val(), "")
				renderOffset.Y, strVal = parsing.PopFirstInt(strVal, "")
			case "blend_mode":
				strBlendMode, _ := parsing.PopFirstString(infile.Val(), "")
				if strBlendMode == "normal" {
					blendMode = renderable.BLEND_NORMAL
				} else if strBlendMode == "add" {
					blendMode = renderable.BLEND_ADD
				} else {
					fmt.Printf("AnimationSet: '%s' is not a valid blend mode.\n", infile.Key())
					blendMode = renderable.BLEND_NORMAL
				}
			case "alpha_mod":
				rawAlphaMod, _ := parsing.PopFirstInt(infile.Val(), "")
				alphaMod = (uint8)(rawAlphaMod)
			case "color_mod":
				colorMod = parsing.ToRGB(infile.Val())
			default:
				return errors.New(fmt.Sprintf("AnimationSet: '%s' is not a valid key.\n", infile.Key()))
			}
		} else {
			// 一个方向或动作
			switch infile.Key() {
			case "position":
				// 第一帧的位置
				rawPosition, _ := parsing.PopFirstInt(infile.Val(), "")
				position = (uint16)(rawPosition)
			case "frames":
				// 该动作的帧数
				frameCount = (uint16)(parsing.ToInt(infile.Val(), 0))
				if this.parent != nil && frameCount != parentAnimFrameCount {
					fmt.Printf("AnimationSet: Frame count %d != %d for matching animation in %s\n", frameCount, parentAnimFrameCount, this.parent.GetName())
					frameCount = parentAnimFrameCount
				}

			case "duration":
				duration = (uint16)(parsing.ToDuration(infile.Val(), settings.Get("max_fps").(int)))
			case "type":
				type1 = infile.Val()
			case "active_frame":
				activeFrames = nil
				nv, strVal := parsing.PopFirstString(infile.Val(), "")
				if nv == "all" {
					// 只有一个-1代表全部设为激活帧
					activeFrames = append(activeFrames, -1)
				} else {
					for nv != "" {
						activeFrames = append(activeFrames, (int16)(parsing.ToInt(nv, 0)))
						nv, strVal = parsing.PopFirstString(strVal, "")
					}

					// 去重
					tmp := map[int16]struct{}{}
					for _, val := range activeFrames {
						tmp[val] = struct{}{}
					}
					activeFrames = nil
					for key, _ := range tmp {
						activeFrames = append(activeFrames, key)
					}

					// 排序
					sort.Slice(activeFrames, func(i, j int) bool { return activeFrames[i] < activeFrames[j] })

				}
			case "frame":
				// 有帧数据，则本节开启压缩加载并分配（第一节和后续节）
				if compressedLoading == false {
					newAnim = NewAnimation(name, type1, this.sprite, blendMode, alphaMod, colorMod)
					newAnim.Setup(frameCount, duration, 8) // 分配空间
					if len(activeFrames) != 0 {
						newAnim.SetActiveFrames(activeFrames)
					}

					activeFrames = nil
					this.animations = append(this.animations, newAnim)

					// 该节后续帧数据都保持
					compressedLoading = true
				}

				r := rect.Construct()
				offset := point.Construct()
				rawIndex, strVal := parsing.PopFirstInt(infile.Val(), "")
				index := (uint16)(rawIndex)
				rawDirection, strVal := parsing.PopFirstInt(strVal, "")
				direction := (uint16)(rawDirection)
				r.X, strVal = parsing.PopFirstInt(strVal, "")
				r.Y, strVal = parsing.PopFirstInt(strVal, "")
				r.W, strVal = parsing.PopFirstInt(strVal, "")
				r.H, strVal = parsing.PopFirstInt(strVal, "")
				offset.X, strVal = parsing.PopFirstInt(strVal, "")
				offset.Y, strVal = parsing.PopFirstInt(strVal, "")
				key := strVal
				// 不同的动作作为key
				if !newAnim.AddFrame(index, direction, r, offset, key) {
					fmt.Printf("AnimationSet: Frame index (%u) is out of bounds [0, %hu].\n", index, frameCount)
				}

			default:
				return errors.New(fmt.Sprintf("AnimationSet: '%s' is not a valid key.\n", infile.Key()))
			}

		}

		if name == "" {
			startingAnimation = infile.GetSection()
		}
		name = infile.GetSection()
	}

	if !compressedLoading {
		// 若最后第二节未启动压缩，则最后一节补充了上节内容，但如果最后一节也没有帧数据，则这里添加最后一节
		a := NewAnimation(name, type1, this.sprite, blendMode, alphaMod, colorMod)
		a.SetupUncompressed(renderSize, renderOffset, position, frameCount, duration, 8)
		if len(activeFrames) != 0 {
			a.SetActiveFrames(activeFrames)
		}
		activeFrames = nil
		this.animations = append(this.animations, a)
	}

	if startingAnimation != "" {
		a := this.GetAnimation(startingAnimation)
		if this.defaultAnimation != nil {
			this.defaultAnimation.Close()
		}
		this.defaultAnimation = a
	}

	return nil
}

func (this *Set) GetAnimation(name string) common.Animation {
	for _, ptr := range this.animations {
		if ptr.GetName() == name {
			return ptr.DeepCopy()
		}
	}

	return this.defaultAnimation.DeepCopy()
}
