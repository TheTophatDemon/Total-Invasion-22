package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func main() {
	const sourceDir = "./engine/tdaudio/c/"
	// Compile to object files
	cmd := exec.Command("gcc", "-c", sourceDir+"td_audio.c", sourceDir+"stb_vorbis.c")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	var libExt string
	switch runtime.GOOS {
	case "windows":
		libExt = "dll"
	case "linux":
		libExt = "so"
	default:
		panic(fmt.Sprintf("unsupported OS %v", runtime.GOOS))
	}

	// Link into library file
	cmd = exec.Command("gcc", "-shared", "-o", fmt.Sprintf("tdaudio.%v", libExt), "td_audio.o", "stb_vorbis.o")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
	fmt.Println("compilation succeeded")

	// Delete object files
	err = os.Remove("td_audio.o")
	if err != nil {
		panic(err)
	}
	err = os.Remove("stb_vorbis.o")
	if err != nil {
		panic(err)
	}
}
