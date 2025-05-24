#!/bin/bash

echo "Testing CLAP Preset Loading with Embedded Presets"
echo "================================================="

# Function to test preset loading using clap-info or similar tool
test_preset_load() {
    local plugin=$1
    local preset=$2
    echo "Testing $plugin with preset '$preset'..."
    
    # Since we don't have a direct CLAP preset testing tool, let's verify the structure
    echo "- Checking plugin binary size (should be larger with embedded presets)..."
    ls -lh build/examples/$plugin/$plugin.clap
    
    echo "- Checking for embedded preset data..."
    if strings build/examples/$plugin/lib$plugin.so | grep -q "$preset"; then
        echo "  ✓ Found preset '$preset' embedded in binary"
    else
        echo "  ✗ Preset '$preset' not found in binary"
    fi
}

# Test synth presets
echo -e "\n1. Testing Synth Plugin Presets"
echo "-------------------------------"
test_preset_load "synth" "lead"
test_preset_load "synth" "pad"

# Test gain presets  
echo -e "\n2. Testing Gain Plugin Presets"
echo "------------------------------"
test_preset_load "gain" "unity"
test_preset_load "gain" "boost"

# Check plugin initialization logs
echo -e "\n3. Running clap-validator to check plugin initialization"
echo "-------------------------------------------------------"
echo "This should show 'Available bundled presets' in the logs..."

# Run validator on synth to see if it logs available presets
timeout 5s clap-validator validate build/examples/synth/synth.clap 2>&1 | grep -E "(Available bundled presets|Synth plugin initialized)" || true

echo -e "\nPreset loading test complete!"