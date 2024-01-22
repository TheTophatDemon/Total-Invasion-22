package shaders

import (
	_ "embed"
	"fmt"
	"log"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type UniformTypes interface {
	int | bool | float32 | mgl32.Mat4 | mgl32.Vec3 | mgl32.Vec4
}

type IUniform interface {
	UniformName() string
}

type Uniform[T UniformTypes] struct {
	name string
}

func (u Uniform[T]) UniformName() string {
	return u.name
}

var (
	UniformModelMatrix  Uniform[mgl32.Mat4] = Uniform[mgl32.Mat4]{"uModelMatrix"}
	UniformViewMatrix   Uniform[mgl32.Mat4] = Uniform[mgl32.Mat4]{"uViewMatrix"}
	UniformProjMatrix   Uniform[mgl32.Mat4] = Uniform[mgl32.Mat4]{"uProjMatrix"}
	UniformFogStart     Uniform[float32]    = Uniform[float32]{"uFogStart"}
	UniformFogLength    Uniform[float32]    = Uniform[float32]{"uFogLength"}
	UniformTex          Uniform[int]        = Uniform[int]{"uTex"}
	UniformLightDir     Uniform[mgl32.Vec3] = Uniform[mgl32.Vec3]{"uLightDir"}
	UniformAmbientColor Uniform[mgl32.Vec3] = Uniform[mgl32.Vec3]{"uAmbientColor"}
	UniformDiffuseColor Uniform[mgl32.Vec4] = Uniform[mgl32.Vec4]{"uDiffuseColor"}
	UniformSrcRect      Uniform[mgl32.Vec4] = Uniform[mgl32.Vec4]{"uSourceRect"}
	UniformFlipHorz     Uniform[bool]       = Uniform[bool]{"uFlipHorz"}
	UniformNoTexture    Uniform[bool]       = Uniform[bool]{"uNoTexture"}
)

var (
	MapShader *Shader
	//go:embed embed/map.vs.glsl
	mapVertShaderSrc string
	//go:embed embed/map.fs.glsl
	mapFragShaderSrc string

	DebugShader *Shader
	//go:embed embed/debug.vs.glsl
	debugVertShaderSrc string
	//go:embed embed/debug.fs.glsl
	debugFragShaderSrc string

	SpriteShader *Shader
	//go:embed embed/sprite.vs.glsl
	spriteVertShaderSrc string
	//go:embed embed/sprite.fs.glsl
	spriteFragShaderSrc string

	UIShader *Shader
	//go:embed embed/ui_box.vs.glsl
	uiVertShaderSrc string
	//go:embed embed/ui_box.fs.glsl
	uiFragShaderSrc string
)

type Shader struct {
	id          uint32
	uniformLocs map[IUniform]int32
}

// Initialize built-in shaders.
func Init() {
	var err error

	MapShader, err = CreateShader(mapVertShaderSrc, mapFragShaderSrc)
	if err != nil {
		log.Fatalln("Couldn't compile map shader: ", err)
	}

	DebugShader, err = CreateShader(debugVertShaderSrc, debugFragShaderSrc)
	if err != nil {
		log.Fatalln("Couldn't compile debug shader: ", err)
	}

	SpriteShader, err = CreateShader(spriteVertShaderSrc, spriteFragShaderSrc)
	if err != nil {
		log.Fatalln("Couldn't compile sprite shader: ", err)
	}

	UIShader, err = CreateShader(uiVertShaderSrc, uiFragShaderSrc)
	if err != nil {
		log.Fatalln("Couldn't compile UI shader: ", err)
	}
}

// Free built-in shaders.
func Free() {
	MapShader.Free()
	DebugShader.Free()
	SpriteShader.Free()
	UIShader.Free()
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
		id:          program,
		uniformLocs: make(map[IUniform]int32),
	}
	return shader, nil
}

func (s *Shader) Free() {
	gl.DeleteProgram(s.id)
}

func (s *Shader) Use() {
	gl.UseProgram(s.id)
}

func (s *Shader) getUniformLoc(u IUniform) (int32, error) {
	loc, ok := s.uniformLocs[u]
	if !ok {
		loc = gl.GetUniformLocation(s.id, gl.Str(u.UniformName()+"\x00"))
		s.uniformLocs[u] = loc
	}
	if loc < 0 {
		return loc, fmt.Errorf("uniform not found: %s", u.UniformName())
	}
	return loc, nil
}

func (s *Shader) SetUniformInt(u Uniform[int], val int) error {
	loc, err := s.getUniformLoc(u)
	if err != nil {
		return err
	}
	gl.Uniform1i(loc, int32(val))
	return nil
}

func (s *Shader) SetUniformBool(u Uniform[bool], val bool) error {
	loc, err := s.getUniformLoc(u)
	if err != nil {
		return err
	}
	if val {
		gl.Uniform1i(loc, 1)
	} else {
		gl.Uniform1i(loc, 0)
	}
	return nil
}

func (s *Shader) SetUniformFloat(u Uniform[float32], val float32) error {
	loc, err := s.getUniformLoc(u)
	if err != nil {
		return err
	}
	gl.Uniform1f(loc, val)
	return nil
}

func (s *Shader) SetUniformMatrix(u Uniform[mgl32.Mat4], val mgl32.Mat4) error {
	loc, err := s.getUniformLoc(u)
	if err != nil {
		return err
	}
	gl.UniformMatrix4fv(loc, 1, false, &val[0])
	return nil
}

func (s *Shader) SetUniformVec3(u Uniform[mgl32.Vec3], val mgl32.Vec3) error {
	loc, err := s.getUniformLoc(u)
	if err != nil {
		return err
	}
	gl.Uniform3fv(loc, 1, &val[0])
	return nil
}

func (s *Shader) SetUniformVec4(u Uniform[mgl32.Vec4], val mgl32.Vec4) error {
	loc, err := s.getUniformLoc(u)
	if err != nil {
		return err
	}
	gl.Uniform4fv(loc, 1, &val[0])
	return nil
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
