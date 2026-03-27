#!/bin/bash
# setup.sh - Setup Weiyun API V3 CLI environment

echo "Setting up Weiyun MCP Shell Environment..."

# Check requirements
command -v curl >/dev/null 2>&1 || { echo >&2 "Require curl but it's not installed. Aborting."; exit 1; }
command -v python3 >/dev/null 2>&1 || { echo >&2 "Require python3 for upload scripts but it's not installed. Aborting."; exit 1; }

# Assuming mcporter is installed via npm or similar
if ! command -v mcporter >/dev/null 2>&1; then
    echo "Warning: mcporter is not installed. You may need to run 'npm install -g @modelcontextprotocol/mcporter' or your equivalent."
fi

# Export environment variables helper
cat << 'EOF' > weiyun_env.sh
#!/bin/bash
if [ -z "$1" ]; then
    echo "Usage: source weiyun_env.sh <YOUR_MCP_TOKEN>"
    return 1
fi
export WEIYUN_MCP_TOKEN="$1"
export WEIYUN_MCP_URL="https://www.weiyun.com/api/v3/mcpserver"
echo "Weiyun Environment Variables Set!"
EOF

chmod +x weiyun_env.sh

echo "Setup complete. To start using:"
echo "1. Run: source weiyun_env.sh YOUR_TOKEN"
echo "2. Use mcporter call --server weiyun --tool weiyun.list limit=10"
