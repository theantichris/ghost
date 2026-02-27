package ui

import (
	"slices"

	"charm.land/bubbles/v2/key"
)

type keyMap struct {
	esc        key.Binding
	enter      key.Binding
	up         key.Binding
	down       key.Binding
	newline    key.Binding
	command    key.Binding
	insert     key.Binding
	pageDown   key.Binding
	pageUp     key.Binding
	goToTop    key.Binding
	goToBottom key.Binding
	new        key.Binding
	quit       key.Binding
	readFile   key.Binding
	threadList key.Binding
}

// matchesCommand is a helper to match the command string to a key.
func matchesCommand(cmd string, binding key.Binding) bool {
	return slices.Contains(binding.Keys(), cmd)
}
