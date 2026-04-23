package ui

import "tophatdemon.com/total-invasion-ii/engine/math2"

// Represents a collection of UI elements that are displayed instead of the game
type Screen interface {
	Layout(queue *RenderQueue, deltaTime float32)
	Enter()
	Exit()
	Bounds() math2.Rect
}
