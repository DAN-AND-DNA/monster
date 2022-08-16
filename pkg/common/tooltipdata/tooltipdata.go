package tooltipdata

import (
	"bufio"
	"monster/pkg/common/color"
	"strings"
)

const (
	STYLE_FLOAT = iota
	STYLE_TOPLABEL
)

type TooltipData struct {
	lines  []string
	colors []color.Color
}

func Construct() TooltipData {
	return TooltipData{}
}

func (this *TooltipData) AddColorText(text string, color color.Color) error {
	if len(text) == 0 {
		this.lines = append(this.lines, " ")
		this.colors = append(this.colors, color)
		return nil
	}

	scanner := bufio.NewScanner(strings.NewReader(text))

	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			this.lines = append(this.lines, " ")
		} else {
			this.lines = append(this.lines, scanner.Text())
		}
		this.colors = append(this.colors, color)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

/*
func (this *TooltipData) AddText(font common.FontEngine, text string) error {
	return this.AddColorText(text, font.GetColor(fontengine.COLOR_WIDGET_NORMAL))
}
*/

func (this *TooltipData) Clear() {
	this.lines = []string{}
	this.colors = []color.Color{}
}

func (this *TooltipData) IsEmpty() bool {
	return len(this.lines) == 0
}

func (this *TooltipData) CompareFirstLine(text string) bool {
	if len(this.lines) == 0 {
		return false
	}

	if this.lines[0] == text {
		return true
	}

	return false
}

func (this *TooltipData) Compare(tip TooltipData) bool {
	tlines := tip.Lines()
	tcolors := tip.Colors()
	if len(this.lines) != len(tlines) {
		return false
	}

	for index, line := range this.lines {
		if line != tlines[index] ||
			this.colors[index].R != tcolors[index].R ||
			this.colors[index].G != tcolors[index].G ||
			this.colors[index].B != tcolors[index].B ||
			this.colors[index].A != tcolors[index].A {
			return false
		}
	}

	return true
}

func (this *TooltipData) Lines() []string {
	return this.lines
}

func (this *TooltipData) Colors() []color.Color {
	return this.colors
}
