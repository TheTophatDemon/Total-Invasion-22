package collision

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/containers"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type Grid struct {
	shape
	cels                  []Shape
	width, height, length int
	spacing               float32
	celsChecked           map[[3]int]bool // Pre-allocated map for tracking cels visited when walking the grid.
	visitBuffer           [][3]int        // Pre-allocated slice of coordinates for keeping track of cels to visit when walking the grid.
}

var _ Shape = (*Grid)(nil)

func NewGrid(width, height, length int, spacing float32) Grid {
	return Grid{
		shape: shape{
			extents: math2.Box{
				Min: mgl32.Vec3{-spacing, -spacing, -spacing},
				Max: mgl32.Vec3{
					spacing + float32(width)*spacing,
					spacing + float32(height)*spacing,
					spacing + float32(length)*spacing,
				},
			},
		},
		cels:        make([]Shape, width*height*length),
		spacing:     spacing,
		width:       width,
		height:      height,
		length:      length,
		celsChecked: make(map[[3]int]bool),
		visitBuffer: make([][3]int, width*height*length),
	}
}

func (grid Grid) String() string {
	return "Grid"
}

func (grid *Grid) Width() int {
	return grid.width
}

func (grid *Grid) Height() int {
	return grid.height
}

func (grid *Grid) Length() int {
	return grid.length
}

func (grid *Grid) Dimensions() (int, int, int) {
	return grid.width, grid.height, grid.length
}

func (grid *Grid) Spacing() float32 {
	return grid.spacing
}

func (grid *Grid) AreCoordsValid(x, y, z int) bool {
	return x >= 0 && y >= 0 && z >= 0 &&
		x < grid.width && y < grid.height && z < grid.length
}

func (grid *Grid) SetShapeAt(x, y, z int, shape Shape) {
	if !grid.AreCoordsValid(x, y, z) {
		return
	}
	grid.cels[grid.FlattenGridPos(x, y, z)] = shape
}

func (grid *Grid) SetShapeAtFlatIndex(index int, shape Shape) {
	if index > 0 && index < len(grid.cels) {
		grid.cels[index] = shape
	}
}

func (grid *Grid) ShapeAt(x, y, z int) Shape {
	if !grid.AreCoordsValid(x, y, z) {
		return nil
	}
	return grid.cels[grid.FlattenGridPos(x, y, z)]
}

func (grid *Grid) UnflattenGridPos(index int) (int, int, int) {
	return (index % grid.width), (index / (grid.width * grid.length)), ((index / grid.width) % grid.length)
}

// Returns the flat index into the Data array for the given integer grid position (does not validate).
func (grid *Grid) FlattenGridPos(x, y, z int) int {
	return x + (z * grid.width) + (y * grid.width * grid.length)
}

func (grid *Grid) WorldToGridPos(worldPos mgl32.Vec3) (int, int, int) {
	var out [3]int
	for i := range out {
		out[i] = int(worldPos[i] / grid.spacing)
	}
	return out[0], out[1], out[2]
}

func (grid *Grid) GridToWorldPos(i, j, k int, center bool) mgl32.Vec3 {
	out := mgl32.Vec3{
		float32(i) * grid.spacing,
		float32(j) * grid.spacing,
		float32(k) * grid.spacing,
	}
	if center {
		out[0] += grid.spacing / 2.0
		out[1] += grid.spacing / 2.0
		out[2] += grid.spacing / 2.0
	}
	return out
}

func (grid Grid) Extents() math2.Box {
	return grid.extents
}

