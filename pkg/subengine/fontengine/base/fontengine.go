package base

import (
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define/fontengine"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/filesystem/logfile"
	"monster/pkg/utils/parsing"
	"strings"
)

const (
	COLOR_COUNT = 17
)

var (
	white = color.Construct(255, 255, 255)
	black = color.Construct(0, 0, 0)
)

type FontEngine struct {
	fontColors map[int]color.Color
	cursorY    int
}

func ConstructFontEngine(mods common.ModManager) FontEngine {
	f := FontEngine{
		fontColors: map[int]color.Color{},
		cursorY:    0,
	}

	f.fontColors[fontengine.COLOR_WHITE] = white
	f.fontColors[fontengine.COLOR_BLACK] = black
	f.fontColors[fontengine.COLOR_MENU_NORMAL] = white
	f.fontColors[fontengine.COLOR_MENU_BONUS] = white
	f.fontColors[fontengine.COLOR_MENU_PENALTY] = white
	f.fontColors[fontengine.COLOR_WIDGET_NORMAL] = white
	f.fontColors[fontengine.COLOR_WIDGET_DISABLED] = white
	f.fontColors[fontengine.COLOR_COMBAT_GIVEDMG] = white
	f.fontColors[fontengine.COLOR_COMBAT_TAKEDMG] = white
	f.fontColors[fontengine.COLOR_COMBAT_CRIT] = white
	f.fontColors[fontengine.COLOR_COMBAT_BUFF] = white
	f.fontColors[fontengine.COLOR_COMBAT_MISS] = white
	f.fontColors[fontengine.COLOR_REQUIREMENTS_NOT_MET] = white
	f.fontColors[fontengine.COLOR_ITEM_BONUS] = white
	f.fontColors[fontengine.COLOR_ITEM_PENALTY] = white
	f.fontColors[fontengine.COLOR_ITEM_FLAVOR] = white
	f.fontColors[fontengine.COLOR_HARDCORE_NAME] = white

	infile := fileparser.New()
	if err := infile.Open("engine/font_colors.txt", true, mods); err != nil {
		panic(err)
	}
	defer infile.Close()

	for infile.Next(mods) {
		if intKey := stringToFontColor(infile.Key()); intKey != -1 {
			f.fontColors[intKey] = parsing.ToRGB(infile.Val())
		} else {
			panic(common.Err_bad_key_in_fontengine)
		}
	}

	return f
}

func (this *FontEngine) Close() {
}

func stringToFontColor(val string) int {
	switch val {
	case "menu_normal":
		return fontengine.COLOR_MENU_NORMAL
	case "menu_bonus":
		return fontengine.COLOR_MENU_BONUS
	case "menu_penalty":
		return fontengine.COLOR_MENU_PENALTY
	case "widget_normal":
		return fontengine.COLOR_WIDGET_NORMAL
	case "widget_disabled":
		return fontengine.COLOR_WIDGET_DISABLED
	case "combat_givedmg":
		return fontengine.COLOR_COMBAT_GIVEDMG
	case "combat_takedmg":
		return fontengine.COLOR_COMBAT_TAKEDMG
	case "combat_crit":
		return fontengine.COLOR_COMBAT_CRIT
	case "combat_buff":
		return fontengine.COLOR_COMBAT_BUFF
	case "combat_miss":
		return fontengine.COLOR_COMBAT_MISS
	case "requirements_not_met":
		return fontengine.COLOR_REQUIREMENTS_NOT_MET
	case "item_bonus":
		return fontengine.COLOR_ITEM_BONUS
	case "item_penalty":
		return fontengine.COLOR_ITEM_PENALTY
	case "item_flavor":
		return fontengine.COLOR_ITEM_FLAVOR
	case "hardcore_color_name":
		return fontengine.COLOR_HARDCORE_NAME
	}

	return -1
}

func (this *FontEngine) GetColor(key int) color.Color {
	if val, ok := this.fontColors[key]; ok {
		return val
	}

	panic(common.Err_bad_key_in_fontengine)
}

// ????????????????????????
func (this *FontEngine) CalcSize(impl common.FontEngine, textWithNewlines string, width int) point.Point {
	newline := "\n"

	text := textWithNewlines
	checkNewline := strings.Index(text, newline)

	// ???????????????
	if checkNewline != -1 {
		p1 := this.CalcSize(impl, textWithNewlines[0:checkNewline], width)
		p2 := this.CalcSize(impl, textWithNewlines[checkNewline+1:], width)
		p3 := point.Construct()

		// ?????????
		if p1.X > p2.X {
			p3.X = p1.X
		} else {
			p3.X = p2.X
		}

		p3.Y = p1.Y + p2.Y // height
		return p3
	}

	height := 0
	maxWidth := 0
	space := ([]byte(" "))[0]
	fulltext := text + " "
	nextWord := ""
	cursor := 0
	builder := "" // ??????2???????????????????????????????????????????????????
	builderPrev := ""
	longToken := ""

	// ?????????cursor??????????????????????????????
	nextWord, cursor = getNextToken(fulltext, cursor, space)

	for {
		if cursor == -1 {
			// ?????????
			break
		}

		builder += nextWord

		if impl.CalcWidth(builder) > width {
			// ??????????????????

			if builderPrev != "" {
				height += impl.GetLineHeight()
				if impl.CalcWidth(builderPrev) > maxWidth {
					maxWidth = impl.CalcWidth(builderPrev)
				}
			}

			builder = ""                                                      // ????????????
			longToken, nextWord = this.popTokenByWidth(impl, nextWord, width) // ?????????nextword?????????????????????

			for {
				if longToken == "" {
					break
				}

				if impl.CalcWidth(nextWord) > maxWidth {
					maxWidth = impl.CalcWidth(nextWord) // ??????????????????
				}

				height += impl.GetLineHeight() // ?????????+1
				nextWord = longToken
				longToken, nextWord = this.popTokenByWidth(impl, nextWord, width) // ?????????nextword?????????????????????
				if longToken == nextWord {
					// ??????????????????????????????????????????,??????
					break
				}
			}

			builder += nextWord + " "
			builderPrev = builder

		} else {
			// ??????????????????????????????????????????????????????
			builder += " "
			builderPrev = builder
		}

		// ??????????????????????????????
		nextWord, cursor = getNextToken(fulltext, cursor, space)
	}

	// ??????????????????" "??????
	builder = strings.TrimSpace(builder)
	if builder != "" {
		height += impl.GetLineHeight()
	}

	if impl.CalcWidth(builder) > maxWidth {
		maxWidth = impl.CalcWidth(builder)
	}

	if textWithNewlines == " " {
		height += impl.GetLineHeight()
	}

	size := point.Construct()
	size.X = maxWidth
	size.Y = height
	return size
}

