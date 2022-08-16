package utils

import (
	"math"
	"monster/pkg/common"
	"monster/pkg/common/define"
	"monster/pkg/common/define/enginesettings"
	"monster/pkg/common/fpoint"
	"monster/pkg/common/point"
	"monster/pkg/common/rect"
	"strconv"
)

func ResizeToScreen(settings common.Settings, eset common.EngineSettings, w, h int, crop bool, align int) rect.Rect {
	var r = rect.Construct()

	ratio := (float32)(settings.GetViewH()) / (float32)(h)
	r.W = (int)(ratio * (float32)(w))
	r.H = settings.GetViewH()

	if !crop {
		panic("!crop")
	}

	r = AlignToScreenEdge(settings, eset, align, r)

	return r
}

// 靠近一个边，对应轴的坐标逼近（0或最大），其他轴的坐标则居中
func AlignToScreenEdge(settings common.Settings, eset common.EngineSettings, alignment int, r rect.Rect) rect.Rect {
	ss := settings

	/*
		    (0,0)--------------->
			| \		|
			|   \		|
			|     \_	|
			|     |_|	|
			|		|
			|_______________|
			v

	*/

	switch alignment {
	case define.ALIGN_TOPLEFT:
		// pass
	case define.ALIGN_TOP:
		r.X = ss.GetViewWHalf() - r.W/2 + r.X
	case define.ALIGN_TOPRIGHT:
		r.X = ss.GetViewW() - r.W/2 + r.X
	case define.ALIGN_LEFT:
		r.Y = ss.GetViewHHalf() - r.H/2 + r.Y
	case define.ALIGN_CENTER:
		r.X = ss.GetViewWHalf() - r.W/2 + r.X
		r.Y = ss.GetViewHHalf() - r.H/2 + r.Y
	case define.ALIGN_RIGHT:
		r.X = ss.GetViewW() - r.W + r.X
		r.Y = ss.GetViewHHalf() - r.H/2 + r.Y
	case define.ALIGN_BOTTOMLEFT:
		r.Y = ss.GetViewH() - r.H + r.Y
	case define.ALIGN_BOTTOM:
		r.X = ss.GetViewWHalf() - r.W/2 + r.X
		r.Y = ss.GetViewH() - r.H + r.Y
	case define.ALIGN_BOTTOMRIGHT:
		r.X = ss.GetViewW() - r.W + r.X
		r.Y = ss.GetViewH() - r.H + r.Y

		// frame
	case define.ALIGN_FRAME_TOPLEFT:
		r.X = (ss.GetViewW()-eset.Get("resolutions", "menu_frame_width").(int))/2 + r.X
		r.Y = (ss.GetViewH()-eset.Get("resolutions", "menu_frame_height").(int))/2 + r.Y
	case define.ALIGN_FRAME_TOP:
		r.X = (ss.GetViewW()-eset.Get("resolutions", "menu_frame_width").(int))/2 + (eset.Get("resolutions", "menu_frame_width").(int)-r.W)/2 + r.X
		r.Y = (ss.GetViewH()-eset.Get("resolutions", "menu_frame_height").(int))/2 + r.Y
	case define.ALIGN_FRAME_TOPRIGHT:
		r.X = (ss.GetViewW()-eset.Get("resolutions", "menu_frame_width").(int))/2 + (eset.Get("resolutions", "menu_frame_width").(int) - r.W) + r.X
		r.Y = (ss.GetViewH()-eset.Get("resolutions", "menu_frame_height").(int))/2 + r.Y
	case define.ALIGN_FRAME_LEFT:
		r.X = (ss.GetViewW()-eset.Get("resolutions", "menu_frame_width").(int))/2 + r.X
		r.Y = (ss.GetViewH()-eset.Get("resolutions", "menu_frame_height").(int))/2 + (eset.Get("resolutions", "menu_frame_height").(int)/2 - r.H/2) + r.Y
	case define.ALIGN_FRAME_CENTER:
		r.X = (ss.GetViewW()-eset.Get("resolutions", "menu_frame_width").(int))/2 + (eset.Get("resolutions", "menu_frame_width").(int)/2 - r.W/2) + r.X
		r.Y = (ss.GetViewH()-eset.Get("resolutions", "menu_frame_height").(int))/2 + (eset.Get("resolutions", "menu_frame_height").(int)/2 - r.H/2) + r.Y
	case define.ALIGN_FRAME_RIGHT:
		r.X = (ss.GetViewW()-eset.Get("resolutions", "menu_frame_width").(int))/2 + (eset.Get("resolutions", "menu_frame_width").(int) - r.W) + r.X
		r.Y = (ss.GetViewH()-eset.Get("resolutions", "menu_frame_height").(int))/2 + (eset.Get("resolutions", "menu_frame_height").(int)/2 - r.H/2) + r.Y
	case define.ALIGN_FRAME_BOTTOMLEFT:
		r.X = (ss.GetViewW()-eset.Get("resolutions", "menu_frame_width").(int))/2 + r.X
		r.Y = (ss.GetViewH()-eset.Get("resolutions", "menu_frame_height").(int))/2 + (eset.Get("resolutions", "menu_frame_height").(int) - r.H) + r.Y
	case define.ALIGN_FRAME_BOTTOM:
		r.X = (ss.GetViewW()-eset.Get("resolutions", "menu_frame_width").(int))/2 + (eset.Get("resolutions", "menu_frame_width").(int)/2 - r.W/2) + r.X
		r.Y = (ss.GetViewH()-eset.Get("resolutions", "menu_frame_height").(int))/2 + (eset.Get("resolutions", "menu_frame_height").(int) - r.H) + r.Y
	case define.ALIGN_FRAME_BOTTOMRIGHT:
		r.X = (ss.GetViewW()-eset.Get("resolutions", "menu_frame_width").(int))/2 + (eset.Get("resolutions", "menu_frame_width").(int) - r.W) + r.X
		r.Y = (ss.GetViewH()-eset.Get("resolutions", "menu_frame_height").(int))/2 + (eset.Get("resolutions", "menu_frame_height").(int) - r.H) + r.Y

	}

	return r
}

