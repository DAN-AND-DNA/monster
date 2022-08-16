package sdlfont

import (
	"fmt"
	"monster/pkg/allocs"
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"monster/pkg/filesystem/fileparser"
	"monster/pkg/filesystem/logfile"
	"monster/pkg/subengine/fontengine/base"
	"monster/pkg/utils/parsing"
	"unicode/utf8"

	"github.com/veandco/go-sdl2/ttf"
)

type FontEngine struct {
	base.FontEngine

	fontStyles []FontStyle
	activeFont *FontStyle
}

func NewFontEngine(settings common.Settings, mods common.ModManager) *FontEngine {
	fe := &FontEngine{}
	_ = (common.FontEngine)(fe)

	// base
	fe.FontEngine = base.ConstructFontEngine(mods)

	// self
	// 初始化 sdl ttf
	if !ttf.WasInit() {
		fmt.Println("ttf init")
		if err := ttf.Init(); err != nil {
			logfile.LogError("SDLFontEngine: TTF_Init: %s", ttf.GetError())
			logfile.LogErrorDialog("SDLFontEngine: TTF_Init: %s", ttf.GetError())
			panic(err)
		}
	}

	// 加载字体
	infile := fileparser.New()
	if err := infile.Open("engine/font_settings.txt", true, mods); err != nil {
		panic(err)
	}
	defer infile.Close()

	for infile.Next(mods) {
		if infile.IsNewSection() && infile.GetSection() == "font" {
			fe.fontStyles = append(fe.fontStyles, ConstructFontStyle())
		} else if infile.GetSection() == "font_fallback" {
			if infile.IsNewSection() {
				logfile.LogError("FontEngine: Support for 'font_fallback' has been removed.")
			}

			continue
		}

		if len(fe.fontStyles) == 0 {
			continue
		}

		ptrStyle := &(fe.fontStyles[len(fe.fontStyles)-1])

		switch infile.Key() {
		case "id":
			ptrStyle.Name = infile.Val()
		case "style":
			strVal := infile.Val()
			lang := ""
			lang, strVal = parsing.PopFirstString(strVal, "")
			if (lang == "default" && ptrStyle.Path == "") || lang == settings.Get("language").(string) {
				blend := ""
				ptrStyle.Path, strVal = parsing.PopFirstString(strVal, "")
				ptrStyle.PtSize, strVal = parsing.PopFirstInt(strVal, "")
				blend, strVal = parsing.PopFirstString(strVal, "")
				ptrStyle.Blend = parsing.ToBool(blend)
				loc, err := mods.Locate(settings, "fonts/"+ptrStyle.Path)
				if err != nil {
					panic(err)
				}

				if ptrStyle.ttfont != nil {
					//ptrStyle.ttfont.Close()
					allocs.Delete(ptrStyle.ttfont)
					ptrStyle.ttfont = nil
				}

				//ptrStyle.ttfont, err = ttf.OpenFont(loc, ptrStyle.PtSize)
				ptrStyle.ttfont, err = allocs.TtfOpenFont(loc, ptrStyle.PtSize)
				if err != nil {
					logfile.LogError("FontEngine: TTF_OpenFont: %s", ttf.GetError())
					panic(err)
				}

				lineSkip := ptrStyle.ttfont.LineSkip()
				ptrStyle.LineHeight = lineSkip
				ptrStyle.FontHeight = lineSkip
			}
		}
	}

	fe.SetFont("font_regular")
	if !fe.IsActiveFontValid() {
		logfile.LogError("FontEngine: Unable to determine default font!")
		logfile.LogErrorDialog("FontEngine: Unable to determine default font!")
	}

	return fe
}

func (this *FontEngine) Close() {

	fmt.Println(len(this.fontStyles))

	for _, fontStyle := range this.fontStyles {
		//fmt.Println("close fs")
		//fontStyle.ttfont.Close()
		allocs.Delete(fontStyle.ttfont)
	}

	ttf.Quit()

	this.FontEngine.Close()
}

func (this *FontEngine) SetFont(font string) {
	for index, fontStyle := range this.fontStyles {
		if fontStyle.ttfont != nil && fontStyle.Name == font {
			this.activeFont = &(this.fontStyles[index])
			return
		}
	}

	for index, fontStyle := range this.fontStyles {
		if fontStyle.ttfont != nil {
			logfile.LogError("FontEngine: Invalid font '%s'. Falling back to '%s'.", font, fontStyle.Name)
			this.activeFont = &(this.fontStyles[index])
			return
		}
	}

	logfile.LogError("FontEngine: Invalid font '%s'. No fallback available.", font)

}

func (this *FontEngine) IsActiveFontValid() bool {
	return this.activeFont != nil && this.activeFont.ttfont != nil
}

