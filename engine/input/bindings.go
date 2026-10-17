package input

import "structs"

type Binding struct {
	_                                     structs.HostLayout
	wasPressed, justPressed, justReleased bool
	lastUpdatedFrame                      uint64
	variant                               [2]byte
}

// Returns true if the input is being activated during this frame
func (binding *Binding) Pressed() bool {
	return inputBindingPressed(binding)
}

// Returns true if the input just started activating on this frame
func (binding *Binding) JustPressed() bool {
	return inputBindingJustPressed(binding)
}

// Returns true if the input just stopped activating on this frame
func (binding *Binding) JustReleased() bool {
	return inputBindingJustReleased(binding)
}

// Returns the input state as a float from -1 to 1
func (binding *Binding) Axis() float32 {
	return inputBindingAxis(binding)
}

// Key for the display name of the binding.
func (binding *Binding) LocalizationKey() string {
	return "" //TODO
}
