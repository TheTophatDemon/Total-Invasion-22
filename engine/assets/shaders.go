package assets

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
)

type Shader struct {
	id       uint32
	uniforms map[string]int32
	attribs  map[string]int32
}

func CreateShader(vertSrc, fragSrc string) (*Shader, error) {
	vertShader, err := compileShader(vertSrc, gl.VERTEX_SHADER)
	if err != nil {
		return nil, fmt.Errorf("vertex shader error: %v", err)
	}
	fragShader, err := compileShader(fragSrc, gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, fmt.Errorf("fragment shader error: %v", err)
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, vertShader)
	gl.AttachShader(program, fragShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return nil, fmt.Errorf("failed to link shader program")
	}

	gl.BindFragDataLocation(program, 0, gl.Str("oColor\x00"))

	gl.DeleteShader(vertShader)
	gl.DeleteShader(fragShader)

	shader := &Shader{
		id:       program,
		uniforms: make(map[string]int32),
		attribs:  make(map[string]int32),
	}
	return shader, nil
}

func (s *Shader) Free() {
	gl.DeleteProgram(s.id)
}

func (s *Shader) Use() {
	gl.UseProgram(s.id)
}

func (s *Shader) GetUniformLoc(name string) int32 {
	loc, ok := s.uniforms[name]
	if !ok {
		loc = gl.GetUniformLocation(s.id, gl.Str(name+"\x00"))
		s.uniforms[name] = loc
	}
	return loc
}

func (s *Shader) GetAttribLoc(name string) int32 {
	loc, ok := s.attribs[name]
	if !ok {
		loc = gl.GetAttribLocation(s.id, gl.Str(name+"\x00"))
		s.attribs[name] = loc
	}
	return loc
}

func compileShader(src string, sType uint32) (uint32, error) {
	shader := gl.CreateShader(sType)

	//Make sure it's null terminated.
	if !strings.HasSuffix(src, "\x00") {
		src += "\x00"
	}
	csources, free := gl.Strs(src)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("%v", log)
	}

	return shader, nil
}
