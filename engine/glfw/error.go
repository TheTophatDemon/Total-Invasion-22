/*
This file is modified from github.com/go-gl/glfw to work with my slapdash purego bindings.
Their license notice is reproduced here:

Copyright (c) 2012 The glfw3-go Authors. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

  - Redistributions of source code must retain the above copyright

notice, this list of conditions and the following disclaimer.
  - Redistributions in binary form must reproduce the above

copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
  - Neither the name of Google Inc. nor the names of its

contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/
package glfw

import (
	"fmt"
	"log"
	"os"
	"unsafe"

	"github.com/ebitengine/purego"
)

// ErrorCode corresponds to an error code.
type ErrorCode int

// Error codes that are translated to panics and the programmer should not
// expect to handle.
const (
	notInitialized   ErrorCode = 0x00010001 // GLFW has not been initialized.
	noCurrentContext ErrorCode = 0x00010002 // No context is current.
	invalidEnum      ErrorCode = 0x00010003 // One of the enum parameters for the function was given an invalid enum.
	invalidValue     ErrorCode = 0x00010004 // One of the parameters for the function was given an invalid value.
	outOfMemory      ErrorCode = 0x00010005 // A memory allocation failed.
	platformError    ErrorCode = 0x00010008 // A platform-specific error occurred that does not match any of the more specific categories.
	noWindowContext  ErrorCode = 0x0001000A // A window that does not have an OpenGL or OpenGL ES context was passed to a function that requires it to have one.
)

const (
	// APIUnavailable is the error code used when GLFW could not find support
	// for the requested client API on the system.
	//
	// The installed graphics driver does not support the requested client API,
	// or does not support it via the chosen context creation backend. Below
	// are a few examples.
	//
	// Some pre-installed Windows graphics drivers do not support OpenGL. AMD
	// only supports OpenGL ES via EGL, while Nvidia and Intel only supports it
	// via a WGL or GLX extension. OS X does not provide OpenGL ES at all. The
	// Mesa EGL, OpenGL and OpenGL ES libraries do not interface with the
	// Nvidia binary driver.
	APIUnavailable ErrorCode = 0x00010006

	// VersionUnavailable is the error code used when the requested OpenGL or
	// OpenGL ES (including any requested profile or context option) is not
	// available on this machine.
	//
	// The machine does not support your requirements. If your application is
	// sufficiently flexible, downgrade your requirements and try again.
	// Otherwise, inform the user that their machine does not match your
	// requirements.
	//
	// Future invalid OpenGL and OpenGL ES versions, for example OpenGL 4.8 if
	// 5.0 comes out before the 4.x series gets that far, also fail with this
	// error and not GLFW_INVALID_VALUE, because GLFW cannot know what future
	// versions will exist.
	VersionUnavailable ErrorCode = 0x00010007

	// FormatUnavailable is the error code used for both window creation and
	// clipboard querying format errors.
	//
	// If emitted during window creation, the requested pixel format is not
	// supported. This means one or more hard constraints did not match any of
	// the available pixel formats. If your application is sufficiently
	// flexible, downgrade your requirements and try again. Otherwise, inform
	// the user that their machine does not match your requirements.
	//
	// If emitted when querying the clipboard, the contents of the clipboard
	// could not be converted to the requested format. You should ignore the
	// error or report it to the user, as appropriate.
	FormatUnavailable ErrorCode = 0x00010009
)

func (e ErrorCode) String() string {
	switch e {
	case notInitialized:
		return "NotInitialized"
	case noCurrentContext:
		return "NoCurrentContext"
	case invalidEnum:
		return "InvalidEnum"
	case invalidValue:
		return "InvalidValue"
	case outOfMemory:
		return "OutOfMemory"
	case platformError:
		return "PlatformError"
	case noWindowContext:
		return "NoWindowContext"
	case APIUnavailable:
		return "APIUnavailable"
	case VersionUnavailable:
		return "VersionUnavailable"
	case FormatUnavailable:
		return "FormatUnavailable"
	default:
		return fmt.Sprintf("ErrorCode(%d)", e)
	}
}

// Error holds error code and description.
type Error struct {
	Code ErrorCode
	Desc string
}

// Error prints the error code and description in a readable format.
func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code.String(), e.Desc)
}

// Note: There are many cryptic caveats to proper error handling here.
// See: https://github.com/go-gl/glfw3/pull/86

// Holds the value of the last error.
var lastError = make(chan *Error, 1)

func setErrorCallback() {
	glfwFuncs.SetErrorCallback(purego.NewCallback(func(code ErrorCode, descPtr *byte) int {
		flushErrors()
		desc := string(unsafe.Slice(descPtr, 1024))
		err := &Error{code, desc}
		select {
		case lastError <- err:
		default:
			fmt.Fprintln(os.Stderr, "go-gl/glfw: internal error: an uncaught error has occurred:", err)
			fmt.Fprintln(os.Stderr, "go-gl/glfw: Please report this in the Go package issue tracker.")
		}
		return 1
	}))
}

// flushErrors is called by Terminate before it actually calls C.glfwTerminate,
// this ensures that any uncaught errors buffered in lastError are printed
// before the program exits.
func flushErrors() {
	err := fetchError()
	if err != nil {
		fmt.Fprintln(os.Stderr, "go-gl/glfw: internal error: an uncaught error has occurred:", err)
		fmt.Fprintln(os.Stderr, "go-gl/glfw: Please report this in the Go package issue tracker.")
	}
}

// acceptError fetches the next error from the error channel, it accepts only
// errors with one of the given error codes. If any other error is encountered,
// a panic will occur.
//
// Platform errors are always printed, for information why please see:
//
//	https://github.com/go-gl/glfw/issues/127
func acceptError(codes ...ErrorCode) error {
	// Grab the next error, if there is one.
	err := fetchError()
	if err == nil {
		return nil
	}

	// Only if the error has the specific error code accepted by the caller, do
	// we return the error.
	for _, code := range codes {
		if err.Code == code {
			return err
		}
	}

	// The error isn't accepted by the caller. If the error code is not a code
	// defined in the GLFW C documentation as a programmer error, then the
	// caller should have accepted it. This is effectively a bug in this
	// package.
	switch err.Code {
	case platformError:
		log.Println(err)
		return nil
	case notInitialized, noCurrentContext, invalidEnum, invalidValue, outOfMemory, noWindowContext:
		panic(err)
	default:
		fmt.Fprintln(os.Stderr, "go-gl/glfw: internal error: an invalid error was not accepted by the caller:", err)
		fmt.Fprintln(os.Stderr, "go-gl/glfw: Please report this in the Go package issue tracker.")
		panic(err)
	}
}

// panicError is a helper used by functions which expect no errors (except
// programmer errors) to occur. It will panic if it finds any such error.
func panicError() {
	err := acceptError()
	if err != nil {
		panic(err)
	}
}

// fetchError fetches the next error from the error channel, it does not block
// and returns nil if there is no error present.
func fetchError() *Error {
	select {
	case err := <-lastError:
		return err
	default:
		return nil
	}
}
