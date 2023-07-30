package ecomps

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/ecs"
	"tophatdemon.com/total-invasion-ii/engine/render"
)

type MeshRender struct {
	Mesh    *assets.Mesh
	Shader  *assets.Shader
	Texture *assets.Texture
	Group   string
}

func AddMeshRender(ent ecs.Entity, Mesh *assets.Mesh, Shader *assets.Shader, Texture *assets.Texture) {
	MeshRenders.Assign(ent, MeshRender{
		Mesh,
		Shader,
		Texture,
		"",
	})
}

// Add a mesh renderer that only renders a specific group in the mesh.
func AddMeshRenderGroup(ent ecs.Entity, Mesh *assets.Mesh, Shader *assets.Shader, Texture *assets.Texture, Group string) {
	MeshRenders.Assign(ent, MeshRender{
		Mesh,
		Shader,
		Texture,
		Group,
	})
}

func (mr *MeshRender) Render(ent ecs.Entity, context *render.Context) {
	// Set defaults
	modelMatrix := mgl32.Ident4()
	transform, ok := Transforms.Get(ent)
	if ok {
		modelMatrix = transform.GetMatrix()
	}

	animPlayer, hasAnimPlayer := AnimationPlayers.Get(ent)

	// Bind resources
	if mr.Mesh == nil || mr.Shader == nil {
		fmt.Println("WARNING: MeshRender is missing mesh or shader.")
	}
	mr.Mesh.Bind()
	mr.Shader.Use()
	if mr.Texture != nil {
		mr.Texture.Bind()
	}

	frame := 0
	if hasAnimPlayer {
		frame = animPlayer.Frame()
	}

	// Set uniforms
	_ = context.SetUniforms(mr.Shader)
	_ = mr.Shader.SetUniformInt(assets.UniformTex, 0)
	_ = mr.Shader.SetUniformInt(assets.UniformAtlas, 1)
	_ = mr.Shader.SetUniformBool(assets.UniformAtlasUsed, mr.Texture.IsAtlas())
	_ = mr.Shader.SetUniformMatrix(assets.UniformModelMatrix, modelMatrix)
	_ = mr.Shader.SetUniformInt(assets.UniformFrame, frame)

	if len(mr.Group) == 0 {
		mr.Mesh.DrawAll()
	} else {
		mr.Mesh.DrawGroup(mr.Group)
	}
}
