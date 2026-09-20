package collision

import (
	"iter"
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/containers"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type GridCell struct {
	ShapeId int  // Index into the shapes array
	Layer   Mask // Collision layer. Will be 0 for an empty tile.
	ZoneId  int
}

type Grid struct {
	extents               math2.Box
	shapes                []Shape
	cells                 []GridCell
	width, height, length int
	spacing               float32
	celsChecked           map[[3]int]bool // Pre-allocated map for tracking cels visited when walking the grid.
	visitBuffer           [][3]int        // Pre-allocated slice of coordinates for keeping track of cels to visit when walking the grid.
	zoneConnections       []bool          // Holds a matrix mapping zones that are connected to each other by a doorway.
	zoneCount             int             // Number of distinct zone IDs in the grid
}

func NewGrid(width, height, length int, spacing float32) Grid {
	return Grid{
		extents: math2.Box{
			Min: mgl32.Vec3{-spacing, -spacing, -spacing},
			Max: mgl32.Vec3{
				spacing + float32(width)*spacing,
				spacing + float32(height)*spacing,
				spacing + float32(length)*spacing,
			},
		},
		cells:       make([]GridCell, width*height*length),
		shapes:      make([]Shape, 0, 32),
		spacing:     spacing,
		width:       width,
		height:      height,
		length:      length,
		celsChecked: make(map[[3]int]bool),
		visitBuffer: make([][3]int, width*height*length),
	}
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

func (grid *Grid) ShapeIndex(shape Shape) int {
	i := 0
	for ; i < len(grid.shapes); i++ {
		if grid.shapes[i] == shape {
			return i
		}
	}
	if i == len(grid.shapes) {
		grid.shapes = append(grid.shapes, shape)
	}
	return i
}

func (grid *Grid) SetShapeAt(x, y, z int, shape Shape, layer Mask) {
	if !grid.AreCoordsValid(x, y, z) {
		return
	}
	grid.cells[grid.FlattenGridPos(x, y, z)] = GridCell{
		ShapeId: grid.ShapeIndex(shape),
		Layer:   layer,
	}
}

func (grid *Grid) SetZoneAt(x, y, z int, zone int) {
	if !grid.AreCoordsValid(x, y, z) {
		return
	}
	grid.cells[grid.FlattenGridPos(x, y, z)].ZoneId = zone
}

func (grid *Grid) SetZoneAtFlatIndex(index int, zone int) {
	if index >= 0 && index < len(grid.cells) {
		grid.cells[index].ZoneId = zone
	}
}

func (grid *Grid) GetZoneAt(x, y, z int) int {
	if !grid.AreCoordsValid(x, y, z) {
		return -1
	}
	return grid.cells[grid.FlattenGridPos(x, y, z)].ZoneId
}

// Iterates over valid zone ids that are intersecting with the given bounding box in world space.
func (grid *Grid) ZonesTouching(bbox math2.Box) iter.Seq[int] {
	return func(yield func(int) bool) {
		minX, minY, minZ := grid.WorldToGridPos(bbox.Min)
		maxX, maxY, maxZ := grid.WorldToGridPos(bbox.Max)
		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				for z := minZ; z <= maxZ; z++ {
					if zone := grid.GetZoneAt(x, y, z); zone > 0 {
						if !yield(zone) {
							return
						}
					}
				}
			}
		}
	}
}