// ????????????????????????????????????
func (this *FontEngine) Position(impl common.FontEngine, text string, x, y, justify int) rect.Rect {
	var destRect rect.Rect

	switch justify {
	case fontengine.JUSTIFY_LEFT:
		destRect.X = x
		destRect.Y = y
	case fontengine.JUSTIFY_RIGHT:
		destRect.X = x - impl.CalcWidth(text)
		destRect.Y = y
	case fontengine.JUSTIFY_CENTER:
		destRect.X = x - impl.CalcWidth(text)/2
		destRect.Y = y
	default:
		logfile.LogError("FontEngine::position() given unhandled 'justify=%d', assuming left", justify)
		destRect.X = x
		destRect.Y = y
	}

	return destRect
}

// ???????????????0??????????????????????????????width??????????????????
func (this *FontEngine) Render(impl common.FontEngine, renderDevice common.RenderDevice, text string, x, y, justify int, target common.Image, width int, color color.Color) error {
	if width == 0 {
		// a width of 0 means we won't try to wrap text
		// ??????????????????????????????
		err := impl.RenderInternal(renderDevice, text, x, y, justify, target, color)
		if err != nil {
			return err
		}
		return nil
	}

	fulltext := text + " " // ??????????????????
	nextWord := ""
	cursor := 0
	this.cursorY = y
	space := ([]byte(" "))[0]
	builder := ""
	builderPrev := ""
	longToken := ""

	// ??????????????????????????????????????????
	nextWord, cursor = getNextToken(fulltext, cursor, space)
	for {
		if cursor == -1 {
			// ?????????
			break
		}

		builder += nextWord
		if impl.CalcWidth(builder) > width {

			if builderPrev != "" {
				impl.RenderInternal(renderDevice, builderPrev, x, this.cursorY, justify, target, color)
				this.cursorY += impl.GetLineHeight()
			}

			builder = ""
			longToken, nextWord = this.popTokenByWidth(impl, nextWord, width) //nextWord????????????????????????????????????

			for {
				if longToken == "" {
					break
				}

				impl.RenderInternal(renderDevice, nextWord, x, this.cursorY, justify, target, color)
				this.cursorY += impl.GetLineHeight()
				nextWord = longToken
				longToken, nextWord = this.popTokenByWidth(impl, nextWord, width)
				if longToken == nextWord { // < width
					break
				}
			}

			builder += nextWord + " "
			builderPrev = builder
		} else {
			builder += " "
			builderPrev = builder
		}

		// ??????????????????????????????????????????
		nextWord, cursor = getNextToken(fulltext, cursor, space)
	}

	impl.RenderInternal(renderDevice, builder, x, this.cursorY, justify, target, color)
	this.cursorY += impl.GetLineHeight()

	return nil
}

// ????????????????????????width????????????????????????
func (this *FontEngine) RenderShadowed(impl common.FontEngine, renderDevice common.RenderDevice, text string, x, y, justify int, target common.Image, width int, color color.Color) error {
	tmpColor := this.GetColor(fontengine.COLOR_BLACK)
	err := this.Render(impl, renderDevice, text, x+1, y+1, justify, target, width, tmpColor)
	if err != nil {
		return err
	}

	err = this.Render(impl, renderDevice, text, x, y, justify, target, width, color)
	if err != nil {
		return err
	}

	return nil
}

// ??????strVal???cursor????????????????????????separator??????????????????
func getNextToken(strVal string, cursor int, separator byte) (string, int) {
	if cursor >= len(strVal) {
		return "", -1
	}

	var bytesSeparator []byte
	bytesSeparator = append(bytesSeparator, separator)

	strVal = strVal[cursor:]
	seppos := strings.Index(strVal, string(bytesSeparator))
	if seppos == -1 {
		return "", -1
	}

	outs := strVal[:seppos]
	return outs, cursor + seppos + 1
}

/*
 * Fits a string, "text", to a pixel "width".
 * The original string is mutated to fit within the width.
 * Returns a string that is the remainder of the original string that could not fit in the width.
 */

// ?????? ??????????????????????????????????????????
func (this *FontEngine) popTokenByWidth(impl common.FontEngine, text string, width int) (string, string) {
	runetext := []rune(text)

	newLength := 0
	for index := 1; index <= len(runetext); index++ {
		if impl.CalcWidth((string)(runetext[:index])) > width {
			break
		}

		newLength = index
	}

	if newLength > 0 {
		return string(runetext[newLength:]), string(runetext[:newLength])
	}

	return text, text
}

func (this *FontEngine) CursorY() int {
	return this.cursorY
}
