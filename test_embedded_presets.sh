#!/bin/bash

# Test embedded presets in the synth plugin
echo "Testing embedded presets..."

# Check if the binary contains the preset data
echo "Checking if presets are embedded in synth plugin..."
if strings build/examples/synth/libsynth.so | grep -q "Lead Synth"; then
    echo "✓ Found 'Lead Synth' preset in binary"
else
    echo "✗ 'Lead Synth' preset not found in binary"
fi

if strings build/examples/synth/libsynth.so | grep -q "Ambient Pad"; then
    echo "✓ Found 'Ambient Pad' preset in binary"
else
    echo "✗ 'Ambient Pad' preset not found in binary"
fi

echo ""
echo "Checking if presets are embedded in gain plugin..."
if strings build/examples/gain/libgain.so | grep -q "Unity Gain"; then
    echo "✓ Found 'Unity Gain' preset in binary"
else
    echo "✗ 'Unity Gain' preset not found in binary"
fi

if strings build/examples/gain/libgain.so | grep -q "6dB Boost"; then
    echo "✓ Found '6dB Boost' preset in binary"
else
    echo "✗ '6dB Boost' preset not found in binary"
fi

echo ""
echo "Embedded presets test complete!"