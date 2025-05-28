// Package util provides common utility functions for the ClapGo framework.
package util

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// FormatParameterValueDB formats a linear gain value as decibels with appropriate precision
func FormatParameterValueDB(linearValue float64, precision int) string {
	db := LinearToDb(linearValue)
	if math.IsInf(db, -1) {
		return "-∞ dB"
	}
	format := fmt.Sprintf("%%.%df dB", precision)
	return fmt.Sprintf(format, db)
}

// ParseParameterValueDB parses a decibel string (e.g., "-6.0 dB") to a linear value
func ParseParameterValueDB(text string) (float64, error) {
	// Remove "dB" suffix and trim spaces
	text = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(text), "dB"))
	
	// Handle infinity
	if text == "-∞" || text == "-inf" {
		return 0.0, nil
	}
	
	// Parse the numeric value
	db, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0.0, err
	}
	
	return DbToLinear(db), nil
}

// FormatParameterValuePercent formats a normalized value (0-1) as a percentage
func FormatParameterValuePercent(value float64, precision int) string {
	format := fmt.Sprintf("%%.%df%%%%", precision)
	return fmt.Sprintf(format, value*100.0)
}

// ParseParameterValuePercent parses a percentage string (e.g., "75%") to a normalized value
func ParseParameterValuePercent(text string) (float64, error) {
	// Remove "%" suffix and trim spaces
	text = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(text), "%"))
	
	// Parse the numeric value
	percent, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0.0, err
	}
	
	return percent / 100.0, nil
}

// FormatParameterValueHz formats a frequency value with appropriate units (Hz or kHz)
func FormatParameterValueHz(freq float64, precision int) string {
	if freq >= 1000.0 {
		format := fmt.Sprintf("%%.%df kHz", precision)
		return fmt.Sprintf(format, freq/1000.0)
	}
	format := fmt.Sprintf("%%.%df Hz", precision)
	return fmt.Sprintf(format, freq)
}

// ParseParameterValueHz parses a frequency string (e.g., "440 Hz" or "1.5 kHz") to Hz
func ParseParameterValueHz(text string) (float64, error) {
	text = strings.TrimSpace(text)
	
	// Check for kHz
	if strings.HasSuffix(text, "kHz") || strings.HasSuffix(text, "khz") {
		text = strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(text, "kHz"), "khz"))
		khz, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return 0.0, err
		}
		return khz * 1000.0, nil
	}
	
	// Default to Hz
	text = strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(text, "Hz"), "hz"))
	return strconv.ParseFloat(text, 64)
}

// FormatParameterValueTime formats a time value with appropriate units (ms or s)
func FormatParameterValueTime(seconds float64, precision int) string {
	if seconds < 1.0 {
		format := fmt.Sprintf("%%.%df ms", precision)
		return fmt.Sprintf(format, seconds*1000.0)
	}
	format := fmt.Sprintf("%%.%df s", precision)
	return fmt.Sprintf(format, seconds)
}

// ParseParameterValueTime parses a time string (e.g., "100 ms" or "1.5 s") to seconds
func ParseParameterValueTime(text string) (float64, error) {
	text = strings.TrimSpace(text)
	
	// Check for milliseconds
	if strings.HasSuffix(text, "ms") {
		text = strings.TrimSpace(strings.TrimSuffix(text, "ms"))
		ms, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return 0.0, err
		}
		return ms / 1000.0, nil
	}
	
	// Default to seconds
	text = strings.TrimSpace(strings.TrimSuffix(text, "s"))
	return strconv.ParseFloat(text, 64)
}

// FormatParameterValueNote formats a MIDI note number as a note name (e.g., "A4")
func FormatParameterValueNote(noteNumber int) string {
	noteNames := []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
	octave := (noteNumber / 12) - 1
	noteName := noteNames[noteNumber%12]
	return fmt.Sprintf("%s%d", noteName, octave)
}

// ParseParameterValueNote parses a note name (e.g., "A4") to a MIDI note number
func ParseParameterValueNote(text string) (int, error) {
	text = strings.TrimSpace(text)
	if len(text) < 2 {
		return 0, fmt.Errorf("invalid note format")
	}
	
	// Extract note name and octave
	notePart := text[:len(text)-1]
	octavePart := text[len(text)-1:]
	
	// Handle sharp notation
	if len(text) >= 3 && text[1] == '#' {
		notePart = text[:2]
		octavePart = text[2:]
	}
	
	// Parse octave
	octave, err := strconv.Atoi(octavePart)
	if err != nil {
		return 0, fmt.Errorf("invalid octave: %v", err)
	}
	
	// Note name to semitone mapping
	noteMap := map[string]int{
		"C": 0, "C#": 1, "C♯": 1,
		"D": 2, "D#": 3, "D♯": 3,
		"E": 4,
		"F": 5, "F#": 6, "F♯": 6,
		"G": 7, "G#": 8, "G♯": 8,
		"A": 9, "A#": 10, "A♯": 10,
		"B": 11,
	}
	
	semitone, ok := noteMap[strings.ToUpper(notePart)]
	if !ok {
		return 0, fmt.Errorf("invalid note name: %s", notePart)
	}
	
	// Calculate MIDI note number
	return (octave+1)*12 + semitone, nil
}