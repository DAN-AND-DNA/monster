package widget

import (
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define/fontengine"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/common/tooltipdata"
	"monster/pkg/filesystem/logfile"
	"monster/pkg/utils"
)

type Tooltip struct {
	bounds     rect.Rect // 边框
	parent     common.WidgetTooltip
	background common.Image
	dataBuf    tooltipdata.TooltipData
	spriteBuf  common.Sprite
}

func NewTooltip(modules common.Modules) *Tooltip {
	tt := &Tooltip{}
	tt.Init(modules)
	return tt
}

func NewTooltip1(settings common.Settings, mods common.ModManager, render common.RenderDevice) *Tooltip {
	var err error

	tt := &Tooltip{}
	tt.background, err = render.LoadImage(settings, mods, "images/menus/tooltips.png")
	if err != nil && !utils.IsNotExist(err) {
		panic(err)
	}

	return tt
}

func (this *Tooltip) Init(modules common.Modules) common.WidgetTooltip {
	settings := modules.Settings()
	render := modules.Render()
	mods := modules.Mods()

	var err error
	this.background, err = render.LoadImage(settings, mods, "images/menus/tooltips.png")
	if err != nil && !utils.IsNotExist(err) {
		panic(err)
	}

	return this
}

func (this *Tooltip) Close() {
	if this.background != nil {
		this.background.UnRef()
	}

	if this.spriteBuf != nil {
		this.spriteBuf.Close()
	}
}

// 计算tip的显示位置
func (this *Tooltip) CalcPosition(s common.Settings, eset common.EngineSettings, style int, pos point.Point, size point.Point) point.Point {
	tipPos := point.Construct()

	// toplabel 固定居中
	if style == tooltipdata.STYLE_TOPLABEL {
		tipPos.X = pos.X - size.X/2
		tipPos.Y = pos.Y - eset.Get("tooltips", "tooltip_offset").(int)

	} else if style == tooltipdata.STYLE_FLOAT {
		// 浮动
		if pos.X < s.GetViewWHalf() && pos.Y < s.GetViewHHalf() {
			// 左上 upper left
			if this.parent != nil {
				tipPos.X = this.parent.GetBounds().X + this.parent.GetBounds().W
			} else {
				tipPos.X = pos.X + eset.Get("tooltips", "tooltip_offset").(int)
			}

			tipPos.Y = pos.Y + eset.Get("tooltips", "tooltip_offset").(int)
		} else if pos.X >= s.GetViewWHalf() && pos.Y < s.GetViewHHalf() {
			// 右上 upper right
			if this.parent != nil {
				tipPos.X = this.parent.GetBounds().X - size.X
			} else {
				tipPos.X = pos.X - eset.Get("tooltips", "tooltip_offset").(int) - size.X
			}

			tipPos.Y = pos.Y + eset.Get("tooltips", "tooltip_offset").(int)
		} else if pos.X < s.GetViewWHalf() && pos.Y >= s.GetViewHHalf() {
			// 左下 lower left
			if this.parent != nil {
				tipPos.X = this.parent.GetBounds().X + this.parent.GetBounds().W
			} else {
				tipPos.X = pos.X + eset.Get("tooltips", "tooltip_offset").(int)
			}

			tipPos.Y = pos.Y - eset.Get("tooltips", "tooltip_offset").(int) - size.Y
		} else if pos.X >= s.GetViewWHalf() && pos.Y >= s.GetViewHHalf() {
			// 右下 lower right
			if this.parent != nil {
				tipPos.X = this.parent.GetBounds().X - size.X
			} else {
				tipPos.X = pos.X - eset.Get("tooltips", "tooltip_offset").(int) - size.X
			}

			tipPos.Y = pos.Y - eset.Get("tooltips", "tooltip_offset").(int) - size.Y
		}

		if tipPos.X+size.X > s.GetViewW() && this.parent != nil {
			tipPos.X = s.GetViewW() - size.X
		}

		if tipPos.Y+size.Y > s.GetViewH() {
			tipPos.Y = s.GetViewH() - size.Y
		}

		if tipPos.X < 0 && this.parent != nil {
			tipPos.X = 0
		}

		if tipPos.Y < 0 {
			tipPos.Y = 0
		}
	}

	return tipPos
}

// 设置精灵大小和显示位置并缓存
func (this *Tooltip) PreRender(settings common.Settings, eset common.EngineSettings, font common.FontEngine, renderDevice common.RenderDevice, tip tooltipdata.TooltipData, pos point.Point, style int) error {

	// 创建缓存
	if this.spriteBuf == nil || !tip.Compare(this.dataBuf) {
		if err := this.CreateBuffer(eset, font, renderDevice, tip); err != nil {
			return err
		}
	}

	var size point.Point
	var err error
	size.X, err = this.spriteBuf.GetGraphicsWidth()
	if err != nil {
		return err
	}

	size.Y, err = this.spriteBuf.GetGraphicsHeight()
	if err != nil {
		return err
	}

	// 计算tip的显示位置
	tipPos := this.CalcPosition(settings, eset, style, pos, size)
	this.spriteBuf.SetDestFromPoint(tipPos)
	this.bounds.X = tipPos.X
	this.bounds.Y = tipPos.Y
	this.bounds.W = size.X
	this.bounds.H = size.Y

	return nil
}

