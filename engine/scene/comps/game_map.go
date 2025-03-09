package comps

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
)

type Map struct {
	GridShape      collision.Grid
	name           string
	body           Body
	tiles          te3.Tiles
	mesh           *geom.Mesh
	triMap         te3.TriMap        // Maps a flattened tile index to its indices in the mesh's triangles array.
	tileAnims      []AnimationPlayer // Animates each texture group of tiles
	groupRenderers []MeshRender      // Renders each texture group of tiles
}

var _ HasBody = (*Map)(nil)

func NewMap(te3File *te3.TE3File, collisionLayer collision.Mask) (Map, error) {
	mesh, triMap, err := te3File.BuildMesh()
	if err != nil {
		return Map{}, err
	}
	cache.TakeMesh(te3File.FilePath(), mesh)

	var gridShape collision.Grid = collision.NewGrid(te3File.Tiles.Width, te3File.Tiles.Height, te3File.Tiles.Length, te3File.Tiles.GridSpacing())

	// Set all tile collision shapes to use the triangle mesh by default.
	for i := range len(te3File.Tiles.Data) {
		if shapeID := te3File.Tiles.Data[i].ShapeID; shapeID >= 0 {
			var shapeMesh *geom.Mesh
			shapeMesh, err = cache.GetMesh(te3File.Tiles.Shapes[shapeID])
			if err != nil {
				return Map{}, err
			}
			tileTransform := te3File.Tiles.Data[i].GetRotationMatrix()
			// Transform the shape by the tile's transform
			trisCopy := append([]math2.Triangle{}, shapeMesh.Triangles()...)
			for i := range trisCopy {
				for p := range trisCopy[i] {
					trisCopy[i][p] = mgl32.TransformNormal(trisCopy[i][p], tileTransform)
				}
			}
			gridShape.SetShapeAtFlatIndex(i, collision.NewMeshFromTriangles(trisCopy))
		}
	}

	gameMap := Map{
		name: te3File.FilePath(),
		body: Body{
			Shape: gridShape,
			Layer: collisionLayer,
		},
		GridShape:      gridShape,
		tiles:          te3File.Tiles,
		mesh:           mesh,
		triMap:         triMap,
		tileAnims:      make([]AnimationPlayer, mesh.GroupCount()),
		groupRenderers: make([]MeshRender, mesh.GroupCount()),
	}

	for g, groupName := range mesh.GroupNames() {
		tex := cache.GetTexture(groupName)
		// Add animations if applicable
		if tex.IsAtlas() {
			anim, _ := tex.GetAnimation(tex.GetAnimationNames()[0])
			gameMap.tileAnims[g] = NewAnimationPlayer(anim, true)
		}

		// Add mesh component
		gameMap.groupRenderers[g] = NewMeshRenderGroup(mesh, shaders.MapShader, tex, groupName)
	}

	return gameMap, nil
}

func (gameMap *Map) Name() string {
	return gameMap.name
}

func (gameMap *Map) Body() *Body {
	return &gameMap.body
}

func (gameMap *Map) GetUntyped(handle scene.Handle) (any, bool) {
	return gameMap, true
}

func (gameMap *Map) Has(handle scene.Handle) bool {
	return true
}

func (gameMap *Map) Remove(handle scene.Handle) {
	// You fool! I am eternal.
}

func (gameMap *Map) TearDown() {
}

func (gm *Map) SetTileCollisionShapes(shapeName string, shape collision.Shape) {
	var shapeID te3.ShapeID = -1
	for id, name := range gm.tiles.Shapes {
		if name == shapeName {
			shapeID = te3.ShapeID(id)
			break
		}
	}
	if shapeID < 0 {
		return
	}
	for index, tile := range gm.tiles.Data {
		if tile.ShapeID == shapeID {
			gm.GridShape.SetShapeAtFlatIndex(index, shape)
		}
	}
	return
}

// Sets the collision shape of all tiles that have the specified shape with the given yaw value.
// The yaw value is 0-3 representing increments of 90 degrees.
func (gm *Map) SetTileCollisionShapesForYaw(shapeName string, yaw uint8, shape collision.Shape) {
	var shapeID te3.ShapeID = -1
	for id, name := range gm.tiles.Shapes {
		if name == shapeName {
			shapeID = te3.ShapeID(id)
			break
		}
	}
	if shapeID < 0 {
		return
	}
	for index, tile := range gm.tiles.Data {
		if tile.ShapeID == shapeID &&
			tile.Yaw == yaw {
			gm.GridShape.SetShapeAtFlatIndex(index, shape)
		}
	}
	return
}

func (gm *Map) Update(deltaTime float32) {
	for i := range gm.tileAnims {
		gm.tileAnims[i].Update(deltaTime)
	}
}

func (gm *Map) Render(context *render.Context) {
	for i := range gm.groupRenderers {
		gm.groupRenderers[i].Render(nil, &gm.tileAnims[i], context)
	}
}
