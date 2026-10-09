package glfw

import "golang.org/x/sys/windows"

func openLibrary() (uintptr, error) {
	dll := windows.NewLazyDLL("glfw3.dll")
	err := dll.Load()
	return dll.Handle(), err
}