func (grid Grid) Raycast(rayOrigin, rayDir, shapeOffset mgl32.Vec3, maxDist float32) RaycastResult {
	if lenSqr := rayDir.LenSqr(); lenSqr == 0.0 {
		return RaycastResult{}
	} else if lenSqr != 1.0 {
		rayDir = rayDir.Normalize()
	}

	rayOrigin = rayOrigin.Sub(shapeOffset)
	var maxDistSqr float32 = maxDist * maxDist

	var pos mgl32.Vec3 = rayOrigin
	var nextPos mgl32.Vec3 = pos

	var i, j, k int
	i, j, k = grid.WorldToGridPos(pos)

	if !grid.AreCoordsValid(i, j, k) {
		return RaycastResult{}
	}

	for {
		// Respond to hit tile
		tileIndex := grid.FlattenGridPos(i, j, k)
		if tileShape := grid.cels[tileIndex]; tileShape != nil {
			tileCenter := grid.GridToWorldPos(i, j, k, true)
			if cast := tileShape.Raycast(rayOrigin, rayDir, tileCenter, maxDist); cast.Hit {
				if cast.Distance*cast.Distance > maxDistSqr {
					return RaycastResult{}
				}
				return RaycastResult{
					Hit:      true,
					Position: cast.Position.Add(shapeOffset),
					Normal:   cast.Normal,
					Distance: cast.Distance,
				}
			}
		}

		// Check if distance is exceeded
		if nextPos.Sub(rayOrigin).LenSqr() > maxDistSqr {
			break
		}

		pos = nextPos

		// Find positions on the XZ, XY, and YZ planes that will be hit next by the ray.
		offsetFromCorner := pos.Sub(grid.GridToWorldPos(i, j, k, false))
		if offsetFromCorner.Z() == 0.0 && rayDir.Z() < 0.0 {
			k -= 1
		}
		if offsetFromCorner.Y() == 0.0 && rayDir.Y() < 0.0 {
			j -= 1
		}
		if offsetFromCorner.X() == 0.0 && rayDir.X() < 0.0 {
			i -= 1
		}
		offsetFromCorner = pos.Sub(grid.GridToWorldPos(i, j, k, false))

		var nextYZPlane, nextXZPlane, nextXYPlane mgl32.Vec3
		var smallestDist float32 = math.MaxFloat32
		nextPos = mgl32.Vec3{}

		// Find next YZ plane
		var skip bool
		if rayDir.X() > 0.0 {
			nextYZPlane = mgl32.Vec3{pos.X() - offsetFromCorner.X() + grid.spacing}
		} else if rayDir.X() < 0.0 {
			nextYZPlane = mgl32.Vec3{pos.X() - offsetFromCorner.X()}
		} else {
			skip = true
		}
		if !skip {
			t := (nextYZPlane.X() - pos.X()) / rayDir.X()
			nextYZPlane = mgl32.Vec3{nextYZPlane.X(), pos.Y() + rayDir.Y()*t, pos.Z() + rayDir.Z()*t}
			if t < smallestDist {
				smallestDist = t
				nextPos = nextYZPlane
			}
		}

		// Find next XZ plane
		skip = false
		if rayDir.Y() > 0.0 {
			nextXZPlane = mgl32.Vec3{0.0, pos.Y() - offsetFromCorner.Y() + grid.spacing}
		} else if rayDir.Y() < 0.0 {
			nextXZPlane = mgl32.Vec3{0.0, pos.Y() - offsetFromCorner.Y()}
		} else {
			skip = true
		}
		if !skip {
			t := (nextXZPlane.Y() - pos.Y()) / rayDir.Y()
			nextXZPlane = mgl32.Vec3{pos.X() + rayDir.X()*t, nextXZPlane.Y(), pos.Z() + rayDir.Z()*t}
			if t < smallestDist {
				smallestDist = t
				nextPos = nextXZPlane
			}
		}

		// Find next XY plane
		skip = false
		if rayDir.Z() > 0.0 {
			nextXYPlane = mgl32.Vec3{0.0, 0.0, pos.Z() - offsetFromCorner.Z() + grid.spacing}
		} else if rayDir.Z() < 0.0 {
			nextXYPlane = mgl32.Vec3{0.0, 0.0, pos.Z() - offsetFromCorner.Z()}
		} else {
			skip = true
		}
		if !skip {
			t := (nextXYPlane.Z() - pos.Z()) / rayDir.Z()
			nextXYPlane = mgl32.Vec3{pos.X() + rayDir.X()*t, pos.Y() + rayDir.Y()*t, nextXYPlane.Z()}
			if t < smallestDist {
				smallestDist = t
				nextPos = nextXYPlane
			}
		}

		// Check for a tile at that grid cell
		i, j, k = grid.WorldToGridPos(nextPos)
		if nextPos.ApproxEqual(nextXYPlane) && rayDir.Z() < 0.0 {
			k -= 1
		} else if nextPos.ApproxEqual(nextXZPlane) && rayDir.Y() < 0.0 {
			j -= 1
		} else if nextPos.ApproxEqual(nextYZPlane) && rayDir.X() < 0.0 {
			i -= 1
		}
		if !grid.AreCoordsValid(i, j, k) {
			break
		}
	}

	return RaycastResult{}
}

