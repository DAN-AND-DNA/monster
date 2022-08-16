package item

type SetBonusData struct {
	BonusData
	Requirement int
}

func ConstructSetBonusData() SetBonusData {
	sb := SetBonusData{
		BonusData: ConstructBonusData(),
	}

	return sb
}
