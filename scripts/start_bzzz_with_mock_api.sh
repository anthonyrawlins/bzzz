#!/bin/bash

# Script to temporarily run bzzz with mock Hive API for testing
# This lets real bzzz agents do actual coordination with fake data

echo "🔧 Configuring Bzzz to use Mock Hive API"
echo "========================================"

# Stop the current bzzz service
echo "Stopping current bzzz service..."
sudo systemctl stop bzzz.service

# Wait a moment
sleep 2

# Set environment variables for mock API
export BZZZ_HIVE_API_URL="http://localhost:5000"
export BZZZ_LOG_LEVEL="debug"

echo "Starting bzzz with mock Hive API..."
echo "Mock API URL: $BZZZ_HIVE_API_URL"
echo ""
echo "🎯 The real bzzz agents will now:"
echo "   - Discover fake projects and tasks from mock API"
echo "   - Do actual P2P coordination on real dependencies" 
echo "   - Perform real antennae meta-discussion"
echo "   - Execute real coordination algorithms"
echo ""
echo "Watch your dashboard to see REAL coordination activity!"
echo ""

# Run bzzz directly with mock API configuration
cd /home/tony/AI/projects/Bzzz
/usr/local/bin/bzzz