package collision

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

func SphereSweepBoxCollision(start, end mgl32.Vec3, sphereRadius float32, box math2.Box) RaycastResult {
	var result RaycastResult
	var nearestPlaneDist float32 = math.MaxFloat32

	sweepDir := end.Sub(start)
	if sweepDir.LenSqr() > 0.0 {
		sweepDir = sweepDir.Normalize()
	}

	boxPlanes := [...]math2.Plane{
		math2.PlaneFromPointAndNormal(box.Min, mgl32.Vec3{-1.0, 0.0, 0.0}), // Left side
		math2.PlaneFromPointAndNormal(box.Min, mgl32.Vec3{0.0, -1.0, 0.0}), // Bottom side
		math2.PlaneFromPointAndNormal(box.Min, mgl32.Vec3{0.0, 0.0, -1.0}), // Front side
		math2.PlaneFromPointAndNormal(box.Max, mgl32.Vec3{1.0, 0.0, 0.0}),  // Right side
		math2.PlaneFromPointAndNormal(box.Max, mgl32.Vec3{0.0, 1.0, 0.0}),  // Top side
		math2.PlaneFromPointAndNormal(box.Max, mgl32.Vec3{0.0, 0.0, 1.0}),  // Back side
	}

	for _, plane := range boxPlanes {
		if cast := RayPlaneCollision(start, sweepDir, plane, false); cast.Hit && cast.Distance < nearestPlaneDist {
			nearestPlaneDist = cast.Distance
			result = cast
		}
	}

	if result.Hit {
		spherePos := math2.ClosestPointOnLine(start, end, result.Position)
		projectedPoint := math2.Vec3Max(math2.Vec3Min(spherePos, box.Max), box.Min)
		diff := spherePos.Sub(projectedPoint)
		distSq := diff.LenSqr()
		if distSq > 0.0 && distSq < sphereRadius*sphereRadius {
			dist := math2.Sqrt(distSq)
			return RaycastResult{
				Hit:      true,
				Position: spherePos.Add(sweepDir.Mul(-sphereRadius)),
				Normal:   diff.Mul(1.0 / dist),
				Distance: spherePos.Sub(start).Len(),
			}
		} else if distSq == 0.0 {
			diffToCenter := spherePos.Sub(box.Center())
			distToCenter := diffToCenter.Len()
			return RaycastResult{
				Hit:      true,
				Position: spherePos.Add(sweepDir.Mul(-sphereRadius)),
				Normal:   diffToCenter.Mul(1.0 / distToCenter),
				Distance: spherePos.Sub(start).Len(),
			}
		}
	}

	return RaycastResult{}
}
