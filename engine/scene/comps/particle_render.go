package comps

import (
	"math/rand"
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
)

const (
	AttrInstancePos = iota + 8
	AttrInstanceCol
	AttrInstanceSize
	AttrInstanceSrcRect
)

const (
	ParticlePositionByteLen = int32(unsafe.Sizeof(mgl32.Vec3{}))
	ParticlePositionByteOfs = unsafe.Offsetof(ParticleForm{}.Position)
	ParticleColorByteLen    = int32(unsafe.Sizeof(mgl32.Vec4{}))
	ParticleColorByteOfs    = unsafe.Offsetof(ParticleForm{}.Color)
	ParticleSizeByteLen     = int32(unsafe.Sizeof(mgl32.Vec2{}))
	ParticleSizeByteOfs     = unsafe.Offsetof(ParticleForm{}.Size)
	ParticleSrcRectByteLen  = int32(unsafe.Sizeof(mgl32.Vec4{}))
	ParticleSrcRectByteOfs  = unsafe.Offsetof(ParticleForm{}.SrcRect)
	ParticleBufferStride    = int32(unsafe.Sizeof(ParticleForm{}))
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
	LocalTransform      Transform  // Transform relative to the position it is rendered at.
	EmissionTimer       float32    // Number of seconds before emission stops. Set this to >0 to start emitting particles.
	SpawnRadius         float32    // The spherical radius within which particles will be spawned.
	SpawnLength         mgl32.Vec3 // Turns the emission sphere into a capsule going in this direction.
	SpawnRate           float32    // The rate at which new particles will be spawned, in seconds per particle.
	BurstCount          int        // Each time a particle is spawned, spawn this many particles.
	VisibilityRadius    float32    // The radius of the invisible sphere that must be visible on camera for these particles to be drawn.
	LocalSpaceParticles bool       // If true, then particle positions will be in the space of the transform passed to the render method.
	MaxCount            int        // Maxmimum number of particles to render at one time

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
		cap(parts.particleForms)*int(ParticleBufferStride),
		nil, gl.STREAM_DRAW)
}

// Updates the particle emitter.
func (parts *ParticleRender) Update(deltaTime float32, spawnPosition mgl32.Vec3) {
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

	var worldTransform mgl32.Mat4
	if !parts.LocalSpaceParticles {
		worldTransform = mgl32.Translate3D(spawnPosition[0], spawnPosition[1], spawnPosition[2]).Mul4(parts.LocalTransform.Matrix())
	}

	// Spawn new particles
	if parts.EmissionTimer > 0.0 {
		parts.spawnTimer += deltaTime
		if parts.spawnTimer > parts.SpawnRate {
			parts.spawnTimer = 0.0

			for range parts.BurstCount {
				dir := math2.RandomDir()
				position := parts.SpawnLength.Mul(rand.Float32()).Add(dir.Mul(parts.SpawnRadius))
				if !parts.LocalSpaceParticles {
					position = mgl32.TransformCoordinate(position, worldTransform)
					dir = mgl32.TransformNormal(dir, worldTransform)
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

func (parts *ParticleRender) Render(position mgl32.Vec3, context *render.Context) {
	worldTransform := mgl32.Translate3D(position[0], position[1], position[2]).Mul4(parts.LocalTransform.Matrix())
	worldPosition := worldTransform.Col(3).Vec3()
	if parts.Mesh == nil || len(parts.particleInfos) == 0 || !context.IsSphereVisible(worldPosition, parts.VisibilityRadius) {
		return
	}

	parts.updateBuffers()

	// Set uniforms
	shader := shaders.ParticlesShader
	shader.Use()
	_ = context.SetUniforms(shader)
	_ = shader.SetUniformInt(shaders.UniformTex, 0)

	if parts.LocalSpaceParticles {
		_ = shader.SetUniformMatrix(shaders.UniformModelMatrix, worldTransform)
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
	gl.EnableVertexAttribArray(AttrInstancePos)
	gl.EnableVertexAttribArray(AttrInstanceCol)
	gl.EnableVertexAttribArray(AttrInstanceSize)
	gl.EnableVertexAttribArray(AttrInstanceSrcRect)

	gl.BindBuffer(gl.ARRAY_BUFFER, parts.particleBuffer)

	gl.VertexAttribPointerWithOffset(AttrInstancePos, 3, gl.FLOAT, false, ParticleBufferStride, ParticlePositionByteOfs)
	gl.VertexAttribPointerWithOffset(AttrInstanceCol, 4, gl.FLOAT, false, ParticleBufferStride, ParticleColorByteOfs)
	gl.VertexAttribPointerWithOffset(AttrInstanceSize, 2, gl.FLOAT, false, ParticleBufferStride, ParticleSizeByteOfs)
	gl.VertexAttribPointerWithOffset(AttrInstanceSrcRect, 4, gl.FLOAT, false, ParticleBufferStride, ParticleSrcRectByteOfs)

	gl.VertexAttribDivisorARB(AttrInstancePos, 1)
	gl.VertexAttribDivisorARB(AttrInstanceCol, 1)
	gl.VertexAttribDivisorARB(AttrInstanceSize, 1)
	gl.VertexAttribDivisorARB(AttrInstanceSrcRect, 1)
}

func (parts *ParticleRender) updateBuffers() {
	if len(parts.particleForms) == 0 {
		return
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, parts.particleBuffer)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0,
		len(parts.particleForms)*int(ParticleBufferStride),
		gl.Ptr(parts.particleForms))
}
