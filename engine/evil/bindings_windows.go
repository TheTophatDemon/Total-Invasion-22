package evil

import "golang.org/x/sys/windows"

func openLibrary() (uintptr, error) {
	dll := windows.NewLazyDLL("evil_engine.dll")
	err := dll.Load()
	return dll.Handle(), err
}