func (this *FontEngine) Render(renderDevice common.RenderDevice, text string, x, y, justify int, target common.Image, width int, color color.Color) error {
	return this.FontEngine.Render(this, renderDevice, text, x, y, justify, target, width, color)
}

func (this *FontEngine) Position(text string, x, y, justify int) rect.Rect {
	return this.FontEngine.Position(this, text, x, y, justify)
}

func (this *FontEngine) CalcSize(textWithNewlines string, width int) point.Point {
	return this.FontEngine.CalcSize(this, textWithNewlines, width)
}

func (this *FontEngine) RenderShadowed(renderDevice common.RenderDevice, text string, x, y, justify int, target common.Image, width int, color color.Color) error {
	return this.FontEngine.RenderShadowed(this, renderDevice, text, x, y, justify, target, width, color)
}

// 计算字符串的像素宽度
func (this *FontEngine) CalcWidth(text string) int {
	if !this.IsActiveFontValid() {
		return 1
	}

	w, _, err := this.activeFont.ttfont.SizeUTF8(text)
	if err != nil {
		panic(err)
	}
	return w
}

func (this *FontEngine) GetLineHeight() int {
	if !this.IsActiveFontValid() {
		return 1
	}

	return this.activeFont.LineHeight
}

func (this *FontEngine) GetFontHeight() int {
	if !this.IsActiveFontValid() {
		return 1
	}

	return this.activeFont.FontHeight
}

// x,y指的是target的位置
func (this *FontEngine) RenderInternal(renderDevice common.RenderDevice, text string, x, y, justify int, target common.Image, color color.Color) error {
	if !this.IsActiveFontValid() || text == "" {
		return nil
	}

	// 计算字符串渲染的起始位置
	destRect := this.Position(text, x, y, justify)

	// 以字符串创建图片
	graphics, err := renderDevice.RenderTextToImage(this.activeFont, text, color, this.activeFont.Blend)
	if err != nil {
		return err
	}
	defer graphics.UnRef()

	// 把上面的图片渲染到目标图片上
	if target != nil {
		var clip rect.Rect
		clip.W, err = graphics.GetWidth()
		if err != nil {
			return err
		}
		clip.H, err = graphics.GetHeight()
		if err != nil {
			return err
		}

		// 把字符串图片渲染到target，以图片大小作为源
		destRect, err = renderDevice.RenderToImage(graphics, clip, target, destRect)
		if err != nil {
			return err
		}

	} else {
		tempSprite, err := graphics.CreateSprite()
		if err != nil {
			return err
		}
		defer tempSprite.Close()

		tempSprite.SetDestFromRect(destRect)
		err = renderDevice.Render(tempSprite)
		if err != nil {
			return err
		}
	}

	return nil
}

// leftPos作为 utf-8的字符位置, 0为全部
func (this *FontEngine) TrimTextToWidth(text string, width int, useEllipsis bool, leftPos int) string {
	if width >= this.CalcWidth(text) {
		return text
	}

	if !utf8.ValidString(text[leftPos:]) {
		panic("no a runes")
	}

	runeText := ([]rune)(text)
	lenRuneText := len(runeText)
	totalWidth := 0 // 像素宽度

	if useEllipsis {

		// 保留起始位置后的像素宽度，多余的用...表示，包含...的像素宽度
		totalWidth = width - this.CalcWidth("...")
	} else {

		// 保留leftPos位置后的像素宽度
		totalWidth = width
	}

	i := lenRuneText
	for ; i > 0; i-- {
		if useEllipsis {
			// rune
			if totalWidth < this.CalcWidth((string)(runeText[0:i])) {
				// 宽度还够，继续
			} else {
				break
			}
		} else {
			// leftPos为一个rune位置
			// 保留leftPos位置后的像素宽度
			if leftPos+i < lenRuneText {
				if totalWidth < this.CalcWidth((string)(runeText[leftPos:i])) {
					// 宽度不够，继续
				} else {
					break
				}
			} else {
				if totalWidth < this.CalcWidth((string)(runeText[lenRuneText-i:])) {
					// 宽度不够，继续
				} else {
					break
				}
			}
		}
	}

	if !useEllipsis {
		// 不使用省略号
		if leftPos+i < lenRuneText {
			return (string)(runeText[leftPos:i])
		} else {
			return (string)(runeText[lenRuneText-i:])
		}
	} else {

		// 使用省略号
		if lenRuneText <= 3 {
			return "..."
		}

		if lenRuneText-i < 3 {
			i = lenRuneText - 3
		}

		retStr := (string)(runeText[:i]) + "..."

		return retStr
	}
}
