package engine

type HasDefault interface {
	InitDefault()
}

type HasFinalizer interface {
	Finalize()
}
