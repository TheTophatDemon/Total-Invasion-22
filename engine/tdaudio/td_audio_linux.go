package tdaudio

import "github.com/ebitengine/purego"

func openLibrary() (uintptr, error) {
	return purego.Dlopen("./tdaudio.so", purego.RTLD_NOW|purego.RTLD_GLOBAL)
}
