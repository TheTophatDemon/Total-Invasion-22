package scene

import (
	"reflect"

	"tophatdemon.com/total-invasion-ii/engine/render"
)

// Given a pointer to a struct, this will find any exported fields that implement StorageOps and call Update(deltaTime) on them.
func UpdateStores(scene any, deltaTime float32) {
	ForEachStorageField(scene, func(storage StorageOps) {
		storage.Update(deltaTime)
	})
}

// Given a pointer to a struct, this will find any exported fields that implement StorageOps and call Render(context) on them.
func RenderStores(scene any, context *render.Context) {
	ForEachStorageField(scene, func(storage StorageOps) {
		storage.Render(context)
	})
}

// Given a pointer to a struct, this will find any exported fields that implement StorageOps and call TearDown() on them.
func TearDownStores(scene any) {
	ForEachStorageField(scene, func(storage StorageOps) {
		storage.TearDown()
	})
}

// Runs the given function on the value of every exported field in the given struct pointer that implements StorageOps.
func ForEachStorageField(scene any, do func(StorageOps)) {
	sceneVal := reflect.ValueOf(scene).Elem()
	if sceneVal.Kind() != reflect.Struct {
		panic("this ain't a struct!")
	}
	for f := range sceneVal.NumField() {
		fieldVal := sceneVal.Field(f)
		if !fieldVal.CanAddr() || !sceneVal.Type().Field(f).IsExported() {
			continue
		}
		var storage StorageOps
		var ok bool
		if storage, ok = fieldVal.Addr().Interface().(StorageOps); !ok {
			if storage, ok = fieldVal.Interface().(StorageOps); !ok {
				continue
			}
		}
		do(storage)
	}
}
