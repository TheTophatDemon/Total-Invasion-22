package rgfw

import "github.com/ebitengine/purego"

func openLibrary() (uintptr, error) {
	return purego.Dlopen("./libRGFW.so", purego.RTLD_NOW|purego.RTLD_GLOBAL)
}
