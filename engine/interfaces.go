package engine

type HasDefault interface {
	InitDefault()
}

type HasFinalizer interface {
	Finalize()
}

type Observer interface {
	ProcessSignal(signal any)
}
