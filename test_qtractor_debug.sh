#!/bin/bash

# Script to monitor debug logs for ClapGo plugin in real-time

echo "=== ClapGo Qtractor Debug Monitor ==="
echo "This script will monitor debug logs from the ClapGo plugin in real-time."
echo "Start Qtractor in another terminal and try to load the gain plugin."
echo ""
echo "Debug logs will be written to:"
echo "  - /tmp/clapgo_factory_calls.log (Factory operations)"
echo "  - /tmp/clapgo_plugin_init.log (Plugin lifecycle)"
echo ""
echo "Press Ctrl+C to stop monitoring."
echo "========================================="
echo ""

# Clear old logs
echo "Clearing old logs..."
rm -f /tmp/clapgo_factory_calls.log
rm -f /tmp/clapgo_plugin_init.log

# Create empty log files
touch /tmp/clapgo_factory_calls.log
touch /tmp/clapgo_plugin_init.log

# Start monitoring in split view using tail
echo "Starting log monitors..."
echo ""
echo "=== FACTORY CALLS LOG ==="

# Use multitail if available, otherwise use regular tail
if command -v multitail &> /dev/null; then
    multitail -i /tmp/clapgo_factory_calls.log -i /tmp/clapgo_plugin_init.log
else
    # Run tail commands in background
    tail -f /tmp/clapgo_factory_calls.log | sed 's/^/[FACTORY] /' &
    TAIL1_PID=$!
    
    tail -f /tmp/clapgo_plugin_init.log | sed 's/^/[PLUGIN]  /' &
    TAIL2_PID=$!
    
    # Wait for Ctrl+C
    trap "kill $TAIL1_PID $TAIL2_PID 2>/dev/null; exit" INT
    wait
fi