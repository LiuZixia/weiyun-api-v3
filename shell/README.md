# Weiyun API V3 - Shell Client

This directory contains utility shell scripts to interact with the Weiyun MCP API via `mcporter`. It acts as a CLI bridge suitable for DevOps integration and general automation.

## Requirements
- `curl`
- `python3` (specifically utilized for local uploads via Python wrappers)
- `mcporter` NPM package (`npm install -g @modelcontextprotocol/mcporter`)

## Basic Setup

Since the API requests authorization via MCP Tokens over headers, we use an interactive tool wrapper format.

1. **Bootstrap the environment**:
   ```bash
   chmod +x setup.sh
   ./setup.sh
   ```

2. **Source variables manually**:
   ```bash
   source weiyun_env.sh "your_secret_mcp_token"
   ```

## Using `mcporter`

Once the token mapping is globally configured (`WEIYUN_MCP_TOKEN`), you can seamlessly script Weiyun functions through terminal RPC mappings.

### Examples

**1. Listing Files**
```bash
mcporter call --server weiyun --tool weiyun.list limit=10
```

**2. Downloading a File**
You must construct the nested JSON parameters for `items`:
```bash
mcporter call --server weiyun --tool weiyun.download \
  --args '{"items": [{"file_id": "file_123", "pdir_key": "parent_456"}]}'
```

**3. Generating a Share Link**
```bash
mcporter call --server weiyun --tool weiyun.gen_share_link \
  --args '{"file_list": [{"file_id": "f_abc", "pdir_key": "d_123"}], "share_name": "My_Share"}'
```

## Available Scripts

- `./setup.sh`: Initializer that dynamically builds `.env` profile mappings.
- `./test.sh`: Provides a swift integrity check to see if `mcporter` and Weiyun authentication bindings trigger a successful root-list lookup.
