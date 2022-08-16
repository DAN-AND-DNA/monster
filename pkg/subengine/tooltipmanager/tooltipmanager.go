package tooltipmanager

import (
	"monster/pkg/common"
	"monster/pkg/common/define/tooltipmanager"
	"monster/pkg/common/point"
	"monster/pkg/common/tooltipdata"
	"monster/pkg/widget"
)

type TooltipManager struct {
	tip     map[int]common.WidgetTooltip
	tipData []tooltipdata.TooltipData
	pos     []point.Point
	style   []int
	context int
}

func New(settings common.Settings, mods common.ModManager, device common.RenderDevice) *TooltipManager {
	tm := &TooltipManager{
		tip: map[int]common.WidgetTooltip{},
	}

	for i := 0; i < 3; i++ {
		tm.tip[i] = widget.NewTooltip1(settings, mods, device)
		if i > 0 {
			tm.tip[i].SetParent(tm.tip[i-1])
		}
		tm.tipData = append(tm.tipData, tooltipdata.Construct())
		tm.pos = append(tm.pos, point.Construct())
		tm.style = append(tm.style, tooltipmanager.CONTEXT_NONE)

	}

	return tm
}

func (this *TooltipManager) Close() {
	for _, ptr := range this.tip {
		ptr.Close()
	}

	this.tip = nil
}

// 清空全部的文字
func (this *TooltipManager) Clear() {
	for index, _ := range this.tipData {
		this.tipData[index].Clear()
	}
}

func (this *TooltipManager) IsEmpty() bool {
	for _, ptr := range this.tipData {
		if ptr.IsEmpty() {
			return false
		}
	}

	return true
}

func (this *TooltipManager) Push(tipData tooltipdata.TooltipData, pos point.Point, style, tipIndex int) {
	if tipData.IsEmpty() || tipIndex >= 3 {
		return
	}

	this.tipData[tipIndex] = tipData
	this.pos[tipIndex] = pos
	this.style[tipIndex] = style
}

func (this *TooltipManager) Render(settings common.Settings, eset common.EngineSettings, font common.FontEngine, render common.RenderDevice) error {
	if !this.IsEmpty() {
		this.context = tooltipmanager.CONTEXT_MENU
	} else if this.context != tooltipmanager.CONTEXT_MAP {
		this.context = tooltipmanager.CONTEXT_NONE
	}

	for i, ptr := range this.tip {
		err := ptr.Render1(settings, eset, font, render, this.tipData[i], this.pos[i], this.style[i])
		if err != nil {
			return err
		}
	}

	return nil
}
