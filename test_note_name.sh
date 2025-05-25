#!/bin/bash

# Test script for Note Name Extension

echo "Testing Note Name Extension..."

# Build the synth example
cd examples/synth
make clean
make

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Build successful!"

# Check if the synth plugin declares note name support
echo "Checking manifest for note name support..."
grep -q '"id": "clap.note-name"' synth.json && grep -A1 '"id": "clap.note-name"' synth.json | grep -q '"supported": true'

if [ $? -eq 0 ]; then
    echo "✓ Note name extension is declared as supported in manifest"
else
    echo "✗ Note name extension is not properly declared in manifest"
    exit 1
fi

# Check if the C exports are defined
echo "Checking for C export symbols..."
nm build/libsynth.so | grep -q "ClapGo_PluginNoteNameCount"
if [ $? -eq 0 ]; then
    echo "✓ ClapGo_PluginNoteNameCount export found"
else
    echo "✗ ClapGo_PluginNoteNameCount export not found"
fi

nm build/libsynth.so | grep -q "ClapGo_PluginNoteNameGet"
if [ $? -eq 0 ]; then
    echo "✓ ClapGo_PluginNoteNameGet export found"
else
    echo "✗ ClapGo_PluginNoteNameGet export not found"
fi

echo "Note Name Extension test complete!"