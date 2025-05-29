package param

import "math"

// ValueMapper defines a function that maps parameter values
type ValueMapper func(paramValue float64) float64

// FrequencyMappers provides common frequency mapping functions
var FrequencyMappers = struct {
    Linear     ValueMapper
    LogMusical ValueMapper
    LogFull    ValueMapper
}{
    // Linear mapping (current behavior)
    Linear: func(paramValue float64) float64 {
        return 20.0 + paramValue*(20000.0-20.0)
    },
    
    // Logarithmic mapping optimized for musical use (20Hz - 8kHz)
    LogMusical: func(paramValue float64) float64 {
        minFreq := 20.0
        maxFreq := 8000.0
        return minFreq * math.Pow(maxFreq/minFreq, paramValue)
    },
    
    // Logarithmic mapping for full audio spectrum (20Hz - 20kHz)
    LogFull: func(paramValue float64) float64 {
        minFreq := 20.0
        maxFreq := 20000.0
        return minFreq * math.Pow(maxFreq/minFreq, paramValue)
    },
}

// ResonanceMappers provides common resonance mapping functions
var ResonanceMappers = struct {
    Linear ValueMapper
    Musical ValueMapper
}{
    // Linear Q mapping (current behavior: 0.5 - 20)
    Linear: func(paramValue float64) float64 {
        return 0.5 + paramValue*19.5
    },
    
    // Musical Q mapping (0.5 - 10, more resolution in useful range)
    Musical: func(paramValue float64) float64 {
        return 0.5 + paramValue*9.5
    },
}

// CreateFrequencyMapper creates a custom frequency mapper
func CreateFrequencyMapper(minHz, maxHz float64, logarithmic bool) ValueMapper {
    if logarithmic {
        return func(paramValue float64) float64 {
            return minHz * math.Pow(maxHz/minHz, paramValue)
        }
    }
    return func(paramValue float64) float64 {
        return minHz + paramValue*(maxHz-minHz)
    }
}

// InverseFrequencyMapper creates the inverse mapping (frequency to param value)
func InverseFrequencyMapper(minHz, maxHz float64, logarithmic bool) ValueMapper {
    if logarithmic {
        logRange := math.Log(maxHz / minHz)
        return func(frequency float64) float64 {
            if frequency <= minHz {
                return 0.0
            }
            if frequency >= maxHz {
                return 1.0
            }
            return math.Log(frequency/minHz) / logRange
        }
    }
    return func(frequency float64) float64 {
        return (frequency - minHz) / (maxHz - minHz)
    }
}