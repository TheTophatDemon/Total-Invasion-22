package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"slices"
)

func main() {
	var libExt string
	compileArgs := []string{"build", "./evil-engine", "-build-mode:shared", "-reloc-mode:pic"}
	if slices.Contains(os.Args, "debug") {
		compileArgs = append(compileArgs, "-debug")
	}
	switch runtime.GOOS {
	case "windows":
		libExt = "dll"
	case "linux":
		libExt = "so"
	default:
		panic(fmt.Sprintf("unsupported OS %v", runtime.GOOS))
	}

	compileArgs = append(compileArgs, "-out:evil_engine."+libExt)

	// Compile to object files
	cmd := exec.Command("odin", compileArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}
