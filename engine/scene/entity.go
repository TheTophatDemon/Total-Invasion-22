package scene

type Entity uint64

const IDX_MASK uint64 = 0x00000000FFFFFFFF
const GEN_MASK uint64 = ^IDX_MASK
const GEN_INCREMENT uint64 = 0x0000000100000000
const ENT_INVALID Entity = Entity(IDX_MASK | GEN_MASK)

func (ent Entity) Index() uint64 {
	return uint64(ent) & IDX_MASK
}

func (ent Entity) Generation() uint64 {
	return uint64(ent) & GEN_MASK
}

// Returns an Entity ID with the same index, but increase the generation count.
func (ent Entity) Renew() Entity {
	return Entity(uint64(ent) + GEN_INCREMENT)
}
