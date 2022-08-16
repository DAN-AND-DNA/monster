package gameres

import (
	"monster/pkg/common"
	"monster/pkg/common/color"
	"monster/pkg/common/define"
	"monster/pkg/common/define/game/stats"
	"monster/pkg/common/event"
	"monster/pkg/common/fpoint"
	"monster/pkg/common/gameres/avatar"
	"monster/pkg/common/gameres/effect"
	"monster/pkg/common/gameres/power"
	"monster/pkg/common/gameres/statblock"
	"monster/pkg/common/item"
	"monster/pkg/common/point"
	"monster/pkg/common/timer"
)

type GameRes interface {
	Stats() Stats
	NewStats(common.Modules) Stats
	Items() ItemManager
	NewItems(common.Modules, Stats) ItemManager
	Camp() CampaignManager
	NewCamp() CampaignManager
	EventManager() EventManager
	NewEventManager() EventManager
	Loot() LootManager
	NewLoot(common.Modules, ItemManager) LootManager
	Pc() Avatar
	NewPc(common.Modules, MapRenderer, Stats, PowerManager, Factory) Avatar
	Menu() MenuManager
	NewMenu(common.Modules, Avatar, PowerManager, Factory) MenuManager
	Mapr() MapRenderer
	NewMapr(common.Modules, Factory) MapRenderer
	Powers() PowerManager
	NewPowers(common.Modules, Stats) PowerManager

	// 工厂方法
	Menuf() Factory
	Resf() Factory
}

type Stats interface {
	Init(common.Modules)
	GetKey(stats.STAT) string
}
type Factory interface {
	New(string) interface{}
}

type ItemManager interface {
	GetItems() map[define.ItemId]item.Item
}

type Menu interface {
	Close()
	Clear()
	SetVisible(bool)
	GetVisible() bool
	Align(common.Modules) error
	Render(common.Modules) error
	Logic(common.Modules, Avatar, PowerManager) error
}

type MenuConfirm interface {
	Menu
	Init(modules common.Modules, buttonMsg, boxMsg string) MenuConfirm
}

type MenuConfig interface {
	Close()
	Init(common.Modules, bool) MenuConfig
	Logic(common.Modules) error
	Render(common.Modules) error
	SetForceRefreshBackground(bool)
	GetForceRefreshBackground() bool
	SetClickedAccept(bool)
	GetClickedAccept() bool
	SetClickedCancel(bool)
	GetClickedCancel() bool
	GetRenderDevice() string
	RefreshMods(common.Modules) (bool, error)
}

type MenuExit interface {
	Menu
	Init(common.Modules, Avatar) MenuExit
}

