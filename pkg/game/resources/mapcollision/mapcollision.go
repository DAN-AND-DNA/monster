package mapcollision

import (
	"math"
	"math/rand"
	"monster/pkg/common"
	"monster/pkg/common/define/game/mapcollision"
	"monster/pkg/common/fpoint"
	"monster/pkg/common/gameres"
	"monster/pkg/common/point"
	"monster/pkg/utils"
)

const (
	MIN_TILE_GAP = 0.001
)

type MapCollision struct {
	mapSize point.Point // 地图大小 宽和高
	colMap  [][]uint16
}

func New() *MapCollision {
	cs := &MapCollision{}
	cs.Init()
	return cs
}

func (this *MapCollision) Init() gameres.MapCollision {
	this.mapSize = point.Construct()
	this.colMap = make([][]uint16, 1)
	this.colMap[0] = make([]uint16, 1)

	return this
}

func (this *MapCollision) SetMap(colMap [][]uint16, w, h uint16) {
	this.colMap = make([][]uint16, w)

	for i := (uint16)(0); i < w; i++ {
		this.colMap[i] = make([]uint16, h)
	}

	for i := (uint16)(0); i < w; i++ {
		for j := (uint16)(0); j < h; j++ {
			this.colMap[i][j] = colMap[i][j]
		}
	}

	this.mapSize.X = (int)(w)
	this.mapSize.Y = (int)(h)
}

// 网格是否在地图外
func (this *MapCollision) IsTileOutsideMap(tileX, tileY int) bool {
	// 0 到 最宽
	// 0 到 最高
	return tileX < 0 || tileY < 0 || tileX >= this.mapSize.X || tileY >= this.mapSize.Y
}

func (this *MapCollision) IsOutsideMap(tileX, tileY float32) bool {
	return this.IsTileOutsideMap((int)(tileX), (int)(tileY))
}

// 一个实体有移动类型和碰撞类型，那么坐标处的瓷砖是否对它而言合法，可否走到该瓷砖
func (this *MapCollision) IsValidTile(modules common.Modules, tileX, tileY int, movementType, collideType int) bool {
	eset := modules.Eset()

	// 地图范围外非法
	if this.IsTileOutsideMap(tileX, tileY) {
		return false
	}

	if collideType == mapcollision.COLLIDE_NORMAL {
		if this.colMap[tileX][tileY] == mapcollision.BLOCKS_ENEMIES {
			return false
		}

		if this.colMap[tileX][tileY] == mapcollision.BLOCKS_ENTITIES {
			return false
		}

	} else if collideType == mapcollision.COLLIDE_HERO {
		if this.colMap[tileX][tileY] == mapcollision.BLOCKS_ENEMIES && !eset.Get("misc", "enable_ally_collision").(bool) {
			return true
		}
	}

	// 无形的
	if movementType == mapcollision.MOVE_INTANGIBLE {
		return true
	}

	// 飞行
	if movementType == mapcollision.MOVE_FLYING {
		return !(this.colMap[tileX][tileY] == mapcollision.BLOCKS_ALL || this.colMap[tileX][tileY] == mapcollision.BLOCKS_ALL_HIDDEN)
	}

	if this.colMap[tileX][tileY] == mapcollision.MAP_ONLY || this.colMap[tileX][tileY] == mapcollision.MAP_ONLY_ALT {
		return true
	}

	// normal should be none
	return this.colMap[tileX][tileY] == mapcollision.BLOCKS_NONE

}

// 一个实体具有移动类型和碰撞类型，该位置是否对其是合法的
func (this *MapCollision) IsValidPosition(modules common.Modules, x, y float32, movementType, collideType int) bool {
	if x < 0 || y < 0 {
		return false
	}

	return this.IsValidTile(modules, int(x), int(y), movementType, collideType)
}

// 实体从x,y出发，走step距离到一个坐标位置，成功则true
func (this *MapCollision) SmallStep(modules common.Modules, x, y, stepX, stepY float32, movementType, collideType int) (float32, float32, bool) {
	// 验证起点
	if !this.IsValidPosition(modules, x, y, movementType, collideType) {
		panic("bad src position")
	}

	if this.IsValidPosition(modules, x+stepX, y+stepY, movementType, collideType) {
		x += stepX
		y += stepY
		return x, y, true
	}

	return x, y, false
}

