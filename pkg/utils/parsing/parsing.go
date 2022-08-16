package parsing

import (
	"bytes"
	"fmt"
	"math"
	"monster/pkg/common/color"
	"monster/pkg/common/define"
	"monster/pkg/common/define/fontengine"
	"monster/pkg/common/item"
	"monster/pkg/common/labelinfo"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"reflect"
	"strconv"
	"strings"
)

func GetKeyPair(s []byte) (string, string) {
	ts := bytes.TrimSpace(s)
	kv := bytes.SplitN(ts, []byte("="), 2)
	if len(kv) != 2 {
		return "", ""
	}

	key := string(bytes.TrimSpace(kv[0]))
	val := string(bytes.TrimSpace(kv[1]))
	/*
		if key == "" || val == "" {
			return "", ""
		}
	*/

	return key, val

}

func PopFirstString(s, separator string) (string, string) {
	seppos := 0
	if s == "" {
		return "", ""
	}

	if separator == "" {
		seppos = strings.Index(s, ",")
		alt_seppos := strings.Index(s, ";")

		if alt_seppos != -1 && alt_seppos < seppos {
			seppos = alt_seppos
		}
	} else {
		seppos = strings.Index(s, separator)
	}

	if seppos == -1 {
		return s, ""
	}

	return s[:seppos], s[seppos+1:]
}

func PopFirstInt(s, separator string) (int, string) {
	firstStr, str := PopFirstString(s, separator)
	return ToInt(firstStr, 0), str
}

func ToInt(s string, defaultVal int) int {
	result, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}

	return result
}

func ToFloat(s string, defaultVal float32) float32 {
	result, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return defaultVal
	}

	return (float32)(result)
}

func ToPowerId(s string, defaultVal define.PowerId) define.PowerId {
	return (define.PowerId)(ToInt(s, (int)(defaultVal)))
}

func ToUnsignedLong(s string, defaultValue uint64) uint64 {
	uint64Val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return defaultValue
	}

	return uint64Val
}

func PopLabelInfo(val string) labelinfo.LabelInfo {
	strVal := val
	info := labelinfo.Construct()

	justify := ""
	valign := ""
	style := ""
	tmp := ""
	tmp, strVal = PopFirstString(strVal, "")
	if tmp == "hidden" {
		info.Hidden = true
	} else {
		info.Hidden = false
		info.X = ToInt(tmp, 0)
		info.Y, strVal = PopFirstInt(strVal, "")
		justify, strVal = PopFirstString(strVal, "")
		valign, strVal = PopFirstString(strVal, "")
		style, strVal = PopFirstString(strVal, "")

		if justify == "left" {
			info.Justify = fontengine.JUSTIFY_LEFT
		} else if justify == "center" {
			info.Justify = fontengine.JUSTIFY_CENTER
		} else if justify == "right" {
			info.Justify = fontengine.JUSTIFY_RIGHT
		}

		if valign == "top" {
			info.Valign = labelinfo.VALIGN_TOP
		} else if valign == "center" {
			info.Valign = labelinfo.VALIGN_CENTER
		} else if valign == "bottom" {
			info.Valign = labelinfo.VALIGN_BOTTOM
		}

		if style != "" {
			info.FontStyle = style
		}
	}

	return info
}

func TryParseValue(strValue string, output interface{}) bool {
	if output == nil {
		return false
	}

	t := reflect.TypeOf(output)

	if t.Kind() != reflect.Ptr {
		return false
	}

	vv := reflect.Indirect(reflect.ValueOf(output)) // ptr to interface
	vvv := vv.Elem()

	switch vvv.Kind() {
	case reflect.Bool:
		if ToBool(strValue) {
			vv.Set(reflect.ValueOf(true))
			return true
		} else {
			vv.Set(reflect.ValueOf(false))
			return true
		}

	case reflect.Int8:
		int8Value, err := strconv.ParseInt(strValue, 10, 8)
		if err != nil {
			return false
		}
		vv.Set(reflect.ValueOf((int8)(int8Value)))
		return true
	case reflect.Int16:
		int16Value, err := strconv.ParseInt(strValue, 10, 16)
		if err != nil {
			return false
		}
		vv.Set(reflect.ValueOf((int16)(int16Value)))
		return true
	case reflect.Int:
		intValue, err := strconv.ParseInt(strValue, 10, 0)
		if err != nil {
			return false
		}
		vv.Set(reflect.ValueOf((int)(intValue)))
		return true
	case reflect.Uint:
		uintValue, err := strconv.ParseUint(strValue, 10, 0)
		if err != nil {
			return false
		}
		vv.Set(reflect.ValueOf(uint(uintValue)))
		return true
	case reflect.Uint16:
		uint16Value, err := strconv.ParseUint(strValue, 10, 16)
		if err != nil {
			return false
		}
		vv.Set(reflect.ValueOf(uint16(uint16Value)))
		return true
	case reflect.Uint8:
		uint8Value, err := strconv.ParseUint(strValue, 10, 8)
		if err != nil {
			return false
		}
		vv.Set(reflect.ValueOf(uint8(uint8Value)))
		return true
	case reflect.Float32:
		float32Value, err := strconv.ParseFloat(strValue, 32)
		if err != nil {
			return false
		}
		vv.Set(reflect.ValueOf(float32(float32Value)))
		return true
	case reflect.Float64:
		float64Value, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			return false
		}
		vv.Set(reflect.ValueOf(float64Value))
		return true
	case reflect.String:
		vv.Set(reflect.ValueOf(strValue))
		return true
	}

	return false
}

