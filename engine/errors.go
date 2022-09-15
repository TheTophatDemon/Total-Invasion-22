package engine

import (
	"log"
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
)

func CheckOpenGLError() {
	if err := gl.GetError(); err != gl.NO_ERROR {
		errName := "Unknown"
		switch err {
		case gl.INVALID_ENUM:
			errName = "Invalid enum"
		case gl.INVALID_VALUE:
			errName = "Invalid value"
		case gl.INVALID_OPERATION:
			errName = "Invalid operation"
		case gl.STACK_OVERFLOW:
			errName = "Stack overflow"
		case gl.STACK_UNDERFLOW:
			errName = "Stack underflow"
		case gl.OUT_OF_MEMORY:
			errName = "Out of memory"
		case gl.INVALID_FRAMEBUFFER_OPERATION:
			errName = "Invalid framebuffer operation"
		case gl.CONTEXT_LOST:
			errName = "Context lost"
		}
		_, fileName, lineNum, _ := runtime.Caller(1)
		log.Printf("**OPENGL ERROR** (%s, %v): %s", fileName, lineNum, errName)
	}
}
