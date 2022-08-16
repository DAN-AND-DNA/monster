package mapcollision

// 碰撞检查类型
const (
	CHECK_MOVEMENT = iota
	CHECK_SIGHT
)

// 实体的碰撞类型
const (
	COLLIDE_NORMAL = iota
	COLLIDE_HERO
	COLLIDE_NO_ENTITY
)

// 移动类型
const (
	MOVE_NORMAL     = iota
	MOVE_FLYING     // can move through BLOCKS_MOVEMENT (e.g. water)
	MOVE_INTANGIBLE // can move through BLOCKS_ALL (e.g. walls)
)

// 瓷砖的碰撞类型
// 0到6为tiled 生成
// 7和8为地图上的实体进行管理
const (
	BLOCKS_NONE = iota
	BLOCKS_ALL
	BLOCKS_MOVEMENT
	BLOCKS_ALL_HIDDEN
	BLOCKS_MOVEMENT_HIDDEN
	MAP_ONLY
	MAP_ONLY_ALT
	BLOCKS_ENTITIES // hero or enemies are blocking this tile, so any other entity is blocked
	BLOCKS_ENEMIES  // an ally is standing on that tile, so the hero could pass if ENABLE_ALLY_COLLISION is false
)
