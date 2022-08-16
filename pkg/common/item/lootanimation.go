package item

type LootAnimation struct {
	Name  string
	Low   int
	Hight int
}

func ConstructLootAnimation() LootAnimation {
	return LootAnimation{}
}
