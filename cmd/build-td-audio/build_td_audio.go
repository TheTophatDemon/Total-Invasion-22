package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func main() {
	var libExt string
	compileArgs := []string{"-c"}
	switch runtime.GOOS {
	case "windows":
		libExt = "dll"
	case "linux":
		libExt = "so"
		compileArgs = append(compileArgs, "-fPIC")
	default:
		panic(fmt.Sprintf("unsupported OS %v", runtime.GOOS))
	}

	const sourceDir = "./engine/tdaudio/c/"

	compileArgs = append(compileArgs, sourceDir+"td_audio.c", sourceDir+"stb_vorbis.c")

	// Compile to object files
	cmd := exec.Command("gcc", compileArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
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
