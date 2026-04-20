#!/bin/bash

# GitHub Actions Runner Start Script
# This script starts the GitHub Actions self-hosted runner

set -e

# Configuration
RUNNER_DIR="/home/rhel/Work/agnos-devops/actions-runner"

# Check if runner directory exists
if [ ! -d "$RUNNER_DIR" ]; then
    echo "Error: Runner directory not found at $RUNNER_DIR"
    exit 1
fi

# Navigate to runner directory
cd "$RUNNER_DIR"

# Check if runner is already running
if pgrep -f "Runner.Listener" > /dev/null; then
    echo "GitHub Actions runner is already running"
    exit 0
fi

# Start the runner
echo "Starting GitHub Actions runner..."
./run.sh

echo "GitHub Actions runner started successfully"
