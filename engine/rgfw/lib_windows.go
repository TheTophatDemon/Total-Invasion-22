package rgfw

import "golang.org/x/sys/windows"

func openLibrary() (uintptr, error) {
	dll := windows.NewLazyDLL("libRGFW.dll")
	err := dll.Load()
	return dll.Handle(), err
}