func (this *Tooltip) Render(modules common.Modules, tip tooltipdata.TooltipData, pos point.Point, style int) error {
	settings := modules.Settings()
	eset := modules.Eset()
	font := modules.Font()
	render := modules.Render()

	return this.Render1(settings, eset, font, render, tip, pos, style)
}

func (this *Tooltip) Render1(settings common.Settings, eset common.EngineSettings, font common.FontEngine, render common.RenderDevice, tip tooltipdata.TooltipData, pos point.Point, style int) error {
	if tip.IsEmpty() {
		// 没内容
		return nil
	}

	err := this.PreRender(settings, eset, font, render, tip, pos, style)
	if err != nil {
		return err
	}

	err = render.Render(this.spriteBuf)
	if err != nil {
		return err
	}

	return nil
}

// 创建精灵来缓存文字
func (this *Tooltip) CreateBuffer(eset common.EngineSettings, font common.FontEngine, renderDevice common.RenderDevice, tip tooltipdata.TooltipData) error {
	fulltext := ""
	lines := tip.Lines()
	lenLines := len(lines)
	for index, line := range lines {
		fulltext += line
		if index != lenLines-1 {
			fulltext += "\n"
		}
	}

	font.SetFont("font_regular")

	// 计算多行宽和高
	size := font.CalcSize(fulltext, eset.Get("tooltips", "tooltip_width").(int))

	// 清理之前的缓存
	if this.spriteBuf != nil {
		this.spriteBuf.Close()
		this.spriteBuf = nil
	}

	graphics, err := renderDevice.CreateImage(size.X+2*eset.Get("tooltips", "tooltip_margin").(int), size.Y+2*eset.Get("tooltips", "tooltip_margin").(int))
	if err != nil {
		logfile.LogError("WidgetTooltip: Could not create tooltip buffer.")
		return err
	}
	defer graphics.UnRef()

	if this.background == nil {
		err := graphics.FillWithColor(color.Construct(0, 0, 0, 255))
		if err != nil {
			return err
		}

	} else {
		var src, dest rect.Rect
		gw, err := graphics.GetWidth()
		if err != nil {
			return err
		}

		gh, err := graphics.GetHeight()
		if err != nil {
			return err
		}

		bkw, err := this.background.GetWidth()
		if err != nil {
			return err
		}

		bkh, err := this.background.GetHeight()
		if err != nil {
			return err
		}

		// 把背景绘制到graphics上
		// 左上
		src.X = 0
		src.Y = 0
		src.W = gw - eset.Get("tooltips", "tooltip_background_border").(int)
		src.H = gh - eset.Get("tooltips", "tooltip_background_border").(int)
		dest.X = 0
		dest.Y = 0
		dest, err = renderDevice.RenderToImage(this.background, src, graphics, dest)
		if err != nil {
			return err
		}

		// 右
		src.X = bkw - eset.Get("tooltips", "tooltip_background_border").(int)
		src.Y = 0
		src.W = eset.Get("tooltips", "tooltip_background_border").(int)
		src.H = gh - eset.Get("tooltips", "tooltip_background_border").(int)
		dest.X = gw - eset.Get("tooltips", "tooltip_background_border").(int)
		dest.Y = 0
		dest, err = renderDevice.RenderToImage(this.background, src, graphics, dest)
		if err != nil {
			return err
		}

		// 底部
		src.X = 0
		src.Y = bkh - eset.Get("tooltips", "tooltip_background_border").(int)
		src.W = gw - eset.Get("tooltips", "tooltip_background_border").(int)
		src.H = eset.Get("tooltips", "tooltip_background_border").(int)
		dest.X = 0
		dest.Y = gh - eset.Get("tooltips", "tooltip_background_border").(int)
		dest, err = renderDevice.RenderToImage(this.background, src, graphics, dest)
		if err != nil {
			return err
		}

		// 右下
		src.X = bkw - eset.Get("tooltips", "tooltip_background_border").(int)
		src.Y = bkh - eset.Get("tooltips", "tooltip_background_border").(int)
		src.W = eset.Get("tooltips", "tooltip_background_border").(int)
		src.H = eset.Get("tooltips", "tooltip_background_border").(int)
		dest.X = gw - eset.Get("tooltips", "tooltip_background_border").(int)
		dest.Y = gh - eset.Get("tooltips", "tooltip_background_border").(int)
		dest, err = renderDevice.RenderToImage(this.background, src, graphics, dest)
		if err != nil {
			return err
		}
	}

	cursorY := eset.Get("tooltips", "tooltip_margin").(int)

	colors := tip.Colors()
	for index, line := range lines {
		if this.background != nil {
			err = font.RenderShadowed(renderDevice, line, eset.Get("tooltips", "tooltip_margin").(int), cursorY, fontengine.JUSTIFY_LEFT, graphics, size.X, colors[index])
			if err != nil {
				return err
			}

		} else {
			err = font.Render(renderDevice, line, eset.Get("tooltips", "tooltip_margin").(int), cursorY, fontengine.JUSTIFY_LEFT, graphics, size.X, colors[index])
			if err != nil {
				return err
			}
		}

		cursorY = font.CursorY()
	}

	this.spriteBuf, err = graphics.CreateSprite()
	if err != nil {
		return err
	}

	this.dataBuf = tip
	return nil
}

func (this *Tooltip) SetParent(t common.WidgetTooltip) {
	this.parent = t
}

func (this *Tooltip) GetBounds() rect.Rect {
	return this.bounds
}
