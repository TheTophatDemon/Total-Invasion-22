package comps

import (
	"log"
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

// Holds properties of particles that are not involved in rendering.
type ParticleInfo struct {
	Velocity, Acceleration mgl32.Vec3
	Lifetime               float32
}

type ParticleRender struct {
	Mesh                *geom.Mesh
	Texture             *textures.Texture
	AnimPlayer          *AnimationPlayer
	Emitting            bool    // Particles will be emitting while this is true
	SpawnRadius         float32 // The spherical radius within which particles will be spawned.
	SpawnRate           float32 // The rate at which new particles will be spawned, in seconds per particle.
	VisibilityRadius    float32 // The radius of the invisible sphere that must be visible on camera for these particles to be drawn.
	LocalSpaceParticles bool    // If true, then particle positions will be in the space of the transform passed to the render method.

	spawnTimer float32

	// Particle instance fields
	positions      []mgl32.Vec3
	positionBuffer uint32
	colors         []mgl32.Vec4
	colorBuffer    uint32
	sizes          []mgl32.Vec2
	sizeBuffer     uint32
	particleInfos  []ParticleInfo
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
		particleInfos: make([]ParticleInfo, 0, maxInstances),
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

// Updates the particle emitter. The transform argument should not be nil.
func (parts *ParticleRender) Update(deltaTime float32, transform *Transform) {
	if transform == nil {
		log.Println("Error: (*ParticleRender).Update must be passed a non-nil transform!")
		return
	}
	if parts.AnimPlayer != nil {
		parts.AnimPlayer.Update(deltaTime)
	}

	//TODO: Particle movement customization
	for i := range parts.particleInfos {
		info := &parts.particleInfos[i]
		info.Lifetime -= deltaTime
		if info.Lifetime <= 0.0 {
			parts.removeParticle(i)
			break
		}
		info.Velocity = info.Velocity.Add(info.Acceleration.Mul(deltaTime))
		parts.positions[i] = parts.positions[i].Add(info.Velocity.Mul(deltaTime))
	}

	if parts.Emitting {
		parts.spawnTimer += deltaTime
		if parts.spawnTimer > parts.SpawnRate {
			parts.spawnTimer = 0.0
			dir := math2.RandomDir()
			position := dir.Mul(parts.SpawnRadius)
			velocity := dir
			if !parts.LocalSpaceParticles {
				position = mgl32.TransformCoordinate(position, transform.Matrix())
				velocity = mgl32.TransformNormal(velocity, transform.Matrix())
			}
			parts.SpawnParticle(position, color.White.Vector(), mgl32.Vec2{0.25, 0.25}, ParticleInfo{
				Velocity: velocity,
				Lifetime: 2.0,
			})
		}
	}
}

func (parts *ParticleRender) Render(
	transform *Transform,
	context *render.Context,
) {
	if parts.Mesh == nil || len(parts.positions) == 0 || !render.IsSphereVisible(context, transform.Position(), parts.VisibilityRadius) {
		return
	}

	parts.updateBuffers()

	// Set uniforms
	shader := shaders.SpriteShaderInstanced
	shader.Use()
	_ = context.SetUniforms(shader)
	_ = shader.SetUniformInt(shaders.UniformTex, 0)

	if parts.LocalSpaceParticles {
		_ = shader.SetUniformMatrix(shaders.UniformModelMatrix, transform.Matrix())
	} else {
		_ = shader.SetUniformMatrix(shaders.UniformModelMatrix, mgl32.Ident4())
	}

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

	context.DrawnParticlesCount += uint32(len(parts.positions))
}

func (parts *ParticleRender) Finalize() {
	parts.Free()
}

func (parts *ParticleRender) SpawnParticle(position mgl32.Vec3, color mgl32.Vec4, size mgl32.Vec2, info ParticleInfo) {
	if len(parts.positions) >= cap(parts.positions) {
		return
	}

	parts.positions = append(parts.positions, position)
	parts.colors = append(parts.colors, color)
	parts.sizes = append(parts.sizes, size)
	parts.particleInfos = append(parts.particleInfos, info)
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
	parts.particleInfos[index] = parts.particleInfos[lastIndex]

	// Shrink the particle arrays by 1
	parts.positions = parts.positions[:lastIndex]
	parts.colors = parts.colors[:lastIndex]
	parts.sizes = parts.sizes[:lastIndex]
	parts.particleInfos = parts.particleInfos[:lastIndex]
}

func (parts *ParticleRender) Free() {
	gl.DeleteBuffers(1, &parts.positionBuffer)
	gl.DeleteBuffers(1, &parts.colorBuffer)
	gl.DeleteBuffers(1, &parts.sizeBuffer)
}

func (parts *ParticleRender) bind() {
	// Position buffer
	gl.EnableVertexAttribArray(ATTR_INSTANCE_POS)
	gl.BindBuffer(gl.ARRAY_BUFFER, parts.positionBuffer)
	gl.VertexAttribPointerWithOffset(ATTR_INSTANCE_POS, 3, gl.FLOAT, false, 0, 0)
	gl.VertexAttribDivisorARB(ATTR_INSTANCE_POS, 1)

	// Color buffer
	gl.EnableVertexAttribArray(ATTR_INSTANCE_COL)
	gl.BindBuffer(gl.ARRAY_BUFFER, parts.colorBuffer)
	gl.VertexAttribPointerWithOffset(ATTR_INSTANCE_COL, 4, gl.FLOAT, false, 0, 0)
	gl.VertexAttribDivisorARB(ATTR_INSTANCE_COL, 1)

	// Size buffer
	gl.EnableVertexAttribArray(ATTR_INSTANCE_SIZE)
	gl.BindBuffer(gl.ARRAY_BUFFER, parts.sizeBuffer)
	gl.VertexAttribPointerWithOffset(ATTR_INSTANCE_SIZE, 2, gl.FLOAT, false, 0, 0)
	gl.VertexAttribDivisorARB(ATTR_INSTANCE_SIZE, 1)
}

func (parts *ParticleRender) updateBuffers() {
	if len(parts.positions) == 0 {
		return
	}

	// Position buffer
	gl.BindBuffer(gl.ARRAY_BUFFER, parts.positionBuffer)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0,
		len(parts.positions)*positionByteSize,
		gl.Ptr(parts.positions))

	// Color buffer
	gl.BindBuffer(gl.ARRAY_BUFFER, parts.colorBuffer)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0,
		len(parts.colors)*colorByteSize,
		gl.Ptr(parts.colors))

	// Size buffer
	gl.BindBuffer(gl.ARRAY_BUFFER, parts.sizeBuffer)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0,
		len(parts.sizes)*sizeByteSize,
		gl.Ptr(parts.sizes))
}
