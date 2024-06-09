package comps

import (
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
)

const (
	ATTR_INSTANCE_POS = iota + 8
	ATTR_INSTANCE_COL
	ATTR_INSTANCE_SIZE
)

var (
	positionByteSize = int(unsafe.Sizeof(mgl32.Vec3{}))
	colorByteSize    = int(unsafe.Sizeof(mgl32.Vec4{}))
	sizeByteSize     = int(unsafe.Sizeof(mgl32.Vec2{}))
)

type ParticleRender struct {
	Mesh       *geom.Mesh
	Texture    *textures.Texture
	AnimPlayer *AnimationPlayer
	Radius     float32 // The sphereical radius within which particles will be spawned.
	SpawnRate  float32 // The rate at which new particles will be spawned, in seconds per particle.

	spawnTimer float32

	// Particle instance fields
	positions                 []mgl32.Vec3
	positionBuffer            uint32
	colors                    []mgl32.Vec4
	colorBuffer               uint32
	sizes                     []mgl32.Vec2
	sizeBuffer                uint32
	velocities, accelerations []mgl32.Vec3
}

func NewParticleRender(
	mesh *geom.Mesh,
	texture *textures.Texture,
	anim *AnimationPlayer,
	maxInstances uint16,
) ParticleRender {
	parts := ParticleRender{
		Mesh:          mesh,
		Texture:       texture,
		AnimPlayer:    anim,
		positions:     make([]mgl32.Vec3, 0, maxInstances),
		colors:        make([]mgl32.Vec4, 0, maxInstances),
		sizes:         make([]mgl32.Vec2, 0, maxInstances),
		velocities:    make([]mgl32.Vec3, 0, maxInstances),
		accelerations: make([]mgl32.Vec3, 0, maxInstances),
	}

	// Position buffer
	gl.GenBuffers(1, &parts.positionBuffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, parts.positionBuffer)
	gl.BufferData(gl.ARRAY_BUFFER,
		cap(parts.positions)*positionByteSize,
		nil, gl.STREAM_DRAW)

	// Color buffer
	gl.GenBuffers(1, &parts.colorBuffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, parts.colorBuffer)
	gl.BufferData(gl.ARRAY_BUFFER,
		cap(parts.colors)*colorByteSize,
		nil, gl.STREAM_DRAW)

	// Size buffer
	gl.GenBuffers(1, &parts.sizeBuffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, parts.sizeBuffer)
	gl.BufferData(gl.ARRAY_BUFFER,
		cap(parts.sizes)*sizeByteSize,
		nil, gl.STREAM_DRAW)

	return parts
}

func (parts *ParticleRender) Update(deltaTime float32) {
	if parts.AnimPlayer != nil {
		parts.AnimPlayer.Update(deltaTime)
	}

	//TODO: Particle movement

	parts.spawnTimer += deltaTime
	if parts.spawnTimer > parts.SpawnRate {
		parts.spawnTimer = 0.0
		parts.SpawnParticle(math2.RandomDir().Mul(parts.Radius), color.White.Vector(), mgl32.Vec2{0.25, 0.25}, mgl32.Vec3{}, mgl32.Vec3{})
	}

	parts.updateBuffers()
}

func (parts *ParticleRender) Render(
	transform *Transform,
	context *render.Context,
) {
	if parts.Mesh == nil {
		return
	}

	// Set uniforms
	shader := shaders.SpriteShaderInstanced
	shader.Use()
	_ = context.SetUniforms(shader)
	_ = shader.SetUniformInt(shaders.UniformTex, 0)
	_ = shader.SetUniformMatrix(shaders.UniformModelMatrix, mgl32.Ident4())
	if parts.AnimPlayer != nil {
		_ = shader.SetUniformVec4(shaders.UniformSrcRect, parts.AnimPlayer.FrameUV().Vec4())
	} else {
		_ = shader.SetUniformVec4(shaders.UniformSrcRect, mgl32.Vec4{0.0, 0.0, 1.0, 1.0})
	}

	parts.Mesh.Bind()
	gl.VertexAttribDivisorARB(geom.ATTR_POS, 0)
	gl.VertexAttribDivisorARB(geom.ATTR_NORMAL, 0)
	gl.VertexAttribDivisorARB(geom.ATTR_COLOR, 0)
	gl.VertexAttribDivisorARB(geom.ATTR_TEXCOORD, 0)

	if parts.Texture != nil {
		parts.Texture.Bind()
	}
	parts.bind()
	gl.DrawArraysInstancedARB(gl.TRIANGLE_STRIP,
		0, int32(len(parts.Mesh.Inds())),
		int32(len(parts.positions)))
}

