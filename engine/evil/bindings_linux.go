package evil

import "github.com/ebitengine/purego"

func openLibrary() (uintptr, error) {
	return purego.Dlopen("./evil_engine.so", purego.RTLD_NOW|purego.RTLD_GLOBAL)
}