func GetSectionTitle(strVal string) string {
	s := strings.TrimSpace(strVal)
	index := strings.Index(s, "]")
	if index < 0 {
		return ""
	}

	return s[1:index]
}

// 时间转化到帧数 ms或s为单位的字符串
func ToDuration(strVal string, maxFps int) int {
	strVal = strings.TrimSpace(strVal)
	isMs := true

	if strings.HasSuffix(strVal, "ms") {
		strVal = strings.TrimSpace(strings.TrimSuffix(strVal, "ms"))
		isMs = true
	} else if strings.HasSuffix(strVal, "s") {
		strVal = strings.TrimSpace(strings.TrimSuffix(strVal, "s"))
		isMs = false
	}
	// 默认ms
	int64Val, err := strconv.ParseInt(strVal, 10, 0)
	if err != nil || int64Val == 0 {
		return 0
	}

	intVal := (int)(int64Val)
	if !isMs {
		intVal *= maxFps
	} else {
		// 向下取整
		intVal = (int)(math.Floor(float64(intVal*maxFps)/1000.0 + 0.5))
	}

	if intVal < 1 {
		return 1
	}

	return (int)(intVal)
}

func ToBool(strVal string) bool {
	strVal = strings.TrimSpace(strVal)

	if strVal == "true" || strVal == "1" || strVal == "yes" {
		return true
	} else if strVal == "false" || strVal == "0" || strVal == "no" {
		return false
	}

	return false

}

func ToRGBA(strVal string) color.Color {
	c := color.Construct()

	intVal := 0
	intVal, strVal = PopFirstInt(strVal, "")
	c.R = (uint8)(intVal)
	intVal, strVal = PopFirstInt(strVal, "")
	c.G = (uint8)(intVal)
	intVal, strVal = PopFirstInt(strVal, "")
	c.B = (uint8)(intVal)
	intVal, strVal = PopFirstInt(strVal, "")
	c.A = (uint8)(intVal)

	return c
}

func ToRGB(strVal string) color.Color {
	c := color.Construct()

	intVal := 0
	intVal, strVal = PopFirstInt(strVal, "")
	c.R = (uint8)(intVal)
	intVal, strVal = PopFirstInt(strVal, "")
	c.G = (uint8)(intVal)
	intVal, strVal = PopFirstInt(strVal, "")
	c.B = (uint8)(intVal)

	return c
}

func ToPoint(strVal string) point.Point {
	p := point.Construct()

	intVal := 0
	intVal, strVal = PopFirstInt(strVal, "")
	p.X = intVal
	intVal, strVal = PopFirstInt(strVal, "")
	p.Y = intVal

	return p
}

func ToRect(strVal string) rect.Rect {
	r := rect.Construct()
	r.X, strVal = PopFirstInt(strVal, "")
	r.Y, strVal = PopFirstInt(strVal, "")
	r.W, strVal = PopFirstInt(strVal, "")
	r.H, strVal = PopFirstInt(strVal, "")

	return r
}

// 对齐的字符串转枚举值
func ToAlignment(s string, defaultValue int) int {
	align := defaultValue

	switch s {
	case "topleft":
		align = define.ALIGN_TOPLEFT
	case "top":
		align = define.ALIGN_TOP
	case "topright":
		align = define.ALIGN_TOPRIGHT
	case "left":
		align = define.ALIGN_LEFT
	case "center":
		align = define.ALIGN_CENTER
	case "right":
		align = define.ALIGN_RIGHT
	case "bottomleft":
		align = define.ALIGN_BOTTOMLEFT
	case "bottom":
		align = define.ALIGN_BOTTOM
	case "bottomright":
		align = define.ALIGN_BOTTOMRIGHT
	case "frame_topleft":
		align = define.ALIGN_FRAME_TOPLEFT
	case "frame_top":
		align = define.ALIGN_FRAME_TOP
	case "frame_topright":
		align = define.ALIGN_FRAME_TOPRIGHT
	case "frame_left":
		align = define.ALIGN_FRAME_LEFT
	case "frame_center":
		align = define.ALIGN_FRAME_CENTER
	case "frame_right":
		align = define.ALIGN_FRAME_RIGHT
	case "frame_bottomleft":
		align = define.ALIGN_FRAME_BOTTOMLEFT
	case "frame_bottom":
		align = define.ALIGN_FRAME_BOTTOM
	case "frame_bottomright":
		align = define.ALIGN_FRAME_BOTTOMRIGHT
	}

	return align
}

func ToDirection(s string) int {
	switch s {
	case "N":
		return 3
	case "NE":
		return 4
	case "E":
		return 5
	case "SE":
		return 6
	case "S":
		return 7
	case "SW":
		return 0
	case "W":
		return 1
	case "NW":
		return 2
	default:
		dir := ToInt(s, 0)
		if dir < 0 || dir > 7 {
			fmt.Printf("UtilsParsing: Direction '%d' is not within range 0-7.\n", dir)
		}
		return dir
	}

	return 0
}

func ToItemQuantityPair(value string) (item.Stack, bool) {
	r := item.ConstructStack()

	checkPair := strings.Contains(value, ":")

	value += ":"
	var first string

	first, value = PopFirstString(value, ":")
	r.Item = define.ItemId(ToInt(first, 0))
	r.Quantity, value = PopFirstInt(value, ":")

	if r.Quantity == 0 {
		r.Quantity = 1
	}

	return r, checkPair
}
