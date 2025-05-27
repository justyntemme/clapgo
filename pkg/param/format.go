package param

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Format specifies how to format parameter values
type Format int

const (
	FormatDefault      Format = iota
	FormatDecibel             // 20 * log10(value) with "dB" suffix
	FormatPercentage          // value * 100 with "%" suffix
	FormatMilliseconds        // value * 1000 with "ms" suffix
	FormatSeconds             // value with "s" suffix
	FormatHertz               // value with "Hz" suffix
	FormatKilohertz           // value / 1000 with "kHz" suffix
)

// FormatValue formats a parameter value according to the specified format
func FormatValue(value float64, format Format) string {
	switch format {
	case FormatDecibel:
		if value <= 0 {
			return "-∞ dB"
		}
		db := 20.0 * math.Log10(value)
		return fmt.Sprintf("%.1f dB", db)
		
	case FormatPercentage:
		return fmt.Sprintf("%.1f%%", value*100.0)
		
	case FormatMilliseconds:
		return fmt.Sprintf("%.0f ms", value*1000.0)
		
	case FormatSeconds:
		return fmt.Sprintf("%.2f s", value)
		
	case FormatHertz:
		return fmt.Sprintf("%.1f Hz", value)
		
	case FormatKilohertz:
		return fmt.Sprintf("%.2f kHz", value/1000.0)
		
	default:
		return fmt.Sprintf("%.3f", value)
	}
}

// Parser handles parsing of formatted parameter values
type Parser struct {
	format Format
}

// NewParser creates a new parameter value parser
func NewParser(format Format) *Parser {
	return &Parser{format: format}
}

// ParseValue parses a formatted string back to a float64 value
func (p *Parser) ParseValue(text string) (float64, error) {
	// Remove extra whitespace
	text = strings.TrimSpace(text)
	
	switch p.format {
	case FormatDecibel:
		// Handle infinity
		if strings.HasPrefix(text, "-∞") || strings.HasPrefix(text, "-inf") {
			return 0, nil
		}
		
		// Remove dB suffix and parse
		re := regexp.MustCompile(`([+-]?\d*\.?\d+)\s*dB?`)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			db, err := strconv.ParseFloat(matches[1], 64)
			if err != nil {
				return 0, err
			}
			// Convert from dB to linear
			return math.Pow(10, db/20.0), nil
		}
		
	case FormatPercentage:
		// Remove % suffix and parse
		re := regexp.MustCompile(`([+-]?\d*\.?\d+)\s*%?`)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			percent, err := strconv.ParseFloat(matches[1], 64)
			if err != nil {
				return 0, err
			}
			return percent / 100.0, nil
		}
		
	case FormatMilliseconds:
		// Remove ms suffix and parse
		re := regexp.MustCompile(`([+-]?\d*\.?\d+)\s*ms?`)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			ms, err := strconv.ParseFloat(matches[1], 64)
			if err != nil {
				return 0, err
			}
			return ms / 1000.0, nil
		}
		
	case FormatSeconds:
		// Remove s suffix and parse
		re := regexp.MustCompile(`([+-]?\d*\.?\d+)\s*s?`)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			return strconv.ParseFloat(matches[1], 64)
		}
		
	case FormatHertz:
		// Remove Hz suffix and parse
		re := regexp.MustCompile(`([+-]?\d*\.?\d+)\s*Hz?`)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			return strconv.ParseFloat(matches[1], 64)
		}
		
	case FormatKilohertz:
		// Remove kHz suffix and parse
		re := regexp.MustCompile(`([+-]?\d*\.?\d+)\s*kHz?`)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			khz, err := strconv.ParseFloat(matches[1], 64)
			if err != nil {
				return 0, err
			}
			return khz * 1000.0, nil
		}
	}
	
	// Default: try to parse as plain number
	return strconv.ParseFloat(text, 64)
}

// ClampValue clamps a value to the given range
func ClampValue(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}