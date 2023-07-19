package engine

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"

	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/comps"
)

type GameMap struct {
	*assets.TE3File

	mesh           *assets.Mesh
	tileAnimations map[string]*comps.AnimationPlayer
}

func LoadGameMap(te3Path string) (*GameMap, error) {
	te3File, err := assets.LoadTE3File(te3Path)
	if err != nil {
		return nil, err
	}

	mesh, err := te3File.BuildMesh()
	if err != nil {
		return nil, err
	}

	//Create tile animation players
	animPlayerMap := make(map[string]*comps.AnimationPlayer)
	for _, groupName := range mesh.GetGroupNames() {
		tex := assets.GetTexture(groupName)
		if tex.AnimationCount() > 0 {
			animPlayerMap[groupName] = comps.NewAnimationPlayerAutoPlay(tex.GetAnimation(0))
		}
	}

	return &GameMap{te3File, mesh, animPlayerMap}, nil
}

func (gm *GameMap) Update(deltaTime float32) {
	for _, anim := range gm.tileAnimations {
		anim.Update(deltaTime)
	}
}

func (gm *GameMap) Render(viewProjection mgl32.Mat4) {
	assets.MapShader.Use()

	gl.UniformMatrix4fv(assets.MapShader.GetUniformLoc("uViewProjection"), 1, false, &viewProjection[0])
	gl.Uniform1i(assets.MapShader.GetUniformLoc("uTex"), 0)
	gl.Uniform1i(assets.MapShader.GetUniformLoc("uAtlas"), 1)
	gl.Uniform1i(assets.MapShader.GetUniformLoc("uAtlasUsed"), 0)

	gl.Uniform1f(assets.MapShader.GetUniformLoc("uFogStart"), 1.0)
	gl.Uniform1f(assets.MapShader.GetUniformLoc("uFogLength"), 50.0)

	//Draw the map
	gm.mesh.Bind()

	model := mgl32.Ident4()
	gl.UniformMatrix4fv(assets.MapShader.GetUniformLoc("uModelTransform"), 1, false, &model[0])

	//Render each geometry group with correct texture
	for _, group := range gm.mesh.GetGroupNames() {
		//Set the animation frame if applicable
		tileAnim, ok := gm.tileAnimations[group]
		if ok {
			gl.Uniform1i(assets.MapShader.GetUniformLoc("uAtlasUsed"), 1)
			frame := tileAnim.Frame()
			gl.Uniform1i(assets.MapShader.GetUniformLoc("uFrame"), int32(frame))
			gl.ActiveTexture(gl.TEXTURE1)
		} else {
			gl.Uniform1i(assets.MapShader.GetUniformLoc("uAtlasUsed"), 0)
			gl.ActiveTexture(gl.TEXTURE0)
		}

		tex := assets.GetTexture(group)
		gl.BindTexture(tex.Target(), tex.ID())
		gm.mesh.DrawGroup(group)
	}

	CheckOpenGLError()
}
