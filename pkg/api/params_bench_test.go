package api

import (
	"testing"
)

// BenchmarkParameterGet tests parameter read performance
func BenchmarkParameterGet(b *testing.B) {
	pm := NewParameterManager()
	
	// Register 100 parameters
	for i := uint32(0); i < 100; i++ {
		pm.RegisterParameter(ParamInfo{
			ID:           i,
			Name:         "Param",
			MinValue:     0.0,
			MaxValue:     1.0,
			DefaultValue: 0.5,
		})
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		// Read parameters in sequence
		for j := uint32(0); j < 100; j++ {
			_ = pm.GetParameterValue(j)
		}
	}
}

// BenchmarkParameterSet tests parameter write performance
func BenchmarkParameterSet(b *testing.B) {
	pm := NewParameterManager()
	
	// Register 100 parameters
	for i := uint32(0); i < 100; i++ {
		pm.RegisterParameter(ParamInfo{
			ID:           i,
			Name:         "Param",
			MinValue:     0.0,
			MaxValue:     1.0,
			DefaultValue: 0.5,
		})
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		// Write parameters in sequence
		for j := uint32(0); j < 100; j++ {
			pm.SetParameterValue(j, float64(i%100)/100.0)
		}
	}
}

// BenchmarkParameterSetWithListeners tests parameter updates with listeners
func BenchmarkParameterSetWithListeners(b *testing.B) {
	pm := NewParameterManager()
	
	// Register 10 parameters
	for i := uint32(0); i < 10; i++ {
		pm.RegisterParameter(ParamInfo{
			ID:           i,
			Name:         "Param",
			MinValue:     0.0,
			MaxValue:     1.0,
			DefaultValue: 0.5,
		})
	}
	
	// Add 5 listeners
	listenerCalled := 0
	for i := 0; i < 5; i++ {
		pm.AddChangeListener(func(paramID uint32, oldValue, newValue float64) {
			listenerCalled++
		})
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		// Update all parameters
		for j := uint32(0); j < 10; j++ {
			pm.SetParameterValue(j, float64(i%100)/100.0)
		}
	}
}

// BenchmarkParameterConcurrentAccess tests concurrent parameter access
func BenchmarkParameterConcurrentAccess(b *testing.B) {
	pm := NewParameterManager()
	
	// Register 100 parameters
	for i := uint32(0); i < 100; i++ {
		pm.RegisterParameter(ParamInfo{
			ID:           i,
			Name:         "Param",
			MinValue:     0.0,
			MaxValue:     1.0,
			DefaultValue: 0.5,
		})
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// Mix reads and writes
			if i%2 == 0 {
				pm.GetParameterValue(uint32(i % 100))
			} else {
				pm.SetParameterValue(uint32(i%100), float64(i%100)/100.0)
			}
			i++
		}
	})
}

// BenchmarkParameterValidation tests parameter validation overhead
func BenchmarkParameterValidation(b *testing.B) {
	pm := NewParameterManager()
	
	// Register parameter with range validation
	pm.RegisterParameter(ParamInfo{
		ID:           0,
		Name:         "ValidatedParam",
		MinValue:     0.0,
		MaxValue:     1.0,
		DefaultValue: 0.5,
	})
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		// Set with valid values - automatic range validation happens
		pm.SetParameterValue(0, float64(i%100)/100.0)
	}
}