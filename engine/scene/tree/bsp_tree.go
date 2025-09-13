package tree

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/containers"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

// A BSP tree that splits objects into sections on the X and Z axes to speed up collision detection.
type BspTree struct {
	nodes []bspNode
}

type bspNode struct {
	splitAxis                   int
	planeOffset                 float32
	objects                     containers.Set[scene.Handle]
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

func BuildBspTree(bodies containers.Set[scene.Handle]) BspTree {
	tree := BspTree{
		nodes: make([]bspNode, 0, len(bodies)),
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

func (tree *BspTree) buildBvhNode(splitAxis, depth int, bodies containers.Set[scene.Handle]) {
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
	for handle := range bodies {
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

	leftBodies := containers.NewSet[scene.Handle](len(bodies))
	rightBodies := containers.NewSet[scene.Handle](len(bodies))
	for handle := range bodies {
		bodyHaver, ok := scene.Get[comps.HasBody](handle)
		if !ok {
			continue
		}

		touchesLeft, touchesRight := node.TouchesChild(bodyHaver.Body().Shape, bodyHaver.Body().Transform.Position())
		if touchesLeft {
			leftBodies.Add(handle)
		}
		if touchesRight {
			rightBodies.Add(handle)
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
func (tree *BspTree) PotentiallyTouchingEnts(pos mgl32.Vec3, shape collision.Shape) containers.Set[scene.Handle] {
	return tree.potentiallyTouchingEntsRecursive(&tree.nodes[0], pos, shape)
}

func (tree *BspTree) potentiallyTouchingEntsRecursive(node *bspNode, pos mgl32.Vec3, shape collision.Shape) containers.Set[scene.Handle] {
	if node.IsLeaf() {
		return node.objects
	}
	res := containers.NewSet[scene.Handle](0)
	touchesLeft, touchesRight := node.TouchesChild(shape, pos)
	if node.rightChildIdx >= 0 && touchesRight {
		for handle := range tree.potentiallyTouchingEntsRecursive(&tree.nodes[node.rightChildIdx], pos, shape) {
			res.Add(handle)
		}
	}
	if node.leftChildIdx >= 0 && touchesLeft {
		for handle := range tree.potentiallyTouchingEntsRecursive(&tree.nodes[node.leftChildIdx], pos, shape) {
			res.Add(handle)
		}
	}
	return res
}