/*
src__________________
|		    |
|                   |  range
|                   |
|___________________|
      range         target
*/

// 调整目标点，使得目标在range * range的矩形内
func ClampDistance(range1 float32, src, target fpoint.FPoint) fpoint.FPoint {
	limitTarget := target

	if range1 > 0 {
		if src.X+range1 < target.X {
			limitTarget.X = src.X + range1
		}

		if src.X-range1 > target.X {
			limitTarget.X = src.X - range1
		}

		if src.Y+range1 < target.Y {
			limitTarget.Y = src.Y + range1
		}

		if src.Y-range1 > target.Y {
			limitTarget.Y = src.Y - range1
		}
	}

	return limitTarget
}

// 某个方向移动距离得到的点
func CalcVector(pos fpoint.FPoint, direction uint8, dist float32) fpoint.FPoint {
	p := pos

	distStraight := dist
	_ = distStraight

	distDiag := dist * 0.7071 //  1/sqrt(2)

	switch direction {
	case 0:
		// 左下
		p.X -= distDiag
		p.Y += distDiag
	case 1:
		// 左
		p.X -= distStraight
	case 2:
		// 左上
		p.X -= distDiag
		p.Y -= distDiag
	case 3:
		// 上
		p.Y -= distStraight
	case 4:
		// 右上
		p.X += distDiag
		p.Y -= distDiag
	case 5:
		// 右
		p.X += distStraight
	case 6:
		p.X += distDiag
		p.Y += distDiag
	case 7:
		p.Y += distStraight
	}

	return p

}

func CalcDist(p1, p2 fpoint.FPoint) float32 {
	return (float32)(math.Sqrt((float64)((p2.X-p1.X)*(p2.X-p1.X) + (p2.Y-p1.Y)*(p2.Y-p1.Y))))
}

// 计算方向，8个方向
func CalcDirection(x0, y0, x1, y1 float32) uint8 {
	theta := CalcTheta(x0, y0, x1, y1)

	val := theta / (math.Pi / 4)

	var dir int
	if val < 0 {
		dir = (int)(math.Ceil(float64(val)-0.5) + 4)
	} else {
		dir = (int)(math.Floor(float64(val)+0.5) + 4)
	}

	dir = (dir + 1) % 8
	if dir >= 0 && dir < 8 {
		return (uint8)(dir)
	}

	return 0
}

