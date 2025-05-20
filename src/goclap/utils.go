package goclap

import (
	"fmt"
	"strconv"
	"strings"
)

// formatInt formats an integer value as a string
func formatInt(value int) string {
	return fmt.Sprintf("%d", value)
}

// formatFloat formats a float value with the given precision
func formatFloat(value float64, precision int) string {
	return fmt.Sprintf("%.*f", precision, value)
}

// parseInt parses an integer from a string
func parseInt(text string) (int, bool) {
	// Remove any whitespace
	text = strings.TrimSpace(text)
	
	// Try to parse as integer
	val, err := strconv.Atoi(text)
	if err != nil {
		return 0, false
	}
	
	return val, true
}

// parseFloat parses a float from a string
func parseFloat(text string) (float64, bool) {
	// Remove any whitespace
	text = strings.TrimSpace(text)
	
	// Try to parse as float
	val, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0, false
	}
	
	return val, true
}