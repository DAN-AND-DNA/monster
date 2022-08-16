package combattext

var (
	defaultComb = New()
)

type CombatText struct {
}

func New() *CombatText {
	return &CombatText{}
}
