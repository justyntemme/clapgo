package main

// Example showing how to implement custom note names for a drum plugin

import (
	"github.com/justyntemme/clapgo/pkg/api"
)

// DrumNoteNameProvider implements custom note names for a drum kit
type DrumNoteNameProvider struct {
	noteNames []api.NoteName
}

// NewDrumNoteNameProvider creates a provider with General MIDI drum note names
func NewDrumNoteNameProvider() *DrumNoteNameProvider {
	return &DrumNoteNameProvider{
		noteNames: api.StandardDrumNoteNames,
	}
}

// GetNoteNameCount returns the number of custom note names
func (d *DrumNoteNameProvider) GetNoteNameCount() uint32 {
	return uint32(len(d.noteNames))
}

// GetNoteName returns the note name at the given index
func (d *DrumNoteNameProvider) GetNoteName(index uint32) *api.NoteName {
	if int(index) >= len(d.noteNames) {
		return nil
	}
	return &d.noteNames[index]
}

// Example of custom note names for a specific sampler/drum machine
var CustomDrumNoteNames = []api.NoteName{
	{Name: "Kick", Port: -1, Key: 36, Channel: -1},
	{Name: "Snare", Port: -1, Key: 38, Channel: -1},
	{Name: "Closed Hat", Port: -1, Key: 42, Channel: -1},
	{Name: "Open Hat", Port: -1, Key: 46, Channel: -1},
	{Name: "Crash", Port: -1, Key: 49, Channel: -1},
	{Name: "Ride", Port: -1, Key: 51, Channel: -1},
	{Name: "Tom 1", Port: -1, Key: 48, Channel: -1},
	{Name: "Tom 2", Port: -1, Key: 45, Channel: -1},
	{Name: "Floor Tom", Port: -1, Key: 41, Channel: -1},
	{Name: "Cowbell", Port: -1, Key: 56, Channel: -1},
	{Name: "Clap", Port: -1, Key: 39, Channel: -1},
	{Name: "Tambourine", Port: -1, Key: 54, Channel: -1},
}

// Example exports for a drum plugin
/*
//export ClapGo_PluginNoteNameCount
func ClapGo_PluginNoteNameCount(plugin unsafe.Pointer) C.uint32_t {
	if plugin == nil {
		return 0
	}
	p := cgo.Handle(plugin).Value().(*DrumPlugin)
	
	if p.noteNameProvider != nil {
		return C.uint32_t(p.noteNameProvider.GetNoteNameCount())
	}
	return 0
}

//export ClapGo_PluginNoteNameGet
func ClapGo_PluginNoteNameGet(plugin unsafe.Pointer, index C.uint32_t, noteName unsafe.Pointer) C.bool {
	if plugin == nil || noteName == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*DrumPlugin)
	
	if p.noteNameProvider == nil {
		return C.bool(false)
	}
	
	name := p.noteNameProvider.GetNoteName(uint32(index))
	if name == nil {
		return C.bool(false)
	}
	
	// Convert to C structure
	api.NoteNameToC(name, noteName)
	
	return C.bool(true)
}
*/