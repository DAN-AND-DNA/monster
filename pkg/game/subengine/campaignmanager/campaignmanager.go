package campaignmanager

import (
	"monster/pkg/common/define"
	"monster/pkg/common/gameres"
	"monster/pkg/utils/tools"
)

type StatusPair struct {
	first  bool
	second string
}

type CampaignManager struct {
	bonusXP float32

	status map[define.StatusId]*StatusPair
}

func New() *CampaignManager {
	cm := &CampaignManager{}
	cm.init()

	return cm
}

func (this *CampaignManager) init() gameres.CampaignManager {
	this.status = map[define.StatusId]*StatusPair{}
	return this
}

func (this *CampaignManager) Close() {
}

// 注册
func (this *CampaignManager) RegisterStatus(s string) define.StatusId {
	if s == "" {
		return 0
	}

	newId := (define.StatusId)(tools.HashString(s))

	if _, ok := this.status[newId]; ok {
		return newId
	}

	this.status[newId] = &StatusPair{false, s}

	return newId
}

func (this *CampaignManager) checkStatus(s define.StatusId) bool {
	if ptr, ok := this.status[s]; ok && ptr.first {
		return true
	}

	return false
}

// 启动、设置
func (this *CampaignManager) SetStatus(s define.StatusId) {
	if this.checkStatus(s) {
		return
	}

	this.status[s].first = true

	// TODO
	// pc check title
}

func (this *CampaignManager) ResetAllStatuses() {
	for _, ptr := range this.status {
		ptr.first = false
	}
}
