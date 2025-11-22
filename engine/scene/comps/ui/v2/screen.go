package ui

// Represents a collection of UI elements that are displayed instead of the game
type Screen interface {
	Layout(queue *RenderQueue, deltaTime float32)
}
