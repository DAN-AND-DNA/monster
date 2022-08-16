package mapcollision

import (
	"math"
	"monster/pkg/common/point"
)

type AstarCloseContainer struct {
	size      uint
	nodeLimit uint
	mapWidth  uint
	mapHeight uint

	nodes  []*AstarNode
	mapPos [][]int16
}

func newAstarCloseContainer(mapWith, mapHeight, nodeLimit uint) *AstarCloseContainer {
	cc := &AstarCloseContainer{}
	cc.init(mapWith, mapHeight, nodeLimit)

	return cc
}

func (this *AstarCloseContainer) init(mapWidth, mapHeight, nodeLimit uint) {
	this.size = 0
	this.nodeLimit = nodeLimit
	this.mapWidth = mapWidth
	this.mapHeight = mapHeight

	this.nodes = make([]*AstarNode, nodeLimit)
	for i := (uint)(0); i < mapWidth; i++ {
		tmp := make([]int16, mapHeight)
		for j := (uint)(0); j < mapHeight; j++ {
			tmp[j] = -1 // 表示没有节点
		}

		this.mapPos[i] = tmp
	}
}

func (this *AstarCloseContainer) Close() {
	for _, ptr := range this.nodes {
		ptr.Close()
	}

	this.nodes = nil
}

func (this *AstarCloseContainer) GetSize() int {
	return (int)(this.size)
}

func (this *AstarCloseContainer) Add(node *AstarNode) {
	if this.size >= this.nodeLimit {
		return
	}

	this.nodes[this.size] = node
	this.mapPos[node.GetX()][node.GetY()] = (int16)(this.size)
	this.size++
}

func (this *AstarCloseContainer) Exists(pos point.Point) bool {
	if pos.X < 0 || pos.Y < 0 || pos.X >= (int)(this.mapWidth) || pos.Y >= (int)(this.mapHeight) {
		return false
	}

	return this.mapPos[pos.X][pos.Y] != -1
}

func (this *AstarCloseContainer) Get(x, y int) (*AstarNode, bool) {
	if x < 0 || y < 0 || x >= int(this.mapWidth) || y >= int(this.mapHeight) || this.mapPos[x][y] < 0 {
		return nil, false
	}

	return this.nodes[this.mapPos[x][y]], true
}

func (this *AstarCloseContainer) GetShortestH() (*AstarNode, bool) {
	var current *AstarNode
	lowestScore := math.MaxFloat32
	found := false
	for _, ptr := range this.nodes {
		if float64(ptr.GetH()) < lowestScore {
			lowestScore = (float64)(ptr.GetH())
			current = ptr
			found = true
		}
	}

	return current, found
}