type StatBlock interface {
	Init(common.Modules, Factory) StatBlock
	Close()
	SetName(string)
	GetName() string
	GetGfxPortrait() string
	SetGfxPortrait(string)
	SetGfxHead(string)
	GetGfxHead() string
	SetGfxBase(string)
	GetGfxBase() string
	GetHero() bool
	SetHero(bool)
	SetCharacterClass(string)
	SetCharacterSubclass(string)
	GetPrimary(int) int
	SetPrimary(int, int)
	SetPrimaryStarting(int, int)
	SetPrimaryAdditional(int, int)
	GetPermadeath() bool
	SetPermadeath(bool)
	Recalc(common.Modules, Stats)
	SetDirection(uint8)
	GetDirection() uint8
	SetLevel(int)
	GetLevel() int
	GetLongClass(common.Modules) string
	SetPerfectAccuracy(bool)
	SetPos(fpoint.FPoint)
	GetPos() fpoint.FPoint
	SetStarting(index, val int)
	GetCorpse() bool
	GetCorpseTimer() *timer.Timer
	GetHeroAlly() bool
	GetEnemyAlly() bool
	SetEncountered(bool)
	GetEncountered() bool
	SetXp(uint64)
	GetXp() uint64
	SetHP(int)
	GetHP() int
	GetPrevHP() int
	SetMP(int)
	GetMP() int
	GetEffects() EffectManager
	SetTeleportation(bool)
	GetTeleportation() bool
	SetTeleportDestination(fpoint.FPoint)
	GetTeleportDestination() fpoint.FPoint
	Get(stats.STAT) int
	GetDamageMax(int) int
	GetDamageMin(int) int
	GetSpeedDefault() float32
	SetKnockbackSrcPos(fpoint.FPoint)
	SetKnockbackDestPos(fpoint.FPoint)
	AddPartyBuff(define.PowerId)
	SetTargetCorpse(StatBlock)
	GetTargetCorpse() StatBlock
	GetTargetNearest() StatBlock
	GetTargetNearestCorpse() StatBlock
	GetTargetNearestDist() float32
	GetTargetNearestCorpseDist() float32
	SetBlockPower(define.PowerId)
	GetBlockPower() define.PowerId
	TakeDamage(dmg int, crit bool, sourceType int)
	GetAlive() bool
	Logic(common.Modules, Avatar, CampaignManager)
	GetSpeed() float32
	SetSpeed(float32)
	GetChargeSpeed() float32
	GetMovementType() int
	SetCurState(statblock.EntityState)
	GetCurState() statblock.EntityState
	SetHumanoid(val bool)
	GetHumanoid() bool
	GetCooldown() *timer.Timer
	GetCooldownHit() *timer.Timer
	GetCooldownHitEnabled() bool
	GetTransformed() bool
	GetBlocking() bool
	SetRefreshStats(bool)
}

type GameSlotPreview interface {
	Init(common.Modules, GameRes) GameSlotPreview
	Close(common.Modules)
	SetStatBlock(StatBlock)
	GetLayerReferenceOrder() []string
	LoadGraphics(common.Modules, []string) error
	SetPos(point.Point)
	Render(common.Modules) error
	Logic()
	SetAnimation(name string)
}

type EffectManager interface {
	Init(common.Modules) EffectManager
	Close()
	GetBonusPrimary(int) int
	GetBonus(int) int
	GetBonusResist(int) int
	SetTriggeredDeath(bool)
	GetTriggeredDeath() bool
	AddEffect(modules common.Modules, def effect.Def, duration, magnitude, sourceType int, powerId define.PowerId)
	RemoveEffectId([]power.RemoveEffectPair)
	GetTriggeredBlock() bool
	SetTriggeredBlock(bool)
	DamageShields(int) int
	GetRevive() bool
	Logic(common.Modules)
	GetRefreshStats() bool
	SetRefreshStats(val bool)
	GetDamage() int
	GetDamagePercent() int
	GetDamageSourceType(int) int
	GetDeathSentence() bool
	GetStun() bool
	GetHPot() int
	GetMPot() int
	GetHPotPercent() int
	GetMPotPercent() int
	GetKnockbackSpeed() float32
	GetConvert() bool
	GetSpeed() int
	GetCurrentColor(color.Color) color.Color
	GetCurrentAlpha(uint8) uint8
	ClearTriggerEffects(trigger int)
}

type CampaignManager interface {
	Close()
	RegisterStatus(string) define.StatusId
	SetStatus(s define.StatusId)
	ResetAllStatuses()
}

type EventManager interface {
	LoadEvent(modules common.Modules, loot LootManager, camp CampaignManager, key, val string, evnt *event.Event) error
	ExecuteEvent(common.Modules, MapRenderer, CampaignManager, *event.Event) bool
	ExecteDelayedEvent(common.Modules, MapRenderer, CampaignManager, *event.Event) bool
	Close()
}

type LootManager interface {
	ParseLoot(common.Modules, string, *event.Component, []event.Component) []event.Component
	Close(common.Modules)
}

type MenuStatBar interface {
	Menu
	Init(common.Modules, uint16) MenuStatBar
	Update(statMin, statCur, statMax uint64)
}

type MenuInventory interface {
	Menu
	Init(common.Modules) MenuInventory
	SetChangedEquipment(bool)
	GetChangedEquipment() bool
	SetCurrency(int)
	GetCurrency() int
}

