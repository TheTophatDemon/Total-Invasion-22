package ents

import "tophatdemon.com/total-invasion-ii/engine/color"

// In order to prevent a circular dependency of packages, entities interact with the World through this interface.
// As a bonus, this prevents entities from doing things with the world that they shouldn't, like running a full update.
type WorldOps interface {
	ShowMessage(text string, duration float32, priority int, colr color.Color)
}
