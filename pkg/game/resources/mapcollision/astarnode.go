package mapcollision

import "monster/pkg/common/point"

const (
	nodeStride = 1 // 节点间的最小间距
)

type AstarNode struct {
	x      int // 位置x轴
	y      int
	g      float32     // 和第一个节点的差
	h      float32     // 和最后一个节点的差
	parent point.Point // 从哪来
}

func newAstarNode(p point.Point) *AstarNode {
	node := &AstarNode{
		x:      p.X,
		y:      p.Y,
		parent: point.Construct(),
	}

	return node
}

func (this *AstarNode) Close() {
}

func (this *AstarNode) GetX() int {
	return this.x
}

func (this *AstarNode) GetY() int {
	return this.y
}

func (this *AstarNode) GetH() float32 {
	return this.h
}

func (this *AstarNode) GetParent() point.Point {
	return this.parent
}

func (this *AstarNode) SetParent(p point.Point) {
	this.parent = p
}

func (this *AstarNode) GetNeighbours(limitX, limitY int) []point.Point {
	var res []point.Point

	if this.x > nodeStride && this.y > nodeStride {
		res = append(res, point.Construct(this.x-nodeStride, this.y-nodeStride))
	}

	if this.x > nodeStride && (limitY == 0 || this.y < limitY-nodeStride) {
		res = append(res, point.Construct(this.x-nodeStride, this.y+nodeStride))
	}

	if this.y > nodeStride && (limitX == 0 || this.x < limitX-nodeStride) {
		res = append(res, point.Construct(this.x+nodeStride, this.y-nodeStride))
	}

	if (limitX == 0 || this.x < limitX-nodeStride) && (limitY == 0 || this.y < limitY-nodeStride) {
		res = append(res, point.Construct(this.x+nodeStride, this.y+nodeStride))
	}

	if this.x > nodeStride {
		res = append(res, point.Construct(this.x-nodeStride, this.y))
	}

	if this.y > nodeStride {
		res = append(res, point.Construct(this.x, this.y-nodeStride))
	}

	if limitX == 0 || this.x < limitX-nodeStride {
		res = append(res, point.Construct(this.x+nodeStride, this.y))
	}

	if limitY == 0 || this.y < limitY-nodeStride {
		res = append(res, point.Construct(this.x, this.y+nodeStride))
	}

	return res
}

func (this *AstarNode) GetActualCost() float32 {
	return this.g
}

func (this *AstarNode) SetActualCost(g float32) {
	this.g = g
}

func (this *AstarNode) SetEstimatedCost(h float32) {
	this.h = h
}

func (this *AstarNode) GetFinalCost() float32 {
	return this.g + this.h*2
}