// 沿一个轴直行
func (this *MapCollision) SmallStepForcedSlideAlongGrid(modules common.Modules, x, y, stepX, stepY float32, movementType, collideType int) (float32, float32, bool) {
	// 验证起点
	if !this.IsValidPosition(modules, x, y, movementType, collideType) {
		panic("bad src position")
	}

	// 2个轴都尝试
	if this.IsValidPosition(modules, x+stepX, y, movementType, collideType) {
		x += stepX
		return x, y, true
	} else if this.IsValidPosition(modules, x, y+stepY, movementType, collideType) {
		y += stepY
		return x, y, true
	}

	return x, y, false
}

// 只获得符合
func sgn(f float32) int {
	if f > 0 {
		return 1
	} else if f < 0 {
		return -1
	}

	return 0
}

/*
         ______________>x
	|      |_|
	|
	|
	|
	|
	|
	v
	y
*/

// 沿轴直行遇到瓷砖障碍物，从另一个轴方向直接绕开该瓷砖，比如遇到墙壁，只能沿着墙移动，一个轴的步长为0
func (this *MapCollision) SmallStepForcedSlide(modules common.Modules, x, y, stepX, stepY float32, movementType, collideType int) (float32, float32, bool) {
	// 验证起点
	if !this.IsValidPosition(modules, x, y, movementType, collideType) {
		panic("bad src position")
	}

	if stepX != 0 {
		// 沿x轴直行遇到瓷砖障碍物，从y轴方向绕开该瓷砖
		if stepY != 0 {
			panic("only step by x")
		}

		dy := (float64)(y) - (math.Floor((float64)(y)))

		if this.IsValidTile(modules, (int)(x), (int)(y)+1, movementType, collideType) &&
			this.IsValidTile(modules, (int)(x)+sgn(stepX), (int)(y)+1, movementType, collideType) &&
			dy > 0.5 {

			// 向下
			y += (float32)(math.Min(1-dy+0.01, math.Abs((float64)(stepX))))
			return x, y, true

		} else if this.IsValidTile(modules, (int)(x), (int)(y)-1, movementType, collideType) &&
			this.IsValidTile(modules, (int)(x)+sgn(stepX), (int)(y)-1, movementType, collideType) &&
			dy < 0.5 {

			//向上
			y -= (float32)(math.Min(dy+0.01, math.Abs((float64)(stepX))))
			return x, y, true
		}

	} else if stepY != 0 {
		// 沿y轴直行遇到瓷砖障碍物，从x轴方向绕开该瓷砖
		if stepX != 0 {
			panic("only step by y")
		}

		dx := (float64)(x) - (math.Floor((float64)(x)))

		if this.IsValidTile(modules, (int)(x)+1, (int)(y), movementType, collideType) &&
			this.IsValidTile(modules, (int)(x)+1, (int)(y)+sgn(stepY), movementType, collideType) &&
			dx > 0.5 {

			// 向右
			x += (float32)(math.Min(1-dx+0.01, math.Abs((float64)(stepY))))
			return x, y, true

		} else if this.IsValidTile(modules, (int)(x)-1, (int)(y), movementType, collideType) &&
			this.IsValidTile(modules, (int)(x)-1, (int)(y)+sgn(stepY), movementType, collideType) &&
			dx < 0.5 {

			// 向左
			x -= (float32)(math.Min(dx+0.01, math.Abs((float64)(stepY))))
			return x, y, true
		}
	}

	return x, y, false
}

