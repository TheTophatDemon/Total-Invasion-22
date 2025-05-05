package comps

import (
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
)

const (
	ATTR_INSTANCE_POS = iota + 8
	ATTR_INSTANCE_COL
	ATTR_INSTANCE_SIZE
	ATTR_INSTANCE_SRC_RECT
)

const (
	PARTICLE_POSITION_BYTE_LEN = int32(unsafe.Sizeof(mgl32.Vec3{}))
	PARTICLE_POSITION_BYTE_OFS = unsafe.Offsetof(ParticleForm{}.Position)
	PARTICLE_COLOR_BYTE_LEN    = int32(unsafe.Sizeof(mgl32.Vec4{}))
	PARTICLE_COLOR_BYTE_OFS    = unsafe.Offsetof(ParticleForm{}.Color)
	PARTICLE_SIZE_BYTE_LEN     = int32(unsafe.Sizeof(mgl32.Vec2{}))
	PARTICLE_SIZE_BYTE_OFS     = unsafe.Offsetof(ParticleForm{}.Size)
	PARTICLE_SRC_RECT_BYTE_LEN = int32(unsafe.Sizeof(mgl32.Vec4{}))
	PARTICLE_SRC_RECT_BYTE_OFS = unsafe.Offsetof(ParticleForm{}.SrcRect)
	PARTICLE_BUFFER_STRIDE     = int32(unsafe.Sizeof(ParticleForm{}))
)

// Describes the appearance of the particle. Stored in a separate buffer and sent to the GPU.
type ParticleForm struct {
	SrcRect  mgl32.Vec4 // The rectangle within the texture representing the current frame of animation. 16 bytes.
	Color    mgl32.Vec4 // R, G, B, A color. 16 bytes.
	Position mgl32.Vec3 // Will be in global space if the particle system has 'LocalSpaceParticles' equal to false. Otherwise in local space. 12 bytes.
	Size     mgl32.Vec2 // Width and height of the quad. 8 bytes.
}

// Describes physics, animation, and movement of the particle. This is not sent to the GPU.
type ParticleInfo struct {
	Velocity, Acceleration mgl32.Vec3
	Lifetime               float32
	AnimPlayer             AnimationPlayer
}

type ParticleRender struct {
	Mesh                *geom.Mesh
	Texture             *textures.Texture
	EmissionTimer       float32 // Number of seconds before emission stops. Set this to >0 to start emitting particles.
	SpawnRadius         float32 // The spherical radius within which particles will be spawned.
	SpawnRate           float32 // The rate at which new particles will be spawned, in seconds per particle.
	BurstCount          int     // Each time a particle is spawned, spawn this many particles.
	VisibilityRadius    float32 // The radius of the invisible sphere that must be visible on camera for these particles to be drawn.
	LocalSpaceParticles bool    // If true, then particle positions will be in the space of the transform passed to the render method.
	MaxCount            int     // Maxmimum number of particles to render at one time

	// Called every frame to move and animate the particles. Velocity and acceleration will be applied later.
	UpdateFunc func(
		deltaTime float32,
		form *ParticleForm,
		info *ParticleInfo,
	)

	// Called after every particle is spawned in order to set up its initial properties.
	// By default, the particle has a random position assigned within the component's emission radius.
	// This position will in the local space of the transform passed into Update() if LocalSpaceParticles is enabled.
	// Otherwise, the position is in global space.
	// The default velocity will be the unit vector pointing away from the center.
	SpawnFunc func(
		index int,
		form *ParticleForm,
		info *ParticleInfo,
	)

	spawnTimer float32

	// Particle instance fields
	particleForms  []ParticleForm
	particleInfos  []ParticleInfo
	particleBuffer uint32
}

// Initializes the particle renderer to support the given maximum number of particles.
func (parts *ParticleRender) Init() {
	if parts.Mesh == nil {
		parts.Mesh = cache.QuadMesh
	}
	if parts.MaxCount == 0 {
		parts.MaxCount = 10
	}
	if parts.VisibilityRadius == 0 {
		parts.VisibilityRadius = 5.0
	}
	if parts.BurstCount == 0 {
		parts.BurstCount = 1
	}

	parts.spawnTimer = parts.SpawnRate // Make particles emit as soon as it's spawned.

	parts.particleForms = make([]ParticleForm, 0, parts.MaxCount)
	parts.particleInfos = make([]ParticleInfo, 0, parts.MaxCount)

	gl.GenBuffers(1, &parts.particleBuffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, parts.particleBuffer)
	gl.BufferData(gl.ARRAY_BUFFER,
		cap(parts.particleForms)*int(PARTICLE_BUFFER_STRIDE),
		nil, gl.STREAM_DRAW)
}

