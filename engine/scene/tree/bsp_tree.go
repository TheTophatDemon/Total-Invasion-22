package tree

import (
	"iter"
	"math"
	"slices"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

// A BSP tree that splits objects into sections on the X and Z axes to speed up collision detection.
type BspTree struct {
	nodes   []bspNode
	visited map[scene.Handle]bool
}

type bspNode struct {
	splitAxis                   int
	planeOffset                 float32
	objects                     []scene.Handle
	leftChildIdx, rightChildIdx int
}

func (node bspNode) IsLeaf() bool {
	return node.leftChildIdx < 0 && node.rightChildIdx < 0
}

// Returns whether a shape at a given position intersects with the left or right region of the node.
func (node bspNode) TouchesChild(shape collision.Shape, shapePosition mgl32.Vec3) (touchesLeft, touchesRight bool) {
	touchesRight = shapePosition[node.splitAxis]+shape.Extents().Max[node.splitAxis] >= node.planeOffset
	touchesLeft = shapePosition[node.splitAxis]+shape.Extents().Min[node.splitAxis] <= node.planeOffset
	return
}

func (node bspNode) TouchesRay(rayOrigin, rayDir mgl32.Vec3, maxDist float32) (touchesLeft, touchesRight bool) {
	touchesRight = rayOrigin[node.splitAxis] >= node.planeOffset || rayOrigin[node.splitAxis]+rayDir[node.splitAxis]*maxDist >= node.planeOffset
	touchesLeft = rayOrigin[node.splitAxis] <= node.planeOffset || rayOrigin[node.splitAxis]+rayDir[node.splitAxis]*maxDist <= node.planeOffset
	return
}

func BuildBspTree(bodies []scene.Handle) BspTree {
	tree := BspTree{
		nodes:   make([]bspNode, 0, len(bodies)),
		visited: make(map[scene.Handle]bool, len(bodies)),
	}

	// Uncomment this to put all the objects in the root node in order
	// to bypass the BSP tree when debugging.
	// tree.nodes = append(tree.nodes, bspNode{
	// 	leftChildIdx:  -1,
	// 	rightChildIdx: -1,
	// 	objects:       bodies,
	// })

	tree.buildBvhNode(0, 0, bodies)

	return tree
}

func (tree *BspTree) buildBvhNode(splitAxis, depth int, bodies []scene.Handle) {
	const maxDepth = 10
	const minObjects = 2
	if len(bodies) <= minObjects || depth >= maxDepth {
		// Create leaf node
		node := bspNode{
			leftChildIdx:  -1,
			rightChildIdx: -1,
			objects:       bodies,
		}
		tree.nodes = append(tree.nodes, node)
		return
	}

	var avgPos float32 = 0.0
	for _, handle := range bodies {
		bodyHaver, ok := handle.Get[comps.HasBody]()
		if !ok {
			continue
		}
		avgPos += bodyHaver.Body().Position[splitAxis]
	}
	avgPos /= float32(len(bodies))

	node := bspNode{
		splitAxis:     splitAxis,
		planeOffset:   avgPos,
		leftChildIdx:  -1,
		rightChildIdx: -1,
	}

	leftBodies := make([]scene.Handle, 0)
	rightBodies := make([]scene.Handle, 0)
	for _, handle := range bodies {
		bodyHaver, ok := handle.Get[comps.HasBody]()
		if !ok {
			continue
		}

		touchesLeft, touchesRight := node.TouchesChild(bodyHaver.Body().Shape, bodyHaver.Body().Position)
		if touchesLeft {
			leftBodies = append(leftBodies, handle)
		}
		if touchesRight {
			rightBodies = append(rightBodies, handle)
		}
	}

	nodeIndex := len(tree.nodes)
	tree.nodes = append(tree.nodes, node)

	var nextSplitAxis int
	if splitAxis == 0 {
		nextSplitAxis = 2
	} else {
		nextSplitAxis = 0
	}
	if len(leftBodies) > 0 {
		tree.nodes[nodeIndex].leftChildIdx = len(tree.nodes)
		tree.buildBvhNode(nextSplitAxis, depth+1, leftBodies)
	}
	if len(rightBodies) > 0 {
		tree.nodes[nodeIndex].rightChildIdx = len(tree.nodes)
		tree.buildBvhNode(nextSplitAxis, depth+1, rightBodies)
	}

}

// Returns handles to entities with physics bodies that are in the leaves of the BSP tree where the given
// collision shape is residing.
func (tree *BspTree) PotentiallyTouchingEnts(pos mgl32.Vec3, shape collision.Shape) iter.Seq[scene.Handle] {
	return func(yield func(scene.Handle) bool) {
		if tree == nil || len(tree.nodes) == 0 {
			return
		}
		clear(tree.visited)
		stack := make([]bspNode, 0, 16)
		stack = append(stack, tree.nodes[0])
		for len(stack) > 0 {
			last := len(stack) - 1
			node := stack[last]
			stack = stack[:last]
			if node.IsLeaf() {
				for _, handle := range node.objects {
					if tree.visited[handle] {
						continue
					}
					tree.visited[handle] = true
					if !yield(handle) {
						return
					}
				}
			} else {
				touchesLeft, touchesRight := node.TouchesChild(shape, pos)
				if node.rightChildIdx >= 0 && touchesRight {
					stack = append(stack, tree.nodes[node.rightChildIdx])
				}
				if node.leftChildIdx >= 0 && touchesLeft {
					stack = append(stack, tree.nodes[node.leftChildIdx])
				}
			}
		}
	}
}

func (tree *BspTree) PotentiallyCrossingEnts(rayOrigin, rayDir mgl32.Vec3, maxDist float32) iter.Seq[scene.Handle] {
	return func(yield func(scene.Handle) bool) {
		if tree == nil || len(tree.nodes) == 0 {
			return
		}
		clear(tree.visited)
		stack := make([]bspNode, 0, 16)
		stack = append(stack, tree.nodes[0])
		for len(stack) > 0 {
			last := len(stack) - 1
			node := stack[last]
			stack = stack[:last]
			if node.IsLeaf() {
				for _, handle := range node.objects {
					if tree.visited[handle] {
						continue
					}
					tree.visited[handle] = true
					if !yield(handle) {
						return
					}
				}
			} else {
				touchesLeft, touchesRight := node.TouchesRay(rayOrigin, rayDir, maxDist)
				if node.rightChildIdx >= 0 && touchesRight {
					stack = append(stack, tree.nodes[node.rightChildIdx])
				}
				if node.leftChildIdx >= 0 && touchesLeft {
					stack = append(stack, tree.nodes[node.leftChildIdx])
				}
			}
		}
	}
}

// Returns the vector that needs to be added to the body's movement to avoid obstacles in the tree.
func (tree *BspTree) ResolveCollisions(
	body *comps.Body,
	movement mgl32.Vec3,
	lockY bool,
	filter collision.Mask,
) mgl32.Vec3 {
	if body == nil || body.Layers == 0 {
		return mgl32.Vec3{}
	}

	push := mgl32.Vec3{}

	{
		shrinkyShape := body.Shape.ShrunkenBy(0.1)
		obstacles := slices.Collect(tree.PotentiallyTouchingEnts(body.Position, body.Shape))
		smallestRes := collision.Result{
			Distance: math.MaxFloat32,
		}
		for _, handle := range obstacles {
			if collidingEnt, ok := handle.Get[comps.HasBody](); ok {
				otherBody := collidingEnt.Body()
				if otherBody == nil || body == otherBody || !otherBody.Layers.On(filter) {
					continue
				}

				res := shrinkyShape.Sweep(body.Position, movement, otherBody.Position, otherBody.Shape)
				if res.Hit && res.Distance < smallestRes.Distance {
					smallestRes = res
				}
			}
		}
		if smallestRes.Hit {
			speed := movement.Len()
			slide := smallestRes.Normal.Mul(-movement.Dot(smallestRes.Normal)).Add(movement)
			if slide != (mgl32.Vec3{}) {
				slide.Mul(speed / slide.Len())
			}
			push = push.Add(slide.Sub(movement))
		}
	}

	for range 4 {
		smallestPush := mgl32.Vec3{}
		smallestPushLenSq := float32(math.MaxFloat32)
		nextPos := body.Position.Add(movement).Add(push)
		obstacles := slices.Collect(tree.PotentiallyTouchingEnts(nextPos, body.Shape))
		for _, handle := range obstacles {
			if collidingEnt, ok := handle.Get[comps.HasBody](); ok {
				otherBody := collidingEnt.Body()
				if otherBody == nil || body == otherBody || !otherBody.Layers.On(filter) {
					continue
				}

				// Bounding box check
				bbox := body.Shape.Extents().Translate(nextPos)
				if !bbox.Intersects(otherBody.Shape.Extents().Translate(otherBody.Position)) {
					continue
				}

				hit, pushVec := body.Shape.PushOutOf(nextPos, otherBody.Position, otherBody.Shape)
				if hit {
					pushVecLenSq := pushVec.LenSqr()
					if pushVecLenSq < smallestPushLenSq {
						smallestPush = pushVec
						smallestPushLenSq = pushVecLenSq
					}
				}
			}
		}
		if smallestPush == (mgl32.Vec3{}) {
			break
		}
		push = push.Add(smallestPush)
	}

	if lockY {
		// Restrict movement to the XZ plane
		push[1] = 0.0
	}

	return push
}