// 尝试移动
// 90度的障碍物停止返回false
// 45度或者135度，尝试从2个轴绕开
func (this *MapCollision) Move(modules common.Modules, x, y, stepX, stepY float32, movementType, collideType int) (float32, float32, bool) {

	// 遇到墙壁(某个方向的步长为0)，则只能尝试另一个轴的移动
	// 提前发现不是沿墙壁走的情况，使得函数尽快返回true
	forceSlide := (stepX != 0 && stepY != 0)

	//尝试各个方向
	for stepX != 0 || stepY != 0 {

		// 重新计算的x轴步长
		newStepX := float32(0)
		if stepX > 0 {
			// 向右
			newStepX = (float32)(math.Min(math.Ceil(float64(x))-float64(x), float64(stepX)))

			if newStepX <= MIN_TILE_GAP {
				newStepX = (float32)(math.Min(1, (float64)(stepX)))
			}

		} else if stepX < 0 {
			// 向左
			newStepX = (float32)(math.Max(math.Floor(float64(x))-float64(x), float64(stepX)))

			if newStepX == 0 {
				newStepX = (float32)(math.Max(-1, (float64)(stepX)))
			}

		}

		newStepY := float32(0)
		if stepY > 0 {
			// 向下

			// 最近的是下个瓷砖还是整个移动距离
			newStepY = (float32)(math.Min(math.Ceil(float64(y))-float64(y), float64(stepY)))

			if newStepY <= MIN_TILE_GAP {
				// 如果已经处在另一个瓷砖的边缘了，就当作已经在另一个瓷砖
				newStepY = (float32)(math.Min(1, (float64)(stepY)))
			}

		} else if stepY < 0 {
			// 向上

			// 最近的是下个瓷砖还是整个移动距离
			newStepY = (float32)(math.Max(math.Floor(float64(y))-float64(y), float64(stepY)))

			if newStepY == 0 {
				// 如果已经处在另一个瓷砖的边缘了，就当作已经在另一个瓷砖
				newStepY = (float32)(math.Max(-1, (float64)(stepY)))
			}
		}

		stepX -= newStepX
		stepY -= newStepY

		ok := false

		// 先尝试普通移动
		x, y, ok = this.SmallStep(modules, x, y, newStepX, newStepY, movementType, collideType)
		if !ok {

			if forceSlide {
				// x,y轴都可以尝试
				// 尝试沿x或y其中一个边直行，失败换另一边
				x, y, ok = this.SmallStepForcedSlideAlongGrid(modules, x, y, newStepX, newStepY, movementType, collideType)
				if !ok {
					return x, y, false
				}
			} else {
				// 只能尝试x或y其中一边，如果失败则强制从另一个轴绕口
				x, y, ok = this.SmallStepForcedSlide(modules, x, y, newStepX, newStepY, movementType, collideType)
				if !ok {
					return x, y, false
				}

			}
		}

	}

	return x, y, true
}

// 无碰撞，但地图外作为边界区域故不为空
func (this *MapCollision) IsEmpty(x, y float32) bool {
	tileX := int(x)
	tileY := int(y)
	if this.IsTileOutsideMap(tileX, tileY) {
		return false
	}

	val := this.colMap[tileX][tileY]
	return val == mapcollision.BLOCKS_NONE || val == mapcollision.MAP_ONLY || val == mapcollision.MAP_ONLY_ALT
}

func (this *MapCollision) IsWall(x, y float32) bool {
	tileX := int(x)
	tileY := int(y)
	if this.IsTileOutsideMap(tileX, tileY) {
		return true
	}

	val := this.colMap[tileX][tileY]
	return val == mapcollision.BLOCKS_ALL || val == mapcollision.BLOCKS_ALL_HIDDEN
}

/*
	 ______________
	|___|          |
	|   |\ x1,y1    |
	|dy | \         |
	|   |  \        |
	|___|___\x2,y2__|
             dx
*/

// 计算2点连线间的点是否可见或者可移动
func (this *MapCollision) LineCheck(modules common.Modules, x1, y1, x2, y2 float32, checkType, movementType int) bool {
	x := x1
	y := y1

	var stepX, stepY float64
	dx := math.Abs(float64(x2 - x1))
	dy := math.Abs(float64(y2 - y1))
	steps := (int)(math.Max(dx, dy))

	// 最大
	if dx > dy {
		stepX = 1
		stepY = dy / dx
	} else {
		stepY = 1
		stepX = dx / dy
	}

	if x1 > x2 {
		stepX = -stepX
	}

	if y1 > y2 {
		stepY = -stepY
	}

	if checkType == mapcollision.CHECK_SIGHT {
		// 可见
		for i := 0; i < steps; i++ {
			x += (float32)(stepX)
			y += (float32)(stepY)
			if this.IsWall(x, y) {
				return false
			}
		}
	} else if checkType == mapcollision.CHECK_MOVEMENT {
		// 移动
		for i := 0; i < steps; i++ {
			x += (float32)(stepX)
			y += (float32)(stepY)
			if !this.IsValidPosition(modules, x, y, movementType, mapcollision.COLLIDE_NORMAL) {
				return false
			}
		}

	}

	return true
}

func (this *MapCollision) LineOfSight(modules common.Modules, x1, y1, x2, y2 float32) bool {
	return this.LineCheck(modules, x1, y1, x2, y2, mapcollision.CHECK_SIGHT, mapcollision.COLLIDE_NORMAL)
}

