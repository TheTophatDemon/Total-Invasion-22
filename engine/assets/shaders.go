package assets

import (
	"fmt"
	"log"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
)

type Shader struct {
	id       uint32
	uniforms map[string]int32
	attribs  map[string]int32
}

var (
	MapShader *Shader
)

//Initialize built-in shaders
func InitShaders() {
	var err error

	MapShader, err = CreateShader(`
		#version 330

		layout(location = 0) in vec3 aPos;
		layout(location = 1) in vec2 aTexCoord;
		layout(location = 2) in vec3 aNormal;

		uniform mat4 uMVP;
		uniform mat4 uModelTransform;

		out vec2 vTexCoord;
		out vec3 vNormal;

		void main() {
			vTexCoord = aTexCoord;
			mat3 rot = mat3(uModelTransform[0].xyz, uModelTransform[1].xyz, uModelTransform[2].xyz);
			vNormal = normalize(rot * aNormal);
			gl_Position = uMVP * vec4(aPos, 1);
		}
	`, `
		#version 330

		const vec3 LIGHT_DIR = normalize(vec3(1.0, 0.0, 1.0));
		const vec3 AMBIENT = vec3(0.5, 0.5, 0.5);

		in vec2 vTexCoord;
		in vec3 vNormal;

		uniform sampler2D uTex;
		uniform float uFogStart;
		uniform float uFogLength;

		out vec4 oColor;

		void main() {
			vec4 diffuse = texture(uTex, vTexCoord);
			if (diffuse.a < 0.5) {
				discard;
			}
			float lightFactor = (dot(-LIGHT_DIR, normalize(vNormal)) + 1.0) / 2.0;
			diffuse.rgb *= AMBIENT + (vec3(1.0) - AMBIENT) * lightFactor;
			float depth = gl_FragCoord.z / gl_FragCoord.w;
			float fog = 1.0 - clamp((depth - uFogStart) / uFogLength, 0.0, 1.0);
    		diffuse.rgb *= fog;
			oColor = diffuse;
		}
	`)
	if err != nil {
		log.Fatalln("Couldn't compile mapShader: ", err)
	}
}

func FreeShaders() {
	MapShader.Free()
}

func CreateShader(vertSrc, fragSrc string) (*Shader, error) {
	vertShader, err := compileShader(vertSrc, gl.VERTEX_SHADER)
	if err != nil {
		return nil, fmt.Errorf("Vertex shader error: %v", err)
	}
	fragShader, err := compileShader(fragSrc, gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, fmt.Errorf("Fragment shader error: %v", err)
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

		return nil, fmt.Errorf("Failed to link shader program")
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
