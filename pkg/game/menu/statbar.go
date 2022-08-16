package menu

import (
	"fmt"
	"math"
	"monster/pkg/common"
	"monster/pkg/common/define/game/menu/statbar"
	"monster/pkg/common/gameres"
	"monster/pkg/common/labelinfo"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/common/timer"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/game/base"
	"monster/pkg/utils/parsing"
)

type StatBar struct {
	base.Menu

	bar              common.Sprite
	label            common.WidgetLabel
	statMin          uint64
	statCur          uint64
	statCurPrev      uint64
	statMax          uint64
	barPos           rect.Rect           // 血条、蓝条等位置，相对整个组件
	textPos          labelinfo.LabelInfo // 文字位置
	orientation      int
	customTextPos    bool // 自定义文字位置
	barGfx           string
	barGfxBackground string
	type1            uint16
	timeout          timer.Timer // 当血条满了，过了对应的时间就进行隐藏，0为不隐藏
	barFillOffset    point.Point // 用于填充的图片偏移位置，相对 barPos而言
	barFillSize      point.Point // 用于填充的图片大小 宽和高
}

func NewStatBar(modules common.Modules, type1 uint16) *StatBar {
	sb := &StatBar{}
	sb.Init(modules, type1)

	return sb
}

func (this *StatBar) Init(modules common.Modules, type1 uint16) gameres.MenuStatBar {
	widgetf := modules.Widgetf()
	mods := modules.Mods()
	settings := modules.Settings()
	render := modules.Render()

	// base
	this.Menu = base.ConstructMenu(modules)

	// self
	this.label = widgetf.New("label").(common.WidgetLabel).Init(modules)
	this.orientation = statbar.HORIZONTAL
	this.type1 = type1
	this.timeout = timer.Construct()
	this.barFillOffset = point.Construct()
	this.barFillSize = point.Construct(-1, -1)

	typeFilename := ""
	switch type1 {
	case statbar.TYPE_HP:
		typeFilename = "hp"
	case statbar.TYPE_MP:
		typeFilename = "mp"
	case statbar.TYPE_XP:
		typeFilename = "xp"
	default:
		panic("bad type for menu stat bar")
	}

	infile := fileparser.New()

	err := infile.Open("menus/"+typeFilename+".txt", true, mods)
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	for infile.Next(mods) {
		key := infile.Key()
		val := infile.Val()

		// 显示的窗口位置，对齐等信息
		if this.Menu.ParseMenuKey(key, val) {
			continue
		}

		switch key {
		case "bar_pos":
			this.barPos = parsing.ToRect(val)
		case "text_pos":
			this.customTextPos = true
			this.textPos = parsing.PopLabelInfo(val)
		case "orientation":
			rawOri := parsing.ToBool(val)
			if !rawOri {
				this.orientation = statbar.HORIZONTAL
			} else {
				this.orientation = statbar.VERTICAL
			}
		case "bar_gfx":
			this.barGfx = val
		case "bar_gfx_background":
			this.barGfxBackground = val
		case "hide_timeout":
			maxFps := settings.Get("max_fps").(int)
			this.timeout.SetDuration((uint)(parsing.ToDuration(val, maxFps)))
		case "bar_fill_offset":
			this.barFillOffset = parsing.ToPoint(val)
		case "bar_fill_size":
			this.barFillSize = parsing.ToPoint(val)
		default:
			panic(fmt.Sprintf("MenuStatBar: '%s' is not a valid key.\n", key))
		}
	}

	if this.barFillSize.X == -1 || this.barFillSize.Y == -1 {
		this.barFillSize.X = this.barPos.W
		this.barFillSize.Y = this.barPos.H
	}

	// 加载图片
	if this.barGfxBackground != "" {
		err := this.Menu.SetBackground(modules, this.barGfxBackground)
		if err != nil {
			panic(err)
		}
	}

	if this.barGfx != "" {
		graphics, err := render.LoadImage(settings, mods, this.barGfx)
		if err != nil {
			panic(err)
		}
		defer graphics.UnRef()

		this.bar, err = graphics.CreateSprite()
		if err != nil {
			panic(err)
		}
	}

	err = this.Menu.Align(modules)
	if err != nil {
		panic(err)
	}

	return this
}

func (this *StatBar) Clear() {
	if this.bar != nil {
		this.bar.Close()
		this.bar = nil
	}

	if this.label != nil {
		this.label.Close()
		this.label = nil
	}
}

func (this *StatBar) Close() {
	this.Menu.Close(this)
}

func (this *StatBar) Logic(modules common.Modules, pc gameres.Avatar, powers gameres.PowerManager) error {
	return nil
}

func (this *StatBar) Render(modules common.Modules) error {
	render := modules.Render()

	if this.disappear() {
		return nil
	}

	src := rect.Construct()
	dest := rect.Construct()
	barDest := this.barPos

	// 显示位置
	barDest.X = this.barPos.X + this.GetWindowArea().X
	barDest.Y = this.barPos.Y + this.GetWindowArea().Y

	dest.X = barDest.X
	dest.Y = barDest.Y
	src.X = 0
	src.Y = 0
	src.W = this.barPos.W
	src.H = this.barPos.H
	this.SetBackgroundClip(src)
	this.SetBackgroundDest(dest)

	// 背景
	err := this.Menu.Render(modules)
	if err != nil {
		return err
	}

	//

	statCurClamped := math.Min((float64)(this.statCur), (float64)(this.statMax))                         // 当前值
	normalizedCur := (uint64)(statCurClamped - math.Min(statCurClamped, (float64)(this.statMin)))        // 当前值和最小值的差
	normalizedMax := this.statMax - (uint64)(math.Min((float64)(this.statMin), (float64)(this.statMax))) // 最大值和最小值的差

	if this.orientation == statbar.HORIZONTAL {
		// 水平
		barLength := (uint64)(0)
		if normalizedMax != 0 {
			// 宽 = 图片的宽 * 占比
			barLength = normalizedCur * (uint64)(this.barFillSize.X) / normalizedMax
		}

		if barLength == 0 && normalizedCur > 0 {
			barLength = 1
		}

		src.X = 0
		src.Y = 0
		src.W = (int)(barLength)
		src.H = this.barFillSize.Y
		dest.X = barDest.X + this.barFillOffset.X
		dest.Y = barDest.Y + this.barFillOffset.Y

	} else if this.orientation == statbar.VERTICAL {
		barLength := (uint64)(0)
		if normalizedMax != 0 {
			// 宽 = 图片的宽 * 占比
			barLength = normalizedCur * (uint64)(this.barFillSize.Y) / normalizedMax
		}

		if barLength == 0 && normalizedCur > 0 {
			barLength = 1
		}

		src.X = 0
		src.Y = this.barFillSize.Y - (int)(barLength)
		src.W = this.barFillSize.X
		src.H = (int)(barLength)
		dest.X = barDest.X + this.barFillOffset.X
		dest.Y = barDest.Y + this.barFillOffset.Y + src.Y

	}

	if this.bar != nil {
		err := this.bar.SetClipFromRect(src)
		if err != nil {
			return err
		}

		this.bar.SetDestFromRect(dest)
		err = render.Render(this.bar)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *StatBar) Update(statMin, statCur, statMax uint64) {
	this.statCurPrev = this.statCur
	this.statMin = statMin
	this.statCur = statCur
	this.statMax = statMax
}

func (this *StatBar) disappear() bool {
	// TODO
	return false
}
