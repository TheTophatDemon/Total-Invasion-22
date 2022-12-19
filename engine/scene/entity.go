package scene

const ID_MASK = 0x00000000FFFFFFFF
const ENT_INVALID = 0xFFFFFFFFFFFFFFFF

type Entity uint64

func (ent Entity) ID() uint64 {
	return uint64(ent) & ID_MASK
}
