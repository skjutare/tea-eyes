//go:build teaeyes

package main

import tea "github.com/charmbracelet/bubbletea"

// TeaEyesNewModel is the white-box entry point used by tea-eyes for in-process
// testing. It is only compiled when the `teaeyes` build tag is set, so it
// never affects normal builds or runtime behavior.
//
// If your model has constructor arguments, write a helper here that supplies
// sensible defaults for testing.
func TeaEyesNewModel() tea.Model {
	return model{}
}
