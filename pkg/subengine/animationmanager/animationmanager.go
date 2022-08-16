package animationmanager

import (
	"fmt"
	"monster/pkg/common"
)

type AnimationManager struct {
	sets   []common.AnimationSet
	names  []string
	counts []int
}

func New() *AnimationManager {
	return &AnimationManager{}
}

func (this *AnimationManager) Close() {
	this.CleanUp()
}

func (this *AnimationManager) GetAnimationSet(settings common.Settings, mods common.ModManager, render common.RenderDevice, resf common.Factory, filename string) (common.AnimationSet, error) {
	index := -1
	for i, val := range this.names {
		if val == filename {
			index = i
			break
		}
	}

	if index >= 0 {
		if this.sets[index] == nil {
			this.sets[index] = resf.New("animationset").(common.AnimationSet).Init(settings, mods, render, filename)

		}

		return this.sets[index], nil
	}

	return nil, fmt.Errorf("AnimationManager::getAnimationSet(): %s not found\n", filename)
}

func (this *AnimationManager) IncreaseCount(name string) {
	index := -1
	for i, val := range this.names {
		if val == name {
			index = i
			break
		}
	}

	if index >= 0 {
		val := this.counts[index]
		val++
		this.counts[index] = val
	} else {
		this.sets = append(this.sets, nil)
		this.names = append(this.names, name)
		this.counts = append(this.counts, 1)
	}
}

func (this *AnimationManager) DecreaseCount(name string) {
	index := -1
	for i, val := range this.names {
		if val == name {
			index = i
			break
		}
	}

	if index >= 0 {
		val := this.counts[index]
		val--
		this.counts[index] = val
	}
}

// 清理无用的动画集
func (this *AnimationManager) CleanUp() {
	i := len(this.sets) - 1

	for i >= 0 {
		if this.counts[i] <= 0 {
			if this.sets[i] != nil {
				this.sets[i].Close()
			}

			leftSets := this.sets[i+1:]
			this.sets = this.sets[:len(this.sets)-1]
			for index, val := range leftSets {
				this.sets[i+index] = val
			}

			leftCounts := this.counts[i+1:]
			this.counts = this.counts[:len(this.counts)-1]
			for index, val := range leftCounts {
				this.counts[i+index] = val
			}

			leftNames := this.names[i+1:]
			this.names = this.names[:len(this.names)-1]
			for index, val := range leftNames {
				this.names[i+index] = val
			}

			//delete(this.sets, i)
			//delete(this.counts, i)
			//delete(this.names, i)
		}

		i--
	}
}
