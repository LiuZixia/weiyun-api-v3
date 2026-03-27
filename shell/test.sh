#!/bin/bash
# test.sh - Example usage of Weiyun Shell Client

echo "Testing Weiyun Shell Client..."

if [ -z "$WEIYUN_MCP_TOKEN" ]; then
    echo "Please set WEIYUN_MCP_TOKEN environment variable first."
    echo "You can do this by running: source weiyun_env.sh <your_token>"
    exit 1
fi

echo "Listing top 5 files..."
# This requires mcporter to be installed
mcporter call --server weiyun --tool weiyun.list limit=5 

echo "Test complete!"
