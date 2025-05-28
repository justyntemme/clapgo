package controls

import (
	"errors"
	"fmt"
)

// RemoteControlsPageBuilder provides a fluent interface for creating remote control pages
type RemoteControlsPageBuilder struct {
	page    RemoteControlsPage
	err     error
	nextIdx int // Next slot to fill in ParamIDs array
}

// NewRemoteControlsPageBuilder creates a new remote controls page builder
func NewRemoteControlsPageBuilder(pageID uint64, pageName string) *RemoteControlsPageBuilder {
	return &RemoteControlsPageBuilder{
		page: RemoteControlsPage{
			SectionName: "",
			PageID:      pageID,
			PageName:    pageName,
			ParamIDs:    [RemoteControlsCount]uint32{}, // Initialize to all zeros
			IsForPreset: false,
		},
		nextIdx: 0,
	}
}

// Section sets the section name for this page
func (b *RemoteControlsPageBuilder) Section(sectionName string) *RemoteControlsPageBuilder {
	if b.err != nil {
		return b
	}
	b.page.SectionName = sectionName
	return b
}

// ForPreset marks this page as preset-specific
func (b *RemoteControlsPageBuilder) ForPreset() *RemoteControlsPageBuilder {
	if b.err != nil {
		return b
	}
	b.page.IsForPreset = true
	return b
}

// AddParameter adds a parameter to the next available slot
func (b *RemoteControlsPageBuilder) AddParameter(paramID uint32) *RemoteControlsPageBuilder {
	if b.err != nil {
		return b
	}
	
	if b.nextIdx >= RemoteControlsCount {
		b.err = fmt.Errorf("cannot add more than %d parameters per page", RemoteControlsCount)
		return b
	}
	
	b.page.ParamIDs[b.nextIdx] = paramID
	b.nextIdx++
	return b
}

// AddParameters adds multiple parameters to consecutive slots
func (b *RemoteControlsPageBuilder) AddParameters(paramIDs ...uint32) *RemoteControlsPageBuilder {
	if b.err != nil {
		return b
	}
	
	if b.nextIdx+len(paramIDs) > RemoteControlsCount {
		b.err = fmt.Errorf("cannot add %d parameters: would exceed maximum of %d", 
			len(paramIDs), RemoteControlsCount)
		return b
	}
	
	for _, paramID := range paramIDs {
		b.page.ParamIDs[b.nextIdx] = paramID
		b.nextIdx++
	}
	
	return b
}

// SetParameter sets a parameter at a specific slot (0-based index)
func (b *RemoteControlsPageBuilder) SetParameter(slot int, paramID uint32) *RemoteControlsPageBuilder {
	if b.err != nil {
		return b
	}
	
	if slot < 0 || slot >= RemoteControlsCount {
		b.err = fmt.Errorf("slot index %d out of range [0, %d)", slot, RemoteControlsCount)
		return b
	}
	
	b.page.ParamIDs[slot] = paramID
	
	// Update nextIdx if we filled a slot beyond the current position
	if slot >= b.nextIdx {
		b.nextIdx = slot + 1
	}
	
	return b
}

// ClearSlot clears a specific slot
func (b *RemoteControlsPageBuilder) ClearSlot(slot int) *RemoteControlsPageBuilder {
	if b.err != nil {
		return b
	}
	
	if slot < 0 || slot >= RemoteControlsCount {
		b.err = fmt.Errorf("slot index %d out of range [0, %d)", slot, RemoteControlsCount)
		return b
	}
	
	b.page.ParamIDs[slot] = 0 // 0 typically means "no parameter"
	return b
}

// FillRemaining fills all remaining slots with a specific parameter ID
func (b *RemoteControlsPageBuilder) FillRemaining(paramID uint32) *RemoteControlsPageBuilder {
	if b.err != nil {
		return b
	}
	
	for i := b.nextIdx; i < RemoteControlsCount; i++ {
		b.page.ParamIDs[i] = paramID
	}
	b.nextIdx = RemoteControlsCount
	
	return b
}