// Updates the particle emitter. The transform argument should not be nil.
func (parts *ParticleRender) Update(deltaTime float32, transform *Transform) {
	if transform == nil {
		failure.LogErrWithLocation("Update must be passed a non-nil transform!")
		return
	}

	// Update existing particles
	for i := range parts.particleInfos {
		form := &parts.particleForms[i]
		info := &parts.particleInfos[i]
		if parts.UpdateFunc != nil {
			parts.UpdateFunc(deltaTime, form, info)
		}
		info.AnimPlayer.Update(deltaTime)
		info.Velocity = info.Velocity.Add(info.Acceleration.Mul(deltaTime))
		form.Position = form.Position.Add(info.Velocity.Mul(deltaTime))
		form.SrcRect = info.AnimPlayer.FrameUV().Vec4()
		info.Lifetime -= deltaTime
	}

	// Remove dead particles
	newLen := len(parts.particleInfos)
	for i := range parts.particleInfos {
		info := &parts.particleInfos[i]
		if info.Lifetime <= 0.0 {
			newLen--
			// Swap the latest particle with the one being removed in order to keep the list contiguous.
			parts.particleForms[i] = parts.particleForms[newLen]
			parts.particleInfos[i] = parts.particleInfos[newLen]
		}
	}
	parts.particleForms = parts.particleForms[:newLen]
	parts.particleInfos = parts.particleInfos[:newLen]

	// Spawn new particles
	if parts.EmissionTimer > 0.0 {
		parts.spawnTimer += deltaTime
		if parts.spawnTimer > parts.SpawnRate {
			parts.spawnTimer = 0.0

			for range parts.BurstCount {
				dir := math2.RandomDir()
				position := dir.Mul(parts.SpawnRadius)
				if !parts.LocalSpaceParticles {
					position = mgl32.TransformCoordinate(position, transform.Matrix())
					dir = mgl32.TransformNormal(dir, transform.Matrix())
				}

				form := ParticleForm{
					SrcRect:  mgl32.Vec4{0.0, 1.0, 1.0, 1.0},
					Color:    color.White.Vector(),
					Position: position,
					Size:     mgl32.Vec2{1.0, 1.0},
				}

				info := ParticleInfo{
					Velocity: dir,
					Lifetime: 1.0,
				}

				if len(parts.particleInfos) != cap(parts.particleInfos) {
					if parts.SpawnFunc != nil {
						parts.SpawnFunc(len(parts.particleInfos), &form, &info)
					}
					if info.Lifetime > 0.0 {
						parts.particleForms = append(parts.particleForms, form)
						parts.particleInfos = append(parts.particleInfos, info)
					}
				}
			}
		}
		parts.EmissionTimer = max(0.0, parts.EmissionTimer-deltaTime)
	}
}

func (parts *ParticleRender) Render(
	transform *Transform,
	context *render.Context,
) {
	if parts.Mesh == nil || len(parts.particleInfos) == 0 || !context.IsSphereVisible(transform.Position(), parts.VisibilityRadius) {
		return
	}

	parts.updateBuffers()

	// Set uniforms
	shader := shaders.ParticlesShader
	shader.Use()
	_ = context.SetUniforms(shader)
	_ = shader.SetUniformInt(shaders.UniformTex, 0)

	if parts.LocalSpaceParticles {
		_ = shader.SetUniformMatrix(shaders.UniformModelMatrix, transform.Matrix())
	} else {
		_ = shader.SetUniformMatrix(shaders.UniformModelMatrix, mgl32.Ident4())
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
		int32(len(parts.particleForms)))

	context.DrawnParticlesCount += uint32(len(parts.particleForms))
}

func (parts *ParticleRender) Finalize() {
	parts.Free()
}

func (parts *ParticleRender) Free() {
	gl.DeleteBuffers(1, &parts.particleBuffer)
}

func (parts *ParticleRender) bind() {
	gl.EnableVertexAttribArray(ATTR_INSTANCE_POS)
	gl.EnableVertexAttribArray(ATTR_INSTANCE_COL)
	gl.EnableVertexAttribArray(ATTR_INSTANCE_SIZE)
	gl.EnableVertexAttribArray(ATTR_INSTANCE_SRC_RECT)

	gl.BindBuffer(gl.ARRAY_BUFFER, parts.particleBuffer)

	gl.VertexAttribPointerWithOffset(ATTR_INSTANCE_POS, 3, gl.FLOAT, false, PARTICLE_BUFFER_STRIDE, PARTICLE_POSITION_BYTE_OFS)
	gl.VertexAttribPointerWithOffset(ATTR_INSTANCE_COL, 4, gl.FLOAT, false, PARTICLE_BUFFER_STRIDE, PARTICLE_COLOR_BYTE_OFS)
	gl.VertexAttribPointerWithOffset(ATTR_INSTANCE_SIZE, 2, gl.FLOAT, false, PARTICLE_BUFFER_STRIDE, PARTICLE_SIZE_BYTE_OFS)
	gl.VertexAttribPointerWithOffset(ATTR_INSTANCE_SRC_RECT, 4, gl.FLOAT, false, PARTICLE_BUFFER_STRIDE, PARTICLE_SRC_RECT_BYTE_OFS)

	gl.VertexAttribDivisorARB(ATTR_INSTANCE_POS, 1)
	gl.VertexAttribDivisorARB(ATTR_INSTANCE_COL, 1)
	gl.VertexAttribDivisorARB(ATTR_INSTANCE_SIZE, 1)
	gl.VertexAttribDivisorARB(ATTR_INSTANCE_SRC_RECT, 1)
}

func (parts *ParticleRender) updateBuffers() {
	if len(parts.particleForms) == 0 {
		return
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, parts.particleBuffer)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0,
		len(parts.particleForms)*int(PARTICLE_BUFFER_STRIDE),
		gl.Ptr(parts.particleForms))
}