func (grid *Grid) SetShapeAtFlatIndex(index int, shape Shape, layer Mask) {
	if index >= 0 && index < len(grid.cells) {
		cell := GridCell{
			ShapeId: grid.ShapeIndex(shape),
			Layer:   layer,
		}
		if shape.pointCount > 0 {
			cell.ZoneId = -1
		}
		grid.cells[index] = cell
	}
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

func (grid Grid) Raycast(rayOrigin, rayDir mgl32.Vec3, maxDist float32, filter Mask) Result {
	if lenSqr := rayDir.LenSqr(); lenSqr == 0.0 {
		return Result{}
	} else if lenSqr != 1.0 {
		rayDir = rayDir.Normalize()
	}

	var maxDistSqr float32 = maxDist * maxDist

	var pos mgl32.Vec3 = rayOrigin
	var nextPos mgl32.Vec3 = pos

	var i, j, k int
	i, j, k = grid.WorldToGridPos(pos)

	if !grid.AreCoordsValid(i, j, k) {
		return Result{}
	}

	for {
		// Respond to hit tile
		tileIndex := grid.FlattenGridPos(i, j, k)
		if cell := grid.cells[tileIndex]; cell.Layer.On(filter) {
			tileCenter := grid.GridToWorldPos(i, j, k, true)
			if cast := grid.shapes[cell.ShapeId].Raycast(tileCenter, rayOrigin, rayDir, maxDist); cast.Hit {
				if cast.Distance*cast.Distance <= maxDistSqr {
					return Result{
						Hit:      true,
						Position: cast.Position,
						Normal:   cast.Normal,
						Distance: cast.Distance,
					}
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

	return Result{}
}

// Call this to resolve collisions another body has with the grid using an optimized grid-walking method.
func (grid *Grid) SweepAgainst(myPosition, theirPosition, theirMovement mgl32.Vec3, theirShape Shape, filters Mask) (Result, Mask) {
	var theirPositionRelative mgl32.Vec3 = theirPosition.Sub(myPosition)

	// Iterate over the subset of tiles that the body occupies
	bbox := theirShape.Extents().Translate(theirPositionRelative).Union(theirShape.Extents().Translate(theirPositionRelative.Add(theirMovement)))
	i, j, k := grid.WorldToGridPos(bbox.Max)
	l, m, n := grid.WorldToGridPos(bbox.Min)
	minX, minY, minZ := max(0, min(i, l)), max(0, min(j, m)), max(0, min(k, n))
	maxX, maxY, maxZ := min(max(i, l), grid.width-1), min(max(j, m), grid.height-1), min(max(k, n), grid.length-1)
	tileCount := (maxX - minX + 1) * (maxZ - minZ + 1) * (maxY - minY + 1)
	if tileCount <= 0 {
		return Result{
			Position: theirPosition.Add(theirMovement),
			Distance: theirMovement.Len(),
		}, 0
	}

	minResult := Result{
		Distance: theirMovement.Len(),
		Position: theirPosition.Add(theirMovement),
	}
	var minMask Mask

	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			for z := minZ; z <= maxZ; z++ {
				if !grid.AreCoordsValid(x, y, z) {
					continue
				}
				if t := grid.FlattenGridPos(x, y, z); grid.cells[t].Layer.On(filters) {
					// Resolve collision against this tile
					tileCenter := grid.GridToWorldPos(x, y, z, true)
					tileShape := grid.shapes[grid.cells[t].ShapeId]

					res := theirShape.Sweep(theirPositionRelative, theirMovement, tileCenter, tileShape)
					if res.Hit && res.Distance < minResult.Distance {
						minResult = res
						minMask = grid.cells[t].Layer
					}
				}
			}
		}
	}

	if minResult.Hit {
		minResult.Position = minResult.Position.Add(myPosition)
	}
	return minResult, minMask
}

func (grid *Grid) PushOut(myPosition, theirPosition mgl32.Vec3, theirShape Shape, filter Mask) mgl32.Vec3 {
	var theirPositionRelative mgl32.Vec3 = theirPosition.Sub(myPosition)

	// Iterate over the subset of tiles that the body occupies
	bbox := theirShape.Extents().Translate(theirPositionRelative)
	i, j, k := grid.WorldToGridPos(bbox.Max)
	l, m, n := grid.WorldToGridPos(bbox.Min)
	minX, minY, minZ := max(0, min(i, l)), max(0, min(j, m)), max(0, min(k, n))
	maxX, maxY, maxZ := min(max(i, l), grid.width-1), min(max(j, m), grid.height-1), min(max(k, n), grid.length-1)
	tileCount := (maxX - minX + 1) * (maxZ - minZ + 1) * (maxY - minY + 1)
	if tileCount <= 0 || tileCount >= len(grid.visitBuffer) {
		return mgl32.Vec3{}
	}

	// Visit each tile within the movement range, starting from the one closest to the current position and then proceeding to its neighbors.
	clear(grid.celsChecked)
	var visitBuffer [][3]int = grid.visitBuffer[:tileCount]
	clear(visitBuffer)
	visitQueue := containers.NewRingBuffer(visitBuffer)
	startX, startY, startZ := grid.WorldToGridPos(theirPositionRelative)

	start := [3]int{startX, startY, startZ}
	visitQueue.Enqueue(start)
	grid.celsChecked[start] = true

	push := mgl32.Vec3{}

	for pos, empty := visitQueue.Dequeue(); !empty; pos, empty = visitQueue.Dequeue() {
		if grid.AreCoordsValid(pos[0], pos[1], pos[2]) {
			if t := grid.FlattenGridPos(pos[0], pos[1], pos[2]); grid.cells[t].Layer.On(filter) {
				// Resolve collision against this tile
				tileCenter := grid.GridToWorldPos(pos[0], pos[1], pos[2], true)
				tileShape := grid.shapes[grid.cells[t].ShapeId]

				hit, pushVec := theirShape.PushOutOf(theirPositionRelative.Add(push), tileCenter, tileShape)
				if hit {
					push = push.Add(pushVec)
				}
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

	return push
}

// Returns true if the other body touches a tile in the grid.
func (grid *Grid) OtherBodyTouches(myPosition, theirPosition mgl32.Vec3, theirShape Shape, filter Mask) bool {
	var theirPositionRelative mgl32.Vec3 = theirPosition.Sub(myPosition)

	// Iterate over the subset of tiles that the body occupies
	bbox := theirShape.Extents().Translate(theirPositionRelative)
	if !bbox.Intersects(grid.extents) {
		return false
	}
	i, j, k := grid.WorldToGridPos(bbox.Max)
	l, m, n := grid.WorldToGridPos(bbox.Min)
	minX, minY, minZ := max(0, min(i, l)), max(0, min(j, m)), max(0, min(k, n))
	maxX, maxY, maxZ := min(max(i, l), grid.width-1), min(max(j, m), grid.height-1), min(max(k, n), grid.length-1)

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			for z := minZ; z <= maxZ; z++ {
				if cell := grid.cells[grid.FlattenGridPos(x, y, z)]; cell.Layer.On(filter) {
					// Resolve collision against this tile
					tileCenter := grid.GridToWorldPos(x, y, z, true)
					if theirShape.Touches(theirPositionRelative, tileCenter, grid.shapes[cell.ShapeId]) {
						return true
					}
				}
			}
		}
	}

	return false
}

// Walks contiguous 2D areas of the grid at the given y coordinate that are not marked with zone ID -1 and gives them each a unique zone ID.
// This can be used, for instance, to detect sounds from within the same room.
func (grid *Grid) MarkZonesXZ(y int) {
	nRows := grid.Length()
	nCols := grid.Width()

	nextZoneId := 1
	for z := range nRows {
		for x := range nCols {
			if grid.GetZoneAt(x, y, z) == 0 {
				grid.floodFillZone(x, y, z, x, y, z, nextZoneId)
				nextZoneId++
			}
		}
	}
	zoneCount := nextZoneId - 1
	grid.zoneConnections = make([]bool, zoneCount*zoneCount)
	grid.zoneCount = zoneCount
	for i := range zoneCount {
		// Mark all zones as connected to themselves.
		grid.zoneConnections[i+(i*grid.zoneCount)] = true
	}
}

func (grid *Grid) ConnectZones(zone1, zone2 int) {
	if zone1 == zone2 {
		return
	}
	if grid.zoneConnections == nil {
		failure.LogErrWithLocation("tried to connect zones before zones were initialized")
		return
	}
	if zone1 < 1 || zone2 < 1 || zone1 > grid.zoneCount || zone2 > grid.zoneCount {
		failure.LogErrWithLocation("cannot connect zones %v and %v", zone1, zone2)
		return
	}
	grid.zoneConnections[(zone1-1)+((zone2-1)*grid.zoneCount)] = true
}

func (grid *Grid) DisconnectZones(zone1, zone2 int) {
	if zone1 == zone2 {
		return
	}
	if grid.zoneConnections == nil {
		failure.LogErrWithLocation("tried to connect zones before zones were initialized")
		return
	}
	if zone1 < 1 || zone2 < 1 || zone1 > grid.zoneCount || zone2 > grid.zoneCount {
		failure.LogErrWithLocation("cannot disconnect zones %v and %v", zone1, zone2)
		return
	}
	grid.zoneConnections[(zone1-1)+((zone2-1)*grid.zoneCount)] = false
}

func (grid *Grid) AreZonesConnected(zone1, zone2 int) bool {
	if zone1 == zone2 {
		return true
	}
	if grid.zoneConnections == nil {
		failure.LogErrWithLocation("tried to query zones before zones were initialized")
		return false
	}
	if zone1 < 1 || zone2 < 1 || zone1 > grid.zoneCount || zone2 > grid.zoneCount {
		failure.LogErrWithLocation("cannot query zones %v and %v", zone1, zone2)
		return false
	}
	return grid.zoneConnections[(zone1-1)+((zone2-1)*grid.zoneCount)]
}

func (grid *Grid) floodFillZone(x, y, z, px, py, pz, zoneId int) {
	if !grid.AreCoordsValid(x, y, z) || grid.GetZoneAt(x, y, z) != 0 {
		return
	}
	grid.SetZoneAt(x, y, z, zoneId)
	nx, ny, nz := x-1, y, z
	if nx != px || ny != py || nz != pz {
		grid.floodFillZone(nx, ny, nz, x, y, z, zoneId)
	}
	nx, ny, nz = x+1, y, z
	if nx != px || ny != py || nz != pz {
		grid.floodFillZone(nx, ny, nz, x, y, z, zoneId)
	}
	nx, ny, nz = x, y, z-1
	if nx != px || ny != py || nz != pz {
		grid.floodFillZone(nx, ny, nz, x, y, z, zoneId)
	}
	nx, ny, nz = x, y, z+1
	if nx != px || ny != py || nz != pz {
		grid.floodFillZone(nx, ny, nz, x, y, z, zoneId)
	}
}

// func (grid *Grid) MarkZones(y int) {
// 	type zoneSpan struct {
// 		startX, endX int
// 		z            int
// 		id           int
// 		parentId     int
// 	}
// 	nRows := grid.Length()
// 	nCols := grid.Width()

// 	// Find contiguous horizontal spans on each row of the tile map
// 	spans := make([]zoneSpan, 0, nRows*2)
// 	for z := range nRows {
// 		currSpan := maybe.None[zoneSpan]()
// 		for x := range nCols {
// 			if grid.GetZoneAt(x, y, z) == 0 {
// 				span, ok := currSpan.Get()
// 				if ok {
// 					span.endX = x
// 					grid.SetZoneAt(x, y, z, span.id)
// 				} else {
// 					newSpan := zoneSpan{startX: x, z: z, id: len(spans) + 1}
// 					currSpan = maybe.Some(newSpan)
// 					grid.SetZoneAt(x, y, z, newSpan.id)
// 				}
// 			} else if span, ok := currSpan.Value(); ok {
// 				spans = append(spans, span)
// 				currSpan = maybe.None[zoneSpan]()
// 			}
// 		}
// 		if span, ok := currSpan.Value(); ok {
// 			span.endX = nCols - 1
// 			spans = append(spans, span)
// 		}
// 	}

// 	// for _, span := range spans {
// 	// 	fmt.Printf("horizontal span #%v at z %v spans from %v to %v\n", span.id, span.z, span.startX, span.endX)
// 	// }

// 	// Connect spans from top to bottom
// 	for _, span := range spans {
// 		if span.z >= nRows-1 {
// 			continue
// 		}
// 		for x := span.startX; x <= span.endX; x++ {
// 			zBelow := span.z + 1
// 			if zoneBelow := grid.GetZoneAt(x, y, zBelow); zoneBelow > 0 && zoneBelow <= len(spans) {
// 				spanBelow := &spans[zoneBelow-1]
// 				if spanBelow.parentId == 0 {
// 					spanBelow.parentId = span.id
// 				}
// 			}
// 		}
// 	}

// 	// fmt.Println("after top to bottom assimilation:")
// 	// for _, span := range spans {
// 	// 	fmt.Printf("horizontal span #%v at z %v spans from %v to %v\n", span.id, span.z, span.startX, span.endX)
// 	// }

// 	// Connect spans from bottom up
// 	for _, span := range slices.Backward(spans) {
// 		if span.z <= 0 {
// 			continue
// 		}
// 		for x := span.startX; x <= span.endX; x++ {
// 			zAbove := span.z - 1
// 			if zoneAbove := grid.GetZoneAt(x, y, zAbove); zoneAbove > 0 && zoneAbove <= len(spans) {
// 				spanAbove := &spans[zoneAbove-1]
// 				if spanAbove.parentId == 0 {
// 					spanAbove.parentId = span.id
// 				}
// 			}
// 		}
// 	}

// 	// fmt.Println("after bottom to top assimilation:")
// 	// for _, span := range spans {
// 	// 	fmt.Printf("horizontal span #%v at z %v spans from %v to %v\n", span.id, span.z, span.startX, span.endX)
// 	// }

// 	// Go over the tiles marked by the spans and set their zone IDs to the earliest ancestor's.
// 	for _, span := range spans {
// 		rootSpan := span
// 		for rootSpan.parentId != 0 {
// 			rootSpan = spans[span.parentId-1]
// 		}
// 		for x := span.startX; x <= span.endX; x++ {
// 			grid.SetZoneAt(x, y, span.z, rootSpan.id)
// 		}
// 	}
// }
