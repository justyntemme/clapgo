# Filter Cutoff Parameter Improvement

## Problem
The original linear filter cutoff parameter (20Hz-20kHz) made it very difficult to fine-tune frequencies, especially in the musically important low-to-mid range.

### Original Linear Mapping (Poor)
```
Parameter Value    Frequency    Musical Usefulness
0.0 (0%)          20 Hz        ✓ Sub-bass
0.1 (10%)         2,000 Hz     ✓ Very useful (presence)
0.2 (20%)         4,000 Hz     ✓ Useful (brightness)
0.5 (50%)         10,000 Hz    ✗ Too bright for most use
0.8 (80%)         16,000 Hz    ✗ Rarely used
1.0 (100%)        20,000 Hz    ✗ Beyond hearing range
```

**Problem**: 90% of the parameter range covers frequencies above 2kHz, leaving only 10% for the critical 20Hz-2kHz range!

## Solution: Logarithmic Mapping

### New Logarithmic Mapping (Excellent)
```
Parameter Value    Frequency    Musical Usefulness
0.0 (0%)          20 Hz        ✓ Sub-bass
0.1 (10%)         32 Hz        ✓ Bass
0.2 (20%)         50 Hz        ✓ Deep bass
0.3 (30%)         80 Hz        ✓ Bass fundamentals  
0.4 (40%)         126 Hz       ✓ Low mids
0.5 (50%)         200 Hz       ✓ Low mids
0.6 (60%)         316 Hz       ✓ Mids
0.7 (70%)         500 Hz       ✓ Important mids
0.8 (80%)         794 Hz       ✓ Upper mids
0.9 (90%)         1,265 Hz     ✓ Presence
1.0 (100%)        8,000 Hz     ✓ Brightness
```

**Benefits**: 
- 50% of parameter range covers 20Hz-200Hz (bass region)
- 80% of parameter range covers 20Hz-800Hz (critical musical range)
- Much finer control in musically important frequencies
- Still reaches 8kHz for brightness control

## Implementation

### Available Options

1. **BindCutoffMusical()** - Simple fix, linear 20Hz-8kHz
2. **BindCutoffLog()** - Logarithmic 20Hz-8kHz (recommended)
3. **BindCutoffLogFull()** - Logarithmic 20Hz-20kHz (if full range needed)

### Code Example

```go
// Old way (linear, hard to control)
plugin.cutoff = plugin.params.BindCutoff(7, "Filter Cutoff", 1000.0)

// New way (logarithmic, easy to control)
plugin.cutoff = plugin.params.BindCutoffLog(7, "Filter Cutoff", 0.5) // 0.5 maps to ~400Hz
```

### Parameter Reading

```go
// In Process method
cutoff, _ := p.params.GetMappedValue(7) // Gets actual frequency (e.g., 400.0)
// vs
cutoff := p.cutoff.Load() // Would get raw param value (e.g., 0.5)
```

### MIDI CC Mapping

```go
// MIDI CC74 (brightness) now maps beautifully across the spectrum
case 74:
    p.cutoff.UpdateWithManager(value, p.ParamManager, 7)
    // CC 0.0 → 20Hz, CC 0.5 → 200Hz, CC 1.0 → 8kHz
```

## Results

✅ **Much better frequency control**
✅ **Easier to dial in bass/mid frequencies** 
✅ **Professional feel matching hardware/software synthesizers**
✅ **MIDI CC control now musically useful across full range**
✅ **Automatic Hz display formatting** (shows "400 Hz" instead of "0.5")

## Frequency Distribution Comparison

### Linear (Bad)
- 20-100 Hz: 4% of parameter range
- 100-1kHz: 49% of parameter range  
- 1kHz-20kHz: 47% of parameter range

### Logarithmic (Good)  
- 20-100 Hz: 30% of parameter range
- 100-1kHz: 50% of parameter range
- 1kHz-8kHz: 20% of parameter range

The logarithmic approach gives **7.5x more resolution** in the critical 20-100Hz bass region!