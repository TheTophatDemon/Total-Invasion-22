package ui

import "testing"

func TestInvalidChar(t *testing.T) {
	txt := Text{}
	txt.SetText("\r")
	mesh, succ := txt.Mesh()
	if !succ {
		t.Fail()
	}
	if len(mesh.Verts().Pos) > 0 {
		t.Fatalf("Mesh should have no vertices but has %v", len(mesh.Verts().Pos))
	}
}
