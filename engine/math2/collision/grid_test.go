package collision

import "testing"

func makeGrid(t *testing.T, zoneIds []int, width, length int) Grid {
	if len(zoneIds) != width*length {
		t.Fatalf("given %v zoneIds, but width and length are %v, %v", len(zoneIds), width, length)
	}
	grid := NewGrid(width, 1, length, 2.0)
	for i, zone := range zoneIds {
		grid.SetZoneAtFlatIndex(i, zone)
	}
	grid.MarkZonesXZ(0)
	return grid
}

func assertGrid(t *testing.T, grid Grid, zoneIds []int) {
	if len(zoneIds) != grid.width*grid.length {
		t.Fatalf("given %v zoneIds, but width and length are %v, %v", len(zoneIds), grid.width, grid.length)
	}
	for i, zone := range zoneIds {
		if grid.cells[i].ZoneId != zone {
			x, _, z := grid.UnflattenGridPos(i)
			t.Fatalf("grid cell did not match at (%v, %v): expected zone ID %v, but got %v", x, z, zone, grid.cells[i].ZoneId)
		}
	}
}

func TestMarkZones(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		assertGrid(
			t,
			makeGrid(t, []int{
				+0, +0, +0,
				+0, -1, +0,
				+0, +0, +0,
			}, 3, 3),
			[]int{
				+1, +1, +1,
				+1, -1, +1,
				+1, +1, +1,
			},
		)
	})
	t.Run("multiple zones", func(t *testing.T) {
		assertGrid(
			t,
			makeGrid(t, []int{
				+0, +0, +0, +0,
				+0, -1, -1, +0,
				+0, -1, +0, -1,
				+0, -1, +0, +0,
			}, 4, 4),
			[]int{
				+1, +1, +1, +1,
				+1, -1, -1, +1,
				+1, -1, +2, -1,
				+1, -1, +2, +2,
			},
		)
	})
	t.Run("enclosures", func(t *testing.T) {
		assertGrid(
			t,
			makeGrid(t, []int{
				-1, -1, -1, +0, +0, +0, +0, -1,
				-1, +0, -1, +0, +0, -1, +0, -1,
				-1, +0, -1, -1, -1, +0, +0, +0,
				-1, -1, -1, +0, -1, -1, +0, +0,
				+0, +0, +0, +0, -1, +0, -1, +0,
				-1, -1, -1, +0, -1, -1, +0, +0,
				+0, +0, -1, -1, -1, +0, -1, +0,
				+0, +0, -1, +0, +0, +0, +0, +0,
			}, 8, 8),
			[]int{
				-1, -1, -1, +1, +1, +1, +1, -1,
				-1, +2, -1, +1, +1, -1, +1, -1,
				-1, +2, -1, -1, -1, +1, +1, +1,
				-1, -1, -1, +3, -1, -1, +1, +1,
				+3, +3, +3, +3, -1, +4, -1, +1,
				-1, -1, -1, +3, -1, -1, +1, +1,
				+5, +5, -1, -1, -1, +1, -1, +1,
				+5, +5, -1, +1, +1, +1, +1, +1,
			},
		)
	})
}
