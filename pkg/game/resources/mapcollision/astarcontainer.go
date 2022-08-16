package mapcollision

import "monster/pkg/common/point"

type AstarContainer struct {
	size      uint // 当前节点数
	nodeLimit uint // 节点上限
	mapWidth  uint
	mapHeight uint
	nodes     []*AstarNode
	mapPos    [][]int16 // 先宽 后高
}

func newAstarContainer(mapWidth, mapHeight, nodeLimit uint) *AstarContainer {
	c := &AstarContainer{}
	c.init(mapWidth, mapHeight, nodeLimit)

	return c
}

func (this *AstarContainer) init(mapWidth, mapHeight, nodeLimit uint) {
	this.size = 0
	this.nodeLimit = nodeLimit
	this.mapWidth = mapWidth
	this.mapHeight = mapHeight
	this.nodes = make([]*AstarNode, nodeLimit)
	this.mapPos = make([][]int16, mapWidth)

	for i := (uint)(0); i < mapWidth; i++ {
		tmp := make([]int16, mapHeight)
		for j := (uint)(0); j < mapHeight; j++ {
			tmp[j] = -1 // 表示没有节点
		}

		this.mapPos[i] = tmp
	}
}

func (this *AstarContainer) Close() {
	for _, ptr := range this.nodes {
		ptr.Close()
	}

	this.nodes = nil
}

// 获得当前节点数
func (this *AstarContainer) GetSize() int {
	return (int)(this.size)
}

// 添加节点
func (this *AstarContainer) Add(node *AstarNode) {
	// 达到上限
	if this.size >= this.nodeLimit {
		return
	}

	this.nodes[this.size] = node
	this.mapPos[node.GetX()][node.GetY()] = int16(this.size) // 写入编号

	m := this.size

	var temp *AstarNode
	for m != 0 {
		if this.nodes[m].GetFinalCost() <= this.nodes[m/2].GetFinalCost() {
			// f代表开销
			// 序号越小为父节点，其开销越小
			// 当前序号的f小于其父的f就交换节点，保证其父节点小于其2个子节点的开销
			temp = this.nodes[m/2]
			this.nodes[m/2] = this.nodes[m]
			this.mapPos[this.nodes[m/2].GetX()][this.nodes[m/2].GetY()] = (int16)(m / 2)
			this.nodes[m] = temp
			this.mapPos[this.nodes[m].GetX()][this.nodes[m].GetY()] = (int16)(m)
			m = m / 2
		} else {
			break
		}
	}

	this.size++
}

func (this *AstarContainer) GetShortestF() *AstarNode {
	return this.nodes[0]
}

func (this *AstarContainer) Remove(node *AstarNode) {
	rawHeapIndexv := this.mapPos[node.GetX()][node.GetY()] // 获得在地图上的序号
	if rawHeapIndexv < 0 {
		panic("bad index in map pos")
	}

	heapIndexv := (uint)(rawHeapIndexv) + 1 // 要删除的序号+1

	// 要删除的节点和最后一个节点（序号最大）进行交换
	this.nodes[heapIndexv-1] = this.nodes[this.size-1]
	this.mapPos[this.nodes[heapIndexv-1].GetX()][this.nodes[heapIndexv-1].GetY()] = (int16)(heapIndexv) - 1

	this.size--

	// 当前节点数为0, 直接返回
	if this.size == 0 {
		this.mapPos[node.GetX()][node.GetY()] = -1
		return
	}

	// 重新排序
	for {
		// 从替换的节点到最后一个节点进行重新排序
		heapIndexu := heapIndexv
		//         0
		//     1       2
		//   3   4   5   6
		// 转化成序号，id为替换的起点
		// 2 * id + 1 <= len - 2
		if 2*heapIndexu+1 <= this.size {
			// 2个子节点

			// 选择最低开销的作为新的父节点
			if this.nodes[heapIndexu-1].GetFinalCost() >= this.nodes[2*heapIndexu-1].GetFinalCost() {
				heapIndexv = 2 * heapIndexu
			}

			if this.nodes[heapIndexv-1].GetFinalCost() >= this.nodes[2*heapIndexu].GetFinalCost() {
				heapIndexv = 2*heapIndexu + 1
			}

			// 转化成序号，id为替换的起点
			// 2 * id <= len - 2
		} else if 2*heapIndexu <= this.size {
			// 1个节点
			if this.nodes[heapIndexu-1].GetFinalCost() >= this.nodes[2*heapIndexu-1].GetFinalCost() {
				heapIndexv = 2 * heapIndexu
			}
		}

		// 父节点的开销大于子的，进行交换
		if heapIndexu != heapIndexv {
			temp := this.nodes[heapIndexu-1]
			this.nodes[heapIndexu-1] = this.nodes[heapIndexv-1]
			this.mapPos[this.nodes[heapIndexu-1].GetX()][this.nodes[heapIndexu-1].GetY()] = int16(heapIndexu) - 1
			this.nodes[heapIndexv-1] = temp
			this.mapPos[this.nodes[heapIndexv-1].GetX()][this.nodes[heapIndexv-1].GetY()] = int16(heapIndexv) - 1
		} else {
			break
		}
	}

	// 在地图上删除该序号
	this.mapPos[node.GetX()][node.GetY()] = -1
}

func (this *AstarContainer) Exists(pos point.Point) bool {
	if pos.X < 0 || pos.Y < 0 || pos.X >= (int)(this.mapWidth) || pos.Y >= (int)(this.mapHeight) {
		return false
	}

	return this.mapPos[pos.X][pos.Y] != -1
}

func (this *AstarContainer) Get(x, y int) (*AstarNode, bool) {
	if x < 0 || y < 0 || x >= int(this.mapWidth) || y >= int(this.mapHeight) || this.mapPos[x][y] < 0 {
		return nil, false
	}

	return this.nodes[this.mapPos[x][y]], true
}

func (this *AstarContainer) IsEmpty() bool {
	return this.size == 0
}

func (this *AstarContainer) UpdateParent(pos, parentPos point.Point, score float32) {
	_, ok := this.Get(parentPos.X, parentPos.Y)
	if !ok {
		return
	}

	ptr, ok := this.Get(pos.X, pos.Y)
	if !ok {
		return
	}

	ptr.SetParent(parentPos) // 更新从那里来而已
	ptr.SetActualCost(score)

	// 重新排序
	m := this.mapPos[pos.X][pos.Y]
	var temp *AstarNode

	for m != 0 {

		// 子节点的开销小于父节点，则交换
		if this.nodes[m].GetFinalCost() <= this.nodes[m/2].GetFinalCost() {
			temp = this.nodes[m/2]
			this.nodes[m/2] = this.nodes[m]
			this.mapPos[this.nodes[m/2].GetX()][this.nodes[m/2].GetY()] = m / 2
			this.nodes[m] = temp
			this.mapPos[this.nodes[m].GetX()][this.nodes[m].GetY()] = m
			m = m / 2

		} else {
			break
		}
	}
}