// 设置地图上空的点为敌人或者实体
func (this *MapCollision) Block(mapX, mapY float32, isAlly bool) {
	tileX := (int)(mapX)
	tileY := (int)(mapY)

	if this.IsTileOutsideMap(tileX, tileY) {
		return
	}

	if this.colMap[tileX][tileY] == mapcollision.BLOCKS_NONE {
		if isAlly {
			this.colMap[tileX][tileY] = mapcollision.BLOCKS_ENEMIES
		} else {
			this.colMap[tileX][tileY] = mapcollision.BLOCKS_ENTITIES
		}
	}
}

// 设置地图上的敌人或实体为空
func (this *MapCollision) Unblock(mapX, mapY float32) {
	tileX := (int)(mapX)
	tileY := (int)(mapY)

	if this.IsTileOutsideMap(tileX, tileY) {
		return
	}

	if this.colMap[tileX][tileY] == mapcollision.BLOCKS_ENTITIES || this.colMap[tileX][tileY] == mapcollision.BLOCKS_ENEMIES {
		this.colMap[tileX][tileY] = mapcollision.BLOCKS_NONE
	}
}

func (this *MapCollision) LineOfMovement(modules common.Modules, x1, y1, x2, y2 float32, movementType int) bool {
	if this.IsOutsideMap(x2, y2) {
		return false
	}

	if movementType == mapcollision.MOVE_INTANGIBLE {
		return true
	}

	tileX := (int)(x2)
	tileY := (int)(y2)
	targetBlocks := false
	targetBlocksType := this.colMap[tileX][tileY]

	if this.colMap[tileX][tileY] == mapcollision.BLOCKS_ENTITIES || this.colMap[tileX][tileY] == mapcollision.BLOCKS_ENEMIES {
		targetBlocks = true
		this.Unblock(x2, y2)
	}

	hasMovement := this.LineCheck(modules, x1, y1, x2, y2, mapcollision.CHECK_MOVEMENT, movementType)

	if targetBlocks {
		this.Block(x2, y2, targetBlocksType == mapcollision.BLOCKS_ENEMIES)
	}

	return hasMovement
}

// 是否点1面向点2
func (this *MapCollision) IsFacing(x1, y1, x2, y2 float32, direction uint8) bool {

	// 180 fov
	switch direction {
	case 2: //north west
		return ((x2 - x1) < ((-1 * y2) - (-1 * y1))) && (((-1 * x2) - (-1 * x1)) > (y2 - y1))
	case 3: //north
		return y2 < y1
	case 4: //north east
		return (((-1 * x2) - (-1 * x1)) < ((-1 * y2) - (-1 * y1))) && ((x2 - x1) > (y2 - y1))
	case 5: //east
		return x2 > x1
	case 6: //south east
		return ((x2 - x1) > ((-1 * y2) - (-1 * y1))) && (((-1 * x2) - (-1 * x1)) < (y2 - y1))
	case 7: //south
		return y2 > y1
	case 0: //south west
		return (((-1 * x2) - (-1 * x1)) > ((-1 * y2) - (-1 * y1))) && ((x2 - x1) < (y2 - y1))
	case 1: //west
		return x2 < x1
	}
	return false
}

func (this *MapCollision) CollisionToMap(p point.Point) fpoint.FPoint {
	ret := fpoint.Construct(float32(p.X)+0.5, float32(p.Y)+0.5)
	return ret
}