type MenuActionBar interface {
	Menu
	Init(common.Modules, PowerManager) MenuActionBar
}

type MenuManager interface {
	Close()
	AlignAll(common.Modules) error
	Render(common.Modules) error
	Get(name string) Menu
	Logic(common.Modules, Avatar, PowerManager)
	MenuAct() MenuActionBar
}

type MapCollision interface {
	Init() MapCollision
	SetMap(colMap [][]uint16, w, h uint16)
	IsEmpty(x, y float32) bool
	IsWall(x, y float32) bool
	GetRandomNeighbor(modules common.Modules, target point.Point, range1 int, ignoreBlocked bool) fpoint.FPoint
	Move(modules common.Modules, x, y, stepX, stepY float32, movementType, collideType int) (float32, float32, bool)
	GetCollideType(isHero bool) int
	Unblock(mapX, mapY float32)
	IsValidPosition(modules common.Modules, x, y float32, movementType, collideType int) bool
}

type MapCamera interface {
	SetTarget(fpoint.FPoint)
	WarpTo(fpoint.FPoint)
	GetPos() fpoint.FPoint
}

type Map interface {
	RegisterDelayedEvent(event.Event)
}

type MapRenderer interface {
	Map
	Clear()
	Close()
	SetTeleportation(bool)
	SetTeleportMapName(string)
	GetTeleportation() bool
	GetTeleportMapName() string
	SetTeleportDestination(fpoint.FPoint)
	GetTeleportDestination() fpoint.FPoint
	Load(modules common.Modules, loot LootManager, camp CampaignManager, eventManager EventManager, gresf Factory, fname string) error
	Render(modules common.Modules, r []common.Renderable, rDead []common.Renderable) error
	Logic(common.Modules)
	ExecuteOnLoadEvent(common.Modules, EventManager, CampaignManager)
	GetIsSpawnMap() bool
	GetCollider() MapCollision
	GetHeroPosEnabled() bool
	GetHeroPos() fpoint.FPoint
	GetCam() MapCamera
	GetFilename() string
}

type Entity interface {
	GetStats() StatBlock
	Clear(common.Modules)
	Close(common.Modules)
}

type Avatar interface {
	Entity
	GetTimePlayed() uint64
	Init(common.Modules, MapRenderer, Stats, PowerManager)
	GetLayerReferenceOrder() []string
	LoadGraphics(common.Modules, []avatar.LayerGfx) error
	Logic(common.Modules, MapRenderer, CampaignManager)
	AddRenders(modules common.Modules, r []common.Renderable) []common.Renderable
	GetPowerCastTimersSize() int
	GetPowerCastTimer(define.PowerId) *timer.Timer
	GetPowerCooldownTimer(define.PowerId) *timer.Timer
}

type PowerManager interface {
	VerifyId(powerId define.PowerId, allowZero bool) define.PowerId
	GetPower(define.PowerId) *power.Power
	GetPowers() map[define.PowerId]*power.Power
	Close()
}

type GameState interface {
	Clear(common.Modules, GameRes)
	Close(common.Modules, GameRes)
	SetLoadingFrame()
	RefreshWidgets(common.Modules, GameRes) error
	SetForceRefreshBackground(bool)
	GetForceRefreshBackground() bool
	ShowLoading(common.Modules) error
	GetHasBackground() bool
	Render(common.Modules, GameRes) error
	GetRequestedGameState() GameState
	GetReloadBackgrounds() bool
	IncrLoadCounter()
	DecrLoadCounter()
	GetLoadCounter() int
	GetExitRequested() bool
	Logic(common.Modules, GameRes) error
}

type GameStateLoad interface {
	GameState
}

type GameStateTitle interface {
	GameState
}

type GameStateConfig interface {
	GameState
}

type GameStateNewGame interface {
	GameState
}

type GameStatePlay interface {
	GameState
	ResetGame(common.Modules, GameRes)
}
