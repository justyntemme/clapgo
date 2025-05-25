package api

import (
	"testing"
)

// BenchmarkIntegrationAudioProcessing simulates complete audio processing
func BenchmarkIntegrationAudioProcessing(b *testing.B) {
	// Setup plugin simulation
	pm := NewParameterManager()
	for i := uint32(0); i < 10; i++ {
		pm.RegisterParameter(ParamInfo{
			ID:           i,
			Name:         "Param",
			MinValue:     0.0,
			MaxValue:     1.0,
			DefaultValue: 0.5,
		})
	}
	
	eventPool := NewEventPool()
	// Note: We can't actually use the logger in benchmarks due to CGO restrictions
	
	// Audio buffers (1024 frames, stereo)
	const frameCount = 1024
	inputL := make([]float32, frameCount)
	inputR := make([]float32, frameCount)
	outputL := make([]float32, frameCount)
	outputR := make([]float32, frameCount)
	
	// Fill input with test signal
	for i := 0; i < frameCount; i++ {
		inputL[i] = float32(i) / frameCount
		inputR[i] = float32(i) / frameCount
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		// Simulate event processing
		for j := 0; j < 10; j++ {
			event := eventPool.GetParamValueEvent()
			event.ParamID = uint32(j)
			event.Value = float64(i%100) / 100.0
			
			// Process parameter change
			pm.SetParameterValue(event.ParamID, event.Value)
			
			eventPool.ReturnParamValueEvent(event)
		}
		
		// Simulate audio processing
		gain := pm.GetParameterValue(0)
		for j := 0; j < frameCount; j++ {
			outputL[j] = inputL[j] * float32(gain)
			outputR[j] = inputR[j] * float32(gain)
		}
		
		// Simulate occasional parameter update (1% of buffers)
		if i%100 == 0 {
			pm.SetParameterValue(0, float64(i%1000)/1000.0)
		}
	}
}

// BenchmarkIntegrationNoteProcessing simulates polyphonic note processing
func BenchmarkIntegrationNoteProcessing(b *testing.B) {
	eventPool := NewEventPool()
	
	// Simulate voice allocation
	type Voice struct {
		noteID    int32
		frequency float64
		amplitude float64
		active    bool
	}
	
	const maxVoices = 32
	voices := make([]Voice, maxVoices)
	activeVoices := 0
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		// Process note events
		for j := 0; j < 4; j++ {
			if j%2 == 0 {
				// Note on
				noteOn := eventPool.GetNoteEvent()
				noteOn.NoteID = int32(j)
				noteOn.Key = int16(60 + j)
				noteOn.Velocity = 0.8
				
				// Allocate voice
				if activeVoices < maxVoices {
					voices[activeVoices] = Voice{
						noteID:    noteOn.NoteID,
						frequency: 440.0 * float64(noteOn.Key-69) / 12.0,
						amplitude: noteOn.Velocity,
						active:    true,
					}
					activeVoices++
				}
				
				eventPool.ReturnNoteEvent(noteOn)
			} else {
				// Note off
				noteOff := eventPool.GetNoteEvent()
				noteOff.NoteID = int32(j - 1)
				
				// Find and deactivate voice
				for k := 0; k < activeVoices; k++ {
					if voices[k].noteID == noteOff.NoteID {
						voices[k].active = false
						// Simple voice stealing - swap with last
						if k < activeVoices-1 {
							voices[k] = voices[activeVoices-1]
						}
						activeVoices--
						break
					}
				}
				
				eventPool.ReturnNoteEvent(noteOff)
			}
		}
		
		// Simulate audio generation for active voices
		const frameCount = 256
		output := make([]float32, frameCount)
		
		for j := 0; j < activeVoices; j++ {
			if voices[j].active {
				// Simple sine wave generation (simplified)
				for k := 0; k < frameCount; k++ {
					output[k] += float32(voices[j].amplitude * 0.1)
				}
			}
		}
	}
}