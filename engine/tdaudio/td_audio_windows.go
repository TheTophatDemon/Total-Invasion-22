package tdaudio

import "golang.org/x/sys/windows"

func openLibrary() (uintptr, error) {
	dll := windows.NewLazyDLL("tdaudio.dll")
	err := dll.Load()
	return dll.Handle(), err
}
