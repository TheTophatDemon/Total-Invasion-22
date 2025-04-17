package tree

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

const BSP_MAX_DEPTH = 10
const BSP_MIN_OBJECTS = 2

// A BSP tree that splits objects into sections on the X and Z axes to speed up collision detection.
type BspTree struct {
	nodes []bspNode
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
	// Covers spheres and cylinders
	type ShapeWithRadius interface {
		Radius() float32
	}

	switch sh := shape.(type) {
	case ShapeWithRadius:
		touchesRight = shapePosition[node.splitAxis]+sh.Radius() >= node.planeOffset
		touchesLeft = shapePosition[node.splitAxis]-sh.Radius() <= node.planeOffset
	case collision.Box, collision.Grid:
		touchesRight = shapePosition[node.splitAxis]+sh.Extents().Max[node.splitAxis] >= node.planeOffset
		touchesLeft = shapePosition[node.splitAxis]+sh.Extents().Min[node.splitAxis] <= node.planeOffset
	case collision.Mesh:
		for _, tri := range sh.Triangles() {
			if touchesLeft && touchesRight {
				break
			}
			for _, vert := range tri {
				coord := vert[node.splitAxis] + shapePosition[node.splitAxis]
				if coord >= node.planeOffset {
					touchesRight = true
				}
				if coord <= node.planeOffset {
					touchesLeft = true
				}
			}
		}
	default:
		panic("bspNode.TouchesChild must be implemented for " + sh.String())
	}

	return
}

func BuildBspTree(bodiesIter comps.BodyIter, exception comps.HasBody) BspTree {
	// Collect iterator into slice that we can sort independently
	bodies := make([]scene.Handle, 0)
	for {
		ent, handle := bodiesIter.Next()
		if ent == nil || ent == exception {
			break
		}
		bodies = append(bodies, handle)
	}

	tree := BspTree{
		nodes: make([]bspNode, 0, len(bodies)),
	}

	// Uncomment this to put all the objects in the root node in order
	// to bypass the BSP tree when debugging.
	// tree.nodes = append(tree.nodes, bvhNode{
	// 	leftChildIdx:  -1,
	// 	rightChildIdx: -1,
	// 	objects:       bodies,
	// })

	tree.buildBvhNode(0, 0, bodies)

	return tree
}

func (tree *BspTree) buildBvhNode(splitAxis, depth int, bodies []scene.Handle) {
	if len(bodies) <= BSP_MIN_OBJECTS || depth >= BSP_MAX_DEPTH {
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
		bodyHaver, ok := scene.Get[comps.HasBody](handle)
		if !ok {
			continue
		}
		avgPos += bodyHaver.Body().Transform.Position()[splitAxis]
	}
	avgPos /= float32(len(bodies))

	node := bspNode{
		splitAxis:     splitAxis,
		planeOffset:   avgPos,
		leftChildIdx:  -1,
		rightChildIdx: -1,
	}

	leftBodies := make([]scene.Handle, 0, len(bodies))
	rightBodies := make([]scene.Handle, 0, len(bodies))
	for _, handle := range bodies {
		bodyHaver, ok := scene.Get[comps.HasBody](handle)
		if !ok {
			continue
		}

		touchesLeft, touchesRight := node.TouchesChild(bodyHaver.Body().Shape, bodyHaver.Body().Transform.Position())
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
func (tree *BspTree) PotentiallyTouchingEnts(pos mgl32.Vec3, shape collision.Shape) []scene.Handle {
	return tree.potentiallyTouchingEntsRecursive(&tree.nodes[0], pos, shape)
}

func (tree *BspTree) potentiallyTouchingEntsRecursive(node *bspNode, pos mgl32.Vec3, shape collision.Shape) []scene.Handle {
	if node.IsLeaf() {
		return node.objects
	}
	res := make([]scene.Handle, 0)
	touchesLeft, touchesRight := node.TouchesChild(shape, pos)
	if node.rightChildIdx >= 0 && touchesRight {
		res = append(res, tree.potentiallyTouchingEntsRecursive(&tree.nodes[node.rightChildIdx], pos, shape)...)
	}
	if node.leftChildIdx >= 0 && touchesLeft {
		res = append(res, tree.potentiallyTouchingEntsRecursive(&tree.nodes[node.leftChildIdx], pos, shape)...)
	}
	return res
}
