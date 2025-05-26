#!/bin/bash

# Script to help test preset discovery with different DAWs

echo "=== CLAP Preset Discovery Test Helper ==="
echo ""
echo "Before testing, make sure plugins are installed:"
ls -la ~/.clap/*/presets/factory/*.json 2>/dev/null
echo ""

# Clear logs
rm -f /tmp/clapgo_factory_calls.log ~/clapgo_preset_debug.log

echo "Logs cleared. Monitoring for preset discovery activity..."
echo ""
echo "DAWs known to support CLAP preset discovery:"
echo "  - Bitwig Studio (excellent CLAP support)"
echo "  - REAPER (with CLAP extension)"
echo "  - Qtractor"
echo "  - Ardour (recent versions)"
echo ""
echo "When testing:"
echo "1. Load the DAW"
echo "2. Look for 'Scan presets' or 'Refresh preset database'"
echo "3. Try to load ClapGo Gain or Synth plugin"
echo "4. Check the preset browser/menu"
echo ""
echo "Monitoring logs (Ctrl+C to stop)..."

# Monitor in background
tail -f /tmp/clapgo_factory_calls.log 2>/dev/null &
TAIL_PID=$!

# Also check for preset discovery calls
while true; do
    if grep -q "preset-discovery" /tmp/clapgo_factory_calls.log 2>/dev/null; then
        echo ""
        echo "!!! PRESET DISCOVERY DETECTED !!!"
        echo "The DAW is requesting preset discovery!"
        break
    fi
    sleep 1
done

kill $TAIL_PID 2>/dev/null