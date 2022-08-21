package ecs

//First 16 bits are the actual ID, the last 16 bits are the generation count.
//The generation count increases with each reuse of the ID to indicate that it's a different entity than the last one.
type EntID uint32

//ID of 0 signifies an absence of an entity, so storage containers can identify non-owned components.
const NO_ENT EntID = 0

const (
	ENT_IDX_MASK  = 0x0000FFFF
	ENT_GEN_MASK  = 0xFFFF0000
	GEN_INC       = 0x00010000 //Number to add to increase the generation by one (relative to the last 2 bytes)
)

const (
	ERR_INDEX =    "Entity index is out of bounds."
	ERR_OWNER =    "Entity does not own the component, or was overwritten."
	ERR_CTYPE =    "Given component is of the wrong type."
	ERR_REGISTER = "Component type was not registered."
)