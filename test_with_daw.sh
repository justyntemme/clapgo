#!/bin/bash

# Clear existing logs
rm -f /tmp/clapgo_factory_calls.log ~/clapgo_preset_debug.log

# Create a marker in the logs
echo "=== DAW Test Started at $(date) ===" > /tmp/clapgo_factory_calls.log
echo "=== DAW Test Started at $(date) ===" > ~/clapgo_preset_debug.log

echo "Starting log monitoring..."
echo "Factory calls log: /tmp/clapgo_factory_calls.log"
echo "Preset debug log: ~/clapgo_preset_debug.log"
echo ""
echo "Please:"
echo "1. Open your DAW"
echo "2. Try to load the ClapGo Gain or Synth plugin"
echo "3. Check if presets are visible"
echo "4. Press Ctrl+C when done"
echo ""

# Monitor both logs
tail -f /tmp/clapgo_factory_calls.log ~/clapgo_preset_debug.log