func (parts *ParticleRender) Finalize() {
	parts.Free()
}

func (parts *ParticleRender) SpawnParticle(position mgl32.Vec3, color mgl32.Vec4, size mgl32.Vec2, velocity, acceleration mgl32.Vec3) {
	if len(parts.positions) >= cap(parts.positions) {
		return
	}

	parts.positions = append(parts.positions, position)
	parts.colors = append(parts.colors, color)
	parts.sizes = append(parts.sizes, size)
	parts.velocities = append(parts.velocities, velocity)
	parts.accelerations = append(parts.accelerations, acceleration)
}

func (parts *ParticleRender) removeParticle(index int) {
	if index >= len(parts.positions) || index < 0 {
		return
	}

	lastIndex := len(parts.positions) - 1
	// Swap the latest particle with the one being removed in order to keep the list contiguous.
	parts.positions[index] = parts.positions[lastIndex]
	parts.colors[index] = parts.colors[lastIndex]
	parts.sizes[index] = parts.sizes[lastIndex]
	parts.velocities[index] = parts.velocities[lastIndex]
	parts.accelerations[index] = parts.accelerations[lastIndex]

	parts.positions = parts.positions[:lastIndex]
	parts.colors = parts.colors[:lastIndex]
	parts.sizes = parts.sizes[:lastIndex]
	parts.velocities = parts.velocities[:lastIndex]
	parts.accelerations = parts.accelerations[:lastIndex]
}

func (parts *ParticleRender) Free() {
	gl.DeleteBuffers(1, &parts.positionBuffer)
	gl.DeleteBuffers(1, &parts.colorBuffer)
	gl.DeleteBuffers(1, &parts.sizeBuffer)
}

func (parts *ParticleRender) bind() {
	gl.EnableVertexAttribArray(ATTR_INSTANCE_POS)
	gl.BindBuffer(gl.ARRAY_BUFFER, parts.positionBuffer)
	gl.VertexAttribPointerWithOffset(parts.positionBuffer, 3, gl.FLOAT, false, 0, 0)
	gl.VertexAttribDivisorARB(ATTR_INSTANCE_POS, 1)

	gl.EnableVertexAttribArray(ATTR_INSTANCE_COL)
	gl.BindBuffer(gl.ARRAY_BUFFER, parts.colorBuffer)
	gl.VertexAttribPointerWithOffset(parts.colorBuffer, 4, gl.FLOAT, false, 0, 0)
	gl.VertexAttribDivisorARB(ATTR_INSTANCE_COL, 1)

	gl.EnableVertexAttribArray(ATTR_INSTANCE_SIZE)
	gl.BindBuffer(gl.ARRAY_BUFFER, parts.sizeBuffer)
	gl.VertexAttribPointerWithOffset(parts.sizeBuffer, 2, gl.FLOAT, false, 0, 0)
	gl.VertexAttribDivisorARB(ATTR_INSTANCE_SIZE, 1)
}

func (parts *ParticleRender) updateBuffers() {
	// 'Buffer orphaning': Allows efficient modification of a buffer that is being
	// used for drawing by having OpenGL allocate a new one.

	// Position buffer
	gl.BindBuffer(gl.ARRAY_BUFFER, parts.positionBuffer)
	gl.BufferData(gl.ARRAY_BUFFER,
		cap(parts.positions)*positionByteSize,
		nil, gl.STREAM_DRAW)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0,
		len(parts.positions)*positionByteSize,
		gl.Ptr(parts.positions))

	// Color buffer
	gl.BindBuffer(gl.ARRAY_BUFFER, parts.colorBuffer)
	gl.BufferData(gl.ARRAY_BUFFER,
		cap(parts.colors)*colorByteSize,
		nil, gl.STREAM_DRAW)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0,
		len(parts.colors)*colorByteSize,
		gl.Ptr(parts.colors))

	// Size buffer
	gl.BindBuffer(gl.ARRAY_BUFFER, parts.sizeBuffer)
	gl.BufferData(gl.ARRAY_BUFFER,
		cap(parts.sizes)*sizeByteSize,
		nil, gl.STREAM_DRAW)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0,
		len(parts.sizes)*sizeByteSize,
		gl.Ptr(parts.sizes))
}