// ClearRemaining clears all remaining slots
func (b *RemoteControlsPageBuilder) ClearRemaining() *RemoteControlsPageBuilder {
	return b.FillRemaining(0)
}

// Build creates the remote controls page, returning an error if validation fails
func (b *RemoteControlsPageBuilder) Build() (RemoteControlsPage, error) {
	if b.err != nil {
		return RemoteControlsPage{}, b.err
	}
	
	// Final validation
	if b.page.PageName == "" {
		return RemoteControlsPage{}, errors.New("page name is required")
	}
	
	// Check for at least one parameter mapped
	hasParameter := false
	for _, paramID := range b.page.ParamIDs {
		if paramID != 0 {
			hasParameter = true
			break
		}
	}
	
	if !hasParameter {
		return RemoteControlsPage{}, errors.New("page must have at least one parameter mapped")
	}
	
	return b.page, nil
}

// MustBuild creates the remote controls page, panicking on error
func (b *RemoteControlsPageBuilder) MustBuild() RemoteControlsPage {
	page, err := b.Build()
	if err != nil {
		panic(err)
	}
	return page
}

// GetUsedSlotCount returns the number of slots that have been filled
func (b *RemoteControlsPageBuilder) GetUsedSlotCount() int {
	return b.nextIdx
}

// GetAvailableSlotCount returns the number of available slots
func (b *RemoteControlsPageBuilder) GetAvailableSlotCount() int {
	return RemoteControlsCount - b.nextIdx
}

// Common preset configurations

// MainControlsPage creates a main controls page with common parameters
func MainControlsPage(pageID uint64, masterVolume, pan, lowEQ, midEQ, highEQ uint32) *RemoteControlsPageBuilder {
	return NewRemoteControlsPageBuilder(pageID, "Main Controls").
		Section("Master").
		AddParameters(masterVolume, pan, lowEQ, midEQ, highEQ)
}

// EffectControlsPage creates an effects control page
func EffectControlsPage(pageID uint64, effectParams ...uint32) *RemoteControlsPageBuilder {
	builder := NewRemoteControlsPageBuilder(pageID, "Effects").
		Section("FX")
	
	if len(effectParams) > 0 {
		builder = builder.AddParameters(effectParams...)
	}
	
	return builder
}

// FilterControlsPage creates a filter controls page
func FilterControlsPage(pageID uint64, cutoff, resonance, envelope, lfo uint32) *RemoteControlsPageBuilder {
	return NewRemoteControlsPageBuilder(pageID, "Filter").
		Section("Filter").
		AddParameters(cutoff, resonance, envelope, lfo)
}

// EnvelopeControlsPage creates an envelope controls page
func EnvelopeControlsPage(pageID uint64, attack, decay, sustain, release uint32) *RemoteControlsPageBuilder {
	return NewRemoteControlsPageBuilder(pageID, "Envelope").
		Section("ADSR").
		AddParameters(attack, decay, sustain, release)
}

// PresetControlsPage creates a preset-specific controls page
func PresetControlsPage(pageID uint64, pageName string, params ...uint32) *RemoteControlsPageBuilder {
	builder := NewRemoteControlsPageBuilder(pageID, pageName).
		ForPreset()
	
	if len(params) > 0 {
		builder = builder.AddParameters(params...)
	}
	
	return builder
}

// Example usage:
// mainPage := MainControlsPage(0, 
//     volumeParamID, 
//     panParamID, 
//     lowEQParamID, 
//     midEQParamID, 
//     highEQParamID).
//     Build()
//
// filterPage := FilterControlsPage(1,
//     cutoffParamID,
//     resonanceParamID,
//     envelopeParamID,
//     lfoParamID).
//     ClearRemaining(). // Clear unused slots
//     Build()
//
// customPage := NewRemoteControlsPageBuilder(2, "Custom").
//     Section("Synth").
//     SetParameter(0, oscParamID).
//     SetParameter(2, filterParamID). // Skip slot 1
//     SetParameter(4, effectParamID). // Skip slot 3
//     Build()