package input

type Binding interface {
	// Returns true if the input is being activated during this frame
	Pressed() bool
	// Returns true if the input just started activating on this frame
	JustPressed() bool
	// Returns true if the input just stopped activating on this frame
	JustReleased() bool
	// Returns the input state as a float from -1 to 1
	Axis() float32
	// Key for the display name of the binding.
	LocalizationKey() string
}

type bindingBase struct {
	wasPressed, justPressed, justReleased bool
	lastUpdatedFrame                      uint64
}

func (binding *bindingBase) updatePressStates(isPressed bool) bool {
	if binding.lastUpdatedFrame != inputFrameNumber {
		binding.justPressed = isPressed && !binding.wasPressed
		binding.justReleased = !isPressed && binding.wasPressed
		binding.wasPressed = isPressed
		binding.lastUpdatedFrame = inputFrameNumber
	}
	return isPressed
}
