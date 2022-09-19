package ecs

import (
	"testing"
)

const MAX_ENTS = 16

type Camera struct {}
type Transform struct {
	dirty bool
}

var world *World
var cameras *ComponentStorage[Camera]
var transforms *ComponentStorage[Transform]

func init() {
	world = CreateWorld(MAX_ENTS)
	cameras = CreateStorage[Camera]()
	transforms = CreateStorage[Transform]()
}

func TestCumulative(t *testing.T) {
	camEnt := world.NewEnt()
	camID, _ := camEnt.Split()
	_, err := cameras.Assign(camEnt, Camera{})
	if err != nil || !cameras.Has(camEnt) {
		t.Fatalf("Could not assign component; %v", err)
	}

	trans, err := transforms.Assign(camEnt, Transform{})
	// trans, err = transforms.Get(camEnt)
	trans.dirty = true

	trans, err = transforms.Get(camEnt)
	if !trans.dirty {
		t.Fatal("Failed to mutate component.")
	}

	err = cameras.Unassign(camEnt)
	if err != nil || cameras.Has(camEnt) {
		t.Fatal("Couldn't remove component.")
	}

	if !world.KillEnt(camEnt) {
		t.Fatal("Couldn't kill entity")
	}

	blamEnt := world.NewEnt()
	blamID, _ := blamEnt.Split()
	if blamID != camID {
		t.Fatal("Entity ID was not reused properly.")
	}
}