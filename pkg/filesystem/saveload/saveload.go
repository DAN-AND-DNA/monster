package saveload

import "monster/pkg/common"

var (
	defaultSaveLoad = new()
)

type SaveLoad struct {
	gameSlot int
}

func new() *SaveLoad {
	return &SaveLoad{}
}

func (this *SaveLoad) saveGame() {
	if this.gameSlot <= 0 {
		return
	}

}

func (this *SaveLoad) createSaveDir(slot int, settings common.Settings) error {

	if slot == 0 {
		return nil
	}

	return nil
}
