package evil

var lib uintptr

func Library() uintptr {
	if lib == 0 {
		var err error
		lib, err = openLibrary()
		if err != nil {
			panic(err)
		}
	}
	return lib
}