func (grid Grid) ResolveCollision(myPosition, theirPosition mgl32.Vec3, theirShape Shape) Result {
	// The map doesn't move, silly!
	return Result{}
}

// Call this to resolve collisions another body has with the grid using an optimized grid-walking method.
func (grid *Grid) ResolveOtherBodysCollision(myPosition, theirPosition mgl32.Vec3, theirShape Shape) Result {
	var firstHitPosition mgl32.Vec3
	var numberOfHits uint
	var theirPositionRelative mgl32.Vec3 = theirPosition.Sub(myPosition)
	var theirOriginalPositionRelative mgl32.Vec3 = theirPositionRelative

	// Iterate over the subset of tiles that the body occupies
	bbox := theirShape.Extents().Translate(theirPositionRelative)
	i, j, k := grid.WorldToGridPos(bbox.Max)
	l, m, n := grid.WorldToGridPos(bbox.Min)
	minX, minY, minZ := max(0, min(i, l)), max(0, min(j, m)), max(0, min(k, n))
	maxX, maxY, maxZ := min(max(i, l), grid.width-1), min(max(j, m), grid.height-1), min(max(k, n), grid.length-1)
	tileCount := (maxX - minX + 1) * (maxZ - minZ + 1) * (maxY - minY + 1)
	if tileCount <= 0 {
		return Result{}
	}

	// Visit each tile within the movement range, starting from the one closest to the current position and then proceeding to its neighbors.
	clear(grid.celsChecked)
	var visitBuffer [][3]int = grid.visitBuffer[:tileCount]
	clear(visitBuffer)
	visitQueue := containers.NewRingBuffer(visitBuffer)
	startX, startY, startZ := grid.WorldToGridPos(theirPositionRelative)
	if !grid.AreCoordsValid(startX, startY, startZ) {
		return Result{}
	}
	start := [3]int{startX, startY, startZ}
	visitQueue.Enqueue(start)
	grid.celsChecked[start] = true

	for pos, empty := visitQueue.Dequeue(); !empty; pos, empty = visitQueue.Dequeue() {
		t := grid.FlattenGridPos(pos[0], pos[1], pos[2])
		if grid.cels[t] != nil {
			// Resolve collision against this tile
			tileCenter := grid.GridToWorldPos(pos[0], pos[1], pos[2], true)
			tileShape := grid.cels[t]
			var res Result = theirShape.ResolveCollision(theirPositionRelative, tileCenter, tileShape)
			if res.Hit {
				theirPositionRelative = theirPositionRelative.Add(res.Normal.Mul(res.Penetration))
				if numberOfHits == 0 {
					firstHitPosition = res.Position
				}
				numberOfHits++
			}
		}

		// Add neighboring tiles to the queue
		neighbors := [...][3]int{
			{pos[0] + 1, pos[1], pos[2]},
			{pos[0] - 1, pos[1], pos[2]},
			{pos[0], pos[1] + 1, pos[2]},
			{pos[0], pos[1] - 1, pos[2]},
			{pos[0], pos[1], pos[2] + 1},
			{pos[0], pos[1], pos[2] - 1},
		}
		for _, n := range neighbors {
			if n[0] < minX || n[1] < minY || n[2] < minZ || n[0] > maxX || n[1] > maxY || n[2] > maxZ {
				continue
			}
			_, v := grid.celsChecked[n]
			if !v {
				visitQueue.Enqueue(n)
				grid.celsChecked[n] = true
			}
		}
	}

	var overallMovement mgl32.Vec3 = theirPositionRelative.Sub(theirOriginalPositionRelative)
	var overallDistance float32 = overallMovement.Len()
	var overallNormal mgl32.Vec3
	if overallDistance > 0.0 {
		overallNormal = overallMovement.Mul(1.0 / overallDistance)
	}
	return Result{
		Hit:         numberOfHits > 0,
		Position:    firstHitPosition.Add(myPosition),
		Normal:      overallNormal,
		Penetration: overallDistance,
	}
}