// 计算角度
func CalcTheta(x1, y1, x2, y2 float32) float32 {

	dx := x2 - x1
	dy := y2 - y1
	exactDx := x2 - x1

	if exactDx == 0 {
		if dy > 0 {
			return math.Pi / 2 // 90 度
		}

		return -math.Pi / 2
	}

	m := math.Atan(float64(dy) / (float64)(dx))

	if dx < 0 && dy >= 0 {
		m += math.Pi
	}

	if dx < 0 && dy < 0 {
		m -= math.Pi
	}

	return float32(m)
}

func IsWithinRect(r rect.Rect, target point.Point) bool {
	return target.X >= r.X && target.Y >= r.Y && target.X < r.X+r.W && target.Y < r.Y+r.H
}

func GetTimeString(time int) string {
	ss := ""
	hours := time / 3600
	if hours < 10 {
		ss = "0" + strconv.Itoa(hours)
	} else {
		ss = strconv.Itoa(hours)
	}

	ss += ":"

	minutes := (time / 60) % 60
	if minutes < 10 {
		ss = ss + "0" + strconv.Itoa(minutes)
	} else {
		ss += strconv.Itoa(minutes)
	}

	ss += ":"
	seconds := time % 60

	if seconds < 10 {
		ss = ss + "0" + strconv.Itoa(seconds)
	} else {
		ss += strconv.Itoa(seconds)
	}

	return ss
}

// 屏幕上的点转换到地图上
func ScreenToMap(settings common.Settings, eset common.EngineSettings, x, y int, camX, camY float32) fpoint.FPoint {

	r := fpoint.Construct()

	if eset.Get("tileset", "orientation").(int) == enginesettings.TILESET_ISOMETRIC {
		// x 和 y 转换到屏幕坐标系，屏幕中间点为和地图坐标换算的偏移
		scrX := (float32)(x-settings.GetViewWHalf()) * 0.5
		scrY := (float32)(y-settings.GetViewHHalf()) * 0.5

		// 摄像头 转换到世界地图的坐标
		r.X = (eset.Get("tileset", "units_per_pixel_x").(float32) * scrX) + (eset.Get("tileset", "units_per_pixel_y").(float32) * scrY) + camX
		r.Y = (eset.Get("tileset", "units_per_pixel_y").(float32) * scrY) - (eset.Get("tileset", "units_per_pixel_x").(float32) * scrX) + camY
	} else {
		r.X = (float32)(x-settings.GetViewWHalf())*eset.Get("tileset", "units_per_pixel_x").(float32) + camX
		r.Y = (float32)(y-settings.GetViewHHalf())*eset.Get("tileset", "units_per_pixel_y").(float32) + camY
	}

	return r
}

func MapToScreen(settings common.Settings, eset common.EngineSettings, x, y, camX, camY float32) point.Point {
	r := point.Construct()

	adjustX := ((float32)(settings.GetViewWHalf()) + 0.5) * eset.Get("tileset", "units_per_pixel_x").(float32)
	adjustY := ((float32)(settings.GetViewHHalf()) + 0.5) * eset.Get("tileset", "units_per_pixel_y").(float32)

	if eset.Get("tileset", "orientation").(int) == enginesettings.TILESET_ISOMETRIC {
		r.X = (int)(math.Floor(float64((x-camX-y+camY+adjustX)/eset.Get("tileset", "units_per_pixel_x").(float32) + 0.5)))
		r.Y = (int)(math.Floor(float64((x-camX+y-camY+adjustY)/eset.Get("tileset", "units_per_pixel_y").(float32) + 0.5)))
	} else {
		r.X = (int)((x - camX + adjustX) / eset.Get("tileset", "units_per_pixel_x").(float32))
		r.Y = (int)((y - camY + adjustY) / eset.Get("tileset", "units_per_pixel_y").(float32))

	}
	return r
}
