package engine

import (
	"errors"
	"fmt"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"

	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/ecomps"
	"tophatdemon.com/total-invasion-ii/engine/scene"
)

type GameMap struct {
	*assets.TE3File

	mesh           *assets.Mesh
	tileAnimations map[string]*ecomps.AnimationPlayer
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
	animPlayerMap := make(map[string]*ecomps.AnimationPlayer)
	for _, groupName := range mesh.GetGroupNames() {
		tex := assets.GetTexture(groupName)
		if tex.AnimationCount() > 0 {
			animPlayerMap[groupName] = ecomps.NewAnimationPlayerAutoPlay(tex.GetAnimation(0))
		}
	}

	return &GameMap{te3File, mesh, animPlayerMap}, nil
}

func (gm *GameMap) Update(deltaTime float32) {
	for _, anim := range gm.tileAnimations {
		anim.Update(deltaTime)
	}
}

func (gm *GameMap) Render(context *scene.RenderContext) {
	shader := assets.MapShader
	shader.Use()

	err := errors.Join(context.SetUniforms(shader),
		shader.SetUniformMatrix(assets.UniformModelMatrix, mgl32.Ident4()),
		shader.SetUniformInt(assets.UniformTex, 0),
		shader.SetUniformInt(assets.UniformAtlas, 1),
		shader.SetUniformBool(assets.UniformAtlasUsed, false))

	//Draw the map
	gm.mesh.Bind()

	//Render each geometry group with correct texture
	for _, group := range gm.mesh.GetGroupNames() {
		//Set the animation frame if applicable
		tileAnim, ok := gm.tileAnimations[group]
		if ok {
			frame := tileAnim.Frame()
			err = errors.Join(err,
				shader.SetUniformBool(assets.UniformAtlasUsed, true),
				shader.SetUniformInt(assets.UniformFrame, frame))
			gl.ActiveTexture(gl.TEXTURE1)
		} else {
			err = errors.Join(err,
				shader.SetUniformBool(assets.UniformAtlasUsed, false))
			gl.ActiveTexture(gl.TEXTURE0)
		}

		tex := assets.GetTexture(group)
		gl.BindTexture(tex.Target(), tex.ID())
		gm.mesh.DrawGroup(group)
	}

	if err != nil {
		fmt.Println(err)
	}
	CheckOpenGLError()
}