// 计算路径
func (this *MapCollision) ComputePath(modules common.Modules, startPos, endPos fpoint.FPoint, movementType int, limit uint) ([]fpoint.FPoint, bool) {
	if this.IsOutsideMap(endPos.X, endPos.Y) {
		return nil, false
	}

	// 为0则默认为地图大小的10%
	if limit == 0 {
		limit = (uint)(this.mapSize.X*this.mapSize.Y) / 10
	}

	start := point.Construct((int)(startPos.X), (int)(startPos.Y))
	end := point.Construct((int)(startPos.X), (int)(startPos.Y))

	// 临时清理
	targetBlocks := false
	targetBlocksType := this.colMap[end.X][end.Y]

	if this.colMap[end.X][end.Y] == mapcollision.BLOCKS_ENTITIES || this.colMap[end.X][end.Y] == mapcollision.BLOCKS_ENEMIES {
		targetBlocks = true
		this.Unblock(endPos.X, endPos.Y)
	}

	current := start
	node := newAstarNode(start)
	node.SetActualCost(0)
	node.SetEstimatedCost(utils.CalcDist(fpoint.Construct(float32(start.X), float32(start.Y)), fpoint.Construct(float32(end.X), float32(end.Y))))
	node.SetParent(current)

	openc := newAstarContainer((uint)(this.mapSize.X), (uint)(this.mapSize.Y), limit)
	defer openc.Close()
	closec := newAstarCloseContainer((uint)(this.mapSize.X), (uint)(this.mapSize.Y), limit)
	defer closec.Close()

	openc.Add(node)

	// 非空且节点未到上限数
	for !openc.IsEmpty() && (uint)(closec.GetSize()) < limit {
		node = openc.GetShortestF() // 获取开销最小的节点

		current.X = node.GetX()
		current.Y = node.GetY()
		closec.Add(node) // 开销最小添加到确认的路里面
		openc.Remove(node)

		if current.X == end.X && current.Y == end.Y {
			break // 路径找到
		}

		// 把当前节点相邻的节点都添加
		neighbours := node.GetNeighbours(this.mapSize.X, this.mapSize.Y)
		for _, val := range neighbours {
			neighbour := val

			if uint(openc.GetSize()) >= limit {
				break
			}

			if !this.IsValidTile(modules, neighbour.X, neighbour.Y, movementType, mapcollision.COLLIDE_NORMAL) {
				continue
			}

			if closec.Exists(neighbour) {
				continue
			}

			if !openc.Exists(neighbour) {
				// 添加新节点
				newNode := newAstarNode(neighbour)
				newNode.SetParent(current)
				// 设置估计开销
				newNode.SetEstimatedCost(utils.CalcDist(fpoint.Construct(float32(neighbour.X), float32(neighbour.Y)), fpoint.Construct(float32(end.X), float32(end.Y))))
				openc.Add(newNode)

			} else {
				i, ok := openc.Get(neighbour.X, neighbour.Y)
				if !ok {
					return nil, false
				}

				tmpVal := utils.CalcDist(fpoint.Construct(float32(current.X), float32(current.Y)), fpoint.Construct(float32(neighbour.X), float32(neighbour.Y)))
				if node.GetActualCost()+tmpVal < i.GetActualCost() {

					// 更新实际开销
					pos := point.Construct(i.GetX(), i.GetY())
					parentPos := point.Construct(node.GetX(), node.GetY())
					openc.UpdateParent(pos, parentPos, node.GetActualCost()+tmpVal) // 更新来源和实际开销
				}
			}
		}
	}

	var path []fpoint.FPoint
	if !(current.X == end.X && current.Y == end.Y) {

		// 没找到路 给一条相对靠近的路
		ok := false
		node, ok = closec.GetShortestH()
		if !ok {
			return nil, false
		}

		current.X = node.GetX()
		current.Y = node.GetY()

		for !(current.X == start.X && current.Y == start.Y) {
			path = append(path, this.CollisionToMap(current))
			tmp, ok := closec.Get(current.X, current.Y)
			if !ok {
				return nil, false
			}

			current = tmp.GetParent()
		}
	} else {
		// 找到
		path = append(path, this.CollisionToMap(end))
		for !(current.X == start.X && current.Y == start.Y) {
			path = append(path, this.CollisionToMap(current))
			tmp, ok := closec.Get(current.X, current.Y)
			if !ok {
				return nil, false
			}

			current = tmp.GetParent()
		}

	}

	if targetBlocks {
		// 恢复
		this.Block(endPos.X, endPos.Y, targetBlocksType == mapcollision.BLOCKS_ENEMIES)
	}

	if len(path) == 0 {
		return nil, false
	}

	return path, true
}

// 返回目标相邻的随机瓷砖
func (this *MapCollision) GetRandomNeighbor(modules common.Modules, target point.Point, range1 int, ignoreBlocked bool) fpoint.FPoint {

	oldTarget := fpoint.Construct(float32(target.X), float32(target.Y))
	newTarget := oldTarget

	var validTiles []fpoint.FPoint

	for i := -range1; i < range1; i++ {
		for j := -range1; i < range1; j++ {
			if i == 0 && j == 0 {
				continue
			}

			newTarget.X = float32(target.X+i) + 0.5
			newTarget.Y = float32(target.Y+j) + 0.5

			if this.IsValidPosition(modules, newTarget.X, newTarget.Y, mapcollision.MOVE_NORMAL, mapcollision.COLLIDE_NORMAL) || ignoreBlocked {
				validTiles = append(validTiles, newTarget)
			}
		}
	}
	if len(validTiles) != 0 {
		return validTiles[rand.Intn(100)%len(validTiles)]
	}

	return oldTarget
}

func (this *MapCollision) GetCollideType(isHero bool) int {
	if isHero {
		return mapcollision.COLLIDE_HERO
	}

	return mapcollision.COLLIDE_NORMAL
}